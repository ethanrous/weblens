import { notifications } from '@mantine/notifications'
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

export function createUser(username, password) {
    const url = new URL(`${API_ENDPOINT}/user`)
    return fetch(url, { method: "POST", body: JSON.stringify({ username: username, password: password }) })
        .then((res) => { if (res.status != 201) { return Promise.reject(`${res.statusText}`) } })
}

export function adminCreateUser(username, password, admin, authHeader?) {
    const url = new URL(`${API_ENDPOINT}/admin/user`)
    return fetch(url, { headers: authHeader, method: "POST", body: JSON.stringify({ username: username, password: password, admin: admin, autoActivate: true }) })
        .then((res) => { if (res.status != 201) { return Promise.reject(`${res.statusText}`) } })
        .catch((reason) => { notifications.show({ message: `Failed to create new user: ${reason}` }) })
}

export function clearCache(authHeader) {
    return fetch(`${API_ENDPOINT}/admin/cache`, { method: "POST", headers: authHeader })
}