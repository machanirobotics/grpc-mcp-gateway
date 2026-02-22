module github.com/machanirobotics/grpc-mcp-gateway/plugin

go 1.25.6

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20260209202127-80ab13bee0bf.1
	github.com/machanirobotics/grpc-mcp-gateway v0.0.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260217215200-42d3e9bedb6d
	google.golang.org/protobuf v1.36.11
)

replace github.com/machanirobotics/grpc-mcp-gateway => ../
