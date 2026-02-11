#!/usr/bin/env python3
# Copyright 2024 The Kubernetes Authors.
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

"""
Read conformance reports from conformance/reports/ and output controller-wizard-data.json
for the Controller Recommendation Wizard. Supports --all (one object keyed by version) or
--version vX.Y.Z (single array for that version). Output path defaults to
site-src/implementations/controller-wizard-data.json.
"""

import argparse
import json
import os
import re
import sys

import yaml


# Profile name (from report) -> conformance label shown in wizard
PROFILE_TO_CONFORMANCE = {
    "GATEWAY-HTTP": "Core",
    "GATEWAY-GRPC": "GRPCRoute",
    "GATEWAY-TLS": "TLSRoute",
}

# Acronyms that should stay uppercase in display names (e.g. GKE not Gke)
DISPLAY_ACRONYMS = frozenset({"gke"})

# Feature definitions for the wizard: id and label match v1.4 conformance table columns
# https://gateway-api.sigs.k8s.io/implementations/v1.4/
FEATURE_DEFINITIONS = {
    "httpGateway": [
        {"id": "GatewayPort8080", "label": "Gateway Port 8080"},
        {"id": "GatewayAddressEmpty", "label": "Gateway Address Empty"},
        {"id": "GatewayHTTPListenerIsolation", "label": "Gateway HTTP Listener Isolation"},
        {"id": "GatewayInfrastructurePropagation", "label": "Gateway Infrastructure Propagation"},
        {"id": "GatewayStaticAddresses", "label": "Gateway Static Addresses"},
    ],
    "httpRoute": [
        {"id": "HTTPRouteHostRewrite", "label": "HTTPRoute Host Rewrite"},
        {"id": "HTTPRoutePathRedirect", "label": "HTTPRoute Path Redirect"},
        {"id": "HTTPRouteRequestMirror", "label": "HTTPRoute Request Mirror"},
        {"id": "HTTPRouteRequestPercentageMirror", "label": "HTTPRoute Request Percentage Mirror"},
        {"id": "HTTPRouteResponseHeaderModification", "label": "HTTPRoute Response Header Modification"},
        {"id": "HTTPRouteSchemeRedirect", "label": "HTTPRoute Scheme Redirect"},
        {"id": "HTTPRouteBackendProtocolH2C", "label": "HTTPRoute Backend Protocol H2C"},
        {"id": "HTTPRouteBackendProtocolWebSocket", "label": "HTTPRoute Backend Protocol Web Socket"},
        {"id": "HTTPRouteBackendRequestHeaderModification", "label": "HTTPRoute Backend Request Header Modification"},
        {"id": "HTTPRouteBackendTimeout", "label": "HTTPRoute Backend Timeout"},
        {"id": "HTTPRouteCORS", "label": "HTTPRoute CORS"},
        {"id": "HTTPRouteDestinationPortMatching", "label": "HTTPRoute Destination Port Matching"},
        {"id": "HTTPRouteMethodMatching", "label": "HTTPRoute Method Matching"},
        {"id": "HTTPRouteNamedRouteRule", "label": "HTTPRoute Named Route Rule"},
        {"id": "HTTPRouteParentRefPort", "label": "HTTPRoute Parent Ref Port"},
        {"id": "HTTPRoutePathRewrite", "label": "HTTPRoute Path Rewrite"},
        {"id": "HTTPRoutePortRedirect", "label": "HTTPRoute Port Redirect"},
        {"id": "HTTPRouteQueryParamMatching", "label": "HTTPRoute Query Param Matching"},
        {"id": "HTTPRouteRequestMultipleMirrors", "label": "HTTPRoute Request Multiple Mirrors"},
        {"id": "HTTPRouteRequestTimeout", "label": "HTTPRoute Request Timeout"},
    ],
    "httpBackendTls": [
        {"id": "BackendTLSPolicy", "label": "Backend TLS Policy"},
        {"id": "BackendTLSPolicySANValidation", "label": "Backend TLS Policy SAN Validation"},
    ],
    "grpc": [
        {"id": "GatewayAddressEmpty", "label": "Gateway Address Empty"},
        {"id": "GatewayHTTPListenerIsolation", "label": "Gateway HTTP Listener Isolation"},
        {"id": "GatewayInfrastructurePropagation", "label": "Gateway Infrastructure Propagation"},
        {"id": "GatewayPort8080", "label": "Gateway Port 8080"},
        {"id": "GatewayStaticAddresses", "label": "Gateway Static Addresses"},
    ],
    "tls": [
        {"id": "GatewayAddressEmpty", "label": "Gateway Address Empty"},
        {"id": "GatewayHTTPListenerIsolation", "label": "Gateway HTTP Listener Isolation"},
        {"id": "GatewayInfrastructurePropagation", "label": "Gateway Infrastructure Propagation"},
        {"id": "GatewayPort8080", "label": "Gateway Port 8080"},
        {"id": "GatewayStaticAddresses", "label": "Gateway Static Addresses"},
    ],
}


