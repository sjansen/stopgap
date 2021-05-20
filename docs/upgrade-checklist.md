# Upgrade Checklist

- `docker-compose.yml`
  - [`https://hub.docker.com/r/amazon/dynamodb-local`](https://hub.docker.com/r/amazon/dynamodb-local)
- `docker/go/Dockerfile`
  - [`https://hub.docker.com/_/golang`](https://hub.docker.com/_/golang)
  - [`https://github.com/golangci/golangci-lint/releases`](https://github.com/golangci/golangci-lint/releases)
- `go.mod`
  - `go get -u`
- `scripts/seqdiag/go.mod`
  - `go get -u`
