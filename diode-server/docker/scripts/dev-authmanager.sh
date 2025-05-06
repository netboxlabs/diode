#!/usr/bin/env bash

join_with_spaces() {
  local IFS=" "
  echo "$*"
}

script_path=$(dirname "$0")
cd ${script_path}/../..
make authmanager_command="$(join_with_spaces "$@")" docker-compose-dev-authmanager
