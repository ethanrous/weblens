GO_SOURCE=$(shell find . -path ./_build -prune -o -iname "*.go")
TS_SOURCE=$(shell find ./weblens-vue/weblens-nuxt/ -iname "*.ts*")

all: run

# WEBLENS_ENV_PATH=./scripts/.env ./_build/bin/weblens
run: gen-ui gen-go
	./_build/bin/weblens

run\:go: gen-go
	./_build/bin/weblens

gen-go: $(GO_SOURCE)
	./scripts/startWeblens

gen-ui: $(TS_SOURCE)
	cd weblens-vue/weblens-nuxt && \
	pnpm i && \
	pnpm generate

ui: $(TS_SOURCE) FORCE
	cd ui && pnpm run dev

test: $(GO_SOURCE) $(TS_SOURCE)
	./scripts/testWeblens

core: FORCE
	./scripts/start.bash

dev: FORCE
	./scripts/start.bash --dev

dev-s: FORCE
	./scripts/start.bash --dev --secure

dev\:backup: FORCE
	./scripts/start.bash --dev -t backup

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

# Publish the full docker image
docker\:build: $(GO_SOURCE) $(TS_SOURCE)
	./scripts/gogogadgetdocker.bash 

docker: test
	./scripts/gogogadgetdocker.bash -p --skip-tests

FORCE:
