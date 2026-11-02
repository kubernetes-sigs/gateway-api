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

"""
Read conformance reports from conformance/reports/ and output controller-wizard-data.json
for the Controller Recommendation Wizard. Supports --all (one object keyed by version) or
--version vX.Y.Z (single array for that version). Output path defaults to
site/static/wizard/data/controller-wizard-data.json. The output directory is created if it does not exist.

When loaded as an MkDocs hook (see mkdocs.yml hooks), runs with --all behavior so the
built site has multi-version wizard data.
"""

import argparse
import json
import logging
import os
import re
import sys

import yaml

log = logging.getLogger(__name__)


# Profile name (from report) -> conformance label shown in wizard
PROFILE_TO_CONFORMANCE = {
    "GATEWAY-HTTP": "Core",
    "GATEWAY-GRPC": "GRPCRoute",
    "GATEWAY-TLS": "TLSRoute",
    "GATEWAY-TCP": "TCPRoute",
    "GATEWAY-UDP": "UDPRoute"
}

# Acronyms that should stay uppercase in display names (e.g. GKE not Gke)
DISPLAY_ACRONYMS = frozenset({"gke"})

# Human-readable labels/descriptions for known feature ids. Features not listed
# here get a label humanized from the id and a short generic description.
FEATURE_LABEL_OVERRIDES = {
    "GatewayPort8080": {"label": "Gateway Port 8080", "description": "Gateway listens on port 8080."},
    "GatewayAddressEmpty": {"label": "Gateway Address Empty", "description": "Gateway status supports empty address list."},
    "GatewayHTTPListenerIsolation": {"label": "Gateway HTTP Listener Isolation", "description": "HTTPRoutes can target a single listener; others are isolated."},
    "GatewayInfrastructurePropagation": {"label": "Gateway Infrastructure Propagation", "description": "Gateway status propagates infrastructure details to Routes."},
    "GatewayStaticAddresses": {"label": "Gateway Static Addresses", "description": "Gateway supports static address assignment."},
    "GatewayBackendClientCertificate": {"label": "Gateway Backend Client Certificate", "description": "Gateway supports client certificates selection for Gateway TLS origination."},
    "GatewayFrontendClientCertificateValidation": {"label": "Gateway Frontend Client Certificate Validation", "description": "Gateway supports client certificate validation for Gateway TLS termination."},
    "ListenerSet": {"label": "ListenerSet", "description": "Gateway supports ListenerSet."},
    "HTTPRouteHostRewrite": {"label": "HTTPRoute Host Rewrite", "description": "Modify the host header of requests forwarded to backends."},
    "HTTPRoutePathRedirect": {"label": "HTTPRoute Path Redirect", "description": "Issue HTTP redirects (e.g. 302) with a modified path."},
    "HTTPRouteRequestMirror": {"label": "HTTPRoute Request Mirror", "description": "Mirror a proportion of traffic to another backend."},
    "HTTPRouteRequestPercentageMirror": {"label": "HTTPRoute Request Percentage Mirror", "description": "Mirror a percentage of requests to another backend."},
    "HTTPRouteResponseHeaderModification": {"label": "HTTPRoute Response Header Modification", "description": "Add, set, or remove response headers."},
    "HTTPRouteCORS": {"label": "HTTPRoute CORS", "description": "CORS policy configuration for HTTPRoute."},
    "BackendTLSPolicy": {"label": "Backend TLS Policy", "description": "Configure TLS when connecting to backends."},
    "BackendTLSPolicySANValidation": {"label": "Backend TLS Policy SAN Validation", "description": "Validate backend certificate SANs."},
    "TLSRouteModeMixed": {"label": "TLSRoute Mode Mixed", "description": "Gateway supports both Passthrough and Terminate TLS listeners on the same port."},
    "TLSRouteModeTerminate": {"label": "TLSRoute Mode Terminate", "description": "Gateway supports TLS termination in addition to passthrough routing."},
}

# Categories rendered by the wizard, in display order.
CATEGORY_ORDER = [
    "httpGateway", "httpRoute", "httpBackendTls", "backend",
    "grpc", "tls", "tcp", "udp", "mesh",
]

# Feature id prefix -> category. More specific prefixes must come first.
CATEGORY_PREFIXES = [
    ("httpBackendTls", ("BackendTLS",)),
    ("backend", ("Backend",)),
    ("httpRoute", ("HTTPRoute",)),
    ("grpc", ("GRPCRoute",)),
    ("tls", ("TLSRoute",)),
    ("tcp", ("TCPRoute",)),
    ("udp", ("UDPRoute",)),
    ("mesh", ("Mesh",)),
    ("httpGateway", ("Gateway",)),
]


