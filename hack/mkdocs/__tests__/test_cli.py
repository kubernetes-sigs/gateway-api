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


class TestCommandLineInterface(unittest.TestCase):
    docs_path: Path

    def setUp(self) -> None:
        """Set up a temporary directory structure for each test."""
        self.test_dir = Path("./temp_test_convert_links")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

        self.docs_path = self.test_dir / "docs"
        self.redirect_map_file = self.test_dir / "redirect_map.json"
        self.docs_path.mkdir(parents=True)

        self.linking_module = sys.modules["hack.mkdocs_linking"]
        self.original_globals = {
            "DOCS_DIR": self.linking_module.DOCS_DIR,
            "REDIRECT_MAP_FILE": self.linking_module.REDIRECT_MAP_FILE,
        }
        self.linking_module.DOCS_DIR = self.docs_path  # type: ignore
        self.linking_module.REDIRECT_MAP_FILE = self.redirect_map_file  # type: ignore

    def test_main_handles_prepare_docs_exceptions(self) -> None:
        """Test main() handles exceptions from prepare_docs gracefully."""
        # Arrange: Mock prepare_docs to raise an exception
        original_prepare_docs = linking.prepare_docs

        def failing_prepare_docs(docs_dir_path=None):
            raise Exception("Test exception from prepare_docs")

        linking.prepare_docs = failing_prepare_docs

        import sys

        original_argv = sys.argv
        sys.argv = ["linking.py", "--prepare"]

        try:
            # Act & Assert: Exception should propagate (this is expected behavior)
            with self.assertRaises(Exception) as context:
                linking.main()

            self.assertIn("Test exception from prepare_docs", str(context.exception))

        finally:
            # Restore everything
            linking.prepare_docs = original_prepare_docs
            sys.argv = original_argv

    def test_main_with_prepare_argument(self) -> None:
        """Test main() function when called with --prepare argument."""
        # Arrange: Create some test files
        (self.docs_path / "test.md").write_text("# Test Document")
        (self.docs_path / "guide.md").write_text("# Guide Document")

        # Mock sys.argv to simulate command line arguments
        import sys

        original_argv = sys.argv
        sys.argv = ["linking.py", "--prepare", "--docs-dir", str(self.docs_path)]

        try:
            # Act: Call main function
            linking.main()

            # Assert: Verify that prepare_docs was executed
            self.assertTrue(self.redirect_map_file.exists())
            redirect_map = json.loads(self.redirect_map_file.read_text())
            self.assertIn("test", redirect_map)
            self.assertIn("guide", redirect_map)

        finally:
            # Restore original argv
            sys.argv = original_argv

    def test_prepare_docs_called_correctly(self) -> None:
        """Test that prepare_docs is called when --prepare is used."""
        # Arrange: Create test files and mock prepare_docs
        (self.docs_path / "sample.md").write_text("# Sample")

        original_prepare_docs = linking.prepare_docs
        prepare_docs_called = False

        def mock_prepare_docs(docs_dir_path=None):
            nonlocal prepare_docs_called
            prepare_docs_called = True
            # Call the original function to ensure it works
            original_prepare_docs(docs_dir_path)

        linking.prepare_docs = mock_prepare_docs

        import sys

        original_argv = sys.argv
        # Pass docs_dir to ensure the redirect map is created in the temp folder
        sys.argv = ["linking.py", "--prepare", "--docs-dir", str(self.docs_path)]

        try:
            # Act: Call main
            linking.main()

            # Assert: Verify prepare_docs was called
            self.assertTrue(prepare_docs_called)

            # Verify it actually worked
            self.assertTrue(self.redirect_map_file.exists())

        finally:
            # Restore everything
            linking.prepare_docs = original_prepare_docs
            sys.argv = original_argv


if __name__ == "__main__":
    unittest.main(verbosity=2)
