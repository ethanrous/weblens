import { AlbumInfo } from '@weblens/api/swag'
import WeblensMedia from '@weblens/types/media/Media'

// Gallery Types
export type MediaWrapperProps = {
    mediaData: WeblensMedia
    scale: number
    width: number
    showMedia: boolean
    rowIndex?: number
    colIndex?: number
    hoverIndex?: { row: number; col: number }
    albumId?: string
    fetchAlbum?: () => void
    menu?: (
        mediaId: string,
        open: boolean,
        setOpen: (open: boolean) => void
    ) => JSX.Element
}

export enum PresentType {
    None = 1,
    InLine,
    Fullscreen,
}

export type TimeOffset = {
    second: 0
    minute: 0
    hour: 0
    day: 0
    month: 0
    year: 0
}

export const newTimeOffset = (): TimeOffset => {
    return {
        second: 0,
        minute: 0,
        hour: 0,
        day: 0,
        month: 0,
        year: 0,
    }
}

export type GalleryStateT = {
    albumsMap: Map<string, AlbumInfo>
    albumsFilter: string[]
    loading: string[]
    newAlbumDialogue: boolean
    blockSearchFocus: boolean
    selecting: boolean
    menuTargetId: string
    imageSize: number
    searchContent: string
    presentingMediaId: string
    presentingMode: PresentType
    timeAdjustOffset: TimeOffset
    hoverIndex: number
    lastSelId: string
    holdingShift: boolean
    albumId: string
}

export type Coordinates = {
    x: number
    y: number
}

export type Dimensions = {
    height: number
    width: number
}

export function ErrorHandler(err: Error) {
    console.error('Caught:', typeof err, err)
}
