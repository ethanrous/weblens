#!/bin/bash
set -euo pipefail

# Stop services and clean up environment for development

source ./scripts/lib/all.bash

mongo_stacks=$(docker compose ls | grep mongo | grep weblens | sed -E 's/weblens-([a-z\-]+)-mongo.*/\1/')

for mongo_stack in $mongo_stacks; do
    cleanup_mongo --stack-name "$mongo_stack"
done

if is_hdir_running; then
    stop_hdir
else
    echo "HDIR is not running."
fi
