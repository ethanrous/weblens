import { notifications } from '@mantine/notifications'
import axios from 'axios'

import { AlbumData, AuthHeaderT, FBDispatchT } from '../types/Types'
import { FileInitT, WeblensFile } from '../classes/File'
import API_ENDPOINT from './ApiEndpoint'
import { WeblensShare } from '../classes/Share'
import { FbModeT } from '../Pages/FileBrowser/FileBrowser'

export function SubToFolder(
    subId: string,
    shareId: string,
    recursive: boolean,
    wsSend
) {
    if (!subId || subId === 'shared') {
        return
    }
    wsSend('subscribe', {
        subscribeType: 'folder',
        subscribeKey: subId,
        shareId: shareId,
        subscribeMeta: JSON.stringify({ recursive: recursive }),
    })
}

export function UnsubFromFolder(subId: string, wsSend) {
    if (!subId) {
        // console.error("Trying to unsub to empty id")
        return
    }
    wsSend('unsubscribe', { subscribeKey: subId })
}

export function TrashFiles(
    fileIds: string[],
    shareId: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/files/trash`)
    if (shareId) {
        url.searchParams.append('shareId', shareId)
    }

    fetch(url.toString(), {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify(fileIds),
    }).catch((r) => console.error(r))
}

export function DeleteFiles(fileIds: string[], authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/files`)

    fetch(url.toString(), {
        method: 'DELETE',
        headers: authHeader,
        body: JSON.stringify(fileIds),
    }).catch((r) => console.error(r))
}

export function UnTrashFiles(fileIds: string[], authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/files/untrash`)

    fetch(url.toString(), {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify(fileIds),
    }).catch((r) =>
        notifications.show({
            title: 'Failed to un-delete file',
            message: String(r),
            color: 'red',
        })
    )
}

async function getSharedWithMe(authHeader: AuthHeaderT) {
    let url = new URL(`${API_ENDPOINT}/files/shared`)
    return fetch(url.toString(), { headers: authHeader })
        .then((res) => res.json())
        .then((data) => {
            const sharedFolder = new WeblensFile({
                id: 'shared',
                isDir: true,
                filename: 'Shared',
            })
            return { children: data.files, self: sharedFolder }

            // dispatch({ type: "set_folder_info", fileInfo: sharedFolder });
            // dispatch({ type: "update_many", files: files, user: user });
        })
}

async function getExternalFiles(contentId: string, authHeader: AuthHeaderT) {
    let url: URL
    url = new URL(`${API_ENDPOINT}/files/external/${contentId}`)
    return fetch(url.toString(), { headers: authHeader })
        .then((res) => res.json())
        .then((data) => {
            const ret = {
                self: data.self,
                parents: data.parents,
                children: [],
            }
            if (data.children) {
                ret.children = data.children
            } else if (data.files) {
                ret.children = data.files
            }
            return ret
        })
}

export async function GetFileInfo(
    fileId: string,
    shareId: string,
    authHeader: AuthHeaderT
): Promise<FileInitT> {
    const url = new URL(`${API_ENDPOINT}/file/${fileId}`)
    if (shareId !== '') {
        url.searchParams.append('shareId', shareId)
    }
    return (await fetch(url.toString(), { headers: authHeader })).json()
}

export async function GetFolderData(
    contentId: string,
    fbMode: FbModeT,
    shareId: string,
    authHeader: AuthHeaderT
) {
    if (fbMode === FbModeT.share && !shareId) {
        return getSharedWithMe(authHeader)
    }
    if (fbMode === FbModeT.external) {
        return getExternalFiles(contentId, authHeader)
    }

    const url = new URL(`${API_ENDPOINT}/folder/${contentId}`)
    if (fbMode === FbModeT.share) {
        url.searchParams.append('shareId', shareId)
    }

    return fetch(url.toString(), { headers: authHeader }).then((res) => {
        if (res.status === 404) {
            return Promise.reject(404)
        } else if (res.status === 401) {
            return Promise.reject('Not Authorized')
        } else {
            try {
                let j = res.json()
                return j
            } catch {
                return Promise.reject('Failed to decode response')
            }
        }
    })
}

export async function CreateFolder(
    parentFolderId: string,
    name: string,
    children: string[],
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT
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
        headers: authHeader,
        body: JSON.stringify({
            parentFolderId: parentFolderId,
            newFolderName: name,
            children: children,
        }),
    })
        .then((res) => res.json())
        .catch((r) => {
            notifications.show({
                title: 'Could not create folder',
                message: String(r),
                color: 'red',
            })
        })
    return dirInfo?.folderId
}

export function moveFile(
    currentParentId,
    newParentId,
    currentFilename,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('currentParentId', currentParentId)
    url.searchParams.append('newParentId', newParentId)
    url.searchParams.append('currentFilename', currentFilename)
    return fetch(url.toString(), { method: 'PUT', headers: authHeader })
}

export function moveFiles(
    fileIds: string[],
    newParentId: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/files`)
    const body = {
        fileIds: fileIds,
        newParentId: newParentId,
    }

    return fetch(url.toString(), {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify(body),
    })
}

