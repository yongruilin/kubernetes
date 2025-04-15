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

// +k8s:validation-gen=TypeMeta
// +k8s:validation-gen-scheme-registry=k8s.io/code-generator/cmd/validation-gen/testscheme.Scheme

// This is a test package.
package ratcheting

import "k8s.io/code-generator/cmd/validation-gen/testscheme"

var localSchemeBuilder = testscheme.New()

type RootStruct struct {
	TypeMeta int

	MyStruct Struct `json:"myStruct"`
}

// +k8s:immutable
type Struct struct {

	// +k8s:listType=map
	// +k8s:listMapKey=key1Field
	// +k8s:eachVal=+k8s:immutable
	ListField []OtherStruct `json:"listField"`

	// +k8s:listType=map
	// +k8s:listMapKey=key1Field
	// +k8s:eachVal=+k8s:immutable
	// +k8s:ratcheting=disabled
	ListTypedefField []OtherTypedefStruct `json:"listTypedefField"`

	// +k8s:minimum=1
	MinField int `json:"minField"`
}

type OtherStruct struct {
	Key1Field string `json:"key1Field"`
	DataField string `json:"dataField"`
}

type OtherTypedefStruct OtherStruct
