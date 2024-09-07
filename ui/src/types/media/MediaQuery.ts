import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import WeblensMedia from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { AuthHeaderT, GalleryStateT } from 'types/Types'

export async function FetchData(
    galleryState: GalleryStateT,
    auth: AuthHeaderT
) {
    try {
        const url = new URL(`${API_ENDPOINT}/media`)
        url.searchParams.append(
            'raw',
            useMediaStore.getState().showRaw.toString()
        )
        url.searchParams.append(
            'hidden',
            useMediaStore.getState().showHidden.toString()
        )
        url.searchParams.append('limit', '100000')
        if (galleryState.albumsFilter) {
            url.searchParams.append(
                'albums',
                JSON.stringify(
                    Array.from(galleryState.albumsMap.values())
                        .filter((v) => galleryState.albumsFilter.includes(v.id))
                        ?.map((v) => v.id)
                )
            )
        }
        const data = await fetch(url.toString(), { headers: auth }).then(
            (res) => {
                if (res.status !== 200) {
                    return Promise.reject('Failed to get media')
                } else {
                    return res.json()
                }
            }
        )
        if (data.Media) {
            const medias = data.Media.map((m) => {
                return new WeblensMedia(m)
            })

            useMediaStore.getState().addMedias(medias)
        }
    } catch (error) {
        console.error(error)
    }
}

export async function getMedia(
    mediaId,
    authHeader: AuthHeaderT
): Promise<WeblensMedia> {
    if (!mediaId) {
        console.error('trying to get media with no mediaId')
        return
    }
    const url = new URL(`${API_ENDPOINT}/media/${mediaId}`)
    url.searchParams.append('meta', 'true')
    const mediaMeta: WeblensMedia = await fetch(url, {
        headers: authHeader,
    }).then((r) => r.json())
    return mediaMeta
}

export async function fetchMediaTypes() {
    const url = new URL(`${API_ENDPOINT}/media/types`)
    return await fetch(url).then((r) => {
        if (r.status === 200) {
            return r.json()
        } else {
            return Promise.reject(r.status)
        }
    })
}

export async function hideMedia(
    mediaIds: string[],
    hidden: boolean,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/media/hide`)
    url.searchParams.append('hidden', hidden.toString())
    return await fetch(url, {
        body: JSON.stringify({ mediaIds: mediaIds }),
        method: 'PATCH',
        headers: authHeader,
    })
}

export async function adjustMediaTime(
    mediaId: string,
    newDate: Date,
    extraMedias: string[],
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/media/date`)
    return await fetch(url, {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify({
            anchorId: mediaId,
            newTime: newDate,
            mediaIds: extraMedias,
        }),
    }).then((r) => {
        if (r.status !== 200) {
            return Promise.reject(r.statusText)
        }
        return r.statusText
    })
}

export async function likeMedia(
    mediaId: string,
    liked: boolean,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/media/${mediaId}/liked`)
    url.searchParams.append('liked', String(liked))

    return fetch(url, { headers: authHeader, method: 'POST' })
}

export async function getRandomThumbs() {
    const url = new URL(`${API_ENDPOINT}/media/random`)
    url.searchParams.append('count', '50')
    return await fetch(url).then((r) => r.json())
}
