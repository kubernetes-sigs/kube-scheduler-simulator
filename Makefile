BUILD      ?= build

USE_BUILDX ?=
HOSTARCH := $(shell uname -m)
ifeq ($(HOSTARCH),x86_64)
HOSTARCH := amd64
endif
BUILDX_PLATFORM ?= linux/$(HOSTARCH)

# Setup buildx flags
ifneq ("$(USE_BUILDX)","")
BUILD = buildx build --platform=$(BUILDX_PLATFORM) -o type=docker
endif

.PHONY: tools
tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
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
docker_build: docker_build_server docker_build_scheduler docker_build_front

.PHONY: docker_build_server
docker_build_server:
	docker $(BUILD) -f simulator/cmd/simulator/Dockerfile -t simulator-server simulator

.PHONY: docker_build_scheduler
docker_build_scheduler:
	docker $(BUILD) -f simulator/cmd/scheduler/Dockerfile -t simulator-scheduler simulator

.PHONY: docker_build_front
docker_build_front:
	docker $(BUILD) -t simulator-frontend ./web/

.PHONY: docker_up
docker_up:
	docker compose up -d

.PHONY: docker_up_local
docker_up_local:
	docker compose -f docker-compose-local.yml up -d

.PHONY: docker_build_and_up
docker_build_and_up: docker_build docker_up_local

.PHONY: docker_down
docker_down:
	docker compose down --volumes

.PHONY: docker_down_local
docker_down_local:
	docker compose -f docker-compose-local.yml down
