# Copyright The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import re
import sys
from pathlib import Path

# Capture readfile shortcode
# E.g. `{{< readfile file="/examples/standard/multicluster/httproute-simple.yaml" code="true" lang="yaml" >}}`
READFILE_RE = re.compile(r"\{\{<\s*readfile\b([\s\S]*?)>\}\}")
# Capture path inside file="..." from readfile shortcode
READFILE_PATH_RE = re.compile(r'\bfile="([^"]+)"')
# Capture reference entry like "#$ - guides/user-guides/tcp.md"
HEADER_REF_RE = re.compile(r"^#\$\s*-\s*(.+)$")

DOC_PATH = [
  Path("site/content/en"),
  Path("geps"),
]

EXAMPLE_PATH = [
  Path("examples"),
]

COMMENT_OUTPUT = Path("example-yaml-reference-audit-comment.md")

# Verification workflow:
# - Added or updated reference Markdown files trigger a full example YAML check
# - Added or updated example YAML files trigger checks for only those YAML files
# - Without changed-file inputs, the script falls back to checking all example YAML files

def markdown_files():
    files = []
    for path in DOC_PATH:
        if path.exists():
            files.extend(path.rglob("*.md"))

    return files

def example_files():
    files = []
    for path in EXAMPLE_PATH:
        if path.exists():
            files.extend(path.rglob("*.yaml"))
            files.extend(path.rglob("*.yml"))

    return files

def changed_example_files():
    changed_examples = os.environ.get("CHANGED_YAMLS", "").split()

    files = []
    for path in changed_examples:
        path = Path(path)
        if path.exists():
            files.append(path)

    return files

def target_example_files():
    changed_examples = changed_example_files()
    changed_references = os.environ.get("CHANGED_REFERENCES", "")

    if changed_references:
        return example_files()
    if changed_examples:
        return changed_examples
    
    return example_files()

def normalize_ref_path(path):
    normalized_path = Path(path).as_posix().removeprefix("/")
    if normalized_path.startswith("site/content/en/"):
        return normalized_path.removeprefix("site/content/en/")
    
    return normalized_path


def extract_example_paths(content):
    # Use set to avoid duplicate example YAML paths
    paths = set()
    for readfile in READFILE_RE.findall(content):
        files = READFILE_PATH_RE.search(readfile)
        if files:
            file = files.group(1)
            if file.startswith(("/examples/", "examples/")) and file.endswith((".yaml", ".yml")):
                paths.add(file.removeprefix("/"))
    
    return sorted(paths)

# Build reference map
def build_map():
    map = {}
    for file in markdown_files():
        content = file.read_text(encoding="utf-8")
        ref = normalize_ref_path(file)

        for example in extract_example_paths(content):
            if example not in map:
                map[example] = set()
            
            map[example].add(ref)
    
    # Convert internal sets to sorted lists
    sorted_map = {}
    for example, refs in map.items():
        sorted_map[example] = sorted(refs)
    
    return sorted_map

def parse_header(content):
    refs = []
    duplicated_refs = []
    existing_refs = set()
    for line in content.splitlines():
        if not line.startswith("#$"):
            break

        ref = HEADER_REF_RE.match(line)
        if ref:
            path = ref.group(1)

            if path in existing_refs:
                duplicated_refs.append(path)
            else:
                existing_refs.add(path)
                refs.append(path)
        
    return sorted(refs), sorted(set(duplicated_refs))


def main():
    map = build_map()
    missing = []
    invalid = []
    duplicated = []
    unused = []

    for example_file in target_example_files():
        example = Path(example_file).as_posix().removeprefix("/")
        expected = map.get(example, [])
        current, duplicated_refs = parse_header(example_file.read_text(encoding="utf-8"))

        missing_refs = [
            ref for ref in expected if ref not in current
        ]
        invalid_refs = [
            ref for ref in current if ref not in expected
        ]

        if missing_refs:
            missing.append((example, missing_refs))
        
        if invalid_refs:
            invalid.append((example, invalid_refs))

        if duplicated_refs:
            duplicated.append((example, duplicated_refs))
        
        if not expected and not current:
            unused.append(example)
    
    output_lines = ["**Example YAML reference check**", ""]

    if missing:
        # https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#example-of-a-workflow-command
        print("::error::Some example YAML files need new reference entries")

        output_lines.append("**Missing reference entries**")
        for example, refs in missing:
            output_lines.append(f"- `{example}`")
            for ref in refs:
                output_lines.append(f"  - add `{ref}`")
    
    if invalid:
        print("::error::Some example YAML files have stale or incorrect references")
        if missing:
            output_lines.append("")

        output_lines.append("**Stale or incorrect reference entries**")
        for example, refs in invalid:
            # https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-a-notice-message
            print(f"::error file={example}::Reference header has stale or incorrect entries")
            output_lines.append(f"- `{example}`")
            for ref in refs:
                output_lines.append(f"  - remove or update `{ref}`")
    
    if duplicated:
        print("::error::Some example YAML files have duplicated reference entries")
        if missing or invalid:
            output_lines.append("")

        output_lines.append("**Duplicated reference entries**")
        for example, refs in duplicated:
            print(f"::error file={example}::Reference header has duplicated entries")
            output_lines.append(f"- `{example}`")
            for ref in refs:
                output_lines.append(f"  - remove duplicated `{ref}`")
    
    if unused:
        print("::warning::Some example YAML files are not referenced by documentation")
        if missing or invalid or duplicated:
            output_lines.append("")

        output_lines.append("**Unreferenced example YAML files**")
        for example in unused:
            output_lines.append(f"- `{example}`")
            output_lines.append("  - no reference header and no documentation usage found")
    
    if not missing and not invalid and not duplicated and not unused:
        output_lines.append("No example YAML reference issues found.")
    
    comment = "\n".join(output_lines).strip()
    COMMENT_OUTPUT.write_text(comment, encoding="utf-8")

    # Do not fail on unused example files yet
    # Detail: https://github.com/kubernetes-sigs/gateway-api/pull/4840#issuecomment-4469495261
    if missing or invalid or duplicated:
        return 1

    return 0

if __name__ == "__main__":
    sys.exit(main())