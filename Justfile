# protoc-mcp-gen Justfile
# Run `just --list` to see all available recipes.

mod := "github.com/machanirobotics/protoc-mcp-gen"
bin := "protoc-gen-mcp"

# Default: list all recipes
default:
    @just --list

# ---------- Build ----------

# Build the protoc-gen-mcp plugin
build:
    go build -o ./bin/{{bin}} ./cmd/protoc-gen-mcp/

# Build with version info baked in (for releases)
build-release version="dev":
    go build -trimpath -ldflags "-s -w -X main.version={{version}}" -o ./bin/{{bin}} ./cmd/protoc-gen-mcp/

# Cross-compile for a specific OS/ARCH
build-cross os arch version="dev":
    GOOS={{os}} GOARCH={{arch}} go build -trimpath -ldflags "-s -w -X main.version={{version}}" \
        -o ./bin/{{bin}}-{{os}}-{{arch}}{{if os == "windows" { ".exe" } else { "" } }} \
        ./cmd/protoc-gen-mcp/

# Build binaries for all release platforms
build-all version="dev":
    just build-cross linux   amd64 {{version}}
    just build-cross linux   arm64 {{version}}
    just build-cross darwin  amd64 {{version}}
    just build-cross darwin  arm64 {{version}}
    just build-cross windows amd64 {{version}}
    just build-cross windows arm64 {{version}}

# ---------- Install ----------

# Install the plugin to $GOPATH/bin
install:
    go install ./cmd/protoc-gen-mcp/

# ---------- Lint ----------

# Run golangci-lint
lint:
    golangci-lint run ./...

# Run go vet
vet:
    go vet ./cmd/... ./pkg/...

# Check formatting
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

# Format all Go files
fmt:
    gofmt -w .

# ---------- Test ----------

# Run all tests
test:
    go test ./cmd/... ./pkg/...

# Run tests with verbose output
test-verbose:
    go test -v ./cmd/... ./pkg/...

# Run tests with race detector
test-race:
    go test -race ./cmd/... ./pkg/...

# Run tests with coverage
test-cover:
    go test -coverprofile=coverage.out ./cmd/... ./pkg/...
    go tool cover -func=coverage.out

# ---------- Generate ----------

# Rebuild the plugin and regenerate example proto code
generate: build
    cp ./bin/{{bin}} $(go env GOPATH)/bin/{{bin}}
    cd examples && buf generate

# ---------- Check ----------

# Run all checks (fmt, vet, lint, test, build)
check: fmt-check vet lint test build

# Quick check (vet + test + build, no lint)
check-quick: vet test build

# ---------- Clean ----------

# Remove build artifacts
clean:
    rm -rf ./bin ./coverage.out
    rm -rf ./dist

# ---------- Release (local) ----------

# Create release archives for all platforms
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
