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
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

// Factory encapsulates the common clients and structures which are needed for
// the execution of some command.
type Factory interface {
	// INTERNAL COMMENT:
	// - The reason for an interface here is to be able to inject this dependency
	//   during unit tests.
	// - The reason for the "factory" pattern is to delay the construction of the
	//   objects for when the commands get run (as opposed to when the commands
	//   are registered in Cobra). This is required because during registration of
	//   the commands, we will need to inject this dependency, but we cannot
	//   construct the dependency because it depends on some flag values which are
	//   only known during the runtime of the command.

	K8sClients() (*common.K8sClients, error)
	PolicyManager() (*policymanager.PolicyManager, error)
}

type factoryImpl struct {
	kubeConfigPath *string

	k8sClients    *common.K8sClients
	policyManager *policymanager.PolicyManager
}

func NewFactory(kubeConfigPath *string) Factory {
	return &factoryImpl{kubeConfigPath: kubeConfigPath}
}

func (f *factoryImpl) K8sClients() (*common.K8sClients, error) {
	if f.k8sClients != nil {
		return f.k8sClients, nil
	}

	if f.kubeConfigPath == nil {
		return nil, fmt.Errorf("kubeConfigPath is nil")
	}

	k8sClients, err := common.NewK8sClients(*f.kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clients: %v", err)
	}
	f.k8sClients = k8sClients
	return f.k8sClients, nil
}

func (f *factoryImpl) PolicyManager() (*policymanager.PolicyManager, error) {
	if f.policyManager != nil {
		return f.policyManager, nil
	}
	k8sClients, err := f.K8sClients()
	if err != nil {
		return nil, err
	}
	policyManager := policymanager.New(k8sClients.DC)
	if err := policyManager.Init(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize policy manager: %v", err)
	}
	f.policyManager = policyManager
	return f.policyManager, nil
}

func MustPolicyManagerForTest(t *testing.T, fakeClients *common.K8sClients) *policymanager.PolicyManager {
	policyManager := policymanager.New(fakeClients.DC)
	if err := policyManager.Init(context.Background()); err != nil {
		t.Fatalf("failed to initialize PolicyManager: %v", err)
	}
	return policyManager
}

type OutputFormat string

const (
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatYAML  OutputFormat = "yaml"
	OutputFormatTable OutputFormat = ""
)

func ValidateAndReturnOutputFormat(format string) (OutputFormat, error) {
	switch format {
	case "json":
		return OutputFormatJSON, nil
	case "yaml":
		return OutputFormatYAML, nil
	case "":
		return OutputFormatTable, nil
	default:
		var zero OutputFormat
		return zero, fmt.Errorf("unknown format %s provided", format)
	}
}

func MarshalWithFormat(content any, format OutputFormat) ([]byte, error) {
	if format == OutputFormatJSON {
		return json.MarshalIndent(content, "", "  ")
	}
	if format == OutputFormatYAML {
		return yaml.Marshal(content)
	}
	return []byte{}, fmt.Errorf("format %s not found to support marshaling", format)
}
