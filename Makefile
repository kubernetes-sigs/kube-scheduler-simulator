.PHONY: tools
tools:
	cd ./tools; \
	cat tools.go | grep "_" | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: generate
generate:
	go generate ./...

.PHONY: lint
lint:
	golangci-lint run --timeout 30m ./...

.PHONY: format
format:
	golangci-lint run --fix ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o ./bin/simulator ./simulator.go

.PHONY: start
serve: build
	./hack/start_simulator.sh

# re-generate openapi file for running api-server
.PHONY: openapi
openapi:
	./hack/openapi.sh

.PHONY: docker_build
docker_build: docker_build_server docker_build_front

.PHONY: docker_build_server
docker_build_server:
	docker build -t simulator-server .

.PHONY: docker_build_front
docker_build_front:
	docker build -t simulator-frontend ./web/

.PHONY: docker_up
docker_up:
	docker-compose up -d

.PHONY: docker_build_and_up
docker_build_and_up: docker_build docker_up

.PHONY: docker_down
docker_down:
	docker-compose down
