#!/bin/bash

source ./scripts/lib/all.bash

if [[ ! -e ./scripts ]]; then
    echo "ERR Could not find ./scripts directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
    exit 1
fi

mkdir -p ./_build/bin
mkdir -p ./_build/logs

docker_tag=devel_$(git rev-parse --abbrev-ref HEAD)
arch=$(uname -m)

# Once the image is built, push it to docker hub
do_push=false

# Skip testing
skip_tests=false

usage="TODO"
dockerfile="Dockerfile"
extras=""

# Path to a local agno checkout to build the image against, instead of the
# release pinned in go.mod
local_agno="${WEBLENS_LOCAL_AGNO:-}"

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-t" | "--tag")
        shift
        docker_tag=$1
        ;;
    "-a" | "--arch")
        shift
        arch=$1
        ;;
    "-p" | "--push")
        do_push=true
        arch=amd64
        ;;
    "-s" | "--skip-tests")
        skip_tests=true
        ;;
    "-h" | "--help")
        echo "$usage"
        exit 0
        ;;
    "-d" | "--dockerfile")
        shift
        dockerfile=$1
        ;;
    "--extras")
        shift
        extras=$1
        ;;
    "--local-agno")
        shift
        local_agno=$1
        ;;
    *)
        echo "Unknown argument: $1"
        echo "$usage"
        exit 1
        ;;
    esac
    shift
done

# `uname -m` names differ from the docker/GOARCH names the build expects
case "$arch" in
"x86_64")
    arch=amd64
    ;;
"aarch64")
    arch=arm64
    ;;
esac

printf "Checking connection to docker..."

dockerc ps &>/dev/null
docker_status=$?

if [[ $docker_status != 0 ]]; then
    printf " FAILED\n"
    echo "Aborting container build. Ensure docker is runnning"
    exit 1
else
    printf " PASS\n"
fi

if [[ $do_push == true && $skip_tests != true ]]; then
    printf "Running tests..."
    if ! ./scripts/test-weblens.bash -a &>./_build/logs/container-build-pretest.log; then
        printf " FAILED\n"
        cat ./_build/logs/container-build-pretest.log
        echo "Aborting container build. Ensure ./scripts/test-weblens.bash passes before building container"
        exit 1
    else
        printf " PASS\n"
    fi
fi

full_tag="ghcr.io/ethanrous/weblens:${docker_tag}"
echo "Using tag: $full_tag"

base_version=$(git rev-parse --short HEAD)
dirty_version=$(git diff | shasum -a 256)
WEBLENS_BUILD_VERSION="${base_version}-devel-${dirty_version:0:7}"

if [[ -n "$local_agno" ]]; then
    if ! build_agno_for_image "$local_agno" "$arch"; then
        exit 1
    fi

    extras="$extras --build-context agno=./_build/agno-image --build-arg AGNO_SOURCE=local"
fi

export WEBLENS_BUILD_VERSION

echo "Weblens build version: $WEBLENS_BUILD_VERSION"

printf "Building Weblens container..."

dockerc rmi "$full_tag" &>/dev/null
if ! dockerc build --platform "linux/$arch" -t "$full_tag" --build-arg WEBLENS_BUILD_VERSION="$WEBLENS_BUILD_VERSION" --build-arg ARCHITECTURE="$arch" $extras -f "./docker/$dockerfile" .; then
    printf "Container build failed\n"
    exit 1
fi

if [[ $do_push == true ]]; then
    if ! dockerc push "$full_tag"; then
        printf "Container push failed\n"
        exit 1
    fi
fi

printf "\nBUILD COMPLETE. Container tag: %s\n" "$full_tag"
