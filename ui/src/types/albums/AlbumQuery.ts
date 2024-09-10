import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import { fetchJson } from '@weblens/api/ApiFetch'
import { MediaDataT } from '@weblens/types/media/Media'
import { AlbumData } from 'types/Types'

export async function getAlbums(includeShared: boolean): Promise<AlbumData[]> {
    const url = new URL(`${API_ENDPOINT}/albums`)
    url.searchParams.append('includeShared', includeShared.toString())

    return await fetch(url)
        .then((r) => r.json())
        .then((j) => j.albums)
        .catch((r) => console.error(r))
}

export async function createAlbum(albumName: string) {
    if (albumName === '') {
        return Promise.reject('No album title')
    }

    const url = new URL(`${API_ENDPOINT}/album`)
    const body = {
        name: albumName,
    }
    await fetch(url.toString(), {
        method: 'POST',
        body: JSON.stringify(body),
    })
}

export async function getAlbumMedia(
    albumId: string,
    includeRaw: boolean
): Promise<{ albumMeta: AlbumData; mediaInfos: MediaDataT[] }> {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    url.searchParams.append('raw', includeRaw.toString())
    const res = await fetch(url.toString())
    if (res.status === 404) {
        return Promise.reject(404)
    } else if (res.status !== 200) {
        return Promise.reject('Unknown error')
    }

    const data = await res.json()
    return { albumMeta: data.albumMeta, mediaInfos: data.medias }
}

export async function addMediaToAlbum(
    albumId: string,
    mediaIds: string[],
    folderIds: string[]
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    const body = {
        newMedia: mediaIds,
        newFolders: folderIds,
    }
    return (
        await fetch(url.toString(), {
            method: 'PATCH',

            body: JSON.stringify(body),
        })
    ).json()
}

export async function RemoveMediaFromAlbum(
    albumId: string,
    mediaIds: string[]
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    const body = {
        removeMedia: mediaIds,
    }
    await fetch(url.toString(), {
        method: 'PATCH',

        body: JSON.stringify(body),
    })
}

export async function ShareAlbum(
    albumId: string,
    users?: string[],
    removeUsers?: string[]
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    const body = {
        users: users,
        removeUsers: removeUsers,
    }
    return fetch(url.toString(), {
        method: 'PATCH',

        body: JSON.stringify(body),
    })
}

export async function LeaveAlbum(albumId: string) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}/leave`)
    return fetch(url, { method: 'POST' })
}

export async function SetAlbumCover(albumId: string, coverMediaId: string) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    const body = {
        cover: coverMediaId,
    }
    return fetch(url.toString(), {
        method: 'PATCH',

        body: JSON.stringify(body),
    })
}

export async function RenameAlbum(albumId: string, newName: string) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    const body = {
        newName: newName,
    }

    return fetchJson(url.toString(), 'PATCH', body)
}

export async function DeleteAlbum(albumId: string) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)

    return await fetch(url.toString(), {
        method: 'DELETE',
    })
}
