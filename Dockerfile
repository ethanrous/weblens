FROM debian:12-slim

RUN apt update
RUN apt-get install -y libvips exiftool ffmpeg

WORKDIR /app
COPY ui/dist /ui/dist
COPY api/weblens /app/
# COPY --from=api /app/weblens /app/weblens
COPY api/config /app/config

EXPOSE 8080

CMD ["/app/weblens"]
