SHELL := /usr/bin/env bash -o pipefail

.DEFAULT_GOAL := docker-compose-run

BUFSTREAM_VERSION := $(shell grep 'bufbuild/bufstream:' docker-compose.yaml | sed -e 's|.*:||')

### Run Bufstream, the demo producer, and the demo consumer on your local machine.
#
# Requires Go to be installed. Targets should be run from separate terminals.

BIN := .tmp

.PHONY: bufstream-run
bufstream-run: $(BIN)/bufstream
	./$(BIN)/bufstream serve --config config/bufstream.yaml --schema .

.PHONY: produce-run
produce-run: # Run the demo producer. Go must be installed.
	go run ./cmd/bufstream-demo-produce --topic orders

.PHONY: consume-run
consume-run: # Run the demo consumer. Go must be installed.
	go run ./cmd/bufstream-demo-consume --topic orders --group order-verifier

.PHONY: consume-dlq-run
consume-dlq-run: # Run the demo DLQ consumer. Go must be installed.
	go run ./cmd/bufstream-demo-consume-dlq --topic orders.dlq --group order-dlq-monitor

.PHONY: bufstream-create-topic
bufstream-create-topic: # Create the "orders" topic.
	./$(BIN)/bufstream kafka topic create orders --partitions 1

.PHONY: bufstream-set-message
bufstream-set-message: # Associate the Cart message with the topic.
	./$(BIN)/bufstream kafka config topic set --topic orders --name buf.registry.value.schema.message --value bufstream.demo.v1.Cart

.PHONY: bufstream-set-message
bufstream-set-message: # Set validation mode to "reject".
	./$(BIN)/bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value reject

.PHONY: bufstream-set-dlq
bufstream-set-dlq: # Set validation mode to "dlq".
	./$(BIN)/bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value dlq

### Run Bufstream, the demo producer, the demo consumer, and AKHQ within Docker Compose.
#
# Requires Docker to be installed, but will work out of the box.

.PHONY: docker-compose-run
docker-compose-run: # Run the demo within docker compose.
	docker compose up --build

.PHONY: docker-compose-clean
docker-compose-clean: # Cleanup docker compose assets.
	docker compose down --rmi all

### Run Bufstream, the demo producer, the demo consumer, and AKHQ within Docker Compose.
#
# Requires Docker to be installed. Targets should be run from separate terminals.

.PHONY: docker-bufstream-run
docker-bufstream-run: # Run Bufstream within Docker.
	docker run --rm -p 9092:9092 -v ./config/bufstream.yaml:/bufstream.yaml \
		-v .:/buf-workspace \
		"bufbuild/bufstream:$(BUFSTREAM_VERSION)" serve \
			--config /bufstream.yaml \
			--schema /buf-workspace

.PHONY: docker-produce-run
docker-produce-run: # Run the demo producer within Docker. If you have Go installed, you can call produce-run.
	docker build -t bufstream/demo-produce -f Dockerfile.produce .
	docker run --rm --network=host bufstream/demo-produce --topic orders

.PHONY: docker-consume-run
docker-consume-run: # Run the demo consumer within Docker. If you have Go installed, you can call consume-run.
	docker build -t bufstream/demo-consume -f Dockerfile.consume .
	docker run --rm --network=host bufstream/demo-consume --topic orders

.PHONY: docker-consume-dlq-run
docker-consume-dlq-run: # Run the demo consumer within Docker. If you have Go installed, you can call consume-run.
	docker build -t bufstream/demo-consume-dlq -f Dockerfile.consume-dlq .
	docker run --rm --network=host bufstream/demo-consume-dlq --topic orders.dlq \
		--group order-dlq-monitor

$(BIN)/bufstream: Makefile
	@rm -f $(BIN)/bufstream
	@mkdir -p $(BIN)
	curl -sSL \
		"https://buf.build/dl/bufstream/v$(BUFSTREAM_VERSION)/bufstream-v$(BUFSTREAM_VERSION)-$(shell uname -s)-$(shell uname -m)" \
		-o $(BIN)/bufstream
	chmod +x $(BIN)/bufstream

### Development commands

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
