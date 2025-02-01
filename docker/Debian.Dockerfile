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
FROM --platform=linux/${ARCHITECTURE} golang:1.23.5-bookworm AS backend

# Install dependencies
RUN apt-get update
RUN apt-get install -y wget
RUN apt-get install -y pkg-config libpng-dev libjpeg-dev libtiff-dev libwebp-dev libraw-dev libltdl-dev libzip-dev

# Build imagemagick
ARG IMAGEMAGICK_VERSION=7.1.1-43

RUN mkdir -p /tools/imagemagick
WORKDIR /tools/imagemagick

RUN wget -O imagemagick-$IMAGEMAGICK_VERSION.tar.gz https://github.com/ImageMagick/ImageMagick/archive/refs/tags/$IMAGEMAGICK_VERSION.tar.gz
RUN tar -xvf imagemagick-$IMAGEMAGICK_VERSION.tar.gz
WORKDIR ImageMagick-$IMAGEMAGICK_VERSION

RUN ./configure --with-modules --enable-shared
RUN make
RUN make install
RUN make distclean
RUN ldconfig

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=${ARCHITECTURE} go build -v -ldflags="-s -w" -o /server/weblens ./cmd/weblens/main.go

#
# Combine into final image
#
FROM --platform=linux/${ARCHITECTURE} debian:bookworm-slim 

ENV NVIDIA_VISIBLE_DEVICES="all"
ENV NVIDIA_DRIVER_CAPABILITIES="compute,video,utility"

RUN apt-get update
RUN apt-get install -y libdav1d6 librav1e0 libde265-0 libx265-199 libjpeg62-turbo libopenh264-7 libpng16-16 libnuma1 zlib1g
RUN apt-get install -y libjpeg62-turbo liblcms2-2 zlib1g libgomp1
RUN apt-get install -y libjxl0.7 liblcms2-2 liblqr-1-0 libdjvulibre21 libjpeg62-turbo libopenjp2-7 libopenexr-3-1-30 libpng16-16 libtiff6 libwebpmux3 libwebpdemux2 libwebp7 libxml2 zlib1g liblzma5 libbz2-1.0 libgomp1
RUN apt-get autoremove -y \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=web dist /app/ui/dist
COPY --from=backend /server/weblens /app/weblens
COPY config/ /app/config
COPY images/brand /app/static
COPY build/ffmpeg /usr/bin/ffmpeg
COPY build/ffprobe /usr/bin/ffprobe

EXPOSE 8080

ENTRYPOINT ["/app/weblens"]
