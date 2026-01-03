#!/bin/bash
set -euo pipefail

is_mongo_running() {
    local mongo_name=${1+x}
    if [[ -z "$mongo_name" ]]; then
        echo "[ERROR] is_mongo_running called with no container name. Aborting"
        exit 1
    fi

    if dockerc ps | grep "$mongo_name" &>/dev/null; then
        return 0
    fi

    return 1
}

export -f is_mongo_running

ensure_repl_set() {
    mongoWaitCount=0
    while [[ $mongoWaitCount -lt 10 ]]; do
        status="$(dockerc inspect "$mongo_name" --format '{{.State.Health.Status}}')"
        if [[ $status == "starting" ]]; then
            mongoWaitCount=$((mongoWaitCount + 1))
            echo "MongoDB is starting, waiting ${mongoWaitCount}s..."
            sleep $mongoWaitCount
            continue
        fi

        if [[ $status == "healthy" ]]; then
            return 0
        fi

        if [[ $status == "unhealthy" ]]; then
            echo "MongoDB container is unhealthy, exiting..."
            exit 1
        fi

        echo "Waiting for MongoDB to be ready... $status"
        mongoWaitCount=$((mongoWaitCount + 1))
        sleep 1
    done

    if [[ $mongoWaitCount -ge 10 ]]; then
        echo "MongoDB did not start in time, exiting..."
        exit 1
    fi

    return 0
}

launch_mongo() {
    local mongo_name="${1?[ERROR] launch_mongo called with no container name. Aborting}"

    if ! dockerc image ls | grep ethrous/weblens-mongo &>/dev/null; then
        ./scripts/build-mongo.bash || exit 1
    fi

    if ! dockerc ps | grep "$mongo_name" &>/dev/null; then
        echo "Starting MongoDB container [$mongo_name] ..."

        dockerc run \
            -d \
            --rm \
            --name "$mongo_name" \
            -v ./_build/log/syslog:/var/log/syslog \
            --mount type=volume,src="$mongo_name",dst=/data/db \
            --publish 27018:27017 \
            --network weblens-net \
            -e WEBLENS_MONGO_HOST_NAME="$mongo_name" \
            ethrous/weblens-mongo || exit 1
    fi

    ensure_repl_set
}

export -f launch_mongo

# Stop all mongo containers and remove mongo volume, if specified
cleanup_mongo() {
    local running_mongos
    running_mongos=$(docker ps | grep -e "weblens" -e "mongo") || true
    if [[ ! -z "$running_mongos" ]]; then
        while IFS= read -r container; do
            local container_id
            container_id=$(sed -E 's/^([^ ]+).*/\1/' <<<"$container")

            echo "Stopping mongo container [$container_id] ..."
            dockerc stop "$container_id"
        done <<<"$running_mongos"
    else
        echo "No running mongo containers found."
    fi

    if [[ -z "${1:-}" ]]; then
        return
    fi

    local mongo_name=$1
    dockerc volume rm "$mongo_name" 2>&1 || true
}

export -f cleanup_mongo
