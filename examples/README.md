# Examples

End-to-end examples demonstrating how to expose a gRPC **TodoService** as an MCP server in Go, Python, and Rust using `grpc-mcp-gateway`.

## Overview

All three languages share the same proto definition and produce identical MCP tool surfaces. Each language example includes separate entrypoints for every supported transport:

| Transport | Description | Default Port |
| ------------------- | ----------------------------------------- | ------------ |
| `streamable-http` | Modern HTTP-based MCP transport | 8082 |
| `stdio` | Stdin/stdout for CLI tools (Claude Desktop) | — |
| `sse` | Server-Sent Events (legacy 2024-11-05 spec) | 8083 |

## Proto Definition

The TodoService proto lives in `proto/todo/v1/` and imports MCP annotations from the published BSR module:

```yaml
# buf.yaml
deps:
  - buf.build/machanirobotics/grpc-mcp-gateway
```

```protobuf
import "mcp/protobuf/annotations.proto";

service TodoService {
  option (mcp.protobuf.service) = {
    app: { name: "Todo App" version: "1.0.0" }
  };

  rpc CreateTodo(CreateTodoRequest) returns (Todo) {
    option (mcp.protobuf.tool) = { ... };
    option (mcp.protobuf.elicitation) = { ... };
  };
}
```

The proto uses:
- **`mcp.protobuf.service`** — app-level metadata
- **`mcp.protobuf.tool`** — per-RPC tool name/description overrides
- **`mcp.protobuf.prompt`** — per-RPC prompt templates with schema-based arguments
- **`mcp.protobuf.elicitation`** — confirmation dialogs with schema-based forms
- **`google.api.resource`** — auto-detected MCP resources from AIP resource annotations

## Code Generation

Install the plugin and generate code for all languages:

```bash
go install github.com/machanirobotics/grpc-mcp-gateway/v2/plugin/cmd/protoc-gen-mcp@latest

cd examples
buf generate
```

This produces:
- `proto/generated/go/` — Go pb + gRPC + MCP files
- `proto/generated/python/` — Python pb + gRPC + MCP files
- `proto/generated/rust/` — Rust pb + gRPC + MCP files

## Generated MCP Tools

All five CRUD operations are exposed as MCP tools:

| Tool Name | Description |
| ------------------------------------ | ----------------------------------------- |
| `todo_v1_TodoService_CreateTodo` | Creates a new todo item |
| `todo_v1_TodoService_GetTodo` | Retrieves a todo by resource name |
| `todo_v1_TodoService_ListTodos` | Lists todos with pagination |
| `todo_v1_TodoService_UpdateTodo` | Updates an existing todo |
| `todo_v1_TodoService_DeleteTodo` | Deletes a todo by resource name |

## Language Examples

| Language | Directory | Details |
| -------- | ----------------------- | ---------------------------------- |
| Go | [`go/`](go/) | Uses `runtime.StartServer` |
| Python | [`python/`](python/) | Uses `FastMCP` / low-level `Server` |
| Rust | [`rust/`](rust/) | Uses `rmcp` SDK with `ServerHandler` |

Each has its own README with setup and run instructions.

## Testing with MCP Inspector

For `streamable-http` or `sse` servers:

```bash
npx @modelcontextprotocol/inspector
# Enter the server URL, e.g. http://localhost:8082/todo/v1/todoservice/mcp
```

For `stdio` servers:

```bash
npx @modelcontextprotocol/inspector -- /absolute/path/to/binary
```
