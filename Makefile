GO_SOURCE=$(shell find . -path ./build -prune -o -iname "*.go")
TS_SOURCE=$(shell find ./ui/src/ -iname "*.ts*")

all: run

# WEBLENS_ENV_PATH=./scripts/.env ./build/bin/weblens
run: gen-ui gen-go
	./build/bin/weblens

run\:go: gen-go
	./build/bin/weblens

gen-go: $(GO_SOURCE)
	./scripts/startWeblens

gen-ui: $(TS_SOURCE)
	cd ui && \
	pnpm i && \
	pnpm run build

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

clean:
	rm -rf ./build/bin/*
	rm -rf ./ui/dist

really-clean:
	rm -rf ./build
	rm -rf ./ui/dist
	rm -rf ./ui/node_modules

# Publish the full docker image
docker\:build: $(GO_SOURCE) $(TS_SOURCE)
	./scripts/gogogadgetdocker.bash 

docker: test
	./scripts/gogogadgetdocker.bash -p

FORCE:
