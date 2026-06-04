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

if [[ "$role" == "core" ]] && ! is_embed_running --containerized false; then
    show_as_subtask "Launching embed service" "orange" -- launch_embed --containerized false
else
    echo "Embed service is already running."
fi
