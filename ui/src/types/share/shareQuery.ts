import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import { fetchJson, wrapRequest } from '@weblens/api/ApiFetch'
import { WeblensFile } from '@weblens/types/files/File'
import { ShareInfo } from '@weblens/types/share/share'

export async function deleteShare(shareId: string) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)
    const res = await fetch(url.toString(), {
        method: 'DELETE',
    })
    return res
}

export async function shareFile(
    file: WeblensFile,
    isPublic: boolean,
    users: string[] = []
): Promise<ShareInfo> {
    return fetchJson(`${API_ENDPOINT}/share/files`, 'POST', {
        fileId: file.Id(),
        users: users,
        public: isPublic,
    })
}

export async function setFileSharePublic(shareId: string, isPublic: boolean) {
    const url = new URL(`${API_ENDPOINT}/share/${shareId}/public`)
    const body = {
        public: isPublic,
    }
    return await wrapRequest(
        fetch(url.toString(), {
            method: 'PATCH',
            body: JSON.stringify(body),
        })
    )
}

export async function setShareAccessors(
    shareId: string,
    addUsers: string[] = [],
    removeUsers: string[] = []
) {
    if (addUsers.length < 1 && removeUsers.length < 1) {
        return null
    }
    const url = new URL(`${API_ENDPOINT}/share/${shareId}/accessors`)
    const body = {
        addUsers: addUsers,
        removeUsers: removeUsers,
    }
    return await wrapRequest(
        fetch(url.toString(), {
            method: 'PATCH',
            body: JSON.stringify(body),
        })
    )
}

export async function getFileShare(shareId: string): Promise<ShareInfo> {
    return fetchJson(`${API_ENDPOINT}/share/${shareId}`)
}
