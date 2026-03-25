#!/bin/bash

is_mongo_running() {
    local stack_name=""
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--stack-name")
            shift
            stack_name="$1-mongo"
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

    if docker compose ls --filter "name=weblens-$stack_name" --format json 2>/dev/null | grep '"running(1)"' &>/dev/null; then
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
            stack_name="$1-mongo"

            # Check if stack_name starts with "weblens-" and if it does, strip it
            if [[ "$stack_name" == weblens-* ]]; then
                stack_name="${stack_name#weblens-}"
            fi
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

    mkdir -p "${MONGO_DATA_ROOT}/mongod" "${MONGO_DATA_ROOT}/configdb" "${MONGO_DATA_ROOT}/mongot"
    chmod 777 "${MONGO_DATA_ROOT}/mongod" "${MONGO_DATA_ROOT}/configdb" "${MONGO_DATA_ROOT}/mongot"

    # If the keyfile doesn't exist, create it with random content and set permissions to 400
    if [ ! -f "${MONGO_DATA_ROOT}keyfile" ]; then
        tr -dc 'a-zA-Z0-9' </dev/urandom | head -c 756 >"${MONGO_DATA_ROOT}keyfile"
        chmod 400 "${MONGO_DATA_ROOT}keyfile"
    fi

    export MONGO_PROJECT_NAME="$stack_name"
    if ! dockerc compose -f ./docker/mongo.compose.yaml --project-name "weblens-$stack_name" up -d; then
        log_dump_file="./_build/logs/mongo/failed-mongo-$stack_name.log"
        mkdir -p "$(dirname "$log_dump_file")"

        echo "dumping mongo container logs to [$log_dump_file]..."
        dockerc logs "weblens-$stack_name-mongo" >"$log_dump_file" || true
        return 1
    else
        # Wait for mongo to be healthy before returning
        local retries=30
        local wait_time=1
        local count=0
        until docker inspect --format='{{json .State.Health}}' weblens-"$stack_name"-mongo 2>/dev/null | grep -q '"healthy"'; do
            if [[ $count -ge $retries ]]; then
                log_dump_file="./_build/logs/mongo/failed-mongo-$stack_name.log"
                mkdir -p "$(dirname "$log_dump_file")"

                echo "MongoDB container failed to become healthy after $((retries * wait_time)) seconds. Check container logs at $log_dump_file for details" >&2

                dockerc logs "weblens-$stack_name-mongo" >"$log_dump_file" || true

                return 1
            fi
            sleep $wait_time
            ((count++)) || true
        done
    fi
}
export -f launch_mongo

dump_mongo_logs() {
    local stack_name=""
    local logfile=""
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--stack-name")
            shift
            stack_name="$1-mongo"
            ;;
        "--logfile")
            shift
            logfile="$1"
            ;;
        *)
            echo "Unknown argument: $1"
            return 1
            ;;
        esac
        shift
    done

    if [[ -z "$stack_name" ]]; then
        echo "dump_mongo_logs requires a stack_name argument. Aborting."
        return 1
    fi

    echo "Dumping MongoDB logs for stack [$stack_name] to [$logfile] ..."

    dockerc logs "weblens-$stack_name-mongo" >"$logfile-mongo.log" || true
}
export -f dump_mongo_logs

# Stop all mongo containers and remove mongo volume, if specified
cleanup_mongo() {
    local stack_name=""

    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--stack-name")
            shift
            stack_name="$1-mongo"
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

    dockerc compose --project-name "weblens-$stack_name" down
}

export -f cleanup_mongo
