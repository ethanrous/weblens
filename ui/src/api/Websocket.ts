import {
    SubToFolder,
    UnsubFromFolder,
    downloadSingleFile,
} from '@weblens/api/FileBrowserApi'
import { useSessionStore } from '@weblens/components/UserInfo'
import { DirViewModeT } from '@weblens/pages/FileBrowser/FileBrowserTypes'
import { StartupTask } from '@weblens/pages/Startup/StartupLogic'
import { ShareRoot, useFileBrowserStore } from '@weblens/store/FBStateControl'
import { useMessagesController } from '@weblens/store/MessagesController.js'
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

import { API_WS_ENDPOINT } from './ApiEndpoint.js'
import { FileInfo, MediaInfo } from './swag/api.js'

export function useWeblensSocket() {
    const user = useSessionStore((state) => state.user)
    const setLastMessage = useWebsocketStore((state) => state.setLastMessage)
    const setReadyState = useWebsocketStore((state) => state.setReadyState)
    const [givenUp, setGivenUp] = useState(false)
    const { sendMessage, lastMessage, lastJsonMessage, readyState } =
        useWebSocket<wsMsgInfo>(API_WS_ENDPOINT, {
            onOpen: () => {
                console.debug('WS Connected')
                setGivenUp(false)
            },
            reconnectAttempts: 5,
            reconnectInterval: (last) => {
                return ((last + 1) ^ 2) * 1000
            },
            shouldReconnect: () => user?.username !== '' && givenUp === false,
            onReconnectStop: () => {
                console.debug('WS Reconnect stopped')
                setGivenUp(true)
            },
        })

    useEffect(() => {
        const send: WsSendFunc = ({
            action,
            subscriptionType,
            subscribeKey,
            content,
        }: WsSendProps) => {
            const msg = {
                // TODO: Add a type for this
                action: action,
                broadcastType: subscriptionType,
                sentAt: Date.now(),
                subscribeKey: subscribeKey,
                content: content,
            }
            // console.debug('WSSend', msg)
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

    useEffect(() => {
        const fn = () => {
            if (user?.username !== '' && givenUp) {
                setGivenUp(false)
            }
        }

        window.addEventListener('focus', fn)
        return () => {
            window.removeEventListener('focus', fn)
        }
    })

    return {
        lastJsonMessage,
    }
}

export type WsSendProps = {
    action: WsAction
    subscriptionType?: WsSubscriptionType
    subscribeKey?: string
    content?: object
}

export type WsSendFunc = (props: WsSendProps) => void

export const useSubscribe = () => {
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const viewMode = useFileBrowserStore((state) => state.viewOpts.dirViewMode)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const { lastJsonMessage } = useWeblensSocket()
    const readyState = useWebsocketStore((state) => state.readyState)
    const user = useSessionStore((state) => state.user)

    useEffect(() => {
        if (readyState !== 1) {
            console.log('Websocket not ready, not subscribing')
            return
        }

        // If we don't have a folderInfo, or we are viewing the past,
        // we can't subscribe to anything
        if (!folderInfo || !folderInfo.fromAPI || pastTime.getTime() !== 0) {
            return
        }
        console.log('FOLDER INFO', folderInfo)

        const folderIds: string[] = []
        if (user.isLoggedIn) {
            folderIds.push(user.homeId)
            folderIds.push(user.trashId)
        }
        if (
            folderInfo.Id() !== user.homeId &&
            folderInfo.Id() !== user.trashId &&
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
            SubToFolder(folderId, shareId)
        }

        return () => {
            for (const folderId of folderIds) {
                console.debug('Unsubscribing from', folderId)
                UnsubFromFolder(folderId)
            }
        }
    }, [folderInfo, shareId, viewMode, pastTime, readyState])

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
    eventTag: WsEvent
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
        // console.debug('WsRecv', lastMessage)
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

function filebrowserWebsocketHandler(
    shareId: string,
    dispatch: FBSubscribeDispatchT
) {
    return (msgData: wsMsgInfo) => {
        switch (msgData.eventTag) {
            case WsEvent.FileCreatedEvent: {
                const fileInfo = msgData.content.fileInfo as FileInfo
                dispatch.addFile(fileInfo)
                break
            }

            case WsEvent.FileMovedEvent:
            case WsEvent.FileUpdatedEvent: {
                const fileInfo = msgData.content.fileInfo as FileInfo
                if (msgData.content.mediaData) {
                    const newM = new WeblensMedia(msgData.content.mediaData)
                    useMediaStore.getState().addMedias([newM])
                    fileInfo.contentId = newM.Id()
                }

                useFileBrowserStore.getState().updateFile(fileInfo)
                break
            }

            case WsEvent.FilesMovedEvent:
            case WsEvent.FilesUpdatedEvent: {
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

            case WsEvent.FileDeletedEvent: {
                const fileInfo = msgData.content.fileInfo as FileInfo
                if (fileInfo.id === undefined) {
                    console.error('FileDeletedEvent missing fileId')
                    break
                }

                dispatch.deleteFile(fileInfo.id)
                break
            }

            case WsEvent.FilesDeletedEvent: {
                if (msgData.content.fileIds === undefined) {
                    console.error(
                        WsEvent.FilesDeletedEvent + ' missing fileIds'
                    )
                    break
                }

                useFileBrowserStore
                    .getState()
                    .deleteFiles(msgData.content.fileIds)

                break
            }

            case WsEvent.TaskCreatedEvent: {
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

            case WsEvent.FolderScanCompleteEvent: {
                useTaskState
                    .getState()
                    .handleTaskCompete(
                        msgData.subscribeKey,
                        msgData.content.runtime,
                        msgData.content.note
                    )
                break
            }

            case WsEvent.TaskFailedEvent: {
                useTaskState
                    .getState()
                    .handleTaskFailure(msgData.subscribeKey, msgData.error)
                break
            }

            case WsEvent.ZipProgressEvent: {
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

            case WsEvent.FileScanCompleteEvent: {
                useTaskState
                    .getState()
                    .updateTaskProgress(msgData.subscribeKey, {
                        progress: msgData.content.percentProgress,
                        tasksComplete: msgData.content.tasksComplete,
                        tasksTotal: msgData.content.tasksTotal,
                        tasksFailed: msgData.content.tasksFailed,
                        finished: msgData.content.filename,
                    })
                break
            }

            case WsEvent.FileScanStartedEvent: {
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

            case WsEvent.ScanDirectoryProgressEvent: {
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

            case WsEvent.ZipCompleteEvent: {
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
                    shareId,
                    null
                ).catch(ErrorHandler)
                break
            }

            case WsEvent.TaskCompleteEvent:
            case WsEvent.PoolCompleteEvent: {
                useTaskState
                    .getState()
                    .handleTaskCompete(msgData.subscribeKey, -1, '')
                break
            }

            case WsEvent.TaskCanceledEvent:
            case WsEvent.PoolCancelledEvent: {
                useTaskState.getState().handleTaskCancel(msgData.subscribeKey)
                break
            }

            case WsEvent.ServerGoingDownEvent: {
                useWebsocketStore.getState().setReadyState(-1)
                setTimeout(() => location.reload(), 5000)
                break
            }

            case WsEvent.WeblensLoadedEvent:
            case WsEvent.BackupProgressEvent:
            case WsEvent.StartupProgressEvent:
            case WsEvent.BackupCompleteEvent:
            case WsEvent.PoolCreatedEvent:
            case WsEvent.RestoreStartedEvent:
            case WsEvent.CopyFileStartedEvent:
            case WsEvent.CopyFileCompleteEvent:
            case WsEvent.RemoteConnectionChangedEvent: {
                // NoOp
                return
            }

            case WsEvent.ErrorEvent: {
                useMessagesController.getState().addMessage({
                    title: 'Websocket Error',
                    text: msgData.error,
                    duration: 5000,
                    severity: 'error',
                })
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
    wsSend: WsSendFunc
    readyState: number
    lastMessage: wsMsgInfo

    setSender: (sender: WsSendFunc) => void
    setReadyState: (readyState: number) => void
    setLastMessage: (msg: wsMsgInfo) => void
}

const WebsocketControl: StateCreator<WebsocketControlT, [], []> = (set) => ({
    wsSend: () => {
        console.error('Websocket not initialized')
    },
    readyState: 0,
    lastMessage: null,

    setSender: (sender: WsSendFunc) => {
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
