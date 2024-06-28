import API_ENDPOINT from '../api/ApiEndpoint'
import { AuthHeaderT, mediaType } from '../types/Types'

export interface MediaDataT {
    contentId: string
    mimeType?: string

    fileIds?: string[]
    thumbnailCacheId?: string
    fullresCacheIds?: string
    blurHash?: string
    owner?: string
    width?: number
    height?: number
    videoLength?: number
    createDate?: string
    recognitionTags?: string[]
    pageCount?: number
    hidden?: boolean
    imported?: boolean

    // Non-api props //

    // Object URLs
    thumbnail?: string
    fullres?: string

    Previous?: WeblensMedia
    Next?: WeblensMedia
    selected?: boolean
    mediaType?: mediaType
    // Display: boolean
    ImgRef?: any

    abort?: AbortController
    index?: number
}

export type PhotoQuality = 'thumbnail' | 'fullres'

class WeblensMedia {
    private data: MediaDataT
    private loadError: PhotoQuality

    constructor(init: MediaDataT) {
        if (typeof init.contentId !== 'string') {
            console.trace()
        }
        this.data = init as MediaDataT
        this.data.selected = false
    }

    Id(): string {
        return this.data.contentId
    }

    IsImported(): boolean {
        return this.data.imported
    }

    IsHidden(): boolean {
        return this.data.hidden
    }

    HighestQualityLoaded(): 'fullres' | 'thumbnail' | '' {
        if (Boolean(this.data.fullres)) {
            return 'fullres'
        } else if (Boolean(this.data.thumbnail)) {
            return 'thumbnail'
        } else {
            return ''
        }
    }

    HasQualityLoaded(q: 'fullres' | 'thumbnail'): boolean {
        if (q == 'fullres') {
            return Boolean(this.data.fullres)
        } else {
            return Boolean(this.data.thumbnail)
        }
    }

    GetMediaType(): mediaType {
        if (!this.data.mediaType) {
            const typeMap = JSON.parse(localStorage.getItem('mediaTypeMap'))
            this.data.mediaType = typeMap.typeMap[this.data.mimeType]
        }
        return this.data.mediaType
    }

    SetThumbnailBytes(buf: ArrayBuffer) {
        // this.data.thumbnail = buf;
    }

    SetFullresBytes(buf: ArrayBuffer) {
        // this.data.fullres = buf;
    }

    GetFileIds(): string[] {
        if (!this.data.fileIds) {
            return []
        }

        return this.data.fileIds
    }

    SetSelected(s: boolean) {
        this.data.selected = s
    }

    IsSelected(): boolean {
        return this.data.selected
    }

    IsDisplayable(): boolean {
        const mt = this.GetMediaType()
        if (!mt) {
            return false
        }
        return this.GetMediaType().IsDisplayable
    }

    GetVideoLength(): number {
        return this.data.videoLength
    }

    HasLoadError(): PhotoQuality {
        return this.loadError
    }

    SetImgRef(r) {
        this.data.ImgRef = r
    }

    GetImgRef() {
        return this.data.ImgRef
    }

    GetHeight(): number {
        return this.data.height
    }

    GetWidth(): number {
        return this.data.width
    }

    SetNextLink(next: WeblensMedia) {
        this.data.Next = next
    }

    Next(): WeblensMedia {
        return this.data.Next
    }

    SetPrevLink(prev: WeblensMedia) {
        this.data.Previous = prev
    }

    Prev(): WeblensMedia {
        return this.data.Previous
    }

    MatchRecogTag(searchTag: string): boolean {
        if (!this.data.recognitionTags) {
            return false
        }

        return this.data.recognitionTags.includes(searchTag)
    }

    GetPageCount(): number {
        return this.data.pageCount
    }

    GetCreateDate(): string {
        return this.data.createDate
    }

    GetAbsIndex(): number {
        return this.data.index
    }

    SetAbsIndex(index: number) {
        this.data.index = index
    }

    GetObjectUrl(quality: 'thumbnail' | 'fullres'): string {
        if (quality == 'thumbnail' || !this.data.fullres) {
            return this.data.thumbnail
        } else if (quality == 'fullres') {
            return this.data.fullres
        }
    }

