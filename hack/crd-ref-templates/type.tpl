{{- define "unionFeatureLinks" -}}
{{- $doc := . -}}
{{- $tags := regexFindAll "<gateway:union:[A-Za-z]+>" $doc -1 -}}
{{- range $i, $tag := $tags -}}
{{- if $i }}, {{ end -}}
{{- $name := trimSuffix ">" (trimPrefix "<gateway:union:" $tag) -}}
[{{ $name }}](#{{ lower $name }})
{{- end -}}
{{- end -}}

{{- define "type" -}}
{{- $type := . -}}
{{- if markdownShouldRenderType $type -}}

#### {{ $type.Name }}

{{ if $type.IsAlias }}_Underlying type:_ _{{ markdownRenderTypeLink $type.UnderlyingType  }}_{{ end }}

{{ $type.Doc }}

{{ if contains "<gateway:union:" $type.Doc -}}
:link: **Union Feature**: works with {{ template "unionFeatureLinks" $type.Doc }}

{{ end -}}
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
| Field | Description | Default | Validation |
| --- | --- | --- | --- |
{{ if $type.GVK -}}
| `apiVersion` _string_ | `{{ $type.GVK.Group }}/{{ $type.GVK.Version }}` | | |
| `kind` _string_ | `{{ $type.GVK.Kind }}` | | |
{{ end -}}

{{ range $type.Members -}}
| `{{ .Name  }}` _{{ markdownRenderType .Type }}_ {{- if contains "<gateway:experimental>" .Doc -}}<br /> :warning: **Experimental**{{ end -}}{{- if contains "<gateway:union:" .Doc -}}<br /> :link: **Union Feature**: works with {{ template "unionFeatureLinks" .Doc }}{{ end -}}| {{ template "type_members" . }} | {{ markdownRenderDefault .Default }} | {{ range .Validation -}} {{ markdownRenderFieldDoc . }} <br />{{ end }} |
{{ end -}}

{{ end -}}

{{ if $type.EnumValues -}} 
| Field | Description |
| --- | --- |
{{ range $type.EnumValues -}}
| `{{ .Name }}` | {{ markdownRenderFieldDoc .Doc }} |
{{ end -}}
{{ end -}}


{{- end -}}
{{- end -}}