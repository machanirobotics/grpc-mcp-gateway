package generator

// MCPServiceOpts is the language-neutral view of MCPServiceOptions for templates.
type MCPServiceOpts struct {
	App       *MCPAppOpts
	Resources []MCPResourceOpts
}

// MCPMethodOpts is the language-neutral view of per-RPC MCP options for templates.
type MCPMethodOpts struct {
	ToolName        string
	ToolDescription string
	Prompt          *MCPPromptOpts
	Elicitation     *MCPElicitationOpts
}

// MCPAppOpts mirrors MCPApp for templates.
type MCPAppOpts struct {
	Name        string
	Version     string
	Description string
}

// MCPPromptOpts mirrors MCPPrompt for templates.
// Arguments are derived from the proto message referenced by Schema.
type MCPPromptOpts struct {
	Name        string
	Description string
	Schema      string
	Arguments   []MCPPromptArgOpts
}

// MCPPromptArgOpts describes a single prompt argument resolved from a schema message.
type MCPPromptArgOpts struct {
	Name        string
	Description string
	Required    bool
	Type        string
	EnumValues  []string
}

// MCPResourceOpts mirrors MCPResource for templates.
type MCPResourceOpts struct {
	URI         string
	URITemplate string
	Name        string
	Description string
	MimeType    string
}

// MCPElicitationOpts mirrors MCPElicitation for templates.
// Fields are derived from the proto message referenced by Schema.
type MCPElicitationOpts struct {
	Message string
	Schema  string
	Fields  []MCPElicitFieldOpts
}

// MCPElicitFieldOpts describes a single elicitation field resolved from a schema message.
type MCPElicitFieldOpts struct {
	Name        string
	Description string
	Required    bool
	Type        string
	EnumValues  []string
}
