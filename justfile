# Default command
_default:
    @just --list --unsorted

# Sync Go modules
tidy:
    go mod tidy
    @echo "All modules synced, Go workspace ready!"

# CLI local run wrapper
devops *args:
    @go run . {{ args }}

# Run all BDD tests
test:
    @echo "Running unit tests!"
    go clean -testcache
    go test -cover ./...

# Build the binary
build:
    #!/usr/bin/env bash
    go mod download all
    CGO_ENABLED=0 GOOS=linux go build -o ./devops .
    echo "Built binary for devops successfully!"

# Update the project dependencies
update-deps:
    @echo "Updating project dependencies..."
    go get -u ./...
    go mod tidy
