{{/*
Create a default fullname using the release name and the chart name.
*/}}
{{- define "vifunction-controller.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Create a name with the chart name.
*/}}
{{- define "vifunction-controller.name" -}}
{{- printf "%s" .Chart.Name -}}
{{- end }}

{{/*
Create a chart name and version.
*/}}
{{- define "vifunction-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end }}