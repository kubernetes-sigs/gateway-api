{{- define "type_members" -}}
{{- $field := . -}}
{{- if eq $field.Name "metadata" -}}
Refer to Kubernetes API documentation for fields of `metadata`.
{{- else -}}
{{- $cleanDoc := regexReplaceAllLiteral "(?s)<gateway:util:excludeFromCRD>.*?</gateway:util:excludeFromCRD>" $field.Doc "" -}}
{{ markdownRenderFieldDoc $cleanDoc }}
{{- end -}}
{{- end -}}