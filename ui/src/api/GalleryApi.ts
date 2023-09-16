export async function fetchData(mediaState, dispatch) {
    try {
        let mediaList = mediaState.mediaList
        let mediaIdMap = mediaState.mediaIdMap
        let dateMap = {}
        const limit = mediaState.maxMediaCount - mediaState.mediaCount
        if (limit < 1) {
            if (limit < 0) {
                console.error(`maxMediaCount (${mediaState.maxMediaCount}) less than mediaCount ${mediaState.mediaCount}`)
            }
            return
        }

        const url = new URL("http://localhost:3000/api/media")
        url.searchParams.append('limit', limit.toString())
        url.searchParams.append('skip', mediaState.mediaCount.toString())
        url.searchParams.append('raw', mediaState.includeRaw.toString())
        const response = await fetch(url.toString())
        const data = await response.json()

        let moreMedia: boolean
        if (data.Media != null) {
            moreMedia = data.MoreMedia

            const prevousLast = mediaList.length > 0 ? mediaList[mediaList.length - 1] : null
            mediaList.push(...data.Media)
            for (const [index, value] of data.Media.entries()) {

                // This is the first media in this fetch, and no prior media exists
                if (index === 0 && prevousLast == null) {
                    mediaIdMap[value.FileHash] = {
                        previous: null,
                        next: null
                    }

                    // This is the first media in this fetch, but prior media does exist
                } else if (index === 0) {
                    mediaIdMap[value.FileHash] = {
                        previous: prevousLast.FileHash,
                        next: null
                    }
                    mediaIdMap[prevousLast.FileHash].next = value.FileHash

                    // Not the first media in this fetch
                } else {
                    mediaIdMap[value.FileHash] = {
                        previous: data.Media[index - 1].FileHash,
                        next: null
                    }
                    mediaIdMap[data.Media[index - 1].FileHash].next = value.FileHash
                }
            }


            for (const item of mediaList) {
                const [date, _] = item.CreateDate.split("T")
                if (dateMap[date] == null) {
                    dateMap[date] = [item]
                } else {
                    dateMap[date].push(item)
                }
            }
        } else {
            moreMedia = false
        }

        dispatch({
            type: 'add_media',
            mediaList: mediaList,
            mediaIdMap: mediaIdMap,
            hasMoreMedia: moreMedia,
            dateMap: dateMap
        })

    } catch (error) {
        console.log(error)
        throw new Error("Error fetching data - Ethan you wrote this, its not a js err")
    }
}