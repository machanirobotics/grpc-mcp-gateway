//! TodoService combined **gRPC + MCP** server example in Rust.
//!
//! | Endpoint | Default                                            |
//! |----------|----------------------------------------------------|
//! | gRPC     | `[::]:50051`                                       |
//! | MCP      | `0.0.0.0:8082` (streamable-http) or stdio          |
//!
//! Environment variables: `MCP_TRANSPORT`, `MCP_HOST`, `MCP_PORT`.

mod proto;

use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::{Arc, Mutex};
use std::time::SystemTime;

use async_trait::async_trait;
use rmcp::ErrorData as McpError;
use serde_json::{json, Value};

use proto::todo_v1::{
    self as pb,
    todo_service_server::{TodoService, TodoServiceServer},
};

#[path = "../../proto/generated/rust/todo/v1/todo_service.mcp.rs"]
#[allow(dead_code)]
mod todo_service_mcp;
use todo_service_mcp::TodoServiceMcpServer;

/// Shared in-memory store. `Clone` is cheap (inner data behind `Arc`).
#[derive(Clone)]
pub struct TodoServer {
    todos: Arc<Mutex<HashMap<String, pb::Todo>>>,
}

impl TodoServer {
    pub fn new() -> Self {
        Self {
            todos: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    /// Current wall-clock as a prost `Timestamp`.
    fn now() -> prost_types::Timestamp {
        let d = SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)
            .unwrap();
        prost_types::Timestamp {
            seconds: d.as_secs() as i64,
            nanos: d.subsec_nanos() as i32,
        }
    }

    /// Serialise an optional timestamp to a simple `"<epoch>Z"` string.
    fn ts_str(ts: &Option<prost_types::Timestamp>) -> String {
        ts.as_ref()
            .map_or(String::new(), |t| format!("{}Z", t.seconds))
    }

    /// Convert a prost `Todo` into a JSON `Value` for MCP responses.
    fn todo_to_json(t: &pb::Todo) -> Value {
        json!({ "name": t.name, "title": t.title, "description": t.description,
                "completed": t.completed,
                "priority": pb::Priority::try_from(t.priority).map(|v| v.as_str_name()).unwrap_or("PRIORITY_UNSPECIFIED"),
                "create_time": Self::ts_str(&t.create_time), "update_time": Self::ts_str(&t.update_time) })
    }

    /// Parse a priority enum string (e.g. `"PRIORITY_HIGH"`) into its i32 value.
    fn pri(s: &str) -> i32 {
        pb::Priority::from_str_name(s)
            .map(|v| v as i32)
            .unwrap_or(0)
    }

    /// Apply partial-update fields from `src` onto `dst`, then bump `update_time`.
    fn apply_update(dst: &mut pb::Todo, src: &pb::Todo) {
        if !src.title.is_empty() {
            dst.title = src.title.clone();
        }
        if !src.description.is_empty() {
            dst.description = src.description.clone();
        }
        dst.completed = src.completed;
        dst.priority = src.priority;
        dst.update_time = Some(Self::now());
    }
}

// -- gRPC (tonic) implementation ---------------------------------------------

#[async_trait]
impl TodoService for TodoServer {
    async fn create_todo(
        &self,
        req: tonic::Request<pb::CreateTodoRequest>,
    ) -> Result<tonic::Response<pb::Todo>, tonic::Status> {
        let r = req.into_inner();
        let name = format!("{}/todos/{}", r.parent, r.todo_id);
        let now = Self::now();
        let mut todo = r.todo.unwrap_or_default();
        todo.name = name.clone();
        todo.create_time = Some(now.clone());
        todo.update_time = Some(now);
        self.todos.lock().unwrap().insert(name, todo.clone());
        Ok(tonic::Response::new(todo))
    }

    async fn get_todo(
        &self,
        req: tonic::Request<pb::GetTodoRequest>,
    ) -> Result<tonic::Response<pb::Todo>, tonic::Status> {
        let name = req.into_inner().name;
        let todos = self.todos.lock().unwrap();
        todos
            .get(&name)
            .cloned()
            .map(tonic::Response::new)
            .ok_or_else(|| tonic::Status::not_found(format!("todo {name:?} not found")))
    }

    async fn list_todos(
        &self,
        _: tonic::Request<pb::ListTodosRequest>,
    ) -> Result<tonic::Response<pb::ListTodosResponse>, tonic::Status> {
        let todos = self.todos.lock().unwrap();
        Ok(tonic::Response::new(pb::ListTodosResponse {
            todos: todos.values().cloned().collect(),
            next_page_token: String::new(),
        }))
    }

    async fn update_todo(
        &self,
        req: tonic::Request<pb::UpdateTodoRequest>,
    ) -> Result<tonic::Response<pb::Todo>, tonic::Status> {
        let upd = req
            .into_inner()
            .todo
            .ok_or_else(|| tonic::Status::invalid_argument("missing todo"))?;
        let mut todos = self.todos.lock().unwrap();
        let existing = todos
            .get_mut(&upd.name)
            .ok_or_else(|| tonic::Status::not_found(format!("todo {:?} not found", upd.name)))?;
        Self::apply_update(existing, &upd);
        Ok(tonic::Response::new(existing.clone()))
    }

