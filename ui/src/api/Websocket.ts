import { useSessionStore } from '@weblens/components/UserInfo'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { WeblensFileParams } from '@weblens/types/files/File'
import WeblensMedia, { MediaDataT } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { UserInfoT } from '@weblens/types/Types'
import { useCallback, useEffect, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { create, StateCreator } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { API_WS_ENDPOINT } from './ApiEndpoint'
import {
    downloadSingleFile,
    SubToFolder,
    UnsubFromFolder,
} from './FileBrowserApi'
import { useTaskState } from '@weblens/pages/FileBrowser/TaskProgress'

export function useWeblensSocket() {
    const user = useSessionStore((state) => state.user)
    const setLastMessage = useWebsocketStore((state) => state.setLastMessage)
    const setReadyState = useWebsocketStore((state) => state.setReadyState)
    const [givenUp, setGivenUp] = useState(false)
    const { sendMessage, lastMessage, readyState } = useWebSocket(
        API_WS_ENDPOINT,
        {
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
        }
    )

    useEffect(() => {
        const send = (action: string, content) => {
            const msg = {
                action: action,
                sentAt: Date.now(),
                content: JSON.stringify(content),
            }
            console.log('WSSend', msg)
            sendMessage(JSON.stringify(msg))
        }

        useWebsocketStore.getState().setSender(send)
    }, [sendMessage])

    useEffect(() => {
        setLastMessage(lastMessage)
    }, [lastMessage])

    useEffect(() => {
        setReadyState(givenUp ? -1 : readyState)
    }, [readyState, givenUp])
    return {
        lastMessage,
    }
}

export type WsSendT = (action: string, content: object) => void

export const useSubscribe = (cId: string, sId: string, usr: UserInfoT) => {
    const { lastMessage } = useWeblensSocket()
    const readyState = useWebsocketStore((state) => state.readyState)
    const wsSend = useWebsocketStore((state) => state.wsSend)

    const fbDispatch: FBSubscribeDispatchT = useFileBrowserStore(
        useShallow((state) => ({
            addFile: state.addToFilesMap,
            updateFile: (f) => state.updateFile(f, usr),
            replaceFile: state.replaceFile,
            deleteFile: state.deleteFile,
        }))
    )

    useEffect(() => {
        if (useFileBrowserStore.getState().pastTime) {
            return
        }
        if (
            readyState === 1 &&
            cId !== null &&
            cId !== 'shared' &&
            usr !== null
        ) {
            if (cId === usr.homeId) {
                SubToFolder(usr.trashId, sId, wsSend)
                return () => UnsubFromFolder(usr.trashId, wsSend)
            }
            SubToFolder(cId, sId, wsSend)
            return () => UnsubFromFolder(cId, wsSend)
        }
    }, [readyState, sId, cId, usr, wsSend])

    // Subscribe to the home folder if we aren't in it, to be able to update the total disk usage
    useEffect(() => {
        if (useFileBrowserStore.getState().pastTime) {
            return
        }
        if (usr === null || readyState !== 1) {
            return
        }

        SubToFolder(usr.homeId, sId, wsSend)
        return () => UnsubFromFolder(usr.homeId, wsSend)
    }, [usr, sId, readyState])

    // Listen for incoming websocket messages
    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            filebrowserWebsocketHandler(sId, fbDispatch)
        )
    }, [lastMessage, usr])

    return { wsSend, readyState }
}

interface wsMsgInfo {
    eventTag: string
    subscribeKey: string
    content: wsMsgContent

    taskType?: string
    error?: string
}

interface wsMsgContent {
    newFile?: WeblensFileParams
    fileInfo?: WeblensFileParams

    note?: string
    oldId?: string
    fileId?: string
    task_id?: string
    filename?: string
    filenames?: string[]
    createdBy?: string
    task_job_name?: string
    task_job_target?: string

    totalFiles?: number
    bytesSoFar?: number
    bytesTotal?: number
    speedBytes?: number
    tasks_total?: number
    tasks_complete?: number
    tasks_failed?: number
    completedFiles?: number
    execution_time?: number
    percent_progress?: number
}

export function HandleWebsocketMessage(
    lastMessage: { data: string },
    handler: (msgData) => void
) {
    if (lastMessage) {
        const msgData: wsMsgInfo = JSON.parse(lastMessage.data)
        console.log('WSRecv', msgData)
        if (msgData.error) {
            console.error(msgData.error)
            return
        }

        try {
            handler(msgData)
        } catch (e) {
            console.error('Exception while handling websocket message', e)
        }
    }
}

export interface FBSubscribeDispatchT {
    addFile: (info: WeblensFileParams) => void
    updateFile: (info: WeblensFileParams) => void
    replaceFile: (oldId: string, newInfo: WeblensFileParams) => void
    deleteFile: (fileId: string) => void
}

