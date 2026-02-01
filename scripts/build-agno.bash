#!/bin/bash
set -euo pipefail

export WEBLENS_VERBOSE=true
source ./scripts/lib/all.bash

build_agno "$1"
