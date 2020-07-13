# syntax = docker/dockerfile:experimental

FROM golang:alpine AS builder

RUN apk add --no-cache git gcc libc-dev openssh

RUN mkdir -p /tmp/prime
COPY go.sum /tmp/prime
COPY go.mod /tmp/prime

WORKDIR /tmp/prime

RUN mkdir -p -m 0600 ~/.ssh \
    && ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
RUN git config --system url."ssh://git@github.com/".insteadOf "https://github.com/"
ENV GOPRIVATE=github.com/hashicorp
RUN --mount=type=ssh go mod download

COPY . /tmp/src
WORKDIR /tmp/src

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=ssh go build -o /tmp/waypoint-hzn -ldflags "-X main.sha1ver=`git rev-parse HEAD` -X main.buildTime=$(date +'+%FT%T.%N%:z')" ./cmd/waypoint-hzn

FROM alpine

COPY --from=builder /tmp/waypoint-hzn /usr/bin/waypoint-hzn

RUN GRPC_HEALTH_PROBE_VERSION=v0.3.2 && \
    wget -qO/usr/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /usr/bin/grpc_health_probe

COPY ./migrations /migrations

ENTRYPOINT ["/usr/bin/waypoint-hzn"]
