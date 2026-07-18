#!/bin/bash
set -euo pipefail

source ./scripts/lib/all.bash

usage="./scripts/quickCore.bash [-r|--rebuild] [-t|--role <role>] [-c|--clean]
	-r, --rebuild   Rebuild the container
	-t, --role      Specify the tower role (default: core)
	-c, --clean     Wipe the mongo container and file data
	-g, --group     Specify the group name to identify the stack (default: quick) e.g. will create a container named 'weblens-quick-core' and a mongo container named 'weblens-quick-core-mongo', assuming the tower role is 'core'
	--local-agno    Path to a local agno checkout. Builds libagno.a from source and uses its Go bindings (rebuilds every run). Also honored via the WEBLENS_LOCAL_AGNO env var
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

start_static_weblens() {
    if [[ "$skip_build" != true ]]; then
        build_frontend false
        build_weblens_binary
    fi

    exec "$WEBLENS_ROOT"/_build/bin/weblens_debug
}

leaks_weblens() {
    if [[ "$skip_build" != true ]]; then
        build_frontend false
        build_weblens_binary
    fi

    local graph_dir="$WEBLENS_ROOT/_build/leaks"
    mkdir -p "$graph_dir"

    # lite mode records allocation stacks in memory so snapshots include them
    MallocStackLogging=lite "$WEBLENS_ROOT"/_build/bin/weblens_debug &

    local server_pid=$!
    trap 'printf "\nWaiting for Weblens to exit...\n"; wait "$server_pid" || true; exit 0' INT TERM

    sleep 2
    printf "\nWeblens running under MallocStackLogging (pid %s)\n" "$server_pid"
    printf "Press \e[34mENTER\e[0m to capture a memgraph snapshot, \e[34mCtrl-C\e[0m to stop\n"

    local graph_file
    while read -r; do
        if ! kill -0 "$server_pid" 2>/dev/null; then
            printf "Weblens has exited\n"
            break
        fi

        graph_file="$graph_dir/weblens-$(date +%Y%m%d-%H%M%S).memgraph"
        leaks "$server_pid" --outputGraph="$graph_file" >/dev/null || true
        printf "Saved \e[34m%s\e[0m\n" "$graph_file"
    done

    wait "$server_pid" || true
}

tower_role="core"
sub_stack_group=""

static=false
valgrind=false
skip_build=false

mongo_port=27017
weblens_port=8080

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-t" | "--role")
        shift
        tower_role="$1"
        ;;
    "--sub-stack-group")
        shift
        sub_stack_group="$1"
        ;;
    "--weblens-port")
        shift
        weblens_port="$1"
        ;;
    "--mongo-port")
        shift
        mongo_port="$1"
        ;;
    "--static")
        static=true
        ;;
    "--valgrind")
        valgrind=true
        ;;
    "--trace")
        export WEBLENS_LOG_LEVEL=trace
        ;;
    "--skip-build")
        skip_build=true
        ;;
    "--local-agno")
        shift
        export WEBLENS_LOCAL_AGNO="$1"
        ;;
    *)
        "Unknown argument: $1"
        echo "$usage"
        exit 1
        ;;
    esac
    shift
done

if [[ "$tower_role" == "" ]]; then
    echo "No tower role specified, defaulting to 'core'"
    tower_role="core"
elif [[ "$tower_role" == "backup" ]]; then
    mongo_port=27018
    weblens_port=8081
fi

stack_name="weblens-$tower_role"

if [[ "$sub_stack_group" != "" ]]; then
    stack_name="$stack_name-$sub_stack_group"
fi

echo "Using stack name: $stack_name"

setup_agno

export WEBLENS_LOG_LEVEL="${WEBLENS_LOG_LEVEL:-debug}"

file_path="$WEBLENS_ROOT/_build/fs/$tower_role"
if [[ "$sub_stack_group" != "" ]]; then
    file_path="$file_path-$sub_stack_group"
fi

export WEBLENS_MONGODB_URI="mongodb://127.0.0.1:$mongo_port/?directConnection=true"
export WEBLENS_EMBED_URI="http://127.0.0.1:5500"
export WEBLENS_DATA_PATH="$file_path/data"
export WEBLENS_CACHE_PATH="$file_path/cache"
export WEBLENS_UI_PATH=./weblens-vue/weblens-nuxt/.output/public
export WEBLENS_STATIC_CONTENT_PATH="$WEBLENS_ROOT/public"
export WEBLENS_LOG_FORMAT="${WEBLENS_LOG_FORMAT:-dev}"
export WEBLENS_DO_CACHE="${WEBLENS_DO_CACHE:-true}"
export WEBLENS_DO_PROFILING="${WEBLENS_DO_PROFILING:-true}"
export WEBLENS_PORT=$weblens_port
export WEBLENS_JWT_SECRET="DEV"

export AGNO_LOG_LEVEL=warn
export AGNO_LOG_FORMAT=human

export GO_BUILD_TAGS="pdfium"

show_as_subtask "Starting services" "blue" -- ./scripts/envup.bash

printf "Running \e[34mWeblens\e[0m locally for development...\n"

if [[ "$valgrind" == true ]]; then
    leaks_weblens
elif [[ "$static" == true ]]; then
    start_static_weblens
else
    devel_weblens_locally
fi
