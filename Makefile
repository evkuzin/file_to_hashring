# load env variables from .env
ENV_PATH ?= ./.env
ifneq ($(wildcard $(ENV_PATH)),)
    include .env
    export
endif

# service
SERVICE = file_to_hashring
# current version
DOCKER_TAG = $(shell git rev-parse --short HEAD)
# docker registry url
DOCKER_URL = evkuzin

FILE = test_photo.jpeg
UPLOAD_DIR = ./misc/uploads
DOWNLOAD_DIR = ./misc/downloads


# Build commands =======================================================================================================
echo:
	echo $(DOCKER_TAG)

dep:
	go env -w GO111MODULE=on
	go env -w GOPRIVATE=github.com/evkuzin/*
	go mod tidy

mock: # generate mocks
	@rm -R ./mocks 2> /dev/null; \
	mockery --all

build:  dep  ## builds the main
	mkdir -p bin
	go build -o bin/ cmd/main.go

artifacts: mock build ## builds and generates all artifacts

run: ## run the service
	./bin/main

# Tests commands =======================================================================================================

test: ## run the tests
	@echo "running tests (skipping integration)"
	go test ./...

test-with-coverage: ## run the tests with coverage
	@echo "running tests with coverage file creation (skipping integration)"
	go test -coverprofile .testCoverage.txt -v ./...

test-integration: ## run the integration tests
	@echo "running integration tests"
	go test -tags integration ./...

# Docker commands =======================================================================================================

docker-build: ## Build the docker images for all services (build inside)
	@echo Building images
	docker build . -f ./Dockerfile -t $(DOCKER_URL)/$(SERVICE):$(DOCKER_TAG) --no-cache

docker-build-test: ## Build the docker images for all services (build inside)
	@echo Building images
	docker build . -f ./Dockerfile-test -t $(DOCKER_URL)/$(SERVICE):$(DOCKER_TAG)-test --no-cache

docker-push: docker-build ## Build and push docker images to the repository
	@echo Pushing images
	docker push $(DOCKER_URL)/$(SERVICE):$(DOCKER_TAG)

docker-push-test: docker-build-test ## Build and push docker images to the repository
	@echo Pushing images
	docker push $(DOCKER_URL)/$(SERVICE):$(DOCKER_TAG)-test

docker-run:
	@echo Running container
	docker run $(DOCKER_URL)/$(SERVICE):$(DOCKER_TAG)

docker-compose-up: docker-build
	@DOCKER_TAG=$(DOCKER_TAG) docker-compose up -d

clean:
	@DOCKER_TAG=$(DOCKER_TAG) docker-compose stop && docker-compose rm -f && rm -rf ./bin && rm -f .$(DOWNLOAD_DIR)/*

upload:
	curl -v -F upload=@$(UPLOAD_DIR)/$(FILE) 127.0.0.1:8080/upload

download:
	curl "127.0.0.1:8080/download?filename=test_photo.jpeg" -o $(DOWNLOAD_DIR)/$(FILE)

check:
	@sha256sum $(UPLOAD_DIR)/$(FILE) || echo "no sha256sum installed, check the other way"
	@sha256sum $(DOWNLOAD_DIR)/$(FILE) || echo "no sha256sum installed, check the other way"
