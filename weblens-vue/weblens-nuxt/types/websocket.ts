import type { TaskType } from './task'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export interface WsMessage<T = any> {
    content: T
    eventTag: WsEvent
    subscribeKey: string
    sentTime: number
    constructedTime: number
    taskType?: TaskType
    taskStartTime?: number
    error?: string
}

export enum WsEvent {
    BackupCompleteEvent = 'backupComplete',
    BackupStartedEvent = 'backupStarted',
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
    RefreshTower = 'refreshTower',
}

export enum WsSubscriptionType {
    Folder = 'folderSubscribe',
    System = 'systemSubscribe',
    Task = 'taskSubscribe',
    TaskType = 'taskTypeSubscribe',
    User = 'userSubscribe',
}
