ARG ARCHITECTURE

ARG NODE_VERSION=22

#
# Build UI
#
FROM node:${NODE_VERSION}-alpine AS web

COPY ui .

RUN --mount=type=cache,target=/root/.npm npm install
RUN npm run build

#
# Build server binary
#
FROM --platform=linux/${ARCHITECTURE} golang:1.23.5-alpine AS backend

# Install dependencies
RUN apk upgrade --no-cache
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev
RUN apk add --no-cache bash build-base pkgconfig
RUN apk add --no-cache tiff-dev libraw-dev libpng-dev libwebp-dev libheif 

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=${ARCHITECTURE} go build -v -ldflags="-s -w" -o /server/weblens ./cmd/weblens/main.go

#
# Combine into final image
#
FROM --platform=linux/${ARCHITECTURE} alpine:latest

RUN apk upgrade --no-cache
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips
RUN apk add --no-cache ffmpeg tiff libraw libpng libwebp libheif libjpeg 

# Build deps
RUN apk add --no-cache build-base gcc jasper-dev libjpeg-turbo-dev lcms2-dev gettext gettext-dev gnu-libiconv-dev

WORKDIR /tmp
# build iconv
# RUN cd /tmp; \
# 		wget https://ftp.gnu.org/pub/gnu/libiconv/libiconv-1.17.tar.gz; \
# 		tar -xzvf libiconv-*.tar.gz ;\
# 		cd libiconv* ; \
# 		./configure ;
# RUN cd /tmp/libiconv* ; \
# 		make ; \
# 		make install ;

RUN wget https://www.dechifro.org/dcraw/archive/dcraw-9.28.0.tar.gz ; \
    tar -xzvf dcraw-*.tar.gz ; \
    cd /tmp/dcraw ; \
    sed 's/-llcms2/-llcms2 -lintl/' <install >install.new && mv install.new install ; \
    chmod 755 install ;
RUN cd /tmp/dcraw; ./install

RUN rm -r /tmp/*

WORKDIR /app
COPY --from=web dist /app/ui/dist
COPY --from=backend /server/weblens /app/weblens
COPY config/ /app/config
COPY images/brand /app/static
# COPY build/ffmpeg /usr/bin/ffmpeg

EXPOSE 8080

ENTRYPOINT ["/app/weblens"]
