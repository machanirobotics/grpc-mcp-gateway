package generator

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// toSnakeCase converts a CamelCase string to snake_case.
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32) // lowercase
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// toScreamingSnakeCase converts a CamelCase or snake_case string to SCREAMING_SNAKE_CASE.
func toScreamingSnakeCase(s string) string {
	return strings.ToUpper(toSnakeCase(s))
}

// pyStringLiteral wraps a string as a Python string literal, using triple-quotes for multiline.
func pyStringLiteral(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	if strings.Contains(s, "\n") {
		return `"""` + s + `"""`
	}
	return `"` + s + `"`
}

// rsStringEscape escapes backslashes and double quotes for use inside a Rust "..." string literal.
func rsStringEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// protoPyModule returns the Python module path for a protobuf message.
// e.g. store.apps.todo.v1.Todo -> store.apps.todo.v1.todo_pb2
func protoPyModule(msg *protogen.Message) string {
	path := msg.Location.SourceFile
	path = strings.TrimSuffix(path, ".proto")
	path = strings.ReplaceAll(path, "/", ".")
	return path + "_pb2"
}

// protoPyType returns the fully-qualified Python type for a protobuf message.
// e.g. store.apps.todo.v1.Todo -> store.apps.todo.v1.todo_pb2.Todo
func protoPyType(msg *protogen.Message) string {
	module := protoPyModule(msg)
	return module + "." + string(msg.Desc.Name())
}
