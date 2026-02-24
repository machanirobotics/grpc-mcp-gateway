package runtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServerConfig holds configuration for starting an MCP server.
// Set Transports (or Transport) to choose one or more wire protocols.
// When multiple transports are specified they run concurrently in the
// same process -- e.g. ["stdio", "streamable-http"].
type MCPServerConfig struct {
	// Name is the MCP server name reported during initialization.
	Name string
	// Version is the MCP server version reported during initialization.
	Version string
	// Transport selects a single wire protocol (for backward compatibility).
	// Ignored when Transports is non-empty.
	Transport Transport
	// Transports selects one or more wire protocols to serve concurrently.
	// Takes precedence over Transport.
	Transports []Transport
	// Addr is the listen address for HTTP-based transports (default ":8080").
	Addr string
	// ServerOptions are passed to mcp.NewServer.
	ServerOptions *mcp.ServerOptions
	// StreamableHTTPOptions are passed to mcp.NewStreamableHTTPHandler.
	StreamableHTTPOptions *mcp.StreamableHTTPOptions
	// SSEOptions are passed to mcp.NewSSEHandler.
	SSEOptions *mcp.SSEOptions
	// BasePath is the HTTP path prefix for the MCP endpoint (default "/mcp").
	BasePath string
	// GeneratedBasePath is the proto-derived default BasePath. If set, it takes precedence over BasePath.
	GeneratedBasePath string
	// HeaderMappings configures HTTP header to gRPC metadata forwarding.
	// Each entry maps an HTTP header name to a gRPC metadata key.
	// Use DefaultHeaderMappings() for common headers.
	HeaderMappings []HeaderMapping
	// OnReady is called after BasePath is resolved, just before the server starts listening.
	// Use this to log or inspect the final endpoint.
	OnReady func(cfg *MCPServerConfig)
}

// NewMCPServer creates an mcp.Server from a MCPServerConfig.
func NewMCPServer(cfg *MCPServerConfig) *mcp.Server {
	opts := cfg.ServerOptions
	if opts == nil {
		opts = &mcp.ServerOptions{}
	}
	return mcp.NewServer(&mcp.Implementation{Name: cfg.Name, Version: cfg.Version}, opts)
}

// ParseTransports splits a comma-separated transport string into a []Transport slice.
// Use with MCP_TRANSPORT env var:
//
//	transports := runtime.ParseTransports(os.Getenv("MCP_TRANSPORT"))
//	if len(transports) == 0 {
//	    transports = []runtime.Transport{runtime.TransportStreamableHTTP}
//	}
func ParseTransports(s string) []Transport {
	parts := strings.Split(s, ",")
	out := make([]Transport, 0, len(parts))
	for _, p := range parts {
		if t := Transport(strings.TrimSpace(p)); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// StartServer starts the MCP server using the configured transport(s).
// Multiple transports run concurrently -- HTTP-based transports share a
// single net/http server while stdio gets its own mcp.Server instance.
// This call blocks until the context is cancelled or an error occurs.
func StartServer(ctx context.Context, cfg *MCPServerConfig, register func(s *mcp.Server)) error {
	transports := cfg.Transports
	if len(transports) == 0 {
		t := cfg.Transport
		if t == "" {
			t = TransportStreamableHTTP
		}
		transports = []Transport{t}
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.GeneratedBasePath != "" {
		cfg.BasePath = cfg.GeneratedBasePath
	} else if cfg.BasePath == "" {
		cfg.BasePath = "/mcp"
	}

	var httpTransports []Transport
	hasStdio := false
	for _, t := range transports {
		switch t {
		case TransportStdio:
			hasStdio = true
		case TransportSSE, TransportStreamableHTTP:
			httpTransports = append(httpTransports, t)
		default:
			return fmt.Errorf("runtime: unsupported transport %q", t)
		}
	}

	// Notify caller that BasePath is resolved.
	if cfg.OnReady != nil {
		cfg.OnReady(cfg)
	}

	// Start HTTP transport(s) if requested.
	if len(httpTransports) > 0 {
		httpServer := NewMCPServer(cfg)
		register(httpServer)
		var handler http.Handler = buildHTTPMux(httpServer, cfg, httpTransports)
		handler = HeadersMiddleware(cfg.HeaderMappings, handler)

		if hasStdio {
			go func() {
				if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
					log.Printf("runtime: HTTP server error: %v", err)
				}
			}()
		} else {
			return http.ListenAndServe(cfg.Addr, handler)
		}
	}

	if hasStdio {
		stdioServer := NewMCPServer(cfg)
		register(stdioServer)
		return serveStdio(ctx, stdioServer)
	}
	return fmt.Errorf("runtime: no transports configured")
}

// buildHTTPMux registers HTTP-based transports on a shared ServeMux.
func buildHTTPMux(server *mcp.Server, cfg *MCPServerConfig, transports []Transport) *http.ServeMux {
	mux := http.NewServeMux()
	for _, t := range transports {
		switch t {
		case TransportStreamableHTTP:
			h := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return server }, cfg.StreamableHTTPOptions)
			mux.Handle(cfg.BasePath, h)
		case TransportSSE:
			h := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server { return server }, cfg.SSEOptions)
			mux.Handle(cfg.BasePath+"/", h)
		}
	}
	return mux
}

func serveStdio(ctx context.Context, server *mcp.Server) error {
	log.SetOutput(os.Stderr)
	return server.Run(ctx, &mcp.StdioTransport{})
}

