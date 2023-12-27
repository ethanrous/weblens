#!/bin/bash
set -e
docker_tag="$1"
# docker build --platform linux/amd64 -t ethrous/weblens:v0.2 --progress plain --push .
if [ -z "$docker_tag" ]
then
    echo "ERR No tag specified"
    exit 1
fi

docker build --platform linux/amd64 -t ethrous/weblens:"$docker_tag" --push .
