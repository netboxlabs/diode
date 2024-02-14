#!/usr/bin/env bash

set -e

changed_files=$1
if [ -z "$changed_files" ]; then
  echo "Please provide a list of changed files as the first argument."
  exit 1
fi

distributor_deps=("diode-server/go.mod" "diode-server/go.sum" "diode-server/cmd/distributor/" "diode-server/distributor/" "diode-server/server/")
ingester_deps=("diode-server/go.mod" "diode-server/go.sum" "diode-server/cmd/ingester" "diode-server/ingester/" "diode-server/server/")
reconciler_deps=("diode-server/go.mod" "diode-server/go.sum" "diode-server/cmd/reconciler/" "diode-server/reconciler/" "diode-server/server/")

repo_root_dir=$(git rev-parse --show-toplevel)

declare -A changed_servers

IFS=';'

check_deps() {
    local -n deps=$1
    local file_path=$2
    file_dir="$( dirname "$file_path" )/"

    for dep in "${deps[@]}"; do
        if [[ "$dep" == "$file_path" ]]; then
            return 0
        fi
        if [[ "$file_dir" =~ ^$dep ]]; then
            return 0
        fi
    done
    return 1
}

read -ra changed_files_arr <<< "$changed_files"

for file in "${changed_files_arr[@]}"; do
  if [[ ! -f "$repo_root_dir/$file" ]]; then
    echo "File $repo_root_dir/$file does not exist, skipping"
    continue
  fi

  if check_deps distributor_deps "$file" && [[ ! -v changed_servers["distributor"] ]]; then
    changed_servers["distributor"]=1
  fi

  if check_deps ingester_deps "$file" && [[ ! -v changed_servers["ingester"] ]]; then
    changed_servers["ingester"]=1
  fi

  if check_deps reconciler_deps "$file" && [[ ! -v changed_servers["reconciler"] ]]; then
    changed_servers["reconciler"]=1
  fi
done

changed_servers_num=${#changed_servers[@]}

matrix_json="{\"include\":["

item_num=0
for server in "${!changed_servers[@]}"; do
  matrix_json+="{"
  matrix_json+="\"server\":\"${server}\""
  matrix_json+="}"
  if [[ $item_num -lt $changed_servers_num-1 ]]; then
    matrix_json+=","
  fi
  item_num=$((item_num+1))
done

matrix_json+="]}"

echo "matrix=$matrix_json" >> $GITHUB_OUTPUT
