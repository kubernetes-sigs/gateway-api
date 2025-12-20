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

###
# Note: At least 3 implementations have to upload their report under a version folder in order for the table to be generated
###

import logging
from io import StringIO
from mkdocs import plugins
from mkdocs.structure.files import File
import yaml
import pandas
from fnmatch import fnmatch
import glob
import os
import re
import semver

log = logging.getLogger(f'mkdocs.plugins.{__name__}')


def process_feature_name(feature):
    """
    Process feature names by splitting camelCase into space-separated words
    """
    # Split camelCase
    words = re.findall(r'HTTPRoute|[A-Z]+(?=[A-Z][a-z])|[A-Z][a-z]+|[A-Z\d]+', feature)
    # Join words with spaces
    return ' '.join(words)


@plugins.event_priority(100)
def on_files(files, config, **kwargs):
    log.info("generating conformance")

    vers = getConformancePaths()
    # Iterate over the list of versions. Exclude the pre 1.0 versions.
    for v in vers[3:]:

        confYamls = getYaml(v)
        releaseVersion = v.split(os.sep)[-2]
        file = generate_conformance_tables(confYamls, releaseVersion, config)

        if file:
          existing_file = files.get_file_from_path(file.src_uri)
          if existing_file:
              # Remove the existing file that is likely present in the
              # repository
              files.remove(existing_file)

          # Add the generated file to the site
          files.append(file)

          # Write the generated file to the site-src directory
          with open(os.path.join("site-src", file.src_uri), "w") as f:
            f.write(file.content_string)

    return files


desc = """
The following tables are populated from the conformance reports [uploaded by project implementations](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports). They are separated into the extended features that each project supports listed in their reports.
Implementations only appear in this page if they pass Core conformance for the resource type, and the features listed should be Extended features. Implementations that submit conformance reports with skipped tests won't appear in the tables.
"""

warning_text = """
???+ warning


    This page is under active development and is not in its final form,
    especially for the project name and the names of the features.
    However, as it is based on submitted conformance reports, the information is correct.
"""



def generate_conformance_tables(reports, currVersion, mkdocsConfig):

    # Enable Pandas copy-on-write
    pandas.options.mode.copy_on_write = True

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

    try:
        f = StringIO()

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

        file_contents = f.getvalue()
    finally:
        f.close()

    new_file = File(
      src_dir=None,
      dest_dir=mkdocsConfig['site_dir'],
      path=f'implementations/{versionFile}.md',
      use_directory_urls=mkdocsConfig['use_directory_urls'],
    )
    new_file.content_string = file_contents
    new_file.generated_by = f'{__name__}'

    return new_file

def generate_profiles_report(reports, route, version):

    http_reports = reports.loc[reports["name"] == route]
    http_reports.set_index('organization')
    http_reports.sort_values(['organization', 'version'], inplace=True)

    http_table = pandas.DataFrame(
        columns=http_reports['organization'])

    http_table = http_reports[['organization', 'project',
                               'version', 'mode', 'core.result', 'extended.supportedFeatures']].T
    http_table.columns = http_table.iloc[0]
    http_table = http_table[1:].T
    # change core.result value

    for row in http_table.itertuples():
        if row._4 == "success":
            http_table.loc[(row.Index, 'core.result')] = ':white_check_mark:'
        else:
            http_table.loc[(row.Index, 'core.result')] = ':x:'

        if type(row._5) is list:
            for feat in row._5:
                # Process feature name before using it as a column
                processed_feat = process_feature_name(feat)
                http_table.loc[(http_table.index == row.Index) & \
                               (http_table['project'] == row.project) & \
                               (http_table['version'] == row.version) & \
                               (http_table['mode'] == row.mode), processed_feat] = ':white_check_mark:'
    http_table = http_table.fillna(':x:')
    http_table = http_table.drop(['extended.supportedFeatures'], axis=1)

    http_table = http_table.rename(
        columns={"project": "Project", "version": "Version", "mode": "Mode", "core.result": "Core"})
    if semver.compare(version.removeprefix('v'), '1.4.0') < 0:
        http_table = http_table.drop(columns=["Core"])
    if version == 'v1.0.0':
        http_table = http_table.drop(columns=["Mode"])
    return http_table


pathTemp = "conformance/reports/*/"
allVersions = set()
reportedImplementationsPath = set()

# returns v1.0.0 and greater, since that's when reports started being generated in the comparison table


def getConformancePaths():
    versions = sorted(glob.glob(pathTemp, recursive=True))
    report_path = versions[-1]+"**"
    for v in versions:
        vers = v.split(os.sep)[-2]
        allVersions.add(vers)
        reportedImplementationsPath.add(v+"**")

    return sorted(list(reportedImplementationsPath))


def getYaml(conf_path):
    yamls = []

    for p in glob.glob(conf_path, recursive=True):

        if fnmatch(p, "*.yaml"):

            x = load_yaml(p)
            if 'profiles' in x:
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
