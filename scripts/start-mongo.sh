#!/bin/bash

mongoName=$1
if [[ -z "$mongoName" ]]; then
    echo "Usage: $0 <mongo-container-name>"
    exit 1
fi

ensure_repl_set() {
    mongoWaitCount=0
    while [[ $mongoWaitCount -lt 10 ]]; do
        status="$(docker inspect "$mongoName" --format '{{.State.Health.Status}}')"
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
    if ! docker image ls | grep ethrous/weblens-mongo; then
        ./scripts/build-mongo.bash || exit 1
    fi

    echo "MONGO VVV"
    docker image ls

    if ! docker ps | grep "$mongoName"; then
        echo "Starting MongoDB container [$mongoName] ..."
        docker run \
            --rm \
            -d \
            --name "$mongoName" \
            --mount type=volume,src="$mongoName",dst=/data/db \
            --network weblens-net \
            -e WEBLENS_MONGO_HOST_NAME="$mongoName" \
            ethrous/weblens-mongo || exit 1
    fi

    ensure_repl_set
}

launch_mongo
