import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import { WeblensFile } from '@weblens/types/files/File'
import { ShareInfo, WeblensShare } from '@weblens/types/share/share'
import { useSessionStore } from 'components/UserInfo'

export async function deleteShare(shareId: string) {
    const auth = useSessionStore.getState().auth

    const url = new URL(`${API_ENDPOINT}/share/${shareId}`)
    const res = await fetch(url.toString(), {
        headers: auth,
        method: 'DELETE',
    })
    return res
}

export async function shareFile(
    file: WeblensFile,
    isPublic: boolean,
    users: string[] = []
): Promise<ShareInfo> {
    const auth = useSessionStore.getState().auth

    const url = new URL(`${API_ENDPOINT}/share/files`)
    const body = {
        fileId: file.Id(),
        users: users,
        public: isPublic,
    }
    return await fetch(url.toString(), {
        headers: auth,
        method: 'POST',
        body: JSON.stringify(body),
    }).then((res) => res.json())
}

export async function setFileSharePublic(shareId: string, isPublic: boolean) {
    const auth = useSessionStore.getState().auth

    const url = new URL(`${API_ENDPOINT}/share/${shareId}/public`)
    const body = {
        public: isPublic,
    }
    return await fetch(url.toString(), {
        headers: auth,
        method: 'PATCH',
        body: JSON.stringify(body),
    })
}

export async function setShareAccessors(
    shareId: string,
    addUsers: string[] = [],
    removeUsers: string[] = []
) {
    const auth = useSessionStore.getState().auth
    if (addUsers.length < 1 && removeUsers.length < 1) {
        return null
    }
    const url = new URL(`${API_ENDPOINT}/share/${shareId}/accessors`)
    const body = {
        addUsers: addUsers,
        removeUsers: removeUsers,
    }
    return await fetch(url.toString(), {
        headers: auth,
        method: 'PATCH',
        body: JSON.stringify(body),
    })
}

export async function getFileShare(shareId: string) {
    const auth = useSessionStore.getState().auth

    const url = new URL(`${API_ENDPOINT}/file/share/${shareId}`)
    return await fetch(url.toString(), { headers: auth })
        .then((res) => res.json())
        .then((j) => {
            return new WeblensShare(j)
        })
        .catch((r) => Promise.reject(r))
}
