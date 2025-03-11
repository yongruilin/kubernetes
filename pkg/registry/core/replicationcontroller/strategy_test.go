/*
Copyright 2015 The Kubernetes Authors.

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

package replicationcontroller

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	fldtest "k8s.io/apimachinery/pkg/util/validation/field/testing"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	podtest "k8s.io/kubernetes/pkg/api/pod/testing"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/utils/ptr"

	// ensure types are installed
	_ "k8s.io/kubernetes/pkg/apis/core/install"
)

func TestControllerStrategy(t *testing.T) {
	ctx := genericapirequest.NewDefaultContext()
	if !Strategy.NamespaceScoped() {
		t.Errorf("ReplicationController must be namespace scoped")
	}
	if Strategy.AllowCreateOnUpdate() {
		t.Errorf("ReplicationController should not allow create on update")
	}

	validSelector := map[string]string{"a": "b"}
	validPodTemplate := api.PodTemplate{
		Template: api.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validSelector,
			},
			Spec: podtest.MakePodSpec(),
		},
	}
	rc := &api.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
		Spec: api.ReplicationControllerSpec{
			Replicas: ptr.To[int32](1),
			Selector: validSelector,
			Template: &validPodTemplate.Template,
		},
		Status: api.ReplicationControllerStatus{
			Replicas:           1,
			ObservedGeneration: int64(10),
		},
	}

	Strategy.PrepareForCreate(ctx, rc)
	if rc.Status.Replicas != 0 {
		t.Error("ReplicationController should not allow setting status.replicas on create")
	}
	if rc.Status.ObservedGeneration != int64(0) {
		t.Error("ReplicationController should not allow setting status.observedGeneration on create")
	}
	errs := Strategy.Validate(ctx, rc)
	if len(errs) != 0 {
		t.Errorf("Unexpected error validating %v", errs)
	}

	invalidRc := &api.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{Name: "bar", ResourceVersion: "4"},
	}
	Strategy.PrepareForUpdate(ctx, invalidRc, rc)
	errs = Strategy.ValidateUpdate(ctx, invalidRc, rc)
	if len(errs) == 0 {
		t.Errorf("Expected a validation error")
	}
	if invalidRc.ResourceVersion != "4" {
		t.Errorf("Incoming resource version on update should not be mutated")
	}
}

func TestControllerStatusStrategy(t *testing.T) {
	ctx := genericapirequest.NewDefaultContext()
	if !StatusStrategy.NamespaceScoped() {
		t.Errorf("ReplicationController must be namespace scoped")
	}
	if StatusStrategy.AllowCreateOnUpdate() {
		t.Errorf("ReplicationController should not allow create on update")
	}
	validSelector := map[string]string{"a": "b"}
	validPodTemplate := api.PodTemplate{
		Template: api.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validSelector,
			},
			Spec: api.PodSpec{
				RestartPolicy: api.RestartPolicyAlways,
				DNSPolicy:     api.DNSClusterFirst,
				Containers:    []api.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent", TerminationMessagePolicy: "File"}},
			},
		},
	}
	oldController := &api.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "10"},
		Spec: api.ReplicationControllerSpec{
			Replicas: ptr.To[int32](3),
			Selector: validSelector,
			Template: &validPodTemplate.Template,
		},
		Status: api.ReplicationControllerStatus{
			Replicas:           1,
			ObservedGeneration: int64(10),
		},
	}
	newController := &api.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "9"},
		Spec: api.ReplicationControllerSpec{
			Replicas: ptr.To[int32](1),
			Selector: validSelector,
			Template: &validPodTemplate.Template,
		},
		Status: api.ReplicationControllerStatus{
			Replicas:           3,
			ObservedGeneration: int64(11),
		},
	}
	StatusStrategy.PrepareForUpdate(ctx, newController, oldController)
	if newController.Status.Replicas != 3 {
		t.Errorf("Replication controller status updates should allow change of replicas: %v", newController.Status.Replicas)
	}
	if *newController.Spec.Replicas != 3 {
		t.Errorf("PrepareForUpdate should have preferred spec")
	}
	errs := StatusStrategy.ValidateUpdate(ctx, newController, oldController)
	if len(errs) != 0 {
		t.Errorf("Unexpected error %v", errs)
	}
}

func TestSelectableFieldLabelConversions(t *testing.T) {
	apitesting.TestSelectableFieldLabelConversionsOfKind(t,
		"v1",
		"ReplicationController",
		ControllerToSelectableFields(&api.ReplicationController{}),
		nil,
	)
}

func TestValidateUpdate(t *testing.T) {
	ctx := genericapirequest.NewDefaultContext()
	validSelector := map[string]string{"a": "b"}
	validPodTemplate := api.PodTemplate{
		Template: api.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validSelector,
			},
			Spec: podtest.MakePodSpec(),
		},
	}
	oldController := &api.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: api.NamespaceDefault, ResourceVersion: "10", Annotations: make(map[string]string)},
		Spec: api.ReplicationControllerSpec{
			Replicas: ptr.To[int32](3),
			Selector: validSelector,
			Template: &validPodTemplate.Template,
		},
		Status: api.ReplicationControllerStatus{
			Replicas:           1,
			ObservedGeneration: int64(10),
		},
	}
	// Conversion sets this annotation
	oldController.Annotations[api.NonConvertibleAnnotationPrefix+"/"+"spec.selector"] = "no way"

	// Deep-copy so we won't mutate both selectors.
	newController := oldController.DeepCopy()

	// Irrelevant (to the selector) update for the replication controller.
	newController.Spec.Replicas = ptr.To[int32](5)

	// If they didn't try to update the selector then we should not return any error.
	errs := Strategy.ValidateUpdate(ctx, newController, oldController)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	// Update the selector - validation should return an error.
	newController.Spec.Selector["shiny"] = "newlabel"
	newController.Spec.Template.Labels["shiny"] = "newlabel"

	errs = Strategy.ValidateUpdate(ctx, newController, oldController)
	for _, err := range errs {
		t.Logf("%#v\n", err)
	}
	if len(errs) != 1 {
		t.Fatalf("expected a validation error")
	}
	if !strings.Contains(errs[0].Error(), "selector") {
		t.Fatalf("expected error related to the selector")
	}
}

// Helper function for RC tests.
func mkValidReplicationController(tweaks ...func(rc *api.ReplicationController)) api.ReplicationController {
	rc := api.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
		Spec: api.ReplicationControllerSpec{
			Replicas: ptr.To[int32](1),
			Selector: map[string]string{"a": "b"},
			Template: &api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b"},
				},
				Spec: podtest.MakePodSpec(),
			},
		},
	}
	for _, tweak := range tweaks {
		tweak(&rc)
	}
	return rc
}

func TestValidateForDeclarative(t *testing.T) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "",
		APIVersion: "v1",
	})
	successCases := []api.ReplicationController{
		mkValidReplicationController(func(rc *api.ReplicationController) {}),
		mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.Replicas = ptr.To[int32](0) }),
		mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.Replicas = ptr.To[int32](1) }),
		mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.Replicas = ptr.To[int32](100) }),
		mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.MinReadySeconds = 0 }),
		mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.MinReadySeconds = 1 }),
		mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.MinReadySeconds = 100 }),
	}
	for _, tc := range successCases {
		for _, gateVal := range []bool{true, false} {
			// We only need to test both gate enabled and disabled together, because
			// 1) the DeclarativeValidationTakeover won't take effect if DeclarativeValidation is disabled.
			// 2) the validation output, when only DeclarativeValidation is enabled, is the same as when both gates are disabled.
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)
			errs := Strategy.Validate(ctx, &tc)
			if len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		}
	}

	errorCases := map[string]struct {
		input        api.ReplicationController
		expectedErrs field.ErrorList
	}{
		"negative replicas": {
			input: mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.Replicas = ptr.To[int32](-1) }),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("spec.replicas"), nil, "").WithOrigin("minimum"),
			},
		},
		"negative minReadySeconds": {
			input: mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.MinReadySeconds = -1 }),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("spec.minReadySeconds"), nil, "").WithOrigin("minimum"),
			},
		},
		"nil replicas": {
			input: mkValidReplicationController(func(rc *api.ReplicationController) { rc.Spec.Replicas = nil }),
			expectedErrs: field.ErrorList{
				field.Required(field.NewPath("spec.replicas"), ""),
			},
		},
	}
	for k, tc := range errorCases {
		t.Run(k, func(t *testing.T) {
			var declarativeTakeoverErrs field.ErrorList
			var imperativeErrs field.ErrorList
			for _, gateVal := range []bool{true, false} {
				// We only need to test both gate enabled and disabled together, because
				// 1) the DeclarativeValidationTakeover won't take effect if DeclarativeValidation is disabled.
				// 2) the validation output, when only DeclarativeValidation is enabled, is the same as when both gates are disabled.
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)

				errs := Strategy.Validate(ctx, &tc.input)
				if gateVal {
					declarativeTakeoverErrs = errs
				} else {
					imperativeErrs = errs
				}
				// The errOutputMatcher is used to verify the output matches the expected errors in test cases.
				errOutputMatcher := fldtest.ErrorMatcher{}.ByType().ByField().ByOrigin()
				errOutputMatcher.Test(t, tc.expectedErrs, errs)
			}
			// The equivalenceMatcher is used to verify the output errors from hand-written imperative validation
			// are equivalent to the output errors when DeclarativeValidationTakeover is enabled.
			equivalenceMatcher := fldtest.ErrorMatcher{}.ByType().ByField().ByOrigin()
			equivalenceMatcher.Test(t, imperativeErrs, declarativeTakeoverErrs)
		})
	}
}

func TestValidateUpdateForDeclarative(t *testing.T) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "",
		APIVersion: "v1",
	})

	successCases := []struct {
		old    api.ReplicationController
		update api.ReplicationController
	}{{
		old: mkValidReplicationController(func(rc *api.ReplicationController) {}),
		update: mkValidReplicationController(func(rc *api.ReplicationController) {
			rc.Spec.Replicas = ptr.To[int32](0)
		}),
	}, {
		old: mkValidReplicationController(func(rc *api.ReplicationController) {}),
		update: mkValidReplicationController(func(rc *api.ReplicationController) {
			rc.Spec.Replicas = ptr.To[int32](3)
		}),
	}, {
		old: mkValidReplicationController(func(rc *api.ReplicationController) {}),
		update: mkValidReplicationController(func(rc *api.ReplicationController) {
			rc.Spec.MinReadySeconds = 0
		}),
	}, {
		old: mkValidReplicationController(func(rc *api.ReplicationController) {}),
		update: mkValidReplicationController(func(rc *api.ReplicationController) {
			rc.Spec.MinReadySeconds = 3
		}),
	}}
	for _, tc := range successCases {
		tc.old.ObjectMeta.ResourceVersion = "1"
		tc.update.ObjectMeta.ResourceVersion = "1"
		for _, gateVal := range []bool{true, false} {
			// We only need to test both gate enabled and disabled together, because
			// 1) the DeclarativeValidationTakeover won't take effect if DeclarativeValidation is disabled.
			// 2) the validation output, when only DeclarativeValidation is enabled, is the same as when both gates are disabled.
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)
			errs := Strategy.ValidateUpdate(ctx, &tc.update, &tc.old)
			if len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		}
	}

	errorCases := map[string]struct {
		old          api.ReplicationController
		update       api.ReplicationController
		expectedErrs field.ErrorList
	}{
		"negative replicas": {
			old: mkValidReplicationController(func(rc *api.ReplicationController) {}),
			update: mkValidReplicationController(func(rc *api.ReplicationController) {
				rc.Spec.Replicas = ptr.To[int32](-1)
			}),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("spec.replicas"), nil, "").WithOrigin("minimum"),
			},
		},
		"nil replicas": {
			old: mkValidReplicationController(func(rc *api.ReplicationController) {}),
			update: mkValidReplicationController(func(rc *api.ReplicationController) {
				rc.Spec.Replicas = nil
			}),
			expectedErrs: field.ErrorList{
				field.Required(field.NewPath("spec.replicas"), ""),
			},
		},
		"negative minReadySeconds": {
			old: mkValidReplicationController(func(rc *api.ReplicationController) {}),
			update: mkValidReplicationController(func(rc *api.ReplicationController) {
				rc.Spec.MinReadySeconds = -1
			}),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("spec.minReadySeconds"), nil, "").WithOrigin("minimum"),
			},
		},
	}
	for k, tc := range errorCases {
		t.Run(k, func(t *testing.T) {
			tc.old.ObjectMeta.ResourceVersion = "1"
			tc.update.ObjectMeta.ResourceVersion = "1"
			var declarativeTakeoverErrs field.ErrorList
			var imperativeErrs field.ErrorList
			for _, gateVal := range []bool{true, false} {
				// We only need to test both gate enabled and disabled together, because
				// 1) the DeclarativeValidationTakeover won't take effect if DeclarativeValidation is disabled.
				// 2) the validation output, when only DeclarativeValidation is enabled, is the same as when both gates are disabled.
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidation, gateVal)
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.DeclarativeValidationTakeover, gateVal)
				errs := Strategy.ValidateUpdate(ctx, &tc.update, &tc.old)
				if gateVal {
					declarativeTakeoverErrs = errs
				} else {
					imperativeErrs = errs
				}
				// The errOutputMatcher is used to verify the output matches the expected errors in test cases.
				errOutputMatcher := fldtest.ErrorMatcher{}.ByType().ByField().ByOrigin()
				errOutputMatcher.Test(t, tc.expectedErrs, errs)
			}
			// The equivalenceMatcher is used to verify the output errors from hand-written imperative validation
			// are equivalent to the output errors when DeclarativeValidationTakeover is enabled.
			equivalenceMatcher := fldtest.ErrorMatcher{}.ByType().ByField().ByOrigin()
			// TODO: remove this once ErrorMatcher fix the bug.
			dedupedImperativeErrs := field.ErrorList{}
			for _, err := range imperativeErrs {
				found := false
				for _, existingErr := range dedupedImperativeErrs {
					if equivalenceMatcher.Matches(existingErr, err) {
						found = true
						break
					}
				}
				if !found {
					dedupedImperativeErrs = append(dedupedImperativeErrs, err)
				}
			}
			equivalenceMatcher.Test(t, dedupedImperativeErrs, declarativeTakeoverErrs)
		})
	}
}
