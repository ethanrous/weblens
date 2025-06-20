ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.24
ARG WEBLENS_ROUX_VERSION=v0

#
# Test server binary
#
FROM --platform=linux/${ARCHITECTURE} ethrous/weblens-roux:${WEBLENS_ROUX_VERSION}


WORKDIR /src
COPY . .
RUN touch .env

WORKDIR /src/ui
RUN --mount=type=cache,target=./node_modules npm install && npm run build

WORKDIR /src

ENV WEBLENS_MONGODB_URI=mongodb://weblens-test-mongo:27017/?replicaSet=rs0
ENV WEBLENS_MONGODB_NAME=weblens-test
ENV WEBLENS_LOG_LEVEL=debug
ENV WEBLENS_LOG_FORMAT=dev
ENV WEBLENS_ENV_PATH=/src/.env
ENV WEBLENS_DO_CACHE=false
ENV GOCACHE=/tmp/go-cache

ENTRYPOINT ["./scripts/testWeblens"]
CMD ["-n"]
