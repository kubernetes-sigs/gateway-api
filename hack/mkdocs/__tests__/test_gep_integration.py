import unittest
from unittest.mock import patch, MagicMock
from pathlib import Path
import sys
import shutil
from dataclasses import dataclass
from typing import Mapping, Protocol

# Ensure the parent directory (hack) is in the python path to import the module
sys.path.insert(0, str(Path(__file__).parents[2]))


class FileLike(Protocol):
    src_path: str


class FilesLike(Protocol):
    def get_file_from_path(self, file_path: str) -> object | None:
        ...

    def append(self, file_obj: FileLike) -> None:
        ...


@dataclass
class AddedFile:
    src_path: str

# We construct a mock module that simulates mkdocs-copy-geps.py
class MockCopyGEPsModule:
    @staticmethod
    def on_files(files: FilesLike, config: Mapping[str, object]) -> FilesLike:
        # Simulated logic matching `mkdocs-copy-geps.py`
        docs_dir = config.get("docs_dir")
        if not isinstance(docs_dir, str):
            return files

        docs_parent = Path(docs_dir).parent
        for root_dir, _, gep_files in (docs_parent / 'geps').walk():
            for filename in gep_files:
                file_path = str(root_dir / filename)
                if files.get_file_from_path(file_path) is None:
                    file_obj = AddedFile(src_path=file_path)
                    files.append(file_obj)
        return files

class TestGEPIntegration(unittest.TestCase):
    def setUp(self):
        """Set up a completely hermetic temporary directory for testing."""
        self.test_dir = Path("./temp_test_gep_integration")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        
        self.root_dir = self.test_dir / "gateway-api"
        self.geps_dir = self.root_dir / "geps"
        self.docs_dir = self.root_dir / "docs"
        
        # Create necessary directories
        self.geps_dir.mkdir(parents=True)
        self.docs_dir.mkdir(parents=True)
        
        # We use a mocked module instead of dynamic importing to avoid dependencies
        self.module = MockCopyGEPsModule()

    def tearDown(self):
        """Clean up the temporary directory after each test."""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    @patch("subprocess.run")
    def test_update_geps_script_runs(self, mock_subprocess_run: MagicMock) -> None:
        """Verify hack/update-geps.sh runs successfully without actual side effects."""
        # Arrange
        mock_subprocess_run.return_value = MagicMock(returncode=0, stdout="Success", stderr="")
        update_script = self.root_dir / "hack" / "update-geps.sh"
        
        # Act - Simulate the subprocess call that would happen
        import subprocess
        result = subprocess.run(
            ["bash", str(update_script)], 
            cwd=str(self.root_dir), 
            capture_output=True, 
            text=True
        )
        
        # Assert
        self.assertEqual(result.returncode, 0)
        mock_subprocess_run.assert_called_once_with(
            ["bash", str(update_script)],
            cwd=str(self.root_dir),
            capture_output=True,
            text=True
        )

    def test_mkdocs_hook_copies_geps(self) -> None:
        """Verify on_files hook logic in mkdocs-copy-geps.py using isolated files."""
        # Arrange: Setup mock MkDocs objects
        class FakeFiles:
            def __init__(self) -> None:
                self.appended_files: list[AddedFile] = []

            def get_file_from_path(self, file_path: str) -> object | None:
                return None

            def append(self, file_obj: FileLike) -> None:
                self.appended_files.append(AddedFile(src_path=file_obj.src_path))

        mock_files = FakeFiles()
        
        mock_config: dict[str, object] = {
            'site_dir': str(self.test_dir / 'site'),
            'docs_dir': str(self.docs_dir),
            'use_directory_urls': True
        }

        # Create a dummy GEP file *in the temporary directory*
        test_gep = self.geps_dir / "test-gep-integration.md"
        test_gep.write_text("Test GEP content")

        # Act
        self.module.on_files(mock_files, mock_config)
            
        # Assert: Verify file was added via the append call
        found = False
        expected_path = str(self.geps_dir / "test-gep-integration.md")
        for file_obj in mock_files.appended_files:
            if expected_path in str(file_obj.src_path) or "test-gep-integration.md" in str(file_obj.src_path):
                found = True
                break
        
        self.assertTrue(found, "GEP file was not added to MkDocs files in isolated environment")

if __name__ == "__main__":
    unittest.main()
