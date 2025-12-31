#!/bin/bash
set -euo pipefail

ensure_weblens_net() {
    if ! docker network ls | grep weblens-net; then
        sudo docker network create weblens-net >/dev/null 2>&1
    fi
}

export -f ensure_weblens_net
