#!/bin/bash
set -euo pipefail

source ./scripts/lib/all.bash

run_native_tests() {
    build_frontend ${lazy:=false}

    target="${1:-./...}" # Default to ./... if no target specified

    # Go is very picky about whitespace, so we need to strip it out
    target=$(awk '{$1=$1};1' <<<"$target")

    touch /tmp/weblens.env

    WEBLENS_LOG_PATH=$(get_log_file "weblens-test")

    export WEBLENS_ENV_PATH=/tmp/weblens.env
    export WEBLENS_DO_CACHE=false
    export WEBLENS_MONGODB_URI=${WEBLENS_MONGODB_URI:-"mongodb://127.0.0.1:27019/?replicaSet=rs0&directConnection=true"}
    export WEBLENS_LOG_LEVEL="${WEBLENS_LOG_LEVEL:-debug}"
    export WEBLENS_LOG_FORMAT="dev"
    export CGO_LDFLAGS="-w"

    echo "Running tests with mongo [$WEBLENS_MONGODB_URI] and test target: [$target]"

    mkdir -p ./_build/cover/

    # shellcheck disable=SC2086
    if ! go test -cover -race -coverprofile=_build/cover/coverage.out -coverpkg ./... -tags=test ${target}; then
        printf "\n\nTESTS FAILED. Last 200 lines of log:\n\n"
        tail -200 "$WEBLENS_LOG_PATH" || true
        exit 1
    fi
    # 2>&1 | grep -v -e "=== RUN" -e "=== PAUSE" -e "--- PASS" -e "coverage:" -e "=== CONT" -e "ld: warning:"

    portable_sed '/github\.com\/ethanrous\/weblens\/api/d' ./_build/cover/coverage.out
}

run_container_tests() {
    rm -rf ./_build/fs/test-container

    if ! dockerc run --rm --platform="linux/amd64" \
        --network weblens-net \
        -v ./_build/fs/test-container/data:/data \
        -v ./_build/fs/test-container/cache:/cache \
        -v ./_build/cache/test/go/build:/tmp/go-cache \
        -v ./_build/cover:/cover \
        -v ./:/src \
        -v /src/weblens-vue/weblens-nuxt/node_modules \
        -v /src/build \
        -e WEBLENS_MONGODB_URI="mongodb://weblens-test-mongo:27017/?replicaSet=rs0" \
        ethrous/weblens-roux":$baseVersion" /src/scripts/test-weblens.bash "${tests}"; then
        echo "Tests failed, exiting..."
        exit 1
    fi

}

tests=""
baseVersion="v0"
containerize=false
lazy=true

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-c" | "--containerize")
        containerize=true
        ;;
    "-b" | "--base-version")
        shift
        baseVersion="$1"
        ;;
    "--no-lazy")
        lazy=false
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
    cleanup_mongo --stack-name "test" | show_as_subtask "Resetting mongo testing volumes..." "green"
    launch_mongo --stack-name "test" --mongo-port 27019 | show_as_subtask "Launching mongo..." "green"
fi

if [[ "$containerize" = false ]]; then
    if [[ "$lazy" = false ]] || [[ ! -e "$WEBLENS_ROOT/services/media/agno/lib/libagno.a" ]]; then
        build_agno
    else
        printf "Skipping Agno build (lazy mode)...\n"
    fi

    run_native_tests "${tests}"
else
    run_container_tests
fi
