import API_ENDPOINT from './ApiEndpoint'

export const fetchMetadata = (fileHash, setMediaData) => {
    var url = new URL(`${API_ENDPOINT}/item/${fileHash}`)
    url.searchParams.append('meta', 'true')

    fetch(url.toString()).then((res) => res.json()).then((data) => setMediaData(data))
}

export function fetchThumb64(fileHash, setMediaData: (thumb64: string) => void) {
    var url = new URL(`${API_ENDPOINT}/item/${fileHash}`)
    url.searchParams.append('thumbnail', 'true')

    fetch(url)
        .then(response => response.blob())
        .then(blob => new Promise((resolve, reject) => {
            const reader = new FileReader()
            reader.onloadend = () => resolve(reader.result)
            reader.onerror = reject
            reader.readAsDataURL(blob)
        }))
        .then(setMediaData)
}

export function fetchFullres64(fileHash, setMediaData: (fullres64: string) => void) {
    var url = new URL(`${API_ENDPOINT}/item/${fileHash}`)
    url.searchParams.append('fullres', 'true')

    fetch(url)
        .then(response => response.blob())
        .then(blob => new Promise((resolve, reject) => {
            const reader = new FileReader()
            reader.onloadend = () => resolve(reader.result)
            reader.onerror = reject
            reader.readAsDataURL(blob)
        }))
        .then(setMediaData)
}

export function login(user: string, pass: string) {
    var url = new URL(`${API_ENDPOINT}/login`)
    let data = {
        username: user,
        password: pass
    }

    return fetch(url.toString(), { method: "POST", body: JSON.stringify(data) })
}

export function createUser(username, password, admin, authHeader, enqueueSnackbar) {
    const url = new URL(`${API_ENDPOINT}/user`)
    fetch(url, { headers: authHeader, method: "POST", body: JSON.stringify({ username: username, password: password, admin: admin, autoActivate: false }) })
        .then((res) => { if (res.status != 201) { return Promise.reject(`${res.statusText}`) } })
        .catch((reason) => { enqueueSnackbar(`Failed to create new user: ${reason}`, { variant: "error" }) })
}