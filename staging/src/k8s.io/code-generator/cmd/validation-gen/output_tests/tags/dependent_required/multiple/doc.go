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

// +k8s:validation-gen=TypeMeta
// +k8s:validation-gen-scheme-registry=k8s.io/code-generator/cmd/validation-gen/testscheme.Scheme

// This is a test package.
// +k8s:validation-gen-nolint
package multiple

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

// Multiple triggers on the same dependent. Each tag is independent (OR
// semantics): Dependent is required if any of T1 or T2 is set. Also covers
// non-pointer trigger (string) and slice trigger to exercise the predicate.
type Struct struct {
	TypeMeta int

	// +k8s:optional
	T1 *string `json:"t1"`

	// +k8s:optional
	T2 []string `json:"t2"`

	// +k8s:optional
	// +k8s:dependentRequired(T1)
	// +k8s:dependentRequired(T2)
	Dependent *string `json:"dependent"`
}
