BUFSTREAM_VERSION := 0.1.0

## Demo run commands

.PHONY: run-bufstream
run-bufstream:
	docker run --rm -p "9092:9092" -v "./config/bufstream.yaml:/bufstream.yaml" \
		us-docker.pkg.dev/buf-images-1/bufstream-public/images/bufstream:$(BUFSTREAM_VERSION) -c "/bufstream.yaml"

.PHONY: run-demo
run-demo:
	docker build -t bufstream/demo .
	docker run --rm bufstream/demo

.PHONY: run-docker-compose
run-docker-compose:
	docker compose up --build

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
