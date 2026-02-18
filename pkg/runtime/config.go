package runtime

// Transport represents the transport protocol for the MCP server.
type Transport string

const (
	// TransportStreamableHTTP is the modern Streamable HTTP transport (default).
	TransportStreamableHTTP Transport = "streamable-http"
	// TransportSSE is the legacy SSE transport (2024-11-05 spec).
	TransportSSE Transport = "sse"
	// TransportStdio runs the MCP server over stdin/stdout.
	TransportStdio Transport = "stdio"
)

// Option is a functional option for configuring MCP handler registration.
type Option func(*Config)

// Config holds runtime configuration for MCP handlers.
type Config struct {
	ExtraProperties []ExtraProperty
	Transport       Transport
	Addr            string
}

// ExtraProperty defines an additional property to inject into tool schemas
// and extract from request arguments into context.
type ExtraProperty struct {
	Name        string
	Description string
	Required    bool
	ContextKey  any
}

// WithExtraProperties returns an Option that adds extra properties to tool
// schemas. These properties are extracted from incoming requests and placed
// into the context using the specified ContextKey.
func WithExtraProperties(properties ...ExtraProperty) Option {
	return func(c *Config) {
		c.ExtraProperties = append(c.ExtraProperties, properties...)
	}
}

// WithTransport sets the transport protocol for the MCP server.
func WithTransport(t Transport) Option {
	return func(c *Config) {
		c.Transport = t
	}
}

// WithAddr sets the listen address (for HTTP-based transports).
// Defaults to ":8080" if not set.
func WithAddr(addr string) Option {
	return func(c *Config) {
		c.Addr = addr
	}
}

// ApplyOptions creates a Config and applies all provided options.
func ApplyOptions(opts ...Option) *Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
