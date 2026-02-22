"""
TodoService MCP server over stdio (for CLI tools like Claude Desktop).

Usage:
    cd examples/python && uv run python stdio/main.py
"""

from __future__ import annotations

import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "..", "proto", "generated", "python"))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from todo.v1.todo_service_pb2_mcp import serve_todo_service_mcp  # noqa: E402
from internal.impl import TodoServer  # noqa: E402


def main() -> None:
    serve_todo_service_mcp(
        TodoServer(),
        transport="stdio",
    )


if __name__ == "__main__":
    main()
