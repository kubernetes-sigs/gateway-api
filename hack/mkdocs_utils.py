# Copyright 2026 The Kubernetes Authors.
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


"""
The core engine of the documentation tools. Contains all logic for ID generation, frontmatter parsing, link conversion, and YAML updates.
"""

import frontmatter
import json
import os
import re
import yaml
from pathlib import Path
from typing import Dict, List, Optional, Tuple


def prepare_docs(docs_dir_input: Optional[str | Path] = None, dry_run: bool = False) -> None:
    """
    Scan the documentation directory for Markdown files, inject a unique permanent
    ID into the frontmatter of each file if missing, and create a redirect map
    (JSON) of page IDs to file paths.
    """
    if docs_dir_input is None:
        docs_dir = DOCS_DIR
    else:
        docs_dir = Path(docs_dir_input)

    print(f"Starting documentation preparation ({'DRY RUN' if dry_run else 'LIVE'})...")

    # Validation: Ensure the documentation directory exists.
    if not docs_dir.exists():
        print(f"  ERROR: Directory {docs_dir} does not exist")
        return

    redirect_map: Dict[str, str] = {}

    # Standardize redirect map file location.
    if docs_dir.resolve() == DOCS_DIR.resolve():
        redirect_map_file = REDIRECT_MAP_FILE
    else:
        redirect_map_file = docs_dir.parent / "redirect_map.json"

    # Step 1: Iterate through all markdown files in the documentation directory.
    md_files = list(docs_dir.rglob("*.md"))
    print(f"  Found {len(md_files)} markdown files.")

    files_needing_ids = []
    files_with_ids = []

    if md_files:
        for md_file in md_files:
            try:
                # Step 2: Read file content and extract existing frontmatter.
                content: str = md_file.read_text("utf-8")
                metadata = get_frontmatter(content)
                page_id: str | None = metadata.get(FRONTMATTER_ID_KEY)

                # Step 3: If no permanent ID exists, generate one from the file path.
                if not page_id:
                    relative_path: Path = md_file.relative_to(docs_dir)
                    page_id = str(relative_path.with_suffix("")).replace(os.path.sep, "-")
                    files_needing_ids.append((md_file, page_id, metadata, content))
                else:
                    files_with_ids.append((md_file, page_id))

                # Step 4: Add the page's ID and current path to our map.
                redirect_map[page_id] = str(md_file.relative_to(docs_dir))
            except Exception as e:
                print(f"  Warning: Could not process {md_file}: {e}")
    else:
        print("  No markdown files found.")

    # Step 5: Report and apply changes.
    if files_needing_ids:
        print(f"\nFiles that {'would be' if dry_run else 'are being'} modified ({len(files_needing_ids)}):")
        for md_file, page_id, metadata, content_raw in files_needing_ids:
            print(f"  {'+' if dry_run else '-'} {md_file.relative_to(docs_dir)} -> ID: '{page_id}'")
            if not dry_run:
                try:
                    post = frontmatter.loads(content_raw)
                except Exception:
                    # Fallback for malformed YAML: manually inject ID line
                    if content_raw.startswith("---"):
                        parts = content_raw.split("---", 2)
                        if len(parts) >= 3:
                            new_fm = parts[1].rstrip() + f"\n{FRONTMATTER_ID_KEY}: {page_id}\n"
                            md_file.write_text(f"---{new_fm}---\n\n{parts[2]}", "utf-8")
                            continue
                    md_file.write_text(f"---\n{FRONTMATTER_ID_KEY}: {page_id}\n---\n\n{content_raw}", "utf-8")
                    continue

                post[FRONTMATTER_ID_KEY] = page_id
                md_file.write_text(frontmatter.dumps(post), "utf-8")
                continue

    if files_with_ids:
        print(f"\nFiles already with IDs ({len(files_with_ids)}):")
        for md_file, page_id in files_with_ids:
            print(f"  * {md_file.relative_to(docs_dir)} -> ID: '{page_id}'")

    # Step 6: Write or preview the completed map.
    if not dry_run:
        redirect_map_file.write_text(json.dumps(redirect_map, indent=2))
        print(f"\nPreparation complete. Map saved to {redirect_map_file}")
    else:
        print(f"\nDry run complete. Would create/update redirect map: {redirect_map_file}")


