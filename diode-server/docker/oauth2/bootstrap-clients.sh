#!/usr/bin/env bash

set -euo pipefail

# Constants
CREDENTIALS_FILE="/etc/config/oauth2/client/client-credentials.json"

# Create the credentials file if it doesn't exist
if [ ! -f "$CREDENTIALS_FILE" ]; then
  echo "ERROR: credentials file $CREDENTIALS_FILE not found"
  exit 1
fi

# Wait for Hydra to be ready
sleep 3

# Function to create client
create_client() {
  local client_id=$1
  local client_secret=$2
  local scope=$3
  local exists_in_hydra=false

  # Check if client exists in Hydra
  if hydra get oauth2-client $client_id --endpoint $HYDRA_ADMIN_URL >/dev/null 2>&1; then
    exists_in_hydra=true
  fi

  # Log client existence status
  if [ "$exists_in_hydra" = true ]; then
    echo "INFO: client $client_id exists in Hydra"
    return 0
  fi

  # Create new client if it doesn't exist in Hydra
  if [ "$exists_in_hydra" = false ]; then
    client_output=$(hydra create oauth2-client --endpoint $HYDRA_ADMIN_URL \
      --id $client_id \
      --secret $client_secret \
      --grant-type "client_credentials" \
      --response-type "token" \
      --scope "$scope" \
      --token-endpoint-auth-method "client_secret_post" \
      --format json)

    echo "INFO: client $client_id created"
  fi
}

# Load client credentials
jq -c '.[]' "$CREDENTIALS_FILE" | while read -r client; do
  client_id=$(echo "$client" | jq -r '.client_id')
  client_secret=$(echo "$client" | jq -r '.client_secret')
  scope=$(echo "$client" | jq -r '.scope')
  create_client "$client_id" "$client_secret" "$scope"
done
