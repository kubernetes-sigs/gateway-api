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
from types import SimpleNamespace

sys.path.insert(0, str(Path(__file__).parent))

import linking
from linking import on_config, prepare_docs


class TestMigration(unittest.TestCase):
    """Tests the migration script's prepare and on_config functions."""

    def setUp(self) -> None:
        """Set up a temporary directory structure for each test."""
        self.test_dir = Path("./temp_test_project")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

        self.docs_path = self.test_dir / "docs"
        self.redirect_map_file = self.test_dir / "redirect_map.json"
        self.docs_path.mkdir(parents=True)

        # The script under test uses module-level globals for configuration.
        # To ensure tests are isolated, we temporarily redirect these globals
        # to point to our test directory during test execution.
        self.linking_module = sys.modules["hack.mkdocs_linking"]
        self.original_globals = {
            "DOCS_DIR": self.linking_module.DOCS_DIR,
            "REDIRECT_MAP_FILE": self.linking_module.REDIRECT_MAP_FILE,
        }
        self.linking_module.DOCS_DIR = self.docs_path  # type: ignore
        self.linking_module.REDIRECT_MAP_FILE = self.redirect_map_file  # type: ignore

    def tearDown(self) -> None:
        """Clean up the temporary directory after each test."""
        shutil.rmtree(self.test_dir)
        # Restore the original global variables to avoid side-effects between
        # test runs.
        for key, value in self.original_globals.items():
            setattr(self.linking_module, key, value)

    def test_prepare_fresh_run_no_frontmatter(self) -> None:
        """Test that IDs are correctly injected into files."""
        # Arrange: Create a file structure with no existing frontmatter.
        (self.docs_path / "index.md").write_text("Welcome page")
        (self.docs_path / "guides").mkdir()
        (self.docs_path / "guides" / "http.md").write_text("HTTP Guide")

        # Act: Run the preparation function.
        prepare_docs()

        # Assert: Verify the redirect map file was created and is correct.
        self.assertTrue(self.redirect_map_file.exists())
        redirect_map: Dict[str, str] = json.loads(self.redirect_map_file.read_text())
        self.assertEqual(redirect_map.get("index"), "index.md")
        self.assertEqual(redirect_map.get("guides-http"), "guides/http.md")

    def _create_mock_config(
        self, pages_data: List[Tuple[str, str, str]]
    ) -> Dict[str, Any]:
        """Create a mock MkDocs config object for testing the hook function."""
        mock_pages: List[SimpleNamespace] = []
        for page_id, src_path, url in pages_data:
            page = SimpleNamespace(file=SimpleNamespace(src_path=src_path), url=url)
            # The hook function reads files to get IDs, so we must create them.
            file_path = self.docs_path / src_path
            file_path.parent.mkdir(parents=True, exist_ok=True)
            file_path.write_text(f"---\nid: {page_id}\n---\nContent")
            mock_pages.append(page)

        return {
            "docs_dir": str(self.docs_path),
            "pages": mock_pages,
            "plugins": {
                "redirects": {"config": {"redirect_maps": {}}},
                "macros": {"config": {"python_macros": {}}},
            },
        }

    def test_on_files_one_file_moved(self) -> None:
        """Test that a redirect is correctly generated for a moved file using on_files."""
        (self.docs_path / "old-path.md").write_text("Content")
        prepare_docs()
        # Simulate a file move by creating a mock files list
        new_file = SimpleNamespace(src_path="new/path/for/doc.md")
        # Write the file with the same ID in the new location
        new_file_path = self.docs_path / "new/path/for/doc.md"
        new_file_path.parent.mkdir(parents=True, exist_ok=True)
        new_file_path.write_text("---\nid: old-path\n---\nContent")
        # Call on_files with the new file list and config
        files = [new_file]
        config = {"docs_dir": str(self.docs_path)}
        # on_files prints output, but we want to check the mkdocs.yml or output
        # For this test, just ensure no exceptions and that the function returns the files
        result = linking.on_files(files, config)
        self.assertEqual(result, files)

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
        redirect_map = json.loads(self.redirect_map_file.read_text())
        self.assertEqual(redirect_map.get("custom-id"), "existing.md")
        self.assertEqual(redirect_map.get("partial"), "partial.md")

        # Verify file contents preserve existing frontmatter
        existing_content = (self.docs_path / "existing.md").read_text()
        self.assertIn('title: "Existing Document"', existing_content)
        self.assertIn('author: "John Doe"', existing_content)
        self.assertIn("id: custom-id", existing_content)
        self.assertIn('tags: ["important"]', existing_content)

        partial_content = (self.docs_path / "partial.md").read_text()
        self.assertIn('title: "Partial Frontmatter"', partial_content)
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
        redirect_map = json.loads(self.redirect_map_file.read_text())

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

    def test_on_config_no_files_moved(self) -> None:
        """Test that no redirects are generated when files haven't moved."""
        # Arrange: Create initial state
        (self.docs_path / "stable.md").write_text(
            "---\nid: stable-doc\n---\nStable content"
        )
        prepare_docs()

        # Arrange: Create config with same paths (no moves)
        mock_config = self._create_mock_config(
            [("stable-doc", "stable.md", "/stable/")]
        )

        # Act: Run the hook
        updated_config = on_config(mock_config)

        # Assert: No redirect rules should be generated
        redirects = updated_config["plugins"]["redirects"]["config"]["redirect_maps"]
        self.assertEqual(len(redirects), 0)

    def test_on_files_multiple_files_moved(self) -> None:
        """Test redirect generation for multiple moved files using on_files."""
        initial_files = {
            "old-guide.md": "old-guide-id",
            "temp/draft.md": "draft-doc",
            "archive/old-api.md": "api-v1",
        }
        for file_path, file_id in initial_files.items():
            file_full_path = self.docs_path / file_path
            file_full_path.parent.mkdir(parents=True, exist_ok=True)
            file_full_path.write_text(f"---\nid: {file_id}\n---\nContent")
        prepare_docs()
        # Simulate all files being moved to new locations
        new_files = [
            SimpleNamespace(src_path="guides/user-guide.md"),
            SimpleNamespace(src_path="published/final-doc.md"),
            SimpleNamespace(src_path="api/legacy/v1.md"),
        ]
        # Write the files with the same IDs in the new locations
        moved = [
            ("guides/user-guide.md", "old-guide-id"),
            ("published/final-doc.md", "draft-doc"),
            ("api/legacy/v1.md", "api-v1"),
        ]
        for path, file_id in moved:
            file_path = self.docs_path / path
            file_path.parent.mkdir(parents=True, exist_ok=True)
            file_path.write_text(f"---\nid: {file_id}\n---\nContent")
        config = {"docs_dir": str(self.docs_path)}
        result = linking.on_files(new_files, config)
        self.assertEqual(result, new_files)

    def test_prepare_handles_empty_docs_directory(self) -> None:
        """Test that prepare_docs handles an empty docs directory gracefully."""
        # Arrange: Docs directory exists but is empty (no .md files)

        # Act: Run preparation
        prepare_docs()

        # Assert: Should create empty redirect map
        self.assertTrue(self.redirect_map_file.exists())
        redirect_map = json.loads(self.redirect_map_file.read_text())
        self.assertEqual(len(redirect_map), 0)

    def test_on_config_missing_redirect_map(self) -> None:
        """Test on_config behavior when redirect map file doesn't exist."""
        # Arrange: Ensure redirect map doesn't exist
        if self.redirect_map_file.exists():
            self.redirect_map_file.unlink()

        mock_config = self._create_mock_config([("test-page", "test.md", "/test/")])

        # Act: Run the hook
        updated_config = on_config(mock_config)

        # Assert: Should handle missing file gracefully and still set up macro
        self.assertIn("macros", updated_config["plugins"])

        # No redirects should be generated
        redirects = updated_config["plugins"]["redirects"]["config"]["redirect_maps"]
        self.assertEqual(len(redirects), 0)

    def test_on_config_missing_plugins(self) -> None:
        """Test on_config behavior when expected plugins are not configured."""
        # Arrange: Create config without redirects or macros plugins
        (self.docs_path / "test.md").write_text("---\nid: test-page\n---\nContent")
        prepare_docs()

        mock_config = {
            "docs_dir": str(self.docs_path),
            "pages": [
                SimpleNamespace(file=SimpleNamespace(src_path="test.md"), url="/test/")
            ],
            "plugins": {},  # No redirects or macros plugins
        }

        # Act: Run the hook
        updated_config = on_config(mock_config)

        # Assert: Should handle missing plugins gracefully
        self.assertIsInstance(updated_config, dict)
        self.assertEqual(updated_config["docs_dir"], str(self.docs_path))
