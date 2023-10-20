import { Ref } from "react"

export type MediaData = {
    BlurHash: string
    CreateDate: string
    FileHash: string
    Filepath: string
    MediaType: {
        FileExtension: []
        FriendlyName: string
        IsRaw: boolean
        IsVideo: boolean
    }
    ThumbFilepath: string
    MediaWidth: number
    MediaHeight: number
    ThumbWidth: number
    ThumbHeight: number
    Thumbnail64: string

    // Local things
    Previous: MediaData
    Next: MediaData
    Display: boolean
    ImgRef: React.MutableRefObject<any>
}

export type MediaStateType = {
    mediaMap: Map<string, MediaData>
    mediaCount: number
    maxMediaCount: number
    hasMoreMedia: boolean
    presentingHash: string
    previousLast: string
    includeRaw: boolean
    showIcons: boolean
    loading: boolean
    scanProgress: number
    searchContent: string
}