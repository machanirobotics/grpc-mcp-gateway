import * as jspb from 'google-protobuf'



export class MCPProgress extends jspb.Message {
  getProgress(): number;
  setProgress(value: number): MCPProgress;

  getTotal(): number;
  setTotal(value: number): MCPProgress;
  hasTotal(): boolean;
  clearTotal(): MCPProgress;

  getMessage(): string;
  setMessage(value: string): MCPProgress;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MCPProgress.AsObject;
  static toObject(includeInstance: boolean, msg: MCPProgress): MCPProgress.AsObject;
  static serializeBinaryToWriter(message: MCPProgress, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MCPProgress;
  static deserializeBinaryFromReader(message: MCPProgress, reader: jspb.BinaryReader): MCPProgress;
}

export namespace MCPProgress {
  export type AsObject = {
    progress: number;
    total?: number;
    message: string;
  };

  export enum TotalCase {
    _TOTAL_NOT_SET = 0,
    TOTAL = 2,
  }
}

