#!/bin/bash
set -e
docker_tag="$1"
if [ -z "$docker_tag" ]
then
    echo "WARN No tag specified. Using tag:"
    docker_tag=devel_$(date +%b.%d.%y)
    echo $docker_tag
fi

docker build --platform linux/amd64 -t ethrous/weblens:"$docker_tag" --build-arg build_tag="$docker_tag" --push .
