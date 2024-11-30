ARG ARCHITECTURE=$ARCHITECTURE

FROM --platform=linux/${ARCHITECTURE} golang:1.23-bookworm

# Install dependencies
RUN apt-get update
RUN apt-get install -y \
    libvips-dev \
    libheif-dev \
    libwebp-dev \
    libglib2.0-dev \
    libexpat1-dev

RUN apt-get update && apt-get install -y wget && \
    apt-get install -y autoconf pkg-config

RUN apt-get update && apt-get install -y wget && \
    apt-get install -y build-essential curl libpng-dev && \
    wget https://github.com/ImageMagick/ImageMagick/archive/refs/tags/7.1.1-41.tar.gz && \
    tar xzf 7.1.1-41.tar.gz && \
    rm 7.1.1-41.tar.gz

RUN sh ./ImageMagick-7.1.1-41/configure --prefix=/usr/local --with-bzlib=yes --with-fontconfig=yes --with-freetype=yes --with-gslib=yes --with-gvc=yes --with-jpeg=yes --with-jp2=yes --with-png=yes --with-tiff=yes --with-xml=yes --with-gs-font-dir=yes && \
    make -j && make install && ldconfig /usr/local/lib/


CMD ['go build -o weblens -ldflags="-s -w" /source/']
