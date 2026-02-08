#!/bin/bash

is_mongo_running() {
    local mongo_name=${1?"[ERROR] is_mongo_running called with no container name. Aborting"}

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
    local mongo_name="${1:-weblens-core-mongo}"
    local mongo_port="${2:-27017}"

    ensure_weblens_net

    if ! dockerc ps | grep "$mongo_name" &>/dev/null; then
        echo "Starting MongoDB container [$mongo_name] on port [:$mongo_port] ..."

        echo "--- Docker debug info ---"
        echo "Docker context: $(docker context show 2>&1 || echo 'N/A')"
        echo "Docker info (brief):"
        dockerc info --format '{{.ServerVersion}} | {{.OSType}}/{{.Architecture}} | containers={{.Containers}}' 2>&1 || true
        echo "Docker networks:"
        dockerc network ls 2>&1 || true
        echo "--- End Docker debug info ---"

        if ! dockerc compose -f ./docker/mongo.compose.yaml --env-file ./docker/mongo-core.env --project-name "$mongo_name" up -d; then
            echo "!!! docker compose up failed !!!"
            echo "--- Container status ---"
            dockerc ps -a --filter "label=com.docker.compose.project=$mongo_name" 2>&1 || true
            echo "--- mongod container logs ---"
            dockerc logs weblens-core-mongod 2>&1 || true
            echo "--- mongod container inspect ---"
            dockerc inspect weblens-core-mongod --format '{{json .State}}' 2>&1 || true
            echo "--- All containers ---"
            dockerc ps -a 2>&1 || true
            exit 1
        fi

        # Verify containers are actually running after compose returns
        sleep 2
        echo "--- Post-launch container status ---"
        dockerc ps -a --filter "label=com.docker.compose.project=$mongo_name" 2>&1 || true

        if ! dockerc ps | grep "weblens-core-mongod" &>/dev/null; then
            echo "!!! mongod container is not running after compose up !!!"
            echo "--- mongod container logs ---"
            dockerc logs weblens-core-mongod 2>&1 || true
            echo "--- mongod container inspect ---"
            dockerc inspect weblens-core-mongod --format '{{json .State}}' 2>&1 || true
            exit 1
        fi
    fi
}

export -f launch_mongo

# Stop all mongo containers and remove mongo volume, if specified
cleanup_mongo() {
    local mongo_name="${1:-weblens-core-mongo}"
    local mongo_port="${2:-27017}"

    dockerc compose --project-name "$mongo_name" down
}

export -f cleanup_mongo
