import axios from 'axios'
import { AlbumData, itemData } from '../types/Types'
import API_ENDPOINT from './ApiEndpoint'
import { notifications } from '@mantine/notifications'

export function DeleteFile(fileId, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('fileId', fileId)
    fetch(url.toString(), { method: "DELETE", headers: authHeader })
}

export function ChangeOwner(updateHashes: string[], user: string, authHeader) {
    const updateData = {
        "owner": user,
        "fileHashes": updateHashes
    }
    var url = new URL(`${API_ENDPOINT}/items`)
    fetch(url.toString(), { method: "PUT", headers: authHeader, body: JSON.stringify(updateData) })
}

function getSharedWithMe(username, dispatch, authHeader) {
    let url = new URL(`${API_ENDPOINT}/share`)
    return fetch(url.toString(), { headers: authHeader })
        .then((res) => res.json())
        .then((data) => {
            let files = data.files?.map((val: itemData) => { return { itemId: val.id, updateInfo: val } })
            if (!files) {
                files = []
            }
            dispatch({ type: 'set_folder_info', folderInfo: { id: "shared", filename: "Shared" } })
            dispatch({ type: 'update_many', items: files, user: username })
            dispatch({ type: "set_loading", loading: false })
        })
}

export async function GetFileInfo(fileId, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file/${fileId}`)
    return (await fetch(url.toString(), {headers: authHeader})).json()
}

export function GetFolderData(folderId, username, dispatch, authHeader) {

    if (folderId === "shared") {
        return getSharedWithMe(username, dispatch, authHeader)
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
        .then((data) => {
            if (data.error) {
                return Promise.reject(data.error)
            }
            let children = data.children?.map((val: itemData) => { return { itemId: val.id, updateInfo: val } })
            if (!children) {
                children = []
            }
            let parents
            if (!data.parents) {
                parents = []
            } else {
                parents = data.parents.reverse()
            }
            dispatch({ type: 'set_folder_info', folderInfo: data.self })
            dispatch({ type: 'update_many', items: children, user: username })
            dispatch({ type: 'set_parents_info', parents: parents })
        })
}

export async function CreateFolder(parentFolderId, name, authHeader): Promise<string> {
    var url = new URL(`${API_ENDPOINT}/folder`)
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

export function downloadSingleItem(fileId: string, authHeader, dispatch, filename?: string, ext?: string) {
    const url = new URL(`${API_ENDPOINT}/download`)
    url.searchParams.append("fileId", fileId)

    const notifId = `download_${fileId}`
    notifications.show({ id: notifId, message: "Starting download", autoClose: false, loading: true })

    axios.get(url.toString(), {
        responseType: 'blob',
        headers: authHeader,
        onDownloadProgress: (p) => {
            notifications.update({ id: notifId, message: `Downloading: (${(p.progress * 100).toFixed(0)}%)` })
            dispatch({ type: "set_scan_progress", progress: (p.progress * 100).toFixed(0) })
        },
    })
        .then(res => new Blob([res.data]))
        .then((blob) => {
            downloadBlob(blob, filename ? filename : `${fileId}.${ext}`)
        })
        .finally(() => { dispatch({ type: "set_scan_progress", progress: 0 }); notifications.hide(notifId) })
}

export function requestZipCreate(items, authHeader) {
    const url = new URL(`${API_ENDPOINT}/takeout`)

    return fetch(url.toString(), { headers: authHeader, method: "POST", body: JSON.stringify({ fileIds: items }) })
        .then(async (res) => {
            if (res.status !== 200 && res.status !== 202) {
                Promise.reject(res.statusText)
            }
            const json = await res.json()
            return { json: json, status: res.status }
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

export async function ShareFiles(files: string[], users: string[], authHeader) {
    const url = new URL(`${API_ENDPOINT}/files/share`)
    console.log(files, users)
    const body = {
        files: files,
        users: users
    }
    const res = await fetch(url.toString(), { headers: authHeader, method: "PATCH", body: JSON.stringify(body) })
    return res
}