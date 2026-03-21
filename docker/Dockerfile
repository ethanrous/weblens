ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.26

# Build UI
FROM node:${NODE_VERSION}-alpine AS web

# Copy both weblens-vue and api/ts maintaining the relative path structure
# package.json has: "@ethanrous/weblens-api": "link:../../api/ts"
# From /src/weblens-nuxt, ../../api/ts resolves to /api/ts
COPY api/ts /api/ts
COPY weblens-vue /src

RUN npm install -g pnpm;

# Build the API client first
WORKDIR /api/ts
RUN --mount=type=cache,target=/pnpm/store \
    pnpm fetch --frozen-lockfile;
RUN --mount=type=cache,target=/pnpm/store \
    pnpm install --frozen-lockfile --prod --offline;
RUN pnpm run build;

# Now build the Nuxt app
WORKDIR /src/weblens-nuxt
RUN --mount=type=cache,target=/pnpm/store \
            pnpm fetch --frozen-lockfile;
RUN --mount=type=cache,target=/pnpm/store \
    pnpm install --frozen-lockfile --prod --offline;
RUN pnpm generate;

# Build agno static library (targeting glibc for Debian compatibility)
FROM rust:slim-bookworm AS agno-builder

# cmake + curl: needed to build libde265 and libheif from source as static libraries
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    cmake \
    curl \
    && rm -rf /var/lib/apt/lists/*;

# Build libde265 as a static library.
# Apt only ships a shared-library package; we need .a so it can be merged into libagno.a.
ARG LIBDE265_VERSION=1.0.15
RUN curl -fsSL "https://github.com/strukturag/libde265/releases/download/v${LIBDE265_VERSION}/libde265-${LIBDE265_VERSION}.tar.gz" \
    | tar -xz \
    && cmake -S "libde265-${LIBDE265_VERSION}" -B libde265-build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX=/usr/local \
        -DBUILD_SHARED_LIBS=OFF \
    && cmake --build libde265-build --parallel "$(nproc)" \
    && cmake --install libde265-build \
    && rm -rf "libde265-${LIBDE265_VERSION}" libde265-build;

# Build libheif 1.21.2 as a static library.
# Debian apt ships ~1.17.x; libheif-sys requires >= 1.21 (enforced by pkg-config at cargo build time).
# Built against our static libde265 so both can be merged into the final libagno.a.
ARG LIBHEIF_VERSION=1.21.2
RUN curl -fsSL "https://github.com/strukturag/libheif/releases/download/v${LIBHEIF_VERSION}/libheif-${LIBHEIF_VERSION}.tar.gz" \
    | tar -xz \
    && cmake -S "libheif-${LIBHEIF_VERSION}" -B libheif-build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX=/usr/local \
        -DBUILD_SHARED_LIBS=OFF \
        -DENABLE_PLUGIN_LOADING=OFF \
        -DBUILD_TESTING=OFF \
        -DWITH_EXAMPLES=OFF \
    && cmake --build libheif-build --parallel "$(nproc)" \
    && cmake --install libheif-build \
    && rm -rf "libheif-${LIBHEIF_VERSION}" libheif-build;

ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/local/lib/x86_64-linux-gnu/pkgconfig:/usr/lib/x86_64-linux-gnu/pkgconfig

WORKDIR /src
COPY ./agno ./agno
COPY ./scripts ./scripts

RUN chmod +x ./scripts/lib/*.bash;

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,id=agno-target-bookworm,target=/src/agno/target \
    ./scripts/build-agno.bash "/build/lib/";
RUN ls -l /build/lib/libagno.a;

# Build server binary
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-bookworm AS backend

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    liblcms2-dev \
    libpng-dev \
    libraw-dev \
    libwebp-dev \
    libtiff-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*;

WORKDIR /src
COPY . .

RUN mkdir -p /src/services/media/agno/lib && cp /src/agno/lib/agno.h /src/services/media/agno/lib/agno.h;
COPY --from=agno-builder /build/lib/libagno.a /src/services/media/agno/lib/libagno.a

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=${ARCHITECTURE} go build -v -ldflags="-s -w" -o /server/weblens ./cmd/weblens/main.go;

# Combine into final image
FROM --platform=linux/${ARCHITECTURE} debian:bookworm-slim AS final

# Install runtime dependencies
# libvulkan1: Vulkan loader for GPU acceleration (wgpu backend)
# NVIDIA drivers + Vulkan ICD are injected via nvidia-container-toolkit
# NOTE: Do NOT install mesa-vulkan-drivers - it provides llvmpipe (CPU software
# renderer) which wgpu may select instead of the actual Nvidia GPU
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    libvulkan1 \
    ca-certificates \
    libvulkan-dev vulkan-tools libegl1 libxext6 \
    && rm -rf /var/lib/apt/lists/*;

WORKDIR /app

COPY --from=web /src/weblens-nuxt/.output/public/ /app/web/
COPY --from=backend /server/weblens /app/weblens
COPY public /app/static

RUN mkdir -p /usr/share/vulkan/icd.d/
RUN echo '{"file_format_version" : "1.0.0", "ICD": { "library_path": "libEGL_nvidia.so.0", "api_version" : "1.3" }}' > /usr/share/vulkan/icd.d/nvidia_icd.json

ARG WEBLENS_BUILD_VERSION=devel
ENV WEBLENS_BUILD_VERSION=$WEBLENS_BUILD_VERSION

EXPOSE 8080

LABEL org.opencontainers.image.source="https://github.com/ethanrous/weblens"

ENTRYPOINT ["/app/weblens"]
