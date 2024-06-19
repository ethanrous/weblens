import { AuthHeaderT, GalleryDispatchT, GalleryStateT } from '../types/Types'
import API_ENDPOINT from '../api/ApiEndpoint'
import WeblensMedia from './Media'

export async function FetchData(
    mediaState: GalleryStateT,
    dispatch: GalleryDispatchT,
    authHeader: AuthHeaderT
) {
    if (
        !authHeader ||
        authHeader.Authorization === ''
        // mediaState.albumsMap.size === 0
    ) {
        console.log('HERE', authHeader, mediaState)
        return
    }

    try {
        const url = new URL(`${API_ENDPOINT}/media`)
        url.searchParams.append('raw', mediaState.includeRaw.toString())
        if (mediaState.albumsFilter) {
            url.searchParams.append(
                'albums',
                JSON.stringify(
                    Array.from(mediaState.albumsMap.values())
                        .filter((v) => mediaState.albumsFilter.includes(v.Id))
                        .map((v) => v.Id)
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

        const medias = data.Media.map((m) => {
            return new WeblensMedia(m)
        })

        dispatch({
            type: 'set_media',
            medias: medias,
        })
    } catch (error) {
        console.error(
            'Error fetching data - Ethan you wrote this, its not a js err'
        )
    }
}
