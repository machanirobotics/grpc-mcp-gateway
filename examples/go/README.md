# Go Examples

TodoService MCP server examples in Go, demonstrating all supported transports.

## Prerequisites

- Go 1.25+
- `protoc-gen-mcp` installed (`go install github.com/machanirobotics/grpc-mcp-gateway/plugin/cmd/protoc-gen-mcp@latest`)
- Generated code already in `proto/generated/go/` (run `buf generate` from `examples/`)

## Structure

```
go/
├── http/          # streamable-http + gRPC side-by-side
│   ├── main.go
│   ├── impl.go
│   └── smoke_test.go
├── stdio/         # stdio transport (for Claude Desktop, MCP Inspector)
│   ├── main.go
│   └── impl.go
├── sse/           # SSE transport (legacy)
│   ├── main.go
│   └── impl.go
└── grpc-gateway/  # gRPC-to-MCP gateway forwarding
    └── main.go
```

Each transport has its own `impl.go` with an in-memory `todoServer` that implements both the gRPC `TodoServiceServer` and MCP `TodoServiceMCPServer` interfaces.

## Running

### Streamable HTTP (+ gRPC)

```bash
cd examples/go/http
go run .
# gRPC  → [::]:50051 (reflection enabled)
# MCP   → 0.0.0.0:8080/todo/v1/todoservice/mcp
```

### Stdio

```bash
cd examples/go/stdio
go run .
# MCP communicates over stdin/stdout
# Logs go to stderr to avoid corrupting JSON-RPC
```

For MCP Inspector:

```bash
go build -o /tmp/todo-stdio ./examples/go/stdio
npx @modelcontextprotocol/inspector -- /tmp/todo-stdio
```

### SSE

```bash
cd examples/go/sse
go run .
# MCP → 0.0.0.0:8080/todo/v1/todoservice/mcp (SSE)
```

### gRPC Gateway

Forwards MCP requests to an upstream gRPC server:

```bash
cd examples/go/grpc-gateway
go run .
# Connects to gRPC at localhost:50051
# MCP → 0.0.0.0:8080
```

## Testing

The `http/` example includes a smoke test that exercises the full CRUD pipeline over an in-memory MCP transport:

```bash
cd examples/go/http
go test -v
```

Test flow:
1. List tools (expects 5)
2. CreateTodo
3. GetTodo
4. ListTodos
5. DeleteTodo
6. Verify deletion

## Architecture

The generated code (`todo_service.pb.mcp.go`) provides:
- `TodoServiceMCPServer` interface — one method per RPC
- `RegisterTodoServiceMCPHandler(s *mcp.Server, impl TodoServiceMCPServer, opts ...runtime.Option)` — registers all tools, prompts, and resources
- `ServeTodoServiceMCP(impl, cfg)` — convenience function that creates the server and starts it

The `runtime` package handles transport selection, multi-transport serving, header-to-metadata forwarding, and schema injection.
