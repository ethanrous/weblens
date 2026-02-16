#!/bin/bash

does_agno_exist() {
    local agno_lib_path="${WEBLENS_ROOT}/services/media/agno/lib/agno.h"
    if [[ -f "$agno_lib_path" ]]; then
        return 0
    fi

    return 1
}

build_agno() {
    local agno_lib_path
    if [[ -n "${1:-}" ]]; then
        agno_lib_path="$1"
    else
        agno_lib_path="${WEBLENS_ROOT}/services/media/agno/lib/"
    fi
    mkdir -p "$agno_lib_path"

    # Remove any existing Agno library files before building
    rm -f "${agno_lib_path}/libagno.a"

    pushd agno >/dev/null || return 1
    show_as_subtask "Building Agno to $agno_lib_path..." "orange" -- "${WEBLENS_ROOT}/agno/build/sh/build-agno.bash" "$agno_lib_path"
    cp ./lib/agno.h "$agno_lib_path" || return 1
    popd >/dev/null || return 1
}
export -f build_agno
