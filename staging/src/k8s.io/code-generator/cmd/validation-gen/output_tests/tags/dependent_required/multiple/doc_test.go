/*
Copyright 2026 The Kubernetes Authors.

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

package multiple

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// No triggers set: valid even without dependent.
	st.Value(&Struct{}).ExpectValid()

	// Dependent set with no triggers: valid.
	st.Value(&Struct{Dependent: ptr.To("d")}).ExpectValid()

	// One trigger set, dependent set: valid (both T1 and T2 cases).
	st.Value(&Struct{T1: ptr.To("t1"), Dependent: ptr.To("d")}).ExpectValid()
	st.Value(&Struct{T2: []string{"t2"}, Dependent: ptr.To("d")}).ExpectValid()

	// Both triggers set, dependent set: valid.
	st.Value(&Struct{T1: ptr.To("t1"), T2: []string{"t2"}, Dependent: ptr.To("d")}).ExpectValid()

	// Only T1 set, dependent missing: error from T1's dependentRequired.
	st.Value(&Struct{T1: ptr.To("t1")}).ExpectMatches(
		field.ErrorMatcher{}.ByType().ByField().ByOrigin(),
		field.ErrorList{
			field.Required(field.NewPath("dependent"), "").WithOrigin("dependentRequired"),
		},
	)

	// Only T2 set, dependent missing: error from T2's dependentRequired.
	st.Value(&Struct{T2: []string{"t2"}}).ExpectMatches(
		field.ErrorMatcher{}.ByType().ByField().ByOrigin(),
		field.ErrorList{
			field.Required(field.NewPath("dependent"), "").WithOrigin("dependentRequired"),
		},
	)

	// Both triggers set, dependent missing: each independent rule fires
	// (OR semantics). The origin-aware matcher accepts one or more errors
	// matching the same pattern; we additionally assert by detail substring
	// to confirm both trigger names appear.
	st.Value(&Struct{T1: ptr.To("t1"), T2: []string{"t2"}}).ExpectMatches(
		field.ErrorMatcher{}.ByType().ByField().ByOrigin(),
		field.ErrorList{
			field.Required(field.NewPath("dependent"), "").WithOrigin("dependentRequired"),
		},
	)
	st.Value(&Struct{T1: ptr.To("t1"), T2: []string{"t2"}}).ExpectMatches(
		field.ErrorMatcher{}.ByType().ByField().ByDetailSubstring(),
		field.ErrorList{
			field.Required(field.NewPath("dependent"), "must be specified when `t1` is set"),
			field.Required(field.NewPath("dependent"), "must be specified when `t2` is set"),
		},
	)
}
