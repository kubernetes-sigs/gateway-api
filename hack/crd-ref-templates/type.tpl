{{- define "type" -}}
{{- $type := . -}}
{{- if markdownShouldRenderType $type -}}

#### {{ $type.Name }}

{{ if $type.IsAlias }}_Underlying type:_ _{{ markdownRenderTypeLink $type.UnderlyingType  }}_{{ end }}

{{ $type.Doc }}

{{ if $type.Validation -}}
_Validation:_
{{- range $type.Validation }}
- {{ . }}
{{- end }}
{{- end }}

{{ if $type.References -}}
_Appears in:_
{{ range $type.SortedReferences }}
- {{ markdownRenderTypeLink . }}
{{- range .Fields -}}
{{- if eq .Type.Name $type.Name -}}
{{- if contains "<gateway:experimental>" .Doc }} :warning: Experimental in `{{.Name}}` field {{- end }}
{{- end }}
{{- end }}
{{ end }}
{{- end }}

{{ if $type.Members -}}
| Field | Description |
| --- | --- |
{{ if $type.GVK -}}
| `apiVersion` _string_ | `{{ $type.GVK.Group }}/{{ $type.GVK.Version }}` |
| `kind` _string_ | `{{ $type.GVK.Kind }}` |
{{ end -}}

{{ range $type.Members -}}
| `{{ .Name  }}`<br />_{{ markdownRenderType .Type }}_{{- if contains "<gateway:experimental>" .Doc -}}<br /> :warning: **Experimental**{{ end -}}{{- if .Default }}<p>Default: {{ markdownRenderDefault .Default }}</p>{{ end -}}<p>**Validations:**<br />{{ range .Validation -}}{{ markdownRenderValidation . }}<br />{{ end }}</p>| {{ template "type_members" . }} |
{{ end -}}

{{ end -}}

{{ if $type.EnumValues -}} 
| Enum Value | Description |
| --- | --- |
{{ range $type.EnumValues -}}
| `{{ .Name }}` | {{ markdownRenderFieldDoc .Doc }} |
{{ end -}}
{{ end -}}


{{- end -}}
{{- end -}}