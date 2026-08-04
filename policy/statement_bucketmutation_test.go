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
	"testing"

	"github.com/minio/pkg/v3/policy/condition"
)

// sensitiveBucketMutationTestActions mirrors sensitiveBucketMutationActions so
// the tests fail loudly if the protected set changes without the tests being
// revisited. Keep the two lists in sync by hand; the mirror test below
// enforces it.
func sensitiveBucketMutationTestActions() []Action {
	return []Action{
		DeleteBucketAction,
		ForceDeleteBucketAction,
		PutBucketPolicyAction,
		DeleteBucketPolicyAction,
		PutReplicationConfigurationAction,
		PutBucketLifecycleAction,
		PutBucketObjectLockConfigurationAction,
		PutBucketVersioningAction,
		PutBucketCorsAction,
		DeleteBucketCorsAction,
		PutBucketQOSAction,
		PutInventoryConfigurationAction,
	}
}

// deliberatelyUnprotectedBucketWrites are bucket-level writes that the
// hardening leaves reachable through an object-only pattern on purpose. None
// of them grants the caller access it does not already hold via the object
// grant, while a tenant told "this bucket is yours" may legitimately use them.
// See the rationale on sensitiveBucketMutationActions.
func deliberatelyUnprotectedBucketWrites() []Action {
	return []Action{
		PutBucketTaggingAction,
		PutBucketEncryptionAction,
		PutBucketNotificationAction,
	}
}

// The test mirror and the protected set must describe the same actions, so a
// change to either one without the other fails immediately.
func TestStatementBucketMutationSetMatchesMirror(t *testing.T) {
	mirror := sensitiveBucketMutationTestActions()
	if len(mirror) != len(sensitiveBucketMutationActions) {
		t.Fatalf("test mirror lists %d actions but the protected set has %d — update both together",
			len(mirror), len(sensitiveBucketMutationActions))
	}
	for _, action := range mirror {
		if !isSensitiveBucketMutation(action) {
			t.Fatalf("mirror action %v is missing from the protected set", action)
		}
	}
}

// Every protected action must be a supported, bucket-only action. An object
// action in the set would be inert at best and misleading at worst (e.g.
// ResetBucketReplicationState IS an object action despite its name).
func TestStatementBucketMutationSetIsBucketOnlyWrites(t *testing.T) {
	for action := range sensitiveBucketMutationActions {
		if _, ok := SupportedActions[action]; !ok {
			t.Fatalf("protected action %v is not in SupportedActions", action)
		}
		if _, ok := SupportedObjectActions[action]; ok {
			t.Fatalf("protected action %v is an object action; the hardening covers bucket-only actions", action)
		}
	}
}

// A bucket-level request (empty object name) for a sensitive bucket-mutation
// action must NOT be authorized by an object-only resource pattern such as
// "arn:aws:s3:::mybucket/*". This is the core of the #20449 hardening: an
// object-scoped grant must not reach these bucket-level writes.
func TestStatementBucketMutationObjectPatternDenied(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, action := range sensitiveBucketMutationTestActions() {
		statement := NewStatement("",
			Allow,
			NewActionSet(action),
			NewResourceSet(NewResource("mybucket/*")),
			condition.NewFunctions(),
		)
		if statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("%v via object pattern mybucket/* should be denied at bucket level", action)
		}
	}
}

// The wildcard action set ("s3:*") granted on an object-only pattern must also
// fail to authorize a sensitive bucket mutation, because the request action is
// concrete even though the statement action is a wildcard.
func TestStatementBucketMutationWildcardActionDenied(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	statement := NewStatement("",
		Allow,
		NewActionSet(AllActions),
		NewResourceSet(NewResource("mybucket/*")),
		condition.NewFunctions(),
	)
	for _, action := range sensitiveBucketMutationTestActions() {
		if statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("s3:* via object pattern mybucket/* should not authorize %v at bucket level", action)
		}
	}
}

// Legitimate bucket-level grants — a bare bucket ARN, the "*" wildcard, a
// slashless prefix wildcard, or the conventional {bucket, bucket/*} pair —
// must keep authorizing the sensitive actions. The hardening only withholds
// object-only patterns, not correctly scoped resources.
func TestStatementBucketMutationLegitimateGrantsAllowed(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	resourceSets := []ResourceSet{
		NewResourceSet(NewResource("mybucket")),
		NewResourceSet(NewResource("*")),
		NewResourceSet(NewResource("mybucket*")),
		NewResourceSet(NewResource("mybucket"), NewResource("mybucket/*")),
	}
	for _, resources := range resourceSets {
		for _, action := range sensitiveBucketMutationTestActions() {
			statement := NewStatement("",
				Allow,
				NewActionSet(action),
				resources,
				condition.NewFunctions(),
			)
			if !statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
				t.Fatalf("%v on resources %v should be allowed at bucket level", action, resources)
			}
		}
	}
}

// Three bucket-level writes are left reachable through an object-only pattern
// ON PURPOSE, because none of them grants the caller access it does not already
// hold and each has a plausible legitimate use by a tenant that owns the bucket.
// This test states that decision so re-adding any of them is a deliberate act
// with a visible compatibility cost, never an accident.
func TestStatementBucketMutationDeliberatelyUnprotected(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, action := range deliberatelyUnprotectedBucketWrites() {
		if isSensitiveBucketMutation(action) {
			t.Fatalf("%v is in the protected set; adding it is a compatibility decision that must be made explicitly, not by editing the set alone", action)
		}
		statement := NewStatement("",
			Allow,
			NewActionSet(action),
			NewResourceSet(NewResource("mybucket/*")),
			condition.NewFunctions(),
		)
		if !statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("%v via mybucket/* must remain allowed (deliberately unprotected)", action)
		}
	}
}

