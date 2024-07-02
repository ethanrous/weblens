import { useCallback, useContext, useEffect, useState } from 'react';
import useWebSocket from 'react-use-websocket';
import { UserContext } from '../Context';

import { FbModeT } from '../Files/filesContext';
import { AuthHeaderT, FBDispatchT, UserContextT, UserInfoT } from '../types/Types';
import { humanFileSize } from '../util';
import { API_WS_ENDPOINT } from './ApiEndpoint';
import { downloadSingleFile, SubToFolder, UnsubFromFolder } from './FileBrowserApi';
import { WeblensFileInfo } from '../Files/File';

function useWeblensSocket() {
    const { usr, authHeader }: UserContextT = useContext(UserContext);
    const [givenUp, setGivenUp] = useState(false);
    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        onOpen: () => {
            setGivenUp(false);
            sendMessage(JSON.stringify({ auth: authHeader.Authorization }));
        },
        reconnectAttempts: 5,
        reconnectInterval: last => {
            return ((last + 1) ^ 2) * 1000;
        },
        shouldReconnect: () => usr.username !== '',
        onReconnectStop: () => {
            setGivenUp(true);
        },
    });
    const wsSend = useCallback(
        (action: string, content: any) => {
            const msg = {
                action: action,
                content: JSON.stringify(content),
            };
            console.log('WSSend', msg);
            sendMessage(JSON.stringify(msg));
        },
        [sendMessage],
    );

    return {
        wsSend,
        lastMessage,
        readyState: givenUp ? -1 : readyState,
    };
}

export function dispatchSync(
    folderIds: string | string[],
    wsSend: (action: string, content: any) => void,
    recursive: boolean,
    full: boolean,
) {
    folderIds = folderIds instanceof Array ? folderIds : [folderIds];
    for (const folderId of folderIds) {
        wsSend('scan_directory', {
            folderId: folderId,
            recursive: recursive,
            full: full,
        });
    }
}

export const useSubscribe = (
    cId: string,
    sId: string,
    mode: FbModeT,
    usr: UserInfoT,
    dispatch: FBDispatchT,
    authHeader: AuthHeaderT,
) => {
    const { wsSend, lastMessage, readyState } = useWeblensSocket();

    useEffect(() => {
        if (readyState === 1 && cId != null && cId !== 'shared' && usr.isLoggedIn) {
            if (cId === usr.homeId) {
                SubToFolder(usr.trashId, sId, wsSend);
                return () => UnsubFromFolder(usr.trashId, wsSend);
            }
            SubToFolder(cId, sId, wsSend);
            return () => UnsubFromFolder(cId, wsSend);
        }
    }, [readyState, sId, cId, usr.homeId, usr.trashId, wsSend]);

    // Subscribe to the home folder if we aren't in it, to be able to update the total disk usage
    useEffect(() => {
        if (!usr.isLoggedIn || readyState !== 1) {
            return;
        }

        SubToFolder(usr.homeId, sId, wsSend);
        return () => UnsubFromFolder(usr.homeId, wsSend);
    }, [usr.homeId, sId, readyState]);

    // Listen for incoming websocket messages
    useEffect(() => {
        HandleWebsocketMessage(lastMessage, sId, usr, dispatch, authHeader);
    }, [lastMessage, usr, authHeader]);

    return { wsSend, readyState };
};

interface wsMsgInfo {
    eventTag: string;
    subscribeKey: string;
    content: wsMsgContent;

    taskType?: string;
    error?: string;
    broadcastType?: boolean;
}

interface wsMsgContent {
    newFile?: WeblensFileInfo;
    fileInfo?: WeblensFileInfo;

    note?: string;
    oldId?: string;
    fileId?: string;
    filename?: string;
    createdBy?: string;
    task_job_name?: string;
    task_job_target?: string;

    totalFiles?: number;
    bytesSoFar?: number;
    bytesTotal?: number;
    speedBytes?: number;
    tasks_total?: number;
    tasks_complete?: number;
    completedFiles?: number;
    execution_time?: number;
    percent_progress?: number;
}

function HandleWebsocketMessage(
    lastMessage: { data: string },
    shareId: string,
    usr: UserInfoT,
    dispatch: FBDispatchT,
    authHeader: AuthHeaderT,
) {
    if (lastMessage) {
        let msgData: wsMsgInfo = JSON.parse(lastMessage.data);
        console.log('WSRecv', msgData);
        if (msgData.error) {
            console.error(msgData.error);
            return;
        }

        try {
            switch (msgData.eventTag) {
                case 'file_created': {
                    dispatch({
                        type: 'create_file',
                        files: [msgData.content.fileInfo],
                        user: usr,
                    });
                    return;
                }

                case 'file_updated': {
                    dispatch({
                        type: 'update_many',
                        files: [msgData.content.fileInfo],
                        user: usr,
                    });
                    return;
                }

                // moved is different from updated because the Id of the file will change
                case 'file_moved': {
                    dispatch({
                        type: 'replace_file',
                        fileId: msgData.content.oldId,
                        fileInfo: msgData.content.newFile,
                    });
                    break;
                }

                case 'file_deleted': {
                    dispatch({
                        type: 'delete_from_map',
                        fileIds: [msgData.content.fileId],
                    });
                    return;
                }

                case 'task_created': {
                    if (msgData.taskType === 'scan_directory') {
                        dispatch({
                            type: 'new_task',
                            serverId: msgData.subscribeKey,
                            taskType: msgData.taskType,
                            target: msgData.content.filename,
                        });
                    }
                    return;
                }

                case 'scan_complete': {
                    dispatch({
                        type: 'task_complete',
                        serverId: msgData.subscribeKey,
                        time: msgData.content.execution_time,
                        note: msgData.content.note,
                    });
                    return;
                }

                case 'task_failure': {
                    dispatch({
                        type: 'task_failure',
                        serverId: msgData.subscribeKey,
                        note: msgData.error,
                    });
                    return;
                }

                // case 'task_progress_update':
                case 'create_zip_progress': {
                    const [size, units] = humanFileSize(msgData.content.speedBytes);
                    dispatch({
                        type: 'update_scan_progress',
                        progress: (msgData.content.bytesSoFar / msgData.content.bytesTotal) * 100,
                        serverId: msgData.subscribeKey,
                        taskType: 'Creating Zip...',
                        target: `Zipping ${msgData.content.totalFiles} files`,
                        fileName: `${size}${units}s`,
                        tasksComplete: msgData.content.completedFiles,
                        tasksTotal: msgData.content.totalFiles,
                        note: 'No note',
                    });
                    return;
                }

                case 'sub_task_complete': {
                    let jobName: string;

                    if (msgData.content.task_job_name) {
                        jobName = msgData.content.task_job_name
                            .replace('_', ' ')
                            .split(' ')
                            .map((s: string) => {
                                return s.charAt(0).toUpperCase() + s.slice(1);
                            })
                            .join(' ')
                            .replace('Directory', 'Folder');
                    }

                    dispatch({
                        type: 'update_scan_progress',
                        progress: msgData.content.percent_progress,
                        serverId: msgData.subscribeKey,
                        taskType: jobName,
                        target: msgData.content.task_job_target,
                        fileName: msgData.content.filename,
                        tasksComplete: msgData.content.tasks_complete,
                        tasksTotal: msgData.content.tasks_total,
                        note: msgData.content.note,
                    });

                    return;
                }

                case 'scan_directory_progress': {
                    dispatch({
                        type: 'set_scan_progress',
                        progress: (1 - msgData.content['remainingTasks'] / msgData.content['totalTasks']) * 100,
                    });
                    return;
                }

                case 'zip_complete': {
                    if (msgData.taskType !== 'create_zip') {
                        return;
                    }
                    dispatch({
                        type: 'task_complete',
                        serverId: msgData.subscribeKey,
                    });
                    downloadSingleFile(msgData.content['takeoutId'], authHeader, dispatch, undefined, 'zip', shareId);
                    return;
                }

                case 'task_complete': {
                    dispatch({
                        type: 'task_complete',
                        serverId: msgData.subscribeKey,
                    });
                    return;
                }

                case 'pool_created': {
                    dispatch({
                        type: 'add_pool_to_progress',
                        serverId: msgData.content.createdBy,
                        poolId: msgData.subscribeKey,
                    });
                    return;
                }

                case 'task_canceled':
                case 'pool_cancelled': {
                    dispatch({
                        type: 'cancel_task',
                        serverId: msgData.subscribeKey,
                    });
                    return;
                }

                case 'error': {
                    console.error(msgData.error);
                    return;
                }

                default: {
                    console.error('Could not parse websocket message type: ', msgData);
                    return;
                }
            }
        } catch (e) {
            console.error('Exception while handling websocket message', e);
        }
    }
}
