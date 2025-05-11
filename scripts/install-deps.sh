apk upgrade --no-cache
apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev vips-poppler
apk add --no-cache bash build-base pkgconfig
apk add --no-cache tiff-dev libraw-dev libpng-dev libwebp-dev libheif

apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-poppler
apk add --no-cache ffmpeg tiff libraw libpng libwebp libheif libjpeg exiftool
apk add --no-cache build-base gcc jasper-dev libjpeg-turbo-dev lcms2-dev gettext gettext-dev gnu-libiconv-dev
