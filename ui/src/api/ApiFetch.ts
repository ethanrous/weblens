import { useSessionStore } from '@weblens/components/UserInfo'
import { WeblensFileInfo } from '@weblens/types/files/File'
import { ApiKeyInfo, ServerInfoT, UserInfoT } from '@weblens/types/Types'
import API_ENDPOINT from './ApiEndpoint'

export async function wrapRequest(rq: Promise<Response>): Promise<Response> {
    return await rq
        .then((r: Response) => {
            if (r.status >= 400) {
                return Promise.reject(r.status)
            }
            return r
        })
        .catch((e) => {
            if (e === 401) {
                useSessionStore
                    .getState()
                    .setUserInfo({ isLoggedIn: false } as UserInfoT)
                useSessionStore.getState().nav('/login')
            }
            return Promise.reject(e)
        })
}

export async function fetchJson<T>(
    url: string,
    method?: string,
    body?: object
): Promise<T> {
    if (!method) {
        method = 'GET'
    }
    const init: RequestInit = {
        method: method,
    }

    if (body) {
        init.body = JSON.stringify(body)
    }

    return await wrapRequest(fetch(url, init)).then((r) => {
        return r.json()
    })
}

export async function login(
    user: string,
    pass: string
): Promise<{ token: string; user: UserInfoT }> {
    const url = new URL(`${API_ENDPOINT}/login`)
    const data = {
        username: user,
        password: pass,
    }

    return fetch(url.toString(), {
        method: 'POST',
        body: JSON.stringify(data),
    }).then((r) => r.json())
}

export async function createUser(username: string, password: string) {
    const url = new URL(`${API_ENDPOINT}/user`)
    return fetch(url, {
        method: 'POST',
        body: JSON.stringify({ username: username, password: password }),
    }).then((res) => {
        if (res.status !== 201) {
            return Promise.reject(`${res.statusText}`)
        }
    })
}

export function adminCreateUser(
    username: string,
    password: string,
    admin: boolean
) {
    const url = new URL(`${API_ENDPOINT}/user`)
    return wrapRequest(
        fetch(url, {
            method: 'POST',
            body: JSON.stringify({
                username: username,
                password: password,
                admin: admin,
                autoActivate: true,
            }),
        })
    )
}

export function clearCache() {
    return wrapRequest(
        fetch(`${API_ENDPOINT}/cache`, {
            method: 'POST',
        })
    )
}

export async function newApiKey() {
    const url = `${API_ENDPOINT}/key`
    return fetchJson(url, 'POST')
}

export async function deleteApiKey(key: string) {
    const url = new URL(`${API_ENDPOINT}/key/${key}`)
    return wrapRequest(
        fetch(url, {
            method: 'DELETE',
        })
    )
}

export async function getApiKeys(): Promise<ApiKeyInfo[]> {
    const url = new URL(`${API_ENDPOINT}/keys`)
    return fetchJson<ApiKeyInfo[]>(url.toString())
}

export async function initServer(
    serverName: string,
    role: 'core' | 'backup',
    username: string,
    password: string,
    coreAddress: string,
    coreKey: string
) {
    const url = new URL(`${API_ENDPOINT}/init`)
    const body = {
        name: serverName,
        role: role,
        username: username,
        password: password,
        coreAddress: coreAddress,
        coreKey: coreKey,
    }
    return await fetch(url, { body: JSON.stringify(body), method: 'POST' })
}

export async function attachNewCore(coreAddress: string, usingKey: string) {
    const url = new URL(`${API_ENDPOINT}/core/attach`)
    const body = {
        coreAddress: coreAddress,
        usingKey: usingKey,
    }
    return wrapRequest(
        fetch(url, { method: 'POST', body: JSON.stringify(body) })
    )
}

export async function getServerInfo() {
    return fetchJson<{
        info: ServerInfoT
        userCount: number
        started: boolean
    }>(`${API_ENDPOINT}/info`)
}

export async function getUsers(): Promise<UserInfoT[]> {
    const url = `${API_ENDPOINT}/users`
    return fetchJson(url)
}

export async function AutocompleteUsers(
    searchValue: string
): Promise<UserInfoT[]> {
    if (searchValue.length < 2) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/users/search`)
    url.searchParams.append('filter', searchValue)
    return (await fetchJson<{ users: UserInfoT[] }>(url.toString())).users
}

export async function doBackup(serverId: string) {
    const url = new URL(`${API_ENDPOINT}/backup`)
    url.searchParams.append('serverId', serverId)
    return wrapRequest(fetch(url, { method: 'POST' }))
}

export async function getRemotes(): Promise<ServerInfoT[]> {
    return fetchJson<ServerInfoT[]>(`${API_ENDPOINT}/remotes`)
}

export async function deleteRemote(remoteId: string) {
    const url = new URL(`${API_ENDPOINT}/remote`)
    return await wrapRequest(
        fetch(url, {
            method: 'DELETE',
            body: JSON.stringify({ remoteId: remoteId }),
        })
    )
}

export async function autocompletePath(pathQuery: string): Promise<{
    folder: WeblensFileInfo
    children: WeblensFileInfo[]
}> {
    if (!pathQuery) {
        return
    }
    const url = new URL(`${API_ENDPOINT}/files/autocomplete`)
    url.searchParams.append('searchPath', pathQuery)
    return fetchJson(url.toString())
}

export async function searchFilenames(
    searchString: string
): Promise<WeblensFileInfo[]> {
    if (searchString.length < 1) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/files/search`)
    url.searchParams.append('search', searchString)
    return fetchJson(url.toString())
}

export async function resetServer() {
    return wrapRequest(fetch(`${API_ENDPOINT}/reset`, { method: 'POST' }))
}
