# Small simple docker container for packaging Weblens

FROM alpine:latest

RUN apk upgrade --no-cache
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips
RUN apk add --no-cache imagemagick exiftool ffmpeg

WORKDIR /app
COPY ui/dist /app/ui/dist
COPY build/bin/weblensbin /app/weblens
COPY config/ /app/config
COPY images/brand /app/static

EXPOSE 8080

CMD ["/app/weblens"]
