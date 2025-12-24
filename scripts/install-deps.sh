set -euox pipefail

buildDeps=false
devDeps=false
agno=false

usage() {
    echo "Usage: $0 [-b|--build] [-d|--dev]"
}

while [ "${1:-}" != "" ]; do
    case "$1" in
    "-b" | "--build")
        buildDeps=true
        ;;
    "-d" | "--dev")
        buildDeps=true
        devDeps=true
        ;;
    "-a" | "--agno")
        agno=true
        ;;
    "-h" | "--help")
        usage
        exit 0
        ;;
    *)
        echo "Unknown argument: $1"
        usage
        exit 1
        ;;
    esac
    shift
done

apk upgrade --no-cache
apk add --no-cache ffmpeg

if [[ $buildDeps == true ]]; then
    apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community

    apk add --no-cache \
        bash \
        build-base \
        curl \
        gcc \
        lcms2-dev \
        libgcc \
        g++ \
        libstdc++ \
        libpng-dev \
        libraw-dev \
        libwebp-dev \
        musl-dbg \
        musl \
        musl-dev \
        npm \
        pnpm \
        pkgconfig \
        tiff-dev \
        rustup

else
    apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community
fi

if [[ $devDeps == true ]]; then
    apk add --no-cache \
        cmake \
        delve \
        gdb \
        git \
        libunwind \
        libunwind-dev \
        pnpm \
        python3 \
        elfutils \
        elfutils-dev \
        boost-dev

    go install github.com/go-delve/delve/cmd/dlv@latest

    go install github.com/air-verse/air@latest
fi
