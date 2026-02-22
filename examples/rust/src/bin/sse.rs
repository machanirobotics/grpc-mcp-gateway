//! TodoService MCP server over SSE (Server-Sent Events).
//!
//! Usage:
//!     cargo run --bin sse

use todo_mcp_example::todo_impl::TodoServer;
use todo_mcp_example::todo_service_mcp;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let host = std::env::var("MCP_HOST").unwrap_or_else(|_| "0.0.0.0".into());
    let port: u16 = std::env::var("MCP_PORT")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(8083);

    eprintln!("MCP starting (transport=sse, {host}:{port})");
    todo_service_mcp::serve_todo_service_mcp(
        TodoServer::new(),
        todo_service_mcp::TodoServiceMcpTransportConfig {
            transport: "sse".into(),
            host,
            port,
            base_path: "/todo/v1/todoservice".into(),
        },
    )
    .await
}
