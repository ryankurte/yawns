default: ons

# Install dependencies
install: lib
	go get -t ./...

install-tools:
	go get -u github.com/golang/lint/golint
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

all: ons lib client

# Build protocol
protocol: protocol/*.proto
	protoc --go_out=import_path=protocol:lib/ protocol/*.proto
	protoc-c --c_out=. protocol/*.proto

# Build ons server
ons: protocol
	go build -ldflags -s ./cmd/ons/

build-linux-x64:
	GOOS=linux GOARCH=amd64 go build ./cmd/ons/

build-osx-x64:
	GOOS=darwin GOARCH=amd64 go build ./cmd/ons/

# Build libons C library
lib: protocol
	/bin/bash -c "cd ./cons && make libs"

# Build libons example client
client: lib
	/bin/bash -c "cd ./cons && make client"

# Run application
run: build
	./ons

# Test application
test: ons lib
	go test -p=1 -timeout=10s -ldflags -s ./...

# Utilities

lint:
	golint ./...

format:
	gofmt -w -s ./

coverage:
	go test -p=1 -cover ./...
	
checks: lint format coverage

.PHONY: ons lib run test lint format coverage
