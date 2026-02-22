# grpc-mcp-gateway Justfile
# Run `just --list` to see all available recipes.

mod := "github.com/machanirobotics/grpc-mcp-gateway/v2"
bin := "protoc-gen-mcp"

# Default: list all recipes
default:
    @just --list

# Build the protoc-gen-mcp plugin
build:
    go build -o ./bin/{{bin}} ./plugin/cmd/protoc-gen-mcp/

# Build with version info baked in (for releases)
build-release version="dev":
    go build -trimpath -ldflags "-s -w -X main.version={{version}}" -o ./bin/{{bin}} ./plugin/cmd/protoc-gen-mcp/

# Cross-compile for a specific OS/ARCH
build-cross os arch version="dev":
    GOOS={{os}} GOARCH={{arch}} go build -trimpath -ldflags "-s -w -X main.version={{version}}" \
        -o ./bin/{{bin}}-{{os}}-{{arch}}{{if os == "windows" { ".exe" } else { "" } }} \
        ./plugin/cmd/protoc-gen-mcp/

# Build binaries for all release platforms
build-all version="dev":
    just build-cross linux   amd64 {{version}}
    just build-cross linux   arm64 {{version}}
    just build-cross darwin  amd64 {{version}}
    just build-cross darwin  arm64 {{version}}
    just build-cross windows amd64 {{version}}
    just build-cross windows arm64 {{version}}

# Install the plugin to $GOPATH/bin
install:
    go install ./plugin/cmd/protoc-gen-mcp/

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run go vet
vet:
    go vet ./plugin/... ./runtime/...

# Check formatting
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

# Format all Go files
fmt:
    gofmt -w .

# Lint proto files
buf-lint:
    cd proto && buf lint

# Run all Go tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Run tests with coverage
test-cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run Python smoke test
test-python:
    cd examples/python && uv run python -m pytest smoke_test.py -v

# Run Rust check
test-rust:
    cd examples/rust && cargo check --all-targets

# Run all tests (Go + Python + Rust)
test-all: test test-python test-rust

# Rebuild the plugin and regenerate example proto code
generate: install
    cd examples && buf generate

# Run all checks (fmt, vet, lint, test, build)
check: fmt-check vet lint test build
# Quick check (vet + test + build, no lint)
check-quick: vet test build

# Remove build artifacts
clean:
    rm -rf ./bin ./coverage.out
    rm -rf ./dist


# Push proto module to buf.build/machanirobotics/grpc-mcp-gateway
buf-push:
    cd proto && buf push

# Push proto module with a specific label (e.g. a release tag)
buf-push-label label:
    cd proto && buf push --label {{label}}

# Create release archives for all platforms and push protos to BSR
release version: clean (build-all version)
    mkdir -p dist
    cd bin && for f in {{bin}}-*; do \
        if echo "$f" | grep -q windows; then \
            zip ../dist/"$f".zip "$f"; \
        else \
            tar czf ../dist/"$f".tar.gz "$f"; \
        fi \
    done
    @echo "Release archives in ./dist/"
    @ls -lh dist/
    @echo ""
    @echo "Pushing proto module to BSR with label {{version}} ..."
    cd proto && buf push --label {{version}}
    @echo "Done. Proto published as buf.build/machanirobotics/grpc-mcp-gateway:{{version}}"
