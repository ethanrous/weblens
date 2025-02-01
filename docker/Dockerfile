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
RUN apk add --no-cache imagemagick imagemagick-dev tiff-dev libraw-dev libpng-dev libwebp-dev libheif imagemagick-pdf

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

ENV NVIDIA_VISIBLE_DEVICES="all"
ENV NVIDIA_DRIVER_CAPABILITIES="compute,video,utility"

RUN apk upgrade --no-cache
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips
RUN apk add --no-cache imagemagick exiftool ffmpeg tiff libraw libpng libwebp libheif imagemagick-pdf

WORKDIR /app
COPY --from=web dist /app/ui/dist
COPY --from=backend /server/weblens /app/weblens
COPY config/ /app/config
COPY images/brand /app/static
COPY build/ffmpeg /usr/bin/ffmpeg

EXPOSE 8080

ENTRYPOINT ["/app/weblens"]
