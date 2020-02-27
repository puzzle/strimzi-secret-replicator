.PHONY: setup-lint lint test build all docker-build docker-push docker-tag

GOLANGCILINT_VERSION ?= 1.23.6
VERSION              ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT           ?= $(shell git rev-parse HEAD)
LDFLAGS              ?= -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -w -s
GOPATH               ?= $(shell go env GOPATH)
IMAGE                ?= docker.io/puzzle/strimzi-secret-replicator

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)"

test:
	go test -v ./...

lint:
	golangci-lint run ./...

all: build lint test

docker-build:
	docker build -t $(IMAGE) .

docker-tag:
	docker tag $(IMAGE) $(IMAGE):$(VERSION)

docker-push:
	docker push $(IMAGE):$(VERSION)

setup-lint:
	mkdir -p $(GOPATH)/bin
	curl -sSfL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCILINT_VERSION)/golangci-lint-$(GOLANGCILINT_VERSION)-linux-amd64.tar.gz | \
		tar -xzv --wildcards --strip-components=1 -C $(GOPATH)/bin '*/golangci-lint'
