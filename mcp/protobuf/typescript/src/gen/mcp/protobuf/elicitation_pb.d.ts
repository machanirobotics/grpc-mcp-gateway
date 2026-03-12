import * as jspb from 'google-protobuf'



export class MCPElicitation extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): MCPElicitation;

  getSchema(): string;
  setSchema(value: string): MCPElicitation;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPElicitation.AsObject;
  static toObject(includeInstance: boolean, msg: MCPElicitation): MCPElicitation.AsObject;
  static serializeBinaryToWriter(message: MCPElicitation, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPElicitation;
  static deserializeBinaryFromReader(message: MCPElicitation, reader: jspb.BinaryReader): MCPElicitation;
}

export namespace MCPElicitation {
  export type AsObject = {
    message: string;
    schema: string;
  };
}

