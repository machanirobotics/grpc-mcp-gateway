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

// ParseTransports splits a comma-separated transport string (e.g.
// "stdio,streamable-http") into a []Transport slice.
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
		mux := buildHTTPMux(httpServer, cfg, httpTransports)
		
		if hasStdio {
			go func() {
				if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
					log.Printf("runtime: HTTP server error: %v", err)
				}
			}()
		} else {
			return http.ListenAndServe(cfg.Addr, mux)
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

// Endpoint represents an MCP server endpoint.
type Endpoint struct {
	// Protocol is "stdio", "http", or "https".
	Protocol  string
	Transport string
	// URL is the full endpoint URL. Empty for stdio.
	URL string
}

// ResolveBasePath returns the effective BasePath for a given config and generated default.
// If cfg.BasePath is empty, it returns the generatedDefault; otherwise it returns cfg.BasePath.
func ResolveBasePath(cfg *MCPServerConfig, generatedDefault string) string {
	if cfg.BasePath == "" {
		return generatedDefault
	}
	return cfg.BasePath
}

// PreferGeneratedBasePath returns the generated default even if cfg.BasePath is set.
// This is useful when you want the proto-derived path to take precedence.
func PreferGeneratedBasePath(generatedDefault string, cfg *MCPServerConfig) string {
	return generatedDefault
}

// ServerEndpoint returns the endpoint for an MCP server based on its config.
// For stdio transport, it returns Endpoint{Protocol: "stdio", URL: ""}.
// For HTTP transports, it returns Endpoint{Protocol: "http|https", URL: "http://host:port/path"}.
func ServerEndpoint(cfg *MCPServerConfig) (*Endpoint, error) {
	// Detect if we're in stdio-only mode.
	hasStdio := cfg.Transport == TransportStdio || (len(cfg.Transports) > 0 && func() bool {
		for _, t := range cfg.Transports {
			if t == TransportStdio {
				return true
			}
		}
		return false
	}())
	hasHTTP := cfg.Transport == TransportStreamableHTTP || cfg.Transport == TransportSSE || func() bool {
		for _, t := range cfg.Transports {
			if t == TransportStreamableHTTP || t == TransportSSE {
				return true
			}
		}
		return false
	}()

	if hasStdio && !hasHTTP {
		return &Endpoint{Protocol: "stdio", Transport: string(TransportStdio), URL: ""}, nil
	}

	// Determine the primary HTTP transport name.
	transportName := string(TransportStreamableHTTP)
	for _, t := range cfg.Transports {
		if t == TransportSSE || t == TransportStreamableHTTP {
			transportName = string(t)
			break
		}
	}
	if cfg.Transport != "" && len(cfg.Transports) == 0 {
		transportName = string(cfg.Transport)
	}

	// HTTP endpoint
	addr := cfg.Addr
	if addr == "" {
		addr = ":8080"
	}

	// Resolve listen address for external access
	host := os.Getenv("MCP_SERVER_HOST")
	port := os.Getenv("MCP_SERVER_PORT")
	
	if host == "" || port == "" {
		if strings.HasPrefix(addr, ":") {
			// :6503 -> localhost:6503
			if host == "" {
				host = "localhost"
			}
			if port == "" {
				port = strings.TrimPrefix(addr, ":")
			}
		} else {
			// host:port or just port
			h, p, found := strings.Cut(addr, ":")
			if found && p != "" {
				// host:port format
				if host == "" {
					host = h
				}
				if port == "" {
					port = p
				}
			} else {
				// no port found, use defaults
				if host == "" {
					host = "localhost"
				}
				if port == "" {
					port = "8080"
				}
			}
		}
	}

	protocol := "http"
	if os.Getenv("MCP_SERVER_TLS") == "true" {
		protocol = "https"
	}

	path := cfg.BasePath
	if cfg.GeneratedBasePath != "" {
		path = cfg.GeneratedBasePath
	} else if path == "" {
		path = "/mcp"
	}

	return &Endpoint{
		Protocol:  protocol,
		Transport: transportName,
		URL:       fmt.Sprintf("%s://%s:%s%s", protocol, host, port, path),
	}, nil
}
