"""
Main module for mkdocs-macros plugin.
This provides the internal_link macro for resilient documentation linking.
"""

import re
from pathlib import Path
from typing import Dict, Optional


class PageResolver:
    """Handles page ID resolution and link generation."""
    
    def __init__(self, docs_dir: Path = Path('docs')):
        self.docs_dir = docs_dir
        self._page_cache: Optional[Dict[str, Path]] = None
    
    def _extract_frontmatter_id(self, content: str) -> Optional[str]:
        """Extract page ID from YAML frontmatter."""
        if not content.startswith('---'):
            return None
        
        match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
        if not match:
            return None
        
        frontmatter = match.group(1)
        for line in frontmatter.split('\n'):
            if ':' in line:
                key, value = line.split(':', 1)
                if key.strip() == 'id':
                    return value.strip()
        return None
    
    def _build_page_cache(self) -> Dict[str, Path]:
        """Build a cache of page ID to file path mappings."""
        cache = {}
        for md_file in self.docs_dir.rglob('*.md'):
            try:
                content = md_file.read_text('utf-8')
                page_id = self._extract_frontmatter_id(content)
                if page_id:
                    cache[page_id] = md_file
            except (OSError, UnicodeDecodeError):
                continue  # Skip files that can't be read
        return cache
    
    def resolve_page_link(self, page_id: str, current_page_path: Optional[str] = None) -> str:
        """Resolve a page ID to its Markdown file reference, relative to current page."""
        import os
        # Build cache on first use
        if self._page_cache is None:
            self._page_cache = self._build_page_cache()

        file_path = self._page_cache.get(page_id)
        if not file_path:
            raise ValueError(f"Page with ID '{page_id}' not found")

        target_path = file_path.relative_to(self.docs_dir)

        # If no current page context, return absolute path from docs root
        if not current_page_path:
            return str(target_path)

        # Calculate relative path from current page to target page
        current_path = Path(current_page_path)
        current_dir = current_path.parent
        rel_path = os.path.relpath(str(target_path), str(current_dir))
        return rel_path.replace(os.path.sep, "/")


def define_env(env):
    """Hook for mkdocs-macros plugin functions and variables."""
    
    resolver = PageResolver()
    
    @env.macro
    def internal_link(page_id: str) -> str:
        """
        Looks up a page by ID and returns its Markdown file reference.
        This provides resilient linking that survives file moves.
        """
        try:
            # Get current page context from mkdocs-macros environment
            current_page_path = None
            if hasattr(env, 'variables') and env.variables:
                page = env.variables.get('page')
                if page and hasattr(page, 'file') and hasattr(page.file, 'src_path'):
                    current_page_path = page.file.src_path
            
            return resolver.resolve_page_link(page_id, current_page_path)
        except Exception:
            # Fallback: use resolver directly
            try:
                return resolver.resolve_page_link(page_id, None)
            except ValueError:
                return f"[LINK ERROR: Page '{page_id}' not found]"