def load_yaml(path):
    with open(path, "r", encoding="utf-8") as f:
        return yaml.safe_load(f)


def parse_version(version_dir_name):
    """Return (major, minor, patch) or (0,0,0) for sorting."""
    m = re.match(r"v?(\d+)\.(\d+)\.(\d+)", version_dir_name)
    if m:
        return (int(m.group(1)), int(m.group(2)), int(m.group(3)))
    return (0, 0, 0)


def get_version_dirs(reports_root, patch_zero_only=False):
    """Return sorted list of version directory paths (e.g. .../v1.4.0)."""
    if not os.path.isdir(reports_root):
        return []
    dirs = [
        os.path.join(reports_root, d)
        for d in os.listdir(reports_root)
        if os.path.isdir(os.path.join(reports_root, d)) and parse_version(d) != (0, 0, 0)
    ]
    if patch_zero_only:
        dirs = [d for d in dirs if parse_version(os.path.basename(d))[2] == 0]
    dirs.sort(key=lambda d: parse_version(os.path.basename(d)), reverse=True)
    return dirs


def normalize_url(url):
    if not url or not isinstance(url, str):
        return ""
    url = url.strip()
    if url and not url.startswith("http://") and not url.startswith("https://"):
        return "https://" + url
    return url


def display_name(organization, project):
    """Human-readable name; project is often the key (e.g. envoy-gateway -> Envoy Gateway)."""
    parts = project.split("-")
    if parts and parts[0].lower() in DISPLAY_ACRONYMS:
        name = parts[0].upper() + (
            " " + " ".join(p.title() for p in parts[1:]) if len(parts) > 1 else ""
        )
    else:
        name = project.replace("-", " ").title()
    if organization and organization != project:
        org_display = (
            organization.upper()
            if organization.lower() in DISPLAY_ACRONYMS
            else organization.replace("-", " ").title()
        )
        if org_display not in name:
            name = org_display + " " + name
    return name


def process_report(path):
    """Load one report YAML and return (implementation_info, conformance_list, features_set)."""
    data = load_yaml(path)
    if not data or "implementation" not in data or "profiles" not in data:
        return None

    impl = data["implementation"]
    organization = str(impl.get("organization") or "").strip()
    project = str(impl.get("project") or "").strip()
    url = normalize_url(str(impl.get("url") or ""))
    version = str(impl.get("version") or "").strip()
    report_date = str(data.get("date") or "").strip()
    if report_date and "T" in report_date:
        report_date = report_date.split("T")[0]
    mode = str(data.get("mode") or "default").strip()

    conformance = []
    features_set = set()

    for profile in data.get("profiles") or []:
        name = str(profile.get("name") or "").strip()
        label = PROFILE_TO_CONFORMANCE.get(name)
        core = profile.get("core") or {}
        core_result = str(core.get("result") or "").strip().lower()
        if label and core_result in ("success", "partial"):
            conformance.append(label)
        extended = profile.get("extended") or {}
        for feat in extended.get("supportedFeatures") or []:
            if isinstance(feat, str) and feat.strip():
                features_set.add(feat.strip())

    return {
        "name": display_name(organization, project),
        "organization": organization,
        "project": project,
        "url": url,
        "version": version,
        "reportDate": report_date,
        "mode": mode,
        "conformance": sorted(conformance),
        "features": sorted(features_set),
    }


