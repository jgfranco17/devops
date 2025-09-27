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
