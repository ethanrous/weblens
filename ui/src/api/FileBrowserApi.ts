import { FbModeT } from '@weblens/store/FBStateControl'

import API_ENDPOINT from './ApiEndpoint.js'
import { WsSendT, useWebsocketStore } from './Websocket'
import {
    FilesApiAxiosParamCreator,
    FilesApiFactory,
    FolderApiFactory,
    FolderInfo,
} from './swag/api.js'

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

export async function downloadSingleFile(
    fileId: string,
    filename: string,
    isZip: boolean,
    shareId: string,
    format: 'webp' | 'jpeg' | null
) {
    const a = document.createElement('a')
    const paramCreator = FilesApiAxiosParamCreator()
    const args = await paramCreator.downloadFile(fileId, shareId, format, isZip)
    const url = API_ENDPOINT + args.url

    if (isZip) {
        filename = 'weblens_download_' + filename
    }

    a.href = url
    a.download = filename
    a.click()
}
