#!/bin/bash
set -e
set -o pipefail

usage="./scripts/quickCore.bash [-r|--rebuild] [-t|--role <role>] [-c|--clean]
	-r, --rebuild   Rebuild the container
	-t, --role      Specify the tower role (default: core)
	-c, --clean     Wipe the mongo container and file data
	-g, --group     Specify the group name to identify the stack (default: quick) e.g. will create a container named 'weblens-quick-core' and a mongo container named 'weblens-quick-core-mongo', assuming the tower role is 'core'
	"

towerRole="core"
groupName="dev"
uiPath="/app/web/"
shouldClean=false
shouldRebuild=false
doDevMode=false
doDynamic=false
local=false

arch=$(uname -m)

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-r" | "--rebuild")
        shouldRebuild=true
        ;;
    "-l" | "--local")
        local=true
        ;;
    "-t" | "--role")
        shift
        towerRole="$1"
        ;;
    "-c" | "--clean")
        shouldClean=true
        ;;
    "-d" | "--dev")
        doDevMode=true
        ;;
    "-y" | "--dynamic")
        doDynamic=true
        ;;
    "-s" | "--secure")
        export VITE_USE_HTTPS=true
        ./scripts/make-cert.bash
        ;;
    "-a" | "--arch")
        shift
        arch="$1"
        ;;
    "-g" | "--group")
        shift
        groupName="$1"
        ;;
    *)
        "Unknown argument: $1"
        echo "$usage"
        exit 1
        ;;
    esac
    shift
done

if [[ "$towerRole" == "" ]]; then
    echo "No tower role specified, defaulting to 'core'"
    towerRole="core"
fi

containerName="weblens-$groupName-$towerRole"
mongoName="$containerName-mongo"
imageName="static"
fsName="$groupName-$towerRole"
dockerfile="Dockerfile"
runCmd=""

echo "Using base name: $containerName"

if [[ $doDevMode == true ]]; then
    echo "Using development image"
    imageName="dev"
    dockerfile="dev.Dockerfile"
    uiPath="./weblens-vue/weblens-nuxt/.output/public"
fi

if [[ $doDynamic == true ]]; then
    WEBLENS_STATIC_CONTENT_PATH=./public
    runCmd="--dynamic"
fi

if [[ $shouldClean == true ]]; then
    docker stop "$mongoName" 2>/dev/null
    rm -rf ./_build/fs/"$fsName"
fi

if [[ $shouldRebuild == true ]]; then
    docker image rm -f ethrous/weblens:"$imageName-$arch" &>/dev/null || :
fi

printf "Removing old '%s' containers... " "$containerName"

docker stop "$containerName" 2>/dev/null || :
docker rm "$containerName" 2>/dev/null || :

printf "Done\n"

if ! docker network ls | grep weblens-net &>/dev/null; then
    printf "Creating weblens docker network... "
    docker network create weblens-net &>/dev/null
    printf "Done\n"
fi

# Build image if it doesn't exist
if [[ $local == false ]] && ! docker image ls | grep "$imageName-$arch" &>/dev/null; then
    echo "Image does not exist, building..."

    printf "\n---- gogogadgetdocker ----\n"
    if ! ./scripts/gogogadgetdocker.bash -t "$imageName" -a "$arch" -d "$dockerfile" | sed 's/^/  /'; then
        printf "\n---- FAILED ----\n\n"
        exit 1
    fi
    printf "\n---- gogogadgetdocker succeeded ----\n\n"
fi

if [[ $WEBLENS_LOG_LEVEL == "" ]]; then
    WEBLENS_LOG_LEVEL="debug"
fi

./scripts/start-mongo.sh "$mongoName" || exit 1

export WEBLENS_DATA_PATH="./_build/fs/$fsName/data"
export WEBLENS_LOG_FORMAT=dev
export WEBLENS_MONGODB_NAME="$containerName"

if [[ $local == true ]]; then
    echo "Running Weblens locally for development..."

    export WEBLENS_MONGODB_URI="mongodb://127.0.0.1:27018/?replicaSet=rs0&directConnection=true"

    ./scripts/devel.bash

    exit 0
fi

echo "Starting development container for Weblens..."

# FIXME: Tempporray fix for cache issues
# rm -rf ./_build/cache/"$fsName"

docker run \
    -t \
    -i \
    --rm \
    --name "$containerName" \
    -p 8080:8080 \
    -p 6060:6060 \
    -p 3001:3000 \
    -v ./_build/fs/"$fsName"/data:/data \
    -v ./_build/fs/"$fsName"/cache:/cache \
    -v .:/src \
    -v ../agno/:/agno \
    -v ./_build/cache/"$fsName"/go/mod:/go/pkg/mod \
    -v ./_build/cache/"$fsName"/go/build:/go/cache \
    -v ./_build/cache/"$fsName"/cargo/registry:/root/.cargo/registry \
    -v /src/weblens-vue/weblens-nuxt/node_modules \
    -e WEBLENS_MONGODB_URI=mongodb://"$containerName"-mongo:27017/?replicaSet=rs0 \
    -e WEBLENS_MONGODB_NAME="$containerName" \
    -e WEBLENS_INIT_ROLE="$towerRole" \
    -e WEBLENS_LOG_LEVEL="$WEBLENS_LOG_LEVEL" \
    -e WEBLENS_LOG_FORMAT="$WEBLENS_LOG_FORMAT" \
    -e WEBLENS_STATIC_CONTENT_PATH="${WEBLENS_STATIC_CONTENT_PATH:-/app/static}" \
    -e WEBLENS_DATA_PATH="$WEBLENS_DATA_PATH" \
    -e WEBLENS_UI_PATH="$uiPath" \
    -e VITE_PROXY_PORT=8080 \
    -e VITE_PROXY_HOST=127.0.0.1 \
    -e VITE_USE_HTTPS="$VITE_USE_HTTPS" \
    -e GOCACHE=/go/cache \
    -e GOMEMLIMIT=10GiB \
    --network weblens-net \
    --platform linux/"$arch" \
    ethrous/weblens:"$imageName-$arch" \
    "$runCmd"
