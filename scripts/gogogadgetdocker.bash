#!/bin/bash
set -e

local=false

while getopts ":t:a:l" opt; do
  case $opt in
    t) docker_tag="$OPTARG"
    ;;
    a) arch="$OPTARG"
    ;;
    l) local=true
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    exit 1
    ;;
  esac

  case $OPTARG in
    -*) echo "Option $opt needs a valid argument"
    exit 1
    ;;
  esac
done

if [ -z "$docker_tag" ]
then
    docker_tag=devel_$(git rev-parse --abbrev-ref HEAD)
    echo "WARN No tag specified"
fi

if [ -z "$arch" ]
then
    arch="amd64"
fi

echo "Using tag: $docker_tag-$arch"

if [ $local == false ] && [ -z "$(sudo docker images -q weblens-go-build-"${arch}" 2> /dev/null)" ]; then
    echo "No weblens-go-build image found, attempting to build now..."
    sudo docker build -t weblens-go-build-"${arch}" --build-arg ARCHITECTURE="$arch" -f ./docker/GoBuild.Dockerfile .
fi

cd ./ui
npm install
export VITE_APP_BUILD_TAG=$docker_tag-$arch
export VITE_BUILD=true
npm run build

cd ..

if [ ! -d ./build/bin ]; then
  echo "Creating new build directory"
  mkdir -p ./build/bin
fi

if [ $local == true ]; then
  GIN_MODE=release CGO_ENABLED=1 GOOS=linux GOARCH=$arch go build -v -ldflags="-s -w" -o ./build/bin/weblensbin ./cmd/weblens/main.go
else
  sudo docker run -v ./:/source -v ./build/.cache/go-pkg:/go -v ./build/.cache/go-build:/root/.cache/go-build --platform "linux/$arch" --rm weblens-go-build-"${arch}" /bin/bash -c \
  "cd /source && GIN_MODE=release CGO_ENABLED=1 GOOS=linux GOARCH=$arch go build -v -ldflags=\"-s -w\" -o ./build/bin/weblensbin ./cmd/weblens/main.go"
fi

sudo docker build --platform "linux/$arch" -t ethrous/weblens:"${docker_tag}-${arch}" --build-arg build_tag="$docker_tag" -f ./docker/Dockerfile .

sudo docker push ethrous/weblens:"${docker_tag}-${arch}"

printf "\nBUILD COMPLETE. Container tag: $docker_tag-$arch\n"
