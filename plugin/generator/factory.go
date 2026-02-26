package generator

import (
	"fmt"
	"sort"

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
	// CppEmitShared, when Lang is Cpp, controls whether to emit shared files
	// (rust/*, Makefile, main.cc). Nil defaults to true.
	CppEmitShared *bool
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
		emitShared := true
		if opts.CppEmitShared != nil {
			emitShared = *opts.CppEmitShared
		}
		NewCppFileGenerator(f, gen).Generate(emitShared)
	default:
		return fmt.Errorf("unsupported language: %q (supported: %s)", opts.Lang, supportedList())
	}
	return nil
}

// GenerateAll runs code generation for every target language on a single
// protobuf file. Cpp is excluded; use GenerateCppBatch for C++.
func GenerateAll(f *protogen.File, gen *protogen.Plugin, packageSuffix string) error {
	for _, lang := range SupportedLanguages() {
		if lang == Cpp {
			continue
		}
		if err := GenerateFile(f, gen, GenerateOptions{
			Lang:          lang,
			PackageSuffix: packageSuffix,
		}); err != nil {
			return err
		}
	}
	return nil
}

// GenerateCppBatch runs C++ generation for all files with services, emitting
// shared files only for the first file to avoid duplicates.
func GenerateCppBatch(gen *protogen.Plugin) error {
	var files []*protogen.File
	for _, f := range gen.Files {
		if !f.Generate || len(f.Services) == 0 {
			continue
		}
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Desc.Path() < files[j].Desc.Path()
	})
	for i, f := range files {
		emitShared := i == 0
		if err := GenerateFile(f, gen, GenerateOptions{
			Lang:          Cpp,
			CppEmitShared: &emitShared,
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
