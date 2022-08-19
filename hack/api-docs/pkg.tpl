{{ define "link" }}
<li>
    <a href="#{{- packageAnchorID . -}}">{{ packageDisplayName . }}</a>
</li>
{{ end }}

{{ define "package" }}
<h2 id="{{- packageAnchorID . -}}">
    {{- packageDisplayName . -}}
</h2>

{{ with (index .GoPackages 0 )}}
    {{ with .DocComments }}
    <p>
        {{ safe (renderComments .) }}
    </p>
    {{ end }}
{{ end }}

Resource Types:
<ul>
{{- range (visibleTypes (sortedTypes .Types)) -}}
    {{ if isExportedType . -}}
    <li>
        <a href="{{ linkForType . }}">{{ typeDisplayName . }}</a>
    </li>
    {{- end }}
{{- end -}}
</ul>

{{ range (visibleTypes (sortedTypes .Types))}}
    {{ template "type" .  }}
{{ end }}
<hr/>
{{ end }}


{{ define "packages" }}

{{/* we manually iterate the packages slice so that we feature beta before alpha */}}

<p>Packages:</p>
<ul>
    {{ template "link" index .packages 1 }}
    {{ template "link" index .packages 0 }}
</ul>

{{ template "package" index .packages 1 }}
{{ template "package" index .packages 0 }}

<p><em>
    Generated with <code>gen-crd-api-reference-docs</code>.
</em></p>

{{ end }}
