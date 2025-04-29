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
    echo "Usage: $0 NETBOX_HOST"
    echo
    echo "Example:"
    echo "  $0 http://localhost:8000"
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

NETBOX_HOST="$1"

if [[ -z "$NETBOX_HOST" ]]; then
    error "NETBOX_HOST is not provided"
    exit 1
fi

if docker compose version >/dev/null 2>&1; then
  DOCKER_COMPOSE="docker compose"
else
  DOCKER_COMPOSE="docker-compose"
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
    sed_args=(-i '')
else
    sed_args=(-i)
fi

ENV_FILE=".env"

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

set_secrets() {
    env_file="$1"
    if ! grep -q "<PLACEHOLDER_SECRET>" "$env_file"; then
        info "File '$env_file' does not contain any placeholders to replace, skipping"
        return
    fi

    temp_file=$(mktemp)
    trap 'rm -f "$temp_file"' RETURN

    while IFS= read -r line; do
        if [[ $line == *"<PLACEHOLDER_SECRET>"* ]]; then
            new_secret=$(generate_secret)
            echo "${line/<PLACEHOLDER_SECRET>/$new_secret}"
        else
            echo "$line"
        fi
    done < "$env_file" > "$temp_file"

    mv "$temp_file" "$env_file"
    ok "Successfully replaced all placeholders in $env_file with unique secrets."
}

# Download docker-compose.yaml
if [ ! -f "docker-compose.yaml" ]; then
    info "Downloading docker-compose.yaml"
    curl -sSfLo docker-compose.yaml https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/docker-compose.yaml
else
    ok "docker-compose.yaml file already exists"
fi

# Download nginx.conf
if [ ! -f "nginx/nginx.conf" ]; then
    info "Downloading nginx/nginx.conf"
    mkdir -p nginx
    curl -sSfLo nginx/nginx.conf https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/nginx/nginx.conf
else
    ok "nginx/nginx.conf file already exists"
fi

# Download .env
if [ ! -f "$ENV_FILE" ]; then
    info "Downloading $ENV_FILE"
    curl -sSfLo "$ENV_FILE" https://raw.githubusercontent.com/netboxlabs/diode/release/diode-server/docker/sample.env
else
    ok "$ENV_FILE file already exists"
fi

# Generate OAuth2 client credentials
if [ ! -f "oauth2/client/client-credentials.json" ]; then
    # Create oauth2/client directory
    if [ ! -d "oauth2/client" ]; then
        mkdir -p oauth2/client
    fi

    info "Generating OAuth2 client credentials in oauth2/client/client-credentials.json"
    generate_client_credentials > oauth2/client/client-credentials.json
else
    ok "Using existing OAuth2 client credentials in oauth2/client/client-credentials.json"
fi

# Set secrets in .env file
set_secrets "$ENV_FILE"

# Set DIODE_TO_NETBOX_CLIENT_SECRET
info "Setting DIODE_TO_NETBOX_CLIENT_SECRET in $ENV_FILE"
DIODE_TO_NETBOX_CLIENT_SECRET=$(jq -r '.[] | select(.client_id == "diode-to-netbox") | .client_secret' oauth2/client/client-credentials.json)

sed "${sed_args[@]}" "s|<PLACEHOLDER_DIODE_TO_NETBOX_CLIENT_SECRET>|$DIODE_TO_NETBOX_CLIENT_SECRET|g" "$ENV_FILE"

# Set NETBOX_HOST
if grep -q "NETBOX_HOST" "$ENV_FILE"; then
    info "Setting NETBOX_HOST in $ENV_FILE"
    sed "${sed_args[@]}" "s|<http://NETBOX_HOST>|$NETBOX_HOST|g" "$ENV_FILE"
fi

# Extract DIODE_NGINX_PORT from .env file, default to 8080 if not found
DIODE_NGINX_PORT=$(grep -oP 'DIODE_NGINX_PORT=\K[0-9]+' "$ENV_FILE" 2>/dev/null || echo "8080")
DIODE_TARGET="grpc://localhost:$DIODE_NGINX_PORT/diode"

# Get diode-ingest client credentials
DIODE_INGEST_CLIENT_ID="diode-ingest"
DIODE_INGEST_CLIENT_SECRET=$(jq -r '.[] | select(.client_id == "'$DIODE_INGEST_CLIENT_ID'") | .client_secret' oauth2/client/client-credentials.json)

echo "----------------------------------------"
ok "Environment setup completed!"
info "You can now start the diode by running:"
info "  $DOCKER_COMPOSE up -d"
info "Configure orb-agent with diode target $DIODE_TARGET to use the following credentials:"
info "  DIODE_CLIENT_ID:     $DIODE_INGEST_CLIENT_ID"
info "  DIODE_CLIENT_SECRET: $DIODE_INGEST_CLIENT_SECRET"
