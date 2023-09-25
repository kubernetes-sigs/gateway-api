/*
Copyright 2023 The Kubernetes Authors.

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

package utils

import (
	"bytes"
	"context"
	"io"
	"testing"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type CmdParams struct {
	K8sClients    *common.K8sClients
	PolicyManager *policymanager.PolicyManager
	Out           io.Writer
}

func MustParamsForTest(t *testing.T, fakeClients *common.K8sClients) *CmdParams {
	policyManager := policymanager.New(fakeClients.DC)
	if err := policyManager.Init(context.Background()); err != nil {
		t.Fatalf("failed to initialize PolicyManager: %v", err)
	}
	return &CmdParams{
		K8sClients:    fakeClients,
		PolicyManager: policyManager,
		Out:           &bytes.Buffer{},
	}
}
