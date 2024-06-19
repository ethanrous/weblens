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
    createDate?: string
    recognitionTags?: string[]
    pageCount?: number
    hidden?: boolean

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
        this.data = init
        this.data.selected = false
    }

    Id(): string {
        return this.data.contentId
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

    GetImgUrl(quality: 'thumbnail' | 'fullres'): string {
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
        if (maxQuality === 'fullres') {
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

    CancelLoad() {
        if (this.data.abort) {
            this.data.abort.abort('Cancelled')
        }
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

export default WeblensMedia
