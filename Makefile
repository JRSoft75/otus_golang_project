BIN := "./bin/resizer"

version: build
	$(BIN) version

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.6

lint: install-lint-deps
	golangci-lint run ./...

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/resizer

run: build
	$(BIN) --config ./configs/config.yaml