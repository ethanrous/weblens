import { useMessagesController } from '@weblens/store/MessagesController'
import WeblensMedia from '@weblens/types/media/Media'
import { AxiosError } from 'axios'

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
    ) => Element
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

export type Coordinates = {
    x: number
    y: number
}

export function validateCoordinates(coordinates: Coordinates): boolean {
    return coordinates.x >= 0 && coordinates.y >= 0
}

export type Dimensions = {
    height: number
    width: number
}

export function ErrorHandler(err: Error, note?: string) {
    note = note ?? ''
    let errMsg = err.message ?? new Error('Unknown error')
    if (err instanceof AxiosError) {
        errMsg = String(err.response.data.error)
    } else {
        errMsg = err.message
    }

    console.error('ErrorHandler caught', errMsg, err.stack)
    useMessagesController.getState().addMessage({
        title: note ?? 'ErrorHandler caught an error',
        text: errMsg,
        duration: 5000,
        severity: 'error',
    })
}
