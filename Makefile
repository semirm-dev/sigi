BINARY  := sigi
MODULE  := github.com/semirm-dev/sigi
BUILD   := .build
BIN     := bin
VERSION := $(shell cat VERSION 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X $(MODULE)/internal/keepalive.Version=$(VERSION)"
COVERFILE := coverprofile

.PHONY: help build install run test test-cover lint order release tidy clean

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  build       Build $(BUILD)/$(BINARY) for this machine."
	@echo "  install     Install sigi into GOBIN."
	@echo "  run         Run sigi from source."
	@echo "  test        Run the test suite with race detection."
	@echo "  test-cover  Run tests and open the coverage report."
	@echo "  lint        gofmt, goimports, declaration order and go vet."
	@echo "  order       Reorder declarations to the house order."
	@echo "  release     Build a release binary for this machine (see CI for the rest)."
	@echo "  tidy        go mod tidy."
	@echo "  clean       Remove build output."

build: ## Build the sigi binary
	@mkdir -p $(BUILD)
	go build $(LDFLAGS) -o $(BUILD)/$(BINARY) ./cmd/sigi

install: ## Install sigi into GOBIN
	go install $(LDFLAGS) ./cmd/sigi

run: ## Run sigi from source
	go run ./cmd/sigi

test: ## Run tests with race detection
	go test ./... -race -count=1

test-cover: ## Run tests and open the coverage report
	go test ./... -coverprofile=$(COVERFILE)
	go tool cover -html=$(COVERFILE) && go tool cover -func $(COVERFILE) && unlink $(COVERFILE)

lint: ## Check formatting, import grouping, declaration order and run vet
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || (echo "gofmt found issues" && exit 1)
	@test -z "$$(go tool goimports -local $(MODULE) -l . | tee /dev/stderr)" || (echo "goimports found issues" && exit 1)
	@python3 scripts/order.py --check $$(find internal cmd -name '*.go') \
		|| (echo "run 'make order' to fix" && exit 1)
	go vet ./...

release: ## Build a release binary for THIS machine only
	@mkdir -p $(BIN)
	@os=$$(go env GOOS); arch=$$(go env GOARCH); ext=""; \
		[ "$$os" = "windows" ] && ext=".exe"; \
		out=$(BIN)/$(BINARY)-$$os-$$arch$$ext; \
		CGO_ENABLED=1 go build -trimpath -buildvcs=false \
			-ldflags "-s -w -X $(MODULE)/internal/keepalive.Version=$(VERSION)" \
			-o $$out ./cmd/sigi; \
		echo "  $$out"
	@echo "  sigi cannot cross-compile: it drives input devices through cgo, so each"
	@echo "  platform must be built on that platform. .github/workflows/release.yml"
	@echo "  does all of them, one runner each."

order: ## Reorder declarations to the house order
	@python3 scripts/order.py $$(find internal cmd -name '*.go')
	@gofmt -w .

tidy: ## Tidy the module graph
	go mod tidy

clean: ## Remove build output
	rm -rf $(BUILD) $(BIN) build
