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
// revisited.
func sensitiveBucketMutationTestActions() []Action {
	return []Action{
		PutBucketPolicyAction,
		DeleteBucketPolicyAction,
		PutReplicationConfigurationAction,
		PutBucketLifecycleAction,
		PutBucketObjectLockConfigurationAction,
		PutBucketVersioningAction,
	}
}

// A bucket-level request (empty object name) for a sensitive bucket-mutation
// action must NOT be authorized by an object-only resource pattern such as
// "arn:aws:s3:::mybucket/*". This is the core of the #20449 hardening: an
// object-scoped grant must not reach these bucket-configuration writes.
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
	if statement.IsAllowed(Args{Action: PutBucketPolicyAction, BucketName: "mybucket"}) {
		t.Fatalf("s3:* via object pattern mybucket/* should not authorize PutBucketPolicy at bucket level")
	}
}

// Legitimate bucket-level grants — a bare bucket ARN or the "*" wildcard
// resource — must keep authorizing the sensitive actions. The hardening only
// withholds object-only patterns, not correctly scoped resources.
func TestStatementBucketMutationLegitimateGrantsAllowed(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	for _, pattern := range []string{"mybucket", "*"} {
		for _, action := range sensitiveBucketMutationTestActions() {
			statement := NewStatement("",
				Allow,
				NewActionSet(action),
				NewResourceSet(NewResource(pattern)),
				condition.NewFunctions(),
			)
			if !statement.IsAllowed(Args{Action: action, BucketName: "mybucket"}) {
				t.Fatalf("%v on resource %q should be allowed at bucket level", action, pattern)
			}
		}
	}
}

// Non-sensitive bucket-level actions (the compatibility-sensitive read/list
// family) must be UNCHANGED: an object-only pattern still authorizes them, so
// deployments relying on "bucket/*" for ListBucket do not break. Object-level
// requests are likewise untouched.
func TestStatementBucketMutationNonSensitiveUnchanged(t *testing.T) {
	SetLegacyBucketResourceMatch(false)

	listStatement := NewStatement("",
		Allow,
		NewActionSet(ListBucketAction),
		NewResourceSet(NewResource("mybucket/*")),
		condition.NewFunctions(),
	)
	if !listStatement.IsAllowed(Args{Action: ListBucketAction, BucketName: "mybucket"}) {
		t.Fatalf("ListBucket via mybucket/* must remain allowed (compatibility preserved)")
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
				NewActionSet(PutBucketPolicyAction),
				NewResourceSet(NewResource("mybucket/*")),
				condition.NewFunctions(),
			),
		},
	}
	if p.IsAllowed(Args{Action: PutBucketPolicyAction, BucketName: "mybucket"}) {
		t.Fatalf("Deny PutBucketPolicy on mybucket/* must still cover the bucket-level request")
	}
}

// The compatibility shim restores the historical behavior when explicitly
// enabled, so operators relying on the old semantics have an escape hatch.
func TestStatementBucketMutationLegacyShimRestores(t *testing.T) {
	SetLegacyBucketResourceMatch(true)
	t.Cleanup(func() { SetLegacyBucketResourceMatch(false) })

	statement := NewStatement("",
		Allow,
		NewActionSet(PutBucketPolicyAction),
		NewResourceSet(NewResource("mybucket/*")),
		condition.NewFunctions(),
	)
	if !statement.IsAllowed(Args{Action: PutBucketPolicyAction, BucketName: "mybucket"}) {
		t.Fatalf("legacy shim should restore mybucket/* granting PutBucketPolicy at bucket level")
	}
}
