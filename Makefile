# GO_SOURCE=$(shell find . -path ./_build -prune -o -iname "*.go")
# TS_SOURCE=$(shell find ./weblens-vue/weblens-nuxt/ -iname "*.ts*")

all: run

# WEBLENS_ENV_PATH=./scripts/.env ./_build/bin/weblens
run: gen-ui gen-go
	./_build/bin/weblens

run\:go: gen-go
	./_build/bin/weblens

gen-go: FORCE
	./scripts/startWeblens

gen-ui: FORCE
	cd weblens-vue/weblens-nuxt && \
	pnpm i && \
	pnpm generate

ui: FORCE
	cd ui && pnpm run dev

test: FORCE
	./scripts/test-weblens.bash

cover:
	go tool cover -func ./_build/cover/coverage.out
cover-view:
	go tool cover -html ./_build/cover/coverage.out

dev: FORCE
	./scripts/start.bash --dynamic "${@:1}"

dev-s: FORCE
	./scripts/start.bash --dev --secure $(ARGS)

dev\:backup: FORCE
	./scripts/start.bash --dev -t backup

dev\:static: FORCE
	./scripts/start.bash --rebuild --dev "${@:2}"

swag: FORCE
	./scripts/swaggo

roux: FORCE
	./scripts/build-base-image.bash -t v0
	docker push ethrous/weblens-roux:v0

clean:
	rm -rf ./_build/bin/*
	rm -rf ./ui/dist

really-clean:
	rm -rf ./_build
	rm -rf ./ui/dist
	rm -rf ./ui/node_modules

lint:
	golangci-lint run ./...
	cd weblens-vue/weblens-nuxt && pnpm run lint

# Publish the full docker image
docker\:build: $(GO_SOURCE) $(TS_SOURCE)
	./scripts/gogogadgetdocker.bash 

docker: FORCE
	./scripts/gogogadgetdocker.bash -p --skip-tests

FORCE:
