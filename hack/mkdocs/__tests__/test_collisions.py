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
import json

sys.path.insert(0, str(Path(__file__).parents[2]))

import mkdocs_utils

class TestIDCollisions(unittest.TestCase):
    def setUp(self):
        self.test_dir = Path("./temp_test_collisions")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.test_dir.mkdir(parents=True)
        self.docs_dir = self.test_dir / "site-src"
        self.docs_dir.mkdir(parents=True)
        self.page_id_map_file = self.test_dir / "page_id_map.json"

        # Patch globals
        self.original_docs_dir = mkdocs_utils.DOCS_DIR
        self.original_id_map = mkdocs_utils.PAGE_ID_MAP_FILE
        mkdocs_utils.DOCS_DIR = self.docs_dir
        mkdocs_utils.PAGE_ID_MAP_FILE = self.page_id_map_file

    def tearDown(self):
        mkdocs_utils.DOCS_DIR = self.original_docs_dir
        mkdocs_utils.PAGE_ID_MAP_FILE = self.original_id_map
        shutil.rmtree(self.test_dir)

    def test_id_collision_generation(self):
        """
        Verify what happens when two different files generate the same ID.
        e.g., 'guides/api.md' and 'guides-api.md' both map to 'guides-api'.
        """
        # Create colliding paths
        (self.docs_dir / "guides").mkdir()
        (self.docs_dir / "guides" / "api.md").write_text("# API Guide")
        (self.docs_dir / "guides-api.md").write_text("# Also API Guide")

        # Act
        mkdocs_utils.prepare_docs(dry_run=False)

        # Assert
        id_map = json.loads(self.page_id_map_file.read_text())
        
        # Check if the collision is handled or if one overwrote the other
        # Current logic is expected to overwrite the first with the second
        self.assertIn("guides-api", id_map)
        
        # This test documents CURRENT behavior. In a perfect world, this would
        # warn or error, but currently it's a "last-one-wins" collision.
        winning_path = id_map["guides-api"]
        self.assertTrue(winning_path in ["guides/api.md", "guides-api.md"])

if __name__ == "__main__":
    unittest.main()
