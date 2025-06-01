#!/bin/bash
set -e
set -o pipefail

devel_weblens_locally() {
    echo "Running WebLens locally for development..."

    cd ./ui
    pnpm install
    pnpm build
    pnpm dev:no-open &

    cd ..

    export WEBLENS_STATIC_CONTENT_PATH=./public

    air

    exit 0
}

ensure_repl_set() {
    if ! docker exec -t "$mongoName" mongosh --eval "rs.status()" &>/dev/null; then
        echo "MongoDB replica set is not initialized, initializing..."
        docker exec -t "$mongoName" mongosh --eval "rs.initiate({_id: 'rs0', members: [ { _id: 0, host: '\"$mongoName\":27017' } ]})"
        return $?
    fi
}

launch_mongo() {
    mountPath=$(docker inspect "$mongoName" | jq -r '.[].Mounts[] | select(.Type=="bind") | .Source')
    if [[ "$mountPath" != "$PWD/build/fs/$fsName/db" ]]; then
        echo "MongoDB mount path does not match, removing old container..."
        echo "Should be: $mountPath -- but found: $PWD/build/fs/$fsName/db"
        docker stop "$mongoName" 2>/dev/null || :
    fi

    mongoWaitCount=0
    if ! docker ps | grep "$mongoName"; then
        echo "Starting MongoDB container [$mongoName] ..."
        docker run \
            --rm \
            -d \
            --name "$mongoName" \
            -v ./build/fs/"$fsName"/db:/data/db \
            --network weblens-net \
            mongo \
            mongod --replSet rs0

        while [[ $mongoWaitCount -lt 10 ]]; do
            if ! ensure_repl_set; then
                echo "Waiting for MongoDB to be ready... $?"
                mongoWaitCount=$((mongoWaitCount + 1))
                sleep 1
            else
                break
            fi
        done

        if [[ $mongoWaitCount -ge 10 ]]; then
            echo "MongoDB did not start in time, exiting..."
            exit 1
        fi
    fi
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

docker stop "$containerName" 2>/dev/null || :
docker rm "$containerName" 2>/dev/null || :

if ! docker network ls | grep weblens-net &>/dev/null; then
    echo "Creating weblens docker network..."
    docker network create weblens-net
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

launch_mongo

docker run \
    -t \
    -i \
    --rm \
    --name "$containerName" \
    -p 8080:8080 \
    -p 3000:3000 \
    -v ./build/fs/"$fsName"/data:/data \
    -v ./build/fs/"$fsName"/cache:/cache \
    -v .:/src \
    -v ./build/cache/"$fsName"/go:/go/pkg/mod \
    -e WEBLENS_MONGODB_URI=mongodb://"$containerName"-mongo:27017/?replicaSet=rs0 \
    -e WEBLENS_MONGODB_NAME="$containerName" \
    -e WEBLENS_LOG_LEVEL=trace \
    -e WEBLENS_LOG_FORMAT=dev \
    --network weblens-net \
    ethrous/weblens:"$imageName-$arch"
