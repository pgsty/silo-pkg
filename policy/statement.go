// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package policy

import (
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/minio/pkg/v3/policy/condition"
	"github.com/zeebo/xxh3"
)

// Statement - iam policy statement.
type Statement struct {
	SID          ID                  `json:"Sid,omitempty"`
	Effect       Effect              `json:"Effect"`
	Actions      ActionSet           `json:"Action,omitempty"`
	NotActions   ActionSet           `json:"NotAction,omitempty"`
	Resources    ResourceSet         `json:"Resource,omitempty"`
	NotResources ResourceSet         `json:"NotResource,omitempty"`
	Conditions   condition.Functions `json:"Condition,omitempty"`
}

// smallBufPool should always return a non-nil *bytes.Buffer
var smallBufPool = sync.Pool{
	New: func() interface{} { return &bytes.Buffer{} },
}

// legacyBucketResourceMatch, when true, restores the historical behavior in
// which a bucket-level request (empty object name) is matched against policy
// resources as "bucket/", so an object-only pattern such as
// "arn:aws:s3:::bucket/*" also matches bucket-level actions. That behavior
// over-grants on Allow statements (and over-blocks on Deny statements) and is
// withheld by default for the sensitive bucket-level write actions listed in
// sensitiveBucketMutationActions. It is meant to be configured once at server
// startup (see SetLegacyBucketResourceMatch) and only read while serving.
var legacyBucketResourceMatch atomic.Bool

// SetLegacyBucketResourceMatch toggles the compatibility shim documented on
// legacyBucketResourceMatch. Call it once during startup, before serving
// requests; it is safe to read concurrently afterwards.
func SetLegacyBucketResourceMatch(enabled bool) {
	legacyBucketResourceMatch.Store(enabled)
}

func init() {
	// Fork-specific escape hatch: MINIO_API_LEGACY_BUCKET_RESOURCE_MATCH=on,
	// read once at package initialization, restores the historical bucket
	// resource matching described on legacyBucketResourceMatch. Kept self
	// contained here (rather than wired from the server) so the hardening and
	// its opt-out ship as a single, independently buildable change. Callers may
	// still override it explicitly via SetLegacyBucketResourceMatch.
	if os.Getenv("MINIO_API_LEGACY_BUCKET_RESOURCE_MATCH") == "on" {
		legacyBucketResourceMatch.Store(true)
	}
}

// sensitiveBucketMutationActions is the set of bucket-level write actions for
// which an object-only resource pattern ("bucket/*") must not be honored as a
// bucket-level grant.
//
// Membership is decided by one question: does reaching this action give the
// caller something the object-scoped grant does not already give them? The
// resource bug only fires when the statement grants the bucket-level action in
// the first place, which in practice means "s3:*" — so the affected principal
// already holds full read/write/delete over every object in the bucket. An
// action that merely reconfigures the bucket adds nothing to that position; an
// action that hands out access, defeats a protection the owner set, or
// outlives the grant does:
//
//   - PutBucketPolicy / DeleteBucketPolicy — grant access to other principals,
//     including anonymous, or grant the caller bucket-level actions it was
//     never given. Self-escalation and public exposure;
//   - PutBucketObjectLockConfiguration — defeat WORM retention, which exists
//     precisely to stop a holder of write access from destroying data;
//   - PutBucketVersioning — defeat version history, same class of protection;
//   - PutReplicationConfiguration — copy the bucket to a chosen target under
//     server credentials, and keep copying after the caller's access is gone;
//   - PutBucketLifecycle — schedule expiry that likewise outlives the grant;
//   - DeleteBucket / ForceDeleteBucket — destroy the bucket entity and its
//     configuration irreversibly. This is the reproduction reported in
//     upstream minio/minio issue #20449;
//   - PutBucketCors / DeleteBucketCors / PutBucketQOS /
//     PutInventoryConfiguration — no MinIO server behavior is attached to
//     these today (no handler, or a handler that returns NotImplemented after
//     the authorization check), so withholding them costs nothing and they are
//     covered in advance should a handler ever be wired.
//
// Deliberately NOT protected, and asserted as such by
// TestStatementBucketMutationDeliberatelyUnprotected:
//
//   - PutBucketTagging, PutBucketEncryption, PutBucketNotification — a tenant
//     handed "s3:*" on "bucket/*" and told the bucket is theirs may legitimately
//     tag it, set default encryption, or wire event notifications. None of the
//     three grants the caller access it does not already have; the harm is to
//     the owner's posture, not to the access boundary. Low security gain against
//     a real compatibility cost, so they keep the historical matching until a
//     migration-gated change;
//   - CreateBucket — targets a bucket that does not exist yet, so there is
//     nothing to mutate or destroy, and provisioning flows commonly create a
//     tenant's bucket with that tenant's own "bucket/*" credentials;
//   - the read/list family (ListBucket, GetBucketLocation, the configuration
//     reads) — many existing deployments do grant those through "bucket/*",
//     and breaking them is what got upstream's own attempt reverted.
//
// See github.com/minio/minio issue #20449.
var sensitiveBucketMutationActions = map[Action]struct{}{
	DeleteBucketAction:                     {},
	ForceDeleteBucketAction:                {},
	PutBucketPolicyAction:                  {},
	DeleteBucketPolicyAction:               {},
	PutReplicationConfigurationAction:      {},
	PutBucketLifecycleAction:               {},
	PutBucketObjectLockConfigurationAction: {},
	PutBucketVersioningAction:              {},
	PutBucketCorsAction:                    {},
	DeleteBucketCorsAction:                 {},
	PutBucketQOSAction:                     {},
	PutInventoryConfigurationAction:        {},
}

