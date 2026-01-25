#!/bin/bash
set -euo pipefail

source ./scripts/lib/all.bash

usage="./scripts/quickCore.bash [-r|--rebuild] [-t|--role <role>] [-c|--clean]
	-r, --rebuild   Rebuild the container
	-t, --role      Specify the tower role (default: core)
	-c, --clean     Wipe the mongo container and file data
	-g, --group     Specify the group name to identify the stack (default: quick) e.g. will create a container named 'weblens-quick-core' and a mongo container named 'weblens-quick-core-mongo', assuming the tower role is 'core'
	"

start_ui() {
    export VITE_PROXY_PORT="$WEBLENS_PORT"
    export VITE_PORT=$((WEBLENS_PORT - 5080))
    echo "Starting Weblens development UI..."

    pushd "$WEBLENS_ROOT/weblens-vue/weblens-nuxt" >/dev/null
    pnpm dev &
    popd >/dev/null
}

devel_weblens_locally() {
    if [[ ! -e "$WEBLENS_ROOT/services/media/agno/lib/libagno.a" ]]; then
        build_agno
    fi

    build_frontend true
    start_ui

    air
}

tower_role="core"
group_name="dev"
uiPath="/app/web/"
shouldClean=false
shouldRebuild=false
doDevMode=false
doDynamic=true
local=true

arch=$(uname -m)

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-r" | "--rebuild")
        shouldRebuild=true
        ;;
    "-z" | "--containerize")
        local=false
        ;;
    "-t" | "--role")
        shift
        tower_role="$1"
        ;;
    "-c" | "--clean")
        shouldClean=true
        ;;
    "-d" | "--dev")
        doDevMode=true
        ;;
    "-s" | "--static")
        doDynamic=false
        ;;
    "-a" | "--arch")
        shift
        arch="$1"
        ;;
    "-g" | "--group")
        shift
        group_name="$1"
        ;;
    *)
        "Unknown argument: $1"
        echo "$usage"
        exit 1
        ;;
    esac
    shift
done

mongo_port=27017
weblens_port=8080

if [[ "$tower_role" == "" ]]; then
    echo "No tower role specified, defaulting to 'core'"
    tower_role="core"
elif [[ "$tower_role" == "backup" ]]; then
    mongo_port=27018
    weblens_port=8081
fi

container_name="weblens-$group_name-$tower_role"
mongo_name="$container_name-mongo"
image_name="static"
fs_name="$group_name-$tower_role"
dockerfile="Dockerfile"
run_cmd=""

echo "Using base name: $container_name"

if [[ $doDevMode == true ]]; then
    echo "Using development image"
    image_name="dev"
    dockerfile="dev.Dockerfile"
    uiPath="./weblens-vue/weblens-nuxt/.output/public"
fi

if [[ $doDynamic == true ]]; then
    WEBLENS_STATIC_CONTENT_PATH=./public
    run_cmd="--dynamic"
fi

if [[ $shouldClean == true ]]; then
    docker stop "$mongo_name" 2>/dev/null
    rm -rf ./_build/fs/"$fs_name"
fi

if [[ $shouldRebuild == true ]]; then
    docker image rm -f ethrous/weblens:"$image_name-$arch" &>/dev/null || :
fi

if [[ ! $(is_mongo_running "$mongo_name") ]]; then
    launch_mongo "$container_name" "$mongo_port" | show_as_subtask "Launching mongo..." "green"
fi
# cleanup_mongo | show_as_subtask "Killing old mongo containers..." "green"

# Build image if it doesn't exist
if [[ $local == false ]] && ! docker image ls | grep "$image_name-$arch" &>/dev/null; then
    printf "Image does not exist, \e[34mbuilding...\e[0m\n"

    if ! ./scripts/gogogadgetdocker.bash -t "$image_name" -a "$arch" -d "$dockerfile" | sed $'s/^/\e[34m| \e[0m/'; then
        printf "\n---- FAILED ----\n\n"
        exit 1
    fi
fi

WEBLENS_LOG_LEVEL="${WEBLENS_LOG_LEVEL:-debug}"

export WEBLENS_MONGODB_URI="mongodb://127.0.0.1:$mongo_port/?replicaSet=rs0&directConnection=true"
export WEBLENS_DATA_PATH="./_build/fs/$fs_name/data"
export WEBLENS_LOG_FORMAT=dev
export WEBLENS_MONGODB_NAME="$container_name"
export WEBLENS_DO_CACHE=false
export WEBLENS_PORT=$weblens_port

if [[ $local == true ]]; then
    printf "Running \e[34mWeblens\e[0m locally for development...\n"

    devel_weblens_locally
    # ./scripts/devel.bash "$run_cmd" | sed $'s/^/\e[34m| \e[0m/'

    exit 0
fi

printf "Starting development container for \e[34mWeblens...\e[0m\n"

docker run \
    -t \
    -i \
    --rm \
    --name "$container_name" \
    -p 8080:8080 \
    -p 6060:6060 \
    -p 3001:3000 \
    -v ./_build/fs/"$fs_name"/data:/data \
    -v ./_build/fs/"$fs_name"/cache:/cache \
    -v .:/src \
    -v ../agno/:/agno \
    -v ./_build/cache/"$fs_name"/go/mod:/go/pkg/mod \
    -v ./_build/cache/"$fs_name"/go/build:/go/cache \
    -v ./_build/cache/"$fs_name"/cargo/registry:/root/.cargo/registry \
    -v /src/weblens-vue/weblens-nuxt/node_modules \
    -e WEBLENS_MONGODB_URI=mongodb://"$container_name"-mongo:27017/?replicaSet=rs0 \
    -e WEBLENS_MONGODB_NAME="$container_name" \
    -e WEBLENS_INIT_ROLE="$tower_role" \
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
    ethrous/weblens:"$image_name-$arch" \
    "$run_cmd" | sed $'s/^/\e[34m[] \e[0m/'
