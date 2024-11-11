import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import MediaApi from '@weblens/api/MediaApi'
import { MediaInfo } from '@weblens/api/swag'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { mediaType } from 'types/Types'

export enum PhotoQuality {
    LowRes = 'thumbnail',
    HighRes = 'fullres',
}

class WeblensMedia {
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

    previous?: WeblensMedia
    next?: WeblensMedia
    selected?: boolean
    mediaType?: mediaType

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
        if (this.fullres) {
            return PhotoQuality.HighRes
        } else if (this.thumbnail) {
            return PhotoQuality.LowRes
        } else {
            return ''
        }
    }

    HasQualityLoaded(q: PhotoQuality.HighRes | PhotoQuality.LowRes): boolean {
        if (q == PhotoQuality.HighRes) {
            return Boolean(this.fullres)
        } else {
            return Boolean(this.thumbnail)
        }
    }

    GetMediaType(): mediaType {
        if (!this.mediaType && this.mimeType) {
            if (useMediaStore.getState().mediaTypeMap) {
                const mediaType =
                    useMediaStore.getState().mediaTypeMap[this.mimeType]
                if (!mediaType) {
                    console.error('Could not get media type', this.mimeType)
                }
                this.mediaType = mediaType
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
        return this.videoLength
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

    SetNextLink(next: WeblensMedia) {
        this.next = next
    }

    Next(): WeblensMedia {
        return this.next
    }

    SetPrevLink(prev: WeblensMedia) {
        this.previous = prev
    }

    Prev(): WeblensMedia {
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

    GetObjectUrl(quality: PhotoQuality.LowRes | PhotoQuality.HighRes): string {
        if (quality == PhotoQuality.LowRes || !this.fullres) {
            return this.thumbnail
        } else if (quality == PhotoQuality.HighRes) {
            return this.fullres
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
        if (!this.thumbnail) {
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
                pageNumber
            )
        }
        if (thumb) {
            await thumb.then((updated: boolean) => {
                if (updated && onThumbFinished) {
                    onThumbFinished()
                }
            })
        }

        if (fullres) {
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
        return MediaApi.getMediaImage(this.contentId, quality, pageNumber, {
            responseType: 'blob',
            signal: signal,
        })
            .then((res) => {
                if (res.status !== 200) {
                    return Promise.reject(res.statusText)
                }

                const blob = new Blob([res.data])
                this[quality] = URL.createObjectURL(blob)
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
