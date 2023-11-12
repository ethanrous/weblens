import { itemData } from '../types/Types'
import API_ENDPOINT from './ApiEndpoint'

export function GetFileInfo(filepath, dispatch) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('path', filepath)
    fetch(url.toString()).then((res) => res.json()).then((data: itemData) => {
        dispatch({
            type: 'update_item', item: data
        })
    })
}

export function DeleteFile(path, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('path', path)
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

export function GetDirectoryData(path, dispatch, navigate, authHeader) {
    dispatch({ type: "set_loading", loading: true })
    var url = new URL(`${API_ENDPOINT}/directory`)
    path = ('/' + path).replace(/\/\/+/g, '/')
    url.searchParams.append('path', path)
    fetch(url.toString(), { headers: authHeader })
        .then((res) => {
            if (res.status == 404) {
                path = path.slice(0, -1)
                let newPath = `/files/${path.slice(0, path.lastIndexOf("/"))}`.replace(/\/\/+/g, '/')
                navigate(newPath, { replace: true })
            } else if (res.status == 401) {
                navigate("/login", { state: { doLogin: false } })
            } else {
                return res.json()
            }
        })
        .then((data: itemData) => {
        dispatch({
            type: 'update_item', items: data
        })
    })
}

export function CreateDirectory(path, authHeader) {
    var url = new URL(`${API_ENDPOINT}/directory`)
    url.searchParams.append('path', ('/' + path).replace(/\/\/+/g, '/'))

    return fetch(url.toString(), { method: "POST", headers: authHeader })
        .then(res => { return res.json() })
}

export function MoveFile(existingPath, newPath, authHeader) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('existingFilepath', existingPath)
    url.searchParams.append('newFilepath', newPath)
    return fetch(url.toString(), { method: "PUT", headers: authHeader })
}

export function RenameFile(existingPath, newName, authHeader) {
    const parentDir = existingPath.replace(/(.*?)[^/]*$/, '$1')
    let newPath = (parentDir + newName)

    MoveFile(existingPath, newPath, authHeader)

    return newPath
}