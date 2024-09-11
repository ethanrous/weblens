export MEDIA_ROOT=$(pwd)/build/fs/test/users
export CACHES_PATH=$(pwd)/build/fs/test/caches

if [[ ! -e ./ui ]]; then
  echo "ERR Could not find ./ui directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
  exit 1
fi

if [[ ! -e ./ui/dist/index.html ]]; then
  cd ui
  npm install
  npm run build
  cd ..
fi

CONFIG_PATH=$(pwd)/config CONFIG_NAME=TEST go test -v ./...
