export type FileEventT = {
    Action: string;
    At: number;
    FileId: string;
    Path: string;
    SecondaryFileId: string;
    Size: number
    SnapshotId: string;
    Timestamp: string;

    // Non-api fields
    fileName: string
    millis: number
    oldPath: string
}