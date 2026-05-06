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

package validators

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/validation-gen/util"
	"k8s.io/gengo/v2/codetags"
	"k8s.io/gengo/v2/parser/tags"
	"k8s.io/gengo/v2/types"
)

const dependentRequiredTagName = "k8s:dependentRequired"

var dependentRequiredValidator = types.Name{Package: libValidationPkg, Name: "DependentRequired"}

func init() {
	RegisterTagValidator(dependentRequiredTagValidator{})
}

type dependentRequiredTagValidator struct{}

func (dependentRequiredTagValidator) Init(_ Config) {}

func (dependentRequiredTagValidator) TagName() string {
	return dependentRequiredTagName
}

var dependentRequiredTagValidScopes = sets.New(ScopeField)

func (dependentRequiredTagValidator) ValidScopes() sets.Set[Scope] {
	return dependentRequiredTagValidScopes
}

func (drtv dependentRequiredTagValidator) GetValidations(context Context, tag codetags.Tag) (Validations, error) {
	triggerArg, ok := tag.PositionalArg()
	if !ok {
		return Validations{}, fmt.Errorf("missing required positional argument naming the trigger field")
	}
	triggerGoName := triggerArg.Value
	if triggerGoName == "" {
		return Validations{}, fmt.Errorf("trigger field name must not be empty")
	}

	// Validate dependent has a JSON tag and supported kind up front so the
	// error fires next to the source tag, not inside the deferred callback.
	if context.Member == nil {
		return Validations{}, fmt.Errorf("dependentRequired requires a struct field context")
	}
	if _, kindOK := isSettableKind(context.Member.Type); !kindOK {
		return Validations{}, fmt.Errorf("dependentRequired field must be a pointer, slice, map, or builtin type")
	}
	if _, jsonOK := tags.LookupJSON(*context.Member); !jsonOK {
		return Validations{}, fmt.Errorf("dependentRequired field %q has no JSON struct field tag", context.Member.Name)
	}

	dependentMember := context.Member

	return Validations{
		Deferred: []DeferredGen{
			Deferred(ParentContext, func() (Validations, error) {
				return emitDependentRequired(context.ParentType, dependentMember, triggerGoName)
			}),
		},
	}, nil
}

func emitDependentRequired(parentType *types.Type, dependent *types.Member, triggerGoName string) (Validations, error) {
	parent := util.NonPointer(util.NativeType(parentType))
	if parent.Kind != types.Struct {
		return Validations{}, fmt.Errorf("dependentRequired must be applied to a field of a struct, got %s", parent.Kind)
	}

	var trigger *types.Member
	for i := range parent.Members {
		if parent.Members[i].Name == triggerGoName {
			trigger = &parent.Members[i]
			break
		}
	}
	if trigger == nil {
		return Validations{}, fmt.Errorf("dependentRequired references unknown sibling field %q on %s (only sibling fields are supported)",
			triggerGoName, parent.String())
	}
	if _, kindOK := isSettableKind(trigger.Type); !kindOK {
		return Validations{}, fmt.Errorf("dependentRequired trigger field %q must be a pointer, slice, map, or builtin type", trigger.Name)
	}

	triggerJSON, ok := tags.LookupJSON(*trigger)
	if !ok {
		return Validations{}, fmt.Errorf("dependentRequired trigger field %q has no JSON struct field tag", trigger.Name)
	}
	dependentJSON, _ := tags.LookupJSON(*dependent) // already checked above

	ptrParent := types.PointerTo(parentType)
	triggerExtractor := createMemberExtractor(ptrParent, trigger)
	dependentExtractor := createMemberExtractor(ptrParent, dependent)

	return Validations{
		Functions: []FunctionGen{
			Function(dependentRequiredTagName, DefaultFlags, dependentRequiredValidator,
				dependentJSON.Name,
				triggerJSON.Name,
				triggerExtractor,
				dependentExtractor,
			),
		},
	}, nil
}

// isSettableKind reports whether a type is one for which "set" has a meaningful
// extractor under the union/zeroOrOneOf predicate (pointer non-nil, slice/map
// non-empty, builtin non-zero).
func isSettableKind(t *types.Type) (types.Kind, bool) {
	k := util.NativeType(t).Kind
	switch k {
	case types.Pointer, types.Slice, types.Map, types.Builtin:
		return k, true
	}
	return k, false
}

func (drtv dependentRequiredTagValidator) Docs() TagDoc {
	return TagDoc{
		Tag:            dependentRequiredTagName,
		StabilityLevel: TagStabilityLevelAlpha,
		Scopes:         sets.List(drtv.ValidScopes()),
		Description:    "Indicates that this field must be specified when a sibling trigger field is set.",
		Docs: "Models JSON Schema's dependentRequired: if the named sibling trigger field is set " +
			"(pointer non-nil, slice/map non-empty, or builtin non-zero), the tagged field must also be set. " +
			"Multiple tags on the same field are independent: the field is required if any of the named triggers is set. " +
			"Sibling-only; non-sibling paths are not supported.",
		Args: []TagArgDoc{{
			Description: "<triggerFieldGoName>",
			Type:        codetags.ArgTypeString,
			Required:    true,
		}},
	}
}
