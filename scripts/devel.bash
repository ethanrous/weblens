#!/bin/bash -eux

start_ui() {
    pushd ./weblens-vue/weblens-nuxt || exit 1

    pnpm dev &>/dev/null &

    popd || exit 1
}

build_ui() {
    pushd ./weblens-vue/weblens-nuxt || exit 1

    pnpm install
    if [[ $1 == true || ! -e ./.output/public/index.html ]]; then
        echo "Rebuilding UI..."
        pnpm generate
    fi

    popd || exit 1
}

build_agno() {
    # pushd /agno/ || exit 1 >/dev/null

    cargo build --release --target aarch64-unknown-linux-musl

    cp target/aarch64-unknown-linux-musl/release/libagno.a services/media/agno/lib/libagno.a

    popd || exit 1 >/dev/null
}

devel_weblens_locally() {
    echo "Running Weblens locally for development..."

    build_agno

    build_ui false
    start_ui

    air
}

debug_weblens() {
    debug_bin="./_build/bin/weblens_debug"

    build_agno

    build_ui true

    export CGO_CFLAGS='-g -O0'
    export CGO_CXXFLAGS='-g -O0'
    export CGO_LDFLAGS='-g'
    export CGO_ENABLED=1
    export GOARCH=arm64
    export GOOS=linux
    export CC=aarch64-linux-musl-gcc
    export CXX=aarch64-linux-musl-g++

    rm -f $debug_bin

    go build -v -gcflags=all="-N -l" -ldflags=-compressdwarf=false -o $debug_bin ./cmd/weblens/main.go

    $debug_bin

    # mkdir -p /root/.config/gdb/
    # echo "set auto-load safe-path /usr/local/go/src/runtime/runtime-gdb.py" >/root/.config/gdb/gdbinit
    # gdb $debug_bin
}

usage() {
    echo "Usage: $0 [-r|--dynamic]"
    echo "  -r, --dynamic   Enable dynamic mode"
}

dynamic=false
while [ "${1:-}" != "" ]; do
    case "$1" in
    "-r" | "--dynamic")
        dynamic=true
        ;;
    *)
        "Unknown argument: $1"
        usage
        exit 1
        ;;
    esac
    shift
done

if [[ $dynamic == true ]]; then
    echo "Dynamic mode enabled"
    devel_weblens_locally
else
    debug_weblens
fi
