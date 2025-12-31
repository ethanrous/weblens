import type { MediaInfo, MediaTypeInfo } from '@ethanrous/weblens-api'
import type { AxiosError, AxiosResponse } from 'axios'
import { API_ENDPOINT, useWeblensAPI } from '~/api/AllApi'

export enum PhotoQuality {
    LowRes = 'thumbnail',
    HighRes = 'fullres',
}

class WeblensMedia implements MediaInfo {
    contentID: string = ''
    mimeType?: string

    fileIDs?: string[]
    location?: [number, number]
    thumbnailCacheID?: string
    fullresCacheIds?: string
    blurHash?: string
    owner: string = ''
    width: number = -1
    height: number = -1
    duration?: number
    createDate: number = -1
    recognitionTags?: string[]
    pageCount: number = -1
    hidden: boolean = false
    imported: boolean = false
    likedBy: string[] = []

    // Non-api props //

    // Object URLs
    thumbnail: string = ''
    fullres: string[]

    previous?: WeblensMedia
    next?: WeblensMedia
    selected: boolean = false
    mediaType?: MediaTypeInfo

    abort?: AbortController
    index: number = -1

    private loadError?: PhotoQuality

    constructor(init: MediaInfo) {
        Object.assign(this, init)

        this.selected = false
        if (this.hidden === undefined) {
            this.hidden = false
        }

        this.fullres = new Array(Math.max(0, this.pageCount)).fill(0).map(() => '')
    }

    public get id(): string {
        return this.contentId
    }

