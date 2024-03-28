FROM golang:1.22 AS build-env

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/simulator

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o ./bin/simulator simulator.go

FROM alpine:3.14.0

COPY --from=build-env /go/src/simulator/bin/simulator /simulator
RUN chmod a+x /simulator
COPY --from=build-env /go/src/simulator/config.yaml /config.yaml

EXPOSE 1212
CMD ["/simulator"]
