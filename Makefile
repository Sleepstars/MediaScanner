.PHONY: build clean test run docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=mediascanner
BINARY_UNIX=$(BINARY_NAME)_unix

# Build parameters
BUILD_DIR=build
MAIN_PATH=cmd/mediascanner/main.go

# Docker parameters
DOCKER_IMAGE=mediascanner
DOCKER_TAG=latest

all: test build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v $(MAIN_PATH)

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	$(GOTEST) -v ./...

run:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME)

deps:
	$(GOMOD) tidy

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) -v $(MAIN_PATH)

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

help:
	@echo "make - Run tests and build binary"
	@echo "make build - Build binary"
	@echo "make clean - Remove build artifacts"
	@echo "make test - Run tests"
	@echo "make run - Build and run locally"
	@echo "make deps - Update dependencies"
	@echo "make build-linux - Cross-compile for Linux"
	@echo "make docker-build - Build Docker image"
	@echo "make docker-run - Run Docker container"
	@echo "make docker-compose-up - Start services with docker-compose"
	@echo "make docker-compose-down - Stop services with docker-compose"
