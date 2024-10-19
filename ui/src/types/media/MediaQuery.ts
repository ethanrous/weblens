import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import { fetchJson, wrapRequest } from '@weblens/api/ApiFetch'
import WeblensMedia, { MediaDataT } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { GalleryStateT } from 'types/Types'

export async function FetchData(galleryState: GalleryStateT) {
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
        const data = await fetchJson<{ Media }>(url.toString(), 'GET')
        // const data = await wrapRequest<{ Media: MediaDataT[] }>(
        //     fetch(url.toString()).then((res) => {
        //         if (res.status !== 200) {
        //             return Promise.reject('Failed to get media')
        //         } else {
        //             return res.json()
        //         }
        //     })
        // )
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

export async function getMedias(mediaIds: string[]): Promise<MediaDataT[]> {
    const url = new URL(`${API_ENDPOINT}/medias`)
    const body = {
        mediaIds: mediaIds,
    }
    const medias = (
        await fetchJson<{ medias: MediaDataT[] }>(url.toString(), 'POST', body)
    ).medias
    return medias ? medias : []
}

export async function fetchMediaTypes() {
    const url = new URL(`${API_ENDPOINT}/media/types`)
    return await wrapRequest(
        fetch(url).then((r) => {
            if (r.status === 200) {
                return r.json()
            } else {
                return Promise.reject(r.status)
            }
        })
    )
}

export async function hideMedia(mediaIds: string[], hidden: boolean) {
    const url = new URL(`${API_ENDPOINT}/media/visibility`)
    url.searchParams.append('hidden', hidden.toString())

    return wrapRequest(
        fetch(url, {
            body: JSON.stringify({ mediaIds: mediaIds }),
            method: 'PATCH',
        })
    )
}

export async function adjustMediaTime(
    mediaId: string,
    newDate: Date,
    extraMedias: string[]
) {
    const url = new URL(`${API_ENDPOINT}/media/date`)
    return await fetch(url, {
        method: 'PATCH',
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

export async function likeMedia(mediaId: string, liked: boolean) {
    const url = new URL(`${API_ENDPOINT}/media/${mediaId}/liked`)
    url.searchParams.append('liked', String(liked))

    return wrapRequest(fetch(url, { method: 'POST' }))
}

export async function getRandomThumbs() {
    const url = new URL(`${API_ENDPOINT}/media/random`)
    url.searchParams.append('count', '50')
    return fetchJson(url.toString())
}