// isSensitiveBucketMutation reports whether action is one of the bucket
// configuration writes protected from object-only ("bucket/*") grants.
func isSensitiveBucketMutation(action Action) bool {
	_, ok := sensitiveBucketMutationActions[action]
	return ok
}

// IsAllowed - checks given policy args is allowed to continue the Rest API.
func (statement Statement) IsAllowed(args Args) bool {
	return statement.IsAllowedPtr(&args)
}

// IsAllowedPtr - checks given policy args is allowed to continue the Rest API.
func (statement Statement) IsAllowedPtr(args *Args) bool {
	check := func() bool {
		if (!statement.Actions.Match(args.Action) && !statement.Actions.IsEmpty()) ||
			statement.NotActions.Match(args.Action) {
			return false
		}

		resource := smallBufPool.Get().(*bytes.Buffer)
		defer smallBufPool.Put(resource)
		resource.Reset()
		resource.WriteString(args.BucketName)
		withheldBucketSlash := false
		if args.ObjectName != "" {
			if !strings.HasPrefix(args.ObjectName, "/") {
				resource.WriteByte('/')
			}
			resource.WriteString(args.ObjectName)
		} else if args.BucketName == "" {
			// Preserve the "/" sentinel used by KMS two-phase authorization
			// (empty bucket and empty object), which the isKMS() path below
			// relies on.
			resource.WriteByte('/')
		} else if legacyBucketResourceMatch.Load() ||
			statement.Effect != Allow ||
			!isSensitiveBucketMutation(args.Action) {
			// Append the trailing slash so an object resource pattern such as
			// "bucket/*" keeps matching this bucket-level request. This is the
			// historical behavior and is retained for every case except an
			// Allow statement being evaluated for one of the sensitive
			// bucket-level writes, where honoring "bucket/*" as a bucket-level
			// grant is a privilege-escalation vector. Deny statements keep the
			// slash too, so no Deny is ever weakened.
			resource.WriteByte('/')
		} else {
			// Bucket-level Allow request for a sensitive bucket-level write,
			// with the legacy shim off: the resource stays the bare bucket
			// name, so an object-only pattern ("bucket/*") no longer
			// authorizes it; bare-bucket and "*" patterns still match. The
			// omission applies to the Resources grant only — remember it so
			// the NotResources check below can restore the historical form.
			withheldBucketSlash = true
		}

		if statement.isTable() && !TableAction(args.Action).IsValid() {
			// When a tables policy statement (for example
			//   "Action":   ["s3tables:GetTableData"],
			//   "Resource": ["arn:aws:s3tables:::bucket/wh/table/uuid"]
			// ) is evaluated for a plain S3 data-path action such as
			// GetObject on (BucketName "wh", ObjectName "uuid[/...]"), the
			// action match succeeds via implicitActions. However, the
			// resource string built from Args ("wh/uuid[/...]") does not
			// look like a tables ARN suffix ("bucket/wh/table/uuid"), so a
			// direct string match against the S3 Tables resource
			// would fail. In this specific case we know:
			//   - the statement is a tables statement,
			//   - the incoming action is covered implicitly (not a table API),
			//   - and the stored policy resource is S3 Tables style.
			// To allow GetObject/ListMultipartUploadParts/etc. when
			// s3tables:GetTableData (or similar) is granted, normalize the
			// S3 data-path resource into the canonical tables form before
			// running the usual resource match.
			if !isTableResourceString(resource.String()) {
				if args.BucketName == "" || args.ObjectName == "" {
					return false
				}
				objectName := args.ObjectName
				if idx := strings.IndexByte(objectName, '/'); idx >= 0 {
					objectName = objectName[:idx]
				}
				resource.Reset()
				resource.WriteString("bucket/")
				resource.WriteString(args.BucketName)
				resource.WriteString("/table/")
				resource.WriteString(objectName)
				if !isTableResourceString(resource.String()) {
					return false
				}
			}
		}

		if statement.isKMS() {
			if resource.Len() == 1 && resource.String() == "/" || len(statement.Resources) == 0 {
				// In previous MinIO versions, KMS statements ignored Resources, so if len(statement.Resources) == 0,
				// allow backward compatibility by not trying to Match.

				// When resource is "/", this allows evaluating KMS statements while explicitly excluding Resource,
				// by passing Args with empty BucketName and ObjectName. This is useful when doing a
				// two-phase authorization of a request.
				return statement.Conditions.Evaluate(args.ConditionValues)
			}
		}

		// Admin actions that do not operate on a bucket resource
		// skip resource matching entirely. For the small set of
		// bucket-scoped admin actions (e.g. SetBucketQuota),
		// resource matching is enforced when Resources are present.
		ignoreResourceMatch := statement.isSTS() || (statement.isAdmin() && !statement.hasAdminResource())

		if !ignoreResourceMatch && len(statement.Resources) > 0 && !statement.Resources.Match(resource.String(), args.ConditionValues) {
			return false
		}

		if withheldBucketSlash {
			// Restore the historical "bucket/" form. It is needed twice below,
			// and in both cases for the same reason: this hardening must only
			// ever remove a grant, never create one.
			resource.WriteByte('/')

			// The bare bucket name is a different string, so a pattern can
			// match it that did not match "bucket/" — a fixed-width wildcard
			// such as "mybucke?" matches "mybucket" but not "mybucket/".
			// Honoring the bare form alone would therefore ADD a grant the
			// historical matcher refused. Requiring both forms makes the
			// protected path an intersection with the historical decision, so
			// it is monotone by construction rather than by argument.
			if !ignoreResourceMatch && len(statement.Resources) > 0 && !statement.Resources.Match(resource.String(), args.ConditionValues) {
				return false
			}

			// NotResource is an exclusion: matching it against the bare bucket
			// name would narrow the exclusion — and thereby broaden this Allow
			// statement — for exactly the writes the hardening protects. The
			// restored form below keeps every pre-existing
			// "NotResource: bucket/*" exclusion at its full reach.
		}

		if !ignoreResourceMatch && len(statement.NotResources) > 0 && statement.NotResources.Match(resource.String(), args.ConditionValues) {
			return false
		}

		return statement.Conditions.Evaluate(args.ConditionValues)
	}

	return statement.Effect.IsAllowed(check())
}

