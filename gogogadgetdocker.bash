#!/bin/bash
set -e
docker_tag="$1"
if [ -z "$docker_tag" ]
then
    echo "WARN No tag specified. Using tag:"
    docker_tag=devel_$(date +%b.%d.%y)
    echo $docker_tag
fi

export VITE_APP_BUILD_TAG=$docker_tag
if [ -z "$(docker images -q buildbuntu 2> /dev/null)" ]; then
    echo "No buildbuntu image found, attempting to build now..."
    docker build -t weblens-go-build -f GoBuild .
fi

cd ./ui
export VITE_BUILD=true
npm run build
cd ..

docker run -v ./api/src:/source --platform linux/amd64 --rm weblens-go-build /bin/bash -c \
'cd /source && export GIN_MODE=release && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w" -o weblens'

docker build --platform linux/amd64 -t ethrous/weblens:"$docker_tag" --build-arg build_tag="$docker_tag" .
docker push ethrous/weblens:"$docker_tag"

# docker build --platform linux/amd64 -t ethrous/weblens-recog:"$docker_tag" --build-arg build_tag="$docker_tag" ./classification
# docker push ethrous/weblens-recog:"$docker_tag"
