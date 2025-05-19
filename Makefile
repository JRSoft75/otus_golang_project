BIN := "./bin/resizer"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)



install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.6

lint: install-lint-deps
	golangci-lint run ./...

building:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/resizer

run: building
	$(BIN)
	#$(BIN) --config ./configs/config.yaml

test:
	go test -v -count=1 -race ./internal/...

version: building
	$(BIN) --version

webtest:
	docker compose up -d resizer-test-nginx