// validateActionTypes rejects statements that mix actions from
// different namespaces (e.g. s3 + admin, admin + sts). Each
// statement must contain actions from exactly one namespace.
func (statement Statement) validateActionTypes() error {
	actions := statement.Actions
	if len(actions) == 0 {
		actions = statement.NotActions
	}
	var hasS3, hasAdmin, hasSTS, hasKMS, hasTable, hasVectors bool
	for action := range actions {
		switch {
		case AdminAction(action).IsValid():
			hasAdmin = true
		case STSAction(action).IsValid():
			hasSTS = true
		case KMSAction(action).IsValid():
			hasKMS = true
		case TableAction(action).IsValid():
			hasTable = true
		case VectorsAction(action).IsValid():
			hasVectors = true
		default:
			hasS3 = true
		}
	}
	count := 0
	for _, b := range []bool{hasS3, hasAdmin, hasSTS, hasKMS, hasTable, hasVectors} {
		if b {
			count++
		}
	}
	if count > 1 {
		return Errorf("mixing action types in the same statement is not allowed")
	}
	return nil
}

func (statement Statement) isAdmin() bool {
	for action := range statement.Actions {
		if AdminAction(action).IsValid() {
			return true
		}
	}
	return false
}

// hasAdminResource reports whether any action in this statement is a
// bucket-scoped admin action (as declared by AdminActionsWithResource).
func (statement Statement) hasAdminResource() bool {
	for action := range statement.Actions {
		if AdminAction(action).HasResource() {
			return true
		}
	}
	return false
}

