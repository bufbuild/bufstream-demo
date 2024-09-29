BUFSTREAM_VERSION := 0.1.0

## Demo run commands

.PHONY: run-docker-compose
run-docker-compose:
	docker compose up --build

.PHONY: run-bufstream
run-bufstream:
	docker run --rm -p "9092:9092" -v "./config/bufstream.yaml:/bufstream.yaml" \
		us-docker.pkg.dev/buf-images-1/bufstream-public/images/bufstream:$(BUFSTREAM_VERSION) -c "/bufstream.yaml"

.PHONY: run-consume
run-consume:
	docker build -t bufstream/demo-consume -f Dockerfile.consume .
	docker run --rm bufstream/demo-consume

.PHONY: run-produce
run-produce:
	docker build -t bufstream/demo-produce -f Dockerfile.produce .
	docker run --rm bufstream/demo-produce

## Development commands

.PHONY: build
build:
	go build ./...
	buf build

.PHONY: lint
lint: buf
	buf lint

.PHONY: generate
generate: buf
	buf generate
	buf format -w

.PHONY: upgrade
upgrade:
	go get -u -t ./...
	go mod tidy -v
	buf dep update

.PHONY: buf
buf:
	go install github.com/bufbuild/buf/cmd/buf@latest
