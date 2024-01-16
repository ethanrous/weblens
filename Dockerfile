FROM --platform=${BUILDPLATFORM:-linux/amd64} node:18 as ui

RUN mkdir -p /app
WORKDIR /app

ARG build_tag
ENV REACT_APP_BUILD_TAG $build_tag

COPY ui/package*.json /app/
RUN npm ci --omit=dev --ignore-scripts

COPY ui /app
RUN npm run build

# FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.21.3-bookworm AS api
FROM --platform=linux/amd64 golang:1.21.3-bookworm AS api

RUN mkdir -p /app
WORKDIR /app

RUN apt update
RUN apt-get install -y libwebp-dev wget

ENV GOOS=linux
ENV GOARCH=amd64

COPY api/go.mod api/go.sum /app/
RUN go mod download

ENV GIN_MODE=release

COPY api /app
RUN go build -v -o weblens .

FROM debian:12-slim

RUN apt update
RUN apt-get install -y libwebp-dev exiftool ffmpeg

WORKDIR /app
COPY --from=ui /app/build /ui/build
COPY --from=api /app/weblens /app/weblens
COPY api/config /app/config

EXPOSE 8080

CMD ["/app/weblens"]
