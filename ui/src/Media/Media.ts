import API_ENDPOINT from '../api/ApiEndpoint'
import { AuthHeaderT, mediaType } from '../types/Types'
import { useMediaStore } from './MediaStateControl'

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
    createDate?: number
    recognitionTags?: string[]
    pageCount?: number
    hidden?: boolean
    imported?: boolean
    likedBy?: string[]

    // Non-api props //

    // Object URLs
    thumbnail?: string
    fullres?: string

    Previous?: WeblensMedia
    Next?: WeblensMedia
    selected?: boolean
    mediaType?: mediaType

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
        if (this.data.hidden === undefined) {
            this.data.hidden = false
        }
    }

    Id(): string {
        return this.data.contentId
    }

    GetOwner(): string {
        return this.data.owner
    }

    IsImported(): boolean {
        return this.data.imported
    }

    IsHidden(): boolean {
        return this.data.hidden
    }

    SetHidden(hidden: boolean) {
        this.data.hidden = hidden
    }

    HighestQualityLoaded(): 'fullres' | 'thumbnail' | '' {
        if (this.data.fullres) {
            return 'fullres'
        } else if (this.data.thumbnail) {
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
        if (!this.data.mediaType && this.data.mimeType) {
            if (useMediaStore.getState().mediaTypeMap) {
                const mediaType =
                    useMediaStore.getState().mediaTypeMap[this.data.mimeType]
                if (!mediaType) {
                    console.error(
                        'Could not get media type',
                        this.data.mimeType
                    )
                }
                this.data.mediaType = mediaType
            }
        }
        return this.data.mediaType
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

    GetCreateDate(): Date {
        if (!this.data.createDate) {
            return new Date()
        }
        return new Date(this.data.createDate)
    }

    GetCreateTimestampUnix(): number {
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

    GetLikedBy(): string[] {
        if (this.data.likedBy === undefined || this.data.likedBy === null) {
            this.data.likedBy = []
        }
        return this.data.likedBy
    }

    SetLikedBy(username: string) {
        const index = this.data.likedBy.indexOf(username)
        if (index >= 0) {
            console.log('Unliking media', index)
            this.data.likedBy.splice(index, 1)
        } else {
            console.log('Liking media')
            this.data.likedBy.push(username)
        }
    }

    async LoadBytes(
        maxQuality: PhotoQuality,
        auth: AuthHeaderT,
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
                auth,
                this.data.abort.signal,
                pageNumber
            )
        }
        let fullres
        if (maxQuality === 'fullres' && !this.GetMediaType().IsVideo) {
            fullres = this.getImageData(
                'fullres',
                auth,
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

    SetThumbnailBytes(bytes: ArrayBuffer) {
        this.data.thumbnail = URL.createObjectURL(new Blob([bytes]))
    }

    async LoadInfo() {
        const url = new URL(`${API_ENDPOINT}/media/${this.data.contentId}/info`)
        await fetch(url.toString())
            .then((r) => r.json())
            .then((r) => {
                console.log(r)
                this.data = {
                    ...this.data,
                    ...(r as MediaDataT),
                }
                this.GetMediaType()
            })
            .catch((e) => {
                console.error('Failed to load media info', e)
            })
    }

    CancelLoad() {
        if (this.data.abort) {
            this.data.abort.abort('Cancelled')
        }
    }

    public StreamVideoUrl(auth: AuthHeaderT): string {
        return `${API_ENDPOINT}/media/${this.data.contentId}/stream`
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
            `${API_ENDPOINT}/media/${this.data.contentId}/${quality}`
        )

        if (pageNumber !== undefined) {
            url.searchParams.append('page', pageNumber.toString())
        }

        const res = fetch(url, {
            headers: {
                ...authHeader,
            },
            signal,
        })
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

// export class MediaStateT {
//     mediaMap: Map<string, WeblensMedia>
//     selectedMap: Map<string, boolean>
//     mediaList: WeblensMedia[]
//
//     private lastSelectedId: string = ''
//     private hovering: WeblensMedia
//
//     private includeRaw: boolean
//     private showHidden: boolean
//
//     private sortDirection: number
//
//     constructor(
//         map?: Map<string, WeblensMedia> | MediaStateT,
//         showRaw?: boolean,
//         showHidden?: boolean
//     ) {
//         this.sortDirection = 1
//
//         if (map && map instanceof Map) {
//             this.mediaMap = map
//             this.mediaList = Array.from(map.values()).sort(
//                 (a: WeblensMedia, b: WeblensMedia) =>
//                     (b.GetCreateTimestampUnix() - a.GetCreateTimestampUnix()) *
//                     this.sortDirection
//             )
//             this.selectedMap = new Map<string, boolean>()
//             this.includeRaw = showRaw
//             this.showHidden = showHidden
//         } else if (map && map instanceof MediaStateT) {
//             this.mediaMap = map.mediaMap
//             this.mediaList = map.mediaList
//             this.selectedMap = map.selectedMap
//             this.includeRaw = map.includeRaw
//             this.showHidden = map.showHidden
//             this.lastSelectedId = map.lastSelectedId
//             this.sortDirection = map.sortDirection
//             this.hovering = map.hovering
//         } else {
//             this.mediaMap = new Map<string, WeblensMedia>()
//             this.mediaList = []
//             this.selectedMap = new Map<string, boolean>()
//             this.includeRaw = showRaw || false
//             this.showHidden = showHidden || false
//         }
//     }
//
//     public getMedias() {
//         const medias = this.mediaList.filter(
//             (m) =>
//                 (!m.GetMediaType().IsRaw || this.includeRaw) &&
//                 (!m.IsHidden() || this.showHidden)
//         )
//         return medias
//     }
//
//     private sortedIndex(target: WeblensMedia) {
//         const targetDate = target.GetCreateTimestampUnix()
//         let low = 0,
//             high = this.mediaList.length
//
//         while (low < high) {
//             const mid = (low + high) >>> 1
//             if (
//                 (this.mediaList[mid].GetCreateTimestampUnix() > targetDate &&
//                     this.sortDirection === 1) ||
//                 (this.mediaList[mid].GetCreateTimestampUnix() < targetDate &&
//                     this.sortDirection === -1)
//             )
//                 low = mid + 1
//             else high = mid
//         }
//         return low
//     }
//
//     public getListIndex(mediaId: string): number {
//         const m = this.mediaMap.get(mediaId)
//         if (!m) {
//             return -1
//         }
//
//         const target = m.GetCreateTimestampUnix()
//
//         let start = 0,
//             end = this.mediaList.length - 1
//
//         // Iterate while start not meets end
//         while (start <= end) {
//             // Find the mid index
//             const mid = Math.floor((start + end) / 2)
//
//             // If element is present at
//             // mid, return True
//             if (this.mediaList[mid].GetCreateTimestampUnix() === target)
//                 return mid
//             // Else look in left or
//             // right half accordingly
//             else if (
//                 (this.mediaList[mid].GetCreateTimestampUnix() < target &&
//                     this.sortDirection === 1) ||
//                 (this.mediaList[mid].GetCreateTimestampUnix() > target &&
//                     this.sortDirection === -1)
//             )
//                 start = mid + 1
//             else end = mid - 1
//         }
//
//         return -1
//     }
//
//     public add(mediaId: string, media: WeblensMedia) {
//         if (!media.IsImported()) {
//             media.LoadInfo()
//         }
//
//         if (this.mediaMap.has(mediaId)) {
//             return
//         }
//
//         if (this.mediaList?.length > 0) {
//             const index = this.sortedIndex(media)
//             this.mediaList.splice(index, 0, media)
//
//             if (index > 0) {
//                 this.mediaList[index - 1].SetNextLink(media)
//                 this.mediaList[index].SetPrevLink(this.mediaList[index - 1])
//             }
//             if (index < this.mediaList.length - 1) {
//                 this.mediaList[index].SetNextLink(this.mediaList[index + 1])
//                 this.mediaList[index + 1].SetPrevLink(this.mediaList[index])
//             }
//         } else {
//             this.mediaList = [media]
//         }
//
//         this.mediaMap.set(mediaId, media)
//     }
//
//     public get(mediaId: string): WeblensMedia {
//         return this.mediaMap.get(mediaId)
//     }
//
//     public remove(mediaId: string): void {
//         this.mediaMap.delete(mediaId)
//         this.selectedMap.delete(mediaId)
//         const index = this.getListIndex(mediaId)
//         if (index != -1) {
//             this.mediaList.splice(index, 1)
//         }
//     }
//
//     private clear() {
//         this.mediaMap = new Map<string, WeblensMedia>()
//         this.selectedMap = new Map<string, boolean>()
//         this.mediaList = []
//     }
//
//     public setShowingRaw(raw: boolean) {
//         localStorage.setItem('showRaws', JSON.stringify(raw))
//         this.clear()
//         this.includeRaw = raw
//     }
//
//     public isShowingRaw(): boolean {
//         if (this.includeRaw === undefined) {
//             this.includeRaw = false
//         }
//         return this.includeRaw
//     }
//
//     public setShowingHidden(showHidden: boolean) {
//         localStorage.setItem('showHidden', JSON.stringify(showHidden))
//         this.clear()
//         this.showHidden = showHidden
//     }
//
//     public isShowingHidden(): boolean {
//         if (this.showHidden === undefined) {
//             this.showHidden = undefined
//         }
//         return this.showHidden
//     }
//
//     public setSelected(mediaId: string, selectMany: boolean) {
//         if (selectMany) {
//             let startIndex = this.mediaMap.get(mediaId).GetAbsIndex()
//             let endIndex
//             if (!this.lastSelectedId) {
//                 endIndex = startIndex
//             } else {
//                 endIndex = this.mediaMap.get(this.lastSelectedId).GetAbsIndex()
//             }
//
//             if (endIndex < startIndex) {
//                 ;[startIndex, endIndex] = [endIndex, startIndex]
//             }
//
//             for (const m of this.mediaList.slice(startIndex, endIndex + 1)) {
//                 this.selectedMap.set(m.Id(), true)
//                 m.SetSelected(true)
//             }
//         } else {
//             if (this.selectedMap.get(mediaId)) {
//                 this.mediaMap.get(mediaId).SetSelected(false)
//                 this.selectedMap.delete(mediaId)
//             } else {
//                 this.mediaMap.get(mediaId).SetSelected(true)
//                 this.selectedMap.set(mediaId, true)
//             }
//         }
//         this.lastSelectedId = mediaId
//     }
//
//     public isSelected(mediaId: string): boolean {
//         return Boolean(this.selectedMap.get(mediaId))
//     }
//
//     public getAllSelectedIds(): string[] {
//         return Array.from(this.selectedMap.keys())
//     }
//
//     public getAllSelectedMedias(): WeblensMedia[] {
//         return this.getAllSelectedIds().map((id) => this.mediaMap.get(id))
//     }
//
//     public clearSelected() {
//         this.selectedMap.forEach((_, sel) => {
//             this.mediaMap.get(sel).SetSelected(false)
//         })
//
//         this.selectedMap.clear()
//         this.lastSelectedId = ''
//     }
//
//     public getLastSelected(): WeblensMedia {
//         return this.mediaMap.get(this.lastSelectedId)
//     }
//
//     public setHoveringId(hoveringId: string) {
//         this.hovering = this.mediaMap.get(hoveringId)
//     }
//
//     public getHovering(): WeblensMedia {
//         return this.hovering
//     }
//
//     public async loadNew(mediaId: string): Promise<WeblensMedia> {
//         if (!mediaId) {
//             return null
//         }
//
//         if (this.mediaMap.has(mediaId)) {
//             return this.mediaMap.get(mediaId)
//         }
//         const url = new URL(`${API_ENDPOINT}/media/${mediaId}/info`)
//         const newData = await fetch(url.toString()).then((r) => r.json())
//         const newM = new WeblensMedia(newData as MediaDataT)
//
//         if (newM) {
//             this.mediaMap.set(mediaId, newM)
//         }
//
//         return newM
//     }
// }

// export type MediaAction = {
//     medias?: WeblensMedia[]
//     media?: WeblensMedia
//
//     type: string
//     mediaId?: string
//     endMediaId?: string
//
//     mediaIds?: string[]
//
//     raw?: boolean
//     hidden?: boolean
//     selectMany?: boolean
//
//     startIndex?: number
//     endIndex?: number
// }

// export function mediaReducer(
//     state: MediaStateT,
//     action: MediaAction
// ): MediaStateT {
//     switch (action.type) {
//         case 'add_medias': {
//             for (const media of action.medias) {
//                 if (state.mediaMap.has(media.Id())) {
//                     continue
//                 }
//                 state.add(media.Id(), media)
//             }
//             break
//         }
//
//         case 'add_media_id': {
//             if (state.mediaMap.has(action.mediaId)) {
//                 return state
//             }
//
//             const newM = new WeblensMedia({ contentId: action.mediaId })
//             state.add(newM.Id(), newM)
//             break
//         }
//
//         case 'add_media_ids': {
//             for (const mId of action.mediaIds) {
//                 const newM = new WeblensMedia({ contentId: mId })
//                 state.add(newM.Id(), newM)
//             }
//             break
//         }
//
//         case 'set_selected': {
//             state.setSelected(action.mediaId, action.selectMany)
//             break
//         }
//
//         case 'remove_by_ids': {
//             for (const mId of action.mediaIds) {
//                 state.remove(mId)
//             }
//             break
//         }
//
//         case 'set_raw_toggle': {
//             if (action.raw === state.isShowingRaw()) {
//                 return state
//             }
//             window.scrollTo({
//                 top: 0,
//                 behavior: 'smooth',
//             })
//
//             state.setShowingRaw(action.raw)
//             break
//         }
//
//         case 'set_hidden_toggle': {
//             if (action.hidden === state.isShowingHidden()) {
//                 return state
//             }
//             window.scrollTo({
//                 top: 0,
//                 behavior: 'smooth',
//             })
//
//             state.setShowingHidden(action.hidden)
//             break
//         }
//
//         case 'clear_selected': {
//             state.clearSelected()
//             break
//         }
//
//         case 'set_hovering': {
//             state.setHoveringId(action.mediaId)
//             break
//         }
//
//         case 'hide_medias': {
//             for (const mediaId of action.mediaIds) {
//                 const m = state.mediaMap.get(mediaId)
//                 if (m) {
//                     m.SetHidden(action.hidden)
//                 } else {
//                     console.error('Trying to hide unknown mediaId', mediaId)
//                 }
//             }
//             break
//         }
//
//         case 'refresh': {
//             break
//         }
//
//         default: {
//             console.error('Unknown action type', action.type)
//             return state
//         }
//     }
//
//     return new MediaStateT(state)
// }

export default WeblensMedia
