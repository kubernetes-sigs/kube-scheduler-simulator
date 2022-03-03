FROM golang:1.16.6 AS build-env

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/simulator-server

COPY go.mod go.sum ./
COPY . .
RUN make build

FROM alpine:3.14.0

COPY --from=build-env /go/src/simulator-server/bin/simulator /simulator
RUN chmod a+x /simulator

EXPOSE 1212
CMD ["/simulator"]
