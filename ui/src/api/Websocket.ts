import { useSessionStore } from '@weblens/components/UserInfo'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { WeblensFileParams } from '@weblens/types/files/File'
import WeblensMedia from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { UserInfoT } from '@weblens/types/Types'
import { Dispatch, useCallback, useEffect, useState } from 'react'
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
    const [givenUp, setGivenUp] = useState(false)
    const { sendMessage, lastMessage, readyState } = useWebSocket(
        API_WS_ENDPOINT,
        {
            onOpen: () => {
                setGivenUp(false)
                useWebsocketStore.getState().setSender(sendMessage)
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
    const wsSend: WsSendT = useCallback(
        (action: string, content) => {
            const msg = {
                action: action,
                sentAt: Date.now(),
                content: JSON.stringify(content),
            }
            console.log('WSSend', msg)
            sendMessage(JSON.stringify(msg))
        },
        [sendMessage]
    )

    useEffect(() => {
        useWebsocketStore.getState().setReadyState(givenUp ? -1 : readyState)
    }, [readyState, givenUp])
    return {
        wsSend,
        lastMessage,
    }
}

export type WsSendT = (action: string, content: object) => void

export const useSubscribe = (cId: string, sId: string, usr: UserInfoT) => {
    const { wsSend, lastMessage } = useWeblensSocket()
    const readyState = useWebsocketStore((state) => state.readyState)

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

function filebrowserWebsocketHandler(
    shareId: string,
    dispatch: FBSubscribeDispatchT
) {
    return (msgData) => {
        switch (msgData.eventTag) {
            case 'file_created': {
                dispatch.addFile(msgData.content.fileInfo)
                return
            }

            case 'file_updated': {
                if (msgData.content.mediaData) {
                    const newM = new WeblensMedia(msgData.content.mediaData)
                    useMediaStore.getState().addMedias([newM])
                    msgData.content.fileInfo.mediaId = newM.Id()
                }

                useFileBrowserStore
                    .getState()
                    .updateFile(
                        msgData.content.fileInfo,
                        useSessionStore.getState().user
                    )
                return
            }

            // moved is different from updated because the Id of the file will change
            case 'file_moved': {
                dispatch.replaceFile(
                    msgData.content.oldId,
                    msgData.content.newFile
                )
                break
            }

            case 'file_deleted': {
                dispatch.deleteFile(msgData.content.fileId)
                break
            }

            case 'task_created': {
                if (msgData.taskType === 'scan_directory') {
                    useTaskState
                        .getState()
                        .addTask(msgData.subscribeKey, msgData.taskType, {
                            target: msgData.content.filename,
                        })
                } else if (msgData.taskType === 'create_zip') {
                    if (!msgData.content.filenames) {
                        return
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
                return
            }

            case 'scan_complete': {
                tasksDispatch({
                    type: 'task_complete',
                    taskId: msgData.subscribeKey,
                    time: msgData.content.execution_time,
                    note: msgData.content.note,
                })
                return
            }

            case 'task_failure': {
                tasksDispatch({
                    type: 'task_failure',
                    taskId: msgData.subscribeKey,
                    note: msgData.error,
                })
                return
            }

            // case 'task_progress_update':
            case 'create_zip_progress': {
                tasksDispatch({
                    type: 'update_scan_progress',
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
                return
            }

            case 'sub_task_complete': {
                // let jobName: string;

                if (msgData.content.task_job_name) {
                    // jobName = msgData.content.task_job_name
                    //     .replace('_', ' ')
                    //     .split(' ')
                    //     .map((s: string) => {
                    //         return s.charAt(0).toUpperCase() + s.slice(1);
                    //     })
                    //     .join(' ')
                    //     .replace('Directory', 'Folder');
                }

                tasksDispatch({
                    type: 'update_scan_progress',
                    taskId: msgData.subscribeKey,
                    progress: msgData.content.percent_progress,
                    tasksComplete: msgData.content.tasks_complete,
                    tasksFailed: msgData.content.tasks_failed,
                    tasksTotal: msgData.content.tasks_total,
                    target: msgData.content.task_job_target,
                    note: msgData.content.note,
                    workingOn: msgData.content.filename,
                    taskType: msgData.taskType,
                })

                return
            }

            case 'scan_directory_progress': {
                tasksDispatch({
                    type: 'set_scan_progress',
                    progress:
                        (1 -
                            msgData.content['remainingTasks'] /
                                msgData.content['totalTasks']) *
                        100,
                })
                return
            }

            case 'zip_complete': {
                if (msgData.taskType !== 'create_zip') {
                    return
                }
                tasksDispatch({
                    type: 'task_complete',
                    taskId: msgData.subscribeKey,
                })
                downloadSingleFile(
                    msgData.content['takeoutId'],
                    tasksDispatch,
                    msgData.content['filename'],
                    true,
                    shareId
                )
                return
            }

            case 'task_complete': {
                tasksDispatch({
                    type: 'task_complete',
                    taskId: msgData.content.task_id,
                })
                return
            }

            case 'pool_complete': {
                tasksDispatch({
                    type: 'task_complete',
                    taskId: msgData.subscribeKey,
                })
                return
            }

            case 'pool_created': {
                tasksDispatch({
                    type: 'add_pool_to_progress',
                    taskId: msgData.content.createdBy,
                    poolId: msgData.subscribeKey,
                    taskType: msgData.taskType,
                })
                return
            }

            case 'task_canceled':
            case 'pool_cancelled': {
                tasksDispatch({
                    type: 'cancel_task',
                    taskId: msgData.subscribeKey,
                })
                return
            }

            case 'going_down': {
                useWebsocketStore.getState().setReadyState(-1)
                setTimeout(() => location.reload(), 5000)
                return
            }

            case 'weblens_loaded': {
                // NoOp if we are already loaded
                return
            }

            case 'error': {
                console.error(msgData.error)
                return
            }

            case 'core_connection_changed': {
                // NoOp
                return
            }

            default: {
                console.error(
                    'Unknown websocket message type: ',
                    msgData.eventTag
                )
                return
            }
        }
    }
}

export interface WebsocketControlT {
    wsSend: (thing) => void
    readyState: number

    setSender: (sender: (thing) => void) => void
    setReadyState: (readyState: number) => void
}

const WebsocketControl: StateCreator<WebsocketControlT, [], []> = (set) => ({
    wsSend: null,
    readyState: 0,

    setSender: (sender: (thing) => void) => {
        set({
            wsSend: sender,
        })
    },

    setReadyState: (readyState: number) => {
        set({ readyState: readyState })
    },
})

export const useWebsocketStore = create<WebsocketControlT>()(WebsocketControl)
