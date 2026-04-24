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


"""
The core engine of the documentation tools. Contains all logic for ID generation, frontmatter parsing, link conversion, and YAML updates.
"""

import frontmatter
import json
import os
import re
import yaml
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Union, Any


def prepare_docs(docs_dir_input: Optional[Union[str, Path]] = None, dry_run: bool = False) -> None:
    """
    Scan the documentation directory for Markdown files, inject a unique permanent
    ID into the frontmatter of each file if missing, and create a redirect map
    (JSON) of page IDs to file paths.

    Args:
        docs_dir_input: Optional override for the documentation directory.
        dry_run: If True, only prints what changes would be made without modifying files.
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

    # Standardize page ID map file location.
    if docs_dir.resolve() == DOCS_DIR.resolve():
        page_id_map_file = PAGE_ID_MAP_FILE
    else:
        page_id_map_file = docs_dir.parent / "page_id_map.json"

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
                page_id: Optional[str] = metadata.get(FRONTMATTER_ID_KEY)

                # Step 3: If no permanent ID exists, generate one from the file path.
                if not page_id:
                    relative_path: Path = md_file.relative_to(docs_dir)
                    page_id = str(relative_path.with_suffix("")).replace(os.path.sep, "-")
                    files_needing_ids.append((md_file, page_id, metadata, content))
                else:
                    files_with_ids.append((md_file, page_id))

                # Step 4: Add the page's ID and current path to our map, checking for collisions.
                if page_id in redirect_map:
                    collision_path = redirect_map[page_id]
                    print(
                        f"  WARNING: ID collision detected for '{page_id}'!"
                    )
                    print(f"    Existing: {collision_path}")
                    print(f"    New:      {md_file.relative_to(docs_dir)}")
                    print("    Links using this ID might resolve incorrectly.")

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
        page_id_map_file.write_text(json.dumps(redirect_map, indent=2))
        print(f"\nPreparation complete. Map saved to {page_id_map_file}")
    else:
        print(f"\nDry run complete. Would create/update page ID map: {page_id_map_file}")


def preview_docs(docs_dir_input: Optional[Union[str, Path]] = None) -> None:
    """
    Simulates the process of preparing documentation files for unique page IDs without making any changes.

    Args:
        docs_dir_input: Optional override for the documentation directory.
    """
    prepare_docs(docs_dir_input, dry_run=True)


def convert_internal_links(docs_dir_input: Optional[Union[str, Path]] = None) -> None:
    """
    Converts all relative Markdown links in a documentation directory to use an internal link macro.

    Args:
        docs_dir_input: Optional override for the documentation directory.
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

            # --- Masking Strategy: Protect code blocks from link conversion ---
            placeholders: List[str] = []

            def mask_code(match: re.Match) -> str:
                idx = len(placeholders)
                placeholders.append(match.group(0))
                return f"__GW_API_INTERNAL_LINK_MASK_{idx}__"

            # Mask fenced code blocks (``` ... ```)
            # Pattern: matches three backticks followed by anything (non-greedy) until the next three backticks.
            # re.DOTALL is crucial here to ensure '.' matches newlines within the block.
            content = re.sub(r"```.*?```", mask_code, content, flags=re.DOTALL)

            # Mask inline code (`...`)
            # Pattern: matches a single backtick NOT preceded by another backtick (lookbehind),
            # followed by any characters that are NOT a backtick or newline, followed by a backtick.
            # This avoids nested backticks and ensures we don't match our own protectors.
            content = re.sub(r"(?<!`)`[^`\n]+`", mask_code, content)

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

            # Perform the conversion on the remaining text
            # Regex Breakdown:
            # \[([^\]]+)\]  -> Matches link text inside square brackets [text]
            # \(            -> Matches opening parenthesis (
            # (?!{{)        -> Negative lookahead: Ensures the link does NOT already start with '{{' (already converted)
            # ([^)]+\.md)   -> Matches the link URL ending in '.md', capturing it in a group
            # \)            -> Matches closing parenthesis )
            content = re.sub(r"\[([^\]]+)\]\((?!{{)([^)]+\.md)\)", replace_link, content)

            # Unmask the code blocks
            for i, original in enumerate(placeholders):
                content = content.replace(f"__GW_API_INTERNAL_LINK_MASK_{i}__", original)

            if content != original_content:
                files_converted += 1
                md_file.write_text(content, "utf-8")
                print(f"  - Converted links in {md_file.relative_to(docs_dir)}")
        except Exception as e:
            print(f"  Warning: Could not convert links in {md_file}: {e}")

    print(f"Link conversion complete. Modified {files_converted} files.")


