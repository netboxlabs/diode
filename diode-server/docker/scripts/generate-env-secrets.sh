#!/usr/bin/env bash

set -euo pipefail

# Function to generate random secret
generate_secret() {
  head -c 32 /dev/urandom | base64 | tr -d '/\n'
}

env_file=$1
# Check if the file exists
if [ ! -f "$env_file" ]; then
    echo "Error: File '$env_file' does not exist."
    exit 1
fi

# Check if the file has .env extension
if [[ "$env_file" != *.env ]]; then
    echo "Error: File '$env_file' must have .env extension."
    exit 1
fi

# Detect OS and set sed command accordingly
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed_cmd="sed -i ''"
else
    # Linux and others
    sed_cmd="sed -i"
fi

# Check if the file contains any placeholders
if ! grep -q "<PLACEHOLDER_SECRET>" "$env_file"; then
    echo "File '$env_file' does not contain any placeholders to replace, skipping"
    exit 0
fi

# Create a temporary file
temp_file=$(mktemp)

# Process the file line by line
while IFS= read -r line; do
  if [[ $line == *"<PLACEHOLDER_SECRET>"* ]]; then
    # Generate a new secret for each placeholder
    new_secret=$(generate_secret)
    echo "${line/<PLACEHOLDER_SECRET>/$new_secret}"
  else
    echo "$line"
  fi
done < "$env_file" > "$temp_file"

# Replace the original file with the processed one
mv "$temp_file" "$env_file"

echo "Successfully replaced all placeholders in $env_file with unique secrets."