def humanize_feature_id(feature_id):
    """Split camelCase ids into words, keeping acronyms like GRPC/TLS/SAN intact."""
    label = re.sub(
        r"(?<=[a-z])(?=[0-9])|(?<=[0-9])(?=[A-Z][a-z])"
        r"|(?<=[a-z0-9])(?=[A-Z])|(?<=[A-Z])(?=[A-Z][a-z])",
        " ",
        feature_id,
    )
    # Keep API resource names as single words, matching the existing labels.
    for words, fixed in (
        ("HTTP Route", "HTTPRoute"),
        ("GRPC Route", "GRPCRoute"),
        ("TLS Route", "TLSRoute"),
        ("TCP Route", "TCPRoute"),
        ("UDP Route", "UDPRoute"),
        ("Web Socket", "WebSocket"),
        ("H2 C", "H2C"),
    ):
        label = label.replace(words, fixed)
    return label


def categorize_feature(feature_id):
    for category, prefixes in CATEGORY_PREFIXES:
        if feature_id.startswith(prefixes):
            return category
    # Unknown ids default to the Gateway section so they are still selectable.
    return "httpGateway"


def load_release_features(path):
    """Return (feature ids, experimental id set) from a release-features yaml, standard first, deduped."""
    data = load_yaml(path) or {}
    ids = []
    for key in ("standard", "experimental"):
        for feature_id in data.get(key) or []:
            feature_id = str(feature_id).strip()
            if feature_id and feature_id not in ids:
                ids.append(feature_id)
    experimental = {str(f).strip() for f in data.get("experimental") or []}
    return ids, experimental


def release_feature_files(repo_root):
    """Map (major, minor) -> release-features yaml path."""
    features_dir = os.path.join(repo_root, "conformance", "release-features")
    files = {}
    if not os.path.isdir(features_dir):
        return files
    for name in os.listdir(features_dir):
        m = re.fullmatch(r"v(\d+)\.(\d+)\.yaml", name)
        if m:
            files[(int(m.group(1)), int(m.group(2)))] = os.path.join(features_dir, name)
    return files


def release_features_for_version(version_dir_name, files):
    """Nearest release-features file for a report version dir, else the oldest one."""
    major, minor, _ = parse_version(version_dir_name)
    path = files.get((major, minor))
    if path is None:
        path = files[min(files)]
    return path, *load_release_features(path)


def build_feature_definitions(release_features, seen_features, experimental=None):
    """Definitions for one version: features in the release yaml that reports actually use.
    Experimental-channel features are tagged in their label."""
    seen = set(seen_features)
    experimental = experimental or set()
    definitions = {category: [] for category in CATEGORY_ORDER}
    for feature_id in release_features:
        if feature_id not in seen:
            continue
        override = FEATURE_LABEL_OVERRIDES.get(feature_id, {})
        label = override.get("label") or humanize_feature_id(feature_id)
        if feature_id in experimental:
            label += " (Experimental)"
        description = override.get("description") or f"Support for {label}."
        definitions[categorize_feature(feature_id)].append(
            {"id": feature_id, "label": label, "description": description}
        )
    for features in definitions.values():
        features.sort(key=lambda f: f["id"])
    return definitions


def check_feature_definitions(definitions, release_features_by_version):
    """Fail loudly if a definition was not sourced from that version's release-features yaml."""
    for version, categories in definitions.items():
        allowed = set(release_features_by_version[version])
        for category, features in categories.items():
            for feature in features:
                if feature["id"] not in allowed:
                    raise SystemExit(
                        f"Feature {feature['id']} in {version}/{category} is not listed in "
                        f"the release-features file used for {version}."
                    )


def load_yaml(path):
    with open(path, "r", encoding="utf-8") as f:
        return yaml.safe_load(f)


def parse_version(version_dir_name):
    """Return (major, minor, patch) or (0,0,0) for sorting."""
    m = re.match(r"^v?(\d+)\.(\d+)(?:\.(\d+))?$", version_dir_name)
    if m:
        return (int(m.group(1)), int(m.group(2)), int(m.group(3) or 0))
    return (0, 0, 0)


