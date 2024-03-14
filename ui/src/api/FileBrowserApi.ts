import { notifications } from '@mantine/notifications'
import axios from 'axios'

import { AlbumData, FileBrowserDispatch, fileData, getBlankFile } from '../types/Types'
import { humanFileSize } from '../util'
import API_ENDPOINT from './ApiEndpoint'

export function SubToFolder(subId: string, recursive: boolean, wsSend) {
    if (!subId) {
        // console.error("Trying to subscribe to empty id")
        return
    }
    wsSend("subscribe", { subscribeType: "folder", subscribeKey: subId, subscribeMeta: JSON.stringify({ recursive: recursive }) })
}

export function UnsubFromFolder(subId: string, wsSend) {
    if (!subId) {
        // console.error("Trying to unsub to empty id")
        return
    }
    wsSend("unsubscribe", { subscribeKey: subId })
}

export function DeleteFiles(fileIds: string[], authHeader) {
    var url = new URL(`${API_ENDPOINT}/files`)

    fetch(url.toString(), { method: "DELETE", headers: authHeader, body: JSON.stringify(fileIds) })
    .catch(r => notifications.show({title: "Failed to delete file", message: String(r), color: 'red'}))
}

function getSharedWithMe(user, dispatch: FileBrowserDispatch, authHeader) {
    let url = new URL(`${API_ENDPOINT}/share`)
    return fetch(url.toString(), { headers: authHeader })
        .then((res) => res.json())
        .then((data) => {
            let files = data.files?.map((val: fileData) => { return { fileId: val.id, updateInfo: val } })
            if (!files) {
                files = []
            }

            const sharedFolder = getBlankFile()
            sharedFolder.isDir = true
            sharedFolder.id = "shared"
            sharedFolder.filename = "Shared"

            dispatch({ type: 'set_folder_info', fileInfo: sharedFolder })
            dispatch({ type: 'update_many', files: files, user: user })
            dispatch({ type: "set_loading", loading: false })
        })
}

function getMyTrash(user, dispatch: FileBrowserDispatch, authHeader) {
    let url = new URL(`${API_ENDPOINT}/trash`)
    return fetch(url.toString(), { headers: authHeader })
        .then((res) => res.json())
        .then((data) => {
            let children = data.children?.map((val: fileData) => { return { fileId: val.id, updateInfo: val } })
            if (!children) {
                children = []
            }
            let parents = data.parents.reverse()
            parents.shift()

            data.self.filename = "Trash"
            dispatch({ type: 'set_folder_info', fileInfo: data.self })
            dispatch({ type: 'update_many', files: children, user: user })
            dispatch({ type: 'set_parents_info', parents: parents })
        })
}

