#!/bin/bash
set -euo pipefail

build_agno() {
    local agno_lib_path="${WEBLENS_ROOT}/services/media/agno/lib/"
    mkdir -p "$agno_lib_path"
    pushd agno >/dev/null
    ./build/sh/build-agno.bash "$agno_lib_path" 2>&1 | show_as_subtask "Building Agno..." "orange"
    cp ./lib/agno.h "$agno_lib_path"
    popd >/dev/null
}
export -f build_agno

build_frontend() {
    local lazy
    if [[ ! -z "${1:-}" ]]; then
        lazy="$1"
    else
        lazy=false
    fi

    if [[ "$lazy" = true ]] && [[ -e "${WEBLENS_ROOT}/weblens-vue/weblens-nuxt/.output/public/index.html" ]]; then
        printf "Skipping UI build (lazy mode)...\n"
        return
    fi

    pushd "${WEBLENS_ROOT}/weblens-vue/weblens-nuxt" >/dev/null

    pnpm install 2>&1 | show_as_subtask "Installing UI Dependencies..."
    pnpm run generate 2>&1 | show_as_subtask "Building UI..."

    popd >/dev/null
}
export -f build_frontend

build_weblens_binary() {
    debug_bin="$WEBLENS_ROOT/_build/bin/weblens_debug"

    export CGO_CFLAGS='-g -O0'
    export CGO_CXXFLAGS='-g -O0'
    export CGO_LDFLAGS='-g'
    export CGO_ENABLED=1

    rm -f "$debug_bin"

    go build -v -gcflags=all="-N -l" -ldflags=-compressdwarf=false -o "$debug_bin" ./cmd/weblens/main.go 2>&1
}
export -f build_weblens_binary
