import * as jspb from 'google-protobuf'

import * as mcp_protobuf_mime_type_pb from '../../mcp/protobuf/mime_type_pb'; // proto import: "mcp/protobuf/mime_type.proto"


export class MCPResource extends jspb.Message {
  getUri(): string;
  setUri(value: string): MCPResource;

  getPattern(): string;
  setPattern(value: string): MCPResource;

  getName(): string;
  setName(value: string): MCPResource;

  getDescription(): string;
  setDescription(value: string): MCPResource;

  getMimeType(): mcp_protobuf_mime_type_pb.MCPMimeType;
  setMimeType(value: mcp_protobuf_mime_type_pb.MCPMimeType): MCPResource;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPResource.AsObject;
  static toObject(includeInstance: boolean, msg: MCPResource): MCPResource.AsObject;
  static serializeBinaryToWriter(message: MCPResource, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPResource;
  static deserializeBinaryFromReader(message: MCPResource, reader: jspb.BinaryReader): MCPResource;
}

export namespace MCPResource {
  export type AsObject = {
    uri: string;
    pattern: string;
    name: string;
    description: string;
    mimeType: mcp_protobuf_mime_type_pb.MCPMimeType;
  };
}