export async function GetFileInfo(fileId: string, shareId: string, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file/${fileId}`)
    if (shareId !== "") {
        url.searchParams.append("shareId", shareId)
    }
    return (await fetch(url.toString(), {headers: authHeader})).json()
}

export function GetFolderData(folderId, user, dispatch: FileBrowserDispatch, authHeader) {
    if (folderId === "shared") {
        return getSharedWithMe(user, dispatch, authHeader)
    }
    if (folderId === "trash") {
        return getMyTrash(user, dispatch, authHeader)
    }

    let url = new URL(`${API_ENDPOINT}/folder/${folderId}`)
    return fetch(url.toString(), { headers: authHeader })
        .then((res) => {
            if (res.status === 404) {
                return Promise.reject(404)
            } else if (res.status === 401) {
                return Promise.reject("Not Authorized")
            } else {
                try {
                    let j = res.json()
                    return j
                } catch {
                    return Promise.reject("Failed to decode response")
                }
            }
        })
}

export async function CreateFolder(parentFolderId, name, isPublic, shareId, authHeader): Promise<string> {
    if (isPublic && !shareId) {
        throw new Error("Attempting to do public upload with no shareId");
    }

    var url
    if (isPublic) {
        url = new URL(`${API_ENDPOINT}/public/folder`)
        url.searchParams.append('shareId', shareId)
    } else {
        url = new URL(`${API_ENDPOINT}/folder`)
    }
    url.searchParams.append('parentFolderId', parentFolderId)
    url.searchParams.append('folderName', name)

    const dirInfo = await fetch(url.toString(), { method: "POST", headers: authHeader }).then(res => res.json()).catch((r) => { notifications.show({ title: "Could not create folder", message: String(r), color: 'red' }) })
    return dirInfo?.folderId
}

export function MoveFile(currentParentId, newParentId, currentFilename, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('currentParentId', currentParentId)
    url.searchParams.append('newParentId', newParentId)
    url.searchParams.append('currentFilename', currentFilename)
    return fetch(url.toString(), { method: "PUT", headers: authHeader })
}

export function MoveFiles(fileIds: string[], newParentId: string, authHeader) {
    var url = new URL(`${API_ENDPOINT}/files`)
    const body = {
        fileIds: fileIds,
        newParentId: newParentId
    }

    return fetch(url.toString(), { method: "PATCH", headers: authHeader, body: JSON.stringify(body) })
}

export async function RenameFile(fileId: string, newName, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('fileId', fileId)
    url.searchParams.append('newFilename', newName)
    fetch(url.toString(), { method: "PATCH", headers: authHeader })
}

function downloadBlob(blob, filename) {
    const aElement = document.createElement("a")
    aElement.setAttribute("download", filename)
    const href = URL.createObjectURL(blob)
    aElement.href = href
    aElement.setAttribute("target", "_blank")
    aElement.click()
    URL.revokeObjectURL(href)
    return
}

export function downloadSingleFile(fileId: string, authHeader, dispatch: FileBrowserDispatch, filename: string, ext: string, shareId: string) {
    const url = new URL(`${API_ENDPOINT}/download`)
    url.searchParams.append("fileId", fileId)
    if (shareId) {
        url.searchParams.append("shareId", shareId)
    }

    const notifId = `download_${fileId}`
    notifications.show({ id: notifId, message: "Starting download", autoClose: false, loading: true })

    axios.get(url.toString(), {
        responseType: 'blob',
        headers: authHeader,
        onDownloadProgress: (p) => {
            dispatch({ type: "set_scan_progress", progress: Number((p.progress * 100).toFixed(0)) })
            const [speed, units] = humanFileSize(p.rate)
            notifications.update({ id: notifId, message: `Downloading ${(p.progress * 100).toFixed(0)}% (${speed}${units}/s)` })
        },
    })
        .then(res => new Blob([res.data]))
        .then((blob) => {
            downloadBlob(blob, filename ? filename : `${fileId}.${ext}`)
        })
        .finally(() => { dispatch({ type: "set_scan_progress", progress: 0 }); notifications.hide(notifId) })
}

export function requestZipCreate(fileIds: string[], shareId: string, authHeader) {
    const url = new URL(`${API_ENDPOINT}/takeout`)
    if (shareId !== "") {
        url.searchParams.append("shareId", shareId)
    }

    return fetch(url.toString(), { headers: authHeader, method: "POST", body: JSON.stringify({ fileIds: fileIds }) })
        .then(async (res) => {
            const json = await res.json()
            return { json: json, status: res.status }
        })
        .catch(r => {
            notifications.show({title: "Failed to request takeout", message: String(r), color: 'red'})
            return {json: null, status: 0}
        })
}

export async function AutocompleteUsers(searchValue, authHeader) {
    if (searchValue.length < 2) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/users`)
    url.searchParams.append('filter', searchValue)
    const res = await fetch(url.toString(), { headers: authHeader }).then(res => res.json())
    return res.users ? res.users : []
}

export async function AutocompleteAlbums(searchValue, authHeader): Promise<AlbumData[]> {
    if (searchValue.length < 2) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/albums`)
    url.searchParams.append('filter', searchValue)
    const res = await fetch(url.toString(), { headers: authHeader }).then(res => res.json())
    return res.albums ? res.albums : []
}

export async function NewWormhole(folderId: string, authHeader) {
    const url = new URL(`${API_ENDPOINT}/share/files`)

    const body = {
        fileIds: [folderId],
        wormhole: true
    }
    const res = await fetch(url.toString(), { headers: authHeader, method: "POST", body: JSON.stringify(body) })
    return res
}

export async function DeleteShare(shareId, authHeader) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)
    const res = await fetch(url.toString(), { headers: authHeader, method: "DELETE" })
    return res
}

export async function GetWormholeInfo(shareId: string, authHeader) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)

    return fetch(url.toString(), { headers: authHeader })
}

export async function GetMediasByFolder(folderId: string, authHeader) {
    const url = new URL(`${API_ENDPOINT}/folder/${folderId}/media`)
    return fetch(url.toString(), { headers: authHeader }).then(res => res.json())
}

export async function ShareFiles(files: string[], isPublic: boolean, users: string[] = [], authHeader) {
    const url = new URL(`${API_ENDPOINT}/share/files`)
    const body = {
        fileIds: files,
        users: users,
        public: isPublic
    }
    const res = await fetch(url.toString(), { headers: authHeader, method: "POST", body: JSON.stringify(body) }).then(res => res.json())
    return res
}

export async function UpdateFileShare(shareId: string, isPublic: boolean, users: string[] = [], authHeader) {
    const url = new URL(`${API_ENDPOINT}/file/share/${shareId}`)
    const body = {
        users: users,
        public: isPublic
    }
    const res = await fetch(url.toString(), { headers: authHeader, method: "PATCH", body: JSON.stringify(body) })
    return res
}

export async function GetFileShare(shareId, authHeader) {
    const url = new URL(`${API_ENDPOINT}/file/share/${shareId}`)
    return await fetch(url.toString(), { headers: authHeader }).then(res => res.json()).catch(r => Promise.reject(r))
}