.PHONY: test lint lint-fix build build-kubectl-x build-kubernetes-mcp build-kustomizelite build-celery fmt vet tidy help

# Default target
help:
	@echo "Available targets:"
	@echo "  test                 - Run all tests"
	@echo "  lint                 - Run golangci-lint"
	@echo "  lint-fix             - Run golangci-lint with auto-fix"
	@echo "  build                - Build all binaries"
	@echo "  build-kubectl-x      - Build the kubectl-x binary"
	@echo "  build-kubernetes-mcp - Build the kubernetes-mcp binary"
	@echo "  build-kustomizelite  - Build the kustomizelite binary"
	@echo "  build-celery         - Build the celery binary"
	@echo "  fmt                  - Format Go code"
	@echo "  vet                  - Run go vet"
	@echo "  tidy                 - Run go mod tidy"

# Run all tests
test:
	cd kubectl-x && go test ./...
	cd kubernetes-mcp && go test ./...
	cd kustomizelite && go test ./...
	cd celery && go test ./...

# Run golangci-lint
lint:
	cd kubectl-x && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 10m
	cd kubernetes-mcp && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 10m
	cd kustomizelite && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 10m
	cd celery && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 10m

# Run golangci-lint with auto-fix
lint-fix:
	cd kubectl-x && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout 10m
	cd kubernetes-mcp && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout 10m
	cd kustomizelite && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout 10m
	cd celery && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout 10m

# Build all binaries
build: build-kubectl-x build-kubernetes-mcp build-kustomizelite build-celery

# Build the kubectl-x binary
build-kubectl-x:
	cd kubectl-x && go build -o ../kubectl-x .

# Build the kubernetes-mcp binary
build-kubernetes-mcp:
	cd kubernetes-mcp && go build -o ../kubernetes-mcp .

# Build the kustomizelite binary
build-kustomizelite:
	cd kustomizelite && go build -o ../kustomizelite .

# Build the celery binary
build-celery:
	cd celery && go build -o ../celery .

# Format Go code
fmt:
	cd kubectl-x && go fmt ./...
	cd kubernetes-mcp && go fmt ./...
	cd kustomizelite && go fmt ./...
	cd celery && go fmt ./...

# Run go vet
vet:
	cd kubectl-x && go vet ./...
	cd kubernetes-mcp && go vet ./...
	cd kustomizelite && go vet ./...
	cd celery && go vet ./...

# Run go mod tidy
tidy:
	cd kubectl-x && go mod tidy
	cd kubernetes-mcp && go mod tidy
	cd kustomizelite && go mod tidy
	cd celery && go mod tidy
	go work sync