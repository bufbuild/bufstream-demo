BUFSTREAM_VERSION := 0.1.0

.DEFAULT_GOAL := run-docker-compose

## Demo run commands

.PHONY: run-docker-compose
run-docker-compose: # Run the demo within docker compose.
	docker compose up --build

.PHONY: run-bufstream
run-bufstream: # Run Bufstream within Docker.
	docker run --rm -p "9092:9092" -v "./config/bufstream.yaml:/bufstream.yaml" \
		us-docker.pkg.dev/buf-images-1/bufstream-public/images/bufstream:$(BUFSTREAM_VERSION) -c "/bufstream.yaml"

.PHONY: run-consume
run-consume: # Run the demo consumer within Docker.
	docker build -t bufstream/demo-consume -f Dockerfile.consume .
	docker run --rm bufstream/demo-consume

.PHONY: run-produce
run-produce: # Run the demo producer within Docker.
	docker build -t bufstream/demo-produce -f Dockerfile.produce .
	docker run --rm bufstream/demo-produce

## Development commands

.PHONY: build
build: # Build all code.
	go build ./...
	buf build

.PHONY: lint
lint: buf # Lint all code.
	buf lint

.PHONY: generate
generate: buf # Regenerate and format code.
	buf generate
	buf format -w
	gofmt -s -w .

.PHONY: upgrade
upgrade: # Upgrade dependencies.
	go get -u -t ./...
	go mod tidy -v
	buf dep update

.PHONY: buf
buf: # Install buf.
	go install github.com/bufbuild/buf/cmd/buf@latest
