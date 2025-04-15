package ratcheting

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/validate/content"
	field "k8s.io/apimachinery/pkg/util/validation/field"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&RootStruct{MyStruct: Struct{MinField: 0}}).ExpectInvalid(field.Invalid(field.NewPath("myStruct.minField"), 0, content.MinError(1)))
	st.Value(&RootStruct{MyStruct: Struct{MinField: 0}}).OldValue(&RootStruct{MyStruct: Struct{MinField: 0}}).ExpectValid()
}
