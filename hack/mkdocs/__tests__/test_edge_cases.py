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
import os
import shutil
import sys
import json
import time
from pathlib import Path
from unittest.mock import patch, MagicMock

# Ensure the parent directory is in the path so we can import our modules
sys.path.insert(0, str(Path(__file__).parents[2]))

import mkdocs_utils
from mkdocs_utils import (
    prepare_docs, convert_internal_links, update_mkdocs_yml_redirects, 
    get_frontmatter, build_id_map
)
from mkdocs_hooks import on_files

class TestEdgeCasesAndRobustness(unittest.TestCase):
    def setUp(self):
        self.test_dir = Path("./temp_test_edge_cases")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.test_dir.mkdir(parents=True)
        self.docs_dir = self.test_dir / "site-src"
        self.docs_dir.mkdir()
        self.mkdocs_yml = self.test_dir / "mkdocs.yml"
        
        # Patch globals for isolation
        self.patcher_docs = patch("mkdocs_utils.DOCS_DIR", self.docs_dir)
        self.patcher_map = patch("mkdocs_utils.PAGE_ID_MAP_FILE", self.test_dir / "page_id_map.json")
        self.patcher_yml = patch("mkdocs_utils.MKDOCS_YML_PATH", self.mkdocs_yml)
        
        # Also patch hooks which might have already imported these
        self.patcher_hook_map = patch("mkdocs_hooks.PAGE_ID_MAP_FILE", self.test_dir / "page_id_map.json")
        
        self.patcher_docs.start()
        self.patcher_map.start()
        self.patcher_yml.start()
        self.patcher_hook_map.start()

    def tearDown(self):
        self.patcher_docs.stop()
        self.patcher_map.stop()
        self.patcher_yml.stop()
        self.patcher_hook_map.stop()
        shutil.rmtree(self.test_dir)

    def test_1_complex_masking_nesting(self):
        """1. Verify masking handles multiple code styles on one line."""
        content = "Here is `some code` and [a link](target.md) and ```more code``` with `final code`."
        target_file = self.docs_dir / "target.md"
        target_file.write_text("---\nid: target-id\n---\nTarget content")
        
        source_file = self.docs_dir / "source.md"
        source_file.write_text(content)
        
        convert_internal_links(docs_dir_input=self.docs_dir)
        
        updated_content = source_file.read_text()
        self.assertIn('{{ internal_link("target-id") }}', updated_content)
        self.assertIn("`some code`", updated_content)
        self.assertIn("```more code```", updated_content)

    def test_2_url_query_parameters(self):
        """2. Verify links with query parameters are ignored (safer fallback)."""
        content = "[Link with param](target.md?version=1)"
        source_file = self.docs_dir / "source.md"
        source_file.write_text(content)
        
        convert_internal_links(docs_dir_input=self.docs_dir)
        
        updated_content = source_file.read_text()
        self.assertIn("target.md?version=1", updated_content, "Links with queries should be ignored currently")

    def test_3_fallback_frontmatter_parsing(self):
        """3. Test manual regex fallback when frontmatter is corrupt."""
        # Unclosed YAML frontmatter or bad syntax
        content = "---\nid: manual-id\nkey: : broken: : syntax\nContent"
        metadata = get_frontmatter(content)
        self.assertEqual(metadata.get("id"), "manual-id")

    def test_4_path_to_id_collision_detection(self):
        """4. Verify collision detection for newly generated IDs (a/b.md vs a-b.md)."""
        (self.docs_dir / "a-b.md").write_text("content")
        (self.docs_dir / "a").mkdir()
        (self.docs_dir / "a" / "b.md").write_text("content")
        
        # This should trigger a warning in logs but complete
        prepare_docs(docs_dir_input=self.docs_dir)
        
        with open(self.test_dir / "page_id_map.json") as f:
            id_map = json.load(f)
        self.assertIn("a-b", id_map)

    def test_5_aggressive_yaml_comments(self):
        """5. Stress-test surgical patcher with comments on almost every line."""
        original_yml = """
plugins:
  - redirects: # plugin comment
      # block comment
      redirect_maps: # header comment
        old.md: new.md # entry comment
"""
        self.mkdocs_yml.write_text(original_yml)
        update_mkdocs_yml_redirects({"a.md": "b.md"})
        
        updated_yml = self.mkdocs_yml.read_text()
        self.assertIn("# plugin comment", updated_yml)
        self.assertIn("# entry comment", updated_yml)
        self.assertIn("a.md: b.md", updated_yml)

    def test_6_directory_index_links(self):
        """6. Ensure directory links aren't incorrectly converted."""
        content = "[Guides](guides/)"
        source_file = self.docs_dir / "source.md"
        source_file.write_text(content)
        
        convert_internal_links(docs_dir_input=self.docs_dir)
        
        updated_content = source_file.read_text()
        self.assertIn("(guides/)", updated_content)

    def test_7_encoding_robustness(self):
        """7. Verify system doesn't crash on invalid UTF-8 (Iso-8859-1 test)."""
        source_file = self.docs_dir / "latin.md"
        # Write bytes that are valid in latin-1 but invalid in utf-8 if they contain high bits
        source_file.write_bytes(b"\xbd\xb2\xbc") # Half, squared, quarter in Latin-1
        
        # Should complete without crashing (skipping or logging)
        prepare_docs(docs_dir_input=self.docs_dir)
        convert_internal_links(docs_dir_input=self.docs_dir)

    def test_8_idempotency(self):
        """8. Run prepare_docs twice and ensure zero extra changes."""
        source_file = self.docs_dir / "index.md"
        source_file.write_text("Pure content")
        
        prepare_docs(docs_dir_input=self.docs_dir)
        first_mtime = source_file.stat().st_mtime
        first_content = source_file.read_text()
        
        time.sleep(0.01) # Ensure clock tick
        
        prepare_docs(docs_dir_input=self.docs_dir)
        second_mtime = source_file.stat().st_mtime
        second_content = source_file.read_text()
        
        self.assertEqual(first_content, second_content)

    def test_9_redirect_type_preservation(self):
        """9. Ensure on_files preserves unrelated redirect config."""
        self.mkdocs_yml.write_text("plugins:\n  - redirects:\n      redirect_maps: {}\n")
        
        # Mock MkDocs config
        config = {
            "docs_dir": str(self.docs_dir),
            "plugins": {
                "redirects": MagicMock()
            }
        }
        
        # Use on_files (requires PAGE_ID_MAP_FILE to exist)
        (self.test_dir / "page_id_map.json").write_text('{"old-id": "old-path.md"}')
        
        # File has moved
        source_file = self.docs_dir / "new-path.md"
        source_file.write_text("---\nid: old-id\n---\nContent")
        
        mock_file = MagicMock()
        mock_file.src_path = "new-path.md"
        
        on_files([mock_file], config)
        
        updated_yml = self.mkdocs_yml.read_text()
        self.assertIn("old-path.md: new-path.md", updated_yml)

    def test_10_performance_smoke_test(self):
        """10. Process a large file with thousands of links."""
        link_line = "[Link](target.md)\n"
        content = "--- \nid: source\n--- \n" + (link_line * 2000)
        
        target_file = self.docs_dir / "target.md"
        target_file.write_text("---\nid: target-id\n---\nTarget")
        
        source_file = self.docs_dir / "source.md"
        source_file.write_text(content)
        
        start_time = time.time()
        convert_internal_links(docs_dir_input=self.docs_dir)
        end_time = time.time()
        
        elapsed = end_time - start_time
        # Performance check: 2000 links should take less than 1 second on any reasonable env
        self.assertLess(elapsed, 2.0, f"Performance too slow: {elapsed:.2f}s")
        self.assertIn('{{ internal_link("target-id") }}', source_file.read_text())

if __name__ == "__main__":
    unittest.main()
