import { fetchJson, wrapRequest } from '@weblens/api/ApiFetch'
import { FbModeT } from '@weblens/pages/FileBrowser/FBStateControl'
import { humanFileSize } from '@weblens/util'
import API_ENDPOINT from './ApiEndpoint'
import { useWebsocketStore, WsSendT } from './Websocket'
import { useTaskState } from '@weblens/pages/FileBrowser/TaskStateControl'
import { FileAction } from '@weblens/pages/FileBrowser/FileBrowserTypes'
import {
    DownloadFileIsTakeoutEnum,
    FilesApiFactory,
    FolderApiFactory,
    FolderInfo,
} from './swag'

export const FileApi = FilesApiFactory(null, API_ENDPOINT)
export const FolderApi = FolderApiFactory(null, API_ENDPOINT)
// new Configuration({ baseOptions: { withCredentials: true } }),

export function SubToFolder(subId: string, shareId: string, wsSend: WsSendT) {
    if (!subId || subId === 'shared') {
        return
    }
    wsSend('folder_subscribe', {
        subscribeKey: subId,
        shareId: shareId,
    })
}

export function SubToTask(
    taskId: string,
    lookingFor: string[],
    wsSend: WsSendT
) {
    wsSend('task_subscribe', {
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
                return Promise.reject(res.statusText)
            }
        })
        .then((blob) => {
            downloadBlob(blob, filename)
        })
}

export async function GetWormholeInfo(shareId: string) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)
    return wrapRequest(fetch(url.toString()))
}

export async function getFilesystemStats(folderId: string): Promise<{
    sizesByExtension: { name: string; size: number }[]
}> {
    return fetchJson(`${API_ENDPOINT}/files/${folderId}/stats`)
}

export async function getFileHistory(
    fileId: string,
    timestamp: Date
): Promise<FileAction[]> {
    if (!fileId) {
        console.error('No fileId trying to get file history')
        return null
    }
    const url = new URL(`${API_ENDPOINT}/file/${fileId}/history`)
    if (timestamp) {
        url.searchParams.append('timestamp', timestamp.getTime().toString())
    }
    return fetchJson(url.toString())
}
