FROM golang:1.22 AS build-env
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=${TARGETOS:-linux}
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

WORKDIR /go/src/simulator

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache go mod download

COPY . .
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    go build -v -o ./bin/scheduler ./cmd/scheduler/scheduler.go

FROM alpine:3.14.0

COPY --from=build-env /go/src/simulator/bin/scheduler /scheduler
RUN chmod a+x /scheduler

CMD ["/scheduler"]
