#!/bin/bash

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

# Verifies that every GEP index.md begins with a Hugo front matter block
# containing a title field, e.g.:
#
#   ---
#   title: "GEP-1234: Some Title"
#   ---
#
# A markdown H1 heading (# GEP-1234: Some Title) produces an empty title in the
# Docsy search index, so the front matter title is required. See #4868.

set -o errexit
set -o nounset
set -o pipefail

readonly KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..
cd "${KUBE_ROOT}"

# Returns 0 if the file starts with a front matter block containing a title.
function has_frontmatter_title {
  awk '
    NR == 1                       { if ($0 != "---") exit 1; next }
    /^---[[:space:]]*$/           { exit (found ? 0 : 1) }
    /^title:[[:space:]]*[^[:space:]]/ { found = 1 }
    END                           { exit (found ? 0 : 1) }
  ' "$1"
}

failed=()
for f in geps/gep-*/index.md; do
  if ! has_frontmatter_title "$f"; then
    failed+=("$f")
  fi
done

if [[ ${#failed[@]} -gt 0 ]]; then
  echo "The following GEP files are missing a Hugo front matter title block:"
  for f in "${failed[@]}"; do
    echo "  - $f"
  done
  echo ""
  echo "Each geps/gep-*/index.md must begin with:"
  echo "  ---"
  echo '  title: "GEP-XXXX: Title"'
  echo "  ---"
  echo ""
  echo "Do not use a markdown H1 heading for the title (see #4868)."
  exit 1
fi

echo "All GEP index.md files have a front matter title block."

# ex: ts=2 sw=2 et filetype=sh