func (statement Statement) isSTS() bool {
	for action := range statement.Actions {
		if STSAction(action).IsValid() {
			return true
		}
	}
	return false
}

func (statement Statement) isKMS() bool {
	for action := range statement.Actions {
		if KMSAction(action).IsValid() {
			return true
		}
	}
	return false
}

func (statement Statement) isTable() bool {
	for action := range statement.Actions {
		if TableAction(action).IsValid() {
			return true
		}
	}
	return false
}

func (statement Statement) isVectors() bool {
	for action := range statement.Actions {
		if VectorsAction(action).IsValid() {
			return true
		}
	}
	return false
}

// isValid - checks whether statement is valid or not.
func (statement Statement) isValid() error {
	if !statement.Effect.IsValid() {
		return Errorf("invalid Effect %v", statement.Effect)
	}

	if len(statement.Actions) == 0 && len(statement.NotActions) == 0 {
		return Errorf("Action must not be empty")
	}

	if len(statement.Actions) > 0 && len(statement.NotActions) > 0 {
		return Errorf("Action and NotAction cannot be specified in the same statement")
	}

	if err := statement.validateActionTypes(); err != nil {
		return err
	}

	if statement.isAdmin() {
		return statement.validateAdmin(false)
	}

	if statement.isSTS() {
		if err := statement.Actions.ValidateSTS(); err != nil {
			return err
		}
		for action := range statement.Actions {
			keys := statement.Conditions.Keys()
			keyDiff := keys.Difference(stsActionConditionKeyMap[action])
			if !keyDiff.IsEmpty() {
				return Errorf("unsupported condition keys '%v' used for action '%v'", keyDiff, action)
			}
		}
		return nil
	}

	if statement.isKMS() {
		if err := statement.Actions.ValidateKMS(); err != nil {
			return err
		}
		if err := statement.Resources.ValidateKMS(); err != nil {
			return err
		}
		if err := statement.NotResources.ValidateKMS(); err != nil {
			return err
		}
		return nil
	}

	if statement.isTable() {
		if err := statement.Actions.ValidateTable(); err != nil {
			return err
		}
		for action := range statement.Actions {
			keys := statement.Conditions.Keys()
			keyDiff := keys.Difference(tableActionConditionKeyMap[action])
			if !keyDiff.IsEmpty() {
				return Errorf("unsupported condition keys '%v' used for action '%v'", keyDiff, action)
			}
		}

		if len(statement.Resources) == 0 && len(statement.NotResources) == 0 {
			return Errorf("Resource must not be empty")
		}

		if len(statement.Resources) > 0 && len(statement.NotResources) > 0 {
			return Errorf("Resource and NotResource cannot be specified in the same statement")
		}

		if err := statement.Resources.ValidateTable(); err != nil {
			return err
		}

		if err := statement.NotResources.ValidateTable(); err != nil {
			return err
		}

		for action := range statement.Actions {
			if len(statement.Resources) > 0 && !statement.Resources.ObjectResourceExists() && !statement.Resources.BucketResourceExists() {
				return Errorf("unsupported Resource found %v for action %v", statement.Resources, action)
			}
			if len(statement.NotResources) > 0 && !statement.NotResources.ObjectResourceExists() && !statement.NotResources.BucketResourceExists() {
				return Errorf("unsupported NotResource found %v for action %v", statement.NotResources, action)
			}
		}

		return nil
	}

	if statement.isVectors() {
		if err := statement.Actions.ValidateVectors(); err != nil {
			return err
		}
		for action := range statement.Actions {
			keys := statement.Conditions.Keys()
			keyDiff := keys.Difference(VectorsActionConditionKeyMap[action])
			if !keyDiff.IsEmpty() {
				return Errorf("unsupported condition keys '%v' used for action '%v'", keyDiff, action)
			}
		}

		if len(statement.Resources) == 0 && len(statement.NotResources) == 0 {
			return Errorf("Resource must not be empty")
		}

		if len(statement.Resources) > 0 && len(statement.NotResources) > 0 {
			return Errorf("Resource and NotResource cannot be specified in the same statement")
		}

		if err := statement.Resources.ValidateVectors(); err != nil {
			return err
		}

		if err := statement.NotResources.ValidateVectors(); err != nil {
			return err
		}

		for action := range statement.Actions {
			if len(statement.Resources) > 0 && !statement.Resources.ObjectResourceExists() && !statement.Resources.BucketResourceExists() {
				return Errorf("unsupported Resource found %v for action %v", statement.Resources, action)
			}
			if len(statement.NotResources) > 0 && !statement.NotResources.ObjectResourceExists() && !statement.NotResources.BucketResourceExists() {
				return Errorf("unsupported NotResource found %v for action %v", statement.NotResources, action)
			}
		}

		return nil
	}

	if !statement.SID.IsValid() {
		return Errorf("invalid SID %v", statement.SID)
	}

	if len(statement.Resources) == 0 && len(statement.NotResources) == 0 {
		return Errorf("Resource must not be empty")
	}

	if len(statement.Resources) > 0 && len(statement.NotResources) > 0 {
		return Errorf("Resource and NotResource cannot be specified in the same statement")
	}

	if err := statement.Resources.ValidateS3(); err != nil {
		return err
	}

	if err := statement.NotResources.ValidateS3(); err != nil {
		return err
	}

	if err := statement.Actions.Validate(); err != nil {
		return err
	}

	for action := range statement.Actions {
		if len(statement.Resources) > 0 && !statement.Resources.ObjectResourceExists() && !statement.Resources.BucketResourceExists() {
			return Errorf("unsupported Resource found %v for action %v", statement.Resources, action)
		}
		if len(statement.NotResources) > 0 && !statement.NotResources.ObjectResourceExists() && !statement.NotResources.BucketResourceExists() {
			return Errorf("unsupported NotResource found %v for action %v", statement.NotResources, action)
		}

		keys := statement.Conditions.Keys()
		keyDiff := keys.Difference(IAMActionConditionKeyMap.Lookup(action))
		if !keyDiff.IsEmpty() {
			return Errorf("unsupported condition keys '%v' used for action '%v'", keyDiff, action)
		}
	}

	return nil
}

