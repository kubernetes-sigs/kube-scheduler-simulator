.PHONY: tools
tools:
	cd ./tools; \
	cat tools.go | grep "_" | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: lint
# run golangci-lint on all modules
lint:
	cd simulator/ && make lint

.PHONY: format
# format codes on all modules
format:
	cd simulator/ && make format

.PHONY: test
# run test on all modules
test:
	cd simulator/ && make test

.PHONY: mod-download
mod-download:
	cd simulator/ && go mod download -x

.PHONY: build
# build all modules
build:
	cd simulator/ && make build

.PHONY: docker_build
docker_build: docker_build_server docker_build_front

.PHONY: docker_build_server
docker_build_server: 
	docker build -t simulator-server ./simulator/

.PHONY: docker_build_front
docker_build_front: 
	docker build -t simulator-frontend ./web/

.PHONY: docker_up
docker_up:
	docker-compose up -d

.PHONY: docker_up_local
docker_up_local:
	docker-compose -f docker-compose-local.yml up -d

.PHONY: docker_build_and_up
docker_build_and_up: docker_build docker_up_local

.PHONY: docker_down
docker_down:
	docker-compose down --volumes

.PHONY: docker_down_local
docker_down_local:
	docker-compose -f docker-compose-local.yml down
