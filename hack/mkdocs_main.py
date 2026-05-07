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
Main module for mkdocs-macros plugin.
This provides the internal_link macro for resilient documentation linking.
"""

import os
import sys
import logging
from pathlib import Path
from typing import Dict, Optional

# Ensure the directory containing this script is in the Python path
# so that internal imports like 'mkdocs_utils' work correctly.
sys.path.append(os.path.dirname(__file__))

from mkdocs_utils import (
    DOCS_DIR,
    build_id_map,
)

logger = logging.getLogger("mkdocs.macros.internal_link")


class PageResolver:
    """Handles page ID resolution and link generation."""

    def __init__(self, docs_dir: Path = DOCS_DIR):
        self.docs_dir = docs_dir
        self._page_cache: Optional[Dict[str, Path]] = None

    def resolve_page_link(
        self, page_id: str, current_page_path: Optional[str] = None
    ) -> str:
        """Resolve a page ID to its Markdown file reference, relative to current page."""
        import os

        # Build cache on first use
        if self._page_cache is None:
            self._page_cache = build_id_map(self.docs_dir)

        file_path = self._page_cache.get(page_id)
        if not file_path:
            raise ValueError(f"Page with ID '{page_id}' not found")

        target_path = file_path.relative_to(self.docs_dir)

        # If no current page context, return absolute path from docs root
        if not current_page_path:
            return str(target_path)

        # Calculate relative path from current page to target page
        current_path = Path(current_page_path)
        current_dir = current_path.parent
        rel_path = os.path.relpath(str(target_path), str(current_dir))
        return rel_path.replace(os.path.sep, "/")


def define_env(env):
    """Hook for mkdocs-macros plugin functions and variables."""

    resolver = PageResolver()

    @env.macro
    def internal_link(page_id: str) -> str:
        """
        Looks up a page by ID and returns its Markdown file reference.
        This provides resilient linking that survives file moves.
        """
        try:
            # Get current page context from mkdocs-macros environment
            current_page_path = None
            if hasattr(env, "variables") and env.variables:
                page = env.variables.get("page")
                if page and hasattr(page, "file") and hasattr(page.file, "src_path"):
                    current_page_path = page.file.src_path

            return resolver.resolve_page_link(page_id, current_page_path)
        except Exception as e:
            # Fallback: try resolving without context
            try:
                return resolver.resolve_page_link(page_id, None)
            except ValueError:
                logger.error(f"Internal link macro error: Page with ID '{page_id}' not found.")
                return f"[LINK ERROR: Page '{page_id}' not found]"
            except Exception as e_inner:
                logger.error(f"Internal link macro unexpected error: {e_inner}")
                return f"[LINK ERROR: {e_inner}]"
