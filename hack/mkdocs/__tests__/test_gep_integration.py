
import unittest
import subprocess
import importlib.util
from pathlib import Path
from unittest.mock import MagicMock

# Helper to import module with dashes
def import_path(path):
    spec = importlib.util.spec_from_file_location("mkdocs_copy_geps", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module

class TestGEPIntegration(unittest.TestCase):
    def setUp(self):
        self.root_dir = Path(__file__).resolve().parents[3] # gateway-api root
        self.geps_dir = self.root_dir / "geps"
        self.script_path = self.root_dir / "hack" / "mkdocs-copy-geps.py"
        self.update_script = self.root_dir / "hack" / "update-geps.sh"
        self.module = import_path(str(self.script_path))

    def test_update_geps_script_runs(self):
        """Verify hack/update-geps.sh runs successfully."""
        result = subprocess.run(
            ["bash", str(self.update_script)], 
            cwd=str(self.root_dir), 
            capture_output=True, 
            text=True
        )
        self.assertEqual(result.returncode, 0, f"Script failed: {result.stderr}")
        # Verify side effects? ensure geps dir exists
        self.assertTrue(self.geps_dir.exists())

    def test_mkdocs_hook_copies_geps(self):
        """Verify on_files hook logic in mkdocs-copy-geps.py."""
        # Mock MkDocs objects
        mock_files = MagicMock()
        mock_files.get_file_from_path.return_value = None
        mock_files.append = MagicMock()
        
        mock_config = {
            'site_dir': 'site',
            'use_directory_urls': True
        }

        # Create a dummy GEP file to test logic
        test_gep = self.geps_dir / "test-gep-integration.md"
        created_file = False
        if not test_gep.exists():
             test_gep.write_text("Test GEP content")
             created_file = True

        try:
            # Run hook
            self.module.on_files(mock_files, mock_config)
            
            # Verify file was added
            # The hook iterates 'geps' dir.
            # We check if mock_files.append was called with a file object having our path
            calls = mock_files.append.call_args_list
            found = False
            for call in calls:
                file_obj = call[0][0]
                # src_path should be consistent with how the hook creates the File object
                # The hook uses str(root_dir / filename) which would be full path or relative
                # Let's check string representation or src_path attribute
                if "test-gep-integration.md" in str(file_obj.src_path):
                    found = True
                    break
            self.assertTrue(found, "GEP file was not added to MkDocs files")
            
        finally:
            if created_file:
                test_gep.unlink()

if __name__ == "__main__":
    unittest.main()
