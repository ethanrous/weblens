#!/bin/bash
set -euox pipefail

build_agno() {
    agno_lib_path="${PWD}/services/media/agno/lib/"
    mkdir -p "$agno_lib_path"
    pushd agno >/dev/null
    printf "Building \e[38;2;255;165;0mAgno...\e[0m\n"
    ./build/sh/buildAgno.bash "$agno_lib_path" 2>&1 | sed $'s/^/\e[38;2;255;165;0m| \e[0m/'
    cp ./lib/agno.h "$agno_lib_path"
    popd >/dev/null
}

export -f build_agno
