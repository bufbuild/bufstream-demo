BUFSTREAM_VERSION := 0.3.22

.DEFAULT_GOAL := docker-compose-run

### Run Bufstream, the demo producer, and the demo consumer on your local machine.
#
# Requires Go to be installed. Targets should be run from separate terminals.

BIN := .tmp

.PHONY: bufstream-run
bufstream-run: $(BIN)/bufstream
	./$(BIN)/bufstream serve --config config/bufstream.yaml

.PHONY: produce-run
produce-run: # Run the demo producer. Go must be installed.
	go run ./cmd/bufstream-demo-produce --topic email-updated --group email-verifier

.PHONY: consume-run
consume-run: # Run the demo consumer. Go must be installed.
	go run ./cmd/bufstream-demo-consume --topic email-updated --group email-verifier \
		--csr-url "https://demo.buf.dev/integrations/confluent/bufstream-demo"

### Run Bufstream, the demo producer, the demo consumer, and AKHQ within Docker Compose.
#
# Requires Docker to be installed, but will work out of the box.

.PHONY: docker-compose-run
docker-compose-run: generate-config # Run the demo within docker compose.
	@echo "Running docker compose up..."
	@docker compose up --build --remove-orphans

.PHONY: docker-compose-clean
docker-compose-clean: # Cleanup docker compose assets.
	@echo "Running docker compose down..."
	@docker compose down --volumes --remove-orphans
	@echo "Cleaning up generated config..."
	@rm -rf .tmp

### Run Bufstream, the demo producer, the demo consumer, and AKHQ within Docker Compose.
#
# Requires Docker to be installed. Targets should be run from separate terminals.

.PHONY: docker-bufstream-run
docker-bufstream-run: # Run Bufstream within Docker.
	docker run --rm -p 9092:9092 -v ./config/bufstream.yaml:/bufstream.yaml \
		"bufbuild/bufstream:$(BUFSTREAM_VERSION)" \
			--config /bufstream.yaml

.PHONY: docker-produce-run
docker-produce-run: # Run the demo producer within Docker. If you have Go installed, you can call produce-run.
	docker build -t bufstream/demo-produce -f Dockerfile.produce .
	docker run --rm --network=host bufstream/demo-produce

.PHONY: docker-consume-run
docker-consume-run: # Run the demo consumer within Docker. If you have Go installed, you can call consume-run.
	docker build -t bufstream/demo-consume -f Dockerfile.consume .
	docker run --rm --network=host bufstream/demo-consume

$(BIN)/bufstream: Makefile
	@rm -f $(BIN)/bufstream
	@mkdir -p $(BIN)
	curl -sSL \
		"https://buf.build/dl/bufstream/v$(BUFSTREAM_VERSION)/bufstream-v$(BUFSTREAM_VERSION)-$(shell uname -s)-$(shell uname -m)" \
		-o $(BIN)/bufstream
	chmod +x $(BIN)/bufstream

### Terraform Commands

# Helper function to source .env and export TF_VAR_ prefixed variables
export_tf_vars = \
	if [ -f .env ]; then \
		set -a; . ./.env; set +a; \
		export TF_VAR_gcp_project_id="$${GCP_PROJECT_ID}"; \
		export TF_VAR_gcp_region="$${GCP_REGION}"; \
		export TF_VAR_gcs_bucket_name="$${GCS_BUCKET_NAME}"; \
		export TF_VAR_bq_dataset_name="$${BQ_DATASET_NAME}"; \
		export TF_VAR_bq_connection_id="$${BQ_CONNECTION_ID}"; \
		export TF_VAR_bq_location="$${BQ_LOCATION}"; \
		if [ -n "$${GCP_CREDENTIALS_FILE}" ]; then \
			export TF_VAR_gcp_credentials_file="$${GCP_CREDENTIALS_FILE}"; \
		fi; \
	else \
		@echo ".env file not found. Please create it from .env.example."; exit 1; \
	fi

.PHONY: tf-init
tf-init: # Initialize Terraform in the terraform directory.
	@echo "Initializing Terraform..."
	@cd terraform && $(export_tf_vars) && terraform init

.PHONY: tf-plan
tf-plan: # Generate a Terraform execution plan.
	@echo "Planning Terraform changes..."
	@$(export_tf_vars) && cd terraform && terraform plan

.PHONY: tf-apply
tf-apply: # Apply the Terraform changes. Requires interactive approval.
	@echo "Applying Terraform changes..."
	@$(export_tf_vars) && cd terraform && terraform apply

.PHONY: tf-apply-auto
tf-apply-auto: # Apply the Terraform changes automatically (no interactive approval). Use with caution.
	@echo "Applying Terraform changes automatically..."
	@$(export_tf_vars) && cd terraform && terraform apply -auto-approve

.PHONY: tf-destroy
tf-destroy: # Destroy the Terraform-managed infrastructure. Requires interactive approval.
	@echo "Destroying Terraform infrastructure..."
	@$(export_tf_vars) && cd terraform && terraform destroy

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

# Ensure the .tmp directory exists
.PHONY: prepare-tmp
prepare-tmp:
	@mkdir -p .tmp

# Generate the bufstream config with environment variables substituted
.PHONY: generate-config
generate-config: prepare-tmp .env
	@echo "Generating bufstream config with substituted variables..."
	@# Source .env, export all variables (-a), then run envsubst
	@set -a && . ./.env && set +a && envsubst < config/bufstream.yaml > .tmp/bufstream.yaml