def aggregate_by_impl(version_dir):
    """
    Walk all YAML reports under version_dir and aggregate by (organization, project, version, mode).
    Returns list of implementation dicts (features merged across profiles).
    """
    seen = {}
    for root, _dirs, files in os.walk(version_dir):
        for f in files:
            if not f.endswith(".yaml") or f.startswith("."):
                continue
            path = os.path.join(root, f)
            row = process_report(path)
            if not row:
                continue
            key = (
                row["organization"],
                row["project"],
                row["version"],
                row["mode"],
            )
            if key not in seen:
                seen[key] = {
                    "name": row["name"],
                    "organization": row["organization"],
                    "project": row["project"],
                    "url": row["url"],
                    "version": row["version"],
                    "reportDate": row["reportDate"],
                    "conformance": list(row["conformance"]),
                    "features": set(row["features"]),
                }
            else:
                seen[key]["conformance"] = sorted(
                    set(seen[key]["conformance"]) | set(row["conformance"])
                )
                seen[key]["features"] |= set(row["features"])

    out = []
    for v in seen.values():
        v["conformance"] = sorted(v["conformance"])
        v["features"] = sorted(v["features"])
        out.append(v)
    out.sort(key=lambda x: (x["organization"], x["project"], x["version"]))
    return out


def main():
    parser = argparse.ArgumentParser(
        description="Generate controller-wizard-data.json from conformance reports."
    )
    parser.add_argument(
        "--all",
        action="store_true",
        help="Output one object keyed by version (v1.4.0, v1.3.0, ...) for version dropdown.",
    )
    parser.add_argument(
        "--version",
        metavar="vX.Y.Z",
        help="Output a single array for this version (e.g. v1.4.0).",
    )
    parser.add_argument(
        "-o",
        "--output",
        default="site-src/implementations/controller-wizard-data.json",
        help="Output JSON path (default: site-src/implementations/controller-wizard-data.json)",
    )
    args = parser.parse_args()

    repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    reports_root = os.path.join(repo_root, "conformance", "reports")
    output = os.path.join(repo_root, args.output)

    if args.all:
        version_dirs = get_version_dirs(reports_root, patch_zero_only=True)
        if not version_dirs:
            print("No version dirs under conformance/reports/", file=sys.stderr)
            sys.exit(1)
        out = {"featureDefinitions": FEATURE_DEFINITIONS}
        for version_dir in version_dirs:
            v = os.path.basename(version_dir)
            out[v] = aggregate_by_impl(version_dir)
    elif args.version:
        version_dir = os.path.join(reports_root, args.version)
        if not os.path.isdir(version_dir):
            print(f"Version directory not found: {version_dir}", file=sys.stderr)
            sys.exit(1)
        impls = aggregate_by_impl(version_dir)
        out = {"featureDefinitions": FEATURE_DEFINITIONS, "implementations": impls}
    else:
        parser.error("Specify --all or --version vX.Y.Z")

    os.makedirs(os.path.dirname(output), exist_ok=True)
    with open(output, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=2)

    if args.all:
        version_keys = [k for k in out if k != "featureDefinitions"]
        total = sum(len(out[v]) for v in version_keys)
        print(f"Wrote {output} with featureDefinitions and {len(version_keys)} version(s), {total} implementation(s).")
    else:
        print(f"Wrote {output} with featureDefinitions and {len(out['implementations'])} implementation(s).")


if __name__ == "__main__":
    main()
