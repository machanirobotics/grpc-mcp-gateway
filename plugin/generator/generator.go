package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const generatedFilenameExtension = ".pb.mcp.go"

// ToolMeta holds the MCP tool name and description for a single RPC method.
type ToolMeta struct {
	Name        string
	Description string
}

// MethodInfo carries the Go type identifiers needed by the code template.
type MethodInfo struct {
	RequestType  string
	ResponseType string
	MethodOpts   *MCPMethodOpts
}

// TplParams is the top-level data fed into the code template.
type TplParams struct {
	Version          string
	SourcePath       string
	GoPackage        string
	ExtraImports     []string // e.g. `emptypb "google.golang.org/.../emptypb"`
	SchemaJSON       map[string]string // key: ServiceName_MethodName -> schema JSON
	ToolMeta         map[string]ToolMeta
	Services         map[string]map[string]MethodInfo
	ServiceBasePaths map[string]string // key: ServiceName -> default base path e.g. "/todo/v1/TodoService"
	ServiceOpts      map[string]*MCPServiceOpts // key: ServiceName
}

// FileGenerator produces a single *.pb.mcp.go file from a protobuf file.
type FileGenerator struct {
	f             *protogen.File
	gen           *protogen.Plugin
	gf            *protogen.GeneratedFile
	genImportPath protogen.GoImportPath
}

// NewFileGenerator creates a FileGenerator for the given protobuf file.
func NewFileGenerator(f *protogen.File, gen *protogen.Plugin) *FileGenerator {
	gen.SupportedFeatures |= uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	return &FileGenerator{f: f, gen: gen}
}

// Generate produces the *.pb.mcp.go output file.  It is a no-op when the
// protobuf file contains no service definitions.
func (g *FileGenerator) Generate(packageSuffix string) {
	file := g.f
	if len(file.Services) == 0 {
		return
	}

	goImportPath := file.GoImportPath
	if packageSuffix != "" {
		if !token.IsIdentifier(packageSuffix) {
			g.gen.Error(fmt.Errorf("package_suffix %q is not a valid Go identifier", packageSuffix))
			return
		}
		file.GoPackageName += protogen.GoPackageName(packageSuffix)
		prefix := filepath.ToSlash(file.GeneratedFilenamePrefix)
		file.GeneratedFilenamePrefix = path.Join(
			path.Dir(prefix),
			string(file.GoPackageName),
			path.Base(prefix),
		)
		goImportPath = protogen.GoImportPath(path.Join(
			string(file.GoImportPath),
			string(file.GoPackageName),
		))
	}

	g.gf = g.gen.NewGeneratedFile(
		file.GeneratedFilenamePrefix+generatedFilenameExtension,
		goImportPath,
	)
	g.genImportPath = goImportPath

	funcMap := template.FuncMap{
		"backtick":     func() string { return "`" },
		"escapeQuotes": func(s string) string { return strings.ReplaceAll(s, `"`, `\"`) },
	}
	tpl, err := template.New("gen").Funcs(funcMap).Parse(codeTemplates[LangGo])
	if err != nil {
		g.gen.Error(err)
		return
	}

	params := g.buildParams()
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, params); err != nil {
		g.gen.Error(err)
		return
	}

	// Validate generated Go source is syntactically correct.
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "", buf.Bytes(), parser.AllErrors); err != nil {
		g.gen.Error(fmt.Errorf("%s: unparsable Go source: %v", file.GeneratedFilenamePrefix+generatedFilenameExtension, err))
		return
	}

	if _, err := g.gf.Write(buf.Bytes()); err != nil {
		g.gen.Error(err)
	}
}

// buildParams iterates over all services/methods and builds the template data.
func (g *FileGenerator) buildParams() TplParams {
	services := make(map[string]map[string]MethodInfo)
	schemaJSON := make(map[string]string)
	toolMeta := make(map[string]ToolMeta)
	serviceBasePaths := make(map[string]string)
	serviceOpts := make(map[string]*MCPServiceOpts)
	extraImportMap := make(map[protogen.GoImportPath]string)

	resolveType := func(ident protogen.GoIdent) string {
		if ident.GoImportPath == g.genImportPath {
			return ident.GoName
		}
		alias := path.Base(string(ident.GoImportPath))
		extraImportMap[ident.GoImportPath] = alias
		return alias + "." + ident.GoName
	}

	for _, svc := range g.f.Services {
		methods := make(map[string]MethodInfo)

		for _, meth := range svc.Methods {
			// Only unary RPCs are supported.
			if meth.Desc.IsStreamingClient() || meth.Desc.IsStreamingServer() {
				continue
			}

			key := string(svc.Desc.Name()) + "_" + meth.GoName
			toolName := BuildToolName(string(meth.Desc.FullName()))
			toolDesc := CleanComment(string(meth.Comments.Leading))

			// Apply method-level option overrides.
			methOpts := ExtractMethodOptions(meth)
			if methOpts != nil {
				if methOpts.ToolName != "" {
					toolName = methOpts.ToolName
				}
				if methOpts.ToolDescription != "" {
					toolDesc = methOpts.ToolDescription
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

			toolMeta[key] = ToolMeta{
				Name:        toolName,
				Description: toolDesc,
			}

			methods[meth.GoName] = MethodInfo{
				RequestType:  resolveType(meth.Input.GoIdent),
				ResponseType: resolveType(meth.Output.GoIdent),
				MethodOpts:   methOpts,
			}
		}

		svcName := string(svc.Desc.Name())
		services[svcName] = methods
		serviceBasePaths[svcName] = "/" + strings.ToLower(strings.ReplaceAll(string(svc.Desc.FullName()), ".", "/")) + "/mcp"

		// Extract explicit MCP service options + auto-detect google.api.resource.
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

	var extraImports []string
	for importPath, alias := range extraImportMap {
		extraImports = append(extraImports, alias+" "+`"`+string(importPath)+`"`)
	}
	sort.Strings(extraImports)

	return TplParams{
		Version:          PluginVersion,
		SourcePath:       g.f.Desc.Path(),
		GoPackage:        string(g.f.GoPackageName),
		ExtraImports:     extraImports,
		SchemaJSON:       schemaJSON,
		ToolMeta:         toolMeta,
		Services:         services,
		ServiceBasePaths: serviceBasePaths,
		ServiceOpts:      serviceOpts,
	}
}
