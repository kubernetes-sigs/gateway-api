{{- define "gvList" -}}
{{- $groupVersions := . -}}

# API Reference

<p>Packages:</p>
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
