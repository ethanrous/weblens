import { useSessionStore } from '@weblens/components/UserInfo.js'
import { FbModeT, useFileBrowserStore } from '@weblens/store/FBStateControl'
import WeblensFile from '@weblens/types/files/File.js'

import API_ENDPOINT from './ApiEndpoint.js'
import { WsAction, WsSubscriptionType, useWebsocketStore } from './Websocket'
import { FilesApiAxiosParamCreator, FilesApiFactory, FolderApiFactory, FolderInfo } from './swag/api.js'
import { Configuration } from './swag/configuration.js'

export const FileApi = FilesApiFactory({} as Configuration, API_ENDPOINT)
export const FolderApi = FolderApiFactory({} as Configuration, API_ENDPOINT)

export function SubToFolder(subId: string, shareId: string) {
    if (!subId) {
        console.trace('Empty subId')
        return
    } else if (subId === 'shared') {
        return
    }

    const wsSend = useWebsocketStore.getState().wsSend

    wsSend({
        action: WsAction.Subscribe,
        subscriptionType: WsSubscriptionType.Folder,
        subscribeKey: subId,
        content: {
            shareId: shareId,
        },
    })
}

export function SubToTask(taskId: string, lookingFor: string[]) {
    const wsSend = useWebsocketStore.getState().wsSend

    wsSend({
        action: WsAction.Subscribe,
        subscriptionType: WsSubscriptionType.Task,
        subscribeKey: taskId,
        content: {
            lookingFor: lookingFor,
        },
    })
}

export function ScanDirectory(directory: WeblensFile) {
    const wsSend = useWebsocketStore.getState().wsSend
    const shareId = useFileBrowserStore.getState().shareId

    wsSend({
        action: WsAction.ScanDirectory,
        content: { folderId: directory.Id(), shareId: shareId },
    })
}

export function CancelTask(taskId: string) {
    const wsSend = useWebsocketStore.getState().wsSend

    wsSend({ action: WsAction.CancelTask, content: { taskId: taskId } })
}

export function UnsubFromFolder(subId: string) {
    if (!subId || useWebsocketStore.getState().readyState < 1) {
        return
    }

    useWebsocketStore.getState().wsSend({
        action: WsAction.Unsubscribe,
        subscribeKey: subId,
    })
}

export async function GetTrashChildIds(): Promise<string[]> {
    const { data: folder } = await FolderApi.getFolder(useSessionStore.getState().user.trashId)

    if (!folder || !folder.children) {
        console.error('No children found in trash folder')
        return []
    }

    const childIds = folder.children.map((file) => file.id).filter((id) => id !== undefined)

    return childIds
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

    if (folderId === '') {
        throw new Error('Folder ID cannot be empty')
    }

    const res = await FolderApi.getFolder(folderId, shareId ? shareId : undefined, viewingTime?.getTime(), {
        withCredentials: true,
    })
    return res.data
}

export type AllowedDownloadFormats = 'webp' | 'jpeg' | 'zip'

export async function downloadSingleFile(fileId: string, filename: string, shareId: string, format?: AllowedDownloadFormats) {
    const a = document.createElement('a')
    const paramCreator = FilesApiAxiosParamCreator()
    const args = await paramCreator.downloadFile(
        fileId,
        shareId,
        format ? `image/${format}` : undefined,
        format === 'zip'
    )
    const url = API_ENDPOINT + args.url

    if (format === 'zip') {
        filename = 'weblens_download_' + filename
    } else if (format) {
        filename = filename.split('.').slice(0, -1).join('.') + '.' + format
    }

    console.log('DOWNLOADING', url, filename)

    a.href = url
    a.download = filename
    a.click()
}
