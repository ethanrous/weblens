import { useSessionStore } from '@weblens/components/UserInfo'
import API_ENDPOINT from './ApiEndpoint'
import User from '@weblens/types/user/User'

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
                const user = new User()
                useSessionStore.getState().setUser(user)

                if (window.location.pathname.startsWith('/files/share')) {
                    console.log('Got 401 on share page')
                    return
                }

                console.debug('Got 401, going to login')
                useSessionStore.getState().nav('/login', {
                    state: { returnTo: window.location.pathname },
                })
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

export function clearCache() {
    return wrapRequest(
        fetch(`${API_ENDPOINT}/cache`, {
            method: 'POST',
        })
    )
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

export async function doBackup(serverId: string) {
    const url = new URL(`${API_ENDPOINT}/backup`)
    url.searchParams.append('serverId', serverId)
    return wrapRequest(fetch(url, { method: 'POST' }))
}

export async function resetServer() {
    return wrapRequest(fetch(`${API_ENDPOINT}/reset`, { method: 'POST' }))
}