    ID(): string {
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
            return this.fullres[this.fullres.length - 1] !== ''
        } else {
            return Boolean(this.thumbnail)
        }
    }

    GetMediaType(): MediaTypeInfo | undefined {
        if (this.mediaType) {
            return this.mediaType
        }

        if (!this.mimeType) {
            for (const mediaType of Object.values(mediaTypes)) {
                if (mediaType.FileExtension.includes(this.contentID.split('.').pop() ?? '')) {
                    this.mediaType = mediaType
                    return mediaType
                }
            }

            console.error('Could not determine media type for', this.contentID)
            return mediaTypes.generic
        }

        this.mediaType = mediaTypes[this.mimeType]
        return this.mediaType
    }

    GetFileIds(): string[] {
        if (!this.fileIDs) {
            return []
        }

        return this.fileIDs
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

        return mt.IsDisplayable ?? false
    }

    IsPdf(): boolean {
        return this.mimeType === 'application/pdf'
    }

    IsVideo(): boolean {
        const mt = this.GetMediaType()
        if (!mt) {
            return false
        }

        return mt.IsVideo ?? false
    }

    GetVideoLength(): number {
        if (!this.duration) {
            console.error('No video length for', this.contentID)
            return 0
        }

        return this.duration
    }

    HasLoadError(): PhotoQuality | undefined {
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

    GetObjectUrl(quality: PhotoQuality.LowRes | PhotoQuality.HighRes, pageNumber?: number): string {
        if (quality === PhotoQuality.LowRes || !this.fullres) {
            return this.thumbnail
        } else if (quality === PhotoQuality.HighRes) {
            return this.fullres[pageNumber ?? 0] ?? ''
        }

        return ''
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

    async LoadBytes(maxQuality: PhotoQuality, pageNumber?: number, controller?: AbortController): Promise<string> {
        if (!controller && (!this.abort || this.abort.signal.aborted)) {
            this.abort = new AbortController()
        }

        if (maxQuality === PhotoQuality.LowRes && this.thumbnail) {
            return this.thumbnail
        } else if (maxQuality === PhotoQuality.HighRes && this.fullres[pageNumber ?? 0]) {
            return this.fullres[pageNumber ?? 0] ?? ''
        }

        const data = await this.getImageData(maxQuality, controller?.signal ?? this.abort!.signal, pageNumber)

        return data
    }

    SetThumbnailBytes(bytes: ArrayBuffer) {
        this.thumbnail = URL.createObjectURL(new Blob([bytes]))
    }

    // This allows us to load the media info from the server with just the contentID.
    // Something like `const media = new WeblensMedia({ contentID: 'xxxx' }).LoadInfo()`
    async LoadInfo(): Promise<WeblensMedia> {
        const res = await useWeblensAPI().MediaAPI.getMediaInfo(this.contentID)
        Object.assign(this, res.data)

        return this
    }

    CancelLoad() {
        if (this.abort) {
            this.abort.abort('Cancelled')
            this.abort = undefined
        }
    }

    public ImgUrl(quality: PhotoQuality = PhotoQuality.LowRes): string {
        let format = 'webp'
        if (this.mimeType === 'application/pdf' && quality === PhotoQuality.HighRes) {
            format = 'pdf'
        }

        return `${API_ENDPOINT.value}/media/${this.contentID}.${format}?quality=${quality}&page=0`
    }

    public MediaUrl(): string {
        return `${window.location.origin}/media/${this.contentID}`
    }

    public StreamVideoUrl(): string {
        return `${API_ENDPOINT.value}/media/${this.contentID}/stream`
    }

    private async getImageData(quality: PhotoQuality, signal: AbortSignal, pageNumber: number = 0): Promise<string> {
        return useWeblensAPI()
            .MediaAPI.getMediaImage(this.contentID, 'webp', quality, pageNumber, {
                responseType: 'blob',
                signal: signal,
            })
            .then((res: AxiosResponse) => {
                if (res.status !== 200) {
                    return Promise.reject(new Error(res.statusText))
                }

                const blob = new Blob([res.data])
                switch (quality) {
                    case PhotoQuality.LowRes: {
                        this.thumbnail = URL.createObjectURL(blob)
                        return this.thumbnail
                    }
                    case PhotoQuality.HighRes: {
                        this.fullres[pageNumber] = URL.createObjectURL(blob)
                        return this.fullres[pageNumber]
                    }
                }
            })
            .catch((r: AxiosError) => {
                if (!signal.aborted) {
                    console.error('Failed to get image from server:', r)
                    this.loadError = quality
                }
                return ''
            })
    }
}

export type GalleryRowItem = { m: WeblensMedia; w: number }

export type GalleryRowInfo = {
    rowHeight: number
    rowWidth: number
    items: GalleryRowItem[]
}

export function isGalleryRow(input?: GalleryRowInfo): input is GalleryRowInfo {
    if (input === undefined) {
        return false
    }

    return input.rowHeight !== undefined && input.rowWidth !== undefined && input.items !== undefined
}

export function GetMediaRows(
    medias: WeblensMedia[],
    baseRowHeight: number,
    viewWidth: number,
    marginSize: number,
    totalMediaCount: number,
): { rows: GalleryRowInfo[]; remainingGap: number } {
    if (medias.length === 0 || viewWidth === -1) {
        return { rows: [], remainingGap: 0 }
    }

    const mediasCpy = [...medias]

    const MAX_ROW_WIDTH = viewWidth

    const rows: GalleryRowInfo[] = []
    let currentRowWidth = 0
    let currentRow: GalleryRowItem[] = []

    while (true) {
        if (mediasCpy.length === 0) {
            if (currentRow.length !== 0) {
                rows.push({
                    rowHeight: baseRowHeight,
                    rowWidth: MAX_ROW_WIDTH,
                    items: currentRow,
                })
            }
            break
        }

        const m = mediasCpy.shift()

        if (!m) {
            break
        }

        if (m.GetHeight() === 0) {
            console.error('Attempt to display media with 0 height:', m.ID())
            continue
        }

        // m.SetAbsIndex(absIndex)
        // absIndex++

        // Calculate width given height "imageBaseScale", keeping aspect ratio
        const newWidth = Math.round((baseRowHeight / m.GetHeight()) * m.GetWidth()) + marginSize

        // If we are out of media, and the image does not overflow this row, add it and break
        if (mediasCpy.length === 0 && !(currentRowWidth + newWidth > MAX_ROW_WIDTH)) {
            currentRow.push({ m: m, w: newWidth })

            rows.push({
                rowHeight: baseRowHeight,
                rowWidth: MAX_ROW_WIDTH,
                items: currentRow,
            })
            break
        }

        // If the image will overflow the window
        else if (currentRowWidth + newWidth > MAX_ROW_WIDTH) {
            const leftover = MAX_ROW_WIDTH - currentRowWidth
            let consuming = false
            if (newWidth / 2 < leftover || currentRow.length === 0) {
                currentRow.push({ m: m, w: newWidth })
                currentRowWidth += newWidth
                consuming = true
            }
            const newRowHeight = (MAX_ROW_WIDTH / currentRowWidth) * baseRowHeight

            currentRow = currentRow.map((v) => {
                v.w = (v.w - marginSize) * (newRowHeight / baseRowHeight) + marginSize
                return v
            })

            rows.push({
                rowHeight: newRowHeight,
                rowWidth: MAX_ROW_WIDTH,
                items: currentRow,
            })
            currentRow = []
            currentRowWidth = 0

            if (consuming) {
                continue
            }
        }
        currentRow.push({ m: m, w: newWidth })
        currentRowWidth += newWidth
    }

    // Add false rows to make scrollbar scale correctly
    const firstRowLength: number = rows[0]?.items.length || 1
    return { rows, remainingGap: baseRowHeight * ((totalMediaCount - medias.length) / firstRowLength) }
}

type MediaType = {
    FriendlyName: string
    FileExtension: string[]
    IsDisplayable: boolean
    IsRaw: boolean
    IsVideo: boolean
    RawThumbExifKey?: string
    SupportsImgRecog?: boolean
    MultiPage?: boolean
}

const mediaTypes: Record<string, MediaType> = {
    generic: {
        FriendlyName: 'File',
        FileExtension: [],
        IsDisplayable: false,
        IsRaw: false,
        IsVideo: false,
        SupportsImgRecog: false,
    },
    'application/zip': {
        FriendlyName: 'Zip',
        FileExtension: ['zip'],
        IsDisplayable: false,
        IsRaw: false,
        IsVideo: false,
    },
    'image/gif': {
        FriendlyName: 'Gif',
        FileExtension: ['gif'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: false,
        SupportsImgRecog: false,
    },
    'image/jpeg': {
        FriendlyName: 'Jpeg',
        FileExtension: ['jpeg', 'jpg', 'JPG'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: false,
        SupportsImgRecog: true,
    },
    'image/png': {
        FriendlyName: 'Png',
        FileExtension: ['png'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: false,
        SupportsImgRecog: true,
    },
    'image/x-nikon-nef': {
        FriendlyName: 'Nikon Raw',
        FileExtension: ['NEF', 'nef'],
        IsDisplayable: true,
        IsRaw: true,
        IsVideo: false,
        RawThumbExifKey: 'JpgFromRaw',
        SupportsImgRecog: true,
    },
    'image/x-sony-arw': {
        FriendlyName: 'Sony ARW',
        FileExtension: ['ARW'],
        IsDisplayable: true,
        IsRaw: true,
        IsVideo: false,
        RawThumbExifKey: 'PreviewImage',
        SupportsImgRecog: true,
    },
    'image/x-canon-cr2': {
        FriendlyName: 'Cannon Raw',
        FileExtension: ['CR2'],
        IsDisplayable: true,
        IsRaw: true,
        IsVideo: false,
        RawThumbExifKey: 'PreviewImage',
        SupportsImgRecog: true,
    },
    'image/heic': {
        FriendlyName: 'HEIC',
        FileExtension: ['HEIC', 'heic'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: false,
        RawThumbExifKey: '',
        SupportsImgRecog: true,
    },
    'image/heif': {
        FriendlyName: 'HEIF',
        FileExtension: ['HEIF', 'heif'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: false,
        RawThumbExifKey: '',
        SupportsImgRecog: true,
    },
    'image/webp': {
        FriendlyName: 'webp',
        FileExtension: ['webp'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: false,
        RawThumbExifKey: '',
        SupportsImgRecog: true,
    },
    'application/pdf': {
        FriendlyName: 'PDF',
        FileExtension: ['pdf'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: false,
        RawThumbExifKey: '',
        SupportsImgRecog: false,
        MultiPage: true,
    },
    'video/mp4': {
        FriendlyName: 'MP4',
        FileExtension: ['MP4', 'mp4', 'MOV'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: true,
        SupportsImgRecog: false,
    },
    'video/quicktime': {
        FriendlyName: 'MP4',
        FileExtension: ['MP4', 'mp4'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: true,
        SupportsImgRecog: false,
    },
    'video/x-matroska': {
        FriendlyName: 'MKV',
        FileExtension: ['MKV', 'mkv'],
        IsDisplayable: true,
        IsRaw: false,
        IsVideo: true,
        SupportsImgRecog: false,
    },
}

export default WeblensMedia
