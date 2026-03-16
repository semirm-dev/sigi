.PHONY: help build run test test-cover

COVERFILE=coverprofile

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build binary for macOS Apple Silicon (darwin/arm64)"
	@echo "  run          Run the application"
	@echo "  test         Run tests"
	@echo "  test-cover   Run tests with coverage report"

build:
	GOOS=darwin GOARCH=arm64 go build -o build/sigi-darwin-arm64 ./cmd/main.go

run:
	go run cmd/main.go

test:
	go test -v ./...
test-cover:
	go test -v ./... -coverprofile=${COVERFILE}
	go tool cover -html=${COVERFILE} && go tool cover -func ${COVERFILE} && unlink ${COVERFILE}
