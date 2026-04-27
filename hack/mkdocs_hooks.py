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
Dedicated module for MkDocs build lifecycle hooks dependent on mkdocs_utils for logic.
"""

import json
from pathlib import Path
from typing import Any, Dict
import mkdocs_utils

# --- Configuration (Exposed for compatibility) ---
PAGE_ID_MAP_FILE = mkdocs_utils.PAGE_ID_MAP_FILE
FRONTMATTER_ID_KEY = mkdocs_utils.FRONTMATTER_ID_KEY


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
    plugins = config.get("plugins")
    if not plugins:
        return config

    # Handled differently depending on whether plugins is a list (config phase) 
    # or a PluginCollection (later phases, though hooks usually see the former).
    if isinstance(plugins, dict) and "redirects" in plugins:
        plugin_instance = plugins["redirects"]
        # Support both dict and object plugin representations
        if isinstance(plugin_instance, dict):
            if "config" not in plugin_instance:
                plugin_instance["config"] = {}
            if "redirect_maps" not in plugin_instance["config"]:
                plugin_instance["config"]["redirect_maps"] = {}
        elif hasattr(plugin_instance, "config"):
            # If it's an object, it should already have a config object from MkDocs.
            # We just ensure redirect_maps is present if it's a known config object.
            if not hasattr(plugin_instance.config, "redirect_maps"):
                try:
                    plugin_instance.config.redirect_maps = {}
                except (AttributeError, TypeError):
                    # If we can't set it, it's likely not the object we expect.
                    pass
        print("  Redirects plugin configuration checked.")
    return config


def on_files(files, config):
    """
    Handles the MkDocs 'files' plugin event to manage page redirects and internal linking.
    """
    print("Running MkDocs migration hook (files phase)...")

    # Step 1: Load the "before" state map
    if not PAGE_ID_MAP_FILE.exists():
        print(f"  Warning: {PAGE_ID_MAP_FILE} not found. Skipping.")
        return files

    try:
        old_paths_map: Dict[str, str] = json.loads(PAGE_ID_MAP_FILE.read_text())
    except Exception as e:
        print(f"  Warning: Could not load page ID map: {e}")
        return files

    # Step 2: Build the "after" state map
    after_paths_map: Dict[str, str] = {}

    for file in files:
        if file.src_path.endswith(".md"):
            abs_path = Path(config["docs_dir"]) / file.src_path
            if abs_path.exists():
                try:
                    content = abs_path.read_text("utf-8")
                    frontmatter = mkdocs_utils.get_frontmatter(content)
                    page_id = frontmatter.get(FRONTMATTER_ID_KEY)

                    if page_id:
                        after_paths_map[page_id] = file.src_path
                except Exception as e:
                    print(f"  Warning: Could not process file in hook: {file.src_path}: {e}")

    # Step 3: Generate redirect rules and write them to mkdocs.yml
    redirect_updates = {}
    count = 0
    for page_id, old_path in old_paths_map.items():
        new_path = after_paths_map.get(page_id)
        if new_path and new_path != old_path:
            redirect_updates[old_path] = new_path
            print(f"    Would add redirect: {old_path} -> {new_path}")
            count += 1

    if count > 0:
        print(f"  Generated {count} redirect rules.")
        if mkdocs_utils.update_mkdocs_yml_redirects(redirect_updates):
            print("  ✓ Updated mkdocs.yml with redirect rules.")
        else:
            print("  ✗ Could not update mkdocs.yml automatically.")
    else:
        print("  No new redirects needed.")

    return files