// Non-sensitive bucket-level actions must be UNCHANGED: an object-only pattern
// still authorizes the compatibility-sensitive read/list family, and — by
// deliberate decision — CreateBucket, which targets a bucket that does not
// exist yet and is commonly performed by provisioning flows holding only
// tenant-scoped ("bucket/*") credentials. Object-level requests are likewise
// untouched.
func TestStatementBucketMutationNonSensitiveUnchanged(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, action := range []Action{ListBucketAction, GetBucketLocationAction, CreateBucketAction} {
		statement := NewStatement("",
			Allow,
			NewActionSet(action),
			NewResourceSet(NewResource("mybucket/*")),
			condition.NewFunctions(),
		)
		if !statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("%v via mybucket/* must remain allowed (compatibility preserved)", action)
		}
	}

	objStatement := NewStatement("",
		Allow,
		NewActionSet(PutObjectAction),
		NewResourceSet(NewResource("mybucket/*")),
		condition.NewFunctions(),
	)
	if !objStatement.IsAllowed(Args{Action: PutObjectAction, BucketName: "mybucket", ObjectName: "obj"}) {
		t.Fatalf("PutObject on an object must remain allowed via mybucket/*")
	}
}

// The fix must not weaken the Deny direction: an admin who blocks a bucket with
// "Deny ... on bucket/*" must keep that protection for bucket-level actions.
func TestStatementBucketMutationDenyNotWeakened(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, action := range sensitiveBucketMutationTestActions() {
		p := Policy{
			Version: DefaultVersion,
			Statements: []Statement{
				NewStatement("",
					Allow,
					NewActionSet(AllActions),
					NewResourceSet(NewResource("*")),
					condition.NewFunctions(),
				),
				NewStatement("",
					Deny,
					NewActionSet(action),
					NewResourceSet(NewResource("mybucket/*")),
					condition.NewFunctions(),
				),
			},
		}
		if p.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("Deny %v on mybucket/* must still cover the bucket-level request", action)
		}
	}
}

// The fix must not weaken NotResource exclusions either. An Allow statement
// with "NotResource: bucket/*" historically did NOT apply to bucket-level
// requests on that bucket (the "bucket/" form matched the exclusion). If the
// hardening matched NotResource against the bare bucket name, the exclusion
// would stop matching and the Allow would BROADEN — the opposite of hardening.
// The exclusion must keep its historical reach while grants are narrowed.
func TestStatementBucketMutationNotResourceExclusionNotWeakened(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, action := range sensitiveBucketMutationTestActions() {
		statement := NewStatementWithNotResource("",
			Allow,
			NewActionSet(AllActions),
			NewResourceSet(NewResource("mybucket/*")),
			condition.NewFunctions(),
		)
		if statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("Allow with NotResource mybucket/* must keep excluding bucket-level %v on mybucket", action)
		}
		if !statement.IsAllowed(Args{Action: action, BucketName: "otherbucket"}) {
			t.Fatalf("Allow with NotResource mybucket/* must still allow %v on other buckets", action)
		}
	}
}

// Withholding the trailing slash changes the string a resource pattern is
// matched against, and a pattern can match the bare bucket name without having
// matched the historical "bucket/" form — a fixed-width wildcard is the clean
// example: "mybucke?" matches "mybucket" but not "mybucket/". Honoring the bare
// form alone would therefore ADD a grant the historical matcher refused, which
// is the one thing this hardening must never do. The protected path requires
// both forms, so no pattern is newly satisfied.
func TestStatementBucketMutationFixedWidthWildcardNotBroadened(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, action := range sensitiveBucketMutationTestActions() {
		statement := NewStatement("",
			Allow,
			NewActionSet(action),
			NewResourceSet(NewResource("mybucke?")),
			condition.NewFunctions(),
		)
		if statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("%v on fixed-width wildcard mybucke? must stay denied; it never matched the historical \"mybucket/\" form", action)
		}
	}

	// The same pattern keeps its historical meaning everywhere else: it never
	// authorized these bucket-level requests, and it still does not.
	listStatement := NewStatement("",
		Allow,
		NewActionSet(ListBucketAction),
		NewResourceSet(NewResource("mybucke?")),
		condition.NewFunctions(),
	)
	if listStatement.IsAllowed(Args{Action: ListBucketAction, BucketName: "mybucket"}) {
		t.Fatalf("ListBucket on mybucke? was never authorized at bucket level; behavior must be unchanged")
	}
}

// A trailing wildcard absorbs the historical slash, so patterns like
// "mybucket*" matched both forms all along and must keep working.
func TestStatementBucketMutationTrailingWildcardStillGrants(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, action := range sensitiveBucketMutationTestActions() {
		statement := NewStatement("",
			Allow,
			NewActionSet(action),
			NewResourceSet(NewResource("mybucket*")),
			condition.NewFunctions(),
		)
		if !statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("%v on mybucket* must remain allowed; the pattern matched the historical form too", action)
		}
	}
}

// The compatibility shim restores the historical behavior when explicitly
// enabled, so operators relying on the old semantics have an escape hatch.
func TestStatementBucketMutationLegacyShimRestores(t *testing.T) {
	SetLegacyBucketResourceMatch(true)
	t.Cleanup(func() { SetLegacyBucketResourceMatch(false) })

	for _, action := range sensitiveBucketMutationTestActions() {
		statement := NewStatement("",
			Allow,
			NewActionSet(action),
			NewResourceSet(NewResource("mybucket/*")),
			condition.NewFunctions(),
		)
		if !statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
			t.Fatalf("legacy shim should restore mybucket/* granting %v at bucket level", action)
		}
	}
}
