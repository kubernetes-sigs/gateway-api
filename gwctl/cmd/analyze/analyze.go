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

package analyze

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/resource"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/directlyattachedpolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/gatewayeffectivepolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/notfoundrefvalidator"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/refgrantvalidator"
	extensionutils "sigs.k8s.io/gateway-api/gwctl/pkg/extension/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
	topologygw "sigs.k8s.io/gateway-api/gwctl/pkg/topology/gateway"
)

func NewCmd(factory common.Factory, iostreams genericiooptions.IOStreams) *cobra.Command {
	flags := &analyzeFlags{
		fileNameFlags: genericclioptions.NewResourceBuilderFlags().FileNameFlags,
	}

	cmd := &cobra.Command{
		Use:   "analyze -f FILENAME|DIRECTORY",
		Short: "Analyze resources by file names or stdin",
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

// analyzeFlags contains the flags used with analyze command.
type analyzeFlags struct {
	fileNameFlags *genericclioptions.FileNameFlags
}

func (f *analyzeFlags) ToOptions(_ []string, factory common.Factory, iostreams genericiooptions.IOStreams) (*analyzeOptions, error) {
	namespace, _, _ := factory.KubeConfigNamespace()

	return &analyzeOptions{
		fileNameOptions: f.fileNameFlags.ToOptions(),
		factory:         factory,
		namespace:       namespace,
		IOStreams:       iostreams,
	}, nil
}

type analyzeOptions struct {
	fileNameOptions resource.FilenameOptions
	factory         common.Factory
	namespace       string

	genericclioptions.IOStreams
}

func (o *analyzeOptions) Run() error {
	fmt.Fprintf(o.IOStreams.Out, "\n")
	fmt.Fprintf(o.IOStreams.Out, "Analyzing %v...\n", strings.Join(o.fileNameOptions.Filenames, ","))
	fmt.Fprintf(o.IOStreams.Out, "\n")

	// Step 1: Parse the files and extract the objects from the files.
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

	// Step 2: Classify whether the object already exists, or not. If it already
	// exists, cache the version which already exists.
	existingObjects := map[*resource.Info]*unstructured.Unstructured{}
	for _, info := range infos {
		helper := resource.NewHelper(info.Client, info.Mapping)
		obj, err := helper.Get(info.Namespace, info.Name) //nolint:govet
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
			existingObjects[info] = nil // Object does not exist.
			continue
		}
		// Object does exist, cache it.
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		u := &unstructured.Unstructured{Object: o}
		existingObjects[info] = u
	}

	// Step 3: Build graph using the provided objects in the files as the
	// source.
	sources := []*unstructured.Unstructured{}
	for _, info := range infos {
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(info.Object) //nolint:govet
		if err != nil {
			return err
		}
		u := &unstructured.Unstructured{Object: o}
		sources = append(sources, u)
	}
	graph, err := topology.NewBuilder(common.NewDefaultGroupKindFetcher(o.factory, common.WithAdditionalResources(sources))).
		StartFrom(sources).
		UseRelationships(topologygw.AllRelations).
		WithMaxDepth(4).
		Build()
	if err != nil {
		return err
	}

	policyManager := policymanager.New(common.NewDefaultGroupKindFetcher(o.factory, common.WithAdditionalResources(sources)))
	if err := policyManager.Init(); err != nil { //nolint:govet
		return err
	}
	// Execute extensions.
	err = extension.ExecuteAll(graph,
		directlyattachedpolicy.NewExtension(policyManager),
		gatewayeffectivepolicy.NewExtension(),
		refgrantvalidator.NewExtension(
			refgrantvalidator.NewDefaultReferenceGrantFetcher(o.factory, refgrantvalidator.WithAdditionalResources(sources)),
		),
		notfoundrefvalidator.NewExtension(),
	)
	if err != nil {
		return err
	}

	// Step 4: Collect errors from the graph. These are the collective set of
	// errors which will be observed after the new changes are applied.
	errorsAfterChanges, err := collectErrors(graph)
	if err != nil {
		return err
	}

	// Step 5: Remove nodes from the graph which are going to be newly created,
	// or revert them to their state before creation. The resulting graph should
	// represent a state which currently exists in the server (before applying
	// the newer changes.)
	for info, existingObject := range existingObjects {
		gknn := common.GKNN{
			Group:     info.Mapping.GroupVersionKind.Group,
			Kind:      info.Mapping.GroupVersionKind.Kind,
			Namespace: info.Namespace,
			Name:      info.Name,
		}
		if existingObject == nil {
			// This means the object would have been newly created, and thus we
			// need to delete it to revert the graph back to it's original
			// state.
			graph.DeleteNodeUsingGKNN(gknn)
		} else if !graph.HasNode(gknn) {
			node := graph.Nodes[gknn.GroupKind()][gknn.NamespacedName()]
			node.Object = existingObject // Revert object back to it's original state which exists in the server.
		}
	}

	// Step 6: Build new graph by running extensions
	policyManager = policymanager.New(common.NewDefaultGroupKindFetcher(o.factory))
	if err := policyManager.Init(); err != nil { //nolint:govet
		return err
	}
	// Execute extensions.
	err = extension.ExecuteAll(graph,
		directlyattachedpolicy.NewExtension(policyManager),
		gatewayeffectivepolicy.NewExtension(),
		refgrantvalidator.NewExtension(
			refgrantvalidator.NewDefaultReferenceGrantFetcher(o.factory),
		),
		notfoundrefvalidator.NewExtension(),
	)
	if err != nil {
		return err
	}

	// Step 6: Collect errors from the graph. These are the collective set of
	// errors which will be observed in the server before the new changes are
	// applied.
	errorsBeforeChanges, err := collectErrors(graph)
	if err != nil {
		return err
	}

	// Step 7: Report analysis

	fmt.Fprintf(o.IOStreams.Out, "Summary:\n")
	fmt.Fprintf(o.IOStreams.Out, "\n")
	created, updated := generateSummary(existingObjects)
	for _, info := range created {
		fmt.Fprintf(o.IOStreams.Out, "\t- Created %v", info.ObjectName())
		if info.Namespaced() {
			fmt.Fprintf(o.IOStreams.Out, " in namespace %v", info.Namespace)
		}
		fmt.Fprintf(o.IOStreams.Out, "\n")
	}
	for _, info := range updated {
		fmt.Fprintf(o.IOStreams.Out, "\t- Updated %v", info.ObjectName())
		if info.Namespaced() {
			fmt.Fprintf(o.IOStreams.Out, " in namespace %v", info.Namespace)
		}
		fmt.Fprintf(o.IOStreams.Out, "\n")
	}
	fmt.Fprintf(o.IOStreams.Out, "\n")

	newIssues, fixedIssues, unchangedIssues := classifyErrors(errorsBeforeChanges, errorsAfterChanges)

	fmt.Fprintf(o.IOStreams.Out, "Potential Issues Introduced\n")
	fmt.Fprintf(o.IOStreams.Out, "(These issues will arise after applying the changes in the analyzed file.):\n")
	fmt.Fprintf(o.IOStreams.Out, "\n")
	for _, s := range newIssues {
		fmt.Fprintf(o.IOStreams.Out, "\t- %v:\n", s)
	}
	if len(newIssues) == 0 {
		fmt.Fprintf(o.IOStreams.Out, "\tNone.\n")
	}
	fmt.Fprintf(o.IOStreams.Out, "\n")

	fmt.Fprintf(o.IOStreams.Out, "Existing Issues Fixed\n")
	fmt.Fprintf(o.IOStreams.Out, "(These issues were present before the changes but will be resolved after applying them.):\n")
	fmt.Fprintf(o.IOStreams.Out, "\n")
	for _, s := range fixedIssues {
		fmt.Fprintf(o.IOStreams.Out, "\t- %v:\n", s)
	}
	if len(fixedIssues) == 0 {
		fmt.Fprintf(o.IOStreams.Out, "\tNone\n")
	}
	fmt.Fprintf(o.IOStreams.Out, "\n")

	fmt.Fprintf(o.IOStreams.Out, "Existing Issues Unchanged\n")
	fmt.Fprintf(o.IOStreams.Out, "(These issues were present before the changes and will remain even after applying them.):\n")
	fmt.Fprintf(o.IOStreams.Out, "\n")
	for _, s := range unchangedIssues {
		fmt.Fprintf(o.IOStreams.Out, "\t- %v:\n", s)
	}
	if len(unchangedIssues) == 0 {
		fmt.Fprintf(o.IOStreams.Out, "\tNone\n")
	}
	fmt.Fprintf(o.IOStreams.Out, "\n")

	return nil
}

func collectErrors(graph *topology.Graph) (map[string]bool, error) {
	errors := map[string]bool{}
	for i := range graph.Nodes {
		for j := range graph.Nodes[i] {
			node := graph.Nodes[i][j]
			aggregateAnalysisErrors, err := extensionutils.AggregateAnalysisErrors(node)
			if err != nil {
				return nil, err
			}
			for _, err := range aggregateAnalysisErrors {
				s := fmt.Sprintf("%v: %v", node.GKNN(), err)
				errors[s] = true
			}
		}
	}
	return errors, nil
}

func generateSummary(objects map[*resource.Info]*unstructured.Unstructured) (created, updated []*resource.Info) {
	for info, existingObject := range objects {
		if existingObject == nil {
			created = append(created, info)
		} else {
			updated = append(updated, info)
		}
	}
	infoComparer := func(a, b *resource.Info) bool {
		p := fmt.Sprintf("%v/%v/%v", a.Object.GetObjectKind().GroupVersionKind().GroupKind(), a.Namespace, a.Name)
		q := fmt.Sprintf("%v/%v/%v", b.Object.GetObjectKind().GroupVersionKind().GroupKind(), b.Namespace, b.Name)
		return p < q
	}
	sort.Slice(created, func(i, j int) bool { return infoComparer(created[i], created[j]) })
	sort.Slice(updated, func(i, j int) bool { return infoComparer(updated[i], updated[j]) })
	return created, updated
}

func classifyErrors(errorsBeforeChanges, errorsAfterChanges map[string]bool) (newIssues, fixedIssues, unchangedIssues []string) {
	for s := range errorsAfterChanges {
		existsBefore := errorsBeforeChanges[s]
		if !existsBefore {
			newIssues = append(newIssues, s)
		} else {
			unchangedIssues = append(unchangedIssues, s)
		}
	}
	for s := range errorsBeforeChanges {
		existsAfter := errorsAfterChanges[s]
		if !existsAfter {
			fixedIssues = append(fixedIssues, s)
		} else {
			unchangedIssues = append(unchangedIssues, s)
		}
	}
	slices.Sort(newIssues)
	slices.Sort(fixedIssues)
	slices.Sort(unchangedIssues)
	return newIssues, fixedIssues, unchangedIssues
}
