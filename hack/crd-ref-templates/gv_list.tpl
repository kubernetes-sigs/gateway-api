{{- define "gvList" -}}
{{- $groupVersions := . -}}
---
title: "API Reference"
linkTitle: LINK_TITLE_PLACEHOLDER
weight: WEIGHT_PLACEHOLDER
---

{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}