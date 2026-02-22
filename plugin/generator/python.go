package generator

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// PyMethodInfo carries Python-specific type identifiers for a single RPC method.
type PyMethodInfo struct {
	PyMethodName   string // snake_case method name
	PyRequestType  string // e.g. todo_pb2.CreateTodoRequest
	PyResponseType string // e.g. todo_pb2.Todo
	ToolName       string // MCP tool name
	MethodOpts     *MCPMethodOpts
}

// PyTplParams is the top-level data fed into the Python code template.
type PyTplParams struct {
	Version          string
	SourcePath       string
	PBImports        string                           // import lines for *_pb2 modules
	SchemaJSON       map[string]string                // key: ServiceName_MethodName -> schema JSON
	ToolMeta         map[string]ToolMeta              // key: ServiceName_MethodName
	Services         map[string]map[string]PyMethodInfo
	ServiceBasePaths map[string]string                // key: ServiceName -> default base path
	ServiceOpts      map[string]*MCPServiceOpts       // key: ServiceName
}

// PythonFileGenerator produces a single *_pb2_mcp.py file from a protobuf file.
type PythonFileGenerator struct {
	f   *protogen.File
	gen *protogen.Plugin
}

// NewPythonFileGenerator creates a PythonFileGenerator for the given protobuf file.
func NewPythonFileGenerator(f *protogen.File, gen *protogen.Plugin) *PythonFileGenerator {
	gen.SupportedFeatures |= uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	return &PythonFileGenerator{f: f, gen: gen}
}

// Generate produces the *_pb2_mcp.py output file.
func (g *PythonFileGenerator) Generate() {
	file := g.f
	if len(file.Services) == 0 {
		return
	}

	// Output file path: same directory as the proto, with _pb2_mcp.py suffix.
	// e.g. store/apps/todo/v1/todo_service_pb2_mcp.py
	outName := file.GeneratedFilenamePrefix + "_pb2_mcp.py"
	gf := g.gen.NewGeneratedFile(outName, "")

	funcMap := template.FuncMap{
		"snakeCase":    toSnakeCase,
		"pyString":     pyStringLiteral,
		"escapeQuotes": func(s string) string { return strings.ReplaceAll(s, `"`, `\"`) },
	}

	tpl, err := template.New("pygen").Funcs(funcMap).Parse(codeTemplates[LangPython])
	if err != nil {
		g.gen.Error(err)
		return
	}

	params := g.buildPyParams()
	if err := tpl.Execute(gf, params); err != nil {
		g.gen.Error(err)
	}
}

// buildPyParams iterates over all services/methods and builds the Python template data.
func (g *PythonFileGenerator) buildPyParams() PyTplParams {
	services := make(map[string]map[string]PyMethodInfo)
	schemaJSON := make(map[string]string)
	toolMeta := make(map[string]ToolMeta)
	serviceBasePaths := make(map[string]string)
	serviceOpts := make(map[string]*MCPServiceOpts)

	// Collect all imported proto files needed for request/response types.
	pbImports := make(map[string]bool)

	for _, svc := range g.f.Services {
		methods := make(map[string]PyMethodInfo)

		for _, meth := range svc.Methods {
			if meth.Desc.IsStreamingClient() || meth.Desc.IsStreamingServer() {
				continue
			}

			key := string(svc.Desc.Name()) + "_" + meth.GoName
			toolName := BuildToolName(string(meth.Desc.FullName()))

			// Apply method-level option overrides.
			methOpts := ExtractMethodOptions(meth)
			if methOpts != nil {
				if methOpts.ToolName != "" {
					toolName = methOpts.ToolName
				}
				// Resolve prompt schema → populate Arguments from proto message fields.
				if methOpts.Prompt != nil && methOpts.Prompt.Schema != "" {
					for _, sf := range ResolveSchemaFields(g.gen, methOpts.Prompt.Schema) {
						methOpts.Prompt.Arguments = append(methOpts.Prompt.Arguments, MCPPromptArgOpts(sf))
					}
				}
				// Resolve elicitation schema → populate Fields from proto message fields.
				if methOpts.Elicitation != nil && methOpts.Elicitation.Schema != "" {
					for _, sf := range ResolveSchemaFields(g.gen, methOpts.Elicitation.Schema) {
						methOpts.Elicitation.Fields = append(methOpts.Elicitation.Fields, MCPElicitFieldOpts(sf))
					}
				}
			}

			// Standard schema
			stdSchema := messageSchema(meth.Input.Desc, false)
			stdBytes, err := json.Marshal(stdSchema)
			if err != nil {
				panic(fmt.Sprintf("marshal standard schema: %v", err))
			}
			schemaJSON[key] = string(stdBytes)

			toolDesc := CleanComment(string(meth.Comments.Leading))
			if methOpts != nil && methOpts.ToolDescription != "" {
				toolDesc = methOpts.ToolDescription
			}

			toolMeta[key] = ToolMeta{
				Name:        toolName,
				Description: toolDesc,
			}

			// Build Python import paths and type references.
			reqModule := protoPyModule(meth.Input)
			respModule := protoPyModule(meth.Output)
			pbImports[reqModule] = true
			pbImports[respModule] = true

			methods[meth.GoName] = PyMethodInfo{
				PyMethodName:   toSnakeCase(meth.GoName),
				PyRequestType:  protoPyType(meth.Input),
				PyResponseType: protoPyType(meth.Output),
				ToolName:       toolName,
				MethodOpts:     methOpts,
			}
		}

		svcName := string(svc.Desc.Name())
		services[svcName] = methods
		serviceBasePaths[svcName] = "/" + strings.ToLower(strings.ReplaceAll(string(svc.Desc.FullName()), ".", "/")) + "/mcp"
		svcOpt := ExtractServiceOptions(svc)
		apiResources := ExtractGoogleAPIResources(svc)
		if len(apiResources) > 0 {
			if svcOpt == nil {
				svcOpt = &MCPServiceOpts{}
			}
			svcOpt.Resources = apiResources
		}
		serviceOpts[svcName] = svcOpt
	}

	// Build import lines.
	var importLines []string
	for mod := range pbImports {
		importLines = append(importLines, fmt.Sprintf("import %s", mod))
	}

	// Add PyDescription to ToolMeta (Python-safe string literal).
	for key, meta := range toolMeta {
		meta.Description = strings.TrimSpace(meta.Description)
		toolMeta[key] = meta
	}

	return PyTplParams{
		Version:          PluginVersion,
		SourcePath:       g.f.Desc.Path(),
		PBImports:        strings.Join(importLines, "\n"),
		SchemaJSON:       schemaJSON,
		ToolMeta:         toolMeta,
		Services:         services,
		ServiceBasePaths: serviceBasePaths,
		ServiceOpts:      serviceOpts,
	}
}

