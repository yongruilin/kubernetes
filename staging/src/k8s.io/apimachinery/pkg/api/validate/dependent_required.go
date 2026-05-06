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

package validate

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// DependentRequired verifies that, if the trigger field on the parent object
// is "set", then the dependent field is also "set". The set predicate is
// supplied by the caller via ExtractorFn; the codegen layer reuses the same
// predicate it uses for union membership: pointer non-nil, slice/map
// non-empty, builtin non-zero.
//
// This is invoked from the parent struct's validator (analogous to Union and
// ZeroOrOneOfUnion). parentPath is the parent struct's path; the error is
// reported on parentPath.Child(dependentName). triggerName is used only for
// the human-readable error message.
//
// On Update, if neither the trigger nor the dependent changed set-state, the
// validation is skipped (ratcheting), so existing invalid objects are not
// blocked from unrelated updates.
func DependentRequired[T any](_ context.Context, op operation.Operation,
	parentPath *field.Path,
	obj, oldObj T,
	dependentName, triggerName string,
	triggerSet, dependentSet ExtractorFn[T, bool],
) field.ErrorList {
	newTriggerSet := triggerSet(obj)
	newDependentSet := dependentSet(obj)

	if op.Type == operation.Update && !reflect.ValueOf(oldObj).IsZero() {
		oldTriggerSet := triggerSet(oldObj)
		oldDependentSet := dependentSet(oldObj)
		if oldTriggerSet == newTriggerSet && oldDependentSet == newDependentSet {
			return nil
		}
	}

	if newTriggerSet && !newDependentSet {
		return field.ErrorList{
			field.Required(parentPath.Child(dependentName),
				fmt.Sprintf("must be specified when `%s` is set", triggerName)),
		}.WithOrigin("dependentRequired")
	}
	return nil
}
