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
import os

log = logging.getLogger('mkdocs')


@plugins.event_priority(100)
def on_pre_build(config, **kwargs):
    log.info("generating conformance")

    vers = getConformancePaths()
    for v in vers[3:]:

        confYamls = getYaml(v)
        releaseVersion = v.split(os.sep)[-2]
        generate_conformance_tables(confYamls, releaseVersion)


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



def generate_conformance_tables(reports, currVersion):

    gateway_tls_table = pandas.DataFrame()
    gateway_grpc_table = pandas.DataFrame()

    if currVersion != 'v1.0.0':
        gateway_http_table = generate_profiles_report(reports, 'GATEWAY-HTTP',currVersion)

        gateway_grpc_table = generate_profiles_report(reports, 'GATEWAY-GRPC',currVersion)
        gateway_grpc_table = gateway_grpc_table.rename_axis('Organization')

        gateway_tls_table = generate_profiles_report(reports, 'GATEWAY-TLS',currVersion)
        gateway_tls_table = gateway_tls_table.rename_axis('Organization')

        mesh_http_table = generate_profiles_report(reports, 'MESH-HTTP',currVersion)
    else:
        gateway_http_table = generate_profiles_report(reports, "HTTP",currVersion)
        mesh_http_table = generate_profiles_report(reports, "MESH",currVersion)

    gateway_http_table = gateway_http_table.rename_axis('Organization')
    mesh_http_table = mesh_http_table.rename_axis('Organization')

    versionFile = ".".join(currVersion.split(".")[:2])
    entries =  gateway_http_table.nunique()

    if entries.Project < 3:
        return

    with open('site-src/implementations/'+versionFile+'.md', 'w') as f:

        f.write(desc)
        f.write("\n\n")

        f.write(warning_text)
        f.write("\n\n")

        f.write("## Gateway Profile\n\n")
        f.write("### HTTPRoute\n\n")
        f.write(gateway_http_table.to_markdown()+'\n\n')
        if currVersion != 'v1.0.0': 
            f.write('### GRPCRoute\n\n')
            f.write(gateway_grpc_table.to_markdown()+'\n\n')
            f.write('### TLSRoute\n\n')
            f.write(gateway_tls_table.to_markdown()+'\n\n')

        f.write("## Mesh Profile\n\n")
        f.write("### HTTPRoute\n\n")
        f.write(mesh_http_table.to_markdown())


def generate_profiles_report(reports, route,version):

    http_reports = reports.loc[reports["name"] == route]
    http_reports.set_index('organization')
    http_reports.sort_values(['organization', 'version'], inplace=True)

    http_table = pandas.DataFrame(
        columns=http_reports['organization'])

    http_table = http_reports[['organization', 'project',
                               'version','mode', 'extended.supportedFeatures']].T
    http_table.columns = http_table.iloc[0]
    http_table = http_table[1:].T
    
    for row in http_table.itertuples():
        if type(row._4) is list:
            for feat in row._4:
                http_table.loc[row.Index, feat] = ':white_check_mark:'
    http_table = http_table.fillna(':x:')
    http_table = http_table.drop(['extended.supportedFeatures'], axis=1)

    http_table = http_table.rename(
        columns={"project": "Project", "version": "Version", "mode":"Mode"})
    if version == 'v1.0.0':
        http_table = http_table.drop(columns=["Mode"])
    return http_table


pathTemp = "conformance/reports/*/"
allVersions = []
reportedImplementationsPath = []

# returns v1.0.0 and greater, since that's when reports started being generated in the comparison table


def getConformancePaths():
    versions = sorted(glob.glob(pathTemp, recursive=True))
    report_path = versions[-1]+"**"
    for v in versions:
        vers = v.split(os.sep)[-2]
        allVersions.append(vers)
        reportedImplementationsPath.append(v+"**")
    return reportedImplementationsPath


def getYaml(conf_path):
    yamls = []

    for p in glob.glob(conf_path, recursive=True):

        if fnmatch(p, "*.yaml"):

            x = load_yaml(p)
            profiles = pandas.json_normalize(
                x, record_path=['profiles'], meta=["mode","implementation"], errors='ignore')

            implementation = pandas.json_normalize(profiles.implementation)
            yamls.append(pandas.concat([implementation, profiles], axis=1))

    yamls = pandas.concat(yamls)
    return yamls


def load_yaml(name):
    with open(name, 'r') as file:
        x = yaml.safe_load(file)

    return x
