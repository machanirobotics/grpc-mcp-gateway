package main

import (
	"flag"
	"fmt"

	"github.com/machanirobotics/protoc-mcp-gen/pkg/generator"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var flags flag.FlagSet
	lang := flags.String(
		"lang",
		"go",
		"Target language for generated MCP code (go, python, rust).",
	)
	packageSuffix := flags.String(
		"package_suffix",
		"",
		"(Go only) Sub-package suffix for generated files (empty = same package as .pb.go files).",
	)

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			switch *lang {
			case "go":
				generator.NewFileGenerator(f, gen).Generate(*packageSuffix)
			case "python":
				generator.NewPythonFileGenerator(f, gen).Generate()
			case "rust":
				generator.NewRustFileGenerator(f, gen).Generate()
			default:
				return fmt.Errorf("unsupported language: %q (supported: go, python, rust)", *lang)
			}
		}
		return nil
	})
}
