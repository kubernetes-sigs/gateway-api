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

	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/utils/clock"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

type OutputFormat string

const (
	OutputFormatWide  OutputFormat = "wide"
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatYAML  OutputFormat = "yaml"
	OutputFormatGraph OutputFormat = "graph"
	OutputFormatTable OutputFormat = ""
)

func ValidateAndReturnOutputFormat(format string) (OutputFormat, error) {
	switch format {
	case "wide":
		return OutputFormatWide, nil
	case "json":
		return OutputFormatJSON, nil
	case "yaml":
		return OutputFormatYAML, nil
	case "graph":
		return OutputFormatGraph, nil
	case "":
		return OutputFormatTable, nil
	default:
		var zero OutputFormat
		return zero, fmt.Errorf("unknown format %s provided", format)
	}
}

func AllowedOutputFormatsForHelp() []string {
	return []string{string(OutputFormatWide), string(OutputFormatJSON), string(OutputFormatYAML), string(OutputFormatGraph)}
}

type PrinterOptions struct { //nolint:revive
	OutputFormat OutputFormat
	Description  bool
	Clock        clock.Clock
	EventFetcher eventFetcher
}

type Printer interface {
	PrintNode(node *topology.Node, w io.Writer) error
	Flush(io.Writer) error
}

func NewPrinter(options PrinterOptions) Printer {
	switch {
	case options.OutputFormat == OutputFormatJSON:
		return NewJSONPrinter()
	case options.OutputFormat == OutputFormatYAML:
		return NewYAMLPrinter()
	case options.Description:
		return &DescriptionPrinter{PrinterOptions: options}
	default:
		return &TablePrinter{PrinterOptions: options}
	}
}

type TablePrinter struct {
	PrinterOptions

	table               *Table
	unknownTablePrinter printers.ResourcePrinter
	curType             string
}

func (p *TablePrinter) PrintNode(node *topology.Node, w io.Writer) error {
	return parseAndPrint(node, w, p)
}

func (p *TablePrinter) Flush(w io.Writer) error {
	return p.checkTypeChange("", w)
}

func (p *TablePrinter) printUnknown(node *topology.Node, w io.Writer) error {
	if p.unknownTablePrinter == nil {
		p.unknownTablePrinter = printers.NewTablePrinter(printers.PrintOptions{})
	}
	return p.unknownTablePrinter.PrintObj(node.Object, w)
}

func (p *TablePrinter) checkTypeChange(newType string, w io.Writer) error {
	var err error
	if p.curType != "" && p.curType != newType && p.table != nil {
		err = p.table.Write(w, 0)
		p.table = nil
	}
	p.curType = newType
	return err
}

type DescriptionPrinter struct {
	PrinterOptions

	printSeparator bool
}

func (p *DescriptionPrinter) PrintNode(node *topology.Node, w io.Writer) error {
	return parseAndPrint(node, w, p)
}

type typedPrinter interface {
	printBackend(*topology.Node, io.Writer) error
	printGatewayClass(*topology.Node, io.Writer) error
	printGateway(*topology.Node, io.Writer) error
	printHTTPRoute(*topology.Node, io.Writer) error
	printNamespace(*topology.Node, io.Writer) error
	printPolicy(*topology.Node, io.Writer) error
	printPolicyCRD(*topology.Node, io.Writer) error
	printUnknown(*topology.Node, io.Writer) error
}

func parseAndPrint(node *topology.Node, w io.Writer, p typedPrinter) error {
	if node.Metadata != nil && node.Metadata[common.PolicyGK.String()] != nil {
		return p.printPolicy(node, w)
	}
	if node.Metadata != nil && node.Metadata[common.PolicyCRDGK.String()] != nil {
		return p.printPolicyCRD(node, w)
	}

	switch node.GKNN().GroupKind() {
	case common.GatewayGK:
		return p.printGateway(node, w)
	case common.GatewayClassGK:
		return p.printGatewayClass(node, w)
	case common.HTTPRouteGK:
		return p.printHTTPRoute(node, w)
	case common.NamespaceGK:
		return p.printNamespace(node, w)
	case common.ServiceGK:
		return p.printBackend(node, w)
	default:
		return p.printUnknown(node, w)
	}
}

func (p *DescriptionPrinter) Flush(io.Writer) error { return nil }

func (p *DescriptionPrinter) printUnknown(node *topology.Node, w io.Writer) error {
	printer := &printers.YAMLPrinter{}
	return printer.PrintObj(node.Object, w)
}

type JSONPrinter struct {
	Delegate *printers.OmitManagedFieldsPrinter
}

func NewJSONPrinter() *JSONPrinter {
	return &JSONPrinter{
		Delegate: &printers.OmitManagedFieldsPrinter{
			Delegate: &printers.JSONPrinter{},
		},
	}
}

func (p *JSONPrinter) PrintNode(node *topology.Node, w io.Writer) error {
	return p.Delegate.PrintObj(node.Object, w)
}

func (p *JSONPrinter) Flush(io.Writer) error { return nil }

type YAMLPrinter struct {
	Delegate *printers.OmitManagedFieldsPrinter
}

func NewYAMLPrinter() *YAMLPrinter {
	return &YAMLPrinter{
		Delegate: &printers.OmitManagedFieldsPrinter{
			Delegate: &printers.YAMLPrinter{},
		},
	}
}

func (p *YAMLPrinter) PrintNode(node *topology.Node, w io.Writer) error {
	return p.Delegate.PrintObj(node.Object, w)
}

func (p *YAMLPrinter) Flush(io.Writer) error { return nil }
