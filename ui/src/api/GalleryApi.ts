import { redirect } from 'react-router-dom'
import { MediaData } from '../types/Types'
import API_ENDPOINT from './ApiEndpoint'

export async function fetchData(mediaState, dispatch, nav, username, token) {
    dispatch({ type: "set_loading", loading: true })

    try {
        // let mediaMap = new Map<string, MediaData>(mediaState.mediaMap)
        let mediaMap: Map<string, MediaData> = mediaState.mediaMap
        let previousLast: string = mediaState.previousLast

        const limit = mediaState.maxMediaCount - mediaState.mediaCount
        if (limit < 1) {
            if (limit < 0) {
                console.error(`maxMediaCount (${mediaState.maxMediaCount}) less than mediaCount ${mediaState.mediaCount}`)
            }
            dispatch({ type: "set_loading", loading: false })
            return
        }

        const url = new URL(`${API_ENDPOINT}/media`)
        url.searchParams.append('limit', limit.toString())
        url.searchParams.append('skip', mediaState.mediaCount.toString())
        url.searchParams.append('raw', mediaState.includeRaw.toString())
        const data = await fetch(url.toString(), { headers: { "Authorization": `${username},${token}` } }).then(res => { if (res.status == 401) { nav('/login') } else { return res.json() } })
        const media: MediaData[] = data.Media

        let hasMoreMedia: boolean
        if (data.Media != null) {
            hasMoreMedia = data.MoreMedia
            for (const [_, value] of media.entries()) {
                mediaMap.set(value.FileHash, value)
                if (previousLast) {
                    mediaMap.get(value.FileHash).Previous = mediaMap.get(previousLast)
                    mediaMap.get(previousLast).Next = mediaMap.get(value.FileHash)
                }
                previousLast = value.FileHash
            }
        } else {
            hasMoreMedia = false
        }

        dispatch({
            type: 'add_media',
            mediaMap: mediaMap,
            hasMoreMedia: hasMoreMedia,
            previousLast: previousLast,
            addedCount: media?.length || 0
        })
        dispatch({ type: "set_loading", loading: false })

    } catch (error) {
        // if (error.name == "TypeError") {
        //     throw new Error("Unauthorized")
        // }
        console.error(error.name)
        dispatch({ type: "set_loading", loading: false })
        // throw new Error("Error fetching data - Ethan you wrote this, its not a js err")
    }
}