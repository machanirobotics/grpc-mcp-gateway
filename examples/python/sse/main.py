"""
TodoService MCP server over SSE (Server-Sent Events).

Usage:
    cd examples/python && uv run python sse/main.py
"""

from __future__ import annotations

import logging
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "..", "proto", "generated", "python"))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from todo.v1.todo_service_pb2_mcp import serve_todo_service_mcp  # noqa: E402
from internal.impl import TodoServer  # noqa: E402


def main() -> None:
    logging.basicConfig(level=logging.INFO)

    host = os.getenv("MCP_HOST", "0.0.0.0")
    port = int(os.getenv("MCP_PORT", "8083"))

    print(f"Starting TodoService MCP SSE server on {host}:{port}")

    serve_todo_service_mcp(
        TodoServer(),
        transport="sse",
        host=host,
        port=port,
    )


if __name__ == "__main__":
    main()
