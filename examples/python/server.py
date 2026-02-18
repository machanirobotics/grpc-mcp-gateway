"""
TodoService combined gRPC + MCP server.

Starts a gRPC server and an MCP server side-by-side, sharing the same
in-memory TodoServer implementation.
"""

from __future__ import annotations

import asyncio
import logging
import os
import sys
import threading

# Make generated code importable.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "proto", "generated", "python"))

import grpc  # noqa: E402
from grpc_reflection.v1alpha import reflection as grpc_reflection  # noqa: E402
from todo.v1 import todo_service_pb2  # noqa: E402
from todo.v1.todo_service_pb2_grpc import add_TodoServiceServicer_to_server  # noqa: E402
from todo.v1.todo_service_pb2_mcp import serve_todo_service_mcp  # noqa: E402

from grpc_servicer import TodoGRPCServicer  # noqa: E402
from impl import TodoServer  # noqa: E402

log = logging.getLogger(__name__)


def _run_grpc(impl: TodoServer, port: int) -> None:
    """Start the async gRPC server in its own event loop (runs in a thread)."""

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


def start(
    *,
    transport: str = "streamable-http",
    host: str = "0.0.0.0",
    mcp_port: int = 8082,
    grpc_port: int = 50051,
) -> None:
    """Start both gRPC and MCP servers (blocking).

    - gRPC runs in a daemon thread on ``grpc_port``.
    - MCP runs in the main thread on ``mcp_port``.
    """
    impl = TodoServer()

    grpc_thread = threading.Thread(
        target=_run_grpc, args=(impl, grpc_port), daemon=True
    )
    grpc_thread.start()

    serve_todo_service_mcp(
        impl,
        transport=transport,
        host=host,
        port=mcp_port,
    )
