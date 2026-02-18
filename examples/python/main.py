"""
CLI entrypoint for the TodoService gRPC + MCP server.

Usage:
    uv run python main.py                              # gRPC :50051 + MCP streamable-http :8082
    MCP_TRANSPORT=stdio uv run python main.py          # stdio (no gRPC)
    MCP_TRANSPORT=streamable-http uv run python main.py
"""

from __future__ import annotations

import logging
import os

import server


def main() -> None:
    logging.basicConfig(level=logging.INFO)

    transport = os.getenv("MCP_TRANSPORT", "streamable-http")
    host = os.getenv("MCP_HOST", "0.0.0.0")
    mcp_port = int(os.getenv("MCP_PORT", "8082"))
    grpc_port = int(os.getenv("GRPC_PORT", "50051"))

    print(
        f"Starting TodoService servers\n"
        f"  gRPC  -> [::]:{grpc_port}\n"
        f"  MCP   -> {host}:{mcp_port} (transport={transport})"
    )

    server.start(
        transport=transport,
        host=host,
        mcp_port=mcp_port,
        grpc_port=grpc_port,
    )


if __name__ == "__main__":
    main()
