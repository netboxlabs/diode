{{/*
Define redis host
*/}}
{{- define "diode.redis.host" -}}
{{- if .Values.redis.enabled -}}
{{- printf "%s-redis-master.%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- else -}}
{{- .Values.externalRedis.host -}}
{{- end -}}
{{- end -}}

{{/*
Define redis port
*/}}
{{- define "diode.redis.port" -}}
{{- if .Values.redis.enabled -}}
{{- .Values.redis.master.containerPorts.redis -}}
{{- else -}}
{{- .Values.externalRedis.port -}}
{{- end -}}
{{- end -}}

{{/*
Define diode-ingester-secret
*/}}
{{- define "diode-ingester.secret" -}}
{{- if .Values.diodeIngester.existingSecret -}}
{{- .Values.diodeIngester.existingSecret -}}
{{- else -}}
{{- printf "%s-secret" .Values.diodeIngester.serviceName -}}
{{- end -}}
{{- end -}}

{{/*
Define diode-reconciler-secret
*/}}
{{- define "diode-reconciler.secret" -}}
{{- if .Values.diodeReconciler.existingSecret -}}
{{- .Values.diodeReconciler.existingSecret -}}
{{- else -}}
{{- printf "%s-secret" .Values.diodeReconciler.serviceName -}}
{{- end -}}
{{- end -}}
