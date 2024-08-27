ARG ARCHITECTURE=$ARCHITECTURE

FROM --platform=linux/${ARCHITECTURE} golang:1.23-bookworm

# Install dependencies
RUN apt-get update
RUN apt-get install -y \
    libvips-dev \
    libheif-dev \
    libwebp-dev \
    libglib2.0-dev \
    libexpat1-dev

CMD ['go build -o weblens -ldflags="-s -w" /source/']
