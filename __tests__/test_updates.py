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
import yaml
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent))

import linking

class TestUpdateMkdocsYml(unittest.TestCase):
    """Tests for the _update_mkdocs_yml_redirects function."""

    def setUp(self) -> None:
        """Set up a temporary directory for each test."""
        self.test_dir = Path("./temp_test_yml_updates")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.test_dir.mkdir()
        self.mkdocs_yml_path = self.test_dir / "mkdocs.yml"
        # Patch the function to use our temporary file path
        self.patcher = patch("linking.Path")
        self.mock_path = self.patcher.start()
        self.mock_path.return_value = self.mkdocs_yml_path

    def tearDown(self) -> None:
        """Clean up the temporary directory after each test."""
        shutil.rmtree(self.test_dir)
        self.patcher.stop()

    def test_updates_yml_with_no_plugins_section(self):
        """Test that the function adds plugins and redirects correctly."""
        # Arrange
        self.mkdocs_yml_path.write_text("site_name: My Docs")
        redirects = {"old/path.md": "new/path.md"}

        # Act
        result = linking._update_mkdocs_yml_redirects(redirects)
        
        # Assert
        self.assertTrue(result)
        with open(self.mkdocs_yml_path, "r") as f:
            config = yaml.safe_load(f)
        self.assertIn("plugins", config)
        self.assertIn({"redirects": {"redirect_maps": redirects}}, config["plugins"])

    def test_updates_yml_with_string_redirects_plugin(self):
        """Test updating when 'redirects' is just a string in the plugins list."""
        # Arrange
        self.mkdocs_yml_path.write_text("plugins:\n  - redirects")
        redirects = {"old/path.md": "new/path.md"}
        
        # Act
        result = linking._update_mkdocs_yml_redirects(redirects)
        
        # Assert
        self.assertTrue(result)
        with open(self.mkdocs_yml_path, "r") as f:
            config = yaml.safe_load(f)
        self.assertIn({"redirects": {"redirect_maps": redirects}}, config["plugins"])
        self.assertNotIn("redirects", config["plugins"])

    def test_updates_yml_with_null_redirect_maps(self):
        """Test handling of 'redirect_maps: null'."""
        # Arrange
        self.mkdocs_yml_path.write_text("plugins:\n  - redirects:\n      redirect_maps:")
        redirects = {"old/path.md": "new/path.md"}
        
        # Act
        result = linking._update_mkdocs_yml_redirects(redirects)
        
        # Assert
        self.assertTrue(result)
        with open(self.mkdocs_yml_path, "r") as f:
            config = yaml.safe_load(f)
        self.assertEqual(config["plugins"][0]["redirects"]["redirect_maps"], redirects)

    def test_does_not_write_if_no_changes_needed(self):
        """Test that the file is not modified if redirects are already present."""
        # Arrange
        redirects = {"old/path.md": "new/path.md"}
        config_dict = {
            "plugins": [{"redirects": {"redirect_maps": redirects}}]
        }
        self.mkdocs_yml_path.write_text(yaml.dump(config_dict))
        initial_mtime = self.mkdocs_yml_path.stat().st_mtime
        
        # Act
        result = linking._update_mkdocs_yml_redirects(redirects)
        
        # Assert
        self.assertTrue(result)
        final_mtime = self.mkdocs_yml_path.stat().st_mtime
        self.assertEqual(initial_mtime, final_mtime)
