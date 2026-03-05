import { API_ENDPOINT, useWeblensAPI } from './AllApi'

export interface TagInfo {
    id: string
    name: string
    color: string
    owner: string
    fileIDs: string[]
    created: string
    updated: string
}

async function tagFetch<T>(path: string, options?: RequestInit): Promise<T> {
    // Ensure API_ENDPOINT is initialized (same as the generated client path)
    useWeblensAPI()

    const res = await fetch(`${API_ENDPOINT.value}/tags${path}`, {
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        ...options,
    })

    if (!res.ok) {
        const text = await res.text().catch(() => res.statusText)
        throw new Error(`Tag API error (${res.status}): ${text}`)
    }

    if (res.status === 204 || res.headers.get('content-length') === '0') {
        return null as T
    }

    return res.json()
}

export function fetchUserTags(): Promise<TagInfo[]> {
    return tagFetch<TagInfo[]>('')
}

export function createTag(name: string, color: string): Promise<TagInfo> {
    return tagFetch<TagInfo>('', {
        method: 'POST',
        body: JSON.stringify({ name, color }),
    })
}

export function updateTag(tagID: string, name?: string, color?: string): Promise<null> {
    return tagFetch<null>(`/${tagID}`, {
        method: 'PATCH',
        body: JSON.stringify({ name, color }),
    })
}

export function deleteTag(tagID: string): Promise<null> {
    return tagFetch<null>(`/${tagID}`, { method: 'DELETE' })
}

export function addFilesToTag(tagID: string, fileIDs: string[]): Promise<null> {
    return tagFetch<null>(`/${tagID}/files`, {
        method: 'POST',
        body: JSON.stringify({ fileIDs }),
    })
}

export function removeFilesFromTag(tagID: string, fileIDs: string[]): Promise<null> {
    return tagFetch<null>(`/${tagID}/files`, {
        method: 'DELETE',
        body: JSON.stringify({ fileIDs }),
    })
}