def update_mkdocs_yml_redirects(
    redirect_updates: Dict[str, str], mkdocs_yml_path_input: Optional[Union[str, Path]] = None
) -> bool:
    """
    Updates the 'mkdocs.yml' configuration file with new redirect rules using regex patching.
    This function ensures that the 'redirects' plugin and its 'redirect_maps' section exist in the
    MkDocs configuration file. It merges the provided redirect rules with any existing rules,
    preserving all comments, indentation, and custom tags throughout the file.

    Args:
        redirect_updates: A dictionary mapping old paths to new paths.
        mkdocs_yml_path_input: Optional override for the mkdocs.yml file path.

    Returns:
        bool: True if the update was successful (or not needed), False otherwise.
    """
    if mkdocs_yml_path_input is None:
        mkdocs_yml_path = MKDOCS_YML_PATH
    else:
        mkdocs_yml_path = Path(mkdocs_yml_path_input)

    if not mkdocs_yml_path.exists():
        print(f"  Warning: {mkdocs_yml_path} not found. Cannot update redirects.")
        return False

    try:
        raw_content = mkdocs_yml_path.read_text("utf-8")
        
        # Merge the new redirects into existing ones
        existing_maps = {}
        
        # 1. Try to find the existing redirect_maps block
        # Robust regex Breakdown:
        # ^(\s*)           -> Start of line, capture leading whitespace (indentation)
        # redirect_maps:   -> Literal key
        # \s*              -> Optional whitespace after colon
        # (?:#.*)?         -> Optional non-capturing group for trailing comments (e.g. # auto)
        # (                -> Start capture group for the value/block state
        #   \{\s*\}        -> Matches empty braces '{}' 
        #   |null          -> OR literal 'null'
        #   |(\n|$)        -> OR a newline (start of block) or end of file
        # )
        match = re.search(r"^(\s*)redirect_maps:\s*(?:#.*)?(\{\s*\}|null|(\n|$))", raw_content, re.MULTILINE)
        if match:
            indent = match.group(1)
            inner_indent = indent + "  "
            is_empty_assignment = "null" in match.group(0) or "{}" in match.group(0)
            
            lines = raw_content[match.end():].splitlines()
            block_lines_count = 0
            if not is_empty_assignment:
                for line in lines:
                    if not line.strip() or line.startswith(inner_indent):
                        block_lines_count += 1
                        # Capture key, value, and optional trailing comment
                        entry_match = re.match(r"^\s*([^#:\s]+)\s*:\s*([^#\s]+)\s*(?:#\s*(.*))?", line)
                        if entry_match:
                            key = entry_match.group(1)
                            val = entry_match.group(2)
                            comment = entry_match.group(3)
                            existing_maps[key] = (val, comment)
                    else:
                        break
            
            # Check if updates are actually needed
            needs_update = False
            for k, v in redirect_updates.items():
                if k not in existing_maps or existing_maps[k][0] != v:
                    needs_update = True
                    break
            
            if not needs_update:
                print("  No new redirect updates needed in mkdocs.yml.")
                return True
                
            updated_maps = existing_maps.copy()
            for k, v in redirect_updates.items():
                if k in updated_maps:
                    # Update value, keep original comment if it exists
                    updated_maps[k] = (v, updated_maps[k][1])
                else:
                    updated_maps[k] = (v, None)
            
            # Build the new block content
            new_lines = [f"{indent}redirect_maps:"]
            for k in sorted(updated_maps.keys()):
                val, comment = updated_maps[k]
                line_str = f"{inner_indent}{k}: {val}"
                if comment:
                    line_str += f" # {comment}"
                new_lines.append(line_str)
            
            # To preserve the rest of the file exactly as it is (comments, spacing, custom tags),
            # we calculate the exact line range where the 'redirect_maps' block lived.
            content_lines = raw_content.splitlines()
            
            # Calculate the line index where our regex match started.
            lines_before = raw_content[:match.start()].count("\n")
            
            # The replacement starts at the matched key line ('redirect_maps:')
            replacement_start_index = lines_before
            # The replacement ends after the key line + all existing entry lines we found.
            replacement_end_index = lines_before + 1 + block_lines_count
            
            # Splice in the new sorted and merged lines between the untouched top and bottom parts.
            new_content_lines = (
                content_lines[:replacement_start_index] + 
                new_lines + 
                content_lines[replacement_end_index:]
            )
            
            final_content = "\n".join(new_content_lines) + "\n"
            mkdocs_yml_path.write_text(final_content, "utf-8")
            print(f"  Surgically updated {mkdocs_yml_path} with {len(redirect_updates)} redirect rules.")
            return True

        # 2. If 'redirect_maps:' not found, try to find the 'redirects' plugin entry
        # Robust regex Breakdown:
        # ^(\s*)           -> Start of line, capture leading whitespace
        # -                -> Literal list dash
        # \s*              -> Internal whitespace
        # redirects(:)?    -> Literal 'redirects', with optional colon (handles string vs dict entries)
        # \s*              -> Whitespace
        # (?:#.*)?         -> Optional trailing comment
        # (\n|$|null)      -> End of line or null value
        match_plugin = re.search(r"^(\s*)-\s*redirects(:)?\s*(?:#.*)?(\n|$|null)", raw_content, re.MULTILINE)
        if match_plugin:
            indent = match_plugin.group(1)
            # Use 4 spaces for the redirect_maps key relative to the list item start
            plugin_key_indent = indent + "    " 
            inner_indent = plugin_key_indent + "  "
            
            new_block = [
                f"{indent}- redirects:",
                f"{plugin_key_indent}redirect_maps:"
            ]
            for k in sorted(redirect_updates.keys()):
                new_block.append(f"{inner_indent}{k}: {redirect_updates[k]}")
                
            content_lines = raw_content.splitlines()
            line_idx = raw_content[:match_plugin.start()].count("\n")
            
            new_content_lines = (
                content_lines[:line_idx] + 
                new_block + 
                content_lines[line_idx + 1:]
            )
            
            mkdocs_yml_path.write_text("\n".join(new_content_lines) + "\n", "utf-8")
            print(f"  Surgically added 'redirect_maps' to existing 'redirects' plugin in {mkdocs_yml_path}")
            return True

        # 3. If everything else fails, try to find 'plugins:' and add the block
        match_plugins = re.search(r"^plugins:\s*(\n|$)", raw_content, re.MULTILINE)
        if match_plugins:
            new_plugin_block = [
                "  - redirects:",
                "      redirect_maps:"
            ]
            for k in sorted(redirect_updates.keys()):
                new_plugin_block.append(f"        {k}: {redirect_updates[k]}")
            
            content_lines = raw_content.splitlines()
            plugins_line_idx = raw_content[:match_plugins.start()].count("\n")
            
            insert_idx = plugins_line_idx + 1
            for i, line in enumerate(content_lines[plugins_line_idx + 1:], plugins_line_idx + 1):
                if line.strip() and not line.startswith(" "):
                    insert_idx = i
                    break
                insert_idx = i + 1
                
            new_content_lines = content_lines[:insert_idx] + new_plugin_block + content_lines[insert_idx:]
            mkdocs_yml_path.write_text("\n".join(new_content_lines) + "\n", "utf-8")
            print(f"  Surgically added 'redirects' plugin block to {mkdocs_yml_path}")
            return True

        # 4. Fallback: Append plugins section
        print(f"  Warning: No plugins section found in {mkdocs_yml_path}. Appending new configuration.")
        new_config = "\nplugins:\n  - redirects:\n      redirect_maps:\n"
        for k in sorted(redirect_updates.keys()):
            new_config += f"        {k}: {redirect_updates[k]}\n"
        mkdocs_yml_path.write_text(raw_content.rstrip() + "\n" + new_config, "utf-8")
        return True
            
    except Exception as e:
        print(f"  Error updating {mkdocs_yml_path} surgically: {e}")
        return False


