#!/bin/bash
set -euo pipefail

# Stop services and clean up environment for development

source ./scripts/lib/all.bash

ignore=""
while [ "${1:-}" != "" ]; do
    case "$1" in
    "--ignore")
        shift
        ignore="$1"
        ;;
    *) ;;
    esac
    shift
done

mongo_stacks=$(docker compose ls | grep mongo | grep weblens | sed -E 's/weblens-([a-z\-]+)-mongo.*/\1/') || true

for mongo_stack in $mongo_stacks; do
    if [[ "$mongo_stack" == "$ignore" ]]; then
        echo "Skipping teardown of mongo stack: '$mongo_stack' (ignored)"
        continue
    fi
    cleanup_mongo --stack-name "$mongo_stack"
done

if is_embed_running; then
    stop_embed
elif is_embed_running --containerized false; then
    stop_embed
else
    echo "Embed service is not running."
fi