export async function RenameFile(
    fileId: string,
    newName: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/file/${fileId}`)
    fetch(url.toString(), {
        method: 'PATCH',
        body: JSON.stringify({ newName: newName }),
        headers: authHeader,
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
    authHeader,
    dispatch: FBDispatchT,
    filename: string,
    ext: string,
    shareId: string
) {
    const url = new URL(`${API_ENDPOINT}/download`)
    url.searchParams.append('fileId', fileId)
    if (shareId) {
        url.searchParams.append('shareId', shareId)
    }

    const notifId = `download_${fileId}`
    notifications.show({
        id: notifId,
        message: 'Starting download',
        autoClose: false,
        loading: true,
    })

    axios
        .get(url.toString(), {
            responseType: 'blob',
            headers: authHeader,
            // onDownloadProgress: (p) => {
            //     dispatch({
            //         type: "set_scan_progress",
            //         progress: Number((p.progress * 100).toFixed(0)),
            //     });
            //     const [speed, units] = humanFileSize(p.rate);
            //     notifications.update({
            //         id: notifId,
            //         message: `Downloading ${(p.progress * 100).toFixed(0)}% (${speed}${units}/s)`,
            //     });
            // },
        })
        .then((res) => new Blob([res.data]))
        .then((blob) => {
            downloadBlob(blob, filename ? filename : `${fileId}.${ext}`)
        })
        .finally(() => {
            // dispatch({ type: "set_scan_progress", progress: 0 });
            notifications.hide(notifId)
        })
}

export async function requestZipCreate(
    fileIds: string[],
    shareId: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/takeout`)
    if (shareId) {
        url.searchParams.append('shareId', shareId)
    }

    return fetch(url.toString(), {
        headers: authHeader,
        method: 'POST',
        body: JSON.stringify({ fileIds: fileIds }),
    })
        .then(async (res) => {
            const json = await res.json()
            return { json: json, status: res.status }
        })
        .catch((r) => {
            notifications.show({
                title: 'Failed to request takeout',
                message: String(r),
                color: 'red',
            })
            return { json: null, status: 0 }
        })
}

export async function AutocompleteAlbums(
    searchValue,
    authHeader: AuthHeaderT
): Promise<AlbumData[]> {
    if (searchValue.length < 2) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/albums`)
    url.searchParams.append('filter', searchValue)
    const res = await fetch(url.toString(), { headers: authHeader }).then(
        (res) => res.json()
    )
    return res.albums ? res.albums : []
}

export async function NewWormhole(folderId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/share/files`)

    const body = {
        fileIds: [folderId],
        wormhole: true,
    }
    const res = await fetch(url.toString(), {
        headers: authHeader,
        method: 'POST',
        body: JSON.stringify(body),
    })
    return res
}

export async function DeleteShare(shareId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)
    const res = await fetch(url.toString(), {
        headers: authHeader,
        method: 'DELETE',
    })
    return res
}

export async function GetWormholeInfo(
    shareId: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)

    return fetch(url.toString(), { headers: authHeader })
}

export async function GetMediasByFolder(
    folderId: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/folder/${folderId}/media`)
    return fetch(url.toString(), { headers: authHeader }).then((res) =>
        res.json()
    )
}

export async function shareFiles(
    files: WeblensFile[],
    isPublic: boolean,
    users: string[] = [],
    authHeader: AuthHeaderT
): Promise<any> {
    const url = new URL(`${API_ENDPOINT}/share/files`)
    const body = {
        fileIds: files.map((f) => f.Id()),
        users: users,
        public: isPublic,
    }
    return await fetch(url.toString(), {
        headers: authHeader,
        method: 'POST',
        body: JSON.stringify(body),
    })
        .then((res) => res.json())
        .then((j) => j.shareData)
}

export async function updateFileShare(
    shareId: string,
    isPublic: boolean,
    users: string[] = [],
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/file/share/${shareId}`)
    const body = {
        users: users,
        public: isPublic,
    }
    return await fetch(url.toString(), {
        headers: authHeader,
        method: 'PATCH',
        body: JSON.stringify(body),
    })
}

export async function getFileShare(shareId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/file/share/${shareId}`)
    return await fetch(url.toString(), { headers: authHeader })
        .then((res) => res.json())
        .then((j) => {
            return new WeblensShare(j)
        })
        .catch((r) => Promise.reject(r))
}

export async function searchFolder(
    folderId: string,
    searchString: string,
    filter: string,
    authHeader: AuthHeaderT
): Promise<FileInitT[]> {
    const url = new URL(`${API_ENDPOINT}/folder/${folderId}/search`)
    url.searchParams.append('search', searchString)
    url.searchParams.append('filter', filter)

    const files: { files: FileInitT[] } = await fetch(url.toString(), {
        headers: authHeader,
    })
        .then((v) => v.json())
        .then((v) => {
            if (v.error) {
                return Promise.reject(v.error)
            }
            return v
        })
        .catch((r) => {
            notifications.show({
                title: 'Failed to search files',
                message: String(r),
                color: 'red',
            })
            return { files: [] }
        })
    return files.files
}

export async function getFilesystemStats(
    folderId: string,
    authHeader: AuthHeaderT
) {
    return await fetch(`${API_ENDPOINT}/files/${folderId}/stats`, {
        headers: authHeader,
    }).then((d) => d.json())
}

export async function getFileHistory(fileId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/file/${fileId}/history`)
    return await fetch(url, { headers: authHeader }).then((r) => {
        if (r.status !== 200) {
            return r.status
        } else {
            return r.json()
        }
    })
}

export async function getSnapshots(authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/snapshots`)
    return await fetch(url, { headers: authHeader }).then((r) => {
        if (r.status !== 200) {
            return r.status
        } else {
            return r.json()
        }
    })
}

export async function getPastFolderInfo(
    folderId: string,
    timestamp: Date,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/history/${folderId}`)
    url.searchParams.append('before', String(timestamp.getTime()))
    return await fetch(url, { headers: authHeader }).then((r) => {
        if (r.status !== 200) {
            return r.status
        } else {
            return r.json()
        }
    })
}

export async function restoreFiles(
    fileIds: string[],
    timestamp: Date,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/history/restore`)
    return await fetch(url, {
        headers: authHeader,
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
