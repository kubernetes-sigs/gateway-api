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

package printer

import (
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

type Printer interface {
	io.Writer
	GetPrintableNodes(resourceModel *resourcediscovery.ResourceModel) []NodeResource
	PrintTable(resourceModel *resourcediscovery.ResourceModel)
}

func Print(p Printer, resourceModel *resourcediscovery.ResourceModel, format utils.OutputFormat) {
	switch format {
	case utils.OutputFormatTable:
		p.PrintTable(resourceModel)
	case utils.OutputFormatJSON, utils.OutputFormatYAML:
		nodes := SortByString(p.GetPrintableNodes(resourceModel))
		clientObjects := ClientObjects(nodes)

		printablePayload, err := renderPrintableObject(clientObjects)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to form the printable payload %v\n", err)
			os.Exit(1)
		}
		output, err := utils.MarshalWithFormat(printablePayload, format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal the object %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(p, string(output))
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized output format: %s\n", format)
		os.Exit(1)
	}
}

func renderPrintableObject(objs []client.Object) (runtime.Object, error) {
	list := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       "List",
			"apiVersion": "v1",
		},
	}
	for _, obj := range objs {
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return list, err
		}
		list.Items = append(list.Items, unstructured.Unstructured{Object: unstructuredObj})
	}
	return list, nil
}
