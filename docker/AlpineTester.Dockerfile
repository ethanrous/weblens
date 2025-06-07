ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.24

#
# Build UI
#
FROM node:${NODE_VERSION}-alpine AS web
WORKDIR /ui

COPY ui .

RUN --mount=type=cache,target=./node_modules npm install && npm run build

#
# Test server binary
#
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine

# Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh
RUN /tmp/install-deps.sh -b --dcraw

WORKDIR /src
COPY . .
# RUN go mod download
COPY --from=web /ui ./ui
RUN touch .env

ENV WEBLENS_MONGODB_URI=mongodb://weblens-test-mongo:27017/?replicaSet=rs0
ENV WEBLENS_MONGODB_NAME=weblens-test
ENV WEBLENS_LOG_LEVEL=debug
ENV WEBLENS_LOG_FORMAT=dev
ENV WEBLENS_ENV_PATH=/src/.env
ENV WEBLENS_DO_CACHE=false
ENV GOCACHE=/tmp/go-cache

ENTRYPOINT ["./scripts/testWeblens"]
CMD ["-n"]
