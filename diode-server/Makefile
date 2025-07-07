UTILITIES        = authmanager
BUILD_UTILITIES  = $(addprefix build-,$(UTILITIES))
SERVICES         := $(filter-out $(UTILITIES),$(shell find ./cmd/* -type d -exec basename {} \;))
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

.PHONY: install-dev-tools
install-dev-tools:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@go install github.com/vektra/mockery/v2@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/mfridman/tparse@latest
	@go install github.com/pressly/goose/v3/cmd/goose@v3.24.3

.PHONY: deps
deps:
	@go mod tidy

.PHONY: lint
lint:
	@golangci-lint run ./... --config ../.github/golangci.yaml

.PHONY: fix-lint
fix-lint:
	@golangci-lint run ./... --config ../.github/golangci.yaml --fix

.PHONY: test
test:
	@go test -race ./...

.PHONY: test-coverage
test-coverage:
	@mkdir -p .coverage
	@go test -race -cover -json -coverprofile=.coverage/cover.out.tmp ./... | grep -Ev "gen|cmd|mocks|sentry" | tparse -format=markdown > .coverage/test-report.md
	@cat .coverage/cover.out.tmp | grep -Ev "gen|cmd|mocks|sentry" > .coverage/cover.out
	@go tool cover -func=.coverage/cover.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}' > .coverage/coverage.txt

.PHONY: test-integration
test-integration:
	@go test -tags integration_test -race ./...

.PHONY: $(BUILD_SERVICES)
$(BUILD_SERVICES):
	@$(eval SVC=$(subst build-,,$@))
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
	go build -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(SVC) ./cmd/$(SVC)

.PHONY: $(BUILD_UTILITIES)
$(BUILD_UTILITIES):
	@$(eval UTIL=$(subst build-,,$@))
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
	go build -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(UTIL) ./cmd/$(UTIL)

.PHONY: build-all
build-all: $(BUILD_SERVICES) $(BUILD_UTILITIES)

.PHONY: $(DOCKER_SERVICES)
$(DOCKER_SERVICES): $(BUILD_UTILITIES)
	@$(eval SVC=$(subst docker-,,$@))

	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(MAKE) build-$(SVC)

	@$(eval DOCKERFILE=Dockerfile.$(SVC))

	docker build \
	  --no-cache \
	  --build-arg SVC=$(SVC) \
	  --build-arg GOOS=$(GOOS) \
	  --build-arg GOARCH=$(GOARCH) \
	  --build-arg GOARM=$(GOARM) \
  	  --tag=$(ORG_NAME)/$(REPO_NAME)-$(SVC):$(DIODE_VERSION)-$(COMMIT_SHA) \
  	  -f docker/$(DOCKERFILE) .

.PHONY: docker-all
docker-all: $(DOCKER_SERVICES)

.PHONY: docker-compose-up
docker-compose-up:
	$(DOCKER_COMPOSE) --env-file docker/sample.env -f docker/docker-compose.yaml up -d --build

.PHONY: docker-compose-down
docker-compose-down:
	$(DOCKER_COMPOSE) --env-file docker/sample.env -f docker/docker-compose.yaml down --remove-orphans

.PHONY: docker-compose-dev-up
docker-compose-dev-up: docker-all
	@DIODE_VERSION=$(DIODE_VERSION) COMMIT_SHA=$(COMMIT_SHA) DIODE_TAG=$(DIODE_VERSION)-$(COMMIT_SHA) PROJECT_NAME=diode-dev \
	$(DOCKER_COMPOSE) --env-file docker/.env -f docker/docker-compose.yaml -f docker/docker-compose.dev.yaml up -d --build

.PHONY: docker-compose-dev-logs
docker-compose-dev-logs:
	@DIODE_VERSION=$(DIODE_VERSION) COMMIT_SHA=$(COMMIT_SHA) DIODE_TAG=$(DIODE_VERSION)-$(COMMIT_SHA) PROJECT_NAME=diode-dev \
	$(DOCKER_COMPOSE) --env-file docker/.env -f docker/docker-compose.yaml -f docker/docker-compose.dev.yaml logs

.PHONY: docker-compose-dev-down
docker-compose-dev-down:
	@DIODE_VERSION=$(DIODE_VERSION) COMMIT_SHA=$(COMMIT_SHA) DIODE_TAG=$(DIODE_VERSION)-$(COMMIT_SHA) PROJECT_NAME=diode-dev \
	$(DOCKER_COMPOSE) --env-file docker/.env -f docker/docker-compose.yaml -f docker/docker-compose.dev.yaml down --remove-orphans

.PHONY: docker-compose-dev-authmanager
docker-compose-dev-authmanager:
	@DIODE_VERSION=$(DIODE_VERSION) COMMIT_SHA=$(COMMIT_SHA) DIODE_TAG=$(DIODE_VERSION)-$(COMMIT_SHA) PROJECT_NAME=diode-dev \
	$(DOCKER_COMPOSE) --env-file docker/.env -f docker/docker-compose.yaml -f docker/docker-compose.dev.yaml run --rm --no-deps diode-auth authmanager $(authmanager_command)

.PHONY: configure-dev
configure-dev:
	./docker/scripts/configure-dev.sh $(NETBOX_HOST)

.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)/*

.PHONY: gen-mocks
gen-mocks:
	mockery

.PHONY: gen-dbstore
gen-dbstore:
	sqlc generate

.PHONY: create-migration
# Call using `make create-migration name=you_migration_name`
create-migration:
	GOOSE_MIGRATION_DIR=dbstore/postgres/migrations/ goose create $(name) sql
