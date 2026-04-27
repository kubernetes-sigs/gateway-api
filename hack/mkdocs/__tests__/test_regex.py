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
            "reference_style": "[ref]: target.md",
            "reference_with_anchor": "[ref-anchor]: target.md#anchor",
            "frontmatter_masking": "---\ndescription: '[link](target.md)'\n---\nBody link: [body](target.md)",
        }
        
        # Create a file for each scenario or one big file
        # For simplicity, we create one big file with all scenarios
        source_content = "\n".join([f"## {k}\n{v}" for k, v in scenarios.items()])
        source_file = self.docs_dir / "source.md"
        source_file.write_text(source_content)
        
        # Also create the file with spaces requested by one scenario
        (self.docs_dir / "target file with spaces.md").write_text("---\nid: spaces-id\n---\n# Spaces")
        
        # Frontmatter masking test in a separate file (it must start at beginning)
        fm_file = self.docs_dir / "fm_mask.md"
        fm_file.write_text("---\ndescription: '[link](target.md)'\n---\nBody link: [body](target.md)")

        # Act
        mkdocs_utils.convert_internal_links(docs_dir_input=self.docs_dir)

        # Assert
        converted_content = source_file.read_text()
        fm_converted = fm_file.read_text()
        
        # 1. Standard should be converted
        self.assertIn('({{ internal_link("target-id") }})', converted_content)
        
        # 2. Parens in title should be handled
        self.assertIn('[Link (with parens)]({{ internal_link("target-id") }})', converted_content)
        
        # 3. Already converted should stay same
        self.assertIn('[Already]({{ internal_link("target-id") }})', converted_content)
        
        # 4. Multiple on one line
        self.assertEqual(converted_content.count('internal_link("target-id")'), 10) # 7 + reference_style + reference_with_anchor + body (frontmatter is masked)
        
        # 5. Spaces - Check if we handle them
        self.assertIn('({{ internal_link("spaces-id") }})', converted_content)

        # 6. Anchors in reference links
        self.assertIn('[ref-anchor]: {{ internal_link("target-id") }}#anchor', converted_content)
        self.assertIn('[ref]: {{ internal_link("target-id") }}', converted_content)

        # 7. Code masking
        self.assertIn("`[code](target.md)`", converted_content)
        self.assertIn("```\n[block](target.md)\n```", converted_content)

        # 8. Frontmatter masking
        # The link inside the frontmatter should NOT be converted
        self.assertIn("description: '[link](target.md)'", fm_converted)
        # The link in the body should be converted
        self.assertIn('[body]({{ internal_link("target-id") }})', fm_converted)
        
if __name__ == "__main__":
    unittest.main()
