#!/bin/bash
set -euo pipefail

publish_api_patch() {
    pushd ./api/ts/
    npm version patch
    npm publish

    new_version=$(package.json | jq -r .version)
    popd

    pushd ./weblens-vue/weblens-nuxt/
    pnpm install @ethanrous/weblens-api@"$new_version"
    popd
}

if [[ ! -e ./scripts ]]; then
    echo "ERR Could not find ./scripts directory, are you at the root of the repo? i.e. ~/repos/weblens and not ~/repos/weblens/scripts"
    exit 1
fi

printf "Generating swagger docs..."
if ! swag init --pd -g router.go -d './routers/router,./routers/api/v1' -q &>./_build/logs/swag.log; then
    echo "FAILED"
    cat ./_build/logs/swag.log
    echo "########## ^ Swag Init Logs ^ ##########"
    exit 1
fi
echo "DONE"
printf "########## END OF SWAG INIT ##########\n\n" >>./_build/logs/swag.log

printf "Generating typescript api..."
# export TS_POST_PROCESS_FILE="prettier --write"
if ! openapi-generator generate -i docs/swagger.json -g typescript-axios -o ./api/ts/generated --additional-properties=useESModules=true &>./_build/logs/swag-typescript.log; then
    echo "FAILED"
    cat ./_build/logs/swag-typescript.log
    echo "########## ^ Openapi Generator Logs ^ ##########"

    echo "openapi-generator (typescript) failed"
    exit 1
fi
echo "DONE"

printf "Compiling typescript api..."
pushd ./api/ts
npm run build
popd
echo "DONE"

printf "Generating go api..."
rm ./api/*.go
if ! openapi-generator generate -i docs/swagger.json -g go --git-user-id ethanrous --git-repo-id weblens/api -o ./api/ &>./_build/logs/swag-go.log; then
    echo "FAILED"
    cat ./_build/logs/swag-go.log
    echo "########## ^ Openapi Generator Logs ^ ##########"

    echo "openapi-generator (go) failed"
    exit 1
fi
echo "DONE"

publish=false
while [ "${1:-}" != "" ]; do
    case "$1" in
    "-p" | "--publish")
        publish=true
        ;;
    *)
        "Unknown argument: $1"
        usage
        exit 1
        ;;
    esac
    shift
done

if [[ $publish == true ]]; then
    printf "Publishing typescript api..."
    publish_api_patch
    echo "DONE"
fi
