# Rust Examples

TodoService MCP server examples in Rust, demonstrating all supported transports.

## Prerequisites

- Rust 1.93.0+ (stable; edition 2024)
- Generated code already in `proto/generated/rust/` (run `buf generate` from `examples/`)

The generated code includes both **TodoService** (`todo/v1/`) and **CounterService** (`counter/v1/`); the Rust examples use TodoService.

## Structure

```
rust/
├── Cargo.toml         # Lib + 3 binary targets
└── src/
    ├── lib.rs         # Re-exports shared modules
    ├── proto.rs       # Prost includes for todo.v1
    ├── impl.rs        # Shared TodoServer (gRPC + MCP impls)
    └── bin/
        ├── http.rs    # streamable-http + gRPC side-by-side
        ├── stdio.rs   # stdio transport (for Claude Desktop)
        └── sse.rs     # SSE transport (legacy)
```

The `impl.rs` module contains the `TodoServer` struct with both the tonic `TodoService` trait impl (gRPC) and the generated `TodoServiceMcpServer` trait impl (MCP JSON ↔ prost bridge). All three binaries share it via the library crate.

## Building

```bash
cd examples/rust
cargo build
```

This produces three binaries: `http`, `stdio`, and `sse`.

## Running

### Streamable HTTP (+ gRPC)

```bash
cargo run --bin http
# gRPC  → [::]:50051 (reflection enabled)
# MCP   → 0.0.0.0:8082 (streamable-http)
```

Environment variables:
- `MCP_HOST` — bind address (default `0.0.0.0`)
- `MCP_PORT` — MCP port (default `8082`)

### Stdio

```bash
cargo run --bin stdio
# MCP communicates over stdin/stdout
```

For MCP Inspector:

```bash
cargo build --bin stdio
npx @modelcontextprotocol/inspector -- ./target/debug/stdio
```

### SSE

```bash
cargo run --bin sse
# MCP → 0.0.0.0:8083 (SSE)
```

Environment variables:
- `MCP_HOST` — bind address (default `0.0.0.0`)
- `MCP_PORT` — MCP port (default `8083`)

## Architecture

The generated code (`todo_service.mcp.rs`) provides:
- `TodoServiceMcpServer` trait — one async method per RPC, taking `serde_json::Value` args and returning `Result<Value, McpError>`
- `TodoServiceMcpHandler<T>` — wraps any `T: TodoServiceMcpServer` and implements `rmcp::ServerHandler` (tools, prompts, resources)
- `serve_todo_service_mcp(impl, config)` — convenience function that creates the handler and starts the transport
- `serve_todo_service_mcp_stdio(impl)` — shortcut for stdio transport

## Dependencies

| Crate | Purpose |
| ------------------- | ----------------------------------------- |
| `rmcp` | MCP Rust SDK (ServerHandler, transports) |
| `async-trait` | Async trait support |
| `serde_json` | JSON serialization for MCP args/results |
| `prost` / `prost-types` | Protocol Buffers runtime |
| `tonic` | gRPC runtime |
| `tonic-reflection` | gRPC server reflection |
| `tokio` | Async runtime |
| `axum` | HTTP server (used by rmcp streamable-http) |
