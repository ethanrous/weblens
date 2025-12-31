#!/bin/bash
set -euox pipefail

source ./scripts/build-agno.bash

usage="Usage: $0 [-n|--native [package_target]]"

run_native_tests() {
    pushd ./weblens-vue/weblens-nuxt
    echo "Installing UI Dependencies..."
    pnpm install >/dev/null 2>&1

    echo "Building UI..."
    pnpm run generate >/dev/null 2>&1

    popd >/dev/null

    target="${1:-./...}" # Default to ./... if no target specified

    # Go is very picky about whitespace, so we need to strip it out
    target=$(awk '{$1=$1};1' <<<"$target")

    touch /tmp/weblens.env
    export WEBLENS_ENV_PATH=/tmp/weblens.env
    export WEBLENS_DO_CACHE=false
    export WEBLENS_MONGODB_URI=${WEBLENS_MONGODB_URI:-"mongodb://127.0.0.1:27018/?replicaSet=rs0&directConnection=true"}

    echo "Running tests with mongo [$WEBLENS_MONGODB_URI] and test target: [$target]"

    mkdir -p ./_build/cover/
    go test -v -cover -race -coverprofile=_build/cover/coverage.out "$target" | grep -v -e "=== RUN" -e "=== PAUSE" -e "--- PASS"
    exit $?
}

run_container_tests() {
    rm -rf ./_build/fs/test-container

    if ! sudo docker run --rm --platform="linux/amd64" \
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

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-c" | "--containerize")
        containerize=true
        ;;
    "-b" | "--base-version")
        shift
        baseVersion="$1"
        ;;
    "-h" | "--help")
        echo "$usage"
        exit 0
        ;;
    *)
        tests="$tests$1 "
        ;;
    esac
    shift
done

sudo docker stop weblens-test-mongo >/dev/null 2>&1 || true
sudo docker rm weblens-test-mongo >/dev/null 2>&1 || true
sudo docker volume rm weblens-test-mongo >/dev/null 2>&1 || true
./scripts/start-mongo.sh "weblens-test-mongo"
if [ "$containerize" = false ]; then
    build_agno
    run_native_tests "${tests}"
else
    run_container_tests
fi
