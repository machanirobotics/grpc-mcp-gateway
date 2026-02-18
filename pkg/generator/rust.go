package generator

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// RsMethodInfo carries Rust-specific identifiers for a single RPC method.
type RsMethodInfo struct {
	RsMethodName string // snake_case method name, e.g. create_todo
	ConstName    string // SCREAMING_SNAKE constant prefix, e.g. TODO_SERVICE_CREATE_TODO
	ToolName     string // MCP tool name, e.g. todo_v1_TodoService_CreateTodo
	Description  string // method description
}

// RsTplParams is the top-level data fed into the Rust code template.
type RsTplParams struct {
	SourcePath       string
	SchemaJSON       map[string]string                 // key: ServiceName_MethodName -> schema JSON
	ToolMeta         map[string]ToolMeta               // key: ServiceName_MethodName
	Services         map[string]map[string]RsMethodInfo // key: ServiceName -> MethodName -> info
	ServiceBasePaths map[string]string                  // key: ServiceName -> default base path
}

// RustFileGenerator produces a single *_mcp.rs file from a protobuf file.
type RustFileGenerator struct {
	f   *protogen.File
	gen *protogen.Plugin
}

// NewRustFileGenerator creates a RustFileGenerator for the given protobuf file.
func NewRustFileGenerator(f *protogen.File, gen *protogen.Plugin) *RustFileGenerator {
	gen.SupportedFeatures |= uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	return &RustFileGenerator{f: f, gen: gen}
}

// Generate produces the *_mcp.rs output file.
func (g *RustFileGenerator) Generate() {
	file := g.f
	if len(file.Services) == 0 {
		return
	}

	// Use the proto file stem (e.g. "audio_service") to produce one
	// MCP file per proto source file (e.g. "todo/v1/audio_service.mcp.rs").
	// This avoids collisions when multiple services share the same proto
	// package -- the previous per-package naming caused only the last
	// service's code to survive.
	dir := filepath.Dir(file.Desc.Path())
	stem := strings.TrimSuffix(filepath.Base(file.Desc.Path()), ".proto")
	outName := filepath.Join(dir, stem+".mcp.rs")
	gf := g.gen.NewGeneratedFile(outName, "")

	funcMap := template.FuncMap{
		"snakeCase":          toSnakeCase,
		"screamingSnakeCase": toScreamingSnakeCase,
		"rsEscape":           rsStringEscape,
	}

	tpl, err := template.New("rsgen").Funcs(funcMap).Parse(codeTemplates[LangRust])
	if err != nil {
		g.gen.Error(err)
		return
	}

	params := g.buildRsParams()
	if err := tpl.Execute(gf, params); err != nil {
		g.gen.Error(err)
	}
}

// buildRsParams iterates over all services/methods and builds the Rust template data.
func (g *RustFileGenerator) buildRsParams() RsTplParams {
	services := make(map[string]map[string]RsMethodInfo)
	schemaJSON := make(map[string]string)
	toolMeta := make(map[string]ToolMeta)
	serviceBasePaths := make(map[string]string)

	for _, svc := range g.f.Services {
		methods := make(map[string]RsMethodInfo)

		for _, meth := range svc.Methods {
			if meth.Desc.IsStreamingClient() || meth.Desc.IsStreamingServer() {
				continue
			}

			key := string(svc.Desc.Name()) + "_" + meth.GoName
			toolName := MangleHeadIfTooLong(
				strings.ReplaceAll(string(meth.Desc.FullName()), ".", "_"), 128,
			)

			// Standard schema
			stdSchema := messageSchema(meth.Input.Desc, false)
			stdBytes, err := json.Marshal(stdSchema)
			if err != nil {
				panic(fmt.Sprintf("marshal standard schema: %v", err))
			}
			schemaJSON[key] = string(stdBytes)

			desc := strings.TrimSpace(CleanComment(string(meth.Comments.Leading)))
			toolMeta[key] = ToolMeta{
				Name:        toolName,
				Description: desc,
			}

			methods[meth.GoName] = RsMethodInfo{
				RsMethodName: toSnakeCase(meth.GoName),
				ConstName:    toScreamingSnakeCase(key),
				ToolName:     toolName,
				Description:  desc,
			}
		}

		svcName := string(svc.Desc.Name())
		services[svcName] = methods
		serviceBasePaths[svcName] = "/" + strings.ToLower(strings.ReplaceAll(string(svc.Desc.FullName()), ".", "/")) + "/mcp"
	}

	return RsTplParams{
		SourcePath:       g.f.Desc.Path(),
		SchemaJSON:       schemaJSON,
		ToolMeta:         toolMeta,
		Services:         services,
		ServiceBasePaths: serviceBasePaths,
	}
}

// toScreamingSnakeCase converts a CamelCase or snake_case string to SCREAMING_SNAKE_CASE.
func toScreamingSnakeCase(s string) string {
	return strings.ToUpper(toSnakeCase(s))
}

// rsStringEscape escapes backslashes and double quotes for use inside a Rust "..." string literal.
func rsStringEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