def build_id_map(docs_dir: Path) -> Dict[str, Path]:
    """
    Scans the documentation directory and builds a mapping between unique
    page IDs and their absolute file paths.

    Args:
        docs_dir: The directory containing documentation Markdown files.

    Returns:
        Dict[str, Path]: A dictionary mapping page IDs to their absolute file paths.
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
DOCS_DIR: Path = Path("site-src")
PAGE_ID_MAP_FILE: Path = Path("hack/page_id_map.json")
MKDOCS_YML_PATH: Path = Path("mkdocs.yml")
FRONTMATTER_ID_KEY: str = "id"


def get_frontmatter(content: str) -> Dict[str, str]:
    """Extracts frontmatter using python-frontmatter.

    Args:
        content: The full text content of a file.

    Returns:
        The metadata as a dictionary.
    """
    metadata = {}
    try:
        metadata, _ = frontmatter.parse(content)
    except Exception:
        pass  # Library failed, we'll try manual fallback below

    # If the library failed to find an ID (common in malformed YAML),
    # use our robust manual regex fallback.
    if FRONTMATTER_ID_KEY not in metadata:
        # Robust fallback: handle both closed blocks and unclosed start blocks
        match = re.search(r"^---\s*\n(.*?)(?:\n---|\Z)", content, re.DOTALL)
        if match:
            fm_text = match.group(1)
            id_match = re.search(r"^id:\s*(.+)$", fm_text, re.MULTILINE)
            if id_match:
                metadata[FRONTMATTER_ID_KEY] = id_match.group(1).strip().strip("\"'")
    
    return metadata


def format_frontmatter_str(data: Dict[str, str]) -> str:
    """
    Formats a dictionary into a YAML frontmatter string block.

    Args:
        data: The dictionary to format.

    Returns:
        str: The YAML formatted string, or an empty string if data is empty.
    """
    if not data:
        return ""
    return yaml.dump(data, default_flow_style=False, sort_keys=False, indent=2)
