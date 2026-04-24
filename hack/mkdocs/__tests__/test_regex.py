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

import unittest
from pathlib import Path
import shutil
import sys
import os

sys.path.insert(0, str(Path(__file__).parents[2]))

import mkdocs_utils

class TestLinkRegexAdversarial(unittest.TestCase):
    def setUp(self):
        self.test_dir = Path("./temp_test_regex")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.test_dir.mkdir(parents=True)
        self.docs_dir = self.test_dir / "site-src"
        self.docs_dir.mkdir(parents=True)
        
        # Patch globals
        self.original_docs_dir = mkdocs_utils.DOCS_DIR
        mkdocs_utils.DOCS_DIR = self.docs_dir

    def tearDown(self):
        mkdocs_utils.DOCS_DIR = self.original_docs_dir
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_link_regex_stress_scenarios(self):
        """
        Test the link conversion regex against complex or adversarial Markdown content.
        """
        # 1. Setup target file that will have an ID
        (self.docs_dir / "target.md").write_text("---\nid: target-id\n---\n# Target")
        
        # 2. Define adversarial scenarios in a source file
        scenarios = {
            "standard": "[Link](target.md)",
            "with_parens_in_title": "[Link (with parens)](target.md)",
            "with_brackets_in_title": "[Link [with brackets]](target.md)",
            "nested_brackets": "[Link [nested [brackets]]](target.md)", # Regex might fail here
            "with_anchor": "[Link](target.md#anchor)",
            "already_converted": '[Already]({{ internal_link("target-id") }})',
            "multiple_on_one_line": "[One](target.md) and [Two](target.md)",
            "with_spaces_in_path": "[Spaces](target file with spaces.md)",
            "inside_code_inline": "See `[code](target.md)` which shouldn't be converted", # Current regex might fail
            "inside_code_block": "```\n[block](target.md)\n```", # Current regex might fail
            "image_link": "![Alt](target.md)", # Should images be converted?
        }
        
        # Create a file for each scenario or one big file
        # For simplicity, we create one big file with all scenarios
        source_content = "\n".join([f"## {k}\n{v}" for k, v in scenarios.items()])
        source_file = self.docs_dir / "source.md"
        source_file.write_text(source_content)
        
        # Also create the file with spaces requested by one scenario
        (self.docs_dir / "target file with spaces.md").write_text("---\nid: spaces-id\n---\n# Spaces")

        # Act
        mkdocs_utils.convert_internal_links(docs_dir_input=self.docs_dir)

        # Assert
        converted_content = source_file.read_text()
        
        # 1. Standard should be converted
        self.assertIn('({{ internal_link("target-id") }})', converted_content)
        
        # 2. Parens in title should be handled (if regex is robust)
        self.assertIn('[Link (with parens)]({{ internal_link("target-id") }})', converted_content)
        
        # 3. Already converted should stay same
        self.assertIn('[Already]({{ internal_link("target-id") }})', converted_content)
        
        # 4. Multiple on one line
        self.assertEqual(converted_content.count('internal_link("target-id")'), 6) # standard + multiple (2) + parens + anchor + already_converted
        
        # 5. Spaces - Check if we handle them
        self.assertIn('({{ internal_link("spaces-id") }})', converted_content)

        # 6. Anchors - Check if we preserve them (Wait, current macro might not take anchors)
        # We should check if the anchor is still there or if we lost it.
        # Actually, internal_link macro only takes the ID. If we want anchors, we need ID#anchor.
        
if __name__ == "__main__":
    unittest.main()
