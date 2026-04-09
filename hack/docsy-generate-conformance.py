#!/usr/bin/env python3
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


def load_feature_channels():
    feature_constants = {}
    feature_channels = {}

    for path in Path("pkg/features").glob("*.go"):
        contents = path.read_text()
        for const_name, feature_name in re.findall(r'([A-Za-z0-9_]+)(?:\s+FeatureName)?\s*=\s*"([^"]+)"', contents):
            feature_constants[const_name] = feature_name
        for const_name, channel in re.findall(
            r'Name:\s*([A-Za-z0-9_]+),\s*[\r\n]+\s*Channel:\s*FeatureChannel(Standard|Experimental)',
            contents,
        ):
            feature_name = feature_constants.get(const_name)
            if feature_name:
                feature_channels[process_feature_name(feature_name)] = channel.lower()

    # Historical reports used the pre-rename feature name.
    feature_channels[process_feature_name("HTTPResponseHeaderModification")] = "standard"

    return feature_channels


FEATURE_CHANNELS = load_feature_channels()


def version_to_words(v_str):
    mapping = {
        "0": "zero",
        "1": "one",
        "2": "two",
        "3": "three",
        "4": "four",
        "5": "five",
        "6": "six",
        "7": "seven",
        "8": "eight",
        "9": "nine",
    }
    parts = v_str.split(".")
    if len(parts) >= 2:
        maj = mapping.get(parts[0], parts[0])
        min = mapping.get(parts[1], parts[1])
        return f"v_{maj}_{min}"
    return f"v_{v_str}"


desc = """
The following tables are populated from the conformance reports [uploaded by project implementations](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports). They are separated into the extended features that each project supports listed in their reports.
Implementations only appear in this page if they pass Core conformance for the resource type, and the features listed should be Extended features. Implementations that submit conformance reports with skipped tests won't appear in the tables.
"""


def generate_conformance_tables(reports, currVersion, out_dir, is_hidden=False):
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

    filename = version_to_words(version_short) + ".md"
    file_path = out_dir / filename
    with open(file_path, "w") as f_out:
        f_out.write(file_contents)
    log.info(f"Generated {file_path}")


def generate_profiles_report(reports, route, version):
    http_reports = reports.loc[reports["name"] == route].copy()
    http_reports.set_index('organization', inplace=True)
    http_reports.sort_values(
        ['organization', 'version'],
        key=lambda col: col.str.casefold() if col.name == 'organization' else col,
        inplace=True,
    )

    http_table = pandas.DataFrame(columns=http_reports.index)

    http_table = http_reports.reset_index()[['organization', 'project',
                               'version', 'mode', 'core.result', 'extended.supportedFeatures']].T
    http_table.columns = http_table.iloc[0]
    http_table = http_table[1:].T

    for row in http_table.itertuples():
        if row._4 == "success":
            http_table.loc[(http_table.index == row.Index), 'core.result'] = '✅'
        else:
            http_table.loc[(http_table.index == row.Index), 'core.result'] = '❌'

        if type(row._5) is list:
            for feat in row._5:
                # Process feature name before using it as a column
                processed_feat = process_feature_name(feat)
                http_table.loc[(http_table.index == row.Index) & \
                               (http_table['project'] == row.project) & \
                               (http_table['version'] == row.version) & \
                               (http_table['mode'] == row.mode), processed_feat] = '✅'
    http_table = http_table.fillna('❌')
    http_table = http_table.drop(['extended.supportedFeatures'], axis=1)

    http_table = http_table.rename(
        columns={"project": "Project", "version": "Version", "mode": "Mode", "core.result": "Core"})
    metadata_columns = ["Project", "Version", "Mode", "Core"]
    feature_columns = [c for c in http_table.columns if c not in metadata_columns]
    standard_feature_columns = [c for c in feature_columns if FEATURE_CHANNELS.get(c, "standard") == "standard"]
    experimental_feature_columns = [c for c in feature_columns if FEATURE_CHANNELS.get(c) == "experimental"]

    def format_total(columns):
        total = len(columns)
        if total == 0:
            return ['0/0'] * len(http_table.index)
        checks = (http_table[columns] == '✅').sum(axis=1)
        return checks.map(lambda count: f"{count}/{total}")

    http_table["Standard Features"] = format_total(standard_feature_columns)
    http_table["Experimental Features"] = format_total(experimental_feature_columns)
    if "Mode" in http_table.columns:
        insert_at = http_table.columns.get_loc("Mode") + 1
    else:
        insert_at = http_table.columns.get_loc("Version") + 1
    experimental_col = http_table.pop("Experimental Features")
    standard_col = http_table.pop("Standard Features")
    http_table.insert(insert_at, "Standard Features", standard_col)
    http_table.insert(insert_at + 1, "Experimental Features", experimental_col)

    if semver.compare(version.lstrip('v'), '1.4.0') < 0:
        http_table = http_table.drop(columns=["Core"])
    if version == 'v1.0.0':
        http_table = http_table.drop(columns=["Mode"])
    return http_table


pathTemp = "conformance/reports/*/"
def parse_release(version):
    return semver.VersionInfo.parse(version.lstrip('v'))


def release_key(version):
    parsed = parse_release(version)
    return (parsed.major, parsed.minor, parsed.patch)


def getConformancePaths():
    """
    Return release paths grouped by minor version.
    """
    versions = []
    for v in glob.glob(pathTemp, recursive=True):
        release = v.split(os.sep)[-2]
        release_semver = parse_release(release)
        # Reports prior to v1.0.0 are not included in generated tables.
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
    # If an implementation/profile appears in multiple patches for the same minor,
    # keep only the newest patch report.
    yamls["reportReleaseKey"] = yamls["reportRelease"].map(release_key)
    # For each implementation project, keep only rows from its newest patch
    # release within this Gateway API minor.
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

    # Ensure an index file exists so Hugo treats it as a section
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
