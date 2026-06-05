BINARY_NAME=plex2pl
DIR ?= ./...
PWD ?= $(shell pwd)
VERSION ?= $(shell head -n 1 VERSION)

define ajv-docker
	docker run --rm -v "${PWD}":/repo weibeld/ajv-cli:5.0.0 ajv --spec draft7
endef

.PHONY: build
build:
	@CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/tx3stn/plex2pl/cmd.Version=${VERSION}" -o ${BINARY_NAME}

.PHONY: build-image
build-image:
	@docker --debug build --tag ${BINARY_NAME}:local .

.PHONY: generate-mocks
generate-mocks:
	@docker run --rm -v "${PWD}":/src -w /src vektra/mockery:3

.PHONY: install
install: build
	@sudo cp ./${BINARY_NAME} /usr/local/bin/${BINARY_NAME}

.PHONY: lint
lint:
	@golangci-lint fmt ${DIR}
	@golangci-lint run --fix -v ${DIR}

.PHONY: lint-schema
lint-schema:
	@$(ajv-docker) compile -s /repo/.schema/schema.json
	@$(ajv-docker) validate -s /repo/.schema/schema.json -d /repo/.schema/example.json

.PHONY: test
test:
	@CGO_ENABLED=1 go test -race -cover ${DIR}

.PHONY: testsum
testsum:
	@CGO_ENABLED=1 gotestsum --format-hide-empty-pkg --format pkgname-and-test-fails -- -race -cover ${DIR}

