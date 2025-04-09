#!/bin/bash

set -e

# Constants
CREDENTIALS_FILE="/etc/config/oauth2/clients/client-credentials.json"
TEMP_CREDENTIALS_FILE="/tmp/client-credentials.json"

# Wait for Hydra to be ready
sleep 3

# Function to generate random secret
generate_secret() {
  head -c 32 /dev/urandom | base64 | tr -d '/\n'
}

# Function to get or create client
get_or_create_client() {
  local client_id=$1
  local client_secret=$2
  local scope=$3
  local exists_in_credentials=false
  local exists_in_hydra=false
  local client_output=""

  # Check if client already exists in the credentials file
  if [ -f "$CREDENTIALS_FILE" ]; then
    if jq -e ".[] | select(.client_id == \"$client_id\")" "$CREDENTIALS_FILE" > /dev/null 2>&1; then
      exists_in_credentials=true
      
      # Extract the client object using jq
      client_output=$(jq -c ".[] | select(.client_id == \"$client_id\")" "$CREDENTIALS_FILE")
            
      # Extract client secret using jq
      client_secret=$(echo "$client_output" | jq -r '.client_secret')
    fi
  fi

  # Check if client exists in Hydra
  if hydra get oauth2-client $client_id --endpoint $HYDRA_ADMIN_URL >/dev/null 2>&1; then
    exists_in_hydra=true
  fi

  # Log client existence status
  if [ "$exists_in_credentials" = true ] && [ "$exists_in_hydra" = true ]; then
    echo "INFO: client $client_id exists in both Hydra and credentials file"
  elif [ "$exists_in_credentials" = true ] && [ "$exists_in_hydra" = false ]; then
    echo "INFO: client $client_id exists in credentials file but not in Hydra"
  elif [ "$exists_in_credentials" = false ] && [ "$exists_in_hydra" = true ]; then
    echo "WARN: client $client_id exists in Hydra but not in credentials file"
  else
    echo "INFO: client $client_id doesn't exist in either Hydra or credentials file"
  fi

  # Create new client if it doesn't exist in Hydra
  if [ "$exists_in_hydra" = false ]; then
    client_output=$(hydra create oauth2-client --endpoint $HYDRA_ADMIN_URL \
      --id $client_id \
      --secret $client_secret \
      --grant-type "client_credentials" \
      --response-type "token" \
      --scope $scope \
      --format json)

    # Filter client_output to keep only the specified fields
    client_output=$(echo "$client_output" | jq '{
      client_id: .client_id,
      client_secret: .client_secret,
      grant_types: .grant_types,
      scope: .scope
    }')
    echo "INFO: client $client_id created"
  fi

  # Add the client output to the temp file
  if [ -n "$client_output" ]; then
    # Check if client already exists in temp file
    if ! jq -e ".[] | select(.client_id == \"$client_id\")" "$TEMP_CREDENTIALS_FILE" > /dev/null 2>&1; then
      jq --argjson new_client "$client_output" '. += [$new_client]' "$TEMP_CREDENTIALS_FILE" > "$TEMP_CREDENTIALS_FILE.tmp" && mv "$TEMP_CREDENTIALS_FILE.tmp" "$TEMP_CREDENTIALS_FILE"
    fi
  fi
}

# Initialize credentials file in temp directory
if [ -f "$CREDENTIALS_FILE" ]; then
  # Copy existing clients to temp file
  cp "$CREDENTIALS_FILE" "$TEMP_CREDENTIALS_FILE"
else
  # Create empty JSON array
  echo "[]" > "$TEMP_CREDENTIALS_FILE"
fi

# Check if TEMP_CREDENTIALS_FILE is empty or contains invalid JSON
if [ ! -s "$TEMP_CREDENTIALS_FILE" ] || ! jq empty "$TEMP_CREDENTIALS_FILE" 2>/dev/null; then
  echo "[]" > "$TEMP_CREDENTIALS_FILE"
fi


# Create client credentials

# Ingest
get_or_create_client \
  "diode-ingest-1" \
  $(generate_secret) \
  "default:diode:ingest"

# Diode to NetBox
get_or_create_client \
  "diode-to-netbox" \
  $(generate_secret) \
  "default:diode:netbox"

# NetBox to Diode
get_or_create_client \
  "netbox-to-diode" \
  $(generate_secret) \
  "default:netbox:diode"

if [ "$(cat "$TEMP_CREDENTIALS_FILE")" != "[]" ]; then
  cat "$TEMP_CREDENTIALS_FILE" > "$CREDENTIALS_FILE"
fi
rm "$TEMP_CREDENTIALS_FILE"
