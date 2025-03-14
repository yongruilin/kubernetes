/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validate

import (
	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	fldtest "k8s.io/apimachinery/pkg/util/validation/field/testing"
)

func TestDNS1123Label(t *testing.T) {
	ctx := context.Background()
	fldPath := field.NewPath("test")

	testCases := []struct {
		name     string
		label    string
		wantErrs field.ErrorList
	}{
		{
			name:     "valid label",
			label:    "valid-label",
			wantErrs: nil,
		},
		{
			name:     "valid single character label",
			label:    "a",
			wantErrs: nil,
		},
		{
			name:     "valid label with numbers",
			label:    "123-abc",
			wantErrs: nil,
		},
		{
			name:  "invalid: uppercase characters",
			label: "Invalid-Label",
			wantErrs: field.ErrorList{
				field.Invalid(fldPath, nil, "").WithOrigin("format=dns-label"),
			},
		},
		{
			name:  "invalid: starts with dash",
			label: "-invalid-label",
			wantErrs: field.ErrorList{
				field.Invalid(fldPath, nil, "").WithOrigin("format=dns-label"),
			},
		},
		{
			name:  "invalid: ends with dash",
			label: "invalid-label-",
			wantErrs: field.ErrorList{
				field.Invalid(fldPath, nil, "").WithOrigin("format=dns-label"),
			},
		},
		{
			name:  "invalid: contains dots",
			label: "invalid.label",
			wantErrs: field.ErrorList{
				field.Invalid(fldPath, nil, "").WithOrigin("format=dns-label"),
			},
		},
		{
			name:  "invalid: contains special characters",
			label: "invalid@label",
			wantErrs: field.ErrorList{
				field.Invalid(fldPath, nil, "").WithOrigin("format=dns-label"),
			},
		},
		{
			name:  "invalid: too long",
			label: "a" + strings.Repeat("b", 62) + "c", // 64 characters
			wantErrs: field.ErrorList{
				field.Invalid(fldPath, nil, "").WithOrigin("format=dns-label"),
			},
		},
		{
			name:     "valid: max length",
			label:    "a" + strings.Repeat("b", 61) + "c", // 63 characters
			wantErrs: nil,
		},
		{
			name:  "invalid: empty string",
			label: "",
			wantErrs: field.ErrorList{
				field.Invalid(fldPath, nil, "").WithOrigin("format=dns-label"),
			},
		},
	}

	matcher := fldtest.ErrorMatcher{}.ByType().ByField().ByOrigin()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.label
			gotErrs := DNSLabel(ctx, operation.Operation{}, fldPath, &value, nil)

			matcher.Test(t, tc.wantErrs, gotErrs)
		})
	}
}
