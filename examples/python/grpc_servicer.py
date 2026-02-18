"""gRPC servicer adapter for TodoServer.

Wraps the async TodoServer (MCP impl) so it can serve gRPC requests
via grpc.aio.
"""

from __future__ import annotations

import grpc

from todo.v1 import todo_pb2
from todo.v1.todo_service_pb2_grpc import TodoServiceServicer

from google.protobuf.empty_pb2 import Empty


class TodoGRPCServicer(TodoServiceServicer):
    """Adapts TodoServer async methods to the gRPC TodoServiceServicer interface."""

    def __init__(self, impl) -> None:
        self._impl = impl

    async def CreateTodo(self, request, context):
        try:
            return await self._impl.create_todo(request)
        except ValueError as e:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(str(e))
            return todo_pb2.Todo()

    async def GetTodo(self, request, context):
        try:
            return await self._impl.get_todo(request)
        except ValueError as e:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(str(e))
            return todo_pb2.Todo()

    async def ListTodos(self, request, context):
        return await self._impl.list_todos(request)

    async def UpdateTodo(self, request, context):
        try:
            return await self._impl.update_todo(request)
        except ValueError as e:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(str(e))
            return todo_pb2.Todo()

    async def DeleteTodo(self, request, context):
        try:
            return await self._impl.delete_todo(request)
        except ValueError as e:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(str(e))
            return Empty()
