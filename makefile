
IDIR=/usr/local

BINS=yawns yawns-mapclient
LIBS=cyawns/build/libyawns.a cyawns/build/libyawns.so

default: yawns

# Install dependencies
deps:
	go get -u github.com/golang/lint/golint
	go get -u github.com/Masterminds/glide
	# Glide manages go deps
	glide install
	# Protobuf binaries must match library version
	go install ./vendor/github.com/golang/protobuf/...

all: yawns lib client

# Build protocol
protocol: protocol/*.proto
	protoc --go_out=import_path=protocol:lib/ protocol/*.proto
	protoc-c --c_out=cyawns/src protocol/*.proto

# Build ons server
yawns: protocol
	go build -ldflags -s ./cmd/yawns-sim
	go build -ldflags -s ./cmd/yawns-mapclient
	go build -ldflags -s ./cmd/yawns-eval

build:
	go build ./cmd/yawns-sim

build-linux-x64:
	GOOS=linux GOARCH=amd64 go build ./cmd/...

build-osx-x64:
	GOOS=darwin GOARCH=amd64 go build ./cmd/...

# Build libyawns C library and example client
lib: protocol
	/bin/bash -c "cd ./cyawns && make all"

client:
	mkdir -p cyawns/build && cd cyawns/build && cmake .. && make

# Run application
run: build
	./yawns-sim -c examples/chain.yml

test-deps: yawns
	./yawns-mapclient -c examples/chain.yml -t satellite
	./yawns-mapclient -c examples/chain.yml -t terrain
    
# Test application
test: yawns lib client
	GODEBUG=cgocheck=0 go test -p=1 -timeout=10s -ldflags -s ./lib/... ./cyawns/...

install: yawns lib
	go install ./cmd/...
	cd cyawns/build && cmake .. && make install; cd ../..

# Utilities

lint:
	golint ./...

format:
	gofmt -w -s ./

coverage:
	go test -p=1 -cover ./...
	
checks: lint format coverage

.PHONY: yawns lib run test lint format coverage protocol client
