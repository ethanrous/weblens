ARG ARCHITECTURE

ARG GO_VERSION=1.24

FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine AS download
COPY . .
RUN go mod download

FROM --platform=linux/${ARCHITECTURE} golang:${GO_VERSION}-alpine

# Install dependencies
COPY scripts/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh
RUN /tmp/install-deps.sh -b --dcraw

COPY --from=download /go/pkg/mod /go/pkg/mod