def preview_docs(docs_dir_input: Optional[str | Path] = None) -> None:
    """
    Simulates the process of preparing documentation files for unique page IDs without making any changes.
    """
    prepare_docs(docs_dir_input, dry_run=True)


def convert_internal_links(docs_dir_input: Optional[str | Path] = None) -> None:
    """
    Converts all relative Markdown links in a documentation directory to use an internal link macro.
    """
    if docs_dir_input is None:
        docs_dir = DOCS_DIR
    else:
        docs_dir = Path(docs_dir_input)

    print("Starting internal link conversion...")

    # Step 1: Build a map of file paths to their unique IDs.
    full_id_map = build_id_map(docs_dir)
    path_to_id_map: Dict[str, str] = {
        path.relative_to(docs_dir).as_posix(): pid for pid, path in full_id_map.items()
    }

    print(f"  - Built a map of {len(path_to_id_map)} page IDs.")

    # Step 2: Iterate through each file and replace its links.
    files_converted = 0
    for md_file in docs_dir.rglob("*.md"):
        try:
            content = md_file.read_text("utf-8")
            original_content = content

            def replace_link(match: re.Match) -> str:
                link_text = match.group(1)
                link_url = match.group(2)

                if link_url.startswith(("http", "#", "mailto:")) or not link_url.endswith(".md"):
                    return match.group(0)

                current_dir = md_file.parent
                target_file = (current_dir / link_url).resolve()

                try:
                    target_relative_path = target_file.relative_to(docs_dir.resolve())
                    target_key = target_relative_path.as_posix()
                except ValueError:
                    return match.group(0)

                target_id = path_to_id_map.get(target_key)
                if target_id:
                    return f'[{link_text}]({{{{ internal_link("{target_id}") }}}})'
                else:
                    return match.group(0)

            content = re.sub(r"\[([^\]]+)\]\((?!{{)([^)]+\.md)\)", replace_link, content)

            if content != original_content:
                files_converted += 1
                md_file.write_text(content, "utf-8")
                print(f"  - Converted links in {md_file.relative_to(docs_dir)}")
        except Exception as e:
            print(f"  Warning: Could not convert links in {md_file}: {e}")

    print(f"Link conversion complete. Modified {files_converted} files.")


