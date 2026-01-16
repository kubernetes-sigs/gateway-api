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

import unittest
from pathlib import Path
import shutil
import sys

sys.path.insert(0, str(Path(__file__).parents[2]))

from mkdocs_main import PageResolver


class TestPageResolver(unittest.TestCase):
    def setUp(self):
        """Set up a temporary directory structure for each test."""
        self.test_dir = Path("./temp_test_main")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.docs_path = self.test_dir / "docs"
        self.docs_path.mkdir(parents=True)

        # Create some test markdown files with frontmatter IDs
        (self.docs_path / "index.md").write_text("---\nid: home\n---\n# Home")
        (self.docs_path / "about.md").write_text("---\nid: about-us\n---\n# About")
        (self.docs_path / "guides").mkdir()
        (self.docs_path / "guides" / "first.md").write_text(
            "---\nid: first-guide\n---\n# First Guide"
        )
        (self.docs_path / "guides" / "second.md").write_text(
            "---\nid: second-guide\n---\n# Second Guide"
        )
        (self.docs_path / "guides" / "subsection").mkdir()
        (self.docs_path / "guides" / "subsection" / "deep.md").write_text(
            "---\nid: deep-page\n---\n# Deep Page"
        )

        self.resolver = PageResolver(docs_dir=self.docs_path)

    def tearDown(self):
        """Clean up the temporary directory after each test."""
        shutil.rmtree(self.test_dir)

    def test_resolve_page_link_no_context(self):
        """Test resolving page links without a current page context."""
        self.assertEqual(self.resolver.resolve_page_link("home"), "index.md")
        self.assertEqual(
            self.resolver.resolve_page_link("first-guide"), "guides/first.md"
        )

    def test_resolve_page_link_from_root(self):
        """Test resolving page links from a page in the docs root."""
        self.assertEqual(
            self.resolver.resolve_page_link("about-us", "index.md"), "about.md"
        )
        self.assertEqual(
            self.resolver.resolve_page_link("first-guide", "index.md"),
            "guides/first.md",
        )

    def test_resolve_page_link_from_subdir(self):
        """Test resolving page links from a page in a subdirectory."""
        self.assertEqual(
            self.resolver.resolve_page_link("home", "guides/first.md"), "../index.md"
        )
        self.assertEqual(
            self.resolver.resolve_page_link("second-guide", "guides/first.md"),
            "second.md",
        )
        self.assertEqual(
            self.resolver.resolve_page_link("about-us", "guides/first.md"),
            "../about.md",
        )
        self.assertEqual(
            self.resolver.resolve_page_link("deep-page", "guides/first.md"),
            "subsection/deep.md",
        )

    def test_resolve_page_link_from_deep_subdir(self):
        """Test resolving page links from a deeply nested page."""
        self.assertEqual(
            self.resolver.resolve_page_link("home", "guides/subsection/deep.md"),
            "../../index.md",
        )
        self.assertEqual(
            self.resolver.resolve_page_link("first-guide", "guides/subsection/deep.md"),
            "../first.md",
        )

    def test_resolve_page_link_not_found(self):
        """Test that resolving a non-existent page ID raises a ValueError."""
        with self.assertRaises(ValueError):
            self.resolver.resolve_page_link("non-existent-id")

    def test_id_changes_are_picked_up(self):
        """Test that the resolver picks up changes to page IDs."""
        self.assertEqual(self.resolver.resolve_page_link("home"), "index.md")

        # Modify the ID in a file
        (self.docs_path / "index.md").write_text("---\nid: new-home\n---\n# Home")

        # Clear the resolver's cache to force it to re-scan the files
        self.resolver._page_cache = None

        # The new ID should now resolve correctly
        self.assertEqual(self.resolver.resolve_page_link("new-home"), "index.md")

        # The old ID should no longer be found
        with self.assertRaises(ValueError):
            self.resolver.resolve_page_link("home")


if __name__ == "__main__":
    unittest.main(verbosity=2)
