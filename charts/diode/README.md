# diode

A Helm chart for Diode

![Version: 1.10.0](https://img.shields.io/badge/Version-1.10.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.5.0](https://img.shields.io/badge/AppVersion-1.5.0-informational?style=flat-square)

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- jq

## Components

The chart includes the following components:

- Diode Auth
- Diode Ingester
- Diode Reconciler
- Hydra (OAuth2 server)
- PostgreSQL (optional)
- Redis (optional)
- Ingress Nginx Controller (optional)
- Cert Manager (optional)

## Installing the Chart

Add diode repository:

```console
helm repo add diode https://netboxlabs.github.io/diode/charts
helm repo update
```

### Quick Installation

Download the quickstart script:

```console
curl -sSfLo quickstart.sh https://raw.githubusercontent.com/netboxlabs/diode/release/charts/diode/scripts/quickstart.sh
chmod +x quickstart.sh
```

Run the following command to create a new namespace and all required resources:

```console
./quickstart.sh [NAMESPACE]
```

Run the following command to install the chart with release name `[RELEASE_NAME]` in namespace `[NAMESPACE]` with default values:

```console
helm install [RELEASE_NAME] diode/diode --namespace [NAMESPACE]
```

To uninstall the `[RELEASE_NAME]` deployment:

```console
helm uninstall [RELEASE_NAME] --namespace [NAMESPACE]
```

### Step-by-step Installation

Create namespace for Diode:

```console
export NAMESPACE=[NAMESPACE]

kubectl create namespace $NAMESPACE
```

Create namespaces for optional components:

```console
# Create namespace for cert-manager if enabled
kubectl create namespace diode-cert-manager
```

Create the following secrets in the `[NAMESPACE]` namespace:

- if not using external resources, generate passwords for Redis and PostgreSQL and set them to the following variables:

```console
REDIS_PASSWORD=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32 | base64 | head -c 32)
POSTGRES_PASSWORD=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32 | base64 | head -c 32)
DIODE_POSTGRES_PASSWORD=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32 | base64 | head -c 32)
HYDRA_POSTGRES_PASSWORD=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32 | base64 | head -c 32)
POSTGRES_HOSTNAME=diode-postgresql.$NAMESPACE.svc.cluster.local
POSTGRES_PORT=5432
```

- if using external resources, set the passwords, host and port of the existing resources:

```console
REDIS_PASSWORD=<redis-password>
POSTGRES_PASSWORD=<postgres-password>
DIODE_POSTGRES_PASSWORD=<diode-postgres-password>
HYDRA_POSTGRES_PASSWORD=<hydra-postgres-password>
POSTGRES_HOSTNAME=<postgresql-hostname>
POSTGRES_PORT=<postgresql-port>
```

- create a secret for PostgreSQL credentials in the `[NAMESPACE]` namespace:

```console
kubectl create secret generic diode-postgresql-secret --namespace $NAMESPACE \
  --from-literal=postgres-database=postgres \
  --from-literal=postgres-username=postgres \
  --from-literal=postgres-password=$POSTGRES_PASSWORD \
  --from-literal=diode-database=diode \
  --from-literal=diode-username=diode \
  --from-literal=diode-password=$DIODE_POSTGRES_PASSWORD \
  --from-literal=hydra-database=hydra \
  --from-literal=hydra-username=hydra \
  --from-literal=hydra-password=$HYDRA_POSTGRES_PASSWORD
```

- create a secret for Redis credentials in the `[NAMESPACE]` namespace:

```console
kubectl create secret generic diode-redis-secret --namespace $NAMESPACE \
  --from-literal=redis-password=$REDIS_PASSWORD
```

- create a secret for Ory Hydra credentials in the `[NAMESPACE]` namespace:

```console
kubectl create secret generic diode-hydra-secret --namespace $NAMESPACE \
  --from-literal=secretsCookie=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32 | base64 | head -c 32) \
  --from-literal=secretsSystem=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32 | base64 | head -c 32) \
  --from-literal=dsn=postgres://hydra:$HYDRA_POSTGRES_PASSWORD@$POSTGRES_HOSTNAME:$POSTGRES_PORT/hydra
```

- generate client credentials for OAuth2 server (Ory Hydra):

```console
curl -o generate-client-credentials.sh https://raw.githubusercontent.com/netboxlabs/diode/release/charts/diode/scripts/generate-client-credentials.sh
chmod +x generate-client-credentials.sh
./generate-client-credentials.sh > <YOUR_PATH>/client-credentials.json
```

- create a secret for the OAuth2 server (Ory Hydra) client credentials with the generated client credentials:

```console
kubectl create secret generic diode-auth-oauth2-secret --namespace $NAMESPACE \
  --from-file=client-credentials.json=<YOUR_PATH>/client-credentials.json
```

- create a secret for the Diode Ingester service:

```console
kubectl create secret generic diode-ingester-secret --namespace $NAMESPACE \
  --from-literal=REDIS_PASSWORD=$REDIS_PASSWORD
```

- create a secret for the Diode Reconciler service:

```console
kubectl create secret generic diode-reconciler-secret --namespace $NAMESPACE \
  --from-literal=REDIS_PASSWORD=$REDIS_PASSWORD \
  --from-literal=POSTGRES_PASSWORD=$DIODE_POSTGRES_PASSWORD \
  --from-literal=DIODE_TO_NETBOX_CLIENT_SECRET=$(jq -r '.[] | select(.client_id == "diode-to-netbox") | .client_secret' <YOUR_PATH>/client-credentials.json)
```

Install chart with release name `[RELEASE_NAME]` in namespace `[NAMESPACE]` with default values:

```console
helm install [RELEASE_NAME] diode/diode --namespace $NAMESPACE --create-namespace
```

Install chart with release name `[RELEASE_NAME]` in namespace `[NAMESPACE]` with your own `values.yaml` (see [Configuration](#configuration)):

```console
helm install [RELEASE_NAME] diode/diode --namespace $NAMESPACE --create-namespace -f values.yaml
```

Install chart with release name `[RELEASE_NAME]` in namespace `[NAMESPACE]` with overridden values using `--set` flag (see [Configuration](#configuration)):

```console
helm install [RELEASE_NAME] diode/diode --namespace $NAMESPACE --create-namespace --set [KEY]=[VALUE]
```

## Uninstalling the Chart

To uninstall the `[RELEASE_NAME]` deployment:

```console
helm uninstall [RELEASE_NAME] --namespace $NAMESPACE
```

## Configuration

Default configuration values are set in `values.yaml` file that can be overridden by providing your own `values.yaml`
file or by using the `--set` flag to override individual values.

See default values in the [Values](#values) or in `values.yaml` file:

```console
helm show values diode/diode
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | common | 2.31.1 |
| https://charts.bitnami.com/bitnami | postgresql | 16.6.3 |
| https://charts.bitnami.com/bitnami | redis | 20.11.5 |
| https://charts.jetstack.io | cert-manager | v1.12.0 |
| https://k8s.ory.sh/helm/charts | hydra | 0.53.0 |
| https://kubernetes.github.io/ingress-nginx | ingress-nginx | 4.12.1 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| certIssuer | object | `{"email":"admin@example.com","enabled":false,"kind":"Issuer","name":"letsencrypt-development"}` | ref: https://cert-manager.io/docs/configuration/acme/ |
| certIssuer.email | string | `"admin@example.com"` | email address for Let's Encrypt notifications |
| certIssuer.enabled | bool | `false` | enable certificate issuer creation |
| certIssuer.kind | string | `"Issuer"` | issuer kind (Issuer or ClusterIssuer) ref: https://cert-manager.io/docs/configuration/acme/ |
| certIssuer.name | string | `"letsencrypt-development"` | issuer name |
| certManager | object | `{"cainjector":{"enabled":true},"crds":{"enabled":true},"enabled":false,"namespace":"diode-cert-manager","prometheus":{"enabled":false},"webhook":{"enabled":true}}` | ref: https://github.com/cert-manager/cert-manager/blob/master/deploy/charts/cert-manager/values.yaml |
| certManager.cainjector | object | `{"enabled":true}` | cainjector enabled |
| certManager.crds | object | `{"enabled":true}` | install CRDs |
| certManager.enabled | bool | `false` | cert-manager enabled |
| certManager.namespace | string | `"diode-cert-manager"` | cert-manager namespace |
| certManager.prometheus | object | `{"enabled":false}` | prometheus enabled |
| certManager.webhook | object | `{"enabled":true}` | webhook enabled |
| diode.environment | string | `"development"` | environment name |
| diodeAuth.annotations | object | `{}` | annotations to add to the auth deployment |
| diodeAuth.config.loggingLevel | string | `"INFO"` | logging level |
| diodeAuth.config.sentryDsn | string | `""` | sentry DSN |
| diodeAuth.config.telemetryEnvironment | string | `"dev"` | telemetry environment |
| diodeAuth.config.telemetryMetricsExporter | string | `"prometheus"` | telemetry metrics exporter |
| diodeAuth.config.telemetryTracesExporter | string | `"none"` | telemetry traces exporter |
| diodeAuth.containerPort | int | `8080` | port to listen on |
| diodeAuth.enabled | bool | `true` | enabled |
| diodeAuth.extraEnvs | string or list | `[]` | extra environment variables to be set on containers' `env` section |
| diodeAuth.extraInitContainers | string or list | `""` | additional containers to run before auth finishes initializing (may contain templating instructions) |
| diodeAuth.image.imagePullSecrets | list | `[]` | secrets with credentials to pull images from a private registry |
| diodeAuth.image.pullPolicy | string | `"IfNotPresent"` | pull policy |
| diodeAuth.image.repository | string | `"docker.io/netboxlabs/diode-auth"` | image repository |
| diodeAuth.image.tag | string | `"1.10.0"` | image tag |
| diodeAuth.replicaCount | int | `1` | replica count |
| diodeAuth.resources | object | `{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | resources |
| diodeAuth.serviceAccount.create | bool | `true` | create service account |
| diodeAuthBootstrap.enabled | bool | `true` | enabled |
| diodeAuthBootstrap.image.imagePullSecrets | list | `[]` | secrets with credentials to pull images from a private registry |
| diodeAuthBootstrap.image.pullPolicy | string | `"IfNotPresent"` | pull policy |
| diodeAuthBootstrap.image.repository | string | `"docker.io/netboxlabs/diode-auth"` | image repository |
| diodeAuthBootstrap.image.tag | string | `"1.10.0"` | image tag |
| diodeAuthBootstrap.job.annotations | object | `{"helm.sh/hook":"post-install, post-upgrade","helm.sh/hook-weight":"2"}` | annotations to add to the auth bootstrap job |
| diodeAuthBootstrap.job.backoffLimit | int | `20` | backoff limit |
| diodeAuthBootstrap.job.extraInitContainers | string or list | `""` | additional initContainers to run during bootstrap (may contain templating instructions) |
| diodeIngester.annotations | object | `{}` | annotations to add to the ingester deployment |
| diodeIngester.config.loggingLevel | string | `"INFO"` | logging level |
| diodeIngester.config.redisStreamDb | int | `1` | redis stream db |
| diodeIngester.config.sentryDsn | string | `""` | sentry DSN |
| diodeIngester.config.telemetryEnvironment | string | `"dev"` | telemetry environment |
| diodeIngester.config.telemetryMetricsExporter | string | `"prometheus"` | telemetry metrics exporter |
| diodeIngester.config.telemetryTracesExporter | string | `"none"` | telemetry traces exporter |
| diodeIngester.containerPort | int | `8081` | port to listen on |
| diodeIngester.enabled | bool | `true` | enabled |
| diodeIngester.existingSecret | string | `"diode-ingester-secret"` | existing secret name |
| diodeIngester.extraEnvs | string or list | `[]` | extra environment variables to be set on containers' `env` section |
| diodeIngester.extraInitContainers | string or list | `""` | additional containers to run before the ingester finishes initializing (may contain templating instructions) |
| diodeIngester.grpc.serviceName | string | `"diode.v1.IngesterService"` | grpc service name |
| diodeIngester.image.imagePullSecrets | list | `[]` | secrets with credentials to pull images from a private registry |
| diodeIngester.image.pullPolicy | string | `"IfNotPresent"` | pull policy |
| diodeIngester.image.repository | string | `"docker.io/netboxlabs/diode-ingester"` | image repository |
| diodeIngester.image.tag | string | `"1.11.0"` | image tag |
| diodeIngester.replicaCount | int | `1` | replica count |
| diodeIngester.resources | object | `{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | resources |
| diodeIngester.serviceAccount.create | bool | `true` | create service account |
| diodeReconciler.annotations | object | `{}` | annotations to add to the reconciler deployment |
| diodeReconciler.config.autoApplyChangesets | string | `"true"` | auto apply changesets |
| diodeReconciler.config.diodeToNetBoxClientId | string | `"diode-to-netbox"` | diode to netbox client id |
| diodeReconciler.config.diodeToNetboxRateLimiterBurst | int | `1` | diode to netbox rate limiter burst |
| diodeReconciler.config.diodeToNetboxRateLimiterRps | int | `20` | diode to netbox rate limiter rps |
| diodeReconciler.config.loggingLevel | string | `"INFO"` | logging level |
| diodeReconciler.config.migrationEnabled | string | `"true"` | migration enabled |
| diodeReconciler.config.netboxDiodePluginApiBaseUrl | string | `"http://localhost:8000/netbox/api/plugins/diode"` | netbox diode plugin api base url |
| diodeReconciler.config.netboxDiodePluginSkipTlsVerify | bool | `false` | netbox diode plugin skip tls verify |
| diodeReconciler.config.postgresDbName | string | `"diode"` | postgres db name |
| diodeReconciler.config.postgresUser | string | `"diode"` | postgres user |
| diodeReconciler.config.reconcilerRateLimiterBurst | int | `1` | reconciler rate limiter burst |
| diodeReconciler.config.reconcilerRateLimiterRps | int | `20` | reconciler rate limiter rps |
| diodeReconciler.config.redisDb | int | `0` | redis db |
| diodeReconciler.config.redisStreamDb | int | `1` | redis stream db |
| diodeReconciler.config.sentryDsn | string | `""` | sentry DSN |
| diodeReconciler.config.telemetryEnvironment | string | `"dev"` | telemetry environment |
| diodeReconciler.config.telemetryMetricsExporter | string | `"prometheus"` | telemetry metrics exporter |
| diodeReconciler.config.telemetryTracesExporter | string | `"none"` | telemetry traces exporter |
| diodeReconciler.containerPort | int | `8081` | port to listen on |
| diodeReconciler.enabled | bool | `true` | enabled |
| diodeReconciler.existingSecret | string | `"diode-reconciler-secret"` | existing secret name |
| diodeReconciler.extraEnvs | string or list | `[]` | extra environment variables to be set on containers' `env` section |
| diodeReconciler.extraInitContainers | string or list | `""` | additional containers to run before the reconciler finishes initializing (may contain templating instructions) |
| diodeReconciler.grpc.serviceName | string | `"diode.v1.ReconcilerService"` | grpc service name |
| diodeReconciler.image.imagePullSecrets | list | `[]` | secrets with credentials to pull images from a private registry |
| diodeReconciler.image.pullPolicy | string | `"IfNotPresent"` | pull policy |
| diodeReconciler.image.repository | string | `"docker.io/netboxlabs/diode-reconciler"` | image repository |
| diodeReconciler.image.tag | string | `"1.11.0"` | image tag |
| diodeReconciler.replicaCount | int | `1` | replica count |
| diodeReconciler.resources | object | `{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | resources |
| diodeReconciler.serviceAccount.create | bool | `true` | create service account |
| externalPostgresql.database | string | `"diode"` | database name |
| externalPostgresql.existingSecretKey | string | `""` | key of password in existing postgresql secret |
| externalPostgresql.existingSecretName | string | `""` | existing postgresql secret |
| externalPostgresql.hostname | string | `"localhost"` | hostname |
| externalPostgresql.password | string | `""` | password |
| externalPostgresql.port | int | `5432` | port |
| externalPostgresql.sslMode | string | `""` | ssl mode |
| externalPostgresql.username | string | `"diode"` | username |
| externalRedis.hostname | string | `"localhost"` | hostname |
| externalRedis.port | int | `6379` | port |
| externalRedis.tls.caPath | string | `""` | path to CA certificate to verify server |
| externalRedis.tls.clientCertPath | string | `""` | path to client certificate for mutual TLS |
| externalRedis.tls.clientKeyPath | string | `""` | path to client private key for mutual TLS |
| externalRedis.tls.enabled | bool | `false` | enable TLS |
| externalRedis.tls.skipVerify | bool | `false` | skip TLS verify |
| externalRedis.username | string | `""` | username (optional, Redis 6+) |
| global.commonAnnotations | object | `{}` | common annotations for all resources |
| global.commonLabels | object | `{}` | common labels for all resources |
| global.diode | object | `{"busybox":{"image":"busybox:latest","imagePullPolicy":"IfNotPresent"},"hydra":{"waitForPostgres":true},"ingester":{"waitForRedis":true},"reconciler":{"waitForPostgres":true,"waitForRedis":true}}` | diode global configuration |
| global.diode.busybox | object | `{"image":"busybox:latest","imagePullPolicy":"IfNotPresent"}` | busybox image configuration |
| global.diode.hydra | object | `{"waitForPostgres":true}` | hydra additional init containers configuration |
| global.diode.hydra.waitForPostgres | bool | `true` | wait for PostgreSQL to be reachable |
| global.diode.ingester.waitForRedis | bool | `true` | wait for Redis to be reachable |
| global.diode.reconciler.waitForPostgres | bool | `true` | wait for PostgreSQL to be reachable |
| global.diode.reconciler.waitForRedis | bool | `true` | wait for Redis to be reachable |
| global.security.allowInsecureImages | bool | `true` |  |
| hydra | object | `{"deployment":{"extraInitContainers":"{{ include \"diode.hydra.extrainitcontainers\" . }}","resources":{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}},"enabled":true,"fullnameOverride":"diode-hydra","hydra":{"automigration":{"enabled":true},"config":{"oidc":{"subject_identifiers":{"supported_types":["public"]}},"strategies":{"access_token":"jwt","jwt":{"scope_claim":"both"}},"ttl":{"access_token":"1h"},"urls":{"self":{"issuer":"http://diode-hydra-public.{{ .Release.Namespace }}.svc.cluster.local:4444"}}},"dev":true,"ingress":{"admin":{"enabled":false},"public":{"enabled":false}},"service":{"admin":{"enabled":true,"port":4445,"type":"ClusterIP"},"public":{"enabled":true,"port":4444,"type":"ClusterIP"}}},"job":{"annotations":{"helm.sh/hook":"post-install, post-upgrade","helm.sh/hook-delete-policy":"hook-succeeded","helm.sh/hook-weight":"1"},"extraInitContainers":"{{ include \"diode.hydra.extrainitcontainers\" . }}"},"secret":{"enabled":false,"nameOverride":"diode-hydra-secret"}}` | ref: https://github.com/ory/k8s/blob/master/helm/charts/hydra/values.yaml |
| hydra.deployment.extraInitContainers | string or list | `"{{ include \"diode.hydra.extrainitcontainers\" . }}"` | extra init containers |
| hydra.deployment.resources | object | `{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | resources |
| hydra.enabled | bool | `true` | enabled |
| hydra.fullnameOverride | string | `"diode-hydra"` | fullname override |
| hydra.hydra.automigration.enabled | bool | `true` | automigration enabled |
| hydra.hydra.config.oidc.subject_identifiers | object | `{"supported_types":["public"]}` | subject identifiers |
| hydra.hydra.config.strategies.access_token | string | `"jwt"` | access token strategy |
| hydra.hydra.config.strategies.jwt.scope_claim | string | `"both"` | scope claim |
| hydra.hydra.config.ttl.access_token | string | `"1h"` | access token TTL |
| hydra.hydra.config.urls.self | object | `{"issuer":"http://diode-hydra-public.{{ .Release.Namespace }}.svc.cluster.local:4444"}` | self issuer |
| hydra.hydra.config.urls.self.issuer | string | `"http://diode-hydra-public.{{ .Release.Namespace }}.svc.cluster.local:4444"` | hydra public url |
| hydra.hydra.dev | bool | `true` | dev mode |
| hydra.hydra.ingress.admin.enabled | bool | `false` | admin ingress enabled |
| hydra.hydra.ingress.public.enabled | bool | `false` | public ingress enabled |
| hydra.hydra.service.admin.enabled | bool | `true` | admin service enabled |
| hydra.hydra.service.admin.port | int | `4445` | admin service port |
| hydra.hydra.service.admin.type | string | `"ClusterIP"` | admin service type |
| hydra.hydra.service.public.enabled | bool | `true` | public service enabled |
| hydra.hydra.service.public.port | int | `4444` | public service port |
| hydra.hydra.service.public.type | string | `"ClusterIP"` | public service type |
| hydra.job.annotations | object | `{"helm.sh/hook":"post-install, post-upgrade","helm.sh/hook-delete-policy":"hook-succeeded","helm.sh/hook-weight":"1"}` | job annotations |
| hydra.job.extraInitContainers | string or list | `"{{ include \"diode.hydra.extrainitcontainers\" . }}"` | extra init containers |
| hydra.secret.enabled | bool | `false` | secret enabled |
| hydra.secret.nameOverride | string | `"diode-hydra-secret"` | existing secret name |
| ingressNginx | object | `{"annotations":{},"controller":{"allowSnippetAnnotations":true,"enabled":true},"enabled":true,"extraHttpPaths":[],"grpcAnnotations":{"nginx.ingress.kubernetes.io/proxy-body-size":"25m","nginx.ingress.kubernetes.io/ssl-redirect":"true"},"hostname":"","httpAnnotations":{"nginx.ingress.kubernetes.io/ssl-redirect":"true"},"ingressClass":"nginx","pathPrefix":"/diode","tls":{}}` | ref: https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/values.yaml |
| ingressNginx.controller | object | `{"allowSnippetAnnotations":true,"enabled":true}` | ingress annotations |
| ingressNginx.controller.allowSnippetAnnotations | bool | `true` | allow snippet annotations |
| ingressNginx.controller.enabled | bool | `true` | deploy an ingress-nginx controller chart in addition to `Ingress` resources |
| ingressNginx.enabled | bool | `true` | ingress-nginx enabled |
| ingressNginx.extraHttpPaths | list | `[]` | ingress extra http paths |
| ingressNginx.grpcAnnotations | object | `{"nginx.ingress.kubernetes.io/proxy-body-size":"25m","nginx.ingress.kubernetes.io/ssl-redirect":"true"}` | ingress grpc annotations |
| ingressNginx.hostname | string | `""` | hostname |
| ingressNginx.httpAnnotations | object | `{"nginx.ingress.kubernetes.io/ssl-redirect":"true"}` | ingress http annotations |
| ingressNginx.ingressClass | string | `"nginx"` | ingress class |
| ingressNginx.pathPrefix | string | `"/diode"` | ingress path prefix |
| ingressNginx.tls | object | `{}` | ingress tls |
| postgresql | object | `{"auth":{"existingSecret":"diode-postgresql-secret","secretKeys":{"adminPasswordKey":"postgres-password"}},"enabled":true,"fullnameOverride":"diode-postgresql","image":{"repository":"bitnamilegacy/postgresql"},"metrics":{"image":{"repository":"bitnamilegacy/postgres-exporter"}},"primary":{"extraVolumeMounts":[{"mountPath":"/docker-entrypoint-initdb.d/init_diode_databases.sh","name":"custom-init-scripts","subPath":"init_diode_databases.sh"}],"initdb":{"scriptsConfigMap":"diode-postgresql-initdb-scripts-configmap"},"livenessProbe":{"enabled":true,"failureThreshold":6,"initialDelaySeconds":30,"periodSeconds":10,"successThreshold":1,"timeoutSeconds":5},"persistence":{"enabled":true,"size":"10Gi"},"readinessProbe":{"enabled":true,"failureThreshold":6,"initialDelaySeconds":5,"periodSeconds":10,"successThreshold":1,"timeoutSeconds":5}},"volumePermissions":{"image":{"repository":"bitnamilegacy/postgresql"}}}` | ref: https://github.com/bitnami/charts/tree/main/bitnami/postgresql |
| postgresql.auth.existingSecret | string | `"diode-postgresql-secret"` | existing secret name |
| postgresql.auth.secretKeys | object | `{"adminPasswordKey":"postgres-password"}` | existing secret password key |
| postgresql.enabled | bool | `true` | enabled |
| postgresql.fullnameOverride | string | `"diode-postgresql"` | fullname override |
| postgresql.image.repository | string | `"bitnamilegacy/postgresql"` | image repository |
| postgresql.metrics.image.repository | string | `"bitnamilegacy/postgres-exporter"` | image repository |
| postgresql.primary.extraVolumeMounts | list | `[{"mountPath":"/docker-entrypoint-initdb.d/init_diode_databases.sh","name":"custom-init-scripts","subPath":"init_diode_databases.sh"}]` | extra volume mounts |
| postgresql.primary.initdb.scriptsConfigMap | string | `"diode-postgresql-initdb-scripts-configmap"` | scripts config map |
| postgresql.primary.livenessProbe | object | `{"enabled":true,"failureThreshold":6,"initialDelaySeconds":30,"periodSeconds":10,"successThreshold":1,"timeoutSeconds":5}` | liveness probe |
| postgresql.primary.persistence.enabled | bool | `true` | persistence enabled |
| postgresql.primary.persistence.size | string | `"10Gi"` | persistence size |
| postgresql.primary.readinessProbe | object | `{"enabled":true,"failureThreshold":6,"initialDelaySeconds":5,"periodSeconds":10,"successThreshold":1,"timeoutSeconds":5}` | readiness probe |
| postgresql.volumePermissions.image.repository | string | `"bitnamilegacy/postgresql"` | image repository |
| redis | object | `{"auth":{"enabled":true,"existingSecret":"diode-redis-secret","existingSecretPasswordKey":"redis-password"},"containerPorts":{"redis":6379},"enabled":true,"fullnameOverride":"diode-redis","image":{"repository":"bitnamilegacy/redis"},"kubectl":{"image":{"repository":"bitnamilegacy/kubectl"}},"metrics":{"image":{"repository":"bitnamilegacy/redis-exporter"}},"persistence":{"enabled":true,"size":"1Gi"},"replica":{"replicaCount":1},"sentinel":{"image":{"repository":"bitnamilegacy/redis-sentinel"}},"service":{"port":6379},"sysctl":{"image":{"repository":"bitnamilegacy/os-shell"}},"volumePermissions":{"image":{"repository":"bitnamilegacy/os-shell"}}}` | ref: https://github.com/bitnami/charts/tree/main/bitnami/redis |
| redis.auth.enabled | bool | `true` | auth enabled |
| redis.auth.existingSecret | string | `"diode-redis-secret"` | existing secret name |
| redis.auth.existingSecretPasswordKey | string | `"redis-password"` | existing secret password key |
| redis.containerPorts | object | `{"redis":6379}` | container ports |
| redis.enabled | bool | `true` | enabled |
| redis.fullnameOverride | string | `"diode-redis"` | fullname override |
| redis.image.repository | string | `"bitnamilegacy/redis"` | image repository |
| redis.kubectl.image.repository | string | `"bitnamilegacy/kubectl"` | image repository |
| redis.metrics.image.repository | string | `"bitnamilegacy/redis-exporter"` | image repository |
| redis.persistence.enabled | bool | `true` | persistence enabled |
| redis.persistence.size | string | `"1Gi"` | persistence size |
| redis.replica.replicaCount | int | `1` | replica count |
| redis.sentinel.image.repository | string | `"bitnamilegacy/redis-sentinel"` | image repository |
| redis.service.port | int | `6379` | service port |
| redis.sysctl.image.repository | string | `"bitnamilegacy/os-shell"` | image repository |
| redis.volumePermissions.image.repository | string | `"bitnamilegacy/os-shell"` | image repository |

## License

Copyright &copy; 2025 NetBox Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
