#!/bin/bash

does_agno_exist() {
    local agno_lib_dir="${WEBLENS_ROOT}/_build/lib"
    local agno_header_path="$agno_lib_dir/agno.h"
    local agno_lib_path="$agno_lib_dir/libagno.a"
    if [[ -f "$agno_header_path" ]] && [[ -f "$agno_lib_path" ]]; then
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

    # Remove any existing Agno library files before building
    rm -f "${lib_agno_path}"

    export AGNO_FEATURES="gpu,pdf-pdfium"

    if [[ ! -e "$lib_agno_path/libpdfium.a" ]]; then
        latest_release=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2026-03-10" '/repos/paulocoutinhox/pdfium-lib/releases?per_page=1' | jq -r '.[0].name')
        curl https://github.com/paulocoutinhox/pdfium-lib/releases/download/"$latest_release"/macos.tgz -L -o /tmp/pdfium.tgz
        tar -xzf /tmp/pdfium.tgz -C /tmp
        mkdir -p "$agno_lib_dir_path"
        mv /tmp/release/lib/libpdfium.a "$agno_lib_dir_path/libpdfium.a"
    fi

    pushd agno >/dev/null || return 1
    show_as_subtask "Building Agno to $lib_agno_path" "orange" -- "${WEBLENS_ROOT}/agno/build/sh/build-agno.bash" "$lib_agno_path"
    cp ./lib/agno.h "$agno_lib_dir_path" || return 1
    popd >/dev/null || return 1

    # Export CGO variables for subsequent Go builds.
    # AGNO_LIB_DIR overrides the default library path.
    local lib_dir="${AGNO_LIB_DIR:-$agno_lib_dir_path}"
    export CGO_CFLAGS="-I${WEBLENS_ROOT}/agno/lib"
    export CGO_LDFLAGS="-L${lib_dir} -lagno -lstdc++ -lm"
}
export -f build_agno

setup_agno_cgo() {
    local lib_dir="${AGNO_LIB_DIR:-${WEBLENS_ROOT}/_build/lib}"
    export CGO_CFLAGS="-I${WEBLENS_ROOT}/agno/lib"
    export CGO_LDFLAGS="-L${lib_dir} -lagno -lstdc++ -lm"
}
export -f setup_agno_cgo
