/*
Copyright The Kubernetes Authors.

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

package crd_test

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	aggregateToViewLabel  = "rbac.authorization.k8s.io/aggregate-to-view"
	aggregateToEditLabel  = "rbac.authorization.k8s.io/aggregate-to-edit"
	aggregateToAdminLabel = "rbac.authorization.k8s.io/aggregate-to-admin"
)

// TestAggregatedClusterRoles makes sure every CRD shipped in a channel is
// readable through the builtin "view", "edit" and "admin" roles, and that the
// aggregated roles never grant more than intended.
func TestAggregatedClusterRoles(t *testing.T) {
	for _, channel := range []string{"standard", "experimental"} {
		t.Run(channel, func(t *testing.T) {
			dir := filepath.Join("..", "..", "config", "crd", channel)
			files, err := filepath.Glob(filepath.Join(dir, "gateway*.yaml"))
			require.NoError(t, err)

			// "group/resource" and "group/resource/subresource" of every shipped CRD.
			crdResources := map[string]bool{}
			var roles []rbacv1.ClusterRole
			for _, file := range files {
				data, err := os.ReadFile(file)
				require.NoError(t, err)
				reader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
				for {
					doc, err := reader.Read()
					if err == io.EOF {
						break
					}
					require.NoError(t, err, file)
					if len(bytes.TrimSpace(doc)) == 0 {
						continue
					}
					var meta struct {
						Kind string `json:"kind"`
					}
					require.NoError(t, utilyaml.Unmarshal(doc, &meta), file)
					switch meta.Kind {
					case "CustomResourceDefinition":
						var crd apiextensionsv1.CustomResourceDefinition
						require.NoError(t, utilyaml.Unmarshal(doc, &crd), file)
						base := crd.Spec.Group + "/" + crd.Spec.Names.Plural
						crdResources[base] = true
						for _, v := range crd.Spec.Versions {
							if v.Subresources != nil && v.Subresources.Status != nil {
								crdResources[base+"/status"] = true
							}
						}
					case "ClusterRole":
						var role rbacv1.ClusterRole
						require.NoError(t, utilyaml.Unmarshal(doc, &role), file)
						roles = append(roles, role)
					}
				}
			}
			require.NotEmpty(t, crdResources)
			require.NotEmpty(t, roles)

			readable := map[string]map[string]bool{}
			for _, label := range []string{aggregateToViewLabel, aggregateToEditLabel, aggregateToAdminLabel} {
				readable[label] = map[string]bool{}
			}
			for _, role := range roles {
				assert.Empty(t, role.AggregationRule, role.Name)

				// A role either reads (aggregated into view, edit and admin, as the builtin
				// edit and admin do not inherit the rules of view) or writes (aggregated
				// into edit and admin only).
				labels := map[string]bool{}
				for _, label := range []string{aggregateToViewLabel, aggregateToEditLabel, aggregateToAdminLabel} {
					if role.Labels[label] == "true" {
						labels[label] = true
					}
				}
				reads := len(labels) == 3
				writes := len(labels) == 2 && labels[aggregateToEditLabel] && labels[aggregateToAdminLabel]
				require.True(t, reads || writes, "ClusterRole %s must carry either all of aggregate-to-view, -edit and -admin, or both aggregate-to-edit and aggregate-to-admin", role.Name)

				for _, rule := range role.Rules {
					assert.Empty(t, rule.NonResourceURLs, role.Name)
					assert.Empty(t, rule.ResourceNames, role.Name)
					for _, group := range rule.APIGroups {
						assert.True(t, strings.HasPrefix(group, "gateway.networking."), "%s: unexpected group %q", role.Name, group)
						for _, resource := range rule.Resources {
							assert.NotContains(t, resource, "*", role.Name)
							key := group + "/" + resource
							assert.True(t, crdResources[key], "%s: %s is not a CRD (or status subresource) in the %s channel", role.Name, key, channel)

							write, read := false, false
							for _, verb := range rule.Verbs {
								assert.NotContains(t, verb, "*", role.Name)
								switch verb {
								case "get", "list", "watch":
									read = true
									for label := range labels {
										readable[label][key] = true
									}
								default:
									write = true
								}
							}
							if writes {
								assert.False(t, read, "%s: write role must not carry read verbs", role.Name)
							}
							if write {
								assert.False(t, reads, "%s: read role must be read-only", role.Name)
								assert.False(t, strings.HasSuffix(resource, "/status"), "%s: status must not be writable", role.Name)
							}
						}
					}
				}
			}

			// Every kind is considered safe (credentials are only referenced by name),
			// so all of them are readable through view, and everything view can read
			// has to be readable through edit and admin as well.
			for key := range crdResources {
				assert.True(t, readable[aggregateToViewLabel][key], "%s is not readable through view", key)
				assert.True(t, readable[aggregateToEditLabel][key], "%s is not readable through edit", key)
				assert.True(t, readable[aggregateToAdminLabel][key], "%s is not readable through admin", key)
			}
			for key := range readable[aggregateToViewLabel] {
				assert.True(t, readable[aggregateToEditLabel][key], "%s is readable through view but not edit", key)
				assert.True(t, readable[aggregateToAdminLabel][key], "%s is readable through view but not admin", key)
			}
		})
	}
}
