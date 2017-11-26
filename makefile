
IDIR=/usr/local

BINS=owns owns-mapclient
LIBS=cowns/build/libowns.a cowns/build/libowns.so

default: owns

# Install dependencies
deps:
	glide install

tools:
	go get -u github.com/golang/lint/golint
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

all: owns lib client

# Build protocol
protocol: protocol/*.proto
	protoc --go_out=import_path=protocol:lib/ protocol/*.proto
	protoc-c --c_out=cowns/src protocol/*.proto

# Build ons server
owns: protocol
	go build -ldflags -s ./cmd/owns-sim
	go build -ldflags -s ./cmd/owns-mapclient
	go build -ldflags -s ./cmd/owns-eval

build:
	go build ./cmd/owns-sim

build-linux-x64:
	GOOS=linux GOARCH=amd64 go build ./cmd/...

build-osx-x64:
	GOOS=darwin GOARCH=amd64 go build ./cmd/...

# Build libons C library and example client
lib: protocol
	/bin/bash -c "cd ./cowns && make all"

client:
	mkdir -p cowns/build && cd cowns/build && cmake .. && make

# Run application
run: build
	./owns-sim -c examples/chain.yml

test-deps: owns
	./owns-mapclient -c examples/chain.yml -t satellite
	./owns-mapclient -c examples/chain.yml -t terrain
    
# Test application
test: owns lib test-deps
	GODEBUG=cgocheck=0 go test -p=1 -timeout=10s -ldflags -s ./lib/... ./cowns/...

install: owns lib
	go install ./cmd/...
	cd cowns/build && cmake .. && make install; cd ../..

# Utilities

lint:
	golint ./...

format:
	gofmt -w -s ./

coverage:
	go test -p=1 -cover ./...
	
checks: lint format coverage

.PHONY: owns lib run test lint format coverage protocol client
