#!/bin/bash

if [[ ! -e ./scripts ]]; then
  echo "ERR Could not find ./scripts directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
  exit 1
fi

mkdir -p ./build/bin
mkdir -p ./build/logs

# Build go binary locally, on host OS, rather than in a container
local=false

# Once the container is build, push it to docker hub
push=false

# Skip testing
skip=false

while getopts ":t:a:lps" opt; do
  case $opt in
  t)
    docker_tag="$OPTARG"
    ;;
  a)
    arch="$OPTARG"
    ;;
  l)
    local=true
    ;;
  p)
    push=true
    ;;
  s)
    skip=true
    ;;
  \?)
    echo "Invalid option -$OPTARG" >&2
    exit 1
    ;;
  esac

  case $OPTARG in
  -*)
    echo "Option $opt needs a valid argument"
    exit 1
    ;;
  esac
done

sudo docker ps &>/dev/null
docker_status=$?

printf "Checking connection to docker..."
if [ $docker_status != 0 ]; then
  printf " FAILED\n"
  echo "Aborting container build. Ensure docker is runnning"
  exit 1
else
  printf " PASS\n"
fi

if [ -z "$docker_tag" ]; then
  docker_tag=devel_$(git rev-parse --abbrev-ref HEAD)
  echo "WARN No tag specified"
fi

if [ -z "$arch" ]; then
  arch="amd64"
fi

echo "Using tag: $docker_tag-$arch"

if [ ! $skip == true ]; then
  printf "Running tests..."
  if ! ./scripts/testWeblens --ui --backend &>./build/logs/container-build-pretest.log; then
    printf " FAILED\n"
    cat ./build/logs/container-build-pretest.log
    echo "Aborting container build. Ensure ./scripts/testWeblens passes before building container"
    exit 1
  else
    printf " PASS\n"
  fi
fi

if [ $local == false ] && [ -z "$(sudo docker images -q weblens-go-build-"${arch}" 2>/dev/null)" ]; then
  echo "No weblens-go-build image found, attempting to build now..."
  if ! sudo docker build -t weblens-go-build-"${arch}" --build-arg ARCHITECTURE="$arch" -f ./docker/GoBuild.Dockerfile .; then
    echo "Failed to build weblens-go-build image"
    exit 1
  fi
fi

cd ./ui || exit
printf "Building UI..."
npm install &>/dev/null
export VITE_APP_BUILD_TAG=$docker_tag-$arch
export VITE_BUILD=true
if ! npm run build &>../build/logs/ui-build.log; then
  printf " FAILED\n"
  echo "Aborting container build. Ensure npm run build completes successfully before building container"
  exit 1
else
  printf " DONE\n"
fi

cd ..

rm -f ./build/bin/weblensbin

printf "Building Weblens binary..."
if [ $local == true ]; then
  GIN_MODE=release CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=$arch go build -v -ldflags="-s -w" -o ./build/bin/weblensbin ./cmd/weblens/main.go &>./build/logs/weblens-build.log
else
  #shellcheck disable=SC2024
  if ! sudo docker run -v ./:/source -v ./build/.cache/go-pkg:/go -v ./build/.cache/go-build:/root/.cache/go-build --platform "linux/$arch" --rm weblens-go-build-"${arch}" /bin/bash -c \
    "cd /source && CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=$arch go build -v -ldflags=\"-s -w\" -o ./build/bin/weblensbin ./cmd/weblens/main.go" &>./build/logs/weblens-build.log; then
    printf " FAILED\n"
    cat ./build/logs/weblens-build.log
    echo "Aborting container build. Ensure go build completes successfully before building container"
    exit 1
  fi
fi

if [[ ! -e ./build/bin/weblensbin ]]; then
  printf " FAILED\n"
  cat ./build/logs/weblens-build.log
  echo "Aborting container build. Could not find ./build/bin/weblensbin"
  exit 1
fi
printf " DONE\n"

sudo docker build --platform "linux/$arch" -t ethrous/weblens:"${docker_tag}-${arch}" --build-arg build_tag="$docker_tag" -f ./docker/Dockerfile .

if [ $push == true ]; then
  sudo docker push ethrous/weblens:"${docker_tag}-${arch}"
fi

printf "\nBUILD COMPLETE. Container tag: ethrous/weblens:%s-%s\n" "$docker_tag" "$arch"
