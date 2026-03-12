import * as jspb from 'google-protobuf'

import * as google_api_resource_pb from '../../google/api/resource_pb'; // proto import: "google/api/resource.proto"


export class MCPApp extends jspb.Message {
  getName(): string;
  setName(value: string): MCPApp;

  getVersion(): string;
  setVersion(value: string): MCPApp;

  getDescription(): string;
  setDescription(value: string): MCPApp;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPApp.AsObject;
  static toObject(includeInstance: boolean, msg: MCPApp): MCPApp.AsObject;
  static serializeBinaryToWriter(message: MCPApp, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPApp;
  static deserializeBinaryFromReader(message: MCPApp, reader: jspb.BinaryReader): MCPApp;
}

export namespace MCPApp {
  export type AsObject = {
    name: string;
    version: string;
    description: string;
  };
}

