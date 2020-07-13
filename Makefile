.PHONY: docker/local
docker/local:
	DOCKER_BUILDKIT=1 docker build --ssh default -t waypoint-hzn:latest .