def get_version_dirs(reports_root):
    """Return sorted list of version directory paths (e.g. .../v1.6, v1.5, ...), newest first.
    New versions added under conformance/reports/ automatically appear at the top of the list.
    Symlinked aliases (e.g. v1.6.0 -> v1.6) are deduplicated; the real directory name wins."""
    if not os.path.isdir(reports_root):
        return []
    candidates = [
        d
        for d in os.listdir(reports_root)
        if os.path.isdir(os.path.join(reports_root, d)) and parse_version(d) != (0, 0, 0)
    ]
    by_real = {}
    for d in candidates:
        real = os.path.realpath(os.path.join(reports_root, d))
        by_real.setdefault(real, []).append(d)
    kept = []
    for names in by_real.values():
        real_names = [n for n in names if not os.path.islink(os.path.join(reports_root, n))]
        kept.append(sorted(real_names or names, key=parse_version)[-1])
    dirs = [os.path.join(reports_root, d) for d in kept]
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
    Walk all YAML reports under version_dir and aggregate by (organization, project, version).
    Multiple reports for the same impl (e.g. different modes) are merged into one row.
    Returns (list of implementation dicts, union of all extended features seen).
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
    features_seen = set()
    for v in seen.values():
        v["conformance"] = sorted(v["conformance"])
        v["features"] = sorted(v["features"])
        features_seen |= set(v["features"])
        out.append(v)
    out.sort(key=lambda x: (x["organization"], x["project"], x["version"]))
    return out, features_seen


def generate_all(repo_root, output_path):
    """
    Generate controller-wizard-data.json with all versions (--all behavior).
    Versions are emitted newest-first so new releases (e.g. v1.5.0) appear at the top of the wizard dropdown.
    Feature definitions are built per version from conformance/release-features/.
    repo_root: repository root (parent of conformance/, site-src/).
    output_path: full path to output JSON file.
    Returns True if successful, False if no version dirs (caller may exit or log).
    """
    reports_root = os.path.join(repo_root, "conformance", "reports")
    version_dirs = get_version_dirs(reports_root)
    if not version_dirs:
        return False
    release_files = release_feature_files(repo_root)
    if not release_files:
        raise SystemExit("No conformance/release-features/vX.Y.yaml files found.")
    definitions = {}
    release_features_by_version = {}
    implementations = {}
    for version_dir in version_dirs:
        v = os.path.basename(version_dir)
        impls, features_seen = aggregate_by_impl(version_dir)
        if not impls:
            continue  # skip empty placeholder dirs (e.g. pre-release v1.7)
        _path, release_features, experimental = release_features_for_version(v, release_files)
        definitions[v] = build_feature_definitions(release_features, features_seen, experimental)
        release_features_by_version[v] = release_features
        implementations[v] = impls
    if not implementations:
        return False
    check_feature_definitions(definitions, release_features_by_version)
    out = {"featureDefinitions": definitions}
    out.update(implementations)
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=2)
    return True


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
        metavar="vX.Y",
        help="Output a single array for this version (e.g. v1.4).",
    )
    parser.add_argument(
        "-o",
        "--output",
        default="site/static/wizard/data/controller-wizard-data.json",
        help="Output JSON path (default: site/static/wizard/data/controller-wizard-data.json)",
    )
    args = parser.parse_args()

    repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    reports_root = os.path.join(repo_root, "conformance", "reports")
    output = os.path.join(repo_root, args.output)

    if args.all:
        if not generate_all(repo_root, output):
            print("No version dirs under conformance/reports/", file=sys.stderr)
            sys.exit(1)
        with open(output, encoding="utf-8") as f:
            out = json.load(f)
        version_keys = [k for k in out if k != "featureDefinitions"]
        total = sum(len(out[v]) for v in version_keys)
        print(f"Wrote {output} with featureDefinitions and {len(version_keys)} version(s), {total} implementation(s).")
        return
    elif args.version:
        version_dir = os.path.join(reports_root, args.version)
        if not os.path.isdir(version_dir):
            print(f"Version directory not found: {version_dir}", file=sys.stderr)
            sys.exit(1)
        release_files = release_feature_files(repo_root)
        if not release_files:
            print("No conformance/release-features/vX.Y.yaml files found.", file=sys.stderr)
            sys.exit(1)
        impls, features_seen = aggregate_by_impl(version_dir)
        _path, release_features, experimental = release_features_for_version(args.version, release_files)
        definitions = {args.version: build_feature_definitions(release_features, features_seen, experimental)}
        check_feature_definitions(definitions, {args.version: release_features})
        out = {"featureDefinitions": definitions, "implementations": impls}
    else:
        parser.error("Specify --all or --version vX.Y")

    os.makedirs(os.path.dirname(output), exist_ok=True)
    with open(output, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=2)

    print(f"Wrote {output} with featureDefinitions and {len(out['implementations'])} implementation(s).")


def on_config(config, **kwargs):
    """MkDocs hook: generate wizard data with --all before build."""
    repo_root = os.path.dirname(config.config_file_path)
    output_path = os.path.join(repo_root, "site", "static", "wizard", "data", "controller-wizard-data.json")
    if not generate_all(repo_root, output_path):
        log.warning("Controller wizard data not generated: no version dirs under conformance/reports/")
    else:
        log.info("Generated controller wizard data (all versions) at %s", output_path)
    return config


if __name__ == "__main__":
    main()
