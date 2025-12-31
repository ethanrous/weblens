import type { TaskType } from './task'

export interface WsMessage {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    content: any
    eventTag: WsEvent
    subscribeKey: string
    sentTime: number
    taskType?: TaskType
    error?: string
}

export enum WsEvent {
    BackupCompleteEvent = 'backupComplete',
    BackupFailedEvent = 'backupFailed',
    BackupProgressEvent = 'backupProgress',
    CopyFileCompleteEvent = 'copyFileComplete',
    CopyFileFailedEvent = 'copyFileFailed',
    CopyFileStartedEvent = 'copyFileStarted',
    ErrorEvent = 'error',
    FileCreatedEvent = 'fileCreated',
    FileDeletedEvent = 'fileDeleted',
    FileMovedEvent = 'fileMoved',
    FileScanCompleteEvent = 'fileScanComplete',
    FileScanStartedEvent = 'fileScanStarted',
    FileScanFailedEvent = 'fileScanFailedEvent',
    FileScanCancelledEvent = 'fileScanCancelledEvent',
    FileUpdatedEvent = 'fileUpdated',
    FilesDeletedEvent = 'filesDeleted',
    FilesMovedEvent = 'filesMoved',
    FilesUpdatedEvent = 'filesUpdated',
    FolderScanCompleteEvent = 'folderScanComplete',
    PoolCancelledEvent = 'poolCancelled',
    PoolCompleteEvent = 'poolComplete',
    PoolCreatedEvent = 'poolCreated',
    RemoteConnectionChangedEvent = 'remoteConnectionChanged',
    RestoreCompleteEvent = 'restoreComplete',
    RestoreFailedEvent = 'restoreFailed',
    RestoreProgressEvent = 'restoreProgress',
    RestoreStartedEvent = 'restoreStarted',
    ScanDirectoryProgressEvent = 'scanDirectoryProgress',
    ServerGoingDownEvent = 'goingDown',
    ShareUpdatedEvent = 'shareUpdated',
    StartupProgressEvent = 'startupProgress',
    TaskCanceledEvent = 'taskCanceled',
    TaskCompleteEvent = 'taskComplete',
    TaskCreatedEvent = 'taskCreated',
    TaskFailedEvent = 'taskFailure',
    WeblensLoadedEvent = 'weblensLoaded',
    ZipCompleteEvent = 'zipComplete',
    ZipProgressEvent = 'createZipProgress',
}

export enum WsAction {
    CancelTask = 'cancelTask',
    ReportError = 'showWebError',
    ScanDirectory = 'scanDirectory',
    Subscribe = 'subscribe',
    Unsubscribe = 'unsubscribe',
}

export enum WsSubscriptionType {
    Folder = 'folderSubscribe',
    System = 'systemSubscribe',
    Task = 'taskSubscribe',
    TaskType = 'taskTypeSubscribe',
    User = 'userSubscribe',
}
