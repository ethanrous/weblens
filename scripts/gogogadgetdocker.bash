#!/bin/bash

if [[ ! -e ./scripts ]]; then
	echo "ERR Could not find ./scripts directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
	exit 1
fi

mkdir -p ./build/bin
mkdir -p ./build/logs

# Once the container is build, push it to docker hub
push=false

# Skip testing
skip=false

while getopts ":t:a:ps" opt; do
	case $opt in
	t)
		docker_tag="$OPTARG"
		;;
	a)
		arch="$OPTARG"
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
	if ! ./scripts/testWeblens -a &>./build/logs/container-build-pretest.log; then
		printf " FAILED\n"
		cat ./build/logs/container-build-pretest.log
		echo "Aborting container build. Ensure ./scripts/testWeblens passes before building container"
		exit 1
	else
		printf " PASS\n"
	fi
fi

if [[ ! -e ./build/ffmpeg ]]; then
	docker run --platform linux/amd64 -v ./scripts/buildFfmpeg.sh:/buildFfmpeg.sh -v ./build:/build --rm alpine /buildFfmpeg.sh
fi

printf "Building Weblens container..."
sudo docker rmi ethrous/weblens:"${docker_tag}-${arch}" &>/dev/null
sudo docker build --platform "linux/$arch" -t ethrous/weblens:"${docker_tag}-${arch}" --build-arg build_tag="$docker_tag" --build-arg ARCHITECTURE=amd64 -f ./docker/Dockerfile .

if [ $push == true ]; then
	sudo docker push ethrous/weblens:"${docker_tag}-${arch}"
fi

printf "\nBUILD COMPLETE. Container tag: ethrous/weblens:%s-%s\n" "$docker_tag" "$arch"
