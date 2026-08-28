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
{{- include "common.labels.standard" ( dict "customLabels" .Values.global.commonLabels "context" . ) }}
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
{{- if .Values.global.diode.busybox.image -}}
{{- .Values.global.diode.busybox.image | quote -}}
{{- else -}}
{{- "busybox:latest" | quote -}}
{{- end -}}
{{- end }}

{{/*
Create the image pull policy of the busybox image to use
*/}}
{{- define "diode.busybox.imagepullpolicy" -}}
{{- if .Values.global.diode.busybox.imagePullPolicy -}}
{{- .Values.global.diode.busybox.imagePullPolicy | quote -}}
{{- else -}}
{{- "IfNotPresent" | quote -}}
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
Create the URL to the auth service
*/}}
{{- define "diode.auth.url" -}}
{{- printf "http://%s:%d" (include "diode.auth.hostname" .) (int .Values.diodeAuth.containerPort) }}
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
Create the name of the ingester grpc service to use
*/}}
{{- define "diode.ingester.grpc.servicename" -}}
{{- if and (hasKey .Values.diodeIngester "grpc") (hasKey .Values.diodeIngester.grpc "serviceName")  (not (empty .Values.diodeIngester.grpc.serviceName)) }}
{{- .Values.diodeIngester.grpc.serviceName }}
{{- else }}
{{- printf "diode.v1.IngesterService" }}
{{- end }}
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
{{- if .Values.diodeIngester.existingSecret }}
{{- .Values.diodeIngester.existingSecret }}
{{- else }}
{{- printf "%s-ingester-secret" (include "diode.name" .) }}
{{- end }}
{{- end }}

{{/*
Create the name of the reconciler service to use
*/}}
{{- define "diode.reconciler.servicename" -}}
{{- printf "%s-reconciler" (include "diode.fullname" .) }}
{{- end }}

{{/*
Create the name of the reconciler grpc service to use
*/}}
{{- define "diode.reconciler.grpc.servicename" -}}
{{- if and (hasKey .Values.diodeReconciler "grpc") (hasKey .Values.diodeReconciler.grpc "serviceName") (not (empty .Values.diodeReconciler.grpc.serviceName)) }}
{{- .Values.diodeReconciler.grpc.serviceName }}
{{- else }}
{{- printf "diode.v1.ReconcilerService" }}
{{- end }}
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
{{- if .Values.diodeReconciler.existingSecret }}
{{- .Values.diodeReconciler.existingSecret }}
{{- else }}
{{- printf "%s-reconciler-secret" (include "diode.name" .) }}
{{- end }}
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
{{- if .Values.postgresql.enabled -}}
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
{{- if .Values.postgresql.enabled -}}
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
{{- if .Values.redis.enabled -}}
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
{{- if .Values.redis.enabled -}}
{{- printf "6379" }}
{{- else if and .Values.externalRedis (hasKey .Values.externalRedis "port") -}}
{{- .Values.externalRedis.port }}
{{- else -}}
{{- fail "externalRedis.port must be defined when redis.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the username of the Redis database
*/}}
{{- define "diode.redis.username" -}}
{{- if .Values.redis.enabled -}}
{{- printf "" }}
{{- else if and .Values.externalRedis (hasKey .Values.externalRedis "username") -}}
{{- .Values.externalRedis.username }}
{{- else -}}
{{- fail "externalRedis.username must be defined when redis.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the database name for PostgreSQL
*/}}
{{- define "diode.postgresql.database" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "diode" }}
{{- else if and .Values.externalPostgresql (hasKey .Values.externalPostgresql "database") -}}
{{- .Values.externalPostgresql.database }}
{{- else -}}
{{- fail "externalPostgresql.database must be defined when postgresql.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the username for PostgreSQL
*/}}
{{- define "diode.postgresql.username" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "diode" }}
{{- else if and .Values.externalPostgresql (hasKey .Values.externalPostgresql "username") -}}
{{- .Values.externalPostgresql.username }}
{{- else -}}
{{- fail "externalPostgresql.username must be defined when postgresql.enabled is false" }}
{{- end }}
{{- end }}

{{/*
Create the secret name for PostgreSQL credentials
*/}}
{{- define "diode.postgresql.secretname" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "diode-postgresql-secret" }}
{{- else if and .Values.externalPostgresql (hasKey .Values.externalPostgresql "existingSecretName") (not (empty .Values.externalPostgresql.existingSecretName)) -}}
{{- .Values.externalPostgresql.existingSecretName }}
{{- else -}}
{{- printf "diode-external-postgresql-secret" }}
{{- end }}
{{- end }}

{{/*
Create the secret key for PostgreSQL password
*/}}
{{- define "diode.postgresql.secretkey" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "postgres-password" }}
{{- else if and .Values.externalPostgresql (hasKey .Values.externalPostgresql "existingSecretKey") (not (empty .Values.externalPostgresql.existingSecretKey)) -}}
{{- .Values.externalPostgresql.existingSecretKey }}
{{- else -}}
{{- printf "postgresql-password" }}
{{- end }}
{{- end }}

{{/*
Create the SSL option for PostgreSQL
*/}}
{{- define "diode.postgresql.sslMode" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "disable" }}
{{- else if and .Values.externalPostgresql (hasKey .Values.externalPostgresql "sslMode") (not (empty .Values.externalPostgresql.sslMode)) -}}
{{- .Values.externalPostgresql.sslMode }}
{{- else -}}
{{- printf "disable" }}
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
Create the name of the extra Hydra init container configmap to use
*/}}
{{- define "diode.hydra.extrainitcontainerconfigmap" -}}
{{- printf "%s-hydra-extra-initcontainer-configmap" .Release.Name }}
{{- end }}

{{/*
Create the name of the extra Hydra init container configmap to use
*/}}
{{- define "diode.hydra.extrainitcontainers" -}}
{{- if .Values.global.diode.hydra.waitForPostgres -}}
- name: wait-for-postgres
  image: {{ .Values.global.diode.busybox.image }}
  imagePullPolicy: {{ .Values.global.diode.busybox.imagePullPolicy }}
  command: ['sh', '-c', 'until nc -zv $POSTGRES_HOST $POSTGRES_PORT; do echo waiting for PostgreSQL; sleep 2; done;']
  envFrom:
    - configMapRef:
        name: {{ include "diode.hydra.extrainitcontainerconfigmap" . }}
{{- end -}}
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
