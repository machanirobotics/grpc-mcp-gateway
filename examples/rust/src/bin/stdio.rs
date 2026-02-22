//! TodoService MCP server over stdio (for CLI tools like Claude Desktop).
//!
//! Usage:
//!     cargo run --bin stdio

use todo_mcp_example::todo_impl::TodoServer;
use todo_mcp_example::todo_service_mcp;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    todo_service_mcp::serve_todo_service_mcp_stdio(TodoServer::new()).await
}
