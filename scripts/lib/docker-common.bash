#!/bin/bash

ensure_weblens_net() {
    if ! dockerc network ls | grep weblens-net; then
        dockerc network create weblens-net >/dev/null 2>&1
    fi
}
export -f ensure_weblens_net

# Wrapper to run docker commands with sudo, when needed
dockerc() {
    # Docker on macos does not need sudo
    if [[ "$(uname -s)" == "Darwin" ]]; then
        docker "${@}"
        return
    fi

    # FIXME: Temp for testing
    docker "${@}"
    # sudo -E docker "${@}"
}
export -f dockerc
