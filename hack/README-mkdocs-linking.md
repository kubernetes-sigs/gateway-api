# SEO Resilience Toolkit

A suite of tools designed to make the Gateway API documentation resilient to refactoring, file moves, and renames. The system is split into four focused modules:

1.  mkdocs_utils.py: The core engine. Contains all logic for ID generation, frontmatter parsing, link conversion, and YAML updates.
2.  mkdocs_linking.py: A CLI wrapper for manual maintenance tasks (preparing docs and converting links).
3.  mkdocs_main.py: The MkDocs macros plugin that provides the `internal_link` macro.
4.  mkdocs_hooks.py: Lifecycle hooks for automatic redirect generation (currently disabled).

---

## Core Features

### 1. Stable Identification (ID Injection)
The system ensures every Markdown file has a unique, permanent ID in its frontmatter. If a file is missing an ID, the `prepare` script generates one based on its path and injects it into the frontmatter. The IDs allow the system to track a file even if its filename or directory changes.

### 2. Resilient Linking (`internal_link` macro)
Instead of hardcoding relative paths like `[text](../guides/api.md)`, the system allows using stable IDs:
`[text]({{ internal_link("guides-api") }})` The macro resolves the ID to the correct current path at build time.

### 3. Automated Link Conversion
The toolkit includes a script to bulk-convert existing standard Markdown links into the resilient macro format with a masking strategy to ensure links inside code blocks or backticks are never touched.

### 4. Automatic Redirect Management
When files move, the system detects the change via the stable IDs and updates the `redirects` plugin in `mkdocs.yml`.

---

## How the Algorithm Works

The system uses a Stable ID approach to track files regardless of their location in the directory structure.

### 1. Stable Identification
Each documentation page is assigned a unique page_id in its frontmatter:
```markdown
---
id: concept-api-overview
---
```

### 2. State Mapping
The algorithm maintains two "states":
-   Before State: A snapshot stored in hack/page_id_map.json which maps every page_id to its known file path (e.g., concepts/api.md).
-   After State: A real-time scan of the site-src/ directory, identifying the current file path for every page_id.

### 3. Change Detection & Rule Generation
By comparing these two states, the script identifies three scenarios:
-   Move: If a page_id exists in both states but the path has changed, a redirect rule is generated: old/path.md -> new/path.md.
-   Rename: Handled the same as a move; as long as the ID is stable, the path change is detected.
-   New/Deleted: New IDs are added to the map; deleted IDs are ignored (or kept in the map for legacy redirects).

### 4. Automatic Configuration Patching
Once rules are generated, the script uses a YAML-aware parser to locate the redirects plugin section in mkdocs.yml and inject the new rules without disturbing other configurations.

---

## Current Status

Note: The hook is currently disabled in the root mkdocs.yml.

It was disabled to prevent destructive mutations to mkdocs.yml during active development and CI runs where side effects are undesirable.

## How to Enable

To re-enable the hook, add `- hack/mkdocs_hooks.py` to the `hooks` list in `mkdocs.yml`:

```yaml
hooks:
  - hack/mkdocs-copy-geps.py
  - hack/mkdocs-generate-conformance.py
  - hack/mkdocs_hooks.py  # Add this line
  - hack/mkdocs-generate-controller-wizard-data.py
```

### 2. Update Redirects Warning
When re-enabling, you should also re-add the warning comment above the `redirects:` plugin section in `mkdocs.yml` to prevent manual entries that will be overwritten:

```yaml
  - macros:
      include_dir: examples
      module_name: hack/mkdocs_main
  # Do not add manual redirects here, they will be overwritten. Add them to 
  # hack/redirect_map.json if needed in exceptional cases, but the mkdocs_linking.py
  # hook and accompanying macro largely negate the need for manual redirects:
  # https://github.com/kubernetes-sigs/gateway-api/pull/3999
  - redirects:
      redirect_maps:
        ...
```

---

## How to Use Manually

The `hack/mkdocs_linking.py` script is your primary interface for maintenance. It is the recommended way to test changes before enabling the hook.

### 1. Safety First: Dry Run
To see what the script would do without actually modifying any files, use the `--dry-run` flag:
```bash
PYTHONPATH=hack python3 hack/mkdocs_linking.py --prepare --dry-run
```

### 2. Prepare Documentation
Scans `site-src/`, injects missing IDs, and updates `hack/page_id_map.json`.
```bash
PYTHONPATH=hack python3 hack/mkdocs_linking.py --prepare
```

### 3. Convert Internal Links
Transforms standard `.md` links into `internal_link` macros across the whole project.
```bash
PYTHONPATH=hack python3 hack/mkdocs_linking.py --convert-links
```

---

## Running Tests

To verify the logic (CLI, link conversion, ID generation, regex patching), run the test suite:

### 1. Using Pytest (Recommended)
```bash
PYTHONPATH=hack pytest hack/mkdocs/__tests__/
```

### 2. Using Unittest
```bash
PYTHONPATH=hack python3 -m unittest discover -s hack/mkdocs/__tests__/ -p 'test_*.py'
```

---

## Troubleshooting & Verification

If you get import errors, ensure you are running from the root of the repository and that `hack` is in your `PYTHONPATH`:
```bash
PYTHONPATH=hack python3 hack/mkdocs_linking.py --prepare
```

## Prerequisites

This toolkit requires Python 3.9+ and several dependencies.

### 1. Dependencies
Install the required libraries using the repository's requirements file:
```bash
pip install -r hack/mkdocs/image/requirements.txt
```

### 2. Using a Virtual Environment
If you are using the virtual environment provided in the repository, prefix your commands with the environment path:
```bash
PYTHONPATH=hack ./.venv/bin/python3 hack/mkdocs_linking.py --prepare
```

### 3. Page ID Map
- `hack/page_id_map.json`: This file contains the authoritative mapping of page IDs to their original paths. It must exist (even if empty) for the scripts to run.
- Frontmatter IDs: Markdown files must have an `id:` field (e.g., `id: geps-101`) to be tracked.