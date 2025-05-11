ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.24

#
# Build UI
#
FROM node:${NODE_VERSION}-alpine AS web
WORKDIR /ui

COPY ui .

RUN npm install --global pnpm
RUN pnpm install
RUN pnpm run build

#
# Build server binary
#
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine AS backend

# Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh
RUN /tmp/install-deps.sh

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 CGO_CFLAGS_ALLOW='-Xpreprocessor' GOOS=linux GOARCH=${ARCHITECTURE} go build -v -ldflags="-s -w" -o /server/weblens ./cmd/weblens/main.go

#
# Combine into final image
#
FROM --platform=linux/${ARCHITECTURE} alpine:latest

COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh
RUN /tmp/install-deps.sh

WORKDIR /tmp
RUN wget https://www.dechifro.org/dcraw/archive/dcraw-9.28.0.tar.gz ; \
    tar -xzvf dcraw-*.tar.gz ; \
    cd /tmp/dcraw ; \
    sed 's/-llcms2/-llcms2 -lintl/' <install >install.new && mv install.new install ; \
    chmod 755 install ;
RUN cd /tmp/dcraw; ./install

RUN rm -r /tmp/*

WORKDIR /app
COPY --from=web /ui/dist /app/web/
COPY --from=backend /server/weblens /app/weblens
COPY config/ /app/config
COPY images/brand /app/static

ARG WEBLENS_BUILD_VERSION=devel
ENV WEBLENS_BUILD_VERSION=$WEBLENS_BUILD_VERSION

EXPOSE 8080

ENTRYPOINT ["/app/weblens"]
