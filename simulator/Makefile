.PHONY: generate
generate:
	go generate ./...
	make format

.PHONY: lint
lint:
	golangci-lint run --timeout 90m ./...

.PHONY: format
format:
	find . -name "*.go" | xargs gci write --section Standard --section Default --section "Prefix(sigs.k8s.io/kube-scheduler-simulator/simulator)"
	golangci-lint run --fix ./...

.PHONY: test
test: 
	go test ./...

.PHONY: mod-download
mod-download: ## Downloads the Go module
	go mod download -x

.PHONY: build
build:  
	go build -o ./bin/simulator ./simulator.go

.PHONY: start
start: build
	./hack/start_simulator.sh
