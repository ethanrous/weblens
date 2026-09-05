ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.26

# For AGNO_URL_OVERRIDE, you can specify a custom URL to download the AGNO library.
ARG AGNO_URL_OVERRIDE

# Where the agno library comes from: "remote" downloads the pinned release,
# "local" takes it from the `agno` build context. Set by
# `gogogadgetdocker.bash --local-agno <path>`.
ARG AGNO_SOURCE=remote

# Build UI
FROM node:${NODE_VERSION}-alpine AS web

# Copy both weblens-vue and api/ts maintaining the relative path structure
# package.json has: "@ethanrous/weblens-api": "link:../../api/ts"
# From /src/weblens-nuxt, ../../api/ts resolves to /api/ts
COPY api/ts /api/ts
COPY weblens-vue /src

RUN npm install -g pnpm@10.18.3;

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

# Placeholder for the local agno checkout, overridden by
# `docker build --build-context agno=<dir>`. Only referenced when AGNO_SOURCE=local.
FROM scratch AS agno

# Download the pre-built agno library. The pinned version is derived from go.mod
# (single source of truth), keeping the native library in lockstep with the Go
# bindings. Runs on the build platform: it only fetches a file.
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-bookworm AS agno-remote

ARG AGNO_URL_OVERRIDE
# ARG GH_TOKEN

RUN apt-get update && apt-get install -y --no-install-recommends unzip && rm -rf /var/lib/apt/lists/*;

WORKDIR /agno
COPY go.mod .

# If AGNO_URL has actions/runs in it, we need to unzip the artifact instead of downloading it directly to the final location
RUN AGNO_VERSION=$(awk '/github.com\/ethanrous\/agno\/bindings\/go\/agno / {print $2}' go.mod); \
    AGNO_URL=${AGNO_URL_OVERRIDE:-https://github.com/ethanrous/agno/releases/download/${AGNO_VERSION}/libagno-linux-x86_64-gpu.a}; \
    if echo ${AGNO_URL} | grep -q "/zip"; then \
        curl -fL -H "Authorization: Bearer ${GH_TOKEN}" \
        -H "Accept: application/vnd.github.v3+json" \
        ${AGNO_URL} -o /tmp/agno.zip && \
        unzip /tmp/agno.zip -d /agno/ && \
        rm /tmp/agno.zip; \
    else \
        curl -fL ${AGNO_URL} -o /agno/libagno.a; \
    fi; \
    rm /agno/go.mod;

# Take the library, and the Go bindings it was built from, out of the build context
FROM scratch AS agno-local

COPY --from=agno libagno.a /agno/libagno.a
COPY --from=agno bindings /agno/bindings

FROM agno-${AGNO_SOURCE} AS agno-lib

# Build server binary
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-bookworm AS backend

ARG ARCHITECTURE

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

# Install the agno library resolved above. In local mode the stage also carries
# the Go bindings, which replace the pinned module so the library and its
# binding declarations stay in lockstep.
COPY --from=agno-lib /agno/ /agno/
RUN cp /agno/libagno.a /usr/local/lib/libagno.a && \
    if [ -f /agno/bindings/go/agno/go.mod ]; then \
        go mod edit -replace github.com/ethanrous/agno/bindings/go/agno=/agno/bindings/go/agno; \
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
