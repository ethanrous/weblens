import { notifications } from '@mantine/notifications'
import API_ENDPOINT from './ApiEndpoint'

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