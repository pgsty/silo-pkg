// Copyright (c) 2015-2026 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package policy

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBareARNDetection(t *testing.T) {
	for prefix := range ARNPrefixToType {
		if prefix == ResourceARNAll.String() {
			continue
		}

		t.Run(prefix, func(t *testing.T) {
			exact, err := ParseResource(prefix)
			if err != nil {
				t.Fatalf("parse exact prefix: %v", err)
			}
			legacy, err := ParseResource(ResourceARNAll.String() + prefix)
			if err != nil {
				t.Fatalf("parse historical spelling: %v", err)
			}
			if exact != legacy {
				t.Fatalf("spellings normalized differently: %#v != %#v", exact, legacy)
			}
			if !exact.IsBareARN() || !legacy.IsBareARN() {
				t.Fatalf("bare ARN not detected: exact=%#v legacy=%#v", exact, legacy)
			}
			if !exact.IsValid() || !legacy.IsValid() {
				t.Fatal("existing bare ARN must remain loadable")
			}
		})
	}

	for _, value := range []string{
		"*",
		"**",
		"***",
		"*foo",
		ResourceARNPrefix + "*",
		ResourceARNPrefix + "bucket",
		ResourceARNPrefix + ResourceARNPrefix,
	} {
		t.Run("not-bare/"+value, func(t *testing.T) {
			r, err := ParseResource(value)
			if err != nil {
				t.Fatalf("parse %q: %v", value, err)
			}
			if r.IsBareARN() {
				t.Fatalf("%q incorrectly detected as a bare ARN: %#v", value, r)
			}
		})
	}

	empty := NewResource("")
	if empty.IsBareARN() {
		t.Fatalf("an empty programmatic resource was detected as a bare ARN: %#v", empty)
	}
	if empty.IsValid() {
		t.Fatalf("an empty programmatic resource became valid: %#v", empty)
	}

	const wildcardDoc = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["**"]}]}`
	p, err := ParseConfig(strings.NewReader(wildcardDoc))
	if err != nil {
		t.Fatalf("historical wildcard policy no longer loads: %v", err)
	}
	if !p.IsAllowed(Args{Action: GetObjectAction, BucketName: "bucket", ObjectName: "object"}) {
		t.Fatal("historical wildcard policy no longer matches")
	}
}

func TestBareARNStrictValidationAcceptsExplicitResources(t *testing.T) {
	for _, value := range []string{
		"*",
		"**",
		ResourceARNPrefix + "*",
		ResourceARNPrefix + "bucket/*",
	} {
		t.Run(value, func(t *testing.T) {
			doc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["` + value + `"]}]}`
			if _, err := ParseConfigStrict(strings.NewReader(doc)); err != nil {
				t.Fatalf("strict parsing rejected an explicit resource: %v", err)
			}
		})
	}

	const mixed = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::","arn:aws:s3:::bucket/*"]}]}`
	if _, err := ParseConfigStrict(strings.NewReader(mixed)); err == nil {
		t.Fatal("strict parsing accepted a set containing a bare ARN")
	}
}

func TestBareARNStrictValidation(t *testing.T) {
	testCases := []struct {
		name   string
		action string
		prefix string
	}{
		{name: "S3", action: "s3:GetObject", prefix: ResourceARNPrefix},
		{name: "S3 Tables", action: "s3tables:GetTableData", prefix: ResourceARNS3TablesPrefix},
		{name: "KMS", action: "kms:ListKeys", prefix: ResourceARNKMSPrefix},
		{name: "Vectors", action: "s3vectors:GetVectors", prefix: ResourceARNPrefix},
		{name: "Admin", action: "admin:SetBucketQuota", prefix: ResourceARNPrefix},
		{name: "STS", action: "sts:AssumeRole", prefix: ResourceARNPrefix},
	}

	for _, tc := range testCases {
		for _, field := range []string{"Resource", "NotResource"} {
			for _, value := range []string{tc.prefix, ResourceARNAll.String() + tc.prefix} {
				name := tc.name + "/" + field + "/" + value
				t.Run(name, func(t *testing.T) {
					doc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["` + tc.action + `"],"` + field + `":["` + value + `"]}]}`
					if _, err := ParseConfig(strings.NewReader(doc)); err != nil {
						t.Fatalf("existing policy must remain loadable: %v", err)
					}
					_, err := ParseConfigStrict(strings.NewReader(doc))
					if err == nil {
						t.Fatal("strict parsing accepted a bare ARN")
					}
					if !strings.Contains(err.Error(), tc.prefix+"*") {
						t.Fatalf("error does not name the explicit wildcard: %v", err)
					}
					if strings.Contains(err.Error(), "use '**'") {
						t.Fatalf("error suggests the wildcard type instead of the ARN prefix: %v", err)
					}
				})
			}
		}
	}
}

func TestBareARNPolicyBehaviorIsUnchanged(t *testing.T) {
	testCases := []struct {
		name    string
		doc     string
		action  Action
		allowed bool
	}{
		{
			name: "Allow Resource matches nothing",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::"]}]}`,
		},
		{
			name: "Deny Resource does not fire",
			doc: `{"Version":"2012-10-17","Statement":[
				{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::*"]},
				{"Effect":"Deny","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::"]}]}`,
			allowed: true,
		},
		{
			name:    "Allow NotResource excludes nothing",
			doc:     `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"NotResource":["arn:aws:s3:::"]}]}`,
			allowed: true,
		},
		{
			name: "Deny NotResource denies everything",
			doc: `{"Version":"2012-10-17","Statement":[
				{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::*"]},
				{"Effect":"Deny","Action":["s3:GetObject"],"NotResource":["arn:aws:s3:::"]}]}`,
		},
		{
			name: "Deny NotAction Resource does not fire",
			doc: `{"Version":"2012-10-17","Statement":[
				{"Effect":"Allow","Action":["s3:*"],"Resource":["arn:aws:s3:::*"]},
				{"Effect":"Deny","NotAction":["s3:GetObject"],"Resource":["arn:aws:s3:::"]}]}`,
			action:  PutObjectAction,
			allowed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ParseConfig(strings.NewReader(tc.doc))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			action := tc.action
			if action == "" {
				action = GetObjectAction
			}
			if got := p.IsAllowed(Args{
				AccountName: "probe",
				Action:      action,
				BucketName:  "secret",
				ObjectName:  "data",
			}); got != tc.allowed {
				t.Fatalf("allowed = %v, want %v", got, tc.allowed)
			}
			if err := p.ValidateStrict(); err == nil {
				t.Fatal("strict validation accepted a bare ARN")
			}
		})
	}
}

func TestBareARNResourceMatchExceptionsAreUnchanged(t *testing.T) {
	testCases := []struct {
		name string
		doc  string
		args Args
	}{
		{
			name: "resource-less Admin action",
			doc: `{"Version":"2012-10-17","Statement":[
				{"Effect":"Allow","Action":["admin:ServerInfo"]},
				{"Effect":"Deny","Action":["admin:ServerInfo"],"Resource":["arn:aws:s3:::"]}]}`,
			args: Args{Action: Action(ServerInfoAdminAction)},
		},
		{
			name: "STS action",
			doc: `{"Version":"2012-10-17","Statement":[
				{"Effect":"Allow","Action":["sts:AssumeRole"]},
				{"Effect":"Deny","Action":["sts:AssumeRole"],"Resource":["arn:aws:s3:::"]}]}`,
			args: Args{Action: Action(AssumeRoleAction)},
		},
		{
			name: "KMS first phase",
			doc: `{"Version":"2012-10-17","Statement":[
				{"Effect":"Allow","Action":["kms:ListKeys"],"Resource":["arn:minio:kms:::*"]},
				{"Effect":"Deny","Action":["kms:ListKeys"],"Resource":["arn:minio:kms:::"]}]}`,
			args: Args{Action: Action(KMSListKeysAction)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ParseConfig(strings.NewReader(tc.doc))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if p.IsAllowed(tc.args) {
				t.Fatal("Deny stopped applying when resource matching was bypassed")
			}
			if err := p.ValidateStrict(); err == nil {
				t.Fatal("strict validation accepted a bare ARN")
			}
		})
	}
}

func TestBareARNSerializationAliasRemainsStrict(t *testing.T) {
	const doc = `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::"]}]}`

	p, err := ParseConfig(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"*arn:aws:s3:::"`) {
		t.Fatalf("historical spelling changed: %s", b)
	}
	if _, err := ParseConfigStrict(strings.NewReader(string(b))); err == nil {
		t.Fatal("strict parsing accepted the historical serialization")
	}
}
