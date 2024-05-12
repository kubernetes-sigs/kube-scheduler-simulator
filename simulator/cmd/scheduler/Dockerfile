FROM golang:1.22 AS build-env

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/simulator

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o ./bin/scheduler ./cmd/scheduler/scheduler.go

FROM alpine:3.14.0

COPY --from=build-env /go/src/simulator/bin/scheduler /scheduler
RUN chmod a+x /scheduler

COPY ./docs/sample/debuggable-scheduler/scheduler.yaml ./docs/sample/debuggable-scheduler/kubeconfig.yaml ./

CMD ["/scheduler", "--config", "scheduler.yaml", "--master", "http://simulator-cluster:3131"]
