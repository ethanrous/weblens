import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import MediaApi from '@weblens/api/MediaApi'
import { MediaInfo, MediaTypeInfo } from '@weblens/api/swag'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'

export enum PhotoQuality {
    LowRes = 'thumbnail',
    HighRes = 'fullres',
}

class WeblensMedia implements MediaInfo {
    contentId: string
    mimeType?: string

    fileIds?: string[]
    thumbnailCacheId?: string
    fullresCacheIds?: string
    blurHash?: string
    owner?: string
    width?: number
    height?: number
    duration?: number
    createDate?: number
    recognitionTags?: string[]
    pageCount?: number
    hidden?: boolean
    imported?: boolean
    likedBy?: string[]

    // Non-api props //

    // Object URLs
    thumbnail?: string
    fullres?: string[]

    previous?: WeblensMedia
    next?: WeblensMedia
    selected?: boolean
    mediaType?: MediaTypeInfo

    abort?: AbortController
    index?: number

    private loadError: PhotoQuality

    constructor(init: MediaInfo) {
        if (typeof init.contentId !== 'string') {
            console.trace(init)
        }

        Object.assign(this, init)

        this.selected = false
        if (this.hidden === undefined) {
            this.hidden = false
        }
        this.fullres = new Array<string>(this.pageCount).fill(null)
    }

    Id(): string {
        return this.contentId
    }

    GetOwner(): string {
        return this.owner
    }

    IsImported(): boolean {
        return this.imported
    }

    IsHidden(): boolean {
        return this.hidden
    }

    SetHidden(hidden: boolean) {
        this.hidden = hidden
    }

    HighestQualityLoaded(): PhotoQuality.HighRes | PhotoQuality.LowRes | '' {
        if (this.fullres[this.fullres.length - 1] !== null) {
            return PhotoQuality.HighRes
        } else if (this.thumbnail) {
            return PhotoQuality.LowRes
        } else {
            return ''
        }
    }

    HasQualityLoaded(q: PhotoQuality.HighRes | PhotoQuality.LowRes): boolean {
        if (q === PhotoQuality.HighRes) {
            return this.fullres[this.fullres.length - 1] !== null
        } else {
            return Boolean(this.thumbnail)
        }
    }

    GetMediaType(): MediaTypeInfo {
        if (!this.mediaType && this.mimeType) {
            const typeMap = useMediaStore.getState().mediaTypeMap
            if (typeMap?.mimeMap) {
                const mediaType = typeMap.mimeMap[this.mimeType]
                if (!mediaType) {
                    console.error('Could not get media type', this.mimeType)
                }
                this.mediaType = mediaType
            } else {
                console.error('Could not get media type map')
            }
        }
        return this.mediaType
    }

    GetFileIds(): string[] {
        if (!this.fileIds) {
            return []
        }

        return this.fileIds
    }

    SetSelected(s: boolean) {
        this.selected = s
    }

    IsSelected(): boolean {
        return this.selected
    }

    IsDisplayable(): boolean {
        const mt = this.GetMediaType()
        if (!mt) {
            return false
        }
        return this.GetMediaType().IsDisplayable
    }

    GetVideoLength(): number {
        if (!this.duration) {
            console.error('No video length for', this.contentId)
            return 0
        }
        return this.duration
    }

    HasLoadError(): PhotoQuality {
        return this.loadError
    }

    GetHeight(): number {
        return this.height
    }

    GetWidth(): number {
        return this.width
    }

    SetNextLink(next?: WeblensMedia) {
        this.next = next
    }

    Next(): WeblensMedia | undefined {
        return this.next
    }

    SetPrevLink(prev?: WeblensMedia) {
        this.previous = prev
    }

    Prev(): WeblensMedia | undefined {
        return this.previous
    }

    MatchRecogTag(searchTag: string): boolean {
        if (!this.recognitionTags) {
            return false
        }

        return this.recognitionTags.includes(searchTag)
    }

    GetPageCount(): number {
        return this.pageCount
    }

    GetCreateDate(): Date {
        if (!this.createDate) {
            return new Date()
        }
        return new Date(this.createDate)
    }

    GetCreateTimestampUnix(): number {
        return this.createDate
    }

    GetAbsIndex(): number {
        return this.index
    }

    SetAbsIndex(index: number) {
        this.index = index
    }

    GetObjectUrl(
        quality: PhotoQuality.LowRes | PhotoQuality.HighRes,
        pageNumber?: number
    ): string {
        if (quality === PhotoQuality.LowRes || !this.fullres) {
            return this.thumbnail
        } else if (quality === PhotoQuality.HighRes) {
            return this.fullres[pageNumber ?? 0]
        }
    }

    GetLikedBy(): string[] {
        if (this.likedBy === undefined || this.likedBy === null) {
            this.likedBy = []
        }
        return this.likedBy
    }

    SetLikedBy(username: string) {
        const index = this.likedBy.indexOf(username)
        if (index >= 0) {
            this.likedBy.splice(index, 1)
        } else {
            this.likedBy.push(username)
        }
    }

    async LoadBytes(
        maxQuality: PhotoQuality,
        pageNumber?: number,
        onThumbFinished?: () => void,
        onFullresFinished?: () => void
    ) {
        if (!this.abort || this.abort.signal.aborted) {
            this.abort = new AbortController()
        }

        let thumb: Promise<boolean>
        if (!this.thumbnail && pageNumber === 0) {
            thumb = this.getImageData(PhotoQuality.LowRes, this.abort.signal)
        }
        let fullres: Promise<boolean>
        if (
            maxQuality === PhotoQuality.HighRes &&
            !this.GetMediaType().IsVideo
        ) {
            fullres = this.getImageData(
                PhotoQuality.HighRes,
                this.abort.signal,
                pageNumber ?? 0
            )
        }

        if (thumb !== undefined) {
            await thumb.then((updated: boolean) => {
                if (updated && onThumbFinished) {
                    onThumbFinished()
                }
            })
        }

        if (fullres !== undefined) {
            await fullres.then((updated: boolean) => {
                if (updated && onFullresFinished) {
                    onFullresFinished()
                }
            })
        }
    }

    SetThumbnailBytes(bytes: ArrayBuffer) {
        this.thumbnail = URL.createObjectURL(new Blob([bytes]))
    }

    async LoadInfo() {
        const res = await MediaApi.getMediaInfo(this.contentId)
        Object.assign(this, res.data)
    }

    CancelLoad() {
        if (this.abort) {
            this.abort.abort('Cancelled')
            this.abort = null
        }
    }

    public StreamVideoUrl(): string {
        return `${API_ENDPOINT}/media/${this.contentId}/stream`
    }

    private async getImageData(
        quality: PhotoQuality,
        signal: AbortSignal,
        pageNumber?: number
    ) {
        return MediaApi.getMediaImage(
            this.contentId,
            'webp',
            quality,
            pageNumber,
            {
                responseType: 'blob',
                signal: signal,
            }
        )
            .then((res) => {
                if (res.status !== 200) {
                    return Promise.reject(new Error(res.statusText))
                }

                const blob = new Blob([res.data])
                switch (quality) {
                    case PhotoQuality.LowRes: {
                        this.thumbnail = URL.createObjectURL(blob)
                        break
                    }
                    case PhotoQuality.HighRes: {
                        this.fullres[pageNumber] = URL.createObjectURL(blob)
                        break
                    }
                }
                return true
            })
            .catch((r) => {
                if (!signal.aborted) {
                    console.error('Failed to get image from server:', r)
                    this.loadError = quality
                }
                return false
            })
    }
}

export default WeblensMedia
