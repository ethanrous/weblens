all: dev

gen-ui: FORCE
	cd weblens-vue/weblens-nuxt && \
	pnpm i && \
	pnpm generate

agno: FORCE
	bash -c 'source ./scripts/lib/all.bash && build_agno'

test-server: FORCE
	./scripts/test-weblens.bash

test-ui: FORCE
	./scripts/test-playwright.bash

test: test-server test-ui

cover-ui: FORCE
	cd weblens-vue/weblens-nuxt && pnpm run test:e2e:coverage

cover:
	go tool cover -func ./_build/cover/coverage.out
cover-view:
	go tool cover -html ./_build/cover/coverage.out

dev: FORCE
	./scripts/start.bash "${@:1}"

swag: FORCE
	./scripts/swaggo.bash

roux: FORCE
	./scripts/build-base-image.bash -t v0
	docker push ethrous/weblens-roux:v0

clean:
	# Go stuff
	rm -rf ./_build/bin/*
	go clean -cache
	go clean -testcache

	# UI stuff
	cd weblens-vue/weblens-nuxt && pnpm run clean

really-clean: clean
	rm -rf ./_build
	rm -rf ./ui/dist
	rm -rf ./ui/node_modules

lint:
	golangci-lint run ./...
	cd weblens-vue/weblens-nuxt && pnpm run lint

lint\:fix:
	golangci-lint run --fix ./...
	cd weblens-vue/weblens-nuxt && pnpm run lint:fix


# Publish the full docker image
container: FORCE
	./scripts/gogogadgetdocker.bash --push --skip-tests --arch amd64

# Test docker image build (same as above but without pushing to registry)
container-test: FORCE
	./scripts/gogogadgetdocker.bash --skip-tests --arch amd64

precommit: lint test-server test-ui

FORCE:
