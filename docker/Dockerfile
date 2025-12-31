ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.25.0

# Build UI
FROM node:${NODE_VERSION}-alpine AS web

COPY weblens-vue /src
WORKDIR /src/weblens-nuxt

RUN npm install -g pnpm;
RUN pnpm install;
RUN pnpm generate;

# Build agno static library
FROM clux/muslrust:stable AS agno-builder
WORKDIR /src
COPY . .

WORKDIR /src
RUN ./scripts/build-agno.bash

# Build server binary
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine AS backend

## Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh;
RUN /tmp/install-deps.sh -b -a;

WORKDIR /src
COPY . .

RUN mkdir -p /src/services/media/agno/lib; cp /src/agno/lib/agno.h /src/services/media/agno/lib/agno.h;
COPY --from=agno-builder /src/services/media/agno/lib/libagno.a /src/services/media/agno/lib/libagno.a

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
