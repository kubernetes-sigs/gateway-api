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

package delete

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func NewCmd(factory common.Factory, iostreams genericiooptions.IOStreams) *cobra.Command {
	fileNameFlags := genericclioptions.NewResourceBuilderFlags().FileNameFlags
	fileNameFlags.Usage = "The files that contain the configurations to apply."
	fileNameFlags.Recursive = ptr.To(false)
	fileNameFlags.Kustomize = ptr.To("")

	flags := &deleteFlags{
		fileNameFlags: fileNameFlags,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources by file names, stdin, resources and names, or by resources and label selector.",
		Run: func(_ *cobra.Command, args []string) {
			o, err := flags.ToOptions(args, factory, iostreams)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				os.Exit(1)
			}

			err = o.Run(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				os.Exit(1)
			}
		},
	}

	flags.fileNameFlags.AddFlags(cmd.Flags())
	return cmd
}

// deleteFlags contains the flags used with delete command.
type deleteFlags struct {
	fileNameFlags *genericclioptions.FileNameFlags
}

func (f *deleteFlags) ToOptions(_ []string, factory common.Factory, iostreams genericiooptions.IOStreams) (*deleteOptions, error) {
	namespace, _, _ := factory.KubeConfigNamespace()

	return &deleteOptions{
		fileNameOptions: f.fileNameFlags.ToOptions(),
		factory:         factory,
		namespace:       namespace,
		IOStreams:       iostreams,
	}, nil
}

type deleteOptions struct {
	fileNameOptions resource.FilenameOptions
	factory         common.Factory
	namespace       string

	genericclioptions.IOStreams
}

func (o *deleteOptions) Run(args []string) error {
	infos, err := o.factory.NewBuilder().
		Unstructured().
		FilenameParam(false, &o.fileNameOptions).
		ResourceTypeOrNameArgs(false, args...).RequireObject(false).
		Flatten().
		NamespaceParam(o.namespace).DefaultNamespace().
		ContinueOnError().
		Do().
		Infos()
	if err != nil {
		return err
	}

	// Loop over all objects from the file(s)/stdin/args.
	for _, info := range infos {
		helper := resource.NewHelper(info.Client, info.Mapping)
		_, err := helper.Delete(info.Namespace, info.Name)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
			fmt.Fprintf(o.IOStreams.Out, "Error when deleting %v: %v\n", info.ObjectName(), err)
		} else {
			fmt.Fprintf(o.IOStreams.Out, "%v deleted\n", info.ObjectName())
		}
	}

	return nil
}
