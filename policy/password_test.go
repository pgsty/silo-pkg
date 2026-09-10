// Copyright (c) 2026 Feng Ruohang
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"strings"
	"testing"
)

func TestPasswordAndUserManagementActions(t *testing.T) {
	for _, tt := range []struct {
		name       string
		statements string
		password   bool
		createUser bool
	}{
		{"implicit self service", `{"Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::*"}`, true, false},
		{"explicit password grant", `{"Effect":"Allow","Action":"admin:ChangeMyPassword"}`, true, false},
		{"explicit user management grant", `{"Effect":"Allow","Action":"admin:CreateUser"}`, true, true},
		{"legacy user management deny", `{"Effect":"Deny","Action":"admin:CreateUser","Resource":"arn:aws:s3:::*"}`, true, false},
		{"password deny", `{"Effect":"Deny","Action":"admin:ChangeMyPassword"}`, false, false},
		{"password deny with user management grant", `{"Effect":"Deny","Action":"admin:ChangeMyPassword"},{"Effect":"Allow","Action":"admin:CreateUser"}`, false, true},
		{"deny overrides allow", `{"Effect":"Allow","Action":"admin:ChangeMyPassword"},{"Effect":"Deny","Action":"admin:ChangeMyPassword"}`, false, false},
		{"wildcard deny", `{"Effect":"Deny","Action":"admin:*"}`, false, false},
		{"both denies", `{"Effect":"Deny","Action":["admin:CreateUser","admin:ChangeMyPassword"]}`, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := `{"Version":"2012-10-17","Statement":[` + tt.statements + `]}`
			for _, mode := range []string{"read", "write", "merge"} {
				t.Run(mode, func(t *testing.T) {
					parser := ParseConfig
					if mode == "write" {
						parser = ParseConfigStrict
					}
					p, err := parser(strings.NewReader(doc))
					if err != nil {
						t.Fatal(err)
					}
					if mode == "merge" {
						merged := MergePolicies(*p, Policy{})
						p = &merged
					}
					actions := p.IsAllowedActions("", "", nil)
					for action, want := range map[Action]bool{
						ChangeMyPasswordAdminAction: tt.password,
						CreateUserAdminAction:       tt.createUser,
					} {
						if got := actions.Contains(action); got != want {
							t.Errorf("reported %s = %v, want %v", action, got, want)
						}
						if got := p.IsAllowed(Args{Action: action, DenyOnly: action == ChangeMyPasswordAdminAction}); got != want {
							t.Errorf("authorized %s = %v, want %v", action, got, want)
						}
					}
				})
			}
		})
	}
}

func TestReadOnlySelfServicePolicies(t *testing.T) {
	for _, name := range []string{"readonly", "consolereadonly"} {
		t.Run(name, func(t *testing.T) {
			var p *Policy
			for _, canned := range DefaultPolicies {
				if canned.Name == name {
					p = &canned.Definition
					break
				}
			}
			if p == nil {
				t.Fatal("missing canned policy")
			}
			actions := p.IsAllowedActions("bucket", "object", nil)
			for action, want := range map[Action]bool{
				GetObjectAction:                 true,
				GetBucketLocationAction:         true,
				ListBucketAction:                name == "consolereadonly",
				PutObjectAction:                 false,
				DeleteObjectAction:              false,
				CreateUserAdminAction:           false,
				ChangeMyPasswordAdminAction:     true,
				CreateServiceAccountAdminAction: true,
			} {
				if got := actions.Contains(action); got != want {
					t.Errorf("%s = %v, want %v", action, got, want)
				}
			}
			// A read-only grant must compose with a separate user-admin policy.
			userAdmin := Policy{Version: DefaultVersion, Statements: []Statement{{
				Effect: Allow, Actions: NewActionSet(CreateUserAdminAction),
			}}}
			merged := MergePolicies(*p, userAdmin)
			if !merged.IsAllowed(Args{Action: CreateUserAdminAction}) {
				t.Error("read-only policy overrides an independent user-management grant")
			}
		})
	}
}
