package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const healthCheckTimeout = 5 * time.Second

// mcpPingResult is the MCP ping response format per
// https://modelcontextprotocol.io/specification/2025-03-26/basic/utilities/ping
type mcpPingResult struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *mcpError   `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HealthCheckHandler returns an HTTP handler that performs a gRPC health check
// and responds with MCP ping format: {"jsonrpc":"2.0","id":"health","result":{}}
// when the backend reports SERVING, or an error object when unhealthy.
// Use for load balancer / k8s probes and MCP clients that expect ping-style responses.
func HealthCheckHandler(conn *grpc.ClientConn) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
		defer cancel()
		client := healthpb.NewHealthClient(conn)
		resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: ""})
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(mcpPingResult{
				JSONRPC: "2.0",
				ID:      "health",
				Error:   &mcpError{Code: -32000, Message: "service unavailable"},
			})
			return
		}
		if resp.Status != healthpb.HealthCheckResponse_SERVING {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(mcpPingResult{
				JSONRPC: "2.0",
				ID:      "health",
				Error:   &mcpError{Code: -32000, Message: "not serving"},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mcpPingResult{
			JSONRPC: "2.0",
			ID:      "health",
			Result:  map[string]interface{}{},
		})
	})
}
