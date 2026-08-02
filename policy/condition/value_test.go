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

package condition

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestValueGetBool(t *testing.T) {
	testCases := []struct {
		value          Value
		expectedResult bool
		expectErr      bool
	}{
		{NewBoolValue(true), true, false},
		{NewIntValue(7), false, true},
		{Value{}, false, true},
	}

	for i, testCase := range testCases {
		result, err := testCase.value.GetBool()
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("case %v: error: expected: %v, got: %v\n", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if result != testCase.expectedResult {
				t.Fatalf("case %v: result: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
			}
		}
	}
}

func TestValueGetInt(t *testing.T) {
	testCases := []struct {
		value          Value
		expectedResult int
		expectErr      bool
	}{
		{NewIntValue(7), 7, false},
		{NewBoolValue(true), 0, true},
		{Value{}, 0, true},
	}

	for i, testCase := range testCases {
		result, err := testCase.value.GetInt()
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("case %v: error: expected: %v, got: %v\n", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if result != testCase.expectedResult {
				t.Fatalf("case %v: result: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
			}
		}
	}
}

func TestValueGetString(t *testing.T) {
	testCases := []struct {
		value          Value
		expectedResult string
		expectErr      bool
	}{
		{NewStringValue("foo"), "foo", false},
		{NewBoolValue(true), "", true},
		{Value{}, "", true},
	}

	for i, testCase := range testCases {
		result, err := testCase.value.GetString()
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("case %v: error: expected: %v, got: %v\n", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if result != testCase.expectedResult {
				t.Fatalf("case %v: result: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
			}
		}
	}
}

func TestValueGetType(t *testing.T) {
	testCases := []struct {
		value          Value
		expectedResult reflect.Kind
	}{
		{NewBoolValue(true), reflect.Bool},
		{NewIntValue(7), reflect.Int},
		{NewStringValue("foo"), reflect.String},
		{Value{}, reflect.Invalid},
	}

	for i, testCase := range testCases {
		result := testCase.value.GetType()

		if result != testCase.expectedResult {
			t.Fatalf("case %v: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
		}
	}
}

func TestValueMarshalJSON(t *testing.T) {
	testCases := []struct {
		value          Value
		expectedResult []byte
		expectErr      bool
	}{
		{NewBoolValue(true), []byte("true"), false},
		{NewIntValue(7), []byte("7"), false},
		{NewStringValue("foo"), []byte(`"foo"`), false},
		{Value{}, nil, true},
	}

	for i, testCase := range testCases {
		result, err := json.Marshal(testCase.value)
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("case %v: error: expected: %v, got: %v\n", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if !reflect.DeepEqual(result, testCase.expectedResult) {
				t.Fatalf("case %v: result: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
			}
		}
	}
}

func TestValueStoreBool(t *testing.T) {
	testCases := []struct {
		value          bool
		expectedResult Value
	}{
		{false, NewBoolValue(false)},
		{true, NewBoolValue(true)},
	}

	for i, testCase := range testCases {
		var result Value
		result.StoreBool(testCase.value)

		if !reflect.DeepEqual(result, testCase.expectedResult) {
			t.Fatalf("case %v: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
		}
	}
}

func TestValueStoreInt(t *testing.T) {
	testCases := []struct {
		value          int
		expectedResult Value
	}{
		{0, NewIntValue(0)},
		{7, NewIntValue(7)},
	}

	for i, testCase := range testCases {
		var result Value
		result.StoreInt(testCase.value)

		if !reflect.DeepEqual(result, testCase.expectedResult) {
			t.Fatalf("case %v: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
		}
	}
}

func TestValueStoreString(t *testing.T) {
	testCases := []struct {
		value          string
		expectedResult Value
	}{
		{"", NewStringValue("")},
		{"foo", NewStringValue("foo")},
	}

	for i, testCase := range testCases {
		var result Value
		result.StoreString(testCase.value)

		if !reflect.DeepEqual(result, testCase.expectedResult) {
			t.Fatalf("case %v: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
		}
	}
}

func TestValueString(t *testing.T) {
	testCases := []struct {
		value          Value
		expectedResult string
	}{
		{NewBoolValue(true), "true"},
		{NewIntValue(7), "7"},
		{NewStringValue("foo"), "foo"},
		{Value{}, ""},
	}

	for i, testCase := range testCases {
		result := testCase.value.String()

		if result != testCase.expectedResult {
			t.Fatalf("case %v: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
		}
	}
}

func TestValueUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		data           []byte
		expectedResult Value
		expectErr      bool
	}{
		{[]byte("true"), NewBoolValue(true), false},
		{[]byte("7"), NewIntValue(7), false},
		{[]byte(`"foo"`), NewStringValue("foo"), false},
		{[]byte("True"), Value{}, true},
		{[]byte("7.1"), Value{}, true},
		{[]byte(`["foo"]`), Value{}, true},
	}

	for i, testCase := range testCases {
		var result Value
		err := json.Unmarshal(testCase.data, &result)
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("case %v: error: expected: %v, got: %v\n", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if !reflect.DeepEqual(result, testCase.expectedResult) {
				t.Fatalf("case %v: result: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
			}
		}
	}
}

// TestGetValuesByKeyPrefersExactName pins the lookup order of
// getValuesByKey(). The map it reads holds two kinds of entries: values the
// server computed itself, keyed exactly as the condition key names them
// (SourceIp, SecureTransport, username, ...), and the request's own headers,
// keyed in canonical MIME form. Looking up the canonical form first lets a
// client shadow a server-computed value by sending a header whose canonical
// spelling collides with it -- `Sourceip: 10.0.0.1` overriding the real
// SourceIp, for instance. The exact name must therefore win, with the
// canonical form kept only as a fallback for keys that genuinely name a
// header, such as s3:x-amz-acl.
func TestGetValuesByKeyPrefersExactName(t *testing.T) {
	testCases := []struct {
		name           string
		key            KeyName
		values         map[string][]string
		expectedResult []string
	}{
		{
			name:           "exact name only",
			key:            AWSSourceIP,
			values:         map[string][]string{"SourceIp": {"192.168.1.1"}},
			expectedResult: []string{"192.168.1.1"},
		},
		{
			name:           "canonical form only, for keys that name a header",
			key:            S3XAmzServerSideEncryption,
			values:         map[string][]string{"X-Amz-Server-Side-Encryption": {"AES256"}},
			expectedResult: []string{"AES256"},
		},
		{
			name: "both present, the exact name wins",
			key:  AWSSourceIP,
			values: map[string][]string{
				"SourceIp": {"192.168.1.1"}, // computed by the server
				"Sourceip": {"10.0.0.1"},    // canonicalised client header
			},
			expectedResult: []string{"192.168.1.1"},
		},
		{
			name: "shadowing an identity key",
			key:  AWSUsername,
			values: map[string][]string{
				"username": {"lowpriv"},
				"Username": {"admin"},
			},
			expectedResult: []string{"lowpriv"},
		},
		{
			name: "shadowing the transport key",
			key:  AWSSecureTransport,
			values: map[string][]string{
				"SecureTransport": {"false"},
				"Securetransport": {"true"},
			},
			expectedResult: []string{"false"},
		},
		{
			name: "shadowing a time key",
			key:  AWSEpochTime,
			values: map[string][]string{
				"EpochTime": {"1700000000"},
				"Epochtime": {"1"},
			},
			expectedResult: []string{"1700000000"},
		},
		{
			name:           "absent",
			key:            AWSSourceIP,
			values:         map[string][]string{"UserAgent": {"curl"}},
			expectedResult: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := getValuesByKey(testCase.values, testCase.key.ToKey())
			if !reflect.DeepEqual(result, testCase.expectedResult) {
				t.Errorf("key %v: expected %v, got %v", testCase.key, testCase.expectedResult, result)
			}
		})
	}
}

// TestGetValuesByKeyResolvesTheAppliedValue pins the case that makes the lookup
// order a correctness question and not only a shadowing one.
//
// MinIO's retention path builds the condition map and then overrides the value
// under the policy key's own spelling with what it is actually about to apply,
// while the request's own X-Amz-Object-Lock-Mode header sits alongside it under
// the canonical spelling. Resolving the canonical form first returns the
// header - so a condition written to constrain the retention being set instead
// reads a value the caller chose, and the two can disagree.
func TestGetValuesByKeyResolvesTheAppliedValue(t *testing.T) {
	values := map[string][]string{
		"Object-Lock-Mode": {"COMPLIANCE"}, // from the request header
		"object-lock-mode": {"GOVERNANCE"}, // what the server is applying
	}
	got := getValuesByKey(values, S3ObjectLockMode.ToKey())
	if !reflect.DeepEqual(got, []string{"GOVERNANCE"}) {
		t.Errorf("expected the applied value [GOVERNANCE], got %v", got)
	}
}
