package generator

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const mcpProgressFQN = "mcp.protobuf.MCPProgress"

// StreamProgressInfo describes a server-streaming RPC that uses MCPProgress for progress updates.
// When non-nil, the streamed message has a oneof with MCPProgress and a result field.
type StreamProgressInfo struct {
	StreamChunkType  string            // Go type for the streamed message
	StreamClientType string            // Go type for gRPC client stream (e.g. "TodoService_CreateTodoClient")
	StreamServerType string            // Go type for gRPC server stream (e.g. "TodoService_CreateTodoServer")
	ResultType       string            // Go type for the final result (resolved)
	ResultMessage    *protogen.Message // Result message (for Python/Rust type resolution)
	ProgressField    string            // oneof field name for progress (e.g. "Progress")
	ResultField      string            // oneof field name for result (e.g. "Result")
	ServiceName      string            // e.g. "TodoService"
	MethodName       string            // e.g. "CreateTodo"
}

// DetectProgressStream returns StreamProgressInfo if the method is server-streaming
// and the streamed message follows the progress convention: a oneof with
// mcp.protobuf.MCPProgress and the result type.
func DetectProgressStream(meth *protogen.Method, resolveType func(protogen.GoIdent) string) *StreamProgressInfo {
	if !meth.Desc.IsStreamingServer() || meth.Desc.IsStreamingClient() {
		return nil
	}
	msg := meth.Output
	if msg == nil {
		return nil
	}
	// Find the oneof that contains both MCPProgress and a result (same oneof).
	var progressField, resultField *protogen.Field
	var resultIdent protogen.GoIdent
	for _, oo := range msg.Oneofs {
		if oo.Desc.IsSynthetic() {
			continue
		}
		var prog, res *protogen.Field
		var resIdent protogen.GoIdent
		for _, f := range oo.Fields {
			if f.Desc.Kind() != protoreflect.MessageKind {
				continue
			}
			fqn := string(f.Message.Desc.FullName())
			if fqn == mcpProgressFQN {
				prog = f
			} else {
				res = f
				resIdent = f.Message.GoIdent
			}
		}
		if prog != nil && res != nil {
			progressField = prog
			resultField = res
			resultIdent = resIdent
			break
		}
	}
	if progressField == nil || resultField == nil {
		return nil
	}
	svcName := string(meth.Parent.Desc.Name())
	return &StreamProgressInfo{
		StreamChunkType:  resolveType(msg.GoIdent),
		StreamClientType: svcName + "_" + meth.GoName + "Client",
		StreamServerType: svcName + "_" + meth.GoName + "Server",
		ResultType:       resolveType(resultIdent),
		ResultMessage:    resultField.Message,
		ProgressField:    progressField.GoName,
		ResultField:      resultField.GoName,
		ServiceName:      svcName,
		MethodName:       meth.GoName,
	}
}
