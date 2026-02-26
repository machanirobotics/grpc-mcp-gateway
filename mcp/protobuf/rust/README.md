# mcp-protobuf

Pre-compiled Protocol Buffer types for [grpc-mcp-gateway](https://github.com/machanirobotics/grpc-mcp-gateway) — the `mcp.protobuf` package containing MCP annotations for gRPC services.

## Install

Use the version matching the [grpc-mcp-gateway release](https://github.com/machanirobotics/grpc-mcp-gateway/releases) you use. See [crates.io](https://crates.io/crates/mcp-protobuf) for the current version.

```bash
cargo add mcp-protobuf
```

Or add to `Cargo.toml` (replace with the version from the latest release):

```toml
[dependencies]
mcp-protobuf = "1.4.1"   # check crates.io for latest
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
- **`MCPFieldOptions`** — Field description, examples, format, deprecated
- **`MCPEnumOptions`**, **`MCPEnumValueOptions`** — Enum and enum-value descriptions
- **`MCPMimeType`**, **`MCPFieldType`** — Enums

## Usage

Import the crate to use MCP-annotated protos in your Rust project:

```rust
use mcp_protobuf::*;
```

When using [protoc-gen-mcp](https://github.com/machanirobotics/grpc-mcp-gateway) with `lang=rust`, the generated code will depend on this crate. Use the same version as your grpc-mcp-gateway release (see [crates.io](https://crates.io/crates/mcp-protobuf) for latest).

## Dependencies

- `prost` 0.14
- `prost-types` 0.14

## Links

- **Source**: [github.com/machanirobotics/grpc-mcp-gateway](https://github.com/machanirobotics/grpc-mcp-gateway)
- **Proto definitions**: [buf.build/machanirobotics/grpc-mcp-gateway](https://buf.build/machanirobotics/grpc-mcp-gateway)
- **License**: Apache-2.0
