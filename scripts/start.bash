#!/bin/bash
set -e
set -o pipefail

devel_weblens_locally() {
    echo "Running Weblens locally for development..."

    cd ./weblens-vue/weblens-nuxt || exit 1

    export WEBLENS_STATIC_CONTENT_PATH=./public
    export WEBLENS_UI_PATH=./weblens-vue/weblens-nuxt/.output/public
    export VITE_PROXY_PORT=8080
    export VITE_PROXY_HOST=127.0.0.1

    pnpm install
    if [[ ! -e ./.output/public/index.html ]]; then
        echo "Rebuilding UI..."
        pnpm generate
    fi
    pnpm dev 1>/dev/null &

    cd ../..

    air

    echo "Weblens development server finished..."
}

usage="./scripts/quickCore.bash [-r|--rebuild] [-t|--role <role>] [-c|--clean]
	-r, --rebuild   Rebuild the container
	-t, --role      Specify the tower role (default: core)
	-c, --clean     Wipe the mongo container and file data
	-g, --group     Specify the group name to identify the stack (default: quick) e.g. will create a container named 'weblens-quick-core' and a mongo container named 'weblens-quick-core-mongo', assuming the tower role is 'core'
	"

towerRole="core"
groupName="quick"
shouldClean=false
shouldRebuild=false
doDevMode=false

arch=$(uname -m)

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-r" | "--rebuild")
        shouldRebuild=true
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
        groupName="dev"
        ;;
    "-l" | "--local")
        devel_weblens_locally
        exit 0
        ;;
    "-s" | "--secure")
        export VITE_USE_HTTPS=true
        ./scripts/make-cert.bash
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
imageName="rc" # Release Candidate? It's not "prod" but not "dev" either. idk
fsName="$groupName-$towerRole"

echo "Using base name: $containerName"

if [[ $doDevMode == true ]]; then
    echo "Using development image"
    imageName="dev"
fi

if [[ $shouldClean == true ]]; then
    docker stop "$mongoName" 2>/dev/null
    rm -rf ./build/fs/"$fsName"
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
    docker network create weblens-net
    printf "Done\n"
fi

# Build image if it doesn't exist
if ! docker image ls | grep "$imageName-$arch" &>/dev/null; then
    echo "Image does not exist, building..."

    dockerfile="Dockerfile"
    if [[ $doDevMode == true ]]; then
        dockerfile="dev.Dockerfile"
    fi

    printf "\n---- gogogadgetdocker ----\n"
    if ! ./scripts/gogogadgetdocker.bash -t "$imageName" -a "$arch" -d "$dockerfile" | sed 's/^/  /'; then
        printf "\n---- FAILED ----\n\n"
        exit 1
    fi
    printf "\n---- gogogadgetdocker succeeded ----\n\n"
fi

./scripts/start-mongo.sh "$mongoName" || exit 1

echo "Starting development container for Weblens..."

docker run \
    -t \
    -i \
    --rm \
    --name "$containerName" \
    -p 8080:8080 \
    -p 3001:3000 \
    -v ./build/fs/"$fsName"/data:/data \
    -v ./build/fs/"$fsName"/cache:/cache \
    -v .:/src \
    -v ./build/cache/"$fsName"/go/mod:/go/pkg/mod \
    -v ./build/cache/"$fsName"/go/build:/go/cache \
    -e WEBLENS_MONGODB_URI=mongodb://"$containerName"-mongo:27017/?replicaSet=rs0 \
    -e WEBLENS_MONGODB_NAME="$containerName" \
    -e WEBLENS_INIT_ROLE="$towerRole" \
    -e WEBLENS_LOG_LEVEL=debug \
    -e WEBLENS_LOG_FORMAT=dev \
    -e VITE_USE_HTTPS="$VITE_USE_HTTPS" \
    -e GOCACHE=/go/cache \
    --network weblens-net \
    ethrous/weblens:"$imageName-$arch"
