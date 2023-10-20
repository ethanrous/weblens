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

export function GetDirectoryData(path, dispatch, navigate) {
    dispatch({ type: "set_loading", loading: true })
    var url = new URL(`${API_ENDPOINT}/directory`)
    path = ('/' + path).replace(/\/\/+/g, '/')
    url.searchParams.append('path', path)
    fetch(url.toString())
        .then((res) => {
            if (res.status == 404) {
                path = path.slice(0, -1)
                let newPath = `/files/${path.slice(0, path.lastIndexOf("/"))}`.replace(/\/\/+/g, '/')
                console.log("NEW!", newPath)
                navigate(newPath)
            } else {
                return res.json()
            }
        })
        .then((data) => {
            dispatch({ type: 'set_loading', loading: false })
        dispatch({
            type: 'update_items', items: data == null ? [] : data
        })
    })
}

export function CreateDirectory(path) {
    var url = new URL(`${API_ENDPOINT}/directory`)
    url.searchParams.append('path', ('/' + path).replace(/\/\/+/g, '/'))

    return fetch(url.toString(), { method: "POST" })
        .then(res => { return res.json() })
}

export function RenameFile(existingPath, newPath) {
    var url = new URL(`${API_ENDPOINT}/file`)
    url.searchParams.append('existingFilepath', existingPath)
    url.searchParams.append('newFilepath', newPath)
    return fetch(url.toString(), { method: "PUT" })
}