    async LoadBytes(
        maxQuality: PhotoQuality,
        authHeader: AuthHeaderT,
        pageNumber?,
        thumbFinished?: () => void,
        fullresFinished?: () => void
    ) {
        if (!this.data.abort || this.data.abort.signal.aborted) {
            this.data.abort = new AbortController()
        }

        let thumb
        if (!this.data.thumbnail) {
            thumb = this.getImageData(
                'thumbnail',
                authHeader,
                this.data.abort.signal,
                pageNumber
            )
        }
        let fullres
        if (maxQuality === 'fullres' && !this.GetMediaType().IsVideo) {
            fullres = this.getImageData(
                'fullres',
                authHeader,
                this.data.abort.signal,
                pageNumber
            )
        }
        if (thumb) {
            await thumb.then((updated: boolean) => {
                if (updated) {
                    thumbFinished()
                }
            })
        }

        if (fullres) {
            await fullres.then((updated: boolean) => {
                if (updated) {
                    fullresFinished()
                }
            })
        }
    }

    async LoadInfo() {
        const url = new URL(`${API_ENDPOINT}/media/${this.data.contentId}/info`)
        fetch(url.toString())
            .then((r) => r.json())
            .then((r) => {
                this.data = {
                    ...this.data,
                    ...(r as MediaDataT),
                }
            })
    }

    CancelLoad() {
        if (this.data.abort) {
            this.data.abort.abort('Cancelled')
        }
    }

    public StreamVideoUrl(authHeader: AuthHeaderT): string {
        return `${API_ENDPOINT}/media/${this.data.contentId}/stream`
        // const url = new URL(
        // )
        // fetch(url, { headers: authHeader })
    }

    private async getImageData(
        quality: PhotoQuality,
        authHeader: AuthHeaderT,
        signal: AbortSignal,
        pageNumber?: number
    ) {
        if (!this.data.contentId) {
            console.error('Trying to get image of media without id')
            return
        }
        const url = new URL(
            `${
                // doPublic ? PUBLIC_ENDPOINT : API_ENDPOINT
                API_ENDPOINT
            }/media/${this.data.contentId}/${quality}`
        )

        if (pageNumber !== undefined) {
            url.searchParams.append('page', pageNumber.toString())
        }

        const res = fetch(url, { headers: authHeader, signal })
            .then((res) => {
                if (res.status !== 200) {
                    return Promise.reject(res.statusText)
                }

                return res.arrayBuffer()
            })
            .then((r) => {
                this.data[quality] = URL.createObjectURL(new Blob([r]))
                return true
            })
            .catch((r) => {
                if (!signal.aborted) {
                    console.error('Failed to get image from server:', r)
                    this.loadError = quality
                }
                return false
            })

        return res
    }
}

export class MediaStateT {
    mediaMap: Map<string, WeblensMedia>
    selectedMap: Map<string, boolean>
    mediaList: WeblensMedia[]

    constructor(map?: Map<string, WeblensMedia> | MediaStateT) {
        if (!map) {
            this.mediaMap = new Map<string, WeblensMedia>()
            this.mediaList = []
            this.selectedMap = new Map<string, boolean>()
        } else if (map instanceof Map) {
            this.mediaMap = map
        } else if (map && map instanceof MediaStateT) {
            this.mediaMap = map.mediaMap
            this.mediaList = map.mediaList
            this.selectedMap = map.selectedMap
        } else {
            console.error('Unable to construct MediaStateT')
        }
    }

    private sortedIndex(target: WeblensMedia) {
        const targetDate = target.GetCreateDate()
        var low = 0,
            high = this.mediaList?.length

        while (low < high) {
            var mid = (low + high) >>> 1
            if (this.mediaList[mid].GetCreateDate() > targetDate) low = mid + 1
            else high = mid
        }
        return low
    }

