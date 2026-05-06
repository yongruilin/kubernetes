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

package simple

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Neither trigger nor dependent set: valid.
	st.Value(&Struct{}).ExpectValid()

	// Only dependent set: valid (the rule is asymmetric).
	st.Value(&Struct{Dependent: ptr.To("d")}).ExpectValid()

	// Both set: valid.
	st.Value(&Struct{Trigger: ptr.To("t"), Dependent: ptr.To("d")}).ExpectValid()

	// Trigger set without dependent: error on dependent's path.
	st.Value(&Struct{Trigger: ptr.To("t")}).ExpectMatches(
		field.ErrorMatcher{}.ByType().ByField().ByOrigin(),
		field.ErrorList{
			field.Required(field.NewPath("dependent"), "").WithOrigin("dependentRequired"),
		},
	)

	// Update ratcheting: same invalid state in old → no error on update.
	st.Value(&Struct{Trigger: ptr.To("t")}).
		OldValue(&Struct{Trigger: ptr.To("t")}).
		ExpectValid()

	// Update transition into invalid state → error.
	st.Value(&Struct{Trigger: ptr.To("t")}).
		OldValue(&Struct{}).
		ExpectMatches(
			field.ErrorMatcher{}.ByType().ByField().ByOrigin(),
			field.ErrorList{
				field.Required(field.NewPath("dependent"), "").WithOrigin("dependentRequired"),
			},
		)

	// Update transition out of invalid state → no error.
	st.Value(&Struct{Trigger: ptr.To("t"), Dependent: ptr.To("d")}).
		OldValue(&Struct{Trigger: ptr.To("t")}).
		ExpectValid()
}
