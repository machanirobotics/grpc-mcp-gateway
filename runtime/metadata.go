package runtime

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

// HeaderMapping maps an HTTP header name to a gRPC metadata key.
// Used with MCPServerConfig.HeaderMappings to forward headers from MCP HTTP
// requests into gRPC outgoing metadata. Use DefaultHeaderMappings() for common ones.
type HeaderMapping struct {
	HTTPHeader string // HTTP header name to read (case-insensitive)
	GRPCKey    string // gRPC metadata key to write (use lowercase)
}

// httpHeadersKey is the context key for storing extracted HTTP headers.
type httpHeadersKeyType struct{}

var httpHeadersKey = httpHeadersKeyType{}

// ForwardMetadata prepares gRPC outgoing metadata on the context by combining:
//
//  1. Incoming gRPC metadata (for gRPC→gRPC proxy scenarios) — all keys
//     except reserved "grpc-" prefixed ones are forwarded automatically.
//  2. HTTP headers stored by HeadersMiddleware (for HTTP→gRPC scenarios) —
//     custom header mappings configured at runtime.
//
// HTTP-extracted headers take precedence over incoming gRPC metadata for the
// same key. This function is called by generated ForwardTo code before every
// gRPC client call.
func ForwardMetadata(ctx context.Context) context.Context {
	md := metadata.MD{}

	// 1. Copy incoming gRPC metadata (proxy pass-through).
	if incoming, ok := metadata.FromIncomingContext(ctx); ok {
		for k, vals := range incoming {
			key := strings.ToLower(k)
			if strings.HasPrefix(key, "grpc-") {
				continue // reserved by gRPC
			}
			md[key] = append(md[key], vals...)
		}
	}

	// 2. Merge HTTP headers stored by HeadersMiddleware (runtime custom headers).
	if pairs, ok := ctx.Value(httpHeadersKey).(map[string]string); ok {
		for k, v := range pairs {
			md.Set(strings.ToLower(k), v) // overwrites duplicates
		}
	}

	if len(md) == 0 {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// HeadersMiddleware returns HTTP middleware that extracts configured headers
// from the incoming request and stores them in the request context.
// These headers are later available via ForwardMetadata.
func HeadersMiddleware(mappings []HeaderMapping, next http.Handler) http.Handler {
	if len(mappings) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pairs := make(map[string]string, len(mappings))
		for _, m := range mappings {
			if v := r.Header.Get(m.HTTPHeader); v != "" {
				pairs[m.GRPCKey] = v
			}
		}
		if len(pairs) > 0 {
			ctx := context.WithValue(r.Context(), httpHeadersKey, pairs)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// DefaultHeaderMappings returns commonly forwarded header mappings:
// Authorization, X-Request-ID, and X-Trace-ID.
func DefaultHeaderMappings() []HeaderMapping {
	return []HeaderMapping{
		{HTTPHeader: "Authorization", GRPCKey: "authorization"},
		{HTTPHeader: "X-Request-Id", GRPCKey: "x-request-id"},
		{HTTPHeader: "X-Trace-Id", GRPCKey: "x-trace-id"},
	}
}
