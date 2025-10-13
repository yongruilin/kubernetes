package resourceslice

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/utils/ptr"

	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/apis/resource"
)

var apiVersions = []string{"v1beta1", "v1beta2", "v1"}

func TestDeclarativeValidate(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidate(t, apiVersion)
		})
	}
}

func testDeclarativeValidate(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "resource.k8s.io",
		APIVersion: apiVersion,
		Resource:   "resourceslices",
	})
	testCases := map[string]struct {
		input        resource.ResourceSlice
		expectedErrs field.ErrorList
	}{
		"valid": {
			input: mkValidResourceSlice(),
		},
		// TODO: Add more test cases
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			apitesting.VerifyValidationEquivalence(t, ctx, &tc.input, Strategy.Validate, tc.expectedErrs)
		})
	}
}

func TestDeclarativeValidateUpdate(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidateUpdate(t, apiVersion)
		})
	}
}

func testDeclarativeValidateUpdate(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "resource.k8s.io",
		APIVersion: apiVersion,
		Resource:   "resourceslices",
	})
	validObj := mkValidResourceSlice()
	testCases := map[string]struct {
		update       resource.ResourceSlice
		old          resource.ResourceSlice
		expectedErrs field.ErrorList
	}{
		"valid": {
			update: validObj,
			old:    validObj,
		},
		// TODO: Add more test cases
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			tc.old.ResourceVersion = "1"
			tc.update.ResourceVersion = "2"
			apitesting.VerifyUpdateValidationEquivalence(t, ctx, &tc.update, &tc.old, Strategy.ValidateUpdate, tc.expectedErrs)
		})
	}
}

func mkValidResourceSlice() resource.ResourceSlice {
	return resource.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: "valid-resource-slice",
		},
		Spec: resource.ResourceSliceSpec{
			NodeName: ptr.To("valid-node-name"),
			Driver:   "testdriver.example.com",
			Pool: resource.ResourcePool{
				Name:               "valid-pool-name",
				ResourceSliceCount: 1,
			},
			Devices: []resource.Device{{
				Name: "device-0",
			}},
		},
	}
}
