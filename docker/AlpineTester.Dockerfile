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
# Test server binary
#
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine

# Install dependencies
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

# COPY go.mod go.sum ./
WORKDIR /src
COPY . .
# RUN go mod download
COPY --from=web /ui ./ui
RUN touch .env
# RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go mod download

ENV WEBLENS_MONGODB_URI=mongodb://weblens-test-mongo:27017/?replicaSet=rs0
ENV WEBLENS_MONGODB_NAME=weblens-test
ENV WEBLENS_LOG_LEVEL=debug
ENV LOG_FORMAT=dev
ENV WEBLENS_ENV_PATH=/src/.env
ENV WEBLENS_DO_CACHE=false
ENV GOCACHE=/tmp/go-cache

ENTRYPOINT ["./scripts/testWeblens"]
CMD ["-n"]
