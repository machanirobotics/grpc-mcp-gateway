import * as jspb from 'google-protobuf'



export class MCPEnumOptions extends jspb.Message {
  getDescription(): string;
  setDescription(value: string): MCPEnumOptions;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPEnumOptions.AsObject;
  static toObject(includeInstance: boolean, msg: MCPEnumOptions): MCPEnumOptions.AsObject;
  static serializeBinaryToWriter(message: MCPEnumOptions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPEnumOptions;
  static deserializeBinaryFromReader(message: MCPEnumOptions, reader: jspb.BinaryReader): MCPEnumOptions;
}

export namespace MCPEnumOptions {
  export type AsObject = {
    description: string;
  };
}

export class MCPEnumValueOptions extends jspb.Message {
  getDescription(): string;
  setDescription(value: string): MCPEnumValueOptions;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPEnumValueOptions.AsObject;
  static toObject(includeInstance: boolean, msg: MCPEnumValueOptions): MCPEnumValueOptions.AsObject;
  static serializeBinaryToWriter(message: MCPEnumValueOptions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPEnumValueOptions;
  static deserializeBinaryFromReader(message: MCPEnumValueOptions, reader: jspb.BinaryReader): MCPEnumValueOptions;
}

export namespace MCPEnumValueOptions {
  export type AsObject = {
    description: string;
  };
}

