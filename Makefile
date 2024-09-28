.PHONY: all
all:
	$(MAKE) generate
	$(MAKE) lint

.PHONY: generate
generate: buf
	buf generate
	buf format -w

.PHONY: lint
lint: buf
	buf lint

.PHONY: upgrade
upgrade:
	buf dep update
	go get -u -t ./...
	go mod tidy -v

.PHONY: buf
buf:
	go install github.com/bufbuild/buf/cmd/buf@latest
