import type WeblensFile from '~/types/weblensFile'
import { useUserStore } from '~/stores/user.js'
import useLocationStore, { FbModeT } from '~/stores/location.js'
import { WsAction, WsSubscriptionType } from '~/types/websocket.js'
import { API_ENDPOINT, useWeblensAPI } from './AllApi.js'
import useFilesStore from '~/stores/files.js'
import { FilesApiAxiosParamCreator, type FolderInfo, type TakeoutInfo } from '@ethanrous/weblens-api'

export function SubToFolder(subID: string, shareID: string) {
    if (!subID) {
        console.warn('Trying to subscribe to folder with no subscription id')
        return
    } else if (subID === 'shared') {
        return
    }

    useWebsocketStore().send({
        action: WsAction.Subscribe,
        subscriptionType: WsSubscriptionType.Folder,
        subscribeKey: subID,
        content: {
            shareID: shareID,
        },
    })
}

export function SubToTask(taskID: string, lookingFor?: string[]) {
    useWebsocketStore().send({
        action: WsAction.Subscribe,
        subscriptionType: WsSubscriptionType.Task,
        subscribeKey: taskID,
        content: {
            lookingFor: lookingFor,
        },
    })
}

export function ScanDirectory(directory: WeblensFile) {
    useWebsocketStore().send({
        action: WsAction.ScanDirectory,
        content: { folderID: directory.ID(), shareID: useLocationStore().activeShareID },
    })
}

export function CancelTask(taskID: string) {
    console.debug('Cancelling task:', taskID)

    useWebsocketStore().send({ action: WsAction.CancelTask, content: { taskID: taskID } })
}

export function UnsubFromFolder(subID: string) {
    console.debug('Unsubscribing from folder:', subID)
    if (!subID || useWebsocketStore().status !== 'OPEN') {
        return
    }

    useWebsocketStore().send({
        action: WsAction.Unsubscribe,
        subscribeKey: subID,
    })
}

export async function GetTrashChildIds(): Promise<string[]> {
    const { data: folder } = await useWeblensAPI().FoldersApi.getFolder(useUserStore().user.trashID)

    if (!folder || !folder.children) {
        console.error('No children found in trash folder')
        return []
    }

    const childIds = folder.children.map((file) => file.id).filter((id) => id !== undefined)

    return childIds
}

export async function GetFolderData(
    folderID: string,
    fbMode: FbModeT,
    shareID?: string,
    viewingTime?: Date,
): Promise<FolderInfo> {
    if (fbMode === FbModeT.share && !shareID) {
        const res = await useWeblensAPI().FilesAPI.getSharedFiles()
        return res.data
    }

    if (fbMode === FbModeT.external) {
        console.error('External files not implemented')
    }

    if (folderID === '') {
        throw new Error('Folder ID cannot be empty')
    }

    const res = await useWeblensAPI().FoldersApi.getFolder(
        folderID,
        shareID ? shareID : undefined,
        viewingTime?.getTime(),
        {
            withCredentials: true,
        },
    )
    return res.data
}

export type AllowedDownloadFormats = 'webp' | 'jpeg' | 'zip'

export async function downloadSingleFile(fileID: string, filename: string, format?: AllowedDownloadFormats) {
    let formatStr: `image/${Exclude<AllowedDownloadFormats, 'zip'>}` | undefined
    if (format && format !== 'zip') {
        formatStr = `image/${format}`
    }

    const args = await FilesApiAxiosParamCreator().downloadFile(
        fileID,
        useLocationStore().activeShareID,
        formatStr,
        format === 'zip',
    )

    const url = API_ENDPOINT.value + args.url

    if (format === 'zip') {
        filename = 'weblens_download_' + filename
    } else if (format) {
        filename = filename.split('.').slice(0, -1).join('.') + '.' + format
    }

    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()

    a.remove()
}

export async function downloadManyFiles(
    fileIDs: string[],
): Promise<{ taskID?: string; takeoutInfo: Promise<TakeoutInfo> }> {
    const res = await useWeblensAPI().FilesAPI.createTakeout(
        {
            fileIDs: fileIDs,
        },
        useLocationStore().activeShareID,
    )

    if (res.status === 202) {
        const taskID = res.data.taskID
        if (!taskID) {
            return Promise.reject(new Error('No task ID returned for takeout creation'))
        }

        SubToTask(taskID)

        return {
            taskID: taskID,
            takeoutInfo: new Promise<TakeoutInfo>((resolve) => {
                useTasksStore().setTaskPromise({ resolve, taskID })
            }),
        }
    } else if (res.status === 200) {
        return {
            takeoutInfo: Promise.resolve(res.data),
        }
    } else {
        return Promise.reject(new Error(`Unexpected response status: ${res.status}`))
    }
}

export async function moveFiles(target: WeblensFile) {
    const filesStore = useFilesStore()

    const selectedIds = [...filesStore.selectedFiles]

    filesStore.setMovedFile(selectedIds, true)

    filesStore.setDragging(false)

    await useWeblensAPI().FilesAPI.moveFiles(
        {
            fileIDs: selectedIds,
            newParentID: target.ID(),
        },
        useLocationStore().activeShareID,
    )
}
