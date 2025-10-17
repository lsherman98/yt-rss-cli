# Makefile for the ytrss CLI

# The binary to build
BINARY_NAME=ytrss

# Default target
all: build

# Build the Go application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .
	@echo "$(BINARY_NAME) built successfully."

# Run the application
run:
	@./$(BINARY_NAME)

# Clean the build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@echo "Cleanup complete."

.PHONY: all build run clean
