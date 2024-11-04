VERSION ?= $(shell date +"%Y%m%d")
OUTPUT_DIR := _output

.PHONY: start-ddn
start-ddn:
	HASURA_DDN_PAT=$$(ddn auth print-pat) docker compose -f compose.yaml -f tests/engine/compose.yaml --env-file tests/engine/.env up --build -d

.PHONY: stop-ddn
stop-ddn:

.PHONY: format
format:
	gofmt -w -s .

.PHONY: test
test:
	go test -v -race -timeout 3m ./...

# Install golangci-lint tool to run lint locally
# https://golangci-lint.run/usage/install
.PHONY: lint
lint:
	golangci-lint run

# clean the output directory
.PHONY: clean
clean:
	rm -rf "$(OUTPUT_DIR)"

.PHONY: build-configuration
build-configuration:
	CGO_ENABLED=0 go build -o _output/ndc-loki ./configuration
	
.PHONY: build-jsonschema
build-jsonschema:
	cd jsonschema && go run .

# build the configuration tool for all given platform/arch
.PHONY: ci-build-configuration
ci-build-configuration: clean
	export CGO_ENABLED=0 && \
	go get github.com/mitchellh/gox && \
	go run github.com/mitchellh/gox -ldflags '-X github.com/hasura/ndc-loki/configuration/version.BuildVersion=$(VERSION) -s -w -extldflags "-static"' \
		-osarch="linux/amd64 linux/arm64 darwin/amd64 windows/amd64 darwin/arm64" \
		-output="$(OUTPUT_DIR)/ndc-loki-{{.OS}}-{{.Arch}}" \
		./configuration

.PHONY: build-supergraph-test
build-supergraph-test:
	docker compose up -d --build ndc-loki
	cd tests/engine && \
		ddn connector-link update loki --add-all-resources --subgraph ./app/subgraph.yaml && \
		ddn supergraph build local
	make start-ddn

.PHONY: generate-api-types
generate-api-types:
	hasura-ndc-go update --directories ./connector/api,./connector/client --connector-dir ./connector/api --schema-format go --style snake-case --type-only

.PHONY: generate-test-config
generate-test-config:
	CONNECTION_URL=http://localhost:3100 \
		go run ./configuration update -d ./tests/configuration --log-level debug