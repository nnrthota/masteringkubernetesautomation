
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "tweet-app.fullname" -}}
{{- default .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tweet-app.selectorLabels" -}}
name: "tweet-pod"
app: {{ include "tweet-app.fullname" . }}
{{- end }}


{{/*
Selector Pod labels
*/}}
{{- define "tweet-app.podLabels" -}}
name: "tweet-pod"
app: {{ include "tweet-app.fullname" . }}
{{- end }}
