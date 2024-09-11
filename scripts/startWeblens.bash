if [[ ! -e ./cmd ]]; then
  echo "ERR Could not find ./cmd directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
  exit 1
fi

if [[ ! -e ./build/bin ]]; then
  mkdir -p ./build/bin
fi

rm -f ./build/bin/weblens

go build -race -o ./build/bin/weblens ./cmd/weblens/main.go

if [[ ! -e ./build/bin/weblens ]]; then
  echo "Failed to build Weblens, exiting..."
  exit 1
fi

if [[ -z "$CONFIG_NAME" ]]; then
  export CONFIG_NAME=DEBUG-CORE
fi

CONFIG_PATH=$(pwd)/config DETACH_UI=true ./build/bin/weblens