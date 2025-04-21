#!/usr/bin/env bash

set -euo pipefail

# Function to generate random secret
generate_secret() {
  head -c 32 /dev/urandom | base64 | tr -d '/\n'
}

# Define client credentials in an associative array
declare -A CLIENT_CREDENTIALS
CLIENT_CREDENTIALS["diode-ingest"]="default:diode:ingest"
CLIENT_CREDENTIALS["diode-to-netbox"]="default:diode:netbox"
CLIENT_CREDENTIALS["netbox-to-diode"]="default:netbox:diode"

output="["
first=true

# Generate credentials for each client
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
