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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/operation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestDependentRequired(t *testing.T) {
	cases := []struct {
		name       string
		trigger    bool
		dependent  bool
		wantErrors bool
	}{
		{name: "neither set", trigger: false, dependent: false, wantErrors: false},
		{name: "trigger only", trigger: true, dependent: false, wantErrors: true},
		{name: "dependent only", trigger: false, dependent: true, wantErrors: false},
		{name: "both set", trigger: true, dependent: true, wantErrors: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			triggerSet := func(_ *testStruct) bool { return tc.trigger }
			dependentSet := func(_ *testStruct) bool { return tc.dependent }

			got := DependentRequired(context.Background(), operation.Operation{},
				field.NewPath("parent"),
				&testStruct{}, (*testStruct)(nil),
				"dependent", "trigger",
				triggerSet, dependentSet)

			if tc.wantErrors && len(got) == 0 {
				t.Fatalf("expected error, got none")
			}
			if !tc.wantErrors && len(got) != 0 {
				t.Fatalf("expected no errors, got %v", got)
			}
			if tc.wantErrors {
				if len(got) != 1 {
					t.Fatalf("expected exactly 1 error, got %d: %v", len(got), got)
				}
				if got[0].Type != field.ErrorTypeRequired {
					t.Errorf("expected ErrorTypeRequired, got %v", got[0].Type)
				}
				if got[0].Origin != "dependentRequired" {
					t.Errorf("expected origin dependentRequired, got %q", got[0].Origin)
				}
			}
		})
	}
}

func TestDependentRequiredRatcheting(t *testing.T) {
	cases := []struct {
		name           string
		oldTrigger     bool
		oldDependent   bool
		newTrigger     bool
		newDependent   bool
		oldStructIsNil bool
		wantErrors     bool
	}{
		{
			name:         "unchanged invalid state - allowed",
			oldTrigger:   true,
			oldDependent: false,
			newTrigger:   true,
			newDependent: false,
			wantErrors:   false,
		},
		{
			name:         "unchanged valid state - allowed",
			oldTrigger:   true,
			oldDependent: true,
			newTrigger:   true,
			newDependent: true,
			wantErrors:   false,
		},
		{
			name:         "transition into invalid state - error",
			oldTrigger:   false,
			oldDependent: false,
			newTrigger:   true,
			newDependent: false,
			wantErrors:   true,
		},
		{
			name:         "transition out of invalid state - allowed",
			oldTrigger:   true,
			oldDependent: false,
			newTrigger:   true,
			newDependent: true,
			wantErrors:   false,
		},
		{
			name:           "nil old object, valid new - allowed",
			newTrigger:     false,
			newDependent:   false,
			oldStructIsNil: true,
			wantErrors:     false,
		},
		{
			name:           "nil old object, invalid new - error",
			newTrigger:     true,
			newDependent:   false,
			oldStructIsNil: true,
			wantErrors:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			triggerSet := func(s *testStruct) bool {
				if s == nil {
					return false
				}
				if reflect.ValueOf(s).Pointer() == reflect.ValueOf(newObj).Pointer() {
					return tc.newTrigger
				}
				return tc.oldTrigger
			}
			dependentSet := func(s *testStruct) bool {
				if s == nil {
					return false
				}
				if reflect.ValueOf(s).Pointer() == reflect.ValueOf(newObj).Pointer() {
					return tc.newDependent
				}
				return tc.oldDependent
			}

			var oldObj *testStruct = oldObjVal
			if tc.oldStructIsNil {
				oldObj = nil
			}

			got := DependentRequired(context.Background(), operation.Operation{Type: operation.Update},
				field.NewPath("parent"),
				newObj, oldObj,
				"dependent", "trigger",
				triggerSet, dependentSet)

			if tc.wantErrors && len(got) == 0 {
				t.Errorf("expected error, got none")
			}
			if !tc.wantErrors && len(got) != 0 {
				t.Errorf("expected no errors, got %v", got)
			}
		})
	}
}

// Sentinel old/new structs so the extractors can tell which one is being asked
// about. Pointer identity is the discriminator.
var (
	newObj    = &testStruct{}
	oldObjVal = &testStruct{}
)
