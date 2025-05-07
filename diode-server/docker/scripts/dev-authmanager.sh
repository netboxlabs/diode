#!/usr/bin/env bash

quote_args() {
    local result=""
    for arg in "$@"; do
        result="${result:+$result }$(printf %q "$arg")"
    done
    echo "$result"
}

script_path=$(dirname "$0")
cd ${script_path}/../..
make authmanager_command="$(quote_args "$@")" docker-compose-dev-authmanager
