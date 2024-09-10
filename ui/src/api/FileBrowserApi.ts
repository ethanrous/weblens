import { fetchJson, wrapRequest } from '@weblens/api/ApiFetch'
import { FbModeT } from '@weblens/pages/FileBrowser/FBStateControl'
import { FileAction } from '@weblens/pages/FileBrowser/FileInfoPane'
import { WeblensFile, WeblensFileParams } from '@weblens/types/files/File'
import { AlbumData, TPDispatchT } from '@weblens/types/Types'
import { humanFileSize } from '@weblens/util'
import axios from 'axios'
import API_ENDPOINT from './ApiEndpoint'
import { useWebsocketStore, WsSendT } from './Websocket'

export function SubToFolder(subId: string, shareId: string, wsSend: WsSendT) {
    if (!subId || subId === 'shared') {
        return
    }
    wsSend('folder_subscribe', {
        subscribeKey: subId,
        shareId: shareId,
    })
}

export function SubToTask(taskId: string, lookingFor: string[], wsSend) {
    wsSend('task_subscribe', {
        subscribeKey: taskId,
        lookingFor: lookingFor,
    })
}

export function UnsubFromFolder(subId: string, wsSend) {
    if (!subId || useWebsocketStore.getState().readyState < 1) {
        return
    }
    wsSend('unsubscribe', { subscribeKey: subId })
}

export function TrashFiles(fileIds: string[], shareId: string) {
    const url = new URL(`${API_ENDPOINT}/files/trash`)
    if (shareId) {
        url.searchParams.append('shareId', shareId)
    }

    return wrapRequest(
        fetch(url.toString(), {
            method: 'PATCH',
            body: JSON.stringify(fileIds),
        })
    )
}

export function DeleteFiles(fileIds: string[]) {
    const url = new URL(`${API_ENDPOINT}/files`)

    return fetch(url.toString(), {
        method: 'DELETE',
        body: JSON.stringify(fileIds),
    }).catch((r) => {
        console.error(r)
        return { ok: false }
    })
}

export function UnTrashFiles(fileIds: string[]) {
    const url = new URL(`${API_ENDPOINT}/files/untrash`)

    return fetch(url.toString(), {
        method: 'PATCH',
        body: JSON.stringify(fileIds),
    })
}

async function getSharedWithMe() {
    const url = new URL(`${API_ENDPOINT}/files/shared`)
    return fetch(url.toString())
        .then((res) => res.json())
        .then((sharedFiles) => {
            const sharedFolder = new WeblensFile({
                id: 'shared',
                isDir: true,
                filename: 'Shared',
            })
            return { children: sharedFiles, self: sharedFolder }
        })
}

async function getExternalFiles(contentId: string) {
    const url = new URL(`${API_ENDPOINT}/files/external/${contentId}`)
    return fetchJson(url.toString())
    // .then((data) => {
    //     const ret = {
    //         self: data.self,
    //         parents: data.parents,
    //         children: [],
    //     }
    //     if (data.children) {
    //         ret.children = data.children
    //     } else if (data.files) {
    //         ret.children = data.files
    //     }
    //     return ret
    // })
}

export async function GetFileInfo(
    fileId: string,
    shareId: string
): Promise<WeblensFileParams> {
    const url = new URL(`${API_ENDPOINT}/file/${fileId}`)
    if (shareId !== '') {
        url.searchParams.append('shareId', shareId)
    }
    return fetchJson(url.toString())
}

export async function GetFolderData(
    contentId: string,
    fbMode: FbModeT,
    shareId: string
) {
    if (fbMode === FbModeT.share && !shareId) {
        return getSharedWithMe()
    }
    if (fbMode === FbModeT.external) {
        return getExternalFiles(contentId)
    }

    if (!contentId) {
        console.error('Tried to get folder with no id')
        return
    }

    const url = new URL(`${API_ENDPOINT}/folder/${contentId}`)
    if (fbMode === FbModeT.share) {
        url.searchParams.append('shareId', shareId)
    }

    return fetchJson(url.toString())
}

export async function CreateFolder(
    parentFolderId: string,
    name: string,
    children: string[],
    isPublic: boolean,
    shareId: string
): Promise<string> {
    if (isPublic && !shareId) {
        throw new Error('Attempting to do public upload with no shareId')
    }

    let url
    if (isPublic) {
        url = new URL(`${API_ENDPOINT}/public/folder`)
        url.searchParams.append('shareId', shareId)
    } else {
        url = new URL(`${API_ENDPOINT}/folder`)
    }

    const dirInfo = await fetch(url.toString(), {
        method: 'POST',

        body: JSON.stringify({
            parentFolderId: parentFolderId,
            newFolderName: name,
            children: children,
        }),
    })
        .then((res) => res.json())
        .catch((r) => {
            console.error(`Could not create folder: ${r}`)
        })
    return dirInfo?.folderId
}

export function moveFiles(fileIds: string[], newParentId: string) {
    const url = new URL(`${API_ENDPOINT}/files`)
    const body = {
        fileIds: fileIds,
        newParentId: newParentId,
    }

    return fetch(url.toString(), {
        method: 'PATCH',

        body: JSON.stringify(body),
    })
}

