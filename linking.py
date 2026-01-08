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


"""
MkDocs Documentation Migration Utility

This module provides CLI and plugin tools to support large-scale documentation
refactoring and robust internal linking for MkDocs sites.

Features:
        - Prepare documentation: Inject a unique permanent ID into the frontmatter
            of each Markdown file and generate a redirect map for safe file moves.
        - Convert internal links: Replace relative Markdown links with a macro-based
            internal_link for resilient cross-page linking.
        - MkDocs plugin hooks: Initialize the redirects plugin and generate redirect
            rules for moved files during the build process (without modifying docs).

Usage:
        Run as a CLI to prepare docs or convert links, or use as a plugin in mkdocs.yml.
"""

import argparse
import json
import os
import re
import yaml
from pathlib import Path
from typing import Any, Dict, List, Tuple, Optional


# --- Configuration ---
# Global constants defining key file paths and metadata keys.
DOCS_DIR: Path = Path("docs")
REDIRECT_MAP_FILE: Path = Path("hack/redirect_map.json")
FRONTMATTER_ID_KEY: str = "id"


def _get_frontmatter(content: str) -> Tuple[Dict[str, str], int]:
    """Extracts simple key-value frontmatter using only the standard library.

    Args:
        content: The full text content of a file.

    Returns:
        A tuple containing the frontmatter data as a dictionary and the
        character position where the frontmatter section ends.
    """
    # Step 1: Look for a YAML frontmatter block (e.g., ---...---) at the
    # beginning of a file.
    match = re.match(r"^---\s*\n(.*?\n)---\s*\n", content, re.DOTALL)
    if not match:
        return {}, 0

    frontmatter_str: str = match.group(1)
    end_pos: int = len(match.group(0))

    # Step 2: Parse the frontmatter block line-by-line for simple key-value pairs.
    data: Dict[str, str] = {}
    for line in frontmatter_str.strip().split("\n"):
        if ":" in line:
            key, value = line.split(":", 1)
            data[key.strip()] = value.strip()
    return data, end_pos


def _format_frontmatter_str(data: Dict[str, str]) -> str:
    """Formats a simple dictionary into a YAML-like string."""
    # Step 1: Convert dictionary to YAML-formatted lines.
    lines: List[str] = [f"{key}: {value}" for key, value in data.items()]
    return "\n".join(lines) + "\n"


def prepare_docs(docs_dir_path: Optional[str] = None) -> None:
    """
    Scan the documentation directory for Markdown files, inject a unique permanent
    ID into the frontmatter of each file if missing, and create a redirect map
    (JSON) of page IDs to file paths. This prepares the docs for safe refactoring
    and enables robust internal linking and redirect generation.

    Args:
        docs_dir_path: Optional path to the documentation directory. If None,
            uses the default DOCS_DIR.

    Returns:
        None. Writes changes to Markdown files and creates/updates the redirect
        map JSON file on disk.
    """
    print("Starting documentation preparation...")
    redirect_map: Dict[str, str] = {}

    # Use global variables for when no path specified
    if docs_dir_path is None:
        docs_dir = DOCS_DIR
        redirect_map_file = REDIRECT_MAP_FILE
    else:
        docs_dir = Path(docs_dir_path)
        redirect_map_file = Path(docs_dir_path).parent / "redirect_map.json"

    # Step 1: Iterate through all markdown files in the documentation directory.
    for md_file in docs_dir.rglob("*.md"):
        # Step 2: Read file content and extract existing frontmatter.
        content: str = md_file.read_text("utf-8")
        frontmatter, end_pos = _get_frontmatter(content)
        page_id: str | None = frontmatter.get(FRONTMATTER_ID_KEY)

        # Step 3: If no permanent ID exists, generate one from the file path,
        # inject it into the frontmatter, and rewrite the file.
        if not page_id:
            relative_path: Path = md_file.relative_to(docs_dir)
            page_id = str(relative_path.with_suffix("")).replace(os.path.sep, "-")
            print(f"  - Assigning ID '{page_id}' to {md_file}")

            frontmatter[FRONTMATTER_ID_KEY] = page_id
            new_frontmatter_str = _format_frontmatter_str(frontmatter)

            body: str = content[end_pos:]
            new_content: str = f"---\n{new_frontmatter_str}---\n{body}"
            md_file.write_text(new_content, "utf-8")
        else:
            print(f"  - Found existing ID '{page_id}' in {md_file}")

        # Step 4: Add the page's ID and current path to our map. This captures
        # the "before" state of the documentation.
        redirect_map[page_id] = str(md_file.relative_to(docs_dir))

    # Step 5: Write the completed map to a JSON file for persistence. This file
    # will be used by the MkDocs hook to generate redirects.
    redirect_map_file.write_text(json.dumps(redirect_map, indent=2))
    print(f"Preparation complete. Map saved to {redirect_map_file}")


