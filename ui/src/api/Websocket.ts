import { useSessionStore } from '@weblens/components/UserInfo'
import { DirViewModeT } from '@weblens/pages/FileBrowser/FileBrowserTypes'
import { StartupTask } from '@weblens/pages/Startup/StartupLogic'
import { ShareRoot, useFileBrowserStore } from '@weblens/store/FBStateControl'
import {
    TaskStageT,
    TaskType,
    useTaskState,
} from '@weblens/store/TaskStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import WeblensMedia from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { useEffect, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { StateCreator, create } from 'zustand'

import { API_WS_ENDPOINT } from './ApiEndpoint'
import {
    SubToFolder,
    UnsubFromFolder,
    downloadSingleFile,
} from './FileBrowserApi'
import { FileInfo, MediaInfo } from './swag'

export function useWeblensSocket() {
    const user = useSessionStore((state) => state.user)
    const setLastMessage = useWebsocketStore((state) => state.setLastMessage)
    const setReadyState = useWebsocketStore((state) => state.setReadyState)
    const [givenUp, setGivenUp] = useState(false)
    const { sendMessage, lastMessage, lastJsonMessage, readyState } =
        useWebSocket<wsMsgInfo>(API_WS_ENDPOINT, {
            onOpen: () => {
                setGivenUp(false)
            },
            reconnectAttempts: 5,
            reconnectInterval: (last) => {
                return ((last + 1) ^ 2) * 1000
            },
            shouldReconnect: () => user?.username !== '',
            onReconnectStop: () => {
                setGivenUp(true)
            },
        })

    useEffect(() => {
        const send = (action: string, content: object) => {
            const msg = {
                action: action,
                sentAt: Date.now(),
                content: JSON.stringify(content),
            }
            console.debug('WSSend', msg)
            sendMessage(JSON.stringify(msg))
        }

        useWebsocketStore.getState().setSender(send)
    }, [sendMessage])

    useEffect(() => {
        setLastMessage(lastJsonMessage)
    }, [lastMessage])

    useEffect(() => {
        setReadyState(givenUp ? -1 : readyState)
    }, [readyState, givenUp])

    return {
        lastJsonMessage,
    }
}

export type WsSendT = (action: string, content: object) => void

export const useSubscribe = () => {
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const viewMode = useFileBrowserStore((state) => state.viewOpts.dirViewMode)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const { lastJsonMessage } = useWeblensSocket()
    const readyState = useWebsocketStore((state) => state.readyState)
    const wsSend = useWebsocketStore((state) => state.wsSend)
    const user = useSessionStore((state) => state.user)

    useEffect(() => {
        // If we don't have a folderInfo, or we are viewing the past,
        // we can't subscribe to anything
        if (!folderInfo || pastTime.getTime() !== 0) {
            return
        }
        const folderIds: string[] = []
        if (user) {
            folderIds.push(user.homeId)
            folderIds.push(user.trashId)
        }
        if (
            folderInfo.Id() !== user?.homeId &&
            folderInfo.Id() !== user?.trashId &&
            folderInfo.Id() !== ShareRoot.Id()
        ) {
            folderIds.push(folderInfo.Id())
        }

        if (viewMode === DirViewModeT.Columns) {
            for (const parent of folderInfo.parents) {
                if (!folderIds.includes(parent.Id())) {
                    folderIds.push(parent.Id())
                }
            }
        }

        for (const folderId of folderIds) {
            console.debug('Subscribing to', folderId)
            SubToFolder(folderId, shareId, wsSend)
        }

        return () => {
            for (const folder of folderIds) {
                UnsubFromFolder(folder, wsSend)
            }
        }
    }, [folderInfo, shareId, viewMode, pastTime])

    // Listen for incoming websocket messages
    useEffect(() => {
        HandleWebsocketMessage(
            lastJsonMessage,
            filebrowserWebsocketHandler(shareId, useFileBrowserStore.getState())
        )
    }, [lastJsonMessage, user])
    return { readyState }
}

export interface wsMsgInfo {
    eventTag: WsMsgEvent
    subscribeKey: string
    content: wsMsgContent

    relaySource?: string

    taskType?: string
    error?: string
}

interface wsMsgContent {
    newFile?: FileInfo
    fileInfo?: FileInfo
    filesInfo?: FileInfo[]

    note?: string
    oldId?: string
    fileId?: string
    fileIds?: string[]
    taskId?: string
    filename?: string
    waitingOn?: StartupTask[]
    filenames?: string[]
    createdBy?: string
    taskJobName?: string
    taskJobTarget?: string
    stage?: string
    filesTotal?: number
    filesRestored?: number
    timestamp?: string
    error?: string
    stages?: TaskStageT[]
    coreId?: string
    totalTime?: number
    taskType?: string

