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

    build_frontend true
    start_ui

    air
}

tower_role="core"
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

stack_name="weblens-$tower_role"

echo "Using stack name: $stack_name"

if ! does_agno_exist; then
    build_agno
fi

show_as_subtask "Launching mongo..." "green" -- launch_mongo --stack-name "$stack_name" --mongo-port "$mongo_port"

# if [[ "$tower_role" == "core" ]] && ! is_hdir_running; then
#     launch_hdir | show_as_subtask "Launching HDIR..." "green"
# fi
#
WEBLENS_LOG_LEVEL="${WEBLENS_LOG_LEVEL:-debug}"

export WEBLENS_MONGODB_URI="mongodb://127.0.0.1:$mongo_port/?replicaSet=rs0&directConnection=true"
export WEBLENS_HDIR_URI="http://127.0.0.1:5001"
export WEBLENS_DATA_PATH="./_build/fs/$tower_role/data"
export WEBLENS_CACHE_PATH="./_build/fs/$tower_role/cache"
export WEBLENS_LOG_FORMAT=dev
export WEBLENS_DO_CACHE=true
export WEBLENS_DO_PROFILING=true
export WEBLENS_PORT=$weblens_port
export AGNO_LOG_LEVEL=warn
export AGNO_LOG_FORMAT=human

printf "Running \e[34mWeblens\e[0m locally for development...\n"

devel_weblens_locally
