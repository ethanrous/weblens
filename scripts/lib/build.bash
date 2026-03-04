#!/bin/bash

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

    pushd "${WEBLENS_ROOT}/weblens-vue/weblens-nuxt" >/dev/null || return 1

    show_as_subtask "Installing UI Dependencies..." -- pnpm install
    VITE_BUILD=true show_as_subtask "Building UI..." -- pnpm run generate

    popd >/dev/null || return 1
}
export -f build_frontend

build_weblens_binary() {
    local lazy=false
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--lazy")
            lazy=true
            ;;
        *)
            "build_weblens_binary unknown argument: $1"
            exit 1
            ;;
        esac
        shift
    done

    local debug_bin="$WEBLENS_ROOT/_build/bin/weblens_debug"

    if [[ "$lazy" = true ]] && [[ -f "$debug_bin" ]]; then
        printf "Skipping Weblens binary build (lazy mode)...\n"
        return
    fi

    export CGO_CFLAGS='-g -O0'
    export CGO_CXXFLAGS='-g -O0'
    export CGO_LDFLAGS='-g'
    export CGO_ENABLED=1

    rm -f "$debug_bin"

    go build -v -gcflags=all="-N -l" -ldflags=-compressdwarf=false -o "$debug_bin" ./cmd/weblens/main.go 2>&1
}
export -f build_weblens_binary

build_hdir() {
    dockerc stop weblens-hdir 2>/dev/null || true
    dockerc image rm weblens_hdir 2>/dev/null || true
    dockerc build -f ./docker/hdir.Dockerfile -t ethrous/weblens_hdir .
}
export -f build_hdir
