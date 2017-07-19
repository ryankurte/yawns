default: owns

# Install dependencies
install: install-deps lib

install-deps:
	glide install

install-tools:
	go get -u github.com/golang/lint/golint
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

all: owns lib client

# Build protocol
protocol: protocol/*.proto
	protoc --go_out=import_path=protocol:lib/ protocol/*.proto
	protoc-c --c_out=cons/src protocol/*.proto

# Build ons server
owns: protocol
	go build -ldflags -s ./cmd/owns/

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
	./owns

# Test application
test: owns lib
	GODEBUG=cgocheck=0 go test -p=1 -timeout=10s -ldflags -s ./lib/... ./cons/...

# Utilities

lint:
	golint ./...

format:
	gofmt -w -s ./

coverage:
	go test -p=1 -cover ./...
	
checks: lint format coverage

.PHONY: owns lib run test lint format coverage protocol
