# Small simple docker container for packaging Weblens

FROM debian:12-slim

RUN apt-get update && apt-get install -y \
wget autoconf pkg-config build-essential curl libpng-dev libjpeg-dev libwebp-dev libraw-dev libheif-dev libvips exiftool ffmpeg darktable ghostscript

ENV IMAGEMAGICK_VERSION=7.1.1-41

RUN wget https://github.com/ImageMagick/ImageMagick/archive/refs/tags/${IMAGEMAGICK_VERSION}.tar.gz && \
    tar xzf ${IMAGEMAGICK_VERSION}.tar.gz && \
    rm ${IMAGEMAGICK_VERSION}.tar.gz

RUN sh ./ImageMagick-${IMAGEMAGICK_VERSION}/configure \
        --without-magick-plus-plus \
        --without-perl \
        --disable-openmp \
        --with-gvc=no \
        --disable-docs && \
    make -j$(nproc) && make install && ldconfig /usr/local/lib/

WORKDIR /app
COPY ui/dist /app/ui/dist
COPY build/bin/weblensbin /app/weblens
COPY config/ /app/config
COPY images/brand /app/static

EXPOSE 8080

CMD ["/app/weblens"]
