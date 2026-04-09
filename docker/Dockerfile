ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.26
ARG AGNO_VERSION=0.0.8

# For AGNO_URL_OVERRIDE, you can specify a custom URL to download the AGNO library.
ARG AGNO_URL_OVERRIDE

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

# Build server binary
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-bookworm AS backend

ARG ARCHITECTURE
ARG AGNO_VERSION

RUN apt-get update && apt-get install -y --no-install-recommends unzip && rm -rf /var/lib/apt/lists/*;

WORKDIR /src

# Copy only what the Go build needs
COPY go.mod go.sum ./
COPY api/*.go api/go.mod api/go.sum api/api/ ./api/
COPY cmd/ ./cmd/
COPY docs/ ./docs/
COPY models/ ./models/
COPY modules/ ./modules/
COPY routers/ ./routers/
COPY services/ ./services/

# Download pre-built agno library to standard lib path
ARG AGNO_URL_OVERRIDE
# ARG GH_TOKEN
ENV AGNO_URL=${AGNO_URL_OVERRIDE:-https://github.com/ethanrous/agno/releases/download/v${AGNO_VERSION}/libagno-linux-x86_64-gpu.a}

# If AGNO_URL has actions/runs in it, we need to unzip the artifact instead of downloading it directly to the final location
RUN if echo ${AGNO_URL} | grep -q "/zip"; then \
        curl -fL -H "Authorization: Bearer ${GH_TOKEN}" \
        -H "Accept: application/vnd.github.v3+json" \
        ${AGNO_URL} -o /tmp/agno.zip && \
        unzip /tmp/agno.zip -d /usr/local/lib/ && \
        rm /tmp/agno.zip; \
    else \
        curl -fL ${AGNO_URL} -o /usr/local/lib/libagno.a; \
    fi;

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' \
    CGO_LDFLAGS="-L/usr/local/lib -lagno -lstdc++ -lm" \
    GOOS=linux GOARCH=${ARCHITECTURE} \
    go build -v -ldflags="-s -w" -o /server/weblens ./cmd/weblens/main.go;

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
    libvulkan-dev \
    vulkan-tools \
    libegl1 \
    libxext6 \
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
LABEL app.weblens.build.version=$WEBLENS_BUILD_VERSION

ENTRYPOINT ["/app/weblens"]
