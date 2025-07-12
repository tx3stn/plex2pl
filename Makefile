BINARY_NAME=plex2m3u
DIR ?= ./...
PWD ?= $(shell pwd)
VERSION ?= $(shell head -n 1 VERSION)

define ajv-docker
	docker run --rm -v "${PWD}":/repo weibeld/ajv-cli:5.0.0 ajv --spec draft7
endef

.PHONY: build
build:
	@CGO_ENABLED=0 go build -ldflags "-X github.com/tx3stn/plex2m3u/cmd.Version=${VERSION}" -o ${BINARY_NAME}

.PHONY: build-image
build-image:
	@docker build --tag ${BINARY_NAME}:local .

.PHONY: install
install: build
	@sudo cp ./${BINARY_NAME} /usr/local/bin/${BINARY_NAME}

.PHONY: lint
lint:
	@golangci-lint run -v ${DIR}

.PHONY: test
test:
	@CGO_ENABLED=1 go test -race -cover -v ${DIR}

.PHONY: validate-schema
validate-schema:
	@$(ajv-docker) compile -s /repo/.schema/schema.json
