package generator

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ResolveSchemaFields looks up a proto message by its fully-qualified name
// (e.g. "todo.v1.CreateTodoConfirmation") across all files in the plugin,
// then extracts each field's name, description (from leading comment),
// required-ness (from google.api.field_behavior), type, and enum values.
func ResolveSchemaFields(gen *protogen.Plugin, schemaFQN string) []SchemaField {
	if schemaFQN == "" {
		return nil
	}

	// Find the message across all files.
	var msg *protogen.Message
	for _, f := range gen.Files {
		msg = findMessage(f.Messages, schemaFQN)
		if msg != nil {
			break
		}
	}
	if msg == nil {
		return nil
	}

	var fields []SchemaField
	for _, field := range msg.Fields {
		desc := getFieldDescription(field.Desc, CleanComment(string(field.Comments.Leading)))
		sf := SchemaField{
			Name:        string(field.Desc.Name()),
			Description: desc,
			Required:    isFieldRequired(field.Desc),
			Type:        protoKindToJSONType(field.Desc.Kind()),
		}
		// If the field is an enum, extract its values (skip UNSPECIFIED).
		if field.Desc.Kind() == protoreflect.EnumKind && field.Enum != nil {
			for _, v := range field.Enum.Values {
				name := string(v.Desc.Name())
				if strings.HasSuffix(name, "_UNSPECIFIED") {
					continue
				}
				friendly := enumValueFriendlyName(name, string(field.Enum.Desc.Name()))
				sf.EnumValues = append(sf.EnumValues, friendly)
			}
			sf.Type = "string" // enums are presented as string choices
		}
		fields = append(fields, sf)
	}
	return fields
}

// SchemaField is a resolved field from a schema proto message.
type SchemaField struct {
	Name        string
	Description string
	Required    bool
	Type        string
	EnumValues  []string
}

// findMessage recursively searches for a message by fully-qualified name.
func findMessage(msgs []*protogen.Message, fqn string) *protogen.Message {
	for _, m := range msgs {
		if string(m.Desc.FullName()) == fqn {
			return m
		}
		if found := findMessage(m.Messages, fqn); found != nil {
			return found
		}
	}
	return nil
}

// protoKindToJSONType maps protobuf field kinds to JSON Schema types.
func protoKindToJSONType(k protoreflect.Kind) string {
	switch k {
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		return "integer"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "number"
	default:
		return "string"
	}
}

// enumValueFriendlyName strips the enum type prefix and lowercases the result.
// E.g. "CONFIRM_ACTION_YES" with enum name "ConfirmAction" → "yes".
func enumValueFriendlyName(valueName, enumName string) string {
	// Convert CamelCase enum name to UPPER_SNAKE prefix.
	// E.g. "ConfirmAction" → "CONFIRM_ACTION_"
	prefix := camelToUpperSnake(enumName) + "_"
	if strings.HasPrefix(valueName, prefix) {
		return strings.ToLower(valueName[len(prefix):])
	}
	return strings.ToLower(valueName)
}

// camelToUpperSnake converts CamelCase to UPPER_SNAKE_CASE.
func camelToUpperSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		if r >= 'a' && r <= 'z' {
			result.WriteByte(byte(r - 32))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
