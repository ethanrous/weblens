ARG ARCHITECTURE

ARG NODE_VERSION=22.14.0
ARG GO_VERSION=1.25.0

#
# Test server binary
#
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine

WORKDIR /tmp

# Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh
RUN /tmp/install-deps.sh --dev 

WORKDIR /src

ENV WEBLENS_LOG_LEVEL=trace
ENV WEBLENS_LOG_FORMAT=dev
ENV WEBLENS_ENV_PATH=/src/.env

ENTRYPOINT ["/src/scripts/devel.bash"]
CMD [""]
