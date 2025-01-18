ARG ARCHITECTURE=$ARCHITECTURE

FROM --platform=linux/${ARCHITECTURE} golang:1.23-alpine

# Install dependencies
RUN apk upgrade --no-cache
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev
RUN apk add --no-cache bash build-base pkgconfig
RUN apk add --no-cache imagemagick imagemagick-dev tiff-dev libraw-dev libpng-dev libwebp-dev

CMD ['go build -o weblens -ldflags="-s -w" /source/']
