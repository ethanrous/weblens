FROM debian:12-slim

RUN apt update
RUN apt-get install -y libvips exiftool ffmpeg

WORKDIR /app
COPY ui/dist /ui/dist
COPY api/src/weblens /app/
# COPY --from=api /app/weblens /app/weblens
COPY api/config /app/config
COPY api/static /app/static

EXPOSE 8080

CMD ["/app/weblens"]
