.PHONY: test lint lint-fix build build-kubectl-x build-kubernetes-mcp build-celery fmt vet tidy install uninstall clean help

# Installation directory
INSTALL_DIR := /usr/local/bin

# Default target
help:
	@echo "Available targets:"
	@echo "  test                 - Run all tests"
	@echo "  lint                 - Run golangci-lint"
	@echo "  lint-fix             - Run golangci-lint with auto-fix"
	@echo "  build                - Build all binaries"
	@echo "  build-kubectl-x      - Build the kubectl-x binary"
	@echo "  build-kubernetes-mcp - Build the kubernetes-mcp binary"
	@echo "  build-celery         - Build the celery binary"
	@echo "  fmt                  - Format Go code"
	@echo "  vet                  - Run go vet"
	@echo "  tidy                 - Run go mod tidy"
	@echo "  install              - Install all binaries to $(INSTALL_DIR)"
	@echo "  uninstall            - Remove installed binaries from $(INSTALL_DIR)"
	@echo "  clean                - Remove built binaries"

# Run all tests
test:
	cd kubectl-x && go test ./...
	cd kubernetes-mcp && go test ./...
	cd celery && go test ./...

# Run golangci-lint
lint:
	cd kubectl-x && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 10m
	cd kubernetes-mcp && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 10m
	cd celery && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 10m

# Run golangci-lint with auto-fix
lint-fix:
	cd kubectl-x && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout 10m
	cd kubernetes-mcp && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout 10m
	cd celery && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout 10m

# Build all binaries
build: build-kubectl-x build-kubernetes-mcp build-celery

# Build the kubectl-x binary
build-kubectl-x:
	cd kubectl-x && go build -o ../bin/kubectl-x .

# Build the kubernetes-mcp binary
build-kubernetes-mcp:
	cd kubernetes-mcp && go build -o ../bin/kubernetes-mcp .

# Build the celery binary
build-celery:
	cd celery && go build -o ../bin/celery .

# Format Go code
fmt:
	cd kubectl-x && go fmt ./...
	cd kubernetes-mcp && go fmt ./...
	cd celery && go fmt ./...

# Run go vet
vet:
	cd kubectl-x && go vet ./...
	cd kubernetes-mcp && go vet ./...
	cd celery && go vet ./...

# Run go mod tidy
tidy:
	cd kubectl-x && go mod tidy
	cd kubernetes-mcp && go mod tidy
	cd celery && go mod tidy
	go work sync

# Install all binaries to /usr/local/bin
install: build
	@echo "Installing binaries to $(INSTALL_DIR)..."
	@mkdir -p bin
	@install -m 755 bin/kubectl-x $(INSTALL_DIR)/kubectl-x
	@install -m 755 bin/kubernetes-mcp $(INSTALL_DIR)/kubernetes-mcp
	@install -m 755 bin/celery $(INSTALL_DIR)/celery
	@echo "Installation complete!"
	@echo "  kubectl-x installed to $(INSTALL_DIR)/kubectl-x"
	@echo "  kubernetes-mcp installed to $(INSTALL_DIR)/kubernetes-mcp"
	@echo "  celery installed to $(INSTALL_DIR)/celery"

# Uninstall binaries from /usr/local/bin
uninstall:
	@echo "Removing binaries from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/kubectl-x
	@rm -f $(INSTALL_DIR)/kubernetes-mcp
	@rm -f $(INSTALL_DIR)/celery
	@echo "Uninstall complete!"

# Clean built binaries
clean:
	@echo "Cleaning built binaries..."
	@rm -rf bin
	@echo "Clean complete!"
