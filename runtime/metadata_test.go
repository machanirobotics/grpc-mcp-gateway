package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestForwardMetadata_IncomingGRPC(t *testing.T) {
	// Simulate incoming gRPC metadata.
	md := metadata.Pairs(
		"authorization", "Bearer token123",
		"x-request-id", "req-abc",
		"grpc-timeout", "5s", // should be filtered
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	ctx = ForwardMetadata(ctx)

	outMD, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got := outMD.Get("authorization"); len(got) == 0 || got[0] != "Bearer token123" {
		t.Errorf("authorization: got %v", got)
	}
	if got := outMD.Get("x-request-id"); len(got) == 0 || got[0] != "req-abc" {
		t.Errorf("x-request-id: got %v", got)
	}
	if got := outMD.Get("grpc-timeout"); len(got) != 0 {
		t.Errorf("grpc-timeout should be filtered, got %v", got)
	}
}

func TestForwardMetadata_HTTPHeaders(t *testing.T) {
	mappings := []HeaderMapping{
		{HTTPHeader: "Authorization", GRPCKey: "authorization"},
		{HTTPHeader: "X-Tenant-Id", GRPCKey: "x-tenant-id"},
	}

	// Simulate HTTP request with headers flowing through middleware.
	var capturedCtx context.Context
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
	})
	handler := HeadersMiddleware(mappings, inner)

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-Tenant-Id", "tenant-42")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if capturedCtx == nil {
		t.Fatal("handler was not called")
	}

	// Now ForwardMetadata should pick up the HTTP headers.
	ctx := ForwardMetadata(capturedCtx)
	outMD, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got := outMD.Get("authorization"); len(got) == 0 || got[0] != "Bearer secret" {
		t.Errorf("authorization: got %v", got)
	}
	if got := outMD.Get("x-tenant-id"); len(got) == 0 || got[0] != "tenant-42" {
		t.Errorf("x-tenant-id: got %v", got)
	}
}

func TestForwardMetadata_MergesBoth(t *testing.T) {
	// Incoming gRPC metadata.
	grpcMD := metadata.Pairs("x-request-id", "from-grpc")
	ctx := metadata.NewIncomingContext(context.Background(), grpcMD)

	// HTTP headers (should override gRPC for same key).
	ctx = context.WithValue(ctx, httpHeadersKey, map[string]string{
		"authorization": "Bearer http-token",
		"x-request-id":  "from-http", // overrides gRPC value
	})

	ctx = ForwardMetadata(ctx)

	outMD, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got := outMD.Get("authorization"); len(got) == 0 || got[0] != "Bearer http-token" {
		t.Errorf("authorization: got %v", got)
	}
	// HTTP should win over gRPC for x-request-id.
	if got := outMD.Get("x-request-id"); len(got) == 0 || got[0] != "from-http" {
		t.Errorf("x-request-id: got %v, want from-http", got)
	}
}

func TestForwardMetadata_NoHeaders(t *testing.T) {
	ctx := context.Background()
	result := ForwardMetadata(ctx)
	// Should return same context unchanged.
	if _, ok := metadata.FromOutgoingContext(result); ok {
		t.Error("expected no outgoing metadata on empty context")
	}
}
