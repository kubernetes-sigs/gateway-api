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
Generate site-src/api-types/conditions.md from Gateway API condition types and
reasons defined in the Go API package. Keeps documentation in sync with the API.

Usage:
  python3 hack/mkdocs-generate-conditions-docs.py
  python3 hack/mkdocs-generate-conditions-docs.py -o site-src/api-types/conditions.md

Output path defaults to site-src/api-types/conditions.md.
"""

import argparse
import re
import sys
from pathlib import Path

# (resource_name, display_name)
SECTIONS = [
    ("GatewayClass", "GatewayClass"),
    ("Gateway", "Gateway"),
    ("Listener", "Listener (Gateway status)"),
    ("Policy", "Policy resources (BackendTLSPolicy, BackendTrafficPolicy)"),
    ("ListenerSet", "ListenerSet"),
    ("ListenerEntry", "ListenerEntry (ListenerSet status)"),
    ("Route", "Routes (HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute)"),
    ("Mesh", "Mesh"),
]

def extract_reasons_from_comment(comment: str) -> dict:
    """Parse 'Possible reasons for this condition to be True/False/Unknown' from comment."""
    result = {"True": [], "False": [], "Unknown": []}
    # Match "Possible reasons for this condition to be True are:" (case insensitive)
    # Section continues until next "Possible reasons" or end
    pattern = r"Possible reasons for this condition to be (True|False|Unknown) are:\s*((?:(?!Possible reasons)[\s\S])*?)(?=Possible reasons|$)"
    for m in re.finditer(pattern, comment, re.IGNORECASE | re.DOTALL):
        status = m.group(1).capitalize()
        if status not in ("True", "False"):
            status = "Unknown"
        section = m.group(2)
        reasons = re.findall(r'\*\s*"([^"]+)"', section)
        result[status] = reasons
    return result


def extract_const_value(line: str) -> tuple[str | None, str | None, str | None]:
    """Extract (name, type, value) from const line. Returns (None, None, None) if not a const."""
    # Match: Name Type = "Value" or Name Type = Value
    m = re.match(r'\s+([A-Za-z0-9_]+)\s+([A-Za-z0-9_]+)\s+=\s+"([^"]+)"', line)
    if m:
        return m.group(1), m.group(2), m.group(3)
    return None, None, None


def get_resource_from_type(type_name: str) -> str | None:
    """Map *ConditionType name to resource (e.g., GatewayConditionType -> Gateway)."""
    for prefix in ["GatewayClass", "Gateway", "ListenerSet", "ListenerEntry", "Listener", "Route", "Policy", "Mesh"]:
        if type_name.startswith(prefix) and "ConditionType" in type_name:
            return prefix
    return None


def parse_file(repo_root: Path, rel_path: str, resource_filter: str | None) -> list[dict]:
    """Parse a Go file and extract condition types and reasons."""
    filepath = repo_root / rel_path
    if not filepath.exists():
        return []

    content = filepath.read_text()
    items = []
    lines = content.split("\n")
    i = 0
    current_comment = []
    in_const = False

    while i < len(lines):
        line = lines[i]
        stripped = line.strip()

        if stripped.startswith("const ("):
            in_const = True
            current_comment = []
            i += 1
            continue

        if in_const and stripped == ")":
            in_const = False
            i += 1
            continue

        if in_const:
            if stripped.startswith("//"):
                current_comment.append(stripped[2:].strip())
                i += 1
                continue

            name, type_name, value = extract_const_value(line)
            if name and type_name and value:
                resource = get_resource_from_type(type_name)
                if resource_filter and resource != resource_filter:
                    current_comment = []
                    i += 1
                    continue

                comment_text = " ".join(current_comment)
                if "ConditionType" in type_name:
                    reasons_map = extract_reasons_from_comment(comment_text)
                    experimental = "<gateway:experimental>" in comment_text
                    deprecated = "Deprecated:" in comment_text or "deprecated" in comment_text.lower()
                    reserved = "reserved for future" in comment_text.lower()

                    items.append({
                        "kind": "condition",
                        "resource": resource,
                        "name": value,
                        "comment": comment_text,
                        "reasons_map": reasons_map,
                        "experimental": experimental,
                        "deprecated": deprecated,
                        "reserved": reserved,
                    })

            current_comment = []
        i += 1

    return items


def discover_go_files(repo_root: Path) -> list[str]:
    roots = [
        repo_root / "apis" / "v1",
        repo_root / "apisx" / "v1alpha1",
    ]
    files: list[str] = []
    for root in roots:
        if not root.exists():
            continue
        for p in root.rglob("*.go"):
            if p.name.endswith("_test.go"):
                continue
            files.append(str(p.relative_to(repo_root)))
    return sorted(files)


def build_reason_map(items: list[dict]) -> tuple[dict, dict]:
    """Build resource -> conditions, and resource -> condition -> reasons with status."""
    by_resource = {}
    condition_reasons = {}

    for item in items:
        r = item.get("resource")
        if not r or item["kind"] != "condition":
            continue
        if r not in by_resource:
            by_resource[r] = []
            condition_reasons[r] = {}

        by_resource[r].append(item)
        condition_reasons[r][item["name"]] = {}
        for status, reasons in item["reasons_map"].items():
            for rv in reasons:
                if rv not in condition_reasons[r][item["name"]]:
                    condition_reasons[r][item["name"]][rv] = []
                if status not in condition_reasons[r][item["name"]][rv]:
                    condition_reasons[r][item["name"]][rv].append(status)

    return by_resource, condition_reasons


def render_markdown(repo_root: Path, output_path: Path) -> None:
    """Generate conditions.md from API sources."""
    all_items = []
    for rel_path in discover_go_files(repo_root):
        items = parse_file(repo_root, rel_path, None)
        all_items.extend(items)

    by_resource, condition_reasons = build_reason_map(all_items)

    out = []
    out.append("# Condition Types and Reasons Reference")
    out.append("")
    out.append("Conditions provide a standardized way for controllers to communicate the status of resources to users. Each condition has a `type`, `status` (True, False, or Unknown), `reason`, and `message`.")
    out.append("")
    out.append("For an introduction to conditions and troubleshooting guidance, see [Troubleshooting and Status](../concepts/troubleshooting.md).")
    out.append("")
    out.append("## Common Conditions")
    out.append("")
    out.append("The following condition types are used across multiple Gateway API resources:")
    out.append("")
    out.append("| Condition | Description |")
    out.append("|-----------|-------------|")
    out.append("| **Accepted** | True when the object is semantically and syntactically valid, will produce some configuration in the underlying data plane, and has been accepted by a controller. |")
    out.append("| **Programmed** | True when an object's configuration has been fully parsed and successfully sent to the data plane. It will be ready \"soon\"—the exact definition depends on the implementation. |")
    out.append("| **ResolvedRefs** | True when all references to other objects (e.g., Secrets, Services) are valid—the objects exist and each reference is valid for the field where it is used. |")
    out.append("")
    out.append("---")
    out.append("")

    for resource, display_name in SECTIONS:
        if resource not in by_resource:
            continue
        conditions = [c for c in by_resource[resource] if c["kind"] == "condition"]
        if not conditions:
            continue

        out.append(f"## {display_name}")
        out.append("")

        if resource == "Mesh":
            out.append('??? experimental "Experimental"')
            out.append("")
            out.append("    See [GEP-3949](../geps/gep-3949/index.md).")
            out.append("")
        elif "Listener (Gateway status)" in display_name:
            out.append("Listeners are defined in `Gateway.spec.listeners`. Their status appears in `Gateway.status.listeners[].conditions`.")
            out.append("")
        elif "ListenerEntry" in display_name:
            out.append("ListenerEntries are defined in `ListenerSet.spec.listeners`. Their status appears in `ListenerSet.status.listeners[].conditions`. ListenerEntries represent listeners from both the Gateway and attached ListenerSets.")
            out.append("")
        elif "Routes" in display_name:
            out.append("Routes share the same condition types. Status appears in `Route.status.parents[].conditions` (per parent) and `Route.status.conditions` (route-level).")
            out.append("")

        for cond in conditions:
            name = cond["name"]
            out.append(f"### {name}")
            out.append("")

            if cond.get("experimental"):
                out.append('??? experimental "Experimental"')
                out.append("")
                out.append("    " + cond["comment"].split("\n")[0][:200])
                out.append("")
            elif cond.get("reserved"):
                out.append('!!! warning "Reserved for future use"')
                out.append("")
                out.append("    Not used by implementations. If used in the future, will represent the final state where all configuration is confirmed good and has completely propagated to the data plane.")
                out.append("")
            elif cond.get("deprecated"):
                out.append('!!! warning "Deprecated"')
                out.append("")
                out.append("    Use Accepted instead.")
                out.append("")

            reasons = condition_reasons.get(resource, {}).get(name, {})
            if reasons:
                out.append('<div class="conditions-compact-table-wrap" markdown="1">')
                out.append("")
                out.append("| Reason | True | False | Unknown |")
                out.append("| --- | --- | --- | --- |")
                for reason_name, statuses in sorted(reasons.items()):
                    out.append(
                        "| "
                        + " | ".join(
                            [
                                reason_name,
                                "✓" if "True" in statuses else "",
                                "✓" if "False" in statuses else "",
                                "✓" if "Unknown" in statuses else "",
                            ]
                        )
                        + " |"
                    )
                out.append("")
                out.append("</div>")
                out.append("")
            out.append("")

        out.append("---")
        out.append("")

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text("\n".join(out))
    print(f"Generated {output_path}", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(description="Generate conditions.md from API sources")
    parser.add_argument("-o", "--output", default="site-src/api-types/conditions.md", help="Output path")
    args = parser.parse_args()

    repo_root = Path(__file__).resolve().parent.parent
    output_path = repo_root / args.output
    render_markdown(repo_root, output_path)


def on_config(config, **kwargs):
    """MkDocs hook: generate conditions.md before build."""
    repo_root = Path(config.config_file_path).resolve().parent
    output_path = repo_root / "site-src" / "api-types" / "conditions.md"
    render_markdown(repo_root, output_path)
    return config


if __name__ == "__main__":
    main()
