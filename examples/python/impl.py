"""In-memory TodoService implementation."""

from __future__ import annotations

from datetime import datetime, timezone

from google.protobuf.empty_pb2 import Empty
from google.protobuf.timestamp_pb2 import Timestamp

from todo.v1 import todo_pb2


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _now() -> Timestamp:
    """Return the current UTC time as a protobuf Timestamp."""
    ts = Timestamp()
    ts.FromDatetime(datetime.now(timezone.utc))
    return ts


# ---------------------------------------------------------------------------
# Implementation
# ---------------------------------------------------------------------------

class TodoServer:
    """Implements the TodoServiceMCPServer protocol with an in-memory store."""

    def __init__(self) -> None:
        self._todos: dict[str, todo_pb2.Todo] = {}

    async def create_todo(
        self, request: todo_pb2.CreateTodoRequest
    ) -> todo_pb2.Todo:
        name = f"{request.parent}/todos/{request.todo_id}"
        now = _now()

        todo = todo_pb2.Todo()
        if request.HasField("todo"):
            todo.CopyFrom(request.todo)
        todo.name = name
        todo.create_time.CopyFrom(now)
        todo.update_time.CopyFrom(now)

        self._todos[name] = todo
        return todo

    async def get_todo(
        self, request: todo_pb2.GetTodoRequest
    ) -> todo_pb2.Todo:
        todo = self._todos.get(request.name)
        if todo is None:
            raise ValueError(f"todo {request.name!r} not found")
        return todo

    async def list_todos(
        self, request: todo_pb2.ListTodosRequest
    ) -> todo_pb2.ListTodosResponse:
        resp = todo_pb2.ListTodosResponse()
        for todo in self._todos.values():
            resp.todos.append(todo)
        return resp

    async def update_todo(
        self, request: todo_pb2.UpdateTodoRequest
    ) -> todo_pb2.Todo:
        existing = self._todos.get(request.todo.name)
        if existing is None:
            raise ValueError(f"todo {request.todo.name!r} not found")

        if request.todo.title:
            existing.title = request.todo.title
        if request.todo.description:
            existing.description = request.todo.description
        existing.completed = request.todo.completed
        existing.priority = request.todo.priority
        existing.update_time.CopyFrom(_now())

        return existing

    async def delete_todo(
        self, request: todo_pb2.DeleteTodoRequest
    ) -> Empty:
        if request.name not in self._todos:
            raise ValueError(f"todo {request.name!r} not found")
        del self._todos[request.name]
        return Empty()
