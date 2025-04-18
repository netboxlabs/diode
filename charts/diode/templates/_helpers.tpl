{{/*
Expand the name of the chart.
*/}}
{{- define "diode.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "diode.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create the namespace of the release.
Allows overriding it for multi-namespace deployments in combined charts.
*/}}
{{- define "diode.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "diode.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "diode.labels" -}}
helm.sh/chart: {{ include "diode.chart" . }}
{{ include "diode.selectorlabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "diode.selectorlabels" -}}
app.kubernetes.io/name: {{ include "diode.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the busybox image to use
*/}}
{{- define "diode.busybox.image" -}}
{{- if and .Values.diode (hasKey .Values.diode "busybox") (hasKey .Values.diode.busybox "image") -}}
{{- .Values.diode.busybox.image | quote -}}
{{- else -}}
{{- "busybox:latest" | quote -}}
{{- end -}}
{{- end }}

{{/*
Create the name of the auth service to use
*/}}
{{- define "diode.auth.servicename" -}}
{{- printf "%s-auth" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the auth bootstrap job to use
*/}}
{{- define "diode.auth.bootstrapjobname" -}}
{{- printf "%s-auth-bootstrap" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the auth bootstrap configmap to use
*/}}
{{- define "diode.auth.bootstrapconfigmapname" -}}
{{- printf "%s-auth-bootstrap-configmap" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the auth migrations configmap to use
*/}}
{{- define "diode.auth.migrations.configmapname" -}}
{{- printf "%s-auth-migrations-configmap" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the auth service account to use
*/}}
{{- define "diode.auth.serviceaccountname" -}}
{{- printf "%s-auth-serviceaccount" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the hostname of the auth service
*/}}
{{- define "diode.auth.hostname" -}}
{{- printf "%s-auth.%s.svc.cluster.local" (include "diode.fullname" .) .Release.Namespace }}
{{- end }}

{{/*
Create the name of the auth configmap to use
*/}}
{{- define "diode.auth.configmap" -}}
{{- printf "%s-auth-configmap" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the auth OAuth2 client credentials secret to use
*/}}
{{- define "diode.auth.oauth2.secret" -}}
{{- printf "%s-auth-oauth2-secret" (include "diode.name" .) }}
{{- end }}

{{/*
Create the name of the ingester service to use
*/}}
{{- define "diode.ingester.servicename" -}}
{{- printf "%s-ingester" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the ingester service account to use
*/}}
{{- define "diode.ingester.serviceaccountname" -}}
{{- printf "%s-ingester-serviceaccount" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the ingester configmap to use
*/}}
{{- define "diode.ingester.configmap" -}}
{{- printf "%s-ingester-configmap" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the ingester secret to use
*/}}
{{- define "diode.ingester.secret" -}}
{{- printf "%s-ingester-secret" (include "diode.name" .) }}
{{- end }}

{{/*
Create the name of the reconciler service to use
*/}}
{{- define "diode.reconciler.servicename" -}}
{{- printf "%s-reconciler" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the reconciler service account to use
*/}}
{{- define "diode.reconciler.serviceaccountname" -}}
{{- printf "%s-reconciler-serviceaccount" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the reconciler configmap to use
*/}}
{{- define "diode.reconciler.configmap" -}}
{{- printf "%s-reconciler-configmap" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the reconciler secret to use
*/}}
{{- define "diode.reconciler.secret" -}}
{{- printf "%s-reconciler-secret" (include "diode.name" .) }}
{{- end }}

{{/*
Create the name of the PostgreSQL initialization database scripts ConfigMap
*/}}
{{- define "diode.postgresql.initdb.scriptsconfigmap" -}}
{{- printf "%s-postgresql-initdb-scripts-configmap" (include "diode.name" .) }}
{{- end }}

{{/*
Create the hostname of the PostgreSQL database
*/}}
{{- define "diode.postgresql.hostname" -}}
{{- if and .Values.postgresql (hasKey .Values.postgresql "enabled") (eq .Values.postgresql.enabled true) -}}
{{- printf "diode-postgresql.%s.svc.cluster.local" .Release.Namespace }}
{{- else if and .Values.externalPostgresql (hasKey .Values.externalPostgresql "hostname") -}}
{{- .Values.externalPostgresql.hostname }}
{{- else -}}
{{- fail "externalPostgresql.hostname must be defined when postgresql.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the port of the PostgreSQL database
*/}}
{{- define "diode.postgresql.port" -}}
{{- if and .Values.postgresql (hasKey .Values.postgresql "enabled") (eq .Values.postgresql.enabled true) -}}
{{- printf "5432" }}
{{- else if and .Values.externalPostgresql (hasKey .Values.externalPostgresql "port") -}}
{{- .Values.externalPostgresql.port }}
{{- else -}}
{{- fail "externalPostgresql.port must be defined when postgresql.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the hostname of the Redis database
*/}}
{{- define "diode.redis.hostname" -}}
{{- if and .Values.redis (hasKey .Values.redis "enabled") (eq .Values.redis.enabled true) -}}
{{- printf "diode-redis-master.%s.svc.cluster.local" .Release.Namespace }}
{{- else if and .Values.externalRedis (hasKey .Values.externalRedis "hostname") -}}
{{- .Values.externalRedis.hostname }}
{{- else -}}
{{- fail "externalRedis.hostname must be defined when redis.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the port of the Redis database
*/}}
{{- define "diode.redis.port" -}}
{{- if and .Values.redis (hasKey .Values.redis "enabled") (eq .Values.redis.enabled true) -}}
{{- printf "6379" }}
{{- else if and .Values.externalRedis (hasKey .Values.externalRedis "port") -}}
{{- .Values.externalRedis.port }}
{{- else -}}
{{- fail "externalRedis.port must be defined when redis.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the hostname of the public Hydra service
*/}}
{{- define "diode.hydra.public.hostname" -}}
{{- printf "diode-hydra-public.%s.svc.cluster.local" .Release.Namespace }}
{{- end }}

{{/*
Create the port of the public Hydra service
*/}}
{{- define "diode.hydra.public.port" -}}
{{- .Values.hydra.hydra.service.public.port | default 4444 | int }}
{{- end }}

{{/*
Create the URL to the public Hydra service
*/}}
{{- define "diode.hydra.public.url" -}}
{{- printf "http://%s:%d" (include "diode.hydra.public.hostname" .) (include "diode.hydra.public.port" . | int) }}
{{- end }}

{{/*
Create the hostname of the admin Hydra service
*/}}
{{- define "diode.hydra.admin.hostname" -}}
{{- printf "diode-hydra-admin.%s.svc.cluster.local" .Release.Namespace }}
{{- end }}

{{/*
Create the port of the admin Hydra service
*/}}
{{- define "diode.hydra.admin.port" -}}
{{- .Values.hydra.hydra.service.admin.port | default 4445 | int }}
{{- end }}

{{/*
Create the URL to the admin Hydra service
*/}}
{{- define "diode.hydra.admin.url" -}}
{{- printf "http://%s:%d" (include "diode.hydra.admin.hostname" .) (include "diode.hydra.admin.port" . | int) }}
{{- end }}

{{/*
Create the name of the ingress to use
*/}}
{{- define "diode.ingress.name" -}}
{{- printf "%s-ingress" (include "diode.fullname" .) }}
{{- end }}

{{- define "diode.ingress.path" -}}
{{- $prefix := .Values.ingressNginx.pathPrefix | default "" -}}
{{- if and $prefix (ne $prefix "") -}}
{{ trimSuffix "/" $prefix }}
{{- end -}}
{{- end -}}
