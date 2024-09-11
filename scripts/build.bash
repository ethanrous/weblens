if [[ ! -e ./cmd ]]; then
  echo "ERR Could not find ./cmd directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
  exit 1
fi

if [[ ! -e ./build/bin ]]; then
  mkdir -p ./build/bin
fi

go build -race -o ./build/bin/weblens ./cmd/weblens/main.go