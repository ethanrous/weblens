import { useCallback, useContext, useEffect, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { API_WS_ENDPOINT } from './ApiEndpoint'
import { UserContext } from '../Context'
import {
    AuthHeaderT,
    FBDispatchT,
    UserContextT,
    UserInfoT,
    WsMessageT,
} from '../types/Types'
import {
    downloadSingleFile,
    SubToFolder,
    UnsubFromFolder,
} from './FileBrowserApi'
import { humanFileSize } from '../util'
import { FbModeT } from '../Pages/FileBrowser/FileBrowser'

function useWeblensSocket() {
    const { usr, authHeader }: UserContextT = useContext(UserContext)
    const [givenUp, setGivenUp] = useState(false)
    const { sendMessage, lastMessage, readyState } = useWebSocket(
        API_WS_ENDPOINT,
        {
            onOpen: () => {
                setGivenUp(false)
                sendMessage(JSON.stringify({ auth: authHeader.Authorization }))
            },
            reconnectAttempts: 5,
            reconnectInterval: (last) => {
                return ((last + 1) ^ 2) * 1000
            },
            shouldReconnect: () => usr.username !== '',
            onReconnectStop: () => {
                setGivenUp(true)
            },
        }
    )
    const wsSend = useCallback(
        (action: string, content: any) => {
            const msg = {
                action: action,
                content: JSON.stringify(content),
            }
            console.log('WSSend', msg)
            sendMessage(JSON.stringify(msg))
        },
        [sendMessage]
    )

    return {
        wsSend,
        lastMessage,
        readyState: givenUp ? -1 : readyState,
    }
}

export function dispatchSync(
    folderIds: string | string[],
    wsSend: (action: string, content: any) => void,
    recursive: boolean,
    full: boolean
) {
    folderIds = folderIds instanceof Array ? folderIds : [folderIds]
    for (const folderId of folderIds) {
        wsSend('scan_directory', {
            folderId: folderId,
            recursive: recursive,
            full: full,
        })
    }
}

export const useSubscribe = (
    cId: string,
    sId: string,
    mode: FbModeT,
    usr: UserInfoT,
    dispatch: FBDispatchT,
    authHeader: AuthHeaderT
) => {
    const { wsSend, lastMessage, readyState } = useWeblensSocket()

    useEffect(() => {
        if (
            readyState === 1 &&
            cId != null &&
            cId !== 'shared' &&
            usr.isLoggedIn
        ) {
            if (cId === usr.homeId) {
                SubToFolder(usr.trashId, sId, wsSend)
                return () => UnsubFromFolder(usr.trashId, wsSend)
            }
            SubToFolder(cId, sId, wsSend)
            return () => UnsubFromFolder(cId, wsSend)
        }
    }, [readyState, sId, cId, usr.homeId, usr.trashId, wsSend])

    // Subscribe to the home folder if we aren't in it, to be able to update the total disk usage
    useEffect(() => {
        if (!usr.isLoggedIn || readyState !== 1) {
            return
        }

        SubToFolder(usr.homeId, sId, wsSend)
        return () => UnsubFromFolder(usr.homeId, wsSend)
    }, [usr.homeId, sId, readyState])

    // Listen for incoming websocket messages
    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            cId,
            mode,
            usr,
            dispatch,
            authHeader
        )
    }, [lastMessage, usr, authHeader])

    return { wsSend, readyState }
}

function HandleWebsocketMessage(
    lastMessage,
    contentId: string,
    fbMode: FbModeT,
    usr: UserInfoT,
    dispatch: FBDispatchT,
    authHeader: AuthHeaderT
) {
    if (lastMessage) {
        let msgData: WsMessageT = JSON.parse(lastMessage.data)
        console.log('WSRecv', msgData)
        if (msgData.error) {
            console.error(msgData.error)
            return
        }
        if (!msgData.content || !msgData.content[0]) {
            console.error('Got empty content in websocket update')
            return
        }

        switch (msgData.eventTag) {
            case 'file_created': {
                dispatch({
                    type: 'create_file',
                    files: msgData.content.map((v) => v.fileInfo),
                    user: usr,
                })
                return
            }

            case 'file_updated': {
                const files = msgData.content.map((v) => {
                    return v.fileInfo
                })
                dispatch({
                    type: 'update_many',
                    files: files,
                    user: usr,
                })
                return
            }

            // moved is different from updated because the Id of the file will change
            case 'file_moved': {
                msgData.content.map((m) => {
                    dispatch({
                        type: 'replace_file',
                        fileId: m.oldId,
                        fileInfo: m.newFile,
                    })
                })
                return
            }

            case 'file_deleted': {
                dispatch({
                    type: 'delete_from_map',
                    fileIds: msgData.content.map((v) => v.fileId),
                })
                return
            }

            case 'task_created': {
                msgData.content.map((e) => {
                    if (e.taskType === 'scan_directory') {
                        dispatch({
                            type: 'new_task',
                            taskId: msgData.subscribeKey,
                            taskType: e.taskType,
                            target: e.directoryName,
                        })
                    }
                })
                return
            }

            case 'scan_complete': {
                dispatch({
                    type: 'scan_complete',
                    taskId: msgData.subscribeKey,
                    time: msgData.content[0].execution_time,
                    note: msgData.content[0].note,
                })
                return
            }

            case 'task_failure': {
                dispatch({
                    type: 'task_failure',
                    taskId: msgData.subscribeKey,
                    note: msgData.content[0].failure_note,
                })
                return
            }

            case 'create_zip_progress': {
                const [size, units] = humanFileSize(
                    msgData.content[0].result.speedBytes
                )
                dispatch({
                    type: 'update_scan_progress',
                    progress:
                        (msgData.content[0].result.completedFiles /
                            msgData.content[0].result.totalFiles) *
                        100,
                    taskId: msgData.subscribeKey,
                    taskType: 'Creating Zip...',
                    target: 'Zip',
                    fileName: `${size}${units}s`,
                    tasksComplete: msgData.content[0].result.completedFiles,
                    tasksTotal: msgData.content[0].result.totalFiles,
                    note: 'No note',
                })
                return
            }

            case 'sub_task_complete': {
                let jobName: string
                if (msgData.content[0].task_job_name) {
                    jobName = msgData.content[0].task_job_name
                        .replace('_', ' ')
                        .split(' ')
                        .map((s: string) => {
                            return s.charAt(0).toUpperCase() + s.slice(1)
                        })
                        .join(' ')
                        .replace('Directory', 'Folder')
                }

                dispatch({
                    type: 'update_scan_progress',
                    progress: msgData.content[0].percent_progress,
                    taskId: msgData.subscribeKey,
                    taskType: jobName,
                    target: msgData.content[0].task_job_target,
                    fileName: msgData.content[0].filename,
                    tasksComplete: msgData.content[0].tasks_complete,
                    tasksTotal: msgData.content[0].tasks_total,
                    note: msgData.content[0].note,
                })
                return
            }

            case 'scan_directory_progress': {
                dispatch({
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
                downloadSingleFile(
                    msgData.content[0].result['takeoutId'],
                    authHeader,
                    dispatch,
                    undefined,
                    'zip',
                    fbMode === FbModeT.share ? contentId : ''
                )
                return
            }
            case 'error': {
                console.error(msgData.error)
                return
            }

            case 'task_complete': {
                return
            }

            default: {
                console.error(
                    'Could not parse websocket message type: ',
                    msgData
                )
                return
            }
        }
    }
}
