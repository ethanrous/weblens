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
    Display: boolean
    ImgRef: Ref<any>
}