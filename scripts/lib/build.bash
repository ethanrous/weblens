#!/bin/bash
set -euo pipefail

build_agno() {
    local agno_lib_path="${WEBLENS_ROOT}/services/media/agno/lib/"
    mkdir -p "$agno_lib_path"
    pushd agno >/dev/null
    ./build/sh/buildAgno.bash "$agno_lib_path" 2>&1 | show_as_subtask "Building Agno..." "orange"
    cp ./lib/agno.h "$agno_lib_path"
    popd >/dev/null
}
export -f build_agno

build_frontend() {
    pushd "${WEBLENS_ROOT}/weblens-vue/weblens-nuxt" >/dev/null

    pnpm install 2>&1 | show_as_subtask "Installing UI Dependencies..."
    pnpm run generate 2>&1 | show_as_subtask "Building UI..."

    popd >/dev/null
}
export -f build_frontend
