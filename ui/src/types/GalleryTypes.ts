import { MediaData } from "./Generic"

export type GalleryBucketProps = {
    date: string
    bucketData: []
    showIcons: boolean
    dispatch: React.Dispatch<any>
}

export type MediaWrapperProps = {
    mediaData: MediaData
    showIcons: boolean
    dispatch
}