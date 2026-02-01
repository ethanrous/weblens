all: run

gen-ui: FORCE
	cd weblens-vue/weblens-nuxt && \
	pnpm i && \
	pnpm generate

agno: FORCE
	bash -c 'source ./scripts/lib/all.bash && build_agno'

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
	./scripts/swaggo.bash

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

dev-container: FORCE
	./scripts/gogogadgetdocker.bash -p -s -a amd64

precommit: lint test

FORCE:
