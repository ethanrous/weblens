set -e

buildDeps=false
devDeps=false
buildDcraw=false

usage() {
    echo "Usage: $0 [-b|--build] [-d|--dev] [--dcraw]"
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
    "--dcraw")
        buildDcraw=true
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
apk add --no-cache ffmpeg exiftool jasper

if [[ $buildDeps == true ]]; then
    apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev

    apk add --no-cache \
        build-base \
        gcc \
        pkgconfig \
        vips-dev \
        lcms2-dev \
        tiff-dev \
        libraw-dev \
        libpng-dev \
        libwebp-dev \
        bash \
        npm
else
    apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-poppler
fi

if [[ $buildDcraw == true ]]; then
    apk add --no-cache --virtual .dcraw-deps \
        build-base \
        gcc \
        pkgconfig \
        lcms2-dev \
        jasper-dev \
        libjpeg-turbo-dev \
        gettext-dev \
        gnu-libiconv-dev

    cd /tmp
    wget https://raw.githubusercontent.com/ncruces/dcraw/refs/heads/master/dcraw.c

    gcc -O4 -march=native -o /usr/local/bin/dcraw dcraw.c \
        -l:libm.a -ljasper -ljpeg -llcms2 -lintl -s -DLOCALEDIR=\"/usr/local/share/locale/\"

    rm -rf /tmp/dcraw*

    apk del .dcraw-deps

    cd /tmp
fi

if [[ $devDeps == true ]]; then
    apk add --no-cache pnpm

    go install github.com/air-verse/air@latest
fi