def update_mkdocs_yml_redirects(redirect_updates: Dict[str, str]) -> bool:
    """
    Updates the 'mkdocs.yml' configuration file with new redirect rules using safe YAML parsing.
    This function ensures that the 'redirects' plugin and its 'redirect_maps' section exist in the
    MkDocs configuration file. It merges the provided redirect rules (mapping old paths to new paths)
    with any existing rules, preserving the overall structure and formatting of the YAML file as much
    as possible.
    """
    mkdocs_yml_path = Path("mkdocs.yml")
    if not mkdocs_yml_path.exists():
        print("  Warning: mkdocs.yml not found. Cannot update redirects.")
        return False

    try:
        with open(mkdocs_yml_path, "r", encoding="utf-8") as f:
            config = yaml.safe_load(f)

        if not isinstance(config, dict):
            print("  Warning: mkdocs.yml is not a valid YAML dictionary.")
            return False

        # Step 1: Ensure 'plugins' section exists
        plugins = config.setdefault("plugins", [])

        # Step 2: Find the first redirects plugin configuration
        redirects_plugin_entry = None
        for plugin in plugins:
            if isinstance(plugin, dict) and "redirects" in plugin:
                redirects_plugin_entry = plugin
                break
            elif isinstance(plugin, str) and plugin == "redirects":
                # Found a string entry, which we will replace with a dict
                redirects_plugin_entry = plugin
                break

        # Step 3: If no entry exists, create a new one
        if redirects_plugin_entry is None:
            redirects_plugin_entry = {"redirects": {"redirect_maps": {}}}
            plugins.append(redirects_plugin_entry)

        # Step 4: If the entry was a string, replace it with a proper dict structure
        if isinstance(redirects_plugin_entry, str):
            plugins[plugins.index(redirects_plugin_entry)] = {
                "redirects": {"redirect_maps": {}}
            }
            redirects_plugin_entry = plugins[
                plugins.index({"redirects": {"redirect_maps": {}}})
            ]

        # Step 5: Get the config dict for the redirects plugin
        redirects_plugin_config = redirects_plugin_entry.setdefault("redirects", {})

        # Step 6: Handle case where config is `redirects: null`
        if redirects_plugin_config is None:
            redirects_plugin_config = {}
            redirects_plugin_entry["redirects"] = redirects_plugin_config

        # Step 7: Ensure 'redirect_maps' exists
        redirect_maps = redirects_plugin_config.setdefault("redirect_maps", {})

        # Step 8: Handle case where `redirect_maps:` is present but empty
        if redirect_maps is None:
            redirect_maps = {}
            redirects_plugin_config["redirect_maps"] = redirect_maps

        # Step 9: Check if there are actual changes to be made before writing the file
        if not any(
            redirect_updates.get(k) != redirect_maps.get(k) for k in redirect_updates
        ):
            print("  No new redirect updates needed in mkdocs.yml.")
            return True

        # Step 10: Update redirect_maps with new redirects
        redirect_maps.update(redirect_updates)

        # Step 11: Write the updated config back to mkdocs.yml
        with open(mkdocs_yml_path, "w", encoding="utf-8") as f:
            yaml.dump(config, f, default_flow_style=False, sort_keys=False, indent=2)

        print(f"  Updated mkdocs.yml with {len(redirect_updates)} redirect rules.")
        return True

    except (yaml.YAMLError, IOError) as e:
        print(f"  Error updating mkdocs.yml: {e}")
        return False


def build_id_map(docs_dir: Path) -> Dict[str, Path]:
    """
    Scans the documentation directory and builds a mapping between unique
    page IDs and their absolute file paths.
    """
    id_map: Dict[str, Path] = {}
    if not docs_dir.exists():
        return id_map

    for md_file in docs_dir.rglob("*.md"):
        try:
            content = md_file.read_text("utf-8")
            frontmatter = get_frontmatter(content)
            page_id = frontmatter.get(FRONTMATTER_ID_KEY)
            if page_id:
                id_map[page_id] = md_file
        except (OSError, UnicodeDecodeError):
            continue  # Skip files that can't be read

    return id_map


# --- Configuration ---
# Global constants defining key file paths and metadata keys.
DOCS_DIR: Path = Path("docs")
REDIRECT_MAP_FILE: Path = Path("hack/redirect_map.json")
FRONTMATTER_ID_KEY: str = "id"


def get_frontmatter(content: str) -> Dict[str, str]:
    """Extracts frontmatter using python-frontmatter.

    Args:
        content: The full text content of a file.

    Returns:
        The metadata as a dictionary.
    """
    try:
        metadata, _ = frontmatter.parse(content)
        return metadata
    except Exception:
        # Fallback for malformed YAML: try to extract 'id:' at least
        metadata = {}
        match = re.search(r"^---\s*\n(.*?)\n---\s*\n", content, re.DOTALL)
        if match:
            fm_text = match.group(1)
            id_match = re.search(r"^id:\s*(.+)$", fm_text, re.MULTILINE)
            if id_match:
                metadata[FRONTMATTER_ID_KEY] = id_match.group(1).strip().strip("\"'")
            return metadata
        return {}


def format_frontmatter_str(data: Dict[str, str]) -> str:
    """Formats a dictionary into a YAML frontmatter string block."""
    if not data:
        return ""
    return yaml.dump(data, default_flow_style=False, sort_keys=False, indent=2)
