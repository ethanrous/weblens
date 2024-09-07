export type FileEventT = {
    Action: string;
    // At: number;
    FileId: string;
    Path: string;
    FromFileId: string;
    FromPath: string;
    // Size: number
    // SnapshotId: string;
    Timestamp: string;

    millis: number;

    // Non-api fields
    fileName: string
    // oldPath: string
}