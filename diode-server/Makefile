.PHONY: deps lint test build-all docker-all docker-compose-up docker-compose-down clean reconciler-proto-gen

SERVICES         := $(shell find ./cmd/* -type d -exec basename {} \;)
BUILD_SERVICES   = $(addprefix build-,$(SERVICES))
DOCKER_SERVICES  = $(addprefix docker-,$(SERVICES))
BUILD_DIR        ?= ./build
LD_FLAGS         ?= -w -s
CGO_ENABLED      ?= 0
GOPATH           ?= $(shell go env GOPATH)
GOARCH           ?= $(shell go env GOARCH)
GOOS             = linux
ORG_NAME         = netboxlabs
REPO_NAME        = diode
ROOT_DIR         := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
VERSION_DIR      = $(ROOT_DIR)/version
DIODE_VERSION    := $(shell cat $(VERSION_DIR)/BUILD_VERSION.txt)
COMMIT_SHA       := $(shell git rev-parse --short HEAD)
VALIDATE_PKG_DIR := $(shell find $(GOPATH)/pkg/mod/github.com/envoyproxy/protoc-gen-validate@* -type d -prune | sort | tail -1)

ifneq ($(shell docker compose version 2>/dev/null),)
  DOCKER_COMPOSE := docker compose
else
  DOCKER_COMPOSE := docker-compose
endif

deps:
	@go mod tidy

lint:
	@golangci-lint run ./... --config ../.github/golangci.yaml

test:
	@go test -race ./...

.PHONY: $(BUILD_SERVICES)
$(BUILD_SERVICES):
	@$(eval SVC=$(subst build-,,$@))
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM)
	@go build -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(SVC) ./cmd/$(SVC)

build-all: $(BUILD_SERVICES)

.PHONY: $(DOCKER_SERVICES)
$(DOCKER_SERVICES):
	@$(eval SVC=$(subst docker-,,$@))

	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(MAKE) build-$(SVC)

	docker build \
	  --no-cache \
	  --build-arg SVC=$(SVC) \
	  --build-arg GOOS=$(GOOS) \
	  --build-arg GOARCH=$(GOARCH) \
	  --build-arg GOARM=$(GOARM) \
  	  --tag=$(ORG_NAME)/$(REPO_NAME)-$(SVC):$(DIODE_VERSION)-$(COMMIT_SHA) \
  	  -f docker/Dockerfile .

docker-all: $(DOCKER_SERVICES)

docker-compose-up: docker-all
	@DIODE_VERSION=$(DIODE_VERSION) COMMIT_SHA=$(COMMIT_SHA) \
	$(DOCKER_COMPOSE) -f docker/docker-compose.yaml -f docker/docker-compose.netbox.yaml up -d --build

docker-compose-down:
	@DIODE_VERSION=$(DIODE_VERSION) COMMIT_SHA=$(COMMIT_SHA) \
	$(DOCKER_COMPOSE) -f docker/docker-compose.yaml -f docker/docker-compose.netbox.yaml down

clean:
	@rm -rf $(BUILD_DIR)/*

reconciler-proto-gen:
	@cd proto/ && buf format -w && buf generate --template buf.gen.yaml
