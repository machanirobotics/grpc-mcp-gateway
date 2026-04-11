# @machanirobotics/grpc-mcp-gateway-protos

TypeScript types for the `mcp.protobuf` protos used by [grpc-mcp-gateway](https://github.com/machanirobotics/grpc-mcp-gateway): service/tool/prompt/elicitation options, field metadata, enums, and related messages.

Code is generated with [protobuf-ts](https://github.com/timostamm/protobuf-ts) and depends on `@protobuf-ts/runtime`.

## Install

```bash
npm install @machanirobotics/grpc-mcp-gateway-protos
# or
bun add @machanirobotics/grpc-mcp-gateway-protos

# Or pin to the same release as the gateway (recommended)
npm install @machanirobotics/grpc-mcp-gateway-protos@1.5.3
bun add @machanirobotics/grpc-mcp-gateway-protos@1.5.3
```

Use the version that matches the gateway / code generator release you rely on.

## Usage

```ts
import { MCPApp, MCPServiceOptions } from "@machanirobotics/grpc-mcp-gateway-protos";

const app: MCPApp = { name: "My MCP", version: "1.0.0" };
```

For `google/protobuf/descriptor.proto` types (e.g. custom tooling), import the subpath:

```ts
import { FileDescriptorSet } from "@machanirobotics/grpc-mcp-gateway-protos/google/protobuf/descriptor";
```

## Regenerating

From the repo `proto/` directory:

```bash
buf generate
```

This refreshes `lib/` (protobuf-ts). This package does **not** ship gRPC-Web clients; generate Connect or gRPC-Web stubs in your application if you need them.

## Development

```bash
bun install --frozen-lockfile
bun run check # TypeScript --noEmit
bun run lint  # Biome (does not lint buf-generated `lib/` — regenerate with `buf generate` instead of editing)
```
