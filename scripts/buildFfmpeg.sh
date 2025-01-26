#!/bin/sh

apk add git gcc g++ gnutls-dev pkgconfig make nasm
git clone https://git.videolan.org/git/ffmpeg/nv-codec-headers.git

(
	cd nv-codec-headers || exit
	make
	make install
	mv /usr/local/lib/pkgconfig/ffnvcodec.pc /usr/lib/pkgconfig/
)

git clone https://github.com/FFmpeg/FFmpeg.git
cd FFmpeg || exit
export PKG_CONFIG_PATH=/usr/lib/pkgconfig
./configure --enable-cuda --enable-nonfree
make

cp ./ffmpeg /build/
