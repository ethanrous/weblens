#!/bin/bash
set -euo pipefail

if [[ -z "${WEBLENS_ROOT:-}" ]]; then
    if [[ $(basename "$PWD") != "weblens" ]]; then
        echo "Please run this script from the repository root (weblens/) folder."
        exit 1
    fi

    WEBLENS_ROOT=$PWD
    export WEBLENS_ROOT
fi

if [[ -z "${WEBLENS_VERBOSE+x}" ]]; then
    echo "Use WEBLENS_VERBOSE=true to enable verbose output"
    export WEBLENS_VERBOSE=false
fi

LIB_DIR="$WEBLENS_ROOT/scripts/lib"

source "$LIB_DIR/meta.bash"
source "$LIB_DIR/build.bash"
source "$LIB_DIR/mongo.bash"
source "$LIB_DIR/docker-common.bash"
