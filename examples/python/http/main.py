"""
TodoService MCP server over streamable-http + gRPC side-by-side.

Usage:
    cd examples/python && uv run python http/main.py
"""

from __future__ import annotations

import asyncio
import logging
import os
import sys
import threading

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "..", "proto", "generated", "python"))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import grpc  # noqa: E402
from grpc_reflection.v1alpha import reflection as grpc_reflection  # noqa: E402
from todo.v1 import todo_service_pb2  # noqa: E402
from todo.v1.todo_service_pb2_grpc import add_TodoServiceServicer_to_server  # noqa: E402
from todo.v1.todo_service_pb2_mcp import serve_todo_service_mcp  # noqa: E402

from internal.grpc_servicer import TodoGRPCServicer  # noqa: E402
from internal.impl import TodoServer  # noqa: E402

log = logging.getLogger(__name__)


def _run_grpc(impl: TodoServer, port: int) -> None:
    async def _serve() -> None:
        server = grpc.aio.server()
        add_TodoServiceServicer_to_server(TodoGRPCServicer(impl), server)
        service_names = (
            todo_service_pb2.DESCRIPTOR.services_by_name["TodoService"].full_name,
            grpc_reflection.SERVICE_NAME,
        )
        grpc_reflection.enable_server_reflection(service_names, server)
        server.add_insecure_port(f"[::]:{port}")
        log.info("gRPC server listening on port %d (reflection enabled)", port)
        await server.start()
        await server.wait_for_termination()

    asyncio.run(_serve())


def main() -> None:
    logging.basicConfig(level=logging.INFO)

    host = os.getenv("MCP_HOST", "0.0.0.0")
    mcp_port = int(os.getenv("MCP_PORT", "8082"))
    grpc_port = int(os.getenv("GRPC_PORT", "50051"))

    print(
        f"Starting TodoService servers\n"
        f"  gRPC  -> [::]:{grpc_port}\n"
        f"  MCP   -> {host}:{mcp_port} (transport=streamable-http)"
    )

    impl = TodoServer()

    grpc_thread = threading.Thread(
        target=_run_grpc, args=(impl, grpc_port), daemon=True
    )
    grpc_thread.start()

    serve_todo_service_mcp(
        impl,
        transport="streamable-http",
        host=host,
        port=mcp_port,
    )


if __name__ == "__main__":
    main()
