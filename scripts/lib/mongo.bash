#!/bin/bash

is_mongo_running() {
    local stack_name=""
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--stack-name")
            shift
            stack_name="$1"
            ;;
        *)
            echo "Unknown argument: $1"
            return 1
            ;;
        esac
        shift
    done

    if [[ -z "$stack_name" ]]; then
        echo "is_mongo_running requires a stack_name argument. Aborting."
        return 1
    fi

    if dockerc ps | grep "$stack_name" &>/dev/null; then
        return 0
    fi

    return 1
}

export -f is_mongo_running

launch_mongo() {
    local stack_name=""
    local mongo_port=27017

    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--stack-name")
            shift
            stack_name="$1"
            ;;
        "-p" | "--mongo-port")
            shift
            mongo_port="$1"
            ;;
        *)
            echo "Unknown argument: $1"
            return 1
            ;;
        esac
        shift
    done

    if [[ -z "$stack_name" ]]; then
        echo "launch_mongo requires a stack_name argument. Aborting."
        return 1
    fi

    ensure_weblens_net

    export MONGO_DATA_ROOT="${WEBLENS_ROOT}/_build/db/$stack_name/"
    export MONGO_HOST_PORT=$mongo_port

    echo "Starting MongoDB container [$stack_name] on port [:$mongo_port] ..."

    mkdir -p "${MONGO_DATA_ROOT}/mongod" "${MONGO_DATA_ROOT}/mongot"
    chmod 777 "${MONGO_DATA_ROOT}/mongod" "${MONGO_DATA_ROOT}/mongot"

    export MONGO_PROJECT_NAME="$stack_name"
    if ! dockerc compose -f ./docker/mongo.compose.yaml --project-name "$stack_name" up -d; then
        echo "!!! docker compose up failed !!!" >&2
        echo "--- mongod container logs ---" >&2
        dockerc logs "weblens-$stack_name-mongod" --tail 150 >&2 2>&1 || true
        echo "--- mongod container inspect ---" >&2
        dockerc inspect "weblens-$stack_name-mongod" --format '{{json .State}}' >&2 2>&1 || true
        echo "--- All containers ---" >&2
        dockerc ps -a >&2 2>&1 || true
        exit 1
    fi
}

export -f launch_mongo

# Stop all mongo containers and remove mongo volume, if specified
cleanup_mongo() {
    local stack_name=""

    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--stack-name")
            shift
            stack_name="$1"
            ;;
        *)
            echo "Unknown argument: $1"
            return 1
            ;;
        esac
        shift
    done

    if [[ -z "$stack_name" ]]; then
        echo "cleanup_mongo requires a stack_name argument. Aborting."
        return 1
    fi

    dockerc compose --project-name "$stack_name" down
}

export -f cleanup_mongo