def convert_internal_links(docs_dir_path: str = "docs"):
    """
    Converts all relative Markdown links in a documentation directory to use an internal link macro.
    This function scans all Markdown (.md) files within the specified documentation directory,
    builds a mapping of file paths to unique page IDs (as defined in each file's frontmatter),
    and then replaces any relative Markdown links to other .md files with a macro of the form:
    [Link Text]({{ internal_link("page_id") }})
    External links, anchor links, mailto links, and links to non-Markdown files are ignored.
    If a link points to a Markdown file that does not have a page ID in its frontmatter,
    the link is left unchanged.
    The function modifies files in-place and prints a summary of the conversion process,
    including the number of files modified.
    Args:
        docs_dir_path (str): Path to the root documentation directory containing Markdown files.
                             Defaults to 'docs'.
    Side Effects:
        - Reads and writes Markdown files in the specified directory.
        - Prints progress and summary information to stdout.
    Requirements:
        - Each Markdown file should have a unique page ID in its frontmatter under the key FRONTMATTER_ID_KEY.
        - The function assumes the existence of helper functions and constants:
            - _get_frontmatter(content): Extracts frontmatter from file content.
            - FRONTMATTER_ID_KEY: The key used to retrieve the page ID from frontmatter.
    Example:
        convert_internal_links('docs')
    """
    print("Starting internal link conversion...")
    docs_dir = Path(docs_dir_path)

    # Step 1: Build a map of file paths to their unique IDs.
    # This is more efficient than reading the target file for every link.
    path_to_id_map: Dict[str, str] = {}
    for md_file in docs_dir.rglob("*.md"):
        content = md_file.read_text("utf-8")
        frontmatter, _ = _get_frontmatter(content)
        page_id = frontmatter.get(FRONTMATTER_ID_KEY)
        if page_id:
            # Use a normalized posix path relative to the docs root as the key
            relative_path_key = md_file.relative_to(docs_dir).as_posix()
            path_to_id_map[relative_path_key] = page_id

    print(f"  - Built a map of {len(path_to_id_map)} page IDs.")

    # Step 2: Iterate through each file and replace its links.
    files_converted = 0
    for md_file in docs_dir.rglob("*.md"):
        content = md_file.read_text("utf-8")
        original_content = content

        # This nested function (closure) captures the current file's context.
        def replace_link(match: re.Match) -> str:
            link_text = match.group(1)
            link_url = match.group(2)

            # Ignore external links, anchors, or non-markdown file links
            if link_url.startswith(("http", "#", "mailto:")) or not link_url.endswith(
                ".md"
            ):
                return match.group(0)

            # Resolve the relative path to an absolute path from the docs root
            current_dir = md_file.parent
            target_file = (current_dir / link_url).resolve()

            # Make the path relative to the docs dir to use as a lookup key
            try:
                target_relative_path = target_file.relative_to(docs_dir.resolve())
                target_key = target_relative_path.as_posix()
            except ValueError:
                # This can happen if the link points outside the docs directory
                return match.group(0)

            # Look up the ID in our map
            target_id = path_to_id_map.get(target_key)
            if target_id:
                # If an ID is found, build the macro
                return f'[{link_text}]({{{{ internal_link("{target_id}") }}}})'
            else:
                # If no ID is found (e.g., broken link), leave it as is
                return match.group(0)

        # Use re.sub with our replacer function to process all links
        # Regex explanation:
        # \[([^\]]+)\]   - Capture the link text inside [ ]
        # \((?!{{)([^)]+)\) - Capture the URL inside ( ), but negative lookahead
        #                    to avoid re-processing our own macros.
        content = re.sub(r"\[([^\]]+)\]\((?!{{)([^)]+\.md)\)", replace_link, content)

        if content != original_content:
            files_converted += 1
            md_file.write_text(content, "utf-8")
            print(f"  - Converted links in {md_file.relative_to(docs_dir)}")

    print(f"Link conversion complete. Modified {files_converted} files.")


