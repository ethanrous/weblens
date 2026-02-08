ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.25.6

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
FROM rust:slim AS agno-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    g++ \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY ./agno ./agno
COPY ./scripts ./scripts

RUN chmod +x ./scripts/lib/*.bash;

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,target=/src/agno/target \
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
COPY config/ /app/config
COPY public /app/static

RUN mkdir -p /usr/share/vulkan/icd.d/
RUN echo '{"file_format_version" : "1.0.0", "ICD": { "library_path": "libEGL_nvidia.so.0", "api_version" : "1.3" }}' > /usr/share/vulkan/icd.d/nvidia_icd.json

ARG WEBLENS_BUILD_VERSION=devel
ENV WEBLENS_BUILD_VERSION=$WEBLENS_BUILD_VERSION

EXPOSE 8080

LABEL org.opencontainers.image.source="https://github.com/ethanrous/weblens"

ENTRYPOINT ["/app/weblens"]
