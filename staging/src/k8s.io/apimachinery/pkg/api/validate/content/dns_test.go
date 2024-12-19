/*
Copyright 2014 The Kubernetes Authors.

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

package content

import (
	"strings"
	"testing"
)

func TestIsDNS1123Label(t *testing.T) {
	// errors expected
	const minSizeError = "must contain at least 1 character"
	const maxSizeError = "must be no more than 63 characters"
	const startEndError = "must start and end with lower-case alphanumeric characters"
	const interiorError = "must contain only lower-case alphanumeric characters or '-'"

	cases := []struct {
		input  string
		expect []string // regexes
	}{
		// Good values
		{"a", nil},
		{"ab", nil},
		{"abc", nil},
		{"a1", nil},
		{"a-1", nil},
		{"a--1--2--b", nil},
		{"0", nil},
		{"01", nil},
		{"012", nil},
		{"1a", nil},
		{"1-a", nil},
		{"1--a--b--2", nil},
		{strings.Repeat("a", 63), nil},

		// Bad values
		{"", mkMsgs(minSizeError)},
		{"A", mkMsgs(startEndError)},
		{"ABC", mkMsgs(startEndError, interiorError)},
		{"aBc", mkMsgs(interiorError)},
		{"AbC", mkMsgs(startEndError)},
		{"A1", mkMsgs(startEndError)},
		{"A-1", mkMsgs(startEndError)},
		{"1-A", mkMsgs(startEndError)},
		{"-", mkMsgs(startEndError)},
		{"a-", mkMsgs(startEndError)},
		{"-a", mkMsgs(startEndError)},
		{"1-", mkMsgs(startEndError)},
		{"-1", mkMsgs(startEndError)},
		{"_", mkMsgs(startEndError)},
		{"a_", mkMsgs(startEndError)},
		{"_a", mkMsgs(startEndError)},
		{"a_b", mkMsgs(interiorError)},
		{"1_", mkMsgs(startEndError)},
		{"_1", mkMsgs(startEndError)},
		{"1_2", mkMsgs(interiorError)},
		{".", mkMsgs(startEndError)},
		{"a.", mkMsgs(startEndError)},
		{".a", mkMsgs(startEndError)},
		{"a.b", mkMsgs(interiorError)},
		{"1.", mkMsgs(startEndError)},
		{".1", mkMsgs(startEndError)},
		{"1.2", mkMsgs(interiorError)},
		{" ", mkMsgs(startEndError)},
		{"a ", mkMsgs(startEndError)},
		{" a", mkMsgs(startEndError)},
		{"a b", mkMsgs(interiorError)},
		{"1 ", mkMsgs(startEndError)},
		{" 1", mkMsgs(startEndError)},
		{"1 2", mkMsgs(interiorError)},
		{strings.Repeat("a", 64), mkMsgs(maxSizeError)},
		{strings.Repeat("-", 64), mkMsgs(maxSizeError)},
		{strings.Repeat(".", 64), mkMsgs(maxSizeError)},
		{strings.Repeat("aBc", 64), mkMsgs(maxSizeError)},
		{strings.Repeat("AbC", 64), mkMsgs(maxSizeError)},
	}

	for i, tc := range cases {
		result := IsDNS1123Label(tc.input)
		testVerify(t, i, tc.input, tc.expect, result)
	}
}
