#!/bin/bash

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
    *)
        echo "Unknown argument: $1"
        echo "$usage"
        exit 1
        ;;
    esac
    shift
done

printf "Checking connection to docker..."

sudo docker ps &>/dev/null
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

full_tag="${docker_tag}-${arch}"
echo "Using tag: $full_tag"

base_version=$(git rev-parse --short HEAD)
dirty_version=$(git diff | shasum -a 256)
WEBLENS_BUILD_VERSION="${base_version}-devel-${dirty_version:0:7}"
export WEBLENS_BUILD_VERSION

echo "Weblens build version: $WEBLENS_BUILD_VERSION"

printf "Building Weblens container..."

sudo docker rmi ethrous/weblens:"$full_tag" &>/dev/null
if ! sudo docker build --platform "linux/$arch" -t ethrous/weblens:"$full_tag" --build-arg WEBLENS_BUILD_VERSION="$WEBLENS_BUILD_VERSION" --build-arg ARCHITECTURE="$arch" -f "./docker/$dockerfile" .; then
    printf "Container build failed\n"
    exit 1
fi

if [[ $do_push == true ]]; then
    sudo docker push ethrous/weblens:"$full_tag"
fi

printf "\nBUILD COMPLETE. Container tag: ethrous/weblens:%s\n" "$full_tag"
