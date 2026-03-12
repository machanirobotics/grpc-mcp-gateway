import * as jspb from 'google-protobuf'

import * as mcp_protobuf_app_pb from '../../mcp/protobuf/app_pb'; // proto import: "mcp/protobuf/app.proto"


export class MCPServiceOptions extends jspb.Message {
  getApp(): mcp_protobuf_app_pb.MCPApp | undefined;
  setApp(value?: mcp_protobuf_app_pb.MCPApp): MCPServiceOptions;
  hasApp(): boolean;
  clearApp(): MCPServiceOptions;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPServiceOptions.AsObject;
  static toObject(includeInstance: boolean, msg: MCPServiceOptions): MCPServiceOptions.AsObject;
  static serializeBinaryToWriter(message: MCPServiceOptions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPServiceOptions;
  static deserializeBinaryFromReader(message: MCPServiceOptions, reader: jspb.BinaryReader): MCPServiceOptions;
}

export namespace MCPServiceOptions {
  export type AsObject = {
    app?: mcp_protobuf_app_pb.MCPApp.AsObject;
  };
}

