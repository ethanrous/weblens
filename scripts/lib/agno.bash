#!/bin/bash

does_agno_exist() {
    local agno_lib_dir="${WEBLENS_ROOT}/services/media/agno/lib"
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
        agno_lib_dir_path="${WEBLENS_ROOT}/services/media/agno/lib"
    fi
    mkdir -p "$agno_lib_dir_path"

    lib_agno_path="$agno_lib_dir_path/libagno.a"

    # Remove any existing Agno library files before building
    rm -f "${lib_agno_path}"

    pushd agno >/dev/null || return 1
    show_as_subtask "Building Agno to $lib_agno_path" "orange" -- "${WEBLENS_ROOT}/agno/build/sh/build-agno.bash" "$lib_agno_path"
    cp ./lib/agno.h "$agno_lib_dir_path" || return 1
    popd >/dev/null || return 1
}
export -f build_agno
