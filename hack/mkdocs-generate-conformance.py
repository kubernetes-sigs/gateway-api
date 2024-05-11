# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import logging
from mkdocs import plugins
import yaml
import pandas
from fnmatch import fnmatch
import glob

log = logging.getLogger('mkdocs')


@plugins.event_priority(100)
def on_pre_build(config, **kwargs):
    log.info("generating conformance")

    yamlReports = getYaml()

    generate_conformance_tables(yamlReports)


desc = """
The following tables are populated from the conformance reports [uploaded by project implementations](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports). They are separated into the extended features that each project supports listed in their reports.
Implementations only appear in this page if they pass Core conformance for the resource type, and the features listed should be Extended features.
"""

warning_text = """
???+ warning


    This page is under active development and is not in its final form,
    especially for the project name and the names of the features.
    However, as it is based on submitted conformance reports, the information is correct.
"""

# NOTE: will have to be updated if new (extended) features are added
httproute_extended_conformance_features_list = ['HTTPRouteBackendRequestHeaderModification', 'HTTPRouteQueryParamMatching', 'HTTPRouteMethodMatching', 'HTTPRouteResponseHeaderModification', 'HTTPRoutePortRedirect', 'HTTPRouteSchemeRedirect',
                                                'HTTPRoutePathRedirect', 'HTTPRouteHostRewrite', 'HTTPRoutePathRewrite', 'HTTPRouteRequestMirror', 'HTTPRouteRequestMultipleMirrors', 'HTTPRouteRequestTimeout', 'HTTPRouteBackendTimeout', 'HTTPRouteParentRefPort']


def generate_conformance_tables(reports):

    gateway_http_table = generate_profiles_report(reports, 'HTTP')
    gateway_http_table = gateway_http_table.rename_axis('Organization')

    # Currently no implementation has extended supported features listed.
    # Can uncomment once a list is needed to keep track
    # gateway_tls_table = generate_profiles_report(reprots,'TLS')

    mesh_http_table = generate_profiles_report(reports, 'MESH')
    mesh_http_table = mesh_http_table.rename_axis('Organization')

    with open('site-src/implementation-table.md', 'w') as f:
        f.write(desc)
        f.write("\n\n")

        f.write(warning_text)
        f.write("\n\n")

        f.write("## Gateway Profile\n\n")
        f.write("### HTTPRoute\n\n")
        f.write(gateway_http_table.to_markdown()+'\n\n')

        f.write("## Mesh Profile\n\n")
        f.write("### HTTPRoute\n\n")
        f.write(mesh_http_table.to_markdown())


def generate_profiles_report(reports, route):

    http_reports = reports.loc[reports["name"] == route]
    http_reports.set_index('organization')
    http_reports.sort_values(['organization', 'version'], inplace=True)

    http_table = pandas.DataFrame(
        columns=http_reports['organization'])
    http_table = http_reports[['organization', 'project',
                               'version', 'extended.supportedFeatures']].T
    http_table.columns = http_table.iloc[0]
    http_table = http_table[1:].T

    for row in http_table.itertuples():
        for feat in row._3:
            http_table.loc[row.Index, feat] = ':white_check_mark:'
    http_table = http_table.fillna(':x:')
    http_table = http_table.drop(['extended.supportedFeatures'], axis=1)

    http_table = http_table.rename(
        columns={"project": "Project", "version": "Version"})

    return http_table


# the path should be changed when there is a new version
conformance_path = "conformance/reports/v1.0.0/**"


def getYaml():
    log.info("parsing conformance reports ============================")
    yamls = []

    for p in glob.glob(conformance_path, recursive=True):

        if fnmatch(p, "*.yaml"):

            x = load_yaml(p)
            profiles = pandas.json_normalize(
                x, record_path='profiles', meta=["implementation"])

            implementation = pandas.json_normalize(profiles.implementation)
            yamls.append(pandas.concat([implementation, profiles], axis=1))

    yamls = pandas.concat(yamls)
    return yamls


def load_yaml(name):
    with open(name, 'r') as file:
        x = yaml.safe_load(file)

    return x
