package runtime

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/status"
)

// grpcError is a lightweight JSON-serialisable representation of a gRPC error.
type grpcError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details,omitempty"`
}

// HandleError converts a gRPC or ConnectRPC error into an MCP tool error result.
// Use it in tool handlers when your gRPC call fails:
//
//	resp, err := srv.CreateTodo(ctx, req)
//	if err != nil {
//	    return runtime.HandleError(err)
//	}
//
// If err is nil both return values are nil. gRPC status codes are preserved
// in the JSON error payload.
func HandleError(err error) (*mcp.CallToolResult, error) {
	if err == nil {
		return nil, nil
	}

	// Try gRPC status.
	if st, ok := status.FromError(err); ok {
		return errorFromGRPC(st), nil
	}

	// Fall back to plain error text.
	return ErrorResult(err.Error()), nil
}

func errorFromGRPC(st *status.Status) *mcp.CallToolResult {
	e := grpcError{
		Code:    st.Code().String(),
		Message: st.Message(),
	}
	e.Details = append(e.Details, st.Details()...)
	return marshalErrorResult(e)
}

func marshalErrorResult(e grpcError) *mcp.CallToolResult {
	b, err := json.Marshal(e)
	if err != nil {
		return ErrorResult(e.Message)
	}
	return ErrorResult(string(b))
}
