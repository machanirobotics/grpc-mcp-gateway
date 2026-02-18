package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/machanirobotics/protoc-mcp-gen/examples/proto/generated/go/todo/todopbv1"
	"github.com/machanirobotics/protoc-mcp-gen/pkg/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	transports := runtime.ParseTransports("streamable-http")
	if t := os.Getenv("MCP_TRANSPORT"); t != "" {
		transports = runtime.ParseTransports(t)
	}
	mcpAddr := ":8082"
	if a := os.Getenv("MCP_ADDR"); a != "" {
		mcpAddr = a
	}
	grpcAddr := ":50051"
	if a := os.Getenv("GRPC_ADDR"); a != "" {
		grpcAddr = a
	}

	srv := newTodoServer()

	// Start gRPC server in a goroutine.
	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("gRPC listen: %v", err)
		}
		gs := grpc.NewServer()
		todopbv1.RegisterTodoServiceServer(gs, srv)
		reflection.Register(gs)
		log.Printf("gRPC server listening on %s (reflection enabled)", grpcAddr)
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("gRPC serve: %v", err)
		}
	}()

	// Start MCP server (blocks).
	cfg := &runtime.MCPServerConfig{
		Name:       "todo-mcp-example",
		Version:    "0.1.0",
		Transports: transports,
		Addr:       mcpAddr,
	}
	log.Printf("MCP server listening on %s (transports=%v)", mcpAddr, transports)
	if err := todopbv1.ServeTodoServiceMCP(context.Background(), srv, cfg); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
