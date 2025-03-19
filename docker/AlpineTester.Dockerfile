ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.24

#
# Build UI
#
FROM node:${NODE_VERSION}-alpine AS web
WORKDIR /ui

COPY ui .

RUN npm install --global pnpm
RUN pnpm install
RUN pnpm run build

#
# Test server binary
#
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine

# Install dependencies
RUN apk upgrade --no-cache
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev vips-poppler
RUN apk add --no-cache bash build-base pkgconfig
RUN apk add --no-cache tiff-dev libraw-dev libpng-dev libwebp-dev libheif exiftool jasper-dev
RUN apk add --no-cache ffmpeg tiff libraw libpng libwebp libheif libjpeg exiftool 
# Build deps
RUN apk add --no-cache build-base gcc jasper-dev libjpeg-turbo-dev lcms2-dev gettext gettext-dev gnu-libiconv-dev 

WORKDIR /tmp
RUN wget https://www.dechifro.org/dcraw/archive/dcraw-9.28.0.tar.gz ; \
    tar -xzvf dcraw-*.tar.gz ; \
    cd /tmp/dcraw ; \
    sed 's/-llcms2/-llcms2 -lintl/' <install >install.new && mv install.new install ; \
    chmod 755 install ;
RUN cd /tmp/dcraw; ./install

# COPY go.mod go.sum ./
WORKDIR /src
COPY . .
COPY --from=web /ui ./ui
RUN go mod download

ENV MONGODB_URI=mongodb://admin:admin@weblens-mongo:27017
ENV LOG_LEVEL=debug
ENV APP_ROOT=/src
# RUN go test -v ./...

ENTRYPOINT ["/usr/local/go/bin/go", "test"]
CMD ["-v", "./..."]
