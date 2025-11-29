ARG ARCHITECTURE

ARG GO_VERSION=1.25.4

# Download Go modules
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine AS download
COPY . .
RUN go mod download;

# Build agno library
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine AS agno-build

COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh;
RUN /tmp/install-deps.sh -b;

WORKDIR /src/agno
COPY ./agno /src/agno

RUN source "$HOME/.cargo/env"; cargo build --release --target x86_64-unknown-linux-musl;
## Copy agno binary to local library
RUN mkdir -p /agno/lib; cp /src/agno/target/x86_64-unknown-linux-musl/release/libagno.a /agno/lib/libagno.a; cp /src/agno/lib/agno.h /agno/lib/agno.h;

# Base image with dependencies installed
FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine

## Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh;
RUN /tmp/install-deps.sh -b;

COPY --from=download /go/pkg/mod /go/pkg/mod
COPY --from=agno-build /agno/lib /agno/lib
