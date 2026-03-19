#!/bin/bash
set -euo pipefail

# Start necessary services for development

source ./scripts/lib/all.bash

role="core"

stack_name="$role"
mongo_port=27017

if ! is_mongo_running --stack-name "$stack_name"; then
    show_as_subtask "Launching mongo" "green" -- launch_mongo --stack-name "$stack_name" --mongo-port "$mongo_port"
else
    echo "MongoDB is already running."
fi

if [[ "$role" == "core" ]] && ! is_hdir_running --containerized false; then
    show_as_subtask "Launching HDIR" "orange" -- launch_hdir --containerized false
else
    echo "HDIR is already running."
fi
