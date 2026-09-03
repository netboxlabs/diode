#!/usr/bin/env bash

set -euo pipefail

# Detect if stdout is a terminal
if [[ -t 1 ]]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color
else
    GREEN=''
    RED=''
    YELLOW=''
    NC=''
fi

timestamp() {
    date +"[%Y-%m-%d %H:%M:%S]"
}

info() {
    echo -e "$(timestamp) ${YELLOW}[INFO]${NC} $*"
}

ok() {
    echo -e "$(timestamp) ${GREEN}[OK]${NC} $*"
}

error() {
    echo -e "$(timestamp) ${RED}[ERROR]${NC} $*" >&2
}

usage() {
    echo "Usage: $0 NAMESPACE"
    echo
    echo "Environment:"
    echo "  CLUSTER_DOMAIN  DNS domain of the k8s cluster (default: cluster.local)"
    echo
    echo "Examples:"
    echo "  $0 diode-dev"
    echo "  CLUSTER_DOMAIN=cluster.example.net $0 diode-dev"
    exit 1
}

if [[ $# -ne 1 ]]; then
    usage
fi

if ((BASH_VERSINFO[0] < 4)); then
  error "This script requires Bash 4.0+ (associative array support)."
  error "On macOS, install newer bash via brew: brew install bash"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  error "jq is required but not installed. Please install jq."
  exit 1
fi

NAMESPACE=$1

# DNS domain of the k8s cluster; override for clusters not using the default.
CLUSTER_DOMAIN="${CLUSTER_DOMAIN:-cluster.local}"
# check if namespace is provided
if [ -z "$NAMESPACE" ]; then
    error "Error: namespace is required"
    usage
fi

# create namespace if it doesn't exist
if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
    info "Creating namespace $NAMESPACE"
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
else
    ok "Namespace $NAMESPACE already exists"
fi

generate_secret() {
  head -c 32 /dev/urandom | base64 | tr -d '/\n'
}

generate_client_credentials() {
    declare -A CLIENT_CREDENTIALS
    CLIENT_CREDENTIALS["diode-ingest"]="diode:ingest"
    CLIENT_CREDENTIALS["diode-to-netbox"]="netbox:read netbox:write"
    CLIENT_CREDENTIALS["netbox-to-diode"]="diode:read diode:write"

    output="["
    first=true

    for client_id in "${!CLIENT_CREDENTIALS[@]}"; do
        if [ "$first" = true ]; then
            first=false
        else
            output+=","
        fi
        output+="\n  {
        \"client_id\": \"$client_id\",
        \"client_secret\": \"$(generate_secret)\",
        \"grant_types\": [\"client_credentials\"],
        \"scope\": \"${CLIENT_CREDENTIALS[$client_id]}\"
    }"
    done

    output+="\n]\n"
    echo -e "$output"
}

if [ ! -f "client-credentials.json" ]; then
    info "Generating OAuth2 client credentials in ${PWD}/client-credentials.json"
    generate_client_credentials > $PWD/client-credentials.json
else
    ok "Using existing OAuth2 client credentials in ${PWD}/client-credentials.json"
fi

info "Generating secrets"

REDIS_PASSWORD=$(generate_secret)
POSTGRES_PASSWORD=$(generate_secret)
DIODE_POSTGRES_PASSWORD=$(generate_secret)
HYDRA_POSTGRES_PASSWORD=$(generate_secret)
POSTGRES_HOSTNAME=diode-postgresql.$NAMESPACE.svc.$CLUSTER_DOMAIN
POSTGRES_PORT=5432

DIODE_POSTGRES_SECRET="diode-postgresql-secret"
DIODE_REDIS_SECRET="diode-redis-secret"
DIODE_HYDRA_SECRET="diode-hydra-secret"
DIODE_AUTH_OAUTH2_SECRET="diode-auth-oauth2-secret"
DIODE_INGESTER_SECRET="diode-ingester-secret"
DIODE_RECONCILER_SECRET="diode-reconciler-secret"

if kubectl get secret $DIODE_POSTGRES_SECRET -n $NAMESPACE &>/dev/null; then
    ok "Secret $DIODE_POSTGRES_SECRET already exists, skipping creation"
else
    kubectl create secret generic $DIODE_POSTGRES_SECRET --namespace $NAMESPACE \
      --from-literal=postgres-database=postgres \
      --from-literal=postgres-username=postgres \
      --from-literal=postgres-password=$POSTGRES_PASSWORD \
      --from-literal=diode-database=diode \
      --from-literal=diode-username=diode \
      --from-literal=diode-password=$DIODE_POSTGRES_PASSWORD \
      --from-literal=hydra-database=hydra \
      --from-literal=hydra-username=hydra \
      --from-literal=hydra-password=$HYDRA_POSTGRES_PASSWORD \
      --dry-run=client -o yaml | kubectl apply -f - > /dev/null

    if [[ $? -eq 0 ]]; then
        ok "Successfully created secret $DIODE_POSTGRES_SECRET in namespace $NAMESPACE"
    else
        error "Failed to create secret $DIODE_POSTGRES_SECRET in namespace $NAMESPACE"
    fi
fi

if kubectl get secret $DIODE_REDIS_SECRET -n $NAMESPACE &>/dev/null; then
    ok "Secret $DIODE_REDIS_SECRET already exists, skipping creation"
else
    kubectl create secret generic $DIODE_REDIS_SECRET --namespace $NAMESPACE \
      --from-literal=redis-password=$REDIS_PASSWORD \
      --dry-run=client -o yaml | kubectl apply -f - > /dev/null

    if [[ $? -eq 0 ]]; then
        ok "Successfully created secret $DIODE_REDIS_SECRET in namespace $NAMESPACE"
    else
        error "Failed to create secret $DIODE_REDIS_SECRET in namespace $NAMESPACE"
    fi
fi

if kubectl get secret $DIODE_HYDRA_SECRET -n $NAMESPACE &>/dev/null; then
    ok "Secret $DIODE_HYDRA_SECRET already exists, skipping creation"
else
    kubectl create secret generic diode-hydra-secret --namespace $NAMESPACE \
      --from-literal=secretsCookie=$(generate_secret) \
      --from-literal=secretsSystem=$(generate_secret) \
      --from-literal=dsn=postgres://hydra:$HYDRA_POSTGRES_PASSWORD@$POSTGRES_HOSTNAME:$POSTGRES_PORT/hydra \
      --dry-run=client -o yaml | kubectl apply -f - > /dev/null

    if [[ $? -eq 0 ]]; then
        ok "Successfully created secret $DIODE_HYDRA_SECRET in namespace $NAMESPACE"
    else
        error "Failed to create secret $DIODE_HYDRA_SECRET in namespace $NAMESPACE"
    fi
fi

if kubectl get secret $DIODE_AUTH_OAUTH2_SECRET -n $NAMESPACE &>/dev/null; then
    ok "Secret $DIODE_AUTH_OAUTH2_SECRET already exists, skipping creation"
else
    kubectl create secret generic $DIODE_AUTH_OAUTH2_SECRET --namespace $NAMESPACE \
      --from-file=client-credentials.json=$PWD/client-credentials.json \
      --dry-run=client -o yaml | kubectl apply -f - > /dev/null

    if [[ $? -eq 0 ]]; then
        ok "Successfully created secret $DIODE_AUTH_OAUTH2_SECRET in namespace $NAMESPACE"
    else
        error "Failed to create secret $DIODE_AUTH_OAUTH2_SECRET in namespace $NAMESPACE"
    fi
fi

if kubectl get secret $DIODE_INGESTER_SECRET -n $NAMESPACE &>/dev/null; then
    ok "Secret $DIODE_INGESTER_SECRET already exists, skipping creation"
else
    kubectl create secret generic $DIODE_INGESTER_SECRET --namespace $NAMESPACE \
      --from-literal=REDIS_PASSWORD=$REDIS_PASSWORD \
      --dry-run=client -o yaml | kubectl apply -f - > /dev/null

    if [[ $? -eq 0 ]]; then
        ok "Successfully created secret $DIODE_INGESTER_SECRET in namespace $NAMESPACE"
    else
        error "Failed to create secret $DIODE_INGESTER_SECRET in namespace $NAMESPACE"
    fi
fi

if kubectl get secret $DIODE_RECONCILER_SECRET -n $NAMESPACE &>/dev/null; then
    ok "Secret $DIODE_RECONCILER_SECRET already exists, skipping creation"
else
    kubectl create secret generic $DIODE_RECONCILER_SECRET --namespace $NAMESPACE \
      --from-literal=REDIS_PASSWORD=$REDIS_PASSWORD \
      --from-literal=POSTGRES_PASSWORD=$DIODE_POSTGRES_PASSWORD \
      --from-literal=DIODE_TO_NETBOX_CLIENT_SECRET=$(jq -r '.[] | select(.client_id == "diode-to-netbox") | .client_secret' $PWD/client-credentials.json) \
      --dry-run=client -o yaml | kubectl apply -f - > /dev/null

    if [[ $? -eq 0 ]]; then
        ok "Successfully created secret $DIODE_RECONCILER_SECRET in namespace $NAMESPACE"
    else
        error "Failed to create secret $DIODE_RECONCILER_SECRET in namespace $NAMESPACE"
    fi
fi

# The NetBox Diode plugin authenticates as netbox-to-diode, a different client
# to the ingest one used by orb-agent. Surfacing it here avoids pasting the
# wrong secret and hitting "Failed to obtain access token".
NETBOX_TO_DIODE_CLIENT_ID="netbox-to-diode"

# Read it from the deployed Secret rather than the local file. When the Secret
# already existed we skipped creating it above, so a client-credentials.json
# regenerated in a fresh working directory holds values that were never applied
# to the cluster; printing those would hand out a credential that cannot
# authenticate. Fall back to the local file only when the Secret is unreadable.
NETBOX_TO_DIODE_CLIENT_SECRET=$(kubectl get secret "$DIODE_AUTH_OAUTH2_SECRET" -n "$NAMESPACE" \
  -o jsonpath='{.data.client-credentials\.json}' 2>/dev/null | base64 -d 2>/dev/null \
  | jq -r --arg id "$NETBOX_TO_DIODE_CLIENT_ID" \
      'try (.[] | select(.client_id == $id) | .client_secret) // empty' 2>/dev/null)

if [[ -z "$NETBOX_TO_DIODE_CLIENT_SECRET" ]]; then
    NETBOX_TO_DIODE_CLIENT_SECRET=$(jq -r --arg id "$NETBOX_TO_DIODE_CLIENT_ID" \
      '.[] | select(.client_id == $id) | .client_secret' "$PWD/client-credentials.json")
fi

echo "----------------------------------------"
ok "Environment setup completed!"
info "Configure the NetBox Diode plugin in configuration.py with:"
info "  netbox_to_diode_client_secret: $NETBOX_TO_DIODE_CLIENT_SECRET"
info "That is the $NETBOX_TO_DIODE_CLIENT_ID client secret, not the ingest one."
info "diode_target_override must point at the ingress, reachable from NetBox."
echo
info "You can now install the diode helm chart by running:"
if [[ "$CLUSTER_DOMAIN" == "cluster.local" ]]; then
    info "  helm install <RELEASE_NAME> diode/diode --namespace $NAMESPACE"
else
    info "  helm install <RELEASE_NAME> diode/diode --namespace $NAMESPACE \\"
    info "    --set global.clusterDomain=$CLUSTER_DOMAIN \\"
    info "    --set redis.clusterDomain=$CLUSTER_DOMAIN \\"
    info "    --set postgresql.clusterDomain=$CLUSTER_DOMAIN"
fi
