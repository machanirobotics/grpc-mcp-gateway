package runtime

import (
	"fmt"
	"os"
	"strings"
)

// Endpoint represents an MCP server endpoint.
// Use ServerEndpoint to compute it from MCPServerConfig.
type Endpoint struct {
	Protocol  string // "stdio", "http", or "https"
	Transport string // "stdio", "streamable-http", or "sse"
	URL       string // Full URL (e.g. "http://localhost:8082/todo/v1/todoservice/mcp"). Empty for stdio.
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
// Use to log the URL before starting:
//
//	if ep, err := runtime.ServerEndpoint(cfg); err == nil {
//	    log.Printf("MCP listening on %s", ep.URL)
//	}
//
// For stdio transport, URL is empty. For HTTP, host/port come from Addr or
// MCP_SERVER_HOST, MCP_SERVER_PORT env vars.
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
			if host == "" {
				host = "localhost"
			}
			if port == "" {
				port = strings.TrimPrefix(addr, ":")
			}
		} else {
			h, p, found := strings.Cut(addr, ":")
			if found && p != "" {
				if host == "" {
					host = h
				}
				if port == "" {
					port = p
				}
			} else {
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
