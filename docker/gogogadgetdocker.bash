#!/bin/bash
set -e

while getopts ":t:a:" opt; do
  case $opt in
    t) docker_tag="$OPTARG"
    ;;
    a) arch="$OPTARG"
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
    docker_tag=devel_$(date +%b.%d.%y)
    echo "WARN No tag specified. Using: $docker_tag"
fi

if [ -z "$arch" ]
then
    arch="amd64"
fi
echo "Building for $arch"

if [ -z "$(sudo docker images -q weblens-go-build-"${arch}" 2> /dev/null)" ]; then
    echo "No weblens-go-build image found, attempting to build now..."
    sudo docker build -t weblens-go-build-"${arch}" --build-arg ARCHITECTURE="$arch" -f ./docker/GoBuild.Dockerfile .
fi

cd ./ui
export VITE_APP_BUILD_TAG=$docker_tag-$arch
export VITE_BUILD=true
npm run build

cd ..

if [ ! -d ./build/bin ]; then
  echo "Creating new build directory"
  mkdir -p ./build/bin
fi

sudo docker run -v ./:/source -v ./build/cache/go-pkg:/go -v ./build/cache/go-build:/root/.cache/go-build --platform "linux/$arch" --rm weblens-go-build-"${arch}" /bin/bash -c \
"cd /source && export GIN_MODE=release && CGO_ENABLED=1 GOOS=linux GOARCH=$arch go build -v -ldflags=\"-s -w\" -o ./build/bin/weblensbin ./cmd/weblens/main.go"

sudo docker build --platform "linux/$arch" -t ethrous/weblens:"${docker_tag}-${arch}" --build-arg build_tag="$docker_tag" -f ./docker/Dockerfile .

sudo docker push ethrous/weblens:"${docker_tag}-${arch}"

printf "\nBUILD COMPLETE. Container tag: $docker_tag-$arch"
