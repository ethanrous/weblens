import { FbModeT } from '@weblens/store/FBStateControl'
import { useTaskState } from '@weblens/store/TaskStateControl'
import { humanFileSize } from '@weblens/util'

import API_ENDPOINT from './ApiEndpoint'
import { WsSendT, useWebsocketStore } from './Websocket'
import { FilesApiFactory, FolderApiFactory, FolderInfo } from './swag'

export const FileApi = FilesApiFactory(null, API_ENDPOINT)
export const FolderApi = FolderApiFactory(null, API_ENDPOINT)

export function SubToFolder(subId: string, shareId: string, wsSend: WsSendT) {
    if (!subId || subId === 'shared') {
        return
    }
    wsSend('folderSubscribe', {
        subscribeKey: subId,
        shareId: shareId,
    })
}

export function SubToTask(
    taskId: string,
    lookingFor: string[],
    wsSend: WsSendT
) {
    wsSend('taskSubscribe', {
        subscribeKey: taskId,
        lookingFor: lookingFor,
    })
}

export function UnsubFromFolder(subId: string, wsSend: WsSendT) {
    if (!subId || useWebsocketStore.getState().readyState < 1) {
        return
    }
    wsSend('unsubscribe', { subscribeKey: subId })
}

export async function GetFolderData(
    folderId: string,
    fbMode: FbModeT,
    shareId?: string,
    viewingTime?: Date
): Promise<FolderInfo> {
    if (fbMode === FbModeT.share && !shareId) {
        const res = await FileApi.getSharedFiles()
        return res.data
    }
    if (fbMode === FbModeT.external) {
        console.error('External files not implemented')
    }

    const res = await FolderApi.getFolder(
        folderId,
        shareId ? shareId : undefined,
        viewingTime?.getTime(),
        { withCredentials: true }
    )
    return res.data
}

function downloadBlob(blob: Blob, filename: string) {
    const aElement = document.createElement('a')
    aElement.setAttribute('download', filename)
    const href = URL.createObjectURL(blob)
    aElement.href = href
    aElement.setAttribute('target', '_blank')
    aElement.click()
    URL.revokeObjectURL(href)
    return
}

export async function downloadSingleFile(
    fileId: string,
    filename: string,
    isZip: boolean,
    shareId: string
) {
    const taskId = `DOWNLOAD_${fileId}`
    useTaskState
        .getState()
        .addTask(taskId, 'download_file', { target: filename })

    return FileApi.downloadFile(fileId, shareId, isZip, {
        responseType: 'blob',
        onDownloadProgress: (p) => {
            const [rateSize, rateUnits] = humanFileSize(p.rate)
            const [bytesSize, bytesUnits] = humanFileSize(p.loaded)
            const [totalSize, totalUnits] = humanFileSize(p.total)
            useTaskState.getState().updateTaskProgress(taskId, {
                progress: p.progress * 100,
                workingOn: `${rateSize}${rateUnits}/s`,
                tasksComplete: `${bytesSize}${bytesUnits}`,
                tasksTotal: `${totalSize}${totalUnits}`,
            })
        },
    })
        .then((res) => {
            if (res.status === 200) {
                useTaskState.getState().handleTaskCompete(taskId, 0, '')
                return new Blob([res.data])
            } else {
                return Promise.reject(new Error(res.statusText))
            }
        })
        .then((blob) => {
            downloadBlob(blob, filename)
        })
}
