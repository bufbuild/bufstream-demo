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
use-reject-mode: # Reject invalid messages.
	./$(BIN)/bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value reject

.PHONY: use-dlq-mode
use-dlq-mode: # Send invalid messages to the DLQ topic.
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
docker-compose-use-reject-mode: # Reject invalid messages.
	docker exec bufstream bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value reject

.PHONY: docker-compose-use-dlq-mode
docker-compose-use-dlq-mode: # Send invalid messages to the DLQ topic.
	docker exec bufstream bufstream kafka config topic set --topic orders --name bufstream.validate.mode --value dlq

.PHONY: docker-compose-clean
docker-compose-clean: # Cleanup docker compose assets.
	docker compose down --rmi all

### Run Bufstream and all services for Iceberg and Spark within Docker Compose.
#
# Requires Docker to be installed, but will work out of the box.

.PHONY: iceberg-run
iceberg-run: # Run Bufstream and other services needed for Iceberg.
	docker compose --file ./iceberg/docker-compose.yaml up

.PHONY: iceberg-config
iceberg-config: # Configure the orders topic for Iceberg.
	docker exec bufstream bufstream kafka config topic set --topic orders --name bufstream.archive.iceberg.catalog --value local-rest-catalog_2  --name bufstream.archive.iceberg.table --value bufstream.orders_2
	docker exec bufstream bufstream kafka config topic set --topic orders --name bufstream.archive.iceberg.table --value bufstream.orders
	docker exec bufstream bufstream kafka config topic set --topic orders --name bufstream.archive.kind --value iceberg

.PHONY: iceberg-produce
iceberg-produce: produce-run

.PHONY: iceberg-clean-topics
iceberg-clean-topics: # Run Bufstream's "clean topics" command.
	docker exec bufstream bufstream admin clean topics
	@echo "Open http://localhost:8888/notebooks/notebooks/bufstream-quickstart.ipynb to run queries."

.PHONY: iceberg-clean
iceberg-clean: # Cleanup Docker Compose assets.
	# --rmi all
	docker compose --file ./iceberg/docker-compose.yaml down
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
