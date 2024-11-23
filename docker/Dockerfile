# Small simple docker container for packaging Weblens

FROM debian:12-slim

RUN apt update
RUN apt-get install -y libvips exiftool ffmpeg libmagickwand-dev

WORKDIR /app
COPY ui/dist /app/ui/dist
COPY build/bin/weblensbin /app/weblens
COPY config/ /app/config
COPY images/brand /app/static

EXPOSE 8080

CMD ["/app/weblens"]
