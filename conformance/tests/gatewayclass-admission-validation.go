/*
Copyright 2022 The Kubernetes Authors.

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

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayClassAdmissionValidation)
}

var GatewayClassAdmissionValidation = suite.ConformanceTest{
	ShortName:   "GatewayClassAdmissionValidation",
	Description: "GatewayClass admission validation behavior",
	Manifests:   []string{"tests/gatewayclass-admission-validation.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("GatewayClass.spec.controllerName is immutable", func(t *testing.T) {
			c := suite.Client
			timeoutConfig := suite.TimeoutConfig
			gwcName := "gatewayclass-immutable"

			ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
			defer cancel()

			gwc := &v1beta1.GatewayClass{}
			err := c.Get(ctx, types.NamespacedName{Name: gwcName}, gwc)

			require.NoErrorf(t, err, "error fetching %s GatewayClass", gwcName)

			gwc.Spec.ControllerName = v1beta1.GatewayController(fmt.Sprintf("%s-modified", gwc.Spec.ControllerName))

			ctx, cancel = context.WithTimeout(context.Background(), timeoutConfig.CreateTimeout)
			defer cancel()

			err = c.Update(ctx, gwc)

			var e *apierrors.StatusError
			require.ErrorAsf(t, err, &e, "updating %s GatewayClass.spec.controllerName should not be permitted", gwcName)
			require.Equal(t, int32(400), e.ErrStatus.Code, "HTTP response code should be 400")
		})
	},
}
