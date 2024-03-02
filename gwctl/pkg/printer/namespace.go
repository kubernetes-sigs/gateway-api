package printer

import (
	"fmt"
	"io"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

type NamespacesPrinter struct {
	Out io.Writer
}

type namespaceDescribeView struct {
	Name                     string                 `json:",omitempty"`
	Namespace                string                 `json:",omitempty"`
	Labels                   map[string]string      `json:",omitempty"`
	Annotations              map[string]string      `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef `json:",omitempty"`
}

func (nsp *NamespacesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	fmt.Print("This is the resource model recieved in namespace.go: ", resourceModel, "\n")
}
