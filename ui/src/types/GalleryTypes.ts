import { MediaData } from "./Generic"

export type GalleryBucketProps = {
    date: string
    bucketData: []
    presentingHash: string
    showIcons: boolean
    dispatch: React.Dispatch<any>
}

export type MediaWrapperProps = {
    mediaData: MediaData
    showIcons: boolean
    presentingHash: string
    dispatch
}