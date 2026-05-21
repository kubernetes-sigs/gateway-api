#!/usr/bin/env python3
# Copyright The Kubernetes Authors.
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
from pathlib import Path
import yaml
import pandas
from fnmatch import fnmatch
import glob
import os
import re
import semver
import sys

logging.basicConfig(level=logging.INFO)
log = logging.getLogger(__name__)


def process_feature_name(feature):
    """
    Process feature names by splitting camelCase into space-separated words
    """
    # Split camelCase
    words = re.findall(r'HTTPRoute|[A-Z]+(?=[A-Z][a-z])|[A-Z][a-z]+|[A-Z\d]+', feature)
    # Join words with spaces
    return ' '.join(words)


desc = """
The following tables are populated from the conformance reports [uploaded by project implementations](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports). They are separated into the extended features that each project supports listed in their reports.
Implementations only appear in this page if they pass Core conformance for the resource type, and the features listed should be Extended features. Implementations that submit conformance reports with skipped tests won't appear in the tables.
"""


def generate_conformance_tables(reports, currVersion, out_dir, is_hidden=False):
    gateway_tls_table = pandas.DataFrame()
    gateway_grpc_table = pandas.DataFrame()

    is_v1_0 = parse_release(currVersion) < semver.VersionInfo.parse('1.1.0')
    if not is_v1_0:
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

    v = ".".join(currVersion.split(".")[:2])
    version_short = v[1:] if v.startswith("v") else v
    entries = gateway_http_table.nunique()

    if entries.Project < 3:
        return

    minor = 0
    try:
        minor = int(version_short.split(".")[1])
    except:
        pass
    weight = (6 - minor) * 10

    try:
        f = StringIO()

        f.write("---\n")
        f.write(f'title: "{currVersion.split(".")[0]}.{minor}"\n')
        f.write(f"weight: {weight}\n")
        if is_hidden:
            f.write("hide_summary: true\n")
            f.write("toc_hide: true\n")
        f.write("---\n\n")

        f.write(desc.strip())
        f.write("\n\n")

        f.write('{{% alert title="Warning" color="warning" %}}\n')
        f.write("This page is under active development and is not in its final form, ")
        f.write("especially for the project name and the names of the features. ")
        f.write("However, as it is based on submitted conformance reports, the information is correct.\n")
        f.write('{{% /alert %}}\n\n')

        f.write("## Gateway Profile\n\n")
        f.write("### HTTPRoute\n\n")
        f.write(gateway_http_table.to_markdown()+'\n\n')
        if not is_v1_0:
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

    # Use vX.Y naming convention
    filename = f"v{version_short}.md"
    file_path = out_dir / filename
    with open(file_path, "w") as f_out:
        f_out.write(file_contents)
    log.info(f"Generated {file_path}")


def generate_profiles_report(reports, route, version):
    http_reports = reports.loc[reports["name"] == route].copy()
    for column in [
        "extended.result",
        "core.statistics.Passed",
        "core.statistics.Failed",
        "core.statistics.Skipped",
        "extended.statistics.Passed",
        "extended.statistics.Failed",
        "extended.statistics.Skipped",
        "extended.supportedFeatures",
        "extended.unsupportedFeatures",
    ]:
        if column not in http_reports.columns:
            http_reports[column] = 0 if ".statistics." in column else None
    http_reports.set_index('organization')
    http_reports.sort_values(
        ['organization', 'version'],
        key=lambda col: col.str.casefold() if col.name == 'organization' else col,
        inplace=True,
    )

    http_table = pandas.DataFrame(columns=http_reports['organization'])
    http_table = http_reports[['organization', 'project',
                               'version', 'mode', 'core.result', 'extended.result',
                               'core.statistics.Passed', 'core.statistics.Failed', 'core.statistics.Skipped',
                               'extended.statistics.Passed', 'extended.statistics.Failed', 'extended.statistics.Skipped',
                               'extended.supportedFeatures',
                               'extended.unsupportedFeatures']].T
    http_table.columns = http_table.iloc[0]
    http_table = http_table[1:].T

    for idx, row in http_table.iterrows():
        row_filter = (http_table.index == idx) & \
                     (http_table['project'] == row['project']) & \
                     (http_table['version'] == row['version']) & \
                     (http_table['mode'] == row['mode'])

        if row['core.result'] == "success":
            http_table.loc[row_filter, 'core.result'] = '✅'
        else:
            http_table.loc[row_filter, 'core.result'] = '❌'

        if isinstance(row['extended.supportedFeatures'], list):
            for feat in row['extended.supportedFeatures']:
                processed_feat = process_feature_name(feat)
                http_table.loc[row_filter, processed_feat] = '✅'
        if isinstance(row['extended.unsupportedFeatures'], list):
            for feat in row['extended.unsupportedFeatures']:
                processed_feat = process_feature_name(feat)
                http_table.loc[row_filter, processed_feat] = '❌'
    http_table = http_table.fillna('❌')

    http_table = http_table.rename(
        columns={
            "project": "Project",
            "version": "Version",
            "mode": "Mode",
            "core.result": "Core",
            "extended.result": "Extended Result",
            "core.statistics.Passed": "Core Passed",
            "core.statistics.Failed": "Core Failed",
            "core.statistics.Skipped": "Core Skipped",
            "extended.statistics.Passed": "Extended Passed",
            "extended.statistics.Failed": "Extended Failed",
            "extended.statistics.Skipped": "Extended Skipped",
        })

    def stat_value(value):
        if pandas.isna(value) or value == '❌':
            return 0
        return int(value)

    def count_features(features):
        if isinstance(features, list):
            return len(features)
        return 0

    def build_features_cell(row):
        core_total = stat_value(row['Core Passed']) + stat_value(row['Core Failed']) + stat_value(row['Core Skipped'])
        core_failed = stat_value(row['Core Failed'])
        extended_total = stat_value(row['Extended Passed']) + stat_value(row['Extended Failed']) + stat_value(row['Extended Skipped'])
        extended_failed = stat_value(row['Extended Failed'])
        extended_skipped = stat_value(row['Extended Skipped'])
        supported_features = count_features(row['extended.supportedFeatures'])
        unsupported_features = count_features(row['extended.unsupportedFeatures'])
        total_features = supported_features + unsupported_features

        lines = []
        if core_failed > 0 and core_total > 0:
            lines.append(f"<b>Failing {core_failed}/{core_total} core tests</b>")
        lines.append(f"{supported_features}/{total_features} features")
        if row['Extended Result'] == 'partial' and extended_total > 0:
            lines.append(f"<b>Partially conformant; {extended_skipped}/{extended_total} tests skipped</b>")
        elif row['Extended Result'] == 'failure' and extended_total > 0:
            lines.append(f"<b>Failing {extended_failed}/{extended_total} tests</b>")
        return '<br>'.join(lines)

    http_table["Extended Features"] = http_table.apply(build_features_cell, axis=1)
    http_table = http_table.drop([
        'Extended Result',
        'Core Passed',
        'Core Failed',
        'Core Skipped',
        'Extended Passed',
        'Extended Failed',
        'Extended Skipped',
        'extended.supportedFeatures',
        'extended.unsupportedFeatures',
    ], axis=1)
    
    if "Mode" in http_table.columns:
        insert_at = http_table.columns.get_loc("Mode") + 1
    else:
        insert_at = http_table.columns.get_loc("Version") + 1
    features_col = http_table.pop("Extended Features")
    http_table.insert(insert_at, "Extended Features", features_col)

    if parse_release(version) < semver.VersionInfo.parse('1.4.0'):
        http_table = http_table.drop(columns=["Core"])
    if version == 'v1.0.0':
        http_table = http_table.drop(columns=["Mode"])
    return http_table


pathTemp = "conformance/reports/*/"
def parse_release(version):
    v = version.lstrip('v')
    parts = v.split('.')
    if len(parts) == 2:
        v = f"{v}.0"
    return semver.VersionInfo.parse(v)


def release_key(version):
    parsed = parse_release(version)
    return (parsed.major, parsed.minor, parsed.patch)


def getConformancePaths():
    versions = []
    for v in glob.glob(pathTemp, recursive=True):
        release = v.split(os.sep)[-2]
        release_semver = parse_release(release)
        if release_semver < semver.VersionInfo.parse("1.0.0"):
            continue
        versions.append((release_semver, release, v))

    versions.sort(key=lambda x: x[0])

    minors = {}
    for release_semver, release, path in versions:
        minor = f"v{release_semver.major}.{release_semver.minor}"
        if minor not in minors:
            minors[minor] = {"minor": minor, "latest": release, "paths": []}
        minors[minor]["paths"].append(path + "**")
        if parse_release(minors[minor]["latest"]) < release_semver:
            minors[minor]["latest"] = release

    return sorted(minors.values(), key=lambda x: parse_release(x["latest"]))


def getYaml(conf_paths):
    yamls = []
    for conf_path in conf_paths:
        release_version = conf_path.split(os.sep)[-2]
        for p in glob.glob(conf_path, recursive=True):
            if fnmatch(p, "*.yaml"):
                x = load_yaml(p)
                if 'profiles' in x:
                    profiles = pandas.json_normalize(
                        x, record_path=['profiles'], meta=["mode","implementation"], errors='ignore')
                    implementation = pandas.json_normalize(profiles.implementation)
                    report = pandas.concat([implementation, profiles], axis=1)
                    report["reportRelease"] = release_version
                    yamls.append(report)

    yamls = pandas.concat(yamls)
    yamls["reportReleaseKey"] = yamls["reportRelease"].map(release_key)
    latest_release_key = yamls.groupby(
        ["organization", "project"]
    )["reportReleaseKey"].transform("max")
    yamls = yamls[yamls["reportReleaseKey"] == latest_release_key]

    yamls = yamls.sort_values("reportReleaseKey").drop_duplicates(
        subset=["organization", "project", "version", "name", "mode"], keep="last"
    )
    yamls = yamls.drop(columns=["reportReleaseKey"])
    return yamls


def load_yaml(name):
    with open(name, 'r') as file:
        x = yaml.safe_load(file)
    return x


def main():
    log.info("Generating conformance tables for Hugo/Docsy...")
    out_dir = Path("site/content/en/docs/implementations/versions")
    out_dir.mkdir(parents=True, exist_ok=True)

    index_path = out_dir / "_index.md"
    if not index_path.exists():
        with open(index_path, "w") as f:
            f.write("---\n")
            f.write('title: "Implementations"\n')
            f.write("weight: 40\n")
            f.write("---\n")
            f.write("Conformance data for Gateway API Implementations.\n")

    release_groups = getConformancePaths()
    for i, group in enumerate(release_groups):
        confYamls = getYaml(group["paths"])
        is_hidden = i < len(release_groups) - 4
        generate_conformance_tables(confYamls, group["latest"], out_dir, is_hidden)


if __name__ == "__main__":
    main()
