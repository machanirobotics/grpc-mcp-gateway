# mcp-protobuf

Pre-compiled Protocol Buffer types for [grpc-mcp-gateway](https://github.com/machanirobotics/grpc-mcp-gateway) — the `mcp.protobuf` package containing MCP annotations for gRPC services.

## Install

```toml
# Cargo.toml
[dependencies]
mcp-protobuf = "0.1"
```

Or:

```bash
cargo add mcp-protobuf
```

## What's included

This crate provides the Rust bindings (via [prost](https://crates.io/crates/prost)) for:

- **`mcp.protobuf`** — Service, tool, prompt, and elicitation options for annotating `.proto` files
- **`MCPServiceOptions`** — App metadata
- **`MCPToolOptions`** — Tool name/description overrides
- **`MCPPrompt`** — Prompt templates
- **`MCPElicitation`** — Confirmation dialogs
- **`MCPResource`** — Resource definitions
- **`MCPApp`** — App info
- **`MCPMimeType`**, **`MCPFieldType`** — Enums

## Usage

Import the crate to use MCP-annotated protos in your Rust project:

```rust
use mcp_protobuf::*;
```

When using [protoc-gen-mcp](https://github.com/machanirobotics/grpc-mcp-gateway) with `lang=rust`, the generated code will depend on this crate. Add it to your `Cargo.toml`:

```toml
[dependencies]
mcp-protobuf = "0.1"
# Your generated MCP stubs will use it
```

## Dependencies

- `prost` 0.14
- `prost-types` 0.14

## Links

- **Source**: [github.com/machanirobotics/grpc-mcp-gateway](https://github.com/machanirobotics/grpc-mcp-gateway)
- **Proto definitions**: [buf.build/machanirobotics/grpc-mcp-gateway](https://buf.build/machanirobotics/grpc-mcp-gateway)
- **License**: Apache-2.0
