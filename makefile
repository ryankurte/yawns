default: build

# Install dependencies
install:
	go get -u github.com/golang/lint/golint
	go get ./...

# Build backend and frontend components
build:
	go build

# Build libons C library
libons:
	@cd ./libons && make libs; cd ..

# Build libons example client
client:
	@cd ./libons && make client; cd ..

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

.PHONY: build run test lint format coverage
