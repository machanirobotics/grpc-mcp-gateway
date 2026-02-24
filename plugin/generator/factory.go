package generator

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

// PluginVersion is set by the protoc-gen-mcp binary before generation.
var PluginVersion = "dev"

// Language represents a supported code generation target.
type Language string

const (
	Go     Language = "go"
	Python Language = "python"
	Rust   Language = "rust"
	Cpp    Language = "cpp"
)

// SupportedLanguages returns all languages the factory can generate for.
func SupportedLanguages() []Language {
	return []Language{Go, Python, Rust, Cpp}
}

// GenerateOptions holds configuration for a single file generation run.
type GenerateOptions struct {
	// Lang selects the target language.
	Lang Language
	// PackageSuffix is Go-specific: sub-package suffix for generated files.
	PackageSuffix string
}

// GenerateFile dispatches code generation for a single protobuf file to the
// appropriate language-specific generator.
func GenerateFile(f *protogen.File, gen *protogen.Plugin, opts GenerateOptions) error {
	switch opts.Lang {
	case Go:
		NewFileGenerator(f, gen).Generate(opts.PackageSuffix)
	case Python:
		NewPythonFileGenerator(f, gen).Generate()
	case Rust:
		NewRustFileGenerator(f, gen).Generate()
	case Cpp:
		NewCppFileGenerator(f, gen).Generate()
	default:
		return fmt.Errorf("unsupported language: %q (supported: %s)", opts.Lang, supportedList())
	}
	return nil
}

// GenerateAll runs code generation for every target language on a single
// protobuf file. Useful for mono-repo setups that publish bindings for
// all languages at once.
func GenerateAll(f *protogen.File, gen *protogen.Plugin, packageSuffix string) error {
	for _, lang := range SupportedLanguages() {
		if err := GenerateFile(f, gen, GenerateOptions{
			Lang:          lang,
			PackageSuffix: packageSuffix,
		}); err != nil {
			return err
		}
	}
	return nil
}

func supportedList() string {
	langs := SupportedLanguages()
	s := make([]string, len(langs))
	for i, l := range langs {
		s[i] = string(l)
	}
	return fmt.Sprintf("%v", s)
}
