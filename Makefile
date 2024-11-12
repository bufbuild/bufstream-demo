BUFSTREAM_VERSION := 0.2.0

.DEFAULT_GOAL := docker-compose-run

## Demo run commands

.PHONY: docker-compose-run
docker-compose-run: # Run the demo within docker compose.
	docker compose up --build

.PHONY: docker-bufstream-run
docker-bufstream-run: # Run Bufstream within Docker.
	docker run --rm -p 9092:9092 -v ./config/bufstream.yaml:/bufstream.yaml \
		"us-docker.pkg.dev/buf-images-1/bufstream-public/images/bufstream:$(BUFSTREAM_VERSION)" \
			--config /bufstream.yaml

.PHONY: docker-produce-build
docker-produce-build:
	docker build -t bufstream/demo-produce -f Dockerfile.produce .

.PHONY: docker-consume-build
docker-consume-build:
	docker build -t bufstream/demo-consume -f Dockerfile.consume .

.PHONY: docker-produce-run
docker-produce-run: docker-produce-build # Run the demo producer within Docker. If you have Go installed, you can call produce-run.
	docker run --rm --network=host bufstream/demo-produce

.PHONY: docker-consume-run
docker-consume-run: docker-consume-build # Run the demo consumer within Docker. If you have Go installed, you can call consume-run.
	docker run --rm --network=host bufstream/demo-consume

.PHONY: produce-run
produce-run: # Run the demo producer. Go must be installed.
	go run ./cmd/bufstream-demo-produce

.PHONY: consume-run
consume-run: # Run the demo consumer. Go must be installed.
	go run ./cmd/bufstream-demo-consume

.PHONY: docker-compose-clean
docker-compose-clean: # Cleanup docker compose assets.
	docker compose down --rmi all

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