// isValidStrict applies additional checks for new policies. It rejects
// admin statements that specify both Resource and NotResource, enforces
// that bucket-scoped admin actions have well-formed Resource ARNs, etc.
// Servers call this path when creating or updating policies; the
// permissive isValid path is used when loading existing policies.
func (statement Statement) isValidStrict() error {
	if !statement.Effect.IsValid() {
		return Errorf("invalid Effect %v", statement.Effect)
	}

	if len(statement.Actions) == 0 && len(statement.NotActions) == 0 {
		return Errorf("Action must not be empty")
	}

	if len(statement.Actions) > 0 && len(statement.NotActions) > 0 {
		return Errorf("Action and NotAction cannot be specified in the same statement")
	}

	if err := statement.validateActionTypes(); err != nil {
		return err
	}

	if statement.isAdmin() {
		return statement.validateAdmin(true)
	}

	// For non-admin types, delegate to the standard path which
	// already enforces Resource/NotResource conflicts.
	return statement.isValid()
}

// validateAdmin validates an admin statement. When strict is true,
// Resource and NotResource in the same statement is rejected, and
// Resources on bucket-scoped actions are validated as S3 ARNs.
func (statement Statement) validateAdmin(strict bool) error {
	if err := statement.Actions.ValidateAdmin(); err != nil {
		return err
	}
	for action := range statement.Actions {
		keys := statement.Conditions.Keys()
		keyDiff := keys.Difference(adminActionConditionKeyMap[action])
		if !keyDiff.IsEmpty() {
			return Errorf("unsupported condition keys '%v' used for action '%v'", keyDiff, action)
		}
	}
	if strict {
		if len(statement.Resources) > 0 && len(statement.NotResources) > 0 {
			return Errorf("Resource and NotResource cannot be specified in the same admin statement")
		}
		if len(statement.Resources) > 0 && statement.hasAdminResource() {
			if err := statement.Resources.ValidateS3(); err != nil {
				return err
			}
		}
		if len(statement.NotResources) > 0 && statement.hasAdminResource() {
			if err := statement.NotResources.ValidateS3(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate - validates Statement is for given bucket or not.
func (statement Statement) Validate() error {
	return statement.isValid()
}

// ValidateStrict validates the statement with strict rules suitable
// for new policy creation. See isValidStrict for details.
func (statement Statement) ValidateStrict() error {
	return statement.isValidStrict()
}

// Equals checks if two statements are equal
func (statement Statement) Equals(st Statement) bool {
	if statement.Effect != st.Effect {
		return false
	}
	if !statement.Actions.Equals(st.Actions) {
		return false
	}
	if !statement.NotActions.Equals(st.NotActions) {
		return false
	}
	if !statement.Resources.Equals(st.Resources) {
		return false
	}
	if !statement.NotResources.Equals(st.NotResources) {
		return false
	}
	if !statement.Conditions.Equals(st.Conditions) {
		return false
	}
	return true
}

// Clone clones Statement structure
func (statement Statement) Clone() Statement {
	return Statement{
		SID:          statement.SID,
		Effect:       statement.Effect,
		Actions:      statement.Actions.Clone(),
		NotActions:   statement.NotActions.Clone(),
		Resources:    statement.Resources.Clone(),
		NotResources: statement.NotResources.Clone(),
		Conditions:   statement.Conditions.Clone(),
	}
}

// NewStatement - creates new statement.
func NewStatement(sid ID, effect Effect, actionSet ActionSet, resourceSet ResourceSet, conditions condition.Functions) Statement {
	return Statement{
		SID:        sid,
		Effect:     effect,
		Actions:    actionSet,
		Resources:  resourceSet,
		Conditions: conditions,
	}
}

// NewStatementWithNotResource - creates new statement with NotAction.
func NewStatementWithNotResource(sid ID, effect Effect, actions ActionSet, notResources ResourceSet, conditions condition.Functions) Statement {
	return Statement{
		SID:          sid,
		Effect:       effect,
		Actions:      actions,
		NotResources: notResources,
		Conditions:   conditions,
	}
}

// NewStatementWithNotAction - creates new statement with NotAction.
func NewStatementWithNotAction(sid ID, effect Effect, notActions ActionSet, resources ResourceSet, conditions condition.Functions) Statement {
	return Statement{
		SID:        sid,
		Effect:     effect,
		NotActions: notActions,
		Resources:  resources,
		Conditions: conditions,
	}
}

// Equals checks if two statements are equal
func (statement Statement) hash(seed uint64) [16]byte {
	// Order independent xor.
	xorTo := func(dst *xxh3.Uint128, v xxh3.Uint128) {
		dst.Lo ^= v.Lo
		dst.Hi ^= v.Hi
	}
	// Add value with seed.
	xorInt := func(dst *xxh3.Uint128, n int, seed uint64) {
		var tmp [8]byte
		binary.LittleEndian.PutUint64(tmp[:], uint64(n))
		xorTo(dst, xxh3.Hash128Seed(tmp[:], seed))
	}

	h := xxh3.HashString128Seed(string(statement.Effect), seed)

	xorInt(&h, len(statement.Actions), seed+1)
	for action := range statement.Actions {
		xorTo(&h, xxh3.HashString128Seed(string(action), seed+2))
	}

	xorInt(&h, len(statement.NotActions), seed+3)
	for action := range statement.NotActions {
		xorTo(&h, xxh3.HashString128Seed(string(action), seed+4))
	}

	xorInt(&h, len(statement.Resources), seed+5)
	for res := range statement.Resources {
		xorTo(&h, xxh3.HashString128Seed(res.Pattern+res.Type.String(), seed+6))
	}

	xorInt(&h, len(statement.Conditions), seed+7)
	for _, cond := range statement.Conditions {
		xorTo(&h, xxh3.HashString128Seed(cond.String(), seed+8))
	}
	return h.Bytes()
}
