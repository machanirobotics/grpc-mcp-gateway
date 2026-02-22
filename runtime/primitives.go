package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DefaultPromptHandler returns a prompt handler that produces a single user
// message containing the prompt description. It is used as a placeholder for
// prompts declared via MCP proto options. Replace it by calling
// server.RemovePrompts / server.AddPrompt with your own handler.
func DefaultPromptHandler(description string) func(context.Context, *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: description,
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: description},
				},
			},
		}, nil
	}
}

// DefaultResourceHandler returns a resource handler that returns an empty JSON
// object. It is used as a placeholder for resources declared via MCP proto
// options. Replace it by calling server.RemoveResources / server.AddResource
// with your own handler.
func DefaultResourceHandler() func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: req.Params.URI, Text: "{}"},
			},
		}, nil
	}
}


// AppResourceURI returns the canonical ui:// resource URI for a service app.
func AppResourceURI(serviceName string) string {
	return fmt.Sprintf("ui://%s/app.html", strings.ToLower(serviceName))
}

// SetToolAppMeta returns a shallow clone of tool with _meta.ui.resourceUri set,
// which makes the tool show up as an MCP App in supporting hosts.
func SetToolAppMeta(tool *mcp.Tool, resourceURI string) *mcp.Tool {
	cloned := *tool
	cloned.Meta = mcp.Meta{
		"ui": map[string]any{
			"resourceUri": resourceURI,
		},
	}
	return &cloned
}

// DefaultAppHTML returns a minimal HTML page for an MCP App placeholder.
func DefaultAppHTML(appName, version, description string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>%s</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 600px; margin: 40px auto; padding: 0 20px; color: #333; }
  h1 { font-size: 1.5rem; } p { color: #666; } .version { font-size: 0.85rem; color: #999; }
</style>
</head>
<body>
  <h1>%s</h1>
  <p class="version">v%s</p>
  <p>%s</p>
  <p>This is a generated MCP App placeholder. Replace this resource with your own UI.</p>
</body>
</html>`, appName, appName, version, description)
}

// DefaultAppResourceHandler returns a resource handler that serves the default
// app HTML page. The returned handler is suitable for registration with
// server.AddResource for the ui:// resource URI.
func DefaultAppResourceHandler(appName, version, description string) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	html := DefaultAppHTML(appName, version, description)
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: req.Params.URI, MIMEType: "text/html", Text: html},
			},
		}, nil
	}
}

// CompletionHandlerFromEnums builds a CompletionHandler that serves autocomplete
// values for prompt arguments. The enumValues map is keyed by "promptName:argName".
func CompletionHandlerFromEnums(enumValues map[string][]string) func(context.Context, *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return func(_ context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
		if req.Params.Ref.Type != "ref/prompt" {
			return &mcp.CompleteResult{Completion: mcp.CompletionResultDetails{Values: []string{}}}, nil
		}
		key := req.Params.Ref.Name + ":" + req.Params.Argument.Name
		values, ok := enumValues[key]
		if !ok {
			return &mcp.CompleteResult{Completion: mcp.CompletionResultDetails{Values: []string{}}}, nil
		}
		prefix := strings.ToLower(req.Params.Argument.Value)
		filtered := []string{}
		for _, v := range values {
			if strings.HasPrefix(strings.ToLower(v), prefix) {
				filtered = append(filtered, v)
			}
		}
		return &mcp.CompleteResult{
			Completion: mcp.CompletionResultDetails{
				Values:  filtered,
				Total:   len(filtered),
				HasMore: false,
			},
		}, nil
	}
}

// ElicitField describes a field for an elicitation request.
type ElicitField struct {
	Name        string
	Description string
	Required    bool
	Type        string // "string", "number", "boolean"
	EnumValues  []string
}

// RunElicitation performs an elicitation request on the server session,
// building a JSON schema from the given fields. Returns the result and any
// error. If the user declines (action != "accept"), the caller should handle
// accordingly.
func RunElicitation(ctx context.Context, session *mcp.ServerSession, message string, fields []ElicitField) (*mcp.ElicitResult, error) {
	props := make(map[string]*jsonschema.Schema, len(fields))
	var required []string
	for _, f := range fields {
		sch := &jsonschema.Schema{Type: f.Type, Description: f.Description}
		if len(f.EnumValues) > 0 {
			for _, v := range f.EnumValues {
				sch.Enum = append(sch.Enum, v)
			}
		}
		props[f.Name] = sch
		if f.Required {
			required = append(required, f.Name)
		}
	}
	return session.Elicit(ctx, &mcp.ElicitParams{
		Message: message,
		RequestedSchema: &jsonschema.Schema{
			Type:       "object",
			Properties: props,
			Required:   required,
		},
	})
}
