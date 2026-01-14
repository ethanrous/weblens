#!/bin/bash
set -euo pipefail

source ./scripts/lib/all.bash

start_ui() {
    export VITE_PROXY_PORT="$WEBLENS_PORT"
    echo "Starting Weblens development UI..."

    pushd "$WEBLENS_ROOT/weblens-vue/weblens-nuxt" >/dev/null
    pnpm dev &>/dev/null &
    popd >/dev/null
}

devel_weblens_locally() {
    if [[ ! -e "$WEBLENS_ROOT/services/media/agno/lib/libagno.a" ]]; then
        build_agno
    fi

    build_frontend true
    start_ui

    air
}

usage() {
    echo "Usage: $0 [-r|--dynamic]"
    echo "  -r, --dynamic   Enable dynamic mode"
}

dynamic=false
while [ "${1:-}" != "" ]; do
    case "$1" in
    "-r" | "--dynamic")
        dynamic=true
        ;;
    *)
        "Unknown argument: $1"
        usage
        exit 1
        ;;
    esac
    shift
done

if [[ $dynamic == true ]]; then
    echo "Dynamic mode enabled"
    devel_weblens_locally
else
    debug_bin="$WEBLENS_ROOT/_build/bin/weblens_debug"
    build_weblens_binary
    "$debug_bin"
fi
