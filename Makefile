.PHONY: all
all:
	$(MAKE) build
	$(MAKE) lint
	$(MAKE) generate

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

.PHONY: dockerbuild
dockerbuild:
	docker build -t bufstream/demo .

.PHONY: buf
buf:
	go install github.com/bufbuild/buf/cmd/buf@latest
