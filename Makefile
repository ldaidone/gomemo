# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod
GOTIDY=$(GOMOD) tidy
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOINSTALL=$(GOCMD) install

# Binary name
BINARY_NAME=gomemo-example
BINARY_OUTPUT=dist

# Build the main example
BUILD_INPUT ?= cmd/examples/*.go
BUILD_OUTPUT ?= $(BINARY_OUTPUT)/$(BINARY_NAME)
.PHONY: build
build:
	mkdir -p $(BINARY_OUTPUT)
	@$(GOBUILD) -ldflags="-s -w" -o $(BUILD_OUTPUT) $(BUILD_INPUT)

# Run the example
RUN_INPUT ?= cmd/examples/*.go
.PHONY: run
run:
	$(GOCMD) run $(RUN_INPUT)

# Install the example
INSTALL_INPUT ?= cmd/examples/main.go
.PHONY: install
install:
	$(GOINSTALL) $(INSTALL_INPUT)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -cover -coverpkg=./... -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run go vet
.PHONY: vet
vet:
	$(GOVET) ./...

# Run go fmt
.PHONY: fmt
fmt:
	$(GOFMT) -w .
	
# Tidy go modules
.PHONY: tidy
tidy:
	$(GOTIDY)

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BINARY_OUTPUT)
	rm -f coverage.out
	rm -f coverage.html

# Run all checks
.PHONY: check
check: vet fmt test

# Build example and run it
.PHONY: demo
NAME ?= ""
demo:
	$(GOCMD) run $(RUN_INPUT) -name $(NAME)

# Show project information
.PHONY: info
info:
	@echo "Project: gomemo"
	@echo "Description: Generic memoization library for Go with pluggable backends"
	@echo "Author: Leo Daidone <leo.daidone@gmail.com>"
	@echo "Main package: memo"
	@echo "Available commands:"
	@echo "  make build <BUILD_INPUT> <BUILD_OUTPUT>			- Build the example binary"
	@echo "  make run <RUN_INPUT>          				- Run the example directly"
	@echo "  make test          						- Run all tests"
	@echo "  make test-coverage 						- Run tests with coverage"
	@echo "  make vet           						- Run go vet"
	@echo "  make fmt           						- Run go fmt"
	@echo "  make tidy          						- Tidy go modules"
	@echo "  make check         						- Run all checks (vet, fmt, test)"
	@echo "  make demo          						- Build and run the demo"
	@echo "  make clean         						- Clean build artifacts"
	@echo "  make info          						- Show this information"