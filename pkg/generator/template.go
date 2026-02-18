package generator

import (
	"fmt"

	"github.com/machanirobotics/protoc-mcp-gen/pkg/generator/templates"
)

// Supported languages for code generation.
const (
	LangGo     = "go"
	LangPython = "python"
	LangRust   = "rust"
)

// codeTemplates maps language name → embedded template content.
var codeTemplates = map[string]string{
	LangGo:     mustReadTemplate("go.tpl"),
	LangPython: mustReadTemplate("python.tpl"),
	LangRust:   mustReadTemplate("rust.tpl"),
}

// GetTemplate returns the code template for the given language.
func GetTemplate(lang string) (string, error) {
	tpl, ok := codeTemplates[lang]
	if !ok {
		return "", fmt.Errorf("unsupported language %q (supported: go, python, rust)", lang)
	}
	return tpl, nil
}

func mustReadTemplate(name string) string {
	b, err := templates.FS.ReadFile(name)
	if err != nil {
		panic("generator: embedded template " + name + " not found: " + err.Error())
	}
	return string(b)
}