type WsMsgContent = {
    newFile?: WeblensFileParams
    fileInfo?: WeblensFileParams
    mediaData?: MediaDataT
    note?: string
    oldId?: string
    fileId?: string
    task_id?: string
    filename?: string
    filenames?: string[]
    createdBy?: string
    task_job_name?: string
    task_job_target?: string
    totalFiles?: number
    bytesSoFar?: number
    bytesTotal?: number
    speedBytes?: number
    tasks_total?: number
    tasks_complete?: number
    tasks_failed?: number
    completedFiles?: number
    runtime?: number
    percent_progress?: number
    takeoutId?: string
}

export enum WsMsgEvent {
    StartupProgressEvent = 'startup_progress',
    TaskCreatedEvent = 'task_created',
    TaskCompleteEvent = 'task_complete',
    BackupCompleteEvent = 'backup_complete',
    TaskFailedEvent = 'task_failure',
    TaskCanceledEvent = 'task_canceled',
    PoolCreatedEvent = 'pool_created',
    PoolCompleteEvent = 'pool_complete',
    PoolCancelledEvent = 'pool_cancelled',
    FolderScanCompleteEvent = 'folder_scan_complete',
    FileScanCompleteEvent = 'file_scan_complete',
    ScanDirectoryProgressEvent = 'scan_directory_progress',
    FileCreatedEvent = 'file_created',
    FileUpdatedEvent = 'file_updated',
    FileMovedEvent = 'file_moved',
    FileDeletedEvent = 'file_deleted',
    ZipProgressEvent = 'create_zip_progress',
    ZipCompleteEvent = 'zip_complete',
    ServerGoingDownEvent = 'going_down',
    RestoreStartedEvent = 'restore_started',
    WeblensLoadedEvent = 'weblens_loaded',
    ErrorEvent = 'error',
    RemoteConnectionChangedEvent = 'remote_connection_changed',
    BackupProgressEvent = 'backup_progress',

    CopyFileStartedEvent = 'copy_file_started',
    CopyFileCompleteEvent = 'copy_file_complete',
}

export type WsMsg = {
    eventTag: WsMsgEvent
    subscribeKey: string
    content: WsMsgContent
    taskType?: string
    error: string
}

function filebrowserWebsocketHandler(
    shareId: string,
    dispatch: FBSubscribeDispatchT
) {
    return (msgData: WsMsg) => {
        switch (msgData.eventTag) {
            case WsMsgEvent.FileCreatedEvent: {
                dispatch.addFile(msgData.content.fileInfo)
                break
            }

            case WsMsgEvent.FileUpdatedEvent: {
                if (msgData.content.mediaData) {
                    const newM = new WeblensMedia(msgData.content.mediaData)
                    useMediaStore.getState().addMedias([newM])
                    console.log(
                        'Media added',
                        newM,
                        msgData.content.fileInfo.contentId
                    )
                    msgData.content.fileInfo.contentId = newM.Id()
                }

                useFileBrowserStore
                    .getState()
                    .updateFile(
                        msgData.content.fileInfo,
                        useSessionStore.getState().user
                    )
                break
            }

            // moved is different from updated because the Id of the file will change
            case WsMsgEvent.FileMovedEvent: {
                if (
                    msgData.content.oldId === undefined ||
                    msgData.content.newFile === undefined
                ) {
                    console.error('FileMovedEvent missing oldId or newFile')
                    break
                }

                dispatch.replaceFile(
                    msgData.content.oldId,
                    msgData.content.newFile
                )
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

            case WsMsgEvent.TaskCreatedEvent: {
                if (msgData.content.totalFiles === undefined) {
                    console.error('TaskCreatedEvent missing totalFiles')
                    break
                }

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
                        taskType: msgData.taskType,
                    })
                break
            }

            case WsMsgEvent.FileScanCompleteEvent: {
                useTaskState
                    .getState()
                    .updateTaskProgress(msgData.subscribeKey, {
                        progress: msgData.content.percent_progress,
                        tasksComplete: msgData.content.tasks_complete,
                        tasksTotal: msgData.content.tasks_total,
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
                )
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
                const _exhaustiveCheck: never = msgData.eventTag
                console.error(
                    'Unknown websocket message type: ',
                    _exhaustiveCheck
                )
                return
            }
        }
    }
}

export interface WebsocketControlT {
    wsSend: (event: string, content) => void
    readyState: number
    lastMessage

    setSender: (sender: (event: string, content) => void) => void
    setReadyState: (readyState: number) => void
    setLastMessage: (msg) => void
}

const WebsocketControl: StateCreator<WebsocketControlT, [], []> = (set) => ({
    wsSend: () => {
        console.error('Websocket not initialized')
    },
    readyState: 0,
    lastMessage: null,

    setSender: (sender: (event: string, content) => void) => {
        set({
            wsSend: sender,
        })
    },

    setReadyState: (readyState: number) => {
        set({ readyState: readyState })
    },

    setLastMessage: (msg) => {
        set({ lastMessage: msg })
    },
})

export const useWebsocketStore = create<WebsocketControlT>()(WebsocketControl)
