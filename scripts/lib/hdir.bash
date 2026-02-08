#!/bin/bash

is_hdir_running() {
    if dockerc ps | grep weblens-hdir &>/dev/null; then
        return 0
    fi

    return 1
}
export -f is_hdir_running

launch_hdir() {
    if ! dockerc image ls | grep weblens_hdir &>/dev/null; then
        build_hdir
    fi

    dockerc run --rm -d --name weblens-hdir --publish 5001:5000 -v "${WEBLENS_ROOT}/_build/fs/core/cache/:/images" -v "${WEBLENS_ROOT}/_build/hdir/model-cache/:/root/.cache/huggingface" --network weblens-net ethrous/weblens_hdir
}
export -f launch_hdir
