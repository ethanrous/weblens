import axios from 'axios'
import { itemData } from '../types/Types'
import API_ENDPOINT from './ApiEndpoint'
import { notifications } from '@mantine/notifications'

export function DeleteFile(parentId, filename, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('parentFolderId', parentId)
    url.searchParams.append('filename', filename)
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
    fetch(url.toString(), { headers: authHeader })
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

export function GetFolderData(folderId, username, dispatch, navigate, authHeader) {
    if (!folderId) {
        navigate("/files/home")
        return
    }
    dispatch({ type: "set_loading", loading: true })

    if (folderId == "shared") {
        getSharedWithMe(username, dispatch, authHeader)
        return
    }

    let url = new URL(`${API_ENDPOINT}/folder/${folderId}`)
    fetch(url.toString(), { headers: authHeader })
        .then((res) => {
            if (res.status === 404) {
                navigate("/fourohfour")
                return Promise.reject("Page not found")
            } else if (res.status === 401) {
                // navigate("/login", { state: { doLogin: false } })
            } else {
                return res.json()
            }
        })
        .then((data) => {
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
            dispatch({ type: "set_loading", loading: false })
        }).catch((r) => console.error(r))
}

export async function CreateFolder(parentFolderId, name, authHeader) {
    var url = new URL(`${API_ENDPOINT}/folder`)
    url.searchParams.append('parentFolderId', parentFolderId)
    url.searchParams.append('folderName', name)

    const dirInfo = await fetch(url.toString(), { method: "POST", headers: authHeader }).then(res => res.json())
    return { folderId: dirInfo.folderId, alreadyExisted: dirInfo.alreadyExisted }
}

export function MoveFile(currentParentId, newParentId, currentFilename, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('currentParentId', currentParentId)
    url.searchParams.append('newParentId', newParentId)
    url.searchParams.append('currentFilename', currentFilename)
    return fetch(url.toString(), { method: "PUT", headers: authHeader })
}

export async function RenameFile(parentId, oldName, newName, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('currentParentId', parentId)
    url.searchParams.append('currentFilename', oldName)
    url.searchParams.append('newFilename', newName)
    const res = await fetch(url.toString(), { method: "PUT", headers: authHeader }).then(res => res.json())
    return res.newItemId
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

export function downloadSingleItem(item: itemData, authHeader, dispatch) {
    const url = new URL(`${API_ENDPOINT}/download`)
    url.searchParams.append("parentFolderId", item.parentFolderId)
    url.searchParams.append("filename", item.filename)

    notifications.show({ id: `download_${item.filename}`, message: "Starting download", autoClose: false, loading: true })

    axios.get(url.toString(), {
        responseType: 'blob',
        headers: authHeader,
        onDownloadProgress: (p) => {
            notifications.update({ id: `download_${item.filename}`, message: `Downloading: (${(p.progress * 100).toFixed(0)}%)` })
            dispatch({ type: "set_scan_progress", progress: (p.progress * 100).toFixed(0) })
        },
    })
        .then(res => new Blob([res.data]))
        .then((blob) => {
            downloadBlob(blob, item.filename)
        })
        .finally(() => { dispatch({ type: "set_scan_progress", progress: 0 }); notifications.hide(`download_${item.filename}`) })
}

export function requestZipCreate(body, authHeader) {
    const url = new URL(`${API_ENDPOINT}/takeout`)

    return fetch(url.toString(), { headers: authHeader, method: "POST", body: JSON.stringify(body) })
        .then(async (res) => {
            if (res.status !== 200 && res.status !== 202) {
                Promise.reject(res.statusText)
            }
            const json = await res.json()
            return { json: json, status: res.status }
        })
}

export function downloadTakeout(takeoutId: string, authHeader, dispatch) {
    const url = new URL(`${API_ENDPOINT}/takeout/${takeoutId}`)
    let filename = "takeout.zip"

    notifications.show({ id: `zip_download_${takeoutId}`, message: "Starting download", autoClose: false, loading: true })

    return axios.get(url.toString(), {
        responseType: 'blob',
        headers: authHeader,
        onDownloadProgress: (p) => {
            notifications.update({ id: `zip_download_${takeoutId}`, message: `Downloading: (${(p.progress * 100).toFixed(0)}%)` })
            dispatch({ type: "set_scan_progress", progress: (p.progress * 100).toFixed(0) })
        },
    })
        .then((res) => {
            notifications.hide(`zip_download_${takeoutId}`)
            if (res.status !== 200) {
                return Promise.reject("Got bad response code while trying to download item")
            }
            filename = `${takeoutId}.zip`

            return new Blob([res.data])
        })
        .then((blob) => downloadBlob(blob, filename))
        .catch((r) => console.error(r))
        .finally(() => { dispatch({ type: "set_scan_progress", progress: 0 }); notifications.hide(`zip_download_${takeoutId}`) })
}

export async function AutocompleteUsers(searchValue, authHeader) {
    if (searchValue.length < 2) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/userSearch`)
    url.searchParams.append('searchValue', searchValue)
    const res = await fetch(url.toString(), { headers: authHeader }).then(res => res.json())
    return res.users ? res.users : []
}

export function ShareFiles(files: { parentFolderId: string, filename: string }[], users: string[], authHeader) {
    console.log(files)
    const url = new URL(`${API_ENDPOINT}/share`)
    const body = {
        files: files,
        users: users
    }
    fetch(url.toString(), { headers: authHeader, method: "POST", body: JSON.stringify(body) })
}