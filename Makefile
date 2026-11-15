NAME    := orchard
BINARY  := terraform-provider-$(NAME)
VERSION ?= 0.0.2
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)

# Local plugin mirror directory — installing here lets `terraform` pick up
# the provider without publishing (see README "Local development").
PLUGIN_DIR := $(HOME)/.local/share/terraform/plugins/registry.terraform.io/vladkanatov/$(NAME)/$(VERSION)/$(OS_ARCH)

default: fmt vet build test

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY) .

.PHONY: install
install: build
	mkdir -p $(PLUGIN_DIR)
	cp bin/$(BINARY) $(PLUGIN_DIR)/

.PHONY: test
test:
	go test -v -timeout=30s ./...

.PHONY: testacc
testacc:
	TF_ACC=1 ORCHARD_BASE_URL=$(ORCHARD_BASE_URL) ORCHARD_TOKEN=$(ORCHARD_TOKEN) \
		go test -v -timeout=30m ./... -run TestAcc

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: fmtcheck
fmtcheck:
	@test -z "$$(gofmt -l .)" || (echo "gofmt issues, run 'make fmt'"; gofmt -l .; exit 1)

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate

.PHONY: clean
clean:
	rm -rf bin/ dist/
