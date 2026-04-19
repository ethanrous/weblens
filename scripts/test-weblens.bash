#!/bin/bash
set -euo pipefail

if [[ "${WEBLENS_DEBUG_SHELL:-0}" == "1" ]]; then
    set -x
fi

source ./scripts/lib/all.bash

run_native_tests() {
    build_frontend ${lazy:=false}

    target="${1:-./...}" # Default to ./... if no target specified
    run="${2:-}"

    # Go is very picky about whitespace, so we need to strip it out
    target=$(awk '{$1=$1};1' <<<"$target")

    # If the target does not start with "./" or "github.com/ethanrous/weblens", prepend "./" to it
    if [[ ! "$target" =~ ^(\./|github\.com/ethanrous/weblens) ]]; then
        target="./$target"
    fi

    # Set up environment variables for testing
    export WEBLENS_DO_CACHE=false
    export WEBLENS_MONGODB_URI=${WEBLENS_MONGODB_URI:-"mongodb://127.0.0.1:27019/?directConnection=true"}
    export WEBLENS_LOG_LEVEL="${WEBLENS_LOG_LEVEL:-debug}"
    export WEBLENS_LOG_FORMAT="dev"

    echo "Running tests with mongo [$WEBLENS_MONGODB_URI] and test target: [$target]"

    # Ensure coverage directory exists
    mkdir -p ./_build/cover/

    # Install gotestsum for better test output formatting
    go install gotest.tools/gotestsum@latest

    gotestsum --packages="${target}" -- -cover -race -coverprofile=_build/cover/coverage.out -json -timeout=1m -coverpkg ./... -tags=test -run "${run}"

    portable_sed '/github\.com\/ethanrous\/weblens\/api/d' ./_build/cover/coverage.out
}

tests=""
run=""
baseVersion="v0"
containerize=false
lazy=true

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-b" | "--base-version")
        shift
        baseVersion="$1"
        ;;
    "--no-lazy")
        lazy=false
        ;;
    "-run")
        shift
        run="$1"
        ;;
    "-h" | "--help")
        usage="Usage: $0 [-n|--native [package_target]]"
        echo "$usage"
        exit 0
        ;;
    *)
        tests="$tests$1 "
        ;;
    esac
    shift
done

if [[ "$lazy" = true ]] && is_mongo_running --stack-name "test"; then
    printf "Skipping mongo container re-deploy (lazy mode)...\n"
else
    show_as_subtask "Resetting mongo testing volumes" "green" -- cleanup_mongo --stack-name "test"
    show_as_subtask "Launching mongo" "green" -- launch_mongo --stack-name "test" --mongo-port 27019
fi

build_agno

run_native_tests "${tests}" "${run}"