def on_config(config: Dict[str, Any]) -> Dict[str, Any]:
    """
    MkDocs plugin hook for the config phase. Ensures the redirects plugin is
    properly initialized in the MkDocs config, supporting both dict and object
    plugin representations. Does not modify files on disk.

    Args:
        config: The MkDocs configuration dictionary.

    Returns:
        The updated MkDocs configuration dictionary with the redirects plugin
        initialized if present.
    """
    print("Running MkDocs migration hook (config phase)...")
    for plugin_name, plugin_instance in config.get("plugins", {}).items():
        if plugin_name == "redirects":
            # Support both dict and object plugin representations
            if isinstance(plugin_instance, dict):
                if "config" not in plugin_instance:
                    plugin_instance["config"] = {}
                if "redirect_maps" not in plugin_instance["config"]:
                    plugin_instance["config"]["redirect_maps"] = {}
            else:
                # Assume object with .config attribute
                if not hasattr(plugin_instance, "config"):
                    plugin_instance.config = type("Config", (), {})()
                if not hasattr(plugin_instance.config, "redirect_maps"):
                    plugin_instance.config.redirect_maps = {}  # type: ignore
            print("  Redirects plugin configured.")
            break
    return config


def _update_mkdocs_yml_redirects(redirect_updates: Dict[str, str]) -> bool:
    """
    Updates the 'mkdocs.yml' configuration file with new redirect rules using safe YAML parsing.
    This function ensures that the 'redirects' plugin and its 'redirect_maps' section exist in the
    MkDocs configuration file. It merges the provided redirect rules (mapping old paths to new paths)
    with any existing rules, preserving the overall structure and formatting of the YAML file as much
    as possible. The function handles both string and dictionary plugin formats, and will create the
    necessary sections if they are missing.
        redirect_updates (Dict[str, str]):
            A dictionary mapping old documentation paths (as keys) to new paths (as values).
            Each entry represents a redirect rule to be added or updated in the configuration.
        bool:
            True if the 'mkdocs.yml' file was successfully updated with new or changed redirect rules.
            False if no changes were needed, if the file does not exist, or if an error occurred.
    Raises:
        None explicitly, but prints warnings or errors to the console if:
            - The 'mkdocs.yml' file is missing or not a valid YAML dictionary.
            - There are issues parsing or writing the YAML file.
    Notes:
        - The function uses PyYAML for parsing and writing YAML.
        - Existing comments and formatting may not be fully preserved due to PyYAML limitations.
        - The function is idempotent: if the provided redirects are already present, no changes are made.
        - If the 'redirects' plugin is not present, it is added in the correct format.
        - If the 'redirect_maps' section is missing, it is created.
        - The function prints informative messages about its actions and any issues encountered.
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

        # Step 8: Handle case where `redirect_maps:` is present but empty (evaluates to None)
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


def on_files(files, config):
    """
    Handles the MkDocs 'files' plugin event to manage page redirects and internal linking.
    This function performs the following steps:
    1. Loads a mapping of page IDs to their previous file paths from a JSON file (REDIRECT_MAP_FILE).
    2. Scans the current set of Markdown files in the documentation, extracting their page IDs from frontmatter,
       and builds a mapping of page IDs to their new file paths.
    3. Compares the old and new mappings to detect moved or renamed pages, and generates redirect rules for any
       pages whose paths have changed.
    4. Attempts to update the 'mkdocs.yml' configuration file with the new redirect rules for use with the
       mkdocs-redirects plugin. If automatic updating fails, prints instructions for manual addition.
    5. Prints debug and status information throughout the process.
    Args:
        files (list): List of MkDocs file objects representing the documentation files.
        config (dict): MkDocs configuration dictionary.
    Returns:
        list: The (possibly unmodified) list of file objects, to be passed along in the MkDocs build process.
    Side Effects:
        - Reads from REDIRECT_MAP_FILE to obtain previous page mappings.
        - May write to 'mkdocs.yml' to update redirect rules.
        - Prints status and debug information to the console.
    Notes:
        - This function assumes the existence of certain global constants and helper functions:
            - REDIRECT_MAP_FILE: Path to the JSON file containing the old page ID to path mapping.
            - FRONTMATTER_ID_KEY: The key in the frontmatter that uniquely identifies a page.
            - _get_frontmatter: Function to extract frontmatter from a Markdown file.
            - _update_mkdocs_yml_redirects: Function to update mkdocs.yml with new redirect rules.
        - The function is designed to be used as a plugin hook in the MkDocs build process.
    """
    """Generates redirects and sets up the internal link macro for MkDocs."""
    print("Running MkDocs migration hook (files phase)...")

    # Step 1: Load the "before" state map
    if not REDIRECT_MAP_FILE.exists():
        print(f"  Warning: {REDIRECT_MAP_FILE} not found. Skipping.")
        return files

    old_paths_map: Dict[str, str] = json.loads(REDIRECT_MAP_FILE.read_text())

    # Step 2: Build the "after" state map
    after_paths_map: Dict[str, str] = {}

    print(
        f"  Debug: MkDocs found these files: {[f.src_path for f in files if f.src_path.endswith('.md')]}"
    )

    for file in files:
        if file.src_path.endswith(".md"):
            abs_path = Path(config["docs_dir"]) / file.src_path
            if abs_path.exists():
                content = abs_path.read_text("utf-8")
                frontmatter, _ = _get_frontmatter(content)
                page_id = frontmatter.get(FRONTMATTER_ID_KEY)

                if page_id:
                    after_paths_map[page_id] = file.src_path

    # Step 3: Generate redirect rules and write them to mkdocs.yml
    # (since programmatic config updates don't seem to work)
    redirect_updates = {}
    count = 0
    for page_id, old_path in old_paths_map.items():
        new_path = after_paths_map.get(page_id)
        if new_path and new_path != old_path:
            # Only create redirect for the .md file path
            # The mkdocs-redirects plugin handles URL generation automatically
            redirect_updates[old_path] = new_path
            print(f"    Would add redirect: {old_path} -> {new_path}")
            count += 1

    if count > 0:
        print(f"  Generated {count} redirect rules.")
        print(f"  Redirect rules: {redirect_updates}")

        # Try to update mkdocs.yml automatically
        if _update_mkdocs_yml_redirects(redirect_updates):
            print("  ✓ Updated mkdocs.yml with redirect rules.")
        else:
            print("  ✗ Could not update mkdocs.yml automatically.")
            print("  Manual addition required:")
            print("  plugins:")
            print("    - redirects:")
            print("        redirect_maps:")
            for old, new in redirect_updates.items():
                print(f"          {old}: {new}")
    else:
        print("  No new redirects needed.")

    print("  `internal_link` macro is ready (provided by main.py).")
    return files


def preview_docs(docs_dir_path: str = "docs") -> None:
    """
    Simulates the process of preparing documentation files for unique page IDs without making any changes.
    This function scans the specified documentation directory (default: 'docs') for Markdown (.md) files.
    It analyzes each file to determine whether it already contains a unique page ID in its frontmatter.
    For files missing an ID, it shows what ID would be generated based on the file's relative path.
    For files with existing IDs, it lists them and their IDs.
    No files are modified; this is a dry-run/preview mode.
    The function also indicates where a redirect map (redirect_map.json) would be created or updated
    if changes were to be applied.
    Args:
        docs_dir_path (str): Path to the documentation directory to scan. Defaults to 'docs'.
    Prints:
        - The number of Markdown files found.
        - A list of files that would be modified with new IDs, showing the generated ID for each.
        - A list of files that already have IDs, showing the existing ID for each.
        - The path where the redirect map would be created or updated.
    Notes:
        - This function does not modify any files or create the redirect map.
        - Errors encountered while reading files are reported, but do not stop the process.
        - Requires the existence of helper functions/constants: _get_frontmatter and FRONTMATTER_ID_KEY.
    """
    """Shows what the prepare_docs function would do without making changes."""
    docs_dir = Path(docs_dir_path)

    print(f"Scanning documentation directory: {docs_dir}")

    # Step 1: Validate that the documentation directory exists.
    if not docs_dir.exists():
        print(f"  ERROR: Directory {docs_dir} does not exist")
        return

    # Step 2: Find all markdown files in the documentation directory.
    md_files = list(docs_dir.rglob("*.md"))
    if not md_files:
        print("  No markdown files found")
        return

    print(f"  Found {len(md_files)} markdown files")

    files_needing_ids = []
    files_with_ids = []

    # Step 3: Analyze each markdown file to determine its current state.
    for md_file in md_files:
        try:
            content = md_file.read_text("utf-8")
            frontmatter, _ = _get_frontmatter(content)
            page_id = frontmatter.get(FRONTMATTER_ID_KEY)

            if page_id:
                files_with_ids.append((md_file, page_id))
            else:
                relative_path = md_file.relative_to(docs_dir)
                generated_id = str(relative_path.with_suffix("")).replace(
                    os.path.sep, "-"
                )
                files_needing_ids.append((md_file, generated_id))
        except Exception as e:
            print(f"  Warning: Could not process {md_file}: {e}")

    # Step 4: Display files that would be modified with new IDs.
    if files_needing_ids:
        print(f"\nFiles that would be modified ({len(files_needing_ids)}):")
        for md_file, generated_id in files_needing_ids:
            relative_path = md_file.relative_to(docs_dir)
            print(f"  + {relative_path} -> ID: '{generated_id}'")

    # Step 5: Display files that already have IDs and would be preserved.
    if files_with_ids:
        print(f"\nFiles already with IDs ({len(files_with_ids)}):")
        for md_file, existing_id in files_with_ids:
            relative_path = md_file.relative_to(docs_dir)
            print(f"  * {relative_path} -> ID: '{existing_id}'")

    # Step 6: Show where the redirect map would be created or updated.
    print(
        f"\nWould create/update redirect map: {Path(docs_dir_path).parent / 'redirect_map.json'}"
    )


def main() -> None:
    """Parses command line arguments and runs the preparation script."""
    parser = argparse.ArgumentParser(
        description="MkDocs migration helper - prepares docs for safe refactoring.",
        prog="linking",
    )
    parser.add_argument(
        "--prepare",
        action="store_true",
        help="Scan docs folder, inject IDs, and create redirect map.",
    )
    parser.add_argument(
        "--convert-links",
        action="store_true",
        help="Convert all relative Markdown links to the internal_link macro.",
    )
    parser.add_argument(
        "--docs-dir", default="docs", help="Documentation directory (default: docs)."
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without making changes.",
    )
    args = parser.parse_args()

    if args.prepare:
        if args.dry_run:
            print("DRY RUN: Preview of changes that would be made")
            print("=" * 50)
            preview_docs(args.docs_dir)
        else:
            prepare_docs(args.docs_dir)
    elif args.convert_links:
        if args.dry_run:
            print(
                "DRY RUN for link conversion is not implemented. This action directly modifies files."
            )
        else:
            convert_internal_links(args.docs_dir)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
