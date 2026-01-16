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

import io
import shutil
import sys
import unittest
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).parents[2]))

import mkdocs_linking as linking

class TestPreviewDocs(unittest.TestCase):
    """Tests for the preview_docs function."""

    def setUp(self) -> None:
        """Set up a temporary directory for each test."""
        self.test_dir = Path("./temp_test_preview")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.docs_dir = self.test_dir / "my-docs"
        self.docs_dir.mkdir(parents=True)

    def tearDown(self) -> None:
        """Clean up the temporary directory after each test."""
        shutil.rmtree(self.test_dir)

    def test_preview_with_mixed_files(self):
        """Test preview output with a mix of files with and without IDs."""
        # Arrange
        (self.docs_dir / "index.md").write_text("# Welcome")
        (self.docs_dir / "has-id.md").write_text("---\nid: existing-id\n---\n# Has ID")
        
        # Act
        captured_output = io.StringIO()
        with patch("sys.stdout", captured_output):
            linking.preview_docs(str(self.docs_dir))
        
        # Assert
        output = captured_output.getvalue()
        self.assertIn("Files that would be modified (1):", output)
        self.assertIn("+ index.md -> ID: 'index'", output)
        self.assertIn("Files already with IDs (1):", output)
        self.assertIn("* has-id.md -> ID: 'existing-id'", output)
        self.assertIn("Would create/update redirect map", output)

    def test_preview_with_empty_directory(self):
        """Test preview output for an empty directory."""
        # Arrange (directory is already empty)
        
        # Act
        captured_output = io.StringIO()
        with patch("sys.stdout", captured_output):
            linking.preview_docs(str(self.docs_dir))
            
        # Assert
        output = captured_output.getvalue()
        self.assertIn("No markdown files found", output)

    def test_preview_with_nonexistent_directory(self):
        """Test preview output for a non-existent directory."""
        # Arrange
        non_existent_path = str(self.test_dir / "non-existent")
        
        # Act
        captured_output = io.StringIO()
        with patch("sys.stdout", captured_output):
            linking.preview_docs(non_existent_path)
            
        # Assert
        output = captured_output.getvalue()
        self.assertIn("ERROR: Directory", output)
        self.assertIn("does not exist", output)

if __name__ == "__main__":
    unittest.main(verbosity=2)
