{{- define "gvList" -}}
{{- $groupVersions := . -}}
---
hide:
- toc
---
# API Reference

{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}