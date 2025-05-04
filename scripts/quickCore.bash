usage="TODO"

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-r" | "--rebuild")
        docker image rm -f ethrous/weblens:quick-core-arm64
        ;;
    "-c" | "--clean")
        docker stop weblens-quick-core-mongo
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

docker stop weblens-quick-core
docker rm weblens-quick-core

arch=$(uname -m)

if ! docker image ls | grep "quick-core-$arch"; then
    echo "Container does not exist, building..."
    if ! ./scripts/gogogadgetdocker.bash -t "quick-core" -a "$arch"; then
        echo "Failed to build container"
        exit 1
    fi
fi

if ! docker ps | grep weblens-quick-core-mongo; then
    docker run \
        --rm \
        -d \
        --name weblens-quick-core-mongo \
        -v ./build/fs/core-container/db:/data/db \
        --network weblens-net \
        mongo \
        mongod --replSet rs0

    while :; do
        if ! docker exec -it weblens-quick-core-mongo mongosh --eval "rs.initiate({
			_id: 'rs0',
			members: [
			{ _id: 0, host: 'weblens-quick-core-mongo' }
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
    --name weblens-quick-core \
    -p 8089:8080 \
    -v ./build/fs/core-container/data:/data \
    -v ./build/fs/core-container/cache:/cache \
    -e WEBLENS_MONGODB_URI=mongodb://weblens-quick-core-mongo:27017/?replicaSet=rs0 \
    -e WEBLENS_MONGODB_NAME=weblens-quick-core \
    -e WEBLENS_LOG_LEVEL=debug \
    -e LOG_FORMAT=dev \
    --network weblens-net \
    ethrous/weblens:quick-core-"$arch"
