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

package apply

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

const (
	fieldManager = "gwctl-server-side-apply"
)

func NewCmd(factory common.Factory, iostreams genericiooptions.IOStreams) *cobra.Command {
	fileNameFlags := genericclioptions.NewResourceBuilderFlags().FileNameFlags
	fileNameFlags.Usage = "The files that contain the configurations to apply."
	fileNameFlags.Recursive = ptr.To(false)
	fileNameFlags.Kustomize = ptr.To("")

	flags := &applyFlags{
		fileNameFlags: fileNameFlags,
	}

	cmd := &cobra.Command{
		Use:   "apply -f FILENAME|DIRECTORY",
		Short: "Apply the provided resources from file or stdin to the cluster.",
		Args:  cobra.ExactArgs(0),
		Run: func(_ *cobra.Command, args []string) {
			o, err := flags.ToOptions(args, factory, iostreams)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				os.Exit(1)
			}

			err = o.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				os.Exit(1)
			}
		},
	}

	flags.fileNameFlags.AddFlags(cmd.Flags())
	return cmd
}

// applyFlags contains the flags used with apply command.
type applyFlags struct {
	fileNameFlags *genericclioptions.FileNameFlags
}

func (f *applyFlags) ToOptions(_ []string, factory common.Factory, iostreams genericiooptions.IOStreams) (*applyOptions, error) {
	namespace, _, _ := factory.KubeConfigNamespace()

	return &applyOptions{
		fileNameOptions: f.fileNameFlags.ToOptions(),
		factory:         factory,
		namespace:       namespace,
		IOStreams:       iostreams,
	}, nil
}

type applyOptions struct {
	fileNameOptions resource.FilenameOptions
	factory         common.Factory
	namespace       string

	genericclioptions.IOStreams
}

func (o *applyOptions) Run() error {
	infos, err := o.factory.NewBuilder().
		Unstructured().
		FilenameParam(false, &o.fileNameOptions).
		Flatten().
		NamespaceParam(o.namespace).DefaultNamespace().
		ContinueOnError().
		Do().
		Infos()
	if err != nil {
		return err
	}

	printer := printers.NamePrinter{Operation: "configured"}

	// Loop over all objects from the file(s) or stdin.
	for _, info := range infos {
		helper := resource.NewHelper(info.Client, info.Mapping).WithFieldManager(fieldManager)

		data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, info.Object)
		if err != nil {
			return fmt.Errorf("%v: %v", info.Source, err)
		}

		obj, err := helper.Patch(
			info.Namespace,
			info.Name,
			types.ApplyPatchType,
			data,
			nil,
		)
		if err != nil {
			return err
		}

		err = printer.PrintObj(obj, o.Out)
		if err != nil {
			return err
		}
	}

	return nil
}
