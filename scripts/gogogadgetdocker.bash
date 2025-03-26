#!/bin/bash

if [[ ! -e ./scripts ]]; then
    echo "ERR Could not find ./scripts directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
    exit 1
fi

mkdir -p ./build/bin
mkdir -p ./build/logs

docker_tag=devel_$(git rev-parse --abbrev-ref HEAD)
arch="amd64"

# Once the container is build, push it to docker hub
do_push=false

# Skip testing
skip_tests=true

# The base image to build from. Alpine is smaller, debian allows for cuda accelerated ffmpeg
base_image="alpine"

usage="TODO"

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
        ;;
    "-s" | "--skip-tests")
        skip_tests=true
        ;;
    "--base-image")
        shift
        base_image=$1
        ;;
    "-h" | "--help")
        echo "$usage"
        exit 0
        ;;
    *)
        "Unknown argument: $1"
        echo "$usage"
        exit 1
        ;;
    esac
    shift
done

sudo docker ps &>/dev/null
docker_status=$?

printf "Checking connection to docker..."
if [ $docker_status != 0 ]; then
    printf " FAILED\n"
    echo "Aborting container build. Ensure docker is runnning"
    exit 1
else
    printf " PASS\n"
fi

if [ ! $skip_tests == true ]; then
    printf "Running tests..."
    if ! ./scripts/testWeblens -a &>./build/logs/container-build-pretest.log; then
        printf " FAILED\n"
        cat ./build/logs/container-build-pretest.log
        echo "Aborting container build. Ensure ./scripts/testWeblens passes before building container"
        exit 1
    else
        printf " PASS\n"
    fi
fi

if [[ ! -e ./build/ffmpeg ]]; then
    docker run --platform linux/amd64 -v ./scripts/buildFfmpeg.sh:/buildFfmpeg.sh -v ./build:/build --rm alpine /buildFfmpeg.sh
fi

df_path="./docker/Dockerfile"
if [ "$base_image" == "debian" ]; then
    df_path="./docker/Debian.Dockerfile"
fi

full_tag="${docker_tag}-${base_image}-${arch}"
echo "Using tag: $full_tag"

base_version=$(git rev-parse --short HEAD)
dirty_version=$(git diff | shasum -a 256)
WEBLENS_BUILD_VERSION="${base_version}-devel-${dirty_version:0:7}"
export WEBLENS_BUILD_VERSION

echo "Weblens build version: $WEBLENS_BUILD_VERSION"

printf "Building Weblens container..."
sudo docker rmi ethrous/weblens:"$full_tag" &>/dev/null
sudo docker build --platform "linux/$arch" -t ethrous/weblens:"$full_tag" --build-arg WEBLENS_BUILD_VERSION="$WEBLENS_BUILD_VERSION" --build-arg ARCHITECTURE="$arch" -f $df_path .

if [ $do_push == true ]; then
    sudo docker push ethrous/weblens:"$full_tag"
fi

printf "\nBUILD COMPLETE. Container tag: ethrous/weblens:%s\n" "$full_tag"
