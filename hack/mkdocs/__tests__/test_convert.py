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

import shutil
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parents[2]))

import mkdocs_linking as linking


class TestConvertFromRelativeLinks(unittest.TestCase):
    """Tests for the convert_internal_links function."""

    test_dir: Path
    docs_path: Path
    redirect_map_file: Path
    linking_module: object
    original_globals: dict

    def setUp(self) -> None:
        """Set up a temporary directory structure for each test."""
        self.test_dir = Path("./temp_test_convert_links")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

        self.docs_path = self.test_dir / "docs"
        self.redirect_map_file = self.test_dir / "redirect_map.json"
        self.docs_path.mkdir(parents=True)

        self.linking_module = sys.modules["mkdocs_linking"]
        self.original_globals = {
            "DOCS_DIR": self.linking_module.DOCS_DIR,
            "REDIRECT_MAP_FILE": self.linking_module.REDIRECT_MAP_FILE,
        }
        self.linking_module.DOCS_DIR = self.docs_path  # type: ignore
        self.linking_module.REDIRECT_MAP_FILE = self.redirect_map_file  # type: ignore

    def test_basic_link_conversion(self) -> None:
        """Test that a simple relative link is converted to a macro."""
        # Arrange
        (self.docs_path / "index.md").write_text("Link to [About](./about.md).")
        (self.docs_path / "about.md").write_text("This is the about page.")
        linking.prepare_docs(str(self.docs_path))

        # Act
        linking.convert_internal_links(str(self.docs_path))

        # Assert
        content = (self.docs_path / "index.md").read_text()
        expected = '---\nid: index\n---\nLink to [About]({{ internal_link("about") }}).'
        self.assertEqual(content, expected)

    def test_file_with_no_links(self) -> None:
        """Test that a file with no links is not modified."""
        # Arrange
        original_content = "This document has no links. Just plain text."
        (self.docs_path / "no-links.md").write_text(original_content)
        linking.prepare_docs(str(self.docs_path))

        # Act
        linking.convert_internal_links(str(self.docs_path))

        # Assert
        final_content = (self.docs_path / "no-links.md").read_text()
        expected_content = "---\nid: no-links\n---\n" + original_content
        self.assertEqual(final_content, expected_content)

    def test_handles_complex_relative_paths(self) -> None:
        """Test conversion of links with complex relative paths like ../.."""
        # Arrange
        (self.docs_path / "guides" / "advanced").mkdir(parents=True)
        (self.docs_path / "api" / "v1").mkdir(parents=True)

        (self.docs_path / "guides" / "advanced" / "config.md").write_text(
            "See the [Auth API](../../api/v1/auth.md) for details."
        )
        (self.docs_path / "api" / "v1" / "auth.md").write_text("Auth API docs.")
        linking.prepare_docs(str(self.docs_path))

        # Act
        linking.convert_internal_links(str(self.docs_path))

        # Assert
        content = (self.docs_path / "guides" / "advanced" / "config.md").read_text()
        expected = '---\nid: guides-advanced-config\n---\nSee the [Auth API]({{ internal_link("api-v1-auth") }}) for details.'
        self.assertEqual(content, expected)

    def test_idempotency_does_not_reconvert_links(self) -> None:
        """Test that running the conversion twice doesn't change already converted links."""
        # Arrange
        (self.docs_path / "index.md").write_text("Link to [About](./about.md).")
        (self.docs_path / "about.md").write_text("This is the about page.")
        linking.prepare_docs(str(self.docs_path))

        # Act
        linking.convert_internal_links(str(self.docs_path))  # First run
        content_after_first_run = (self.docs_path / "index.md").read_text()

        linking.convert_internal_links(str(self.docs_path))  # Second run
        content_after_second_run = (self.docs_path / "index.md").read_text()

        # Assert
        expected = '---\nid: index\n---\nLink to [About]({{ internal_link("about") }}).'
        self.assertEqual(content_after_first_run, expected)
        self.assertEqual(
            content_after_second_run,
            expected,
            "Content should not change on the second run.",
        )

    def test_leaves_broken_links_unchanged(self) -> None:
        """Test that a link to a non-existent .md file is not converted."""
        # Arrange
        original_content = "This is a [Broken Link](./nonexistent.md)."
        (self.docs_path / "index.md").write_text(original_content)
        linking.prepare_docs(str(self.docs_path))

        # Act
        linking.convert_internal_links(str(self.docs_path))

        # Assert
        final_content = (self.docs_path / "index.md").read_text()
        expected_content = "---\nid: index\n---\n" + original_content
        self.assertEqual(
            final_content, expected_content, "Broken link should not be modified."
        )


if __name__ == "__main__":
    unittest.main(verbosity=2)
