.PHONY: all build test clean install

BINARY_NAME=r
BIN_DIR=bin

all: test build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/r

install:
	go install ./cmd/r

test:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR)
