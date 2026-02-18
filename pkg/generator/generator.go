package generator

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"go/token"
	"math/big"
	"path"
	"path/filepath"
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
}

// TplParams is the top-level data fed into the code template.
type TplParams struct {
	SourcePath       string
	GoPackage        string
	SchemaJSON       map[string]string // key: ServiceName_MethodName -> schema JSON
	ToolMeta         map[string]ToolMeta
	Services         map[string]map[string]MethodInfo
	ServiceBasePaths map[string]string // key: ServiceName -> default base path e.g. "/todo/v1/TodoService"
}

// FileGenerator produces a single *.pb.mcp.go file from a protobuf file.
type FileGenerator struct {
	f   *protogen.File
	gen *protogen.Plugin
	gf  *protogen.GeneratedFile
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
	if packageSuffix != "" {
		g.gf.Import(file.GoImportPath)
	}

	funcMap := template.FuncMap{
		"backtick": func() string { return "`" },
	}
	tpl, err := template.New("gen").Funcs(funcMap).Parse(codeTemplates[LangGo])
	if err != nil {
		g.gen.Error(err)
		return
	}

	params := g.buildParams()
	if err := tpl.Execute(g.gf, params); err != nil {
		g.gen.Error(err)
	}
}

// buildParams iterates over all services/methods and builds the template data.
func (g *FileGenerator) buildParams() TplParams {
	services := make(map[string]map[string]MethodInfo)
	schemaJSON := make(map[string]string)
	toolMeta := make(map[string]ToolMeta)
	serviceBasePaths := make(map[string]string)

	for _, svc := range g.f.Services {
		methods := make(map[string]MethodInfo)

		for _, meth := range svc.Methods {
			// Only unary RPCs are supported.
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

			toolMeta[key] = ToolMeta{
				Name:        toolName,
				Description: CleanComment(string(meth.Comments.Leading)),
			}

			methods[meth.GoName] = MethodInfo{
				RequestType:  g.gf.QualifiedGoIdent(meth.Input.GoIdent),
				ResponseType: g.gf.QualifiedGoIdent(meth.Output.GoIdent),
			}
		}

		svcName := string(svc.Desc.Name())
		services[svcName] = methods
		serviceBasePaths[svcName] = "/" + strings.ToLower(strings.ReplaceAll(string(svc.Desc.FullName()), ".", "/")) + "/mcp"
	}

	return TplParams{
		SourcePath:       g.f.Desc.Path(),
		GoPackage:        string(g.f.GoPackageName),
		SchemaJSON:       schemaJSON,
		ToolMeta:         toolMeta,
		Services:         services,
		ServiceBasePaths: serviceBasePaths,
	}
}

// ---------------------------------------------------------------------------
// Tool-name helpers
// ---------------------------------------------------------------------------

// MangleHeadIfTooLong truncates the head of name and prepends a short hash
// when name exceeds maxLen.  The tail (most-specific part) is preserved.
func MangleHeadIfTooLong(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	hash := sha1.Sum([]byte(name))
	prefix := base36(hash[:])[:6]
	available := maxLen - len(prefix) - 1
	if available <= 0 {
		return prefix
	}
	return prefix + "_" + name[len(name)-available:]
}

func base36(b []byte) string {
	n := new(big.Int).SetBytes(b)
	return n.Text(36)
}
