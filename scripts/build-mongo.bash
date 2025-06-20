#!/bin/bash
if [[ ! -e ./scripts ]]; then
    echo "ERR Could not find ./scripts directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
    exit 1
fi

mkdir -p ./build/logs

arch=$(uname -m)

# Once the image is built, push it to docker hub
do_push=false

usage="TODO"

while [ "${1:-}" != "" ]; do
    case "$1" in
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

if [ $docker_status != 0 ]; then
    printf " FAILED\n"
    echo "Aborting image build. Ensure docker is runnning"
    exit 1
else
    printf " PASS\n"
fi

printf "Building Weblens mongodb image..."

sudo docker rmi ethrous/weblens-mongo:latest &>/dev/null
if ! sudo docker build --platform "linux/$arch" -t ethrous/weblens-mongo:latest --build-arg ARCHITECTURE="$arch" -f "./docker/mongo.Dockerfile" .; then
    printf "mongodb image build failed\n"
    exit 1
fi

if ! docker image inspect ethrous/weblens-mongo:latest; then
    printf "mongodb image build failed (image does not exist)\n"
    exit 1
fi

if [ $do_push == true ]; then
    sudo docker push ethrous/weblens-mongo:latest
fi

printf "\nBUILD COMPLETE. Image tag: ethrous/weblens-mongo:latest\n"
