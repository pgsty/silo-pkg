// Copyright (c) 2026 PGSTY
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"strings"
	"testing"
)

func TestParseConfigStrictRejectsUnknownFieldsAndNonObjects(t *testing.T) {
	invalidDocs := []struct {
		name string
		doc  string
	}{
		{
			name: "top-level misspelled Statements",
			doc:  `{"Version":"2012-10-17","Statements":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::only-one/*"]}]}`,
		},
		{
			name: "top-level misspelled Statment", //nolint:misspell // Intentional typo tests strict rejection.
			doc:  `{"Statment":[]}`,               //nolint:misspell // Intentional typo tests strict rejection.
		},
		{
			name: "statement misspelled Conditions",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::b/*"],"Conditions":{"IpAddress":{"aws:SourceIp":["10.0.0.1/32"]}}}]}`,
		},
		{
			name: "statement misspelled NotResources",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::*"],"NotResources":["arn:aws:s3:::secret/*"]}]}`,
		},
		{
			name: "null literal",
			doc:  `null`,
		},
		{
			name: "whitespace padded null",
			doc:  " \n\t null \t\n ",
		},
		{
			name: "empty input",
			doc:  "   ",
		},
		{
			name: "top-level array",
			doc:  `[]`,
		},
		{
			name: "trailing JSON object",
			doc:  `{} {}`,
		},
		{
			name: "trailing garbage",
			doc:  `{} trailing`,
		},
	}

	for _, tc := range invalidDocs {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := ParseConfigStrict(strings.NewReader(tc.doc)); err == nil {
				t.Fatalf("ParseConfigStrict(%s) succeeded with %+v, want error", tc.name, got)
			}
		})
	}

	// Explicit empty JSON objects remain valid policy-reset documents.
	for _, validDoc := range []string{
		`{}`,
		`{"Statement":[]}`,
		`{"Version":"2012-10-17","Statement":[]}`,
	} {
		parsed, err := ParseConfigStrict(strings.NewReader(validDoc))
		if err != nil {
			t.Fatalf("ParseConfigStrict(%q) failed: %v", validDoc, err)
		}
		if !parsed.IsEmpty() {
			t.Fatalf("ParseConfigStrict(%q) expected empty policy, got %+v", validDoc, parsed)
		}
	}
}

func TestParseBucketPolicyConfigRejectsUnknownFieldsAndNonObjects(t *testing.T) {
	invalidDocs := []struct {
		name string
		doc  string
	}{
		{
			name: "top-level misspelled Statements",
			doc:  `{"Version":"2012-10-17","Statements":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::mybucket/*"]}]}`,
		},
		{
			name: "statement misspelled Conditions",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::mybucket/*"],"Conditions":{"IpAddress":{"aws:SourceIp":["10.0.0.1/32"]}}}]}`,
		},
		{
			name: "principal unknown field CanonicalUser",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"CanonicalUser":"abc"},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::mybucket/*"]}]}`,
		},
		{
			name: "principal empty object",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::mybucket/*"]}]}`,
		},
		{
			name: "statement unsupported NotPrincipal",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","NotPrincipal":{"AWS":["arn:aws:iam::123456789012:root"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::mybucket/*"]}]}`,
		},
		{
			name: "null literal",
			doc:  `null`,
		},
		{
			name: "trailing JSON object",
			doc:  `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::mybucket/*"]}]} {}`,
		},
	}

	for _, tc := range invalidDocs {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := ParseBucketPolicyConfig(strings.NewReader(tc.doc), "mybucket"); err == nil {
				t.Fatalf("ParseBucketPolicyConfig(%s) succeeded with %+v, want error", tc.name, got)
			}
		})
	}
}
