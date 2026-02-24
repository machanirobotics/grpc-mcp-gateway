pub mod mcp {
    pub mod protobuf {
        include!("gen/mcp/protobuf/mcp.protobuf.rs");
    }
}

pub use mcp::protobuf::*;