    async fn delete_todo(
        &self,
        req: tonic::Request<pb::DeleteTodoRequest>,
    ) -> Result<tonic::Response<()>, tonic::Status> {
        let name = req.into_inner().name;
        self.todos
            .lock()
            .unwrap()
            .remove(&name)
            .map(|_| tonic::Response::new(()))
            .ok_or_else(|| tonic::Status::not_found(format!("todo {name:?} not found")))
    }
}

// -- MCP implementation (JSON ↔ prost bridge) --------------------------------

#[async_trait]
impl TodoServiceMcpServer for TodoServer {
    async fn create_todo(&self, args: Value) -> Result<Value, McpError> {
        let (parent, tid) = (
            args["parent"].as_str().unwrap_or_default(),
            args["todo_id"].as_str().unwrap_or_default(),
        );
        let name = format!("{parent}/todos/{tid}");
        let now = Self::now();
        let j = args.get("todo");
        let todo = pb::Todo {
            name: name.clone(),
            title: j
                .and_then(|v| v["title"].as_str())
                .unwrap_or_default()
                .into(),
            description: j
                .and_then(|v| v["description"].as_str())
                .unwrap_or_default()
                .into(),
            completed: j.and_then(|v| v["completed"].as_bool()).unwrap_or(false),
            priority: j
                .and_then(|v| v["priority"].as_str())
                .map(Self::pri)
                .unwrap_or(0),
            create_time: Some(now.clone()),
            update_time: Some(now),
        };
        self.todos.lock().unwrap().insert(name, todo.clone());
        Ok(Self::todo_to_json(&todo))
    }

    async fn get_todo(&self, args: Value) -> Result<Value, McpError> {
        let name = args["name"].as_str().unwrap_or_default();
        let todos = self.todos.lock().unwrap();
        todos
            .get(name)
            .map(Self::todo_to_json)
            .ok_or_else(|| McpError::invalid_params(format!("todo {name:?} not found"), None))
    }

    async fn list_todos(&self, _: Value) -> Result<Value, McpError> {
        let todos = self.todos.lock().unwrap();
        Ok(
            json!({ "todos": todos.values().map(Self::todo_to_json).collect::<Vec<_>>(), "next_page_token": "" }),
        )
    }

    async fn update_todo(&self, args: Value) -> Result<Value, McpError> {
        let j = args
            .get("todo")
            .ok_or_else(|| McpError::invalid_params("missing 'todo'", None))?;
        let name = j["name"].as_str().unwrap_or_default();
        let mut todos = self.todos.lock().unwrap();
        let e = todos
            .get_mut(name)
            .ok_or_else(|| McpError::invalid_params(format!("todo {name:?} not found"), None))?;
        if let Some(s) = j["title"].as_str() {
            if !s.is_empty() {
                e.title = s.into();
            }
        }
        if let Some(s) = j["description"].as_str() {
            if !s.is_empty() {
                e.description = s.into();
            }
        }
        if let Some(b) = j["completed"].as_bool() {
            e.completed = b;
        }
        if let Some(s) = j["priority"].as_str() {
            e.priority = Self::pri(s);
        }
        e.update_time = Some(Self::now());
        Ok(Self::todo_to_json(e))
    }

    async fn delete_todo(&self, args: Value) -> Result<Value, McpError> {
        let name = args["name"].as_str().unwrap_or_default();
        self.todos
            .lock()
            .unwrap()
            .remove(name)
            .map(|_| json!({}))
            .ok_or_else(|| McpError::invalid_params(format!("todo {name:?} not found"), None))
    }
}

// -- Entrypoint: gRPC + MCP side-by-side -------------------------------------

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let srv = TodoServer::new();

    let transport = std::env::var("MCP_TRANSPORT").unwrap_or_else(|_| "streamable-http".into());
    let host = std::env::var("MCP_HOST").unwrap_or_else(|_| "0.0.0.0".into());
    let port: u16 = std::env::var("MCP_PORT")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(8082);

    // gRPC in background task (with reflection).
    let grpc_addr: SocketAddr = "[::]:50051".parse()?;
    let grpc_srv = srv.clone();
    tokio::spawn(async move {
        const FILE_DESCRIPTOR_SET: &[u8] =
            include_bytes!("../../proto/generated/rust/descriptor.binpb");
        let reflection_service = tonic_reflection::server::Builder::configure()
            .register_encoded_file_descriptor_set(FILE_DESCRIPTOR_SET)
            .build_v1()
            .expect("failed to build reflection service");
        eprintln!("gRPC listening on {grpc_addr} (reflection enabled)");
        tonic::transport::Server::builder()
            .add_service(TodoServiceServer::new(grpc_srv))
            .add_service(reflection_service)
            .serve(grpc_addr)
            .await
            .expect("gRPC server failed");
    });

    // MCP (blocks).
    eprintln!("MCP starting (transport={transport}, {host}:{port})");
    todo_service_mcp::serve_todo_service_mcp(
        srv,
        todo_service_mcp::TodoServiceMcpTransportConfig {
            transport,
            host,
            port,
            base_path: "/todo/v1/todoservice".into(),
        },
    )
    .await
}
