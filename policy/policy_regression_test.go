// Copyright (c) 2026 Feng Ruohang
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

func TestPolicyParsingPreservesNotResourceDenies(t *testing.T) {
	for _, count := range []int{4, 10, 11, 20} {
		p := Policy{
			Version: DefaultVersion,
			Statements: []Statement{
				{Effect: Allow, Actions: NewActionSet(GetObjectAction), Resources: NewResourceSet(NewResource("*"))},
				{Effect: Deny, Actions: NewActionSet(GetObjectAction), NotResources: NewResourceSet(NewResource("public/*"), NewResource("shared/*"))},
				{Effect: Deny, Actions: NewActionSet(GetObjectAction), NotResources: NewResourceSet(NewResource("other/*"), NewResource("shared/*"))},
			},
		}
		for len(p.Statements) < count {
			p.Statements = append(p.Statements, Statement{
				Effect: Allow, Actions: NewActionSet(PutObjectAction),
				Resources: NewResourceSet(NewResource(fmt.Sprintf("filler%d/*", len(p.Statements)))),
			})
		}
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, parser := range []struct {
			name  string
			parse func(io.Reader) (*Policy, error)
		}{
			{"read", ParseConfig},
			{"write", ParseConfigStrict},
		} {
			t.Run(fmt.Sprintf("%s/%d", parser.name, count), func(t *testing.T) {
				parsed, err := parser.parse(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				if got := len(parsed.Statements); got != count {
					t.Errorf("kept %d statements, want %d", got, count)
				}
				for _, bucket := range []string{"public", "other", "shared"} {
					want := bucket == "shared"
					if got := parsed.IsAllowed(Args{Action: GetObjectAction, BucketName: bucket, ObjectName: "file"}); got != want {
						t.Errorf("GetObject %s/file = %v, want %v", bucket, got, want)
					}
				}
			})
		}
	}
}

func TestMergePoliciesPreservesNotResourceDenies(t *testing.T) {
	allow := Statement{Effect: Allow, Actions: NewActionSet(GetObjectAction), Resources: NewResourceSet(NewResource("*"))}
	public := Statement{Effect: Deny, Actions: NewActionSet(GetObjectAction), NotResources: NewResourceSet(NewResource("public/*"), NewResource("shared/*"))}
	other := Statement{Effect: Deny, Actions: NewActionSet(GetObjectAction), NotResources: NewResourceSet(NewResource("other/*"), NewResource("shared/*"))}
	merged := MergePolicies(
		Policy{Version: DefaultVersion, Statements: []Statement{allow, public}},
		Policy{Version: DefaultVersion, Statements: []Statement{other, public.Clone()}},
	)
	if got := len(merged.Statements); got != 3 {
		t.Errorf("kept %d statements, want 3 distinct statements", got)
	}
	for _, bucket := range []string{"public", "other", "shared"} {
		want := bucket == "shared"
		if got := merged.IsAllowed(Args{Action: GetObjectAction, BucketName: bucket, ObjectName: "file"}); got != want {
			t.Errorf("GetObject %s/file = %v, want %v", bucket, got, want)
		}
	}
}

func TestHasDenyStatementWithoutParsing(t *testing.T) {
	for _, tt := range []struct {
		name    string
		effects []Effect
		want    bool
	}{
		{"empty", nil, false},
		{"allow", []Effect{Allow}, false},
		{"deny", []Effect{Deny}, true},
		{"allow-then-deny", []Effect{Allow, Deny}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := Policy{Version: DefaultVersion}
			for _, effect := range tt.effects {
				p.Statements = append(p.Statements, Statement{
					Effect: effect, Actions: NewActionSet(GetObjectAction),
					Resources: NewResourceSet(NewResource("*")),
				})
			}
			if got := p.HasDenyStatement(); got != tt.want {
				t.Errorf("HasDenyStatement() = %v, want %v", got, tt.want)
			}
			p.updateActionIndex()
			if got := p.HasDenyStatement(); got != tt.want {
				t.Errorf("indexed HasDenyStatement() = %v, want %v", got, tt.want)
			}
		})
	}
	for _, p := range DefaultPolicies {
		if p.Name == "readonly" {
			if !p.Definition.HasDenyStatement() {
				t.Error("readonly's explicit Deny must be reported before parsing")
			}
			return
		}
	}
	t.Fatal("readonly policy not found")
}
