# Waypoint Horizon API (waypoint-hzn)

This repository contains the Waypoint API frontend to the [Horizon](https://github.com/hashicorp/horizon)
service. Horizon is an internal API that isn't fully exposed to the public.
It expects applications to implement application-specific services in front of
it to perform application-specific business logic. This project (waypoint-hzn)
is the Waypoint-specific API frontend for Horizon.

## Development

The project is made to be developed primarily through a unit-test driven
workflow via `go test`. This tests fully communicating to an in-memory Horizon
server to verify behaviors.

Some dependencies must be running for the unit tests. These are all contained
in the Docker Compose configuration. Therefore, to run all tests:

```
$ docker-compuse up -d
$ go test ./... -p 1
```

To build the project, you can use `go build ./cmd/waypoint-hzn` directly or
build the included Docker image with `make docker/local`.
