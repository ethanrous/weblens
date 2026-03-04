#!/bin/bash
set -euo pipefail

# Stop services and clean up environment for development

source ./scripts/lib/all.bash

role="core"

stack_name="$role"

if is_mongo_running --stack-name "$stack_name"; then
    cleanup_mongo --stack-name "$stack_name"
else
    echo "MongoDB is not running."
fi

if is_hdir_running; then
    stop_hdir
else
    echo "HDIR is not running."
fi
