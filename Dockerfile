FROM --platform=${BUILDPLATFORM:-linux/amd64} node:18 as ui

RUN mkdir -p /app
WORKDIR /app

ARG build_tag
ENV REACT_APP_BUILD_TAG $build_tag

COPY ui/package*.json /app/
RUN npm ci --omit=dev --ignore-scripts

COPY ui /app
RUN npm run build

# Build backend binary #
FROM --platform=linux/amd64 golang:1.22-bookworm AS api

RUN mkdir -p /app
WORKDIR /app

# Install dependencies
RUN apt-get update
RUN apt-get install -y \
  libvips-dev \
  libheif-dev \
  libwebp-dev \
  libglib2.0-dev \
  libexpat1-dev

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=1

COPY api/go.mod api/go.sum /app/
RUN go mod download

ENV GIN_MODE=release

COPY api /app
RUN go build -o weblens .

FROM debian:12-slim

RUN apt update
RUN apt-get install -y libvips exiftool ffmpeg

WORKDIR /app
COPY --from=ui /app/build /ui/build
COPY --from=api /app/weblens /app/weblens
COPY api/config /app/config

EXPOSE 8080

CMD ["/app/weblens"]
