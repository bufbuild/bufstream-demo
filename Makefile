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
	go run ./cmd/bufstream-demo-produce --topic orders \
		--topic-config buf.registry.value.schema.message=bufstream.demo.v1.Cart

.PHONY: consume-run
consume-run: # Run the demo consumer. Go must be installed.
	go run ./cmd/bufstream-demo-consume --topic orders --group order-verifier

.PHONY: use-reject-mode
use-reject-mode: # Reject invalid messages
	./$(BIN)/bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value reject

.PHONY: use-dlq-mode
use-dlq-mode: # Send invalid messages to the DLQ topic
	./$(BIN)/bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value dlq

.PHONY: consume-dlq-run
consume-dlq-run: # Run the demo DLQ consumer. Go must be installed.
	go run ./cmd/bufstream-demo-consume-dlq --topic orders.dlq --group order-dlq-monitor

### Run Bufstream, the demo producer, the demo consumer, and AKHQ within Docker Compose.
#
# Requires Docker to be installed, but will work out of the box.

.PHONY: docker-compose-run
docker-compose-run: # Run the demo within Docker Compose.
	docker compose up --build

.PHONY: docker-compose-use-reject-mode
docker-compose-use-reject-mode: # Reject invalid messages
	docker exec bufstream bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value reject

.PHONY: docker-compose-use-dlq-mode
docker-compose-use-dlq-mode: # Send invalid messages to the DLQ topic
	docker exec bufstream bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value dlq

.PHONY: docker-compose-clean
docker-compose-clean: # Cleanup docker compose assets.
	docker compose down --rmi all

### Run Iceberg-capable Bufstream within Docker Compose.
#
# Requires Docker to be installed, but will work out of the box.

.PHONY: iceberg-run
iceberg-run: # Run the Iceberg demo within Docker Compose.
	docker compose --file ./iceberg/docker-compose.yaml up --detach
	@echo "Waiting 10s for sample records to be produced..."
	@sleep 10s
	docker exec bufstream bufstream admin clean topics
	@echo "Order data created. Open http://localhost:8888/notebooks/notebooks/bufstream-quickstart.ipynb to run queries."

.PHONY: iceberg-clean
iceberg-clean: # Cleanup docker compose assets.
	docker compose --file ./iceberg/docker-compose.yaml down --rmi all
	rm -rf ./iceberg/data


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
