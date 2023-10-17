import { MediaData } from '../types/Generic'
import API_ENDPOINT from './ApiEndpoint'

export async function fetchData(mediaState, dispatch) {
    try {
        let mediaMap = { ...mediaState.mediaMap }
        let dateMap = { ...mediaState.dateMap }
        let previousLast: string = mediaState.previousLast

        const limit = mediaState.maxMediaCount - mediaState.mediaCount
        if (limit < 1) {
            if (limit < 0) {
                console.error(`maxMediaCount (${mediaState.maxMediaCount}) less than mediaCount ${mediaState.mediaCount}`)
            }
            return
        }

        const url = new URL(`${API_ENDPOINT}/media`)
        url.searchParams.append('limit', limit.toString())
        url.searchParams.append('skip', mediaState.mediaCount.toString())
        url.searchParams.append('raw', mediaState.includeRaw.toString())
        const data = await fetch(url.toString()).then(res => res.json())
        const media: MediaData[] = data.Media

        let hasMoreMedia: boolean
        if (data.Media != null) {
            hasMoreMedia = data.MoreMedia
            for (const [_, value] of media.entries()) {
                mediaMap[value.FileHash] = value
                if (previousLast) {
                    mediaMap[value.FileHash].previous = mediaMap[previousLast]

                    mediaMap[previousLast].next = mediaMap[value.FileHash]
                }
                previousLast = value.FileHash
                const [date, _] = value.CreateDate.split("T")
                if (dateMap[date] == null) {
                    dateMap[date] = [value]
                } else {
                    dateMap[date].push(value)
                }
            }
        } else {
            hasMoreMedia = false
        }

        dispatch({
            type: 'add_media',
            mediaMap: mediaMap,
            dateMap: dateMap,
            hasMoreMedia: hasMoreMedia,
            previousLast: previousLast,
            addedCount: media?.length || 0
        })

    } catch (error) {
        console.error(error)
        throw new Error("Error fetching data - Ethan you wrote this, its not a js err")
    }
}