#!/bin/bash
set -euo pipefail

export WEBLENS_VERBOSE=true
source ./scripts/lib/all.bash

agnoLibPath="${1:-${WEBLENS_ROOT}/services/media/agno/lib/}"

build_agno "$agnoLibPath"
