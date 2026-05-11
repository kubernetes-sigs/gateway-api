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

sys.path.insert(0, str(Path(__file__).parents[2]))

from mkdocs_utils import update_mkdocs_yml_redirects

class TestYAMLFormattingPreservation(unittest.TestCase):
    def setUp(self):
        self.test_dir = Path("./temp_test_formatting")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.test_dir.mkdir(parents=True)
        self.mkdocs_yml = self.test_dir / "mkdocs.yml"

    def tearDown(self):
        shutil.rmtree(self.test_dir)

    def test_update_preserves_comments_and_tags(self):
        """
        Verify that updating the redirects doesn't strip comments or 
        !!python/name tags from the mkdocs.yml file.
        """
        original_content = """# Main Configuration
site_name: Gateway API

markdown_extensions:
  - pymdownx.emoji:
      # These tags are often lost by standard YAML dumpers
      emoji_index: !!python/name:material.extensions.emoji.twemoji
      emoji_generator: !!python/name:material.extensions.emoji.to_svg

plugins:
  - search
  - redirects:
      redirect_maps:
        old.md: new.md
"""
        self.mkdocs_yml.write_text(original_content)

        updates = {"moved.md": "destination.md"}
        
        # Act
        success = update_mkdocs_yml_redirects(updates, mkdocs_yml_path_input=self.mkdocs_yml)
        self.assertTrue(success)

        # Assert
        new_content = self.mkdocs_yml.read_text()
        
        # Check for original comment survivability
        self.assertIn("# Main Configuration", new_content, "Main comment was lost")
        self.assertIn("# These tags are often lost", new_content, "Inline comment was lost")
        
        # Check for tag survivability
        self.assertIn("!!python/name:material.extensions.emoji.twemoji", new_content, 
                     "!!python/name tag was lost or mangled")
        
        # Verify the new redirect was actually added
        self.assertIn("moved.md: destination.md", new_content, "New redirect was not added")

if __name__ == "__main__":
    unittest.main()
