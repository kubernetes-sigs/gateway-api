import frontmatter
content = "---\nid: manual-id\nkey: : broken: : syntax\nContent"
try:
    metadata, content_text = frontmatter.parse(content)
    print(f"Metadata: {metadata}")
except Exception as e:
    print(f"Error: {e}")
