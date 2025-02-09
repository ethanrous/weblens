ARG ARCHITECTURE

ARG NODE_VERSION=22

# FROM --platform=linux/${ARCHITECTURE} debian:bookworm-slim AS tools
FROM --platform=linux/${ARCHITECTURE} nvidia/cuda:12.8.0-base-ubuntu24.04 AS tools
RUN apt-get update
RUN apt-get install -y git gcc nasm pkg-config make wget
# RUN wget https://developer.download.nvidia.com/compute/cuda/12.8.0/local_installers/cuda-repo-debian12-12-8-local_12.8.0-570.86.10-1_amd64.deb
# RUN dpkg -i cuda-repo-debian12-12-8-local_12.8.0-570.86.10-1_amd64.deb
# RUN cp /var/cuda-repo-debian12-12-8-local/cuda-*-keyring.gpg /usr/share/keyrings/
# RUN apt-get update
# RUN apt-get -y install cuda-toolkit-12-8
# RUN apt-get install -y nvidia-open

WORKDIR /tools/
RUN git clone git://git.videolan.org/ffmpeg.git
WORKDIR /tools/ffmpeg
RUN mkdir bin
RUN export PATH=$PATH:/usr/local/cuda/bin
# RUN ls -l /usr/local/cuda/bin
RUN ./configure --prefix=/usr --bindir="/tools/ffmpeg/bin" --enable-nonfree --enable-cuda-nvcc --enable-libnpp --extra-cflags=-I/usr/local/cuda/include --extra-ldflags=-L/usr/local/cuda/lib64 --disable-static --enable-shared
RUN make -j 8
RUN ls -l bin

#
# Build UI
#
FROM node:${NODE_VERSION}-bookworm-slim AS web

COPY ui .

RUN --mount=type=cache,target=/root/.npm npm install
RUN npm run build

#
# Build server binary
#
FROM --platform=linux/${ARCHITECTURE} golang:1.23.5-bookworm AS backend

# Install dependencies
RUN apt-get update
RUN apt-get install -y wget make g++
RUN apt-get install -y pkg-config libpng-dev libjpeg-dev libtiff-dev libwebp-dev libraw-dev libltdl-dev libzip-dev

# Build imagemagick
ARG IMAGEMAGICK_VERSION=7.1.1-43

RUN mkdir -p /tools/imagemagick
WORKDIR /tools/imagemagick

RUN wget -O imagemagick-$IMAGEMAGICK_VERSION.tar.gz https://github.com/ImageMagick/ImageMagick/archive/refs/tags/$IMAGEMAGICK_VERSION.tar.gz
RUN tar -xvf imagemagick-$IMAGEMAGICK_VERSION.tar.gz
WORKDIR ImageMagick-$IMAGEMAGICK_VERSION

RUN ./configure --with-modules --enable-shared
RUN make -j 8
RUN make install
RUN rm ./.gitignore
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
RUN apt-get install -y make gcc g++ exiftool ffmpeg 
RUN apt-get install -y pkg-config libpng-dev libjpeg-dev libtiff-dev libwebp-dev libraw-dev libltdl-dev libzip-dev
RUN apt-get install -y libdav1d6 librav1e0 libde265-0 libx265-199 libjpeg62-turbo libopenh264-7 libpng16-16 libnuma1 zlib1g 
RUN apt-get install -y libjpeg62-turbo liblcms2-2 zlib1g libgomp1
RUN apt-get install -y libjxl0.7 liblcms2-2 liblqr-1-0 libdjvulibre21 libjpeg62-turbo libopenjp2-7 libopenexr-3-1-30 libpng16-16 libtiff6 libwebpmux3 libwebpdemux2 libwebp7 libxml2 zlib1g liblzma5 libbz2-1.0 libgomp1
RUN apt-get autoremove -y \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

ARG IMAGEMAGICK_VERSION=7.1.1-43
COPY --from=backend /tools/imagemagick /tools/imagemagick
WORKDIR /tools/imagemagick/ImageMagick-$IMAGEMAGICK_VERSION
RUN make install
RUN make distclean
RUN ldconfig

WORKDIR /app
COPY --from=web dist /app/ui/dist
COPY --from=backend /server/weblens /app/weblens
COPY config/ /app/config
COPY images/brand /app/static

COPY --from=tools /tools/ffmpeg/bin/ffmpeg /usr/bin/ffmpeg
COPY --from=tools /tools/ffmpeg/bin/ffprobe /usr/bin/ffprobe

EXPOSE 8080

ENTRYPOINT ["/app/weblens"]
