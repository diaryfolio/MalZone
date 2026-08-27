{{- define "malzone.name" -}}
malzone
{{- end }}

{{- define "malzone.labels" -}}
app.kubernetes.io/name: {{ include "malzone.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end }}

{{- define "malzone.image" -}}
{{ printf "%s:%s" .Values.image.repository .Values.image.tag }}
{{- end }}
