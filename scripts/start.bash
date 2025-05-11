usage="./scripts/quickCore.bash [-r|--rebuild] [-t|--role <role>] [-c|--clean]
	-r, --rebuild   Rebuild the container
	-t, --role      Specify the tower role (default: core)
	-c, --clean     Wipe the mongo container and file data"

towerRole="core"

arch=$(uname -m)

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-r" | "--rebuild")
        docker image rm -f ethrous/weblens:quick-"$towerRole"-"$arch" &>/dev/null
        ;;
    "-t" | "--role")
        shift
        towerRole="$1"
        ;;
    "-c" | "--clean")
        docker stop weblens-quick-"$towerRole"-mongo
        rm -rf ./build/fs/core-container
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

docker stop weblens-quick-"$towerRole" 2>/dev/null
docker rm weblens-quick-"$towerRole" 2>/dev/null

if ! docker network ls | grep weblens-net &>/dev/null; then
    echo "Creating weblens docker network..."
    docker network create weblens-net
fi

if ! docker image ls | grep "quick-$towerRole-$arch" &>/dev/null; then
    echo "Container does not exist, building..."
    if ! ./scripts/gogogadgetdocker.bash -t quick-"$towerRole" -a "$arch"; then
        echo "Failed to build container"
        exit 1
    fi
fi

if ! docker ps | grep weblens-quick-"$towerRole"-mongo; then
    docker run \
        --rm \
        -d \
        --name weblens-quick-"$towerRole"-mongo \
        -v ./build/fs/core-container/db:/data/db \
        --network weblens-net \
        mongo \
        mongod --replSet rs0

    while :; do
        if ! docker exec -it weblens-quick-"$towerRole"-mongo mongosh --eval "rs.initiate({
			_id: 'rs0',
			members: [
			{ _id: 0, host: 'weblens-quick-\"$towerRole\"-mongo' }
			]
		})"; then
            echo "Waiting for MongoDB to be ready..."
            sleep 1
        else
            break
        fi
    done
fi

docker run \
    --rm \
    --name weblens-quick-"$towerRole" \
    -p 8089:8080 \
    -v ./build/fs/core-container/data:/data \
    -v ./build/fs/core-container/cache:/cache \
    -e WEBLENS_MONGODB_URI=mongodb://weblens-quick-"$towerRole"-mongo:27017/?replicaSet=rs0 \
    -e WEBLENS_MONGODB_NAME=weblens-quick-"$towerRole" \
    -e WEBLENS_LOG_LEVEL=debug \
    -e LOG_FORMAT=dev \
    --network weblens-net \
    ethrous/weblens:quick-"$towerRole"-"$arch"
