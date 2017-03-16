default: build

# Install dependencies
install:
	go get -u github.com/golang/lint/golint
	go get ./...

# Build backend and frontend components
build:
	go build ./cmd/ons/

build-linux-x64:
	GOOS=linux GOARCH=amd64 go build ./cmd/ons/

build-osx-x64:
	GOOS=darwin GOARCH=amd64 go build ./cmd/ons/

# Build libons C library
libons:
	/bin/bash -c "cd ./libons && make libs"

# Build libons example client
client:
	/bin/bash -c "cd ./libons && make client;"

# Run application
run: build
	./ons

# Test application
test: libons
	go test -p=1 ./...

# Utilities

lint:
	golint ./...

format:
	gofmt -w -s ./

coverage:
	go test -p=1 -cover ./...
	
checks: lint format coverage

.PHONY: build run test lint format coverage libons
