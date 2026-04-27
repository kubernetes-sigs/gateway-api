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

import json
import shutil
import sys
from typing import Any, Dict, List, Tuple
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parents[2]))

import mkdocs_linking as linking
from mkdocs_linking import prepare_docs


class TestMigration(unittest.TestCase):
    """Tests the migration script's prepare function."""

    def setUp(self) -> None:
        """Set up a temporary directory structure for each test."""
        self.test_dir = Path("./temp_test_project")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

        self.docs_path = self.test_dir / "site-src"
        self.page_id_map_file = self.test_dir / "page_id_map.json"
        self.mkdocs_yml_path = self.test_dir / "mkdocs.yml"
        self.docs_path.mkdir(parents=True)

        # Create a mock mkdocs.yml for hooks that update it
        self.mkdocs_yml_path.write_text("plugins:\n  - redirects:\n      redirect_maps:\n")

        # The script under test uses module-level globals for configuration.
        self.utils_module = sys.modules["mkdocs_utils"]
        self.original_globals = {
            "DOCS_DIR": self.utils_module.DOCS_DIR,
            "PAGE_ID_MAP_FILE": self.utils_module.PAGE_ID_MAP_FILE,
            "MKDOCS_YML_PATH": self.utils_module.MKDOCS_YML_PATH,
        }
        self.utils_module.DOCS_DIR = self.docs_path  # type: ignore
        self.utils_module.PAGE_ID_MAP_FILE = self.page_id_map_file  # type: ignore
        self.utils_module.MKDOCS_YML_PATH = self.mkdocs_yml_path  # type: ignore

        # Patch linking for those that import it directly
        if "mkdocs_linking" in sys.modules:
            self.linking_module = sys.modules["mkdocs_linking"]
            self.linking_module.DOCS_DIR = self.docs_path  # type: ignore
            self.linking_module.PAGE_ID_MAP_FILE = self.page_id_map_file  # type: ignore

    def tearDown(self) -> None:
        """Clean up the temporary directory after each test."""
        shutil.rmtree(self.test_dir)
        # Restore the original global variables to avoid side-effects between
        # test runs.
        self.utils_module.DOCS_DIR = self.original_globals["DOCS_DIR"]
        self.utils_module.PAGE_ID_MAP_FILE = self.original_globals["PAGE_ID_MAP_FILE"]
        self.utils_module.MKDOCS_YML_PATH = self.original_globals["MKDOCS_YML_PATH"]

        if "mkdocs_linking" in sys.modules:
            self.linking_module.DOCS_DIR = self.original_globals["DOCS_DIR"]
            self.linking_module.PAGE_ID_MAP_FILE = self.original_globals["PAGE_ID_MAP_FILE"]

    def test_prepare_fresh_run_no_frontmatter(self) -> None:
        """Test that IDs are correctly injected into files."""
        # Arrange: Create a file structure with no existing frontmatter.
        (self.docs_path / "index.md").write_text("Welcome page")
        (self.docs_path / "guides").mkdir()
        (self.docs_path / "guides" / "http.md").write_text("HTTP Guide")

        # Act: Run the preparation function.
        prepare_docs()

        # Assert: Verify the page ID map file was created and is correct.
        self.assertTrue(self.page_id_map_file.exists())
        redirect_map: Dict[str, str] = json.loads(self.page_id_map_file.read_text())
        self.assertEqual(redirect_map.get("index"), "index.md")
        self.assertEqual(redirect_map.get("guides-http"), "guides/http.md")

    def test_prepare_with_existing_frontmatter(self) -> None:
        """Test that existing frontmatter is preserved and IDs are respected."""
        # Arrange: Create files with existing frontmatter
        (self.docs_path / "existing.md").write_text("""---
title: "Existing Document"
author: "John Doe"
id: custom-id
tags: ["important"]
---
# Existing Document
This has frontmatter already.""")

        (self.docs_path / "partial.md").write_text("""---
title: "Partial Frontmatter"
description: "No ID yet"
---
# Partial Document""")

        # Act: Run the preparation function
        prepare_docs()

        # Assert: Check that existing ID is preserved and new ID is added
        redirect_map = json.loads(self.page_id_map_file.read_text())
        self.assertEqual(redirect_map.get("custom-id"), "existing.md")
        self.assertEqual(redirect_map.get("partial"), "partial.md")

        # Verify file contents preserve existing frontmatter
        existing_content = (self.docs_path / "existing.md").read_text()
        self.assertIn('title: "Existing Document"', existing_content)
        self.assertIn('author: "John Doe"', existing_content)
        self.assertIn("id: custom-id", existing_content)
        self.assertIn('tags: ["important"]', existing_content)

        partial_content = (self.docs_path / "partial.md").read_text()
        self.assertIn("title: Partial Frontmatter", partial_content)
        self.assertIn("id: partial", partial_content)

    def test_prepare_multiple_subdirectories(self) -> None:
        """Test ID generation for files in multiple nested directories."""
        # Arrange: Create a complex directory structure
        (self.docs_path / "api" / "v1").mkdir(parents=True)
        (self.docs_path / "api" / "v2").mkdir(parents=True)
        (self.docs_path / "guides" / "getting-started").mkdir(parents=True)
        (self.docs_path / "guides" / "advanced").mkdir(parents=True)

        files_to_create = {
            "api/v1/auth.md": "# Authentication v1",
            "api/v1/users.md": "# Users API v1",
            "api/v2/auth.md": "# Authentication v2",
            "api/v2/users.md": "# Users API v2",
            "guides/getting-started/installation.md": "# Installation",
            "guides/getting-started/quickstart.md": "# Quick Start",
            "guides/advanced/configuration.md": "# Advanced Configuration",
            "guides/advanced/deployment.md": "# Deployment Guide",
        }

        for file_path, content in files_to_create.items():
            (self.docs_path / file_path).write_text(content)

        # Act: Run preparation
        prepare_docs()

        # Assert: Verify all files get appropriate IDs
        redirect_map = json.loads(self.page_id_map_file.read_text())

        expected_mappings = {
            "api-v1-auth": "api/v1/auth.md",
            "api-v1-users": "api/v1/users.md",
            "api-v2-auth": "api/v2/auth.md",
            "api-v2-users": "api/v2/users.md",
            "guides-getting-started-installation": "guides/getting-started/installation.md",
            "guides-getting-started-quickstart": "guides/getting-started/quickstart.md",
            "guides-advanced-configuration": "guides/advanced/configuration.md",
            "guides-advanced-deployment": "guides/advanced/deployment.md",
        }

        for expected_id, expected_path in expected_mappings.items():
            self.assertEqual(redirect_map.get(expected_id), expected_path)

    def test_prepare_handles_empty_docs_directory(self) -> None:
        """Test that prepare_docs handles an empty docs directory gracefully."""
        # Arrange: Docs directory exists but is empty (no .md files)

        # Act: Run preparation
        prepare_docs()

        # Assert: Should create empty redirect map
        self.assertTrue(self.page_id_map_file.exists())
        redirect_map = json.loads(self.page_id_map_file.read_text())
        self.assertEqual(len(redirect_map), 0)


if __name__ == "__main__":
    unittest.main()
