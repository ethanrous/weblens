import API_ENDPOINT from './ApiEndpoint'

export function GetFileInfo(filepath, dispatch) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('path', filepath)
    fetch(url.toString()).then((res) => res.json()).then((data) => {
        dispatch({
            type: 'update_items', items: data == null ? [] : [data]
        })
    })
}

export function DeleteFile(path) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('path', path)
    fetch(url.toString(), { method: "DELETE" })
}

export function GetDirectoryData(path, dispatch) {
    var url = new URL(`${API_ENDPOINT}/directory`)
    url.searchParams.append('path', ('/' + path).replace(/\/\/+/g, '/'))
    fetch(url.toString()).then((res) => res.json()).then((data) => {
        dispatch({
            type: 'update_items', items: data == null ? [] : data
        })
    })
}

export function CreateDirectory(path, dispatch) {
    console.log("HJERE")
    var url = new URL(`${API_ENDPOINT}/directory`)
    url.searchParams.append('path', ('/' + path).replace(/\/\/+/g, '/'))
    console.log("AFTR")
    return fetch(url.toString(), { method: "POST" }).then(() => GetDirectoryData(path, dispatch))
}

export function RenameFile(existingPath, newPath) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('existingFilepath', existingPath)
    url.searchParams.append('newFilepath', newPath)
    return fetch(url.toString(), { method: "PUT" })
}