export async function RenameFile(fileId: string, newName: string) {
    const url = new URL(`${API_ENDPOINT}/file/${fileId}`)
    fetch(url.toString(), {
        method: 'PATCH',
        body: JSON.stringify({ newName: newName }),
    })
}

function downloadBlob(blob, filename) {
    const aElement = document.createElement('a')
    aElement.setAttribute('download', filename)
    const href = URL.createObjectURL(blob)
    aElement.href = href
    aElement.setAttribute('target', '_blank')
    aElement.click()
    URL.revokeObjectURL(href)
    return
}

export function downloadSingleFile(
    fileId: string,
    progDispatch: TPDispatchT,
    filename: string,
    isZip: boolean,
    shareId: string
) {
    if (!fileId) {
        console.error('Trying to download without file id!')
        return
    }

    let url
    if (isZip) {
        url = new URL(`${API_ENDPOINT}/takeout/${fileId}`)
    } else {
        url = new URL(`${API_ENDPOINT}/file/${fileId}/download`)
    }
    if (shareId) {
        url.searchParams.append('shareId', shareId)
    }

    const taskId = `DOWNLOAD_${fileId}`
    progDispatch({
        type: 'new_task',
        taskId: taskId,
        taskType: 'download_file',
        target: filename,
    })

    return axios
        .get(url.toString(), {
            responseType: 'blob',

            onDownloadProgress: (p) => {
                const [rateSize, rateUnits] = humanFileSize(p.rate)
                const [bytesSize, bytesUnits] = humanFileSize(p.loaded)
                const [totalSize, totalUnits] = humanFileSize(p.total)
                progDispatch({
                    type: 'update_scan_progress',
                    progress: p.progress * 100,
                    taskId: taskId,
                    workingOn: `${rateSize}${rateUnits}/s`,
                    tasksComplete: `${bytesSize}${bytesUnits}`,
                    tasksTotal: `${totalSize}${totalUnits}`,
                    note: 'No note',
                })
            },
        })
        .then((res) => {
            if (res.status === 200) {
                progDispatch({
                    type: 'task_complete',
                    taskId: taskId,
                })
                return new Blob([res.data])
            } else {
                return Promise.reject(res.statusText)
            }
        })
        .then((blob) => {
            downloadBlob(blob, filename)
        })
}

export async function requestZipCreate(fileIds: string[], shareId: string) {
    const url = new URL(`${API_ENDPOINT}/takeout`)
    if (shareId) {
        url.searchParams.append('shareId', shareId)
    }

    return fetch(url.toString(), {
        method: 'POST',
        body: JSON.stringify({ fileIds: fileIds }),
    }).then(async (res) => {
        const json = await res.json()
        return { json: json, status: res.status }
    })
}

export async function AutocompleteAlbums(searchValue): Promise<AlbumData[]> {
    if (searchValue.length < 2) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/albums`)
    url.searchParams.append('filter', searchValue)
    return fetchJson(url.toString())
}

export async function NewWormhole(folderId: string) {
    const url = new URL(`${API_ENDPOINT}/share/files`)

    const body = {
        fileIds: [folderId],
        wormhole: true,
    }
    const res = await fetch(url.toString(), {
        method: 'POST',
        body: JSON.stringify(body),
    })
    return res
}

export async function GetWormholeInfo(shareId: string) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)
    return wrapRequest(fetch(url.toString()))
}

export async function searchFolder(
    folderId: string,
    searchString: string,
    filter: string
): Promise<WeblensFileParams[]> {
    const url = new URL(`${API_ENDPOINT}/folder/${folderId}/search`)
    url.searchParams.append('search', searchString)
    url.searchParams.append('filter', filter)

    const files: { files: WeblensFileParams[] } = await fetch(
        url.toString(),
        {}
    )
        .then((v) => v.json())
        .then((v) => {
            if (v.error) {
                return Promise.reject(v.error)
            }
            return v
        })
        .catch((r) => {
            console.error(`Failed to search files: ${r}`)
            return { files: [] }
        })
    return files.files
}

export async function getFilesystemStats(folderId: string): Promise<{sizesByExtension: {name: string, size: number}[]}> {
    return fetchJson(`${API_ENDPOINT}/files/${folderId}/stats`)
}

export async function getFileHistory(fileId: string): Promise<FileAction[]> {
    if (!fileId) {
        console.error('No fileId trying to get file history')
        return null
    }
    return fetchJson(`${API_ENDPOINT}/file/${fileId}/history`)
}

export async function getPastFolderInfo(folderId: string, timestamp: Date) {
    const millis = timestamp.getTime()
    const url = new URL(`${API_ENDPOINT}/file/rewind/${folderId}/${millis}`)

    return fetchJson(url.toString())
}

export async function restoreFiles(fileIds: string[], timestamp: Date) {
    const url = new URL(`${API_ENDPOINT}/history/restore`)
    return await fetch(url, {
        method: 'POST',
        body: JSON.stringify({
            fileIds: fileIds,
            timestamp: timestamp.getTime(),
        }),
    }).then((r) => {
        if (r.status !== 200) {
            return Promise.reject(r.statusText)
        } else {
            return
        }
    })
}

export async function GetFileText(fileId: string) {
    const url = new URL(`${API_ENDPOINT}/file/${fileId}/text`)
    return await fetch(url, {}).then((r) => r.text())
}
