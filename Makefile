.PHONY: docker/local
docker/local:
	DOCKER_BUILDKIT=1 docker build \
					--ssh default \
					--secret id=ssh.config,src="${HOME}/.ssh/config" \
					--secret id=ssh.key,src="${HOME}/.ssh/config" \
					-t waypoint-hzn:latest \
					.

.PHONY: docker/evanphx
docker/evanphx:
	DOCKER_BUILDKIT=1 docker build -f hack/Dockerfile.evanphx \
					--ssh default \
					-t waypoint-hzn:latest \
					.

.PHONY: proto
proto:
	protoc -I proto/ \
		--go_out=pkg/pb/ --go_opt=paths=source_relative \
		--go-grpc_out=pkg/pb/ --go-grpc_opt=paths=source_relative \
		--validate_out="lang=go:pkg/pb/" \
		--validate_opt=paths=source_relative \
		proto/server.proto