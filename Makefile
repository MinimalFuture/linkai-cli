VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X github.com/MinimalFuture/linkai-cli/cmd.version=$(VERSION) \
           -X github.com/MinimalFuture/linkai-cli/cmd.buildDate=$(DATE)

.PHONY: build install test lint tidy clean

build:
	go build -ldflags '$(LDFLAGS)' -o linkai .

install:
	go install -ldflags '$(LDFLAGS)' .

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

clean:
	rm -f linkai
