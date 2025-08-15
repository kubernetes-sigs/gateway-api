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
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

import linking
from linking import prepare_docs


class TestYAMLFrontmatterEdgeCases(unittest.TestCase):
    """Tests focused on YAML frontmatter parsing edge cases."""

    def setUp(self) -> None:
        """Set up test environment."""
        self.test_dir = Path("./temp_test_yaml")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

        self.original_docs_dir = linking.DOCS_DIR
        self.original_redirect_file = linking.REDIRECT_MAP_FILE

        linking.DOCS_DIR = self.test_dir / "docs"
        linking.REDIRECT_MAP_FILE = self.test_dir / "redirect_map.json"
        linking.DOCS_DIR.mkdir(parents=True, exist_ok=True)

    def tearDown(self) -> None:
        """Clean up test environment."""
        linking.DOCS_DIR = self.original_docs_dir
        linking.REDIRECT_MAP_FILE = self.original_redirect_file
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_malformed_yaml_frontmatter(self) -> None:
        """Test handling of various malformed YAML frontmatter."""
        malformed_files = {
            "unclosed-quotes.md": """---
title: "This quote is never closed
id: broken-yaml
---
# Content""",
            "invalid-structure.md": """---
title: Valid Title
invalid-yaml: [unclosed list
id: another-broken
---
# Content""",
            "wrong-delimiters.md": """+++
title: Hugo-style frontmatter
id: wrong-format
+++
# Content""",
            "no-end-delimiter.md": """---
title: Missing end delimiter
id: incomplete
# This should be treated as content""",
        }

        for filename, content in malformed_files.items():
            (linking.DOCS_DIR / filename).write_text(content)

        # Should handle malformed YAML gracefully
        prepare_docs()

        # Check that redirect map was created (files with valid structure should work)
        self.assertTrue(linking.REDIRECT_MAP_FILE.exists())
        redirect_map = json.loads(linking.REDIRECT_MAP_FILE.read_text())

        # Files with malformed YAML should get auto-generated IDs or extracted IDs
        self.assertIn("broken-yaml", redirect_map)  # From unclosed-quotes.md
        self.assertIn("another-broken", redirect_map)  # From invalid-structure.md
        self.assertIn("wrong-delimiters", redirect_map)
        self.assertIn("no-end-delimiter", redirect_map)

    def test_complex_yaml_structures(self) -> None:
        """Test handling of complex YAML structures in frontmatter."""
        complex_content = """---
title: "Complex YAML Test"
id: complex-yaml
tags:
  - testing
  - yaml
  - complex
metadata:
  author: 
    name: "John Doe"
    email: "john@example.com"
  created: 2023-01-01
  updated: 2023-12-31
  nested:
    deeply:
      very: "deep value"
categories: ["cat1", "cat2", "cat3"]
boolean_value: true
null_value: null
number_value: 42
float_value: 3.14
multiline: |
  This is a multiline
  string that spans
  multiple lines
---
# Complex YAML Test"""

        (linking.DOCS_DIR / "complex.md").write_text(complex_content)
        prepare_docs()

        # Should preserve all the complex YAML structure
        updated_content = (linking.DOCS_DIR / "complex.md").read_text()

        # Verify the ID was preserved and complex structure remains
        self.assertIn("id: complex-yaml", updated_content)
        self.assertIn("deeply:", updated_content)
        self.assertIn("multiline: |", updated_content)

        # Verify redirect map was created correctly
        redirect_map = json.loads(linking.REDIRECT_MAP_FILE.read_text())
        self.assertEqual(redirect_map["complex-yaml"], "complex.md")

    def test_unicode_in_yaml_frontmatter(self) -> None:
        """Test handling of Unicode characters in YAML frontmatter."""
        unicode_content = """---
title: "测试文档 🚀 Café"
id: unicode-test
description: "This contains émojis 🎉 and ñoñ-ASCII çhars"
author: "José García-Martínez"
tags: ["日本語", "español", "français"]
---
# Unicode Test Document"""

        (linking.DOCS_DIR / "unicode.md").write_text(unicode_content, encoding="utf-8")
        prepare_docs()

        # Should handle Unicode correctly
        updated_content = (linking.DOCS_DIR / "unicode.md").read_text(encoding="utf-8")
        self.assertIn("🚀", updated_content)
        self.assertIn("José García-Martínez", updated_content)
        self.assertIn("日本語", updated_content)

        redirect_map = json.loads(linking.REDIRECT_MAP_FILE.read_text())
        self.assertEqual(redirect_map["unicode-test"], "unicode.md")


if __name__ == "__main__":
    unittest.main(verbosity=2)
