import * as jspb from 'google-protobuf'



export class MCPFieldOptions extends jspb.Message {
  getDescription(): string;
  setDescription(value: string): MCPFieldOptions;

  getExamplesList(): Array<string>;
  setExamplesList(value: Array<string>): MCPFieldOptions;
  clearExamplesList(): MCPFieldOptions;
  addExamples(value: string, index?: number): MCPFieldOptions;

  getDeprecated(): boolean;
  setDeprecated(value: boolean): MCPFieldOptions;

  getFormat(): string;
  setFormat(value: string): MCPFieldOptions;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPFieldOptions.AsObject;
  static toObject(includeInstance: boolean, msg: MCPFieldOptions): MCPFieldOptions.AsObject;
  static serializeBinaryToWriter(message: MCPFieldOptions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPFieldOptions;
  static deserializeBinaryFromReader(message: MCPFieldOptions, reader: jspb.BinaryReader): MCPFieldOptions;
}

export namespace MCPFieldOptions {
  export type AsObject = {
    description: string;
    examplesList: Array<string>;
    deprecated: boolean;
    format: string;
  };
}

