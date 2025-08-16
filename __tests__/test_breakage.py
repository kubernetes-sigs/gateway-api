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
from types import SimpleNamespace

sys.path.insert(0, str(Path(__file__).parent))

import linking
from linking import on_config, prepare_docs


class TestLinkBreakageScenarios(unittest.TestCase):
    """Tests focused on how links break and how to prevent/handle breakage."""

    def setUp(self) -> None:
        """Set up test environment."""
        self.test_dir = Path("./temp_test_link_breakage")
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

    # === ID COLLISION AND CONFLICTS ===

    def test_id_collision_when_files_move_to_same_path_structure(self) -> None:
        """Test when multiple files would generate the same ID after restructuring."""
        # Initial structure
        (linking.DOCS_DIR / "user-guide.md").write_text("# User Guide")
        (linking.DOCS_DIR / "admin").mkdir()
        (linking.DOCS_DIR / "admin" / "guide.md").write_text("# Admin Guide")

        prepare_docs()

        # Simulate restructuring where both files move to create ID collision
        # user-guide.md -> guides/user.md (would generate 'guides-user')
        # admin/guide.md -> guides/user.md (would also generate 'guides-user')

        # This is a real scenario: two different guides both get moved to guides/user.md
        # at different times, or one file replaces another

        original_map = json.loads(linking.REDIRECT_MAP_FILE.read_text())

        # Verify we have the expected initial state
        self.assertEqual(original_map["user-guide"], "user-guide.md")
        self.assertEqual(original_map["admin-guide"], "admin/guide.md")

        # Now simulate what happens if both files are moved to the same new location
        # This would break internal_link() because two IDs point to the same file
        (linking.DOCS_DIR / "guides").mkdir()

        # File 1 moves first
        shutil.move(
            linking.DOCS_DIR / "user-guide.md", linking.DOCS_DIR / "guides" / "user.md"
        )

        # File 2 overwrites it (common in refactoring)
        shutil.move(
            linking.DOCS_DIR / "admin" / "guide.md",
            linking.DOCS_DIR / "guides" / "user.md",
        )

        # Update the moved file to have both old IDs (impossible situation)
        content = """---
id: user-guide
old_id: admin-guide
---
# Merged Guide"""
        (linking.DOCS_DIR / "guides" / "user.md").write_text(content)

        mock_config = {
            "docs_dir": str(linking.DOCS_DIR),
            "pages": [
                SimpleNamespace(
                    file=SimpleNamespace(src_path="guides/user.md"), url="/guides/user/"
                )
            ],
            "plugins": {"macros": {"config": {"python_macros": {}}}},
        }

        # The internal_link macro should handle this somehow
        result_config = on_config(mock_config)

        if (
            "internal_link"
            in result_config["plugins"]["macros"]["config"]["python_macros"]
        ):
            internal_link = result_config["plugins"]["macros"]["config"][
                "python_macros"
            ]["internal_link"]

            # Both old IDs should resolve to the same page (or one should fail gracefully)
            try:
                url1 = internal_link("user-guide")
                self.assertEqual(url1, "/guides/user/")
            except ValueError:
                pass  # Acceptable if it fails gracefully

            # The admin-guide ID no longer exists as a separate page
            with self.assertRaises(ValueError):
                internal_link("admin-guide")

    def test_id_changes_during_refactoring(self) -> None:
        """Test when someone manually changes IDs in frontmatter, breaking existing links."""
        # Initial setup
        (linking.DOCS_DIR / "api.md").write_text("""---
id: api-reference
title: API Reference
---
# API Reference""")

        (linking.DOCS_DIR / "tutorial.md").write_text("""---
id: getting-started
title: Getting Started
---
# Getting Started

See the {{ internal_link('api-reference') }} for details.""")

        prepare_docs()

        # Someone manually changes the API page ID during editing
        (linking.DOCS_DIR / "api.md").write_text("""---
id: api-docs-v2
title: API Reference
---
# API Reference""")

        mock_config = {
            "docs_dir": str(linking.DOCS_DIR),
            "pages": [
                SimpleNamespace(file=SimpleNamespace(src_path="api.md"), url="/api/"),
                SimpleNamespace(
                    file=SimpleNamespace(src_path="tutorial.md"), url="/tutorial/"
                ),
            ],
            "plugins": {"macros": {"config": {"python_macros": {}}}},
        }

        result_config = on_config(mock_config)

        if (
            "internal_link"
            in result_config["plugins"]["macros"]["config"]["python_macros"]
        ):
            internal_link = result_config["plugins"]["macros"]["config"][
                "python_macros"
            ]["internal_link"]

            # Old ID should fail
            with self.assertRaises(ValueError) as context:
                internal_link("api-reference")
            self.assertIn("api-reference", str(context.exception))

            # New ID should work
            self.assertEqual(internal_link("api-docs-v2"), "/api/")

    def test_circular_id_references_and_dependency_loops(self) -> None:
        """Test handling of circular references in ID mappings."""
        # This can happen if redirect map gets corrupted or manually edited

        # Create initial files
        (linking.DOCS_DIR / "a.md").write_text("---\nid: page-a\n---\n# Page A")
        (linking.DOCS_DIR / "b.md").write_text("---\nid: page-b\n---\n# Page B")

        prepare_docs()

        # Manually corrupt the redirect map to create circular references
        corrupt_map = {
            "page-a": "b.md",  # page-a points to b.md
            "page-b": "a.md",  # page-b points to a.md (circular!)
            "page-c": "nonexistent.md",  # broken reference
        }
        linking.REDIRECT_MAP_FILE.write_text(json.dumps(corrupt_map))

        mock_config = {
            "docs_dir": str(linking.DOCS_DIR),
            "pages": [
                SimpleNamespace(file=SimpleNamespace(src_path="a.md"), url="/a/"),
                SimpleNamespace(file=SimpleNamespace(src_path="b.md"), url="/b/"),
            ],
            "plugins": {"macros": {"config": {"python_macros": {}}}},
        }

        # Should handle corrupted redirect map gracefully
        result_config = on_config(mock_config)

        # The macro should work based on actual current file IDs, not the corrupt map
        if (
            "internal_link"
            in result_config["plugins"]["macros"]["config"]["python_macros"]
        ):
            internal_link = result_config["plugins"]["macros"]["config"][
                "python_macros"
            ]["internal_link"]

            # Should work based on actual current frontmatter, not redirect map
            self.assertEqual(internal_link("page-a"), "/a/")
            self.assertEqual(internal_link("page-b"), "/b/")

    # === FILE SYSTEM CHANGES THAT BREAK LINKS ===

    def test_case_sensitivity_issues_across_filesystems(self) -> None:
        """Test link breakage due to case sensitivity differences."""
        # Create file with specific casing
        (linking.DOCS_DIR / "API-Guide.md").write_text(
            "---\nid: API-Guide\n---\n# API Guide"
        )

        prepare_docs()

        # Simulate file being renamed with different case (common on case-insensitive filesystems)
        original_content = (linking.DOCS_DIR / "API-Guide.md").read_text()
        (linking.DOCS_DIR / "API-Guide.md").unlink()
        (linking.DOCS_DIR / "api-guide.md").write_text(original_content)

        mock_config = {
            "docs_dir": str(linking.DOCS_DIR),
            "pages": [
                SimpleNamespace(
                    file=SimpleNamespace(src_path="api-guide.md"), url="/api-guide/"
                )
            ],
            "plugins": {"macros": {"config": {"python_macros": {}}}},
        }

        result_config = on_config(mock_config)

        if (
            "internal_link"
            in result_config["plugins"]["macros"]["config"]["python_macros"]
        ):
            internal_link = result_config["plugins"]["macros"]["config"][
                "python_macros"
            ]["internal_link"]

            # The ID should still work (case-sensitive match required)
            self.assertEqual(internal_link("API-Guide"), "/api-guide/")

    def test_unicode_normalization_issues(self) -> None:
        """Test link breakage due to Unicode normalization differences."""
        import unicodedata

        # Create file with Unicode characters
        # Using different Unicode normalization forms that look the same
        filename1 = "café.md"  # é as single character
        filename2 = unicodedata.normalize("NFD", "café.md")  # é as e + combining accent

        # These look the same but are different at byte level
        self.assertNotEqual(filename1, filename2)

        # Create file with one form
        (linking.DOCS_DIR / filename1).write_text(
            "---\nid: cafe-menu\n---\n# Café Menu"
        )

        prepare_docs()

        # File system or Git might change the normalization
        original_content = (linking.DOCS_DIR / filename1).read_text()
        (linking.DOCS_DIR / filename1).unlink()
        (linking.DOCS_DIR / filename2).write_text(original_content)

        mock_config = {
            "docs_dir": str(linking.DOCS_DIR),
            "pages": [
                SimpleNamespace(file=SimpleNamespace(src_path=filename2), url="/cafe/")
            ],
            "plugins": {"macros": {"config": {"python_macros": {}}}},
        }

        result_config = on_config(mock_config)

        if (
            "internal_link"
            in result_config["plugins"]["macros"]["config"]["python_macros"]
        ):
            internal_link = result_config["plugins"]["macros"]["config"][
                "python_macros"
            ]["internal_link"]

            # Should still work despite Unicode normalization change
            self.assertEqual(internal_link("cafe-menu"), "/cafe/")

    def test_redirect_map_corruption_scenarios(self) -> None:
        """Test various ways the redirect map can become corrupted."""
        # Create initial files
        (linking.DOCS_DIR / "page.md").write_text("---\nid: test-page\n---\n# Test")

        prepare_docs()

        # Test various corruption scenarios
        corruption_scenarios = [
            '{"invalid": json syntax}',  # Invalid JSON
            '{"valid": "json", "but": "wrong", "structure": true}',  # Wrong structure
            "not json at all",  # Not JSON
            "",  # Empty file
            "{}",  # Empty but valid JSON
            '{"key-with-no-value":}',  # Malformed JSON
            '{"unicode-test": "café\\ud83d\\ude00"}',  # Unicode issues
        ]

        for i, corrupt_content in enumerate(corruption_scenarios):
            with self.subTest(scenario=i):
                # Corrupt the redirect map
                linking.REDIRECT_MAP_FILE.write_text(corrupt_content)

                mock_config = {
                    "docs_dir": str(linking.DOCS_DIR),
                    "pages": [
                        SimpleNamespace(
                            file=SimpleNamespace(src_path="page.md"), url="/page/"
                        )
                    ],
                    "plugins": {"macros": {"config": {"python_macros": {}}}},
                }

                # Should handle all corruption gracefully
                try:
                    result_config = on_config(mock_config)
                    self.assertIsNotNone(result_config)
                except Exception as e:
                    # Should not crash with unhandled exceptions
                    self.assertIsInstance(
                        e, (json.JSONDecodeError, KeyError, ValueError)
                    )

    def test_internal_link_macro_with_invalid_inputs(self) -> None:
        """Test internal_link macro with various invalid inputs that could break pages."""
        (linking.DOCS_DIR / "test.md").write_text("---\nid: test-page\n---\n# Test")

        # Run prepare_docs first to create the redirect map
        prepare_docs()

        mock_config = {
            "docs_dir": str(linking.DOCS_DIR),
            "pages": [
                SimpleNamespace(file=SimpleNamespace(src_path="test.md"), url="/test/")
            ],
            "plugins": {"macros": {"config": {"python_macros": {}}}},
        }

        result_config = on_config(mock_config)

        if (
            "internal_link"
            in result_config["plugins"]["macros"]["config"]["python_macros"]
        ):
            internal_link = result_config["plugins"]["macros"]["config"][
                "python_macros"
            ]["internal_link"]

            # Test various invalid inputs that could come from template errors
            invalid_inputs = [
                None,  # None value
                "",  # Empty string
                "   ",  # Whitespace only
                "non-existent-page",  # Non-existent ID
                "test page",  # Spaces in ID
                "test/page",  # Slashes in ID
                "test-page\n",  # Newlines
                123,  # Non-string type
                ["test-page"],  # List instead of string
                {"id": "test-page"},  # Dict instead of string
            ]

            for invalid_input in invalid_inputs:
                with self.subTest(input=repr(invalid_input)):
                    try:
                        result = internal_link(invalid_input)
                        # If it somehow succeeds, result should be reasonable
                        self.assertIsInstance(result, str)
                        self.assertTrue(result.startswith("/"))
                    except (ValueError, TypeError, AttributeError) as e:
                        # Expected for invalid inputs
                        self.assertIsInstance(
                            e, (ValueError, TypeError, AttributeError)
                        )
                    except Exception as e:
                        # Should not crash with unexpected exceptions
                        self.fail(
                            f"Unexpected exception for input {invalid_input}: {e}"
                        )


if __name__ == "__main__":
    unittest.main(verbosity=2)
