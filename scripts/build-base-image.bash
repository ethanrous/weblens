#!/bin/bash

if [[ ! -e ./scripts ]]; then
    echo "ERR Could not find ./scripts directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
    exit 1
fi

docker_tag=""
arch=$(uname -m)

# Once the image is built, push it to docker hub
do_push=false

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
        arch=amd64
        ;;
    "-h" | "--help")
        echo "$usage"
        exit 0
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

if [[ -z "$docker_tag" ]]; then
    base_version=$(git rev-parse --short HEAD)
    dirty_version=$(git diff | shasum -a 256)
    docker_tag="${base_version}-devel-${dirty_version:0:7}"
fi

echo "Using tag: $docker_tag"

printf "Building Weblens base image..."

sudo docker rmi ethrous/weblens-roux:"$docker_tag" &>/dev/null
if ! sudo docker build --platform "linux/$arch" -t ethrous/weblens-roux:"$docker_tag" --build-arg ARCHITECTURE="$arch" -f "./docker/Base.Dockerfile" .; then
    printf "Container build failed\n"
    exit 1
fi

if [[ $do_push == true ]]; then
    sudo docker push ethrous/weblens-roux:"$docker_tag"
fi

printf "\nBUILD COMPLETE. Container tag: ethrous/weblens-roux:%s\n" "$docker_tag"
