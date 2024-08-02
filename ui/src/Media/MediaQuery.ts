import { AuthHeaderT, GalleryStateT } from '../types/Types'
import API_ENDPOINT from '../api/ApiEndpoint'
import WeblensMedia, { MediaAction, MediaStateT } from './Media'

export async function FetchData(
    galleryState: GalleryStateT,
    mediaState: MediaStateT,
    mediaDispatch: (a: MediaAction) => void,
    authHeader: AuthHeaderT
) {
    if (
        !authHeader ||
        authHeader.Authorization === ''
        // mediaState.albumsMap.size === 0
    ) {
        return
    }

    try {
        const url = new URL(`${API_ENDPOINT}/media`)
        url.searchParams.append('raw', mediaState.isShowingRaw().toString())
        url.searchParams.append(
            'hidden',
            mediaState.isShowingHidden().toString()
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
        const data = await fetch(url.toString(), { headers: authHeader }).then(
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

            mediaDispatch({ type: 'add_medias', medias: medias })
        }
    } catch (error) {
        console.error(error)
    }
}
