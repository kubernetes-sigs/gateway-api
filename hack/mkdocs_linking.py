# Copyright 2025 The Kubernetes Authors.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


"""
A CLI wrapper for documentation tools.

The documentation logic is split into four focused modules:

mkdocs_utils.py
: The core engine of the documentation tools. Contains all logic for ID generation, frontmatter parsing, link conversion, and YAML updates.

mkdocs_hooks.py
: Dedicated module for MkDocs build lifecycle hooks, now purely dependent on mkdocs_utils for logic.

mkdocs_main.py
: The macros plugin, simplified to use the centralized ID map scanner.

mkdocs_linking.py
: A pure CLI wrapper. It contains no core logic, only command-line argument parsing and invocation of the functions in mkdocs_utils.
"""

import argparse
from pathlib import Path
import mkdocs_utils

# --- Configuration (Exposed for test compatibility) ---
DOCS_DIR = mkdocs_utils.DOCS_DIR
REDIRECT_MAP_FILE = mkdocs_utils.REDIRECT_MAP_FILE
FRONTMATTER_ID_KEY = mkdocs_utils.FRONTMATTER_ID_KEY


from mkdocs_utils import (
    prepare_docs,
    convert_internal_links,
)


def main() -> None:
    """Parses command line arguments and runs the preparation script."""
    parser = argparse.ArgumentParser(
        description="MkDocs migration helper - prepares docs for safe refactoring.",
        prog="linking",
    )
    parser.add_argument(
        "--prepare",
        action="store_true",
        help="Scan docs folder, inject IDs, and create redirect map.",
    )
    parser.add_argument(
        "--convert-links",
        action="store_true",
        help="Convert all relative Markdown links to the internal_link macro.",
    )
    parser.add_argument(
        "--docs-dir", default="docs", help="Documentation directory (default: docs)."
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without making changes.",
    )
    args = parser.parse_args()

    docs_dir = Path(args.docs_dir)
    if not docs_dir.exists():
        print(f"ERROR: Documentation directory '{docs_dir}' does not exist.")
        return

    if args.prepare:
        prepare_docs(docs_dir, dry_run=args.dry_run)
    elif args.convert_links:
        if args.dry_run:
            print(
                "DRY RUN for link conversion is not implemented. This action directly modifies files."
            )
        else:
            convert_internal_links(docs_dir)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
