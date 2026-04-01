# Justfile for md-go-validator

# Default recipe - run tests
default:
    go test ./...

# Run all tests
test:
    go test ./...

# Run tests with coverage
cover:
    go test -cover ./...

# Build the binary
build:
    go build ./cmd/md-go-validator

# Install locally (used by install-local.sh)
install-local:
    go install ./cmd/md-go-validator

# Run the validator
run *args:
    go run ./cmd/md-go-validator {{args}}

# Clean build artifacts
clean:
    rm -f md-go-validator

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run
