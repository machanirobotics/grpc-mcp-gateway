import * as jspb from 'google-protobuf'



export class MCPPrompt extends jspb.Message {
  getName(): string;
  setName(value: string): MCPPrompt;

  getDescription(): string;
  setDescription(value: string): MCPPrompt;

  getSchema(): string;
  setSchema(value: string): MCPPrompt;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPPrompt.AsObject;
  static toObject(includeInstance: boolean, msg: MCPPrompt): MCPPrompt.AsObject;
  static serializeBinaryToWriter(message: MCPPrompt, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPPrompt;
  static deserializeBinaryFromReader(message: MCPPrompt, reader: jspb.BinaryReader): MCPPrompt;
}

export namespace MCPPrompt {
  export type AsObject = {
    name: string;
    description: string;
    schema: string;
  };
}

export class MCPToolOptions extends jspb.Message {
  getName(): string;
  setName(value: string): MCPToolOptions;

  getDescription(): string;
  setDescription(value: string): MCPToolOptions;

  getProgress(): boolean;
  setProgress(value: boolean): MCPToolOptions;
  hasProgress(): boolean;
  clearProgress(): MCPToolOptions;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPToolOptions.AsObject;
  static toObject(includeInstance: boolean, msg: MCPToolOptions): MCPToolOptions.AsObject;
  static serializeBinaryToWriter(message: MCPToolOptions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPToolOptions;
  static deserializeBinaryFromReader(message: MCPToolOptions, reader: jspb.BinaryReader): MCPToolOptions;
}

export namespace MCPToolOptions {
  export type AsObject = {
    name: string;
    description: string;
    progress?: boolean;
  };

  export enum ProgressCase {
    _PROGRESS_NOT_SET = 0,
    PROGRESS = 3,
  }
}