    totalFiles?: number
    bytesSoFar?: number
    bytesTotal?: number
    speedBytes?: number
    tasksTotal?: number
    tasksComplete?: number
    tasksFailed?: number
    completedFiles?: number
    executionTime?: number
    percentProgress?: number
    mediaData?: MediaInfo
    mediaDatas?: MediaInfo[]
    runtime?: number
    takeoutId?: string
}

export function HandleWebsocketMessage(
    lastMessage: wsMsgInfo,
    handler: (msgData: wsMsgInfo) => void
) {
    if (lastMessage) {
        console.debug('WsRecv', lastMessage)
        try {
            handler(lastMessage)
        } catch (e) {
            console.error('Exception while handling websocket message', e)
        }
    }
}

export interface FBSubscribeDispatchT {
    addFile: (info: FileInfo) => void
    updateFile: (info: FileInfo) => void
    deleteFile: (fileId: string) => void
}

export enum WsMsgEvent {
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

export enum WsMsgAction {
    CancelTaskAction = 'cancelTask',
}

function filebrowserWebsocketHandler(
    shareId: string,
    dispatch: FBSubscribeDispatchT
) {
    return (msgData: wsMsgInfo) => {
        switch (msgData.eventTag) {
            case WsMsgEvent.FileCreatedEvent: {
                dispatch.addFile(msgData.content.fileInfo)
                break
            }

            case WsMsgEvent.FileMovedEvent:
            case WsMsgEvent.FileUpdatedEvent: {
                if (msgData.content.mediaData) {
                    const newM = new WeblensMedia(msgData.content.mediaData)
                    useMediaStore.getState().addMedias([newM])
                    msgData.content.fileInfo.contentId = newM.Id()
                }

                useFileBrowserStore
                    .getState()
                    .updateFile(msgData.content.fileInfo)
                break
            }

            case WsMsgEvent.FilesMovedEvent:
            case WsMsgEvent.FilesUpdatedEvent: {
                if (msgData.content.mediaDatas) {
                    const newMs: WeblensMedia[] = []
                    for (const mediaData of msgData.content.mediaDatas) {
                        newMs.push(new WeblensMedia(mediaData))
                    }
                    useMediaStore.getState().addMedias(newMs)
                }

                if (!msgData.content?.filesInfo) {
                    console.error('FilesMovedEvent missing filesInfo')
                    break
                }
                const state = useFileBrowserStore.getState()
                const homeId = useSessionStore.getState().user.homeId
                if (
                    state.folderInfo.Id() != homeId &&
                    msgData.subscribeKey === homeId
                ) {
                    break
                }
                new Promise(() => {
                    state.updateFiles(msgData.content.filesInfo)
                }).catch(ErrorHandler)
                break
            }

            case WsMsgEvent.FileDeletedEvent: {
                if (msgData.content.fileId === undefined) {
                    console.error('FileDeletedEvent missing fileId')
                    break
                }

                dispatch.deleteFile(msgData.content.fileId)
                break
            }

            case WsMsgEvent.FilesDeletedEvent: {
                if (msgData.content.fileIds === undefined) {
                    console.error(
                        WsMsgEvent.FilesDeletedEvent + ' missing fileIds'
                    )
                    break
                }

                useFileBrowserStore
                    .getState()
                    .deleteFiles(msgData.content.fileIds)

                break
            }

            case WsMsgEvent.TaskCreatedEvent: {
                // if (msgData.content.totalFiles === undefined) {
                //     console.error('TaskCreatedEvent missing totalFiles')
                //     break
                // }

                if (msgData.taskType === 'scan_directory') {
                    useTaskState
                        .getState()
                        .addTask(msgData.subscribeKey, msgData.taskType, {
                            target: msgData.content.filename,
                        })
                } else if (msgData.taskType === 'create_zip') {
                    if (!msgData.content.filenames) {
                        break
                    }
                    let target = msgData.content.filenames[0]
                    if (msgData.content.filenames.length > 1) {
                        target = `${target} +${msgData.content.filenames.length - 1}`
                    }
                    useTaskState
                        .getState()
                        .addTask(msgData.subscribeKey, msgData.taskType, {
                            target: target,
                            progress: 0 / msgData.content.totalFiles,
                        })
                }
                break
            }

            case WsMsgEvent.FolderScanCompleteEvent: {
                useTaskState
                    .getState()
                    .handleTaskCompete(
                        msgData.subscribeKey,
                        msgData.content.runtime,
                        msgData.content.note
                    )
                break
            }

            case WsMsgEvent.TaskFailedEvent: {
                useTaskState
                    .getState()
                    .handleTaskFailure(msgData.subscribeKey, msgData.error)
                break
            }

            case WsMsgEvent.ZipProgressEvent: {
                if (
                    msgData.content.bytesSoFar === undefined ||
                    msgData.content.bytesTotal === undefined
                ) {
                    console.error(
                        'ZipProgressEvent missing bytesSoFar or bytesTotal'
                    )
                    break
                }

                useTaskState
                    .getState()
                    .updateTaskProgress(msgData.subscribeKey, {
                        progress:
                            (msgData.content.bytesSoFar /
                                msgData.content.bytesTotal) *
                            100,
                        taskId: msgData.subscribeKey,
                        tasksComplete: msgData.content.completedFiles,
                        tasksTotal: msgData.content.totalFiles,
                        note: 'No note',
                        taskType: msgData.taskType as TaskType,
                    })
                break
            }

            case WsMsgEvent.FileScanCompleteEvent: {
                useTaskState
                    .getState()
                    .updateTaskProgress(msgData.subscribeKey, {
                        progress: msgData.content.percentProgress,
                        tasksComplete: msgData.content.tasksComplete,
                        tasksTotal: msgData.content.tasksTotal,
                        tasksFailed: msgData.content.tasksFailed,
                        workingOn: msgData.content.filename,
                    })
                break
            }

            case WsMsgEvent.ScanDirectoryProgressEvent: {
                useTaskState
                    .getState()
                    .updateTaskProgress(msgData.subscribeKey, {
                        progress:
                            (1 -
                                msgData.content['remainingTasks'] /
                                    msgData.content['totalTasks']) *
                            100,
                    })
                break
            }

            case WsMsgEvent.ZipCompleteEvent: {
                if (msgData.taskType !== 'create_zip') {
                    break
                }

                if (
                    msgData.content.takeoutId === undefined ||
                    msgData.content.filename === undefined
                ) {
                    console.error(
                        'ZipCompleteEvent missing takeoutId or filename'
                    )
                    break
                }

                useTaskState
                    .getState()
                    .handleTaskCompete(msgData.subscribeKey, -1, '')

                downloadSingleFile(
                    msgData.content.takeoutId,
                    msgData.content.filename,
                    true,
                    shareId
                ).catch(ErrorHandler)
                break
            }

            case WsMsgEvent.TaskCompleteEvent:
            case WsMsgEvent.PoolCompleteEvent: {
                useTaskState
                    .getState()
                    .handleTaskCompete(msgData.subscribeKey, -1, '')
                break
            }

            case WsMsgEvent.TaskCanceledEvent:
            case WsMsgEvent.PoolCancelledEvent: {
                useTaskState.getState().handleTaskCancel(msgData.subscribeKey)
                break
            }

            case WsMsgEvent.ServerGoingDownEvent: {
                useWebsocketStore.getState().setReadyState(-1)
                setTimeout(() => location.reload(), 5000)
                break
            }

            case WsMsgEvent.WeblensLoadedEvent:
            case WsMsgEvent.BackupProgressEvent:
            case WsMsgEvent.StartupProgressEvent:
            case WsMsgEvent.BackupCompleteEvent:
            case WsMsgEvent.PoolCreatedEvent:
            case WsMsgEvent.RestoreStartedEvent:
            case WsMsgEvent.CopyFileStartedEvent:
            case WsMsgEvent.CopyFileCompleteEvent:
            case WsMsgEvent.RemoteConnectionChangedEvent: {
                // NoOp
                return
            }

            case WsMsgEvent.ErrorEvent: {
                console.error(msgData.error)
                return
            }

            default: {
                // const _exhaustiveCheck: never = msgData.eventTag
                console.error(
                    'Unknown websocket message type: '
                    // _exhaustiveCheck
                )
                return
            }
        }
    }
}

export interface WebsocketControlT {
    wsSend: WsSendT
    readyState: number
    lastMessage: wsMsgInfo

    setSender: (sender: WsSendT) => void
    setReadyState: (readyState: number) => void
    setLastMessage: (msg: wsMsgInfo) => void
}

const WebsocketControl: StateCreator<WebsocketControlT, [], []> = (set) => ({
    wsSend: () => {
        console.error('Websocket not initialized')
    },
    readyState: 0,
    lastMessage: null,

    setSender: (sender: WsSendT) => {
        set({
            wsSend: sender,
        })
    },

    setReadyState: (readyState: number) => {
        set({ readyState: readyState })
    },

    setLastMessage: (msg: wsMsgInfo) => {
        set({ lastMessage: msg })
    },
})

export const useWebsocketStore = create<WebsocketControlT>()(WebsocketControl)