    public getListIndex(mediaId: string): number {
        const m = this.mediaMap.get(mediaId)
        if (!m) {
            return -1
        }

        const target = m.GetCreateDate()

        let start = 0,
            end = this.mediaList.length - 1

        // Iterate while start not meets end
        while (start <= end) {
            // Find the mid index
            let mid = Math.floor((start + end) / 2)

            // If element is present at
            // mid, return True
            if (this.mediaList[mid].GetCreateDate() === target) return mid
            // Else look in left or
            // right half accordingly
            else if (this.mediaList[mid].GetCreateDate() < target)
                start = mid + 1
            else end = mid - 1
        }

        return -1
    }

    public add(mediaId: string, media: WeblensMedia) {
        if (!media.IsImported()) {
            media.LoadInfo()
        }

        if (this.mediaList?.length > 0) {
            const index = this.sortedIndex(media)
            this.mediaList.splice(index, 0, media)

            if (index > 0) {
                this.mediaList[index - 1].SetNextLink(media)
                this.mediaList[index].SetPrevLink(this.mediaList[index - 1])
            }
            if (index < this.mediaList.length - 1) {
                this.mediaList[index].SetNextLink(this.mediaList[index + 1])
                this.mediaList[index + 1].SetPrevLink(this.mediaList[index])
            }
        } else {
            this.mediaList = [media]
        }

        this.mediaMap.set(mediaId, media)
        this.selectedMap.set(mediaId, false)
    }

    public setSelected(mediaId: string, endMediaId?: string) {
        if (endMediaId !== undefined) {
            let startIndex = this.getListIndex(mediaId)
            let endIndex = this.getListIndex(endMediaId)

            if (endIndex < startIndex) {
                ;[startIndex, endIndex] = [endIndex, startIndex]
            }

            for (const m of this.mediaList.slice(startIndex, endIndex)) {
                this.selectedMap.set(m.Id(), true)
            }
        } else {
            this.selectedMap.set(mediaId, true)
        }
    }

    public isSelected(mediaId: string): boolean {
        return this.selectedMap.get(mediaId)
    }

    public getAllSelected(): string[] {
        return Array.from(this.selectedMap.entries())
            .filter(([_, s]) => s)
            .map(([m, _]) => m)
    }

    public async loadNew(mediaId: string): Promise<WeblensMedia> {
        if (!mediaId) {
            return null
        }

        if (this.mediaMap.has(mediaId)) {
            return this.mediaMap.get(mediaId)
        }

        const url = new URL(`${API_ENDPOINT}/media/${mediaId}/info`)
        const newData = await fetch(url.toString()).then((r) => r.json())
        const newM = new WeblensMedia(newData as MediaDataT)

        if (newM) {
            this.mediaMap.set(mediaId, newM)
        }

        return newM
    }

    public get(mediaId: string) {
        const m = this.mediaMap.get(mediaId)
        if (m) return
    }
}

export type MediaAction = {
    type: string
    medias?: WeblensMedia[]
    media?: WeblensMedia
    mediaId?: string
    endMediaId?: string
    mediaIds?: string[]
}

export function mediaReducer(
    state: MediaStateT,
    action: MediaAction
): MediaStateT {
    switch (action.type) {
        case 'add_media': {
            if (state.mediaMap.has(action.media.Id())) {
                return state
            }

            state.add(action.media.Id(), action.media)
            break
        }
        case 'add_media_id': {
            if (state.mediaMap.has(action.mediaId)) {
                return state
            }

            const newM = new WeblensMedia({ contentId: action.mediaId })
            state.add(newM.Id(), newM)
            break
        }
        case 'add_media_ids': {
            for (const mId of action.mediaIds) {
                const newM = new WeblensMedia({ contentId: mId })
                state.add(newM.Id(), newM)
            }
            break
        }

        case 'set_selected': {
            state.setSelected(action.mediaId, action.endMediaId)
            break
        }

        case 'remove_by_ids': {
            for (const mId of action.mediaIds) {
                state.mediaMap.delete(mId)
                state.selectedMap.delete(mId)
                const index = state.getListIndex(mId)
                if (index != -1) {
                    state.mediaList.splice(index, 1)
                }
            }
            action.mediaIds
        }

        case 'refresh': {
            break
        }
        default: {
            console.error('Unknown action type', action.type)
            return state
        }
    }

    return new MediaStateT(state)
    // new Map<string, WeblensMedia>(state)
}

export default WeblensMedia
