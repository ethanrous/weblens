ARG NODE_VERSION=24.3.0

# Build UI
FROM --platform=linux/amd64 node:${NODE_VERSION}-alpine AS web
WORKDIR /src
COPY weblens-nuxt .
RUN npm install -g pnpm
RUN pnpm install 
RUN pnpm generate

FROM --platform=linux/amd64 golang:1.24-alpine AS backend
WORKDIR /src
COPY go.mod proxy.go ./
RUN go build -o ./proxy

FROM --platform=linux/amd64 alpine:latest
WORKDIR /app
COPY --from=web /src/.output/public/ .
COPY --from=backend /src/proxy .
ENV WEBLENS_NUXT_UI_PATH="/app"
EXPOSE 8989
CMD ["./proxy"]

# sudo docker build --platform linux/amd64" -t ethrous/weblens-vue:v0.1 ."
