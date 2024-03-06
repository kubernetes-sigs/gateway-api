package printer

import (
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/yaml"
)

type NamespacesPrinter struct {
	Out io.Writer
}

type namespaceDescribeView struct {
	Name                     string                 `json:",omitempty"`
	Labels                   map[string]string      `json:",omitempty"`
	Annotations              map[string]string      `json:",omitempty"`
	Status                   string                 `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef `json:",omitempty"`
}

func (nsp *NamespacesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, namespaceNode := range resourceModel.Namespaces {
		index++

		views := []namespaceDescribeView{
			{
				Name: namespaceNode.Namespace.Name,
			},
			{
				Annotations: namespaceNode.Namespace.Annotations,
				Labels:      namespaceNode.Namespace.Labels,
			},
			{
				Status: string(namespaceNode.Namespace.Status.Phase),
			},
		}

		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(namespaceNode.Policies); len(policyRefs) != 0 {
			views = append(views, namespaceDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		
		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal to yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprint(nsp.Out, string(b))
		}

		if index+1 <= len(resourceModel.Namespaces) {
			fmt.Fprintf(nsp.Out, "\n\n")
		}
	}
}
