# Variables
BINARY_NAME=ghsurf_server
CMD_PATH=./cmd/server/main.go

# Phony targets (commands that don't produce files with the same name)
.PHONY: all test build run clean help

# Default target (runs when 'make' is called without arguments)
# Let's make running tests the default action.
all: test

## test: Run all unit tests verbosely
test:
	@echo "==> Running tests..."
	go test -v ./...

## build: Build the server executable
build:
	@echo "==> Building server..."
	go build -o $(BINARY_NAME) $(CMD_PATH)

## run: Build and run the server (requires config like GHSURF_GITHUB_TOKEN)
run: build
	@echo "==> Running server (ensure GHSURF_GITHUB_TOKEN is set or .env configured)..."
	./$(BINARY_NAME)

## clean: Remove the built server executable
clean:
	@echo "==> Cleaning up..."
	rm -f $(BINARY_NAME)
