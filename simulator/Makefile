.PHONY: generate
generate: vendor
	go generate ./... && \
	./hack/openapi.sh

.PHONY: lint
lint:
	golangci-lint run --timeout 30m ./...

.PHONY: format
format:
	golangci-lint run --fix ./...

.PHONY: test
test: 
	go test ./...

.PHONY: mod-download
mod-download: ## Downloads the Go module
	go mod download -x

.PHONY: vendor
vendor: mod-download
	go mod vendor

.PHONY: build
build:  
	go build -o ./bin/simulator ./simulator.go

.PHONY: start
start: build
	./hack/start_simulator.sh
