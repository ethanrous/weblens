set -e

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
# apk add --no-cache ffmpeg jasper poppler-glib fontconfig libraw
# apk add --update --no-cache --virtual .ms-fonts msttcorefonts-installer &&
#     update-ms-fonts 2>/dev/null &&
#     fc-cache -fv &&
#     apk del .ms-fonts

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

    rustup-init -y --no-modify-path
    . "$HOME/.cargo/env"
    TRIPLE_VENDOR="x86_64-unknown-linux-musl"
    TRIPLE="x86_64-linux-musl"
    rustup target add $TRIPLE_VENDOR || exit 1

    mkdir -p /opt/musl

    MUSL_VERSION="aarch64-linux-musl-cross"
    curl -L "https://musl.cc/${MUSL_VERSION}.tgz" | tar xz -C /opt/musl

    ln -sf /opt/musl/"${MUSL_VERSION}"/bin/"${TRIPLE}"-gcc /usr/local/bin/"${TRIPLE}"-gcc &&
        ln -sf /opt/musl/"${MUSL_VERSION}"/bin/"${TRIPLE}"-g++ /usr/local/bin/"${TRIPLE}"-g++ &&
        ln -sf /opt/musl/"${MUSL_VERSION}"/bin/"${TRIPLE}"-ar /usr/local/bin/"${TRIPLE}"-ar &&
        ln -sf /opt/musl/"${MUSL_VERSION}"/bin/"${TRIPLE}"-nm /usr/local/bin/"${TRIPLE}"-nm &&
        ln -sf /opt/musl/"${MUSL_VERSION}"/bin/"${TRIPLE}"-strip /usr/local/bin/"${TRIPLE}"-strip &&
        ln -sf /opt/musl/"${MUSL_VERSION}"/bin/"${TRIPLE}"-ranlib /usr/local/bin/"${TRIPLE}"-ranlib

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

    # mkdir /debug && cd /debug || exit 1
    # git clone https://github.com/KDE/heaptrack.git
    # cd heaptrack || exit 1
    # mkdir build && cd build || exit 1
    # cmake -DCMAKE_BUILD_TYPE=Release ..
    # make -j$(nproc)

    go install github.com/go-delve/delve/cmd/dlv@latest

    go install github.com/air-verse/air@latest
fi

# if [[ $agno == true ]]; then
# cd /agno/ || exit 1
#
# . "$HOME/.cargo/env"
# echo '[target.aarch64-unknown-linux-musl]
# linker = "aarch64-linux-musl-g++"' >~/.cargo/config
# export CARGO_TARGET_AARCH64_UNKNOWN_LINUX_MUSL_LINKER=aarch64-linux-musl-g++
# PDFIUM_STATIC_LIB_PATH="/agno/libpdfium" RUSTFLAGS='-C link-arg=-lpdfium -C link-arg=-lstdc++' cargo build --release --target aarch64-unknown-linux-musl
# cp target/aarch64-unknown-linux-musl/release/libagno.a ./lib/
# fi
