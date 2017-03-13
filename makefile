default: build

# Install dependencies
install:
	go get -u github.com/golang/lint/golint
	go get ./...

# Build backend and frontend components
build:
	go build

# Run application
run: build
	./ons

# Test application
test:
	go test -p=1 ./...

# Utilities

lint:
	golint ./...

format:
	gofmt -w -s ./

coverage:
	go test -p=1 -cover ./...
	
checks: lint format coverage
