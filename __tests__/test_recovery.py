import json
import shutil
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

import linking
from linking import prepare_docs


class TestErrorRecovery(unittest.TestCase):
    """Tests focused on error recovery and system robustness."""

    def setUp(self) -> None:
        """Set up test environment."""
        self.test_dir = Path("./temp_test_robustness")
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

    def test_concurrent_access_simulation(self) -> None:
        """Test robustness when files are modified during processing."""
        import threading
        import time

        # Create initial files
        for i in range(5):
            (linking.DOCS_DIR / f"concurrent-{i}.md").write_text(f"""---
id: concurrent-{i}
---
# Concurrent Test {i}""")

        def modify_files():
            """Simulate another process modifying files during processing."""
            time.sleep(0.1)  # Give prepare_docs a chance to start
            try:
                # Modify a file during processing
                (linking.DOCS_DIR / "concurrent-2.md").write_text("""---
id: concurrent-2-modified
---
# Modified During Processing""")
            except FileNotFoundError:
                pass  # File might be temporarily locked

        # Start background modification
        modifier_thread = threading.Thread(target=modify_files)
        modifier_thread.start()

        # Run preparation while files are being modified
        prepare_docs()

        modifier_thread.join()

        # Should complete successfully despite concurrent modifications
        self.assertTrue(linking.REDIRECT_MAP_FILE.exists())
        redirect_map = json.loads(linking.REDIRECT_MAP_FILE.read_text())

        # Should have processed most files
        self.assertGreaterEqual(len(redirect_map), 4)

    def test_special_characters_in_filenames(self) -> None:
        """Test handling of files with special characters in names."""
        special_files = {
            "file with spaces.md": "file-with-spaces",
            "file-with-üñïçødé.md": "file-with-unicode",
            "file.with.dots.md": "file-with-dots",
            "file[with]brackets.md": "file-with-brackets",
            "file(with)parens.md": "file-with-parens",
            "file&with&symbols.md": "file-with-symbols",
        }

        for filename, expected_id in special_files.items():
            try:
                (linking.DOCS_DIR / filename).write_text(f"""---
id: {expected_id}
---
# Test File""")
            except OSError:
                # Some filesystems don't support certain characters
                continue

        prepare_docs()

        # Should handle special characters in filenames
        redirect_map = json.loads(linking.REDIRECT_MAP_FILE.read_text())

        # Check that files were processed (those that could be created)
        for filename, expected_id in special_files.items():
            if (linking.DOCS_DIR / filename).exists():
                self.assertIn(expected_id, redirect_map)
                self.assertEqual(redirect_map[expected_id], filename)
