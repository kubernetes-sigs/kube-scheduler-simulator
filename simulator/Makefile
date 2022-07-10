.PHONY: generate
generate: format
	go generate ./...

# re-generate openapi file for running api-server
.PHONY: openapi
openapi: vendor
	./hack/openapi.sh

.PHONY: lint
lint:
	golangci-lint run --timeout 30m ./...

.PHONY: format
format:
	find . -name "*.go" | grep -v "vendor" | grep -v "submodule" | xargs gci write --Section Standard --Section Default --Section "Prefix(sigs.k8s.io/kube-scheduler-simulator/simulator)"
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
