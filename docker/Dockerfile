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
RUN pnpm install;
RUN pnpm run build;

# Now build the Nuxt app
WORKDIR /src/weblens-nuxt
RUN pnpm install;
RUN pnpm generate;

# Build agno static library
FROM rust:slim AS agno-builder
  RUN apt-get update && apt-get install -y --no-install-recommends \
      g++ \
      musl-tools \
      && rm -rf /var/lib/apt/lists/*;

WORKDIR /src
COPY ./agno ./agno
COPY ./scripts ./scripts

RUN chmod +x ./scripts/lib/*.bash;
RUN rustup target add x86_64-unknown-linux-musl;

ENV CC_x86_64_unknown_linux_musl=musl-gcc

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,target=/src/agno/target \
    ./scripts/build-agno.bash "/build/lib/";
RUN ls -l /build/lib/libagno.a;

# Build server binary
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine AS backend

## Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh;
RUN /tmp/install-deps.sh -b -a;

WORKDIR /src
COPY . .

RUN mkdir -p /src/services/media/agno/lib; cp /src/agno/lib/agno.h /src/services/media/agno/lib/agno.h;
COPY --from=agno-builder /build/lib/libagno.a /src/services/media/agno/lib/libagno.a

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=${ARCHITECTURE} go build -v -ldflags="-s -w" -o /server/weblens ./cmd/weblens/main.go;

# Combine into final image
FROM --platform=linux/${ARCHITECTURE} alpine:latest AS final

COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh;
RUN /tmp/install-deps.sh;

WORKDIR /app

COPY --from=web /src/weblens-nuxt/.output/public/ /app/web/
COPY --from=backend /server/weblens /app/weblens
COPY config/ /app/config
COPY public /app/static

ARG WEBLENS_BUILD_VERSION=devel
ENV WEBLENS_BUILD_VERSION=$WEBLENS_BUILD_VERSION

EXPOSE 8080

ENTRYPOINT ["/app/weblens"]
