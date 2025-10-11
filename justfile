# Devops: Development scripts

INSTALL_PATH := "$HOME/.local"

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

# Start the API server
start:
    @echo "Starting API server..."
    @uv run python app.py

# Run Go tests
test:
    go clean -testcache
    go test -cover ./...

# Run Python tests
pytest *args:
    uv run pytest {{ args }}

# Build the binary
build:
    #!/usr/bin/env bash
    # Detect OS and architecture
    case "$(uname -s)" in
        Linux*) OS="linux" ;;
        Darwin*) OS="darwin" ;;
        *) echo "Error: Unsupported OS (${OS})"; exit 1 ;;
    esac
    case "$(uname -m)" in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64) ARCH="arm64" ;;
        *) echo "Error: Unsupported architecture (${ENV_ARCH})"; exit 1 ;;
    esac

    echo "Building devops for ${OS}/${ARCH}..."
    go mod download all
    CGO_ENABLED=0 GOOS="${OS}" GOARCH="${ARCH}" go build -o ./devops .
    echo "Built binary for devops successfully!"

# Install the binary locally
install-local: build
    #!/usr/bin/env bash
    set -eux
    echo "Installing devops locally..."
    BIN_PATH="{{ INSTALL_PATH }}/bin/devops"
    cp ./devops "${BIN_PATH}"
    chmod +x "${BIN_PATH}"
    echo "Installed devops locally!"

# Remove the local binary
uninstall-local:
    #!/usr/bin/env bash
    set -eux
    echo "Uninstalling devops..."
    BIN_PATH="{{ INSTALL_PATH }}/bin/devops"
    rm "${BIN_PATH}"
    echo "Uninstalled devops!"

# Update the project dependencies
update-deps:
    @echo "Updating project dependencies..."
    go get -u ./...
    go mod tidy

# Run the docs server locally
docs:
    mkdocs build --strict --clean
    mkdocs serve --open

# Run linters
lint:
    @echo "Running linters..."
    uv run flake8 .
    uv run black --check .
    uv run isort --check-only .
    uv run mypy .
