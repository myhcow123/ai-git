.PHONY: build clean test install run

BINARY_NAME=ai-git
MAIN_PATH=.

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

clean:
	go clean
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

install: build
	go install

run: build
	./$(BINARY_NAME)

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run

.PHONY: all
all: deps fmt test build
