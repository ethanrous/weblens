ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.24

# Build UI
FROM node:${NODE_VERSION}-alpine AS web
WORKDIR /ui

COPY ui .

RUN --mount=type=cache,target=./node_modules npm install && npm run build

# Build server binary
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine AS backend

## Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh
RUN /tmp/install-deps.sh -b

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=${ARCHITECTURE} go build -v -ldflags="-s -w" -o /server/weblens ./cmd/weblens/main.go

# Combine into final image
FROM --platform=linux/${ARCHITECTURE} alpine:latest

COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh
RUN /tmp/install-deps.sh --dcraw

WORKDIR /app
COPY --from=web /ui/dist /app/web/
COPY --from=backend /server/weblens /app/weblens
COPY config/ /app/config
COPY public /app/static

ARG WEBLENS_BUILD_VERSION=devel
ENV WEBLENS_BUILD_VERSION=$WEBLENS_BUILD_VERSION

EXPOSE 8080

ENTRYPOINT ["/app/weblens"]
