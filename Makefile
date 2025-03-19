GO_SOURCE=$(shell find . -path ./build -prune -o -iname "*.go")
TS_SOURCE=$(shell find ./ui/src/ -iname "*.ts*")

all: run

run: gen-ui gen-go
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
	./scripts/testWeblens -a 

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

docker-push: test
	./scripts/gogogadgetdocker.bash -p

docker\:run: 
	docker compose --file ./scripts/docker-compose.yml --env-file ./scripts/.env up -d

docker: docker-build docker-run

FORCE:
