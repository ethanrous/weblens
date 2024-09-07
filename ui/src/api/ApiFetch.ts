import { notifications } from '@mantine/notifications'
import { WeblensFileInfo } from '@weblens/types/files/File'
import { ApiKeyInfo, AuthHeaderT, UserInfoT } from '@weblens/types/Types'
import API_ENDPOINT from './ApiEndpoint'

export function login(
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

export function createUser(username: string, password: string) {
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
    username,
    password,
    admin,
    authHeader?: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/user`)
    return fetch(url, {
        headers: authHeader,
        method: 'POST',
        body: JSON.stringify({
            username: username,
            password: password,
            admin: admin,
            autoActivate: true,
        }),
    })
        .then((res) => {
            if (res.status !== 201) {
                return Promise.reject(`${res.statusText}`)
            }
        })
        .catch((r) => {
            notifications.show({
                message: `Failed to create new user: ${String(r)}`,
            })
        })
}

export function clearCache(authHeader: AuthHeaderT) {
    return fetch(`${API_ENDPOINT}/cache`, {
        method: 'POST',
        headers: authHeader,
    })
}

export async function newApiKey(authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/key`)
    return await fetch(url, { headers: authHeader, method: 'POST' }).then((r) =>
        r.json()
    )
}

export async function deleteApiKey(key: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/key/${key}`)
    return await fetch(url, {
        headers: authHeader,
        method: 'DELETE',
    })
}

export async function getApiKeys(
    authHeader: AuthHeaderT
): Promise<ApiKeyInfo[]> {
    const url = new URL(`${API_ENDPOINT}/keys`)
    return (await fetch(url, { headers: authHeader }).then((r) => r.json()))
        .keys
}

export async function initServer(
    serverName: string,
    role: 'core' | 'backup',
    username: string,
    password: string,
    coreAddress: string,
    coreKey: string
) {
    const url = new URL(`${API_ENDPOINT}/initialize`)
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

export async function getServerInfo() {
    const url = new URL(`${API_ENDPOINT}/info`)
    return await fetch(url).then((r) => {
        if (r.status === 200) {
            return r.json()
        } else if (r.status === 307) {
            return 307
        } else {
            return Promise.reject(r.statusText)
        }
    })
}

export async function getUsers(authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/users`)
    let res
    if (authHeader && authHeader.Authorization !== '') {
        res = fetch(url, { headers: authHeader })
    } else {
        res = fetch(url)
    }
    return await res.then((r) => {
        if (r.status === 200) {
            return r.json()
        } else {
            return Promise.reject(r.statusText)
        }
    })
}

export async function AutocompleteUsers(
    searchValue: string,
    authHeader: AuthHeaderT
): Promise<UserInfoT[]> {
    if (searchValue.length < 2) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/users/search`)
    url.searchParams.append('filter', searchValue)
    const res = await fetch(url.toString(), { headers: authHeader }).then(
        (res) => res.json()
    )
    return res.users ? res.users : []
}

export async function doBackup(authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/backup`)
    return await fetch(url, { method: 'POST', headers: authHeader }).then(
        (r) => r.status
    )
}

export async function getRemotes(authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/remotes`)
    return await fetch(url, { method: 'GET', headers: authHeader }).then(
        (r) => {
            if (r.status !== 200) {
                return r.status
            } else {
                return r.json()
            }
        }
    )
}

export async function deleteRemote(remoteId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/remote`)
    return await fetch(url, {
        method: 'DELETE',
        headers: authHeader,
        body: JSON.stringify({ remoteId: remoteId }),
    })
}

export async function getAlbumPreview(
    albumId: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}/preview`)
    return await fetch(url, {
        method: 'GET',
        headers: authHeader,
    }).then((r) => {
        return r.json()
    })
}

export async function autocompletePath(
    pathQuery: string,
    authHeader: AuthHeaderT
): Promise<{ folder: WeblensFileInfo; children: WeblensFileInfo[] }> {
    if (!pathQuery) {
        return
    }
    const url = new URL(`${API_ENDPOINT}/files/autocomplete`)
    url.searchParams.append('searchPath', pathQuery)
    return await fetch(url, { headers: authHeader }).then((r) => r.json())
}

export async function getFileDataByPath(
    pathQuery: string,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/file/path`)
    url.searchParams.append('searchPath', pathQuery)
    return await fetch(url, { headers: authHeader }).then((r) => r.json())
}

export async function searchFilenames(
    searchString: string,
    authHeader: AuthHeaderT
): Promise<WeblensFileInfo[]> {
    if (searchString.length < 1) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/files/search`)
    url.searchParams.append('search', searchString)
    return fetch(url, { headers: authHeader }).then((r) => r.json())
}
