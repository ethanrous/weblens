#!/bin/bash

AGNO_VERSION="v0.0.8"

does_agno_exist() {
    local agno_lib_dir="${WEBLENS_ROOT}/_build/lib"
    local agno_lib_path="$agno_lib_dir/libagno.a"
    if [[ -f "$agno_lib_path" ]]; then
        return 0
    fi

    return 1
}

build_agno() {
    local agno_lib_dir_path
    if [[ -n "${1:-}" ]]; then
        agno_lib_dir_path="$1"
    else
        agno_lib_dir_path="${WEBLENS_ROOT}/_build/lib"
    fi
    mkdir -p "$agno_lib_dir_path"

    lib_agno_path="$agno_lib_dir_path/libagno.a"

    if [[ ! -e "$lib_agno_path" ]]; then
        latest_release=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2026-03-10" '/repos/ethanrous/agno/releases?per_page=1' | jq -r '.[0].name')
        if [[ "$latest_release" != "$AGNO_VERSION" ]]; then
            orange=$(get_color_code "orange")
            printf "\e[%sUpdate available: Latest release of agno is $latest_release, but we are loading $AGNO_VERSION. Consider updating AGNO_VERSION in ${BASH_SOURCE[0]} to the latest release.\n\e[0m" "$orange"
        fi

        curl https://github.com/ethanrous/agno/releases/download/"${AGNO_VERSION}"/libagno-macos-aarch64-gpu.a -L -o "$lib_agno_path"
    fi

    setup_agno_cgo
}
export -f build_agno

setup_agno_cgo() {
    echo "Setting up CGO flags for AGNO..."
    local lib_dir="${AGNO_LIB_DIR:-${WEBLENS_ROOT}/_build/lib}"
    export CGO_CFLAGS="-I${WEBLENS_ROOT}/agno/lib"
    export CGO_LDFLAGS="-L${lib_dir} -lagno -lstdc++ -lm"
}
export -f setup_agno_cgo
