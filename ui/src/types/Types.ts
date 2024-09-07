import { TasksProgressAction } from '@weblens/pages/FileBrowser/TaskProgress'
import { GalleryAction } from '@weblens/pages/Gallery/GalleryLogic'
import WeblensMedia from '@weblens/types/media/Media'
import { Dispatch } from 'react'

export type AuthHeaderT = {
    Authorization: string
}

export const LOGIN_TOKEN_COOKIE_KEY = 'weblens-login-token'
export const USERNAME_COOKIE_KEY = 'weblens-username'

export type UserInfoT = {
    homeId: string
    trashId: string
    username: string
    admin: boolean
    owner: boolean
    activated: boolean
    isLoggedIn: boolean
}

export type ServerInfoT = {
    id: string
    name: string
    role: string
    coreAddress: string
}

export type mediaType = {
    FileExtension: []
    FriendlyName: string
    IsRaw: boolean
    IsVideo: boolean
    IsDisplayable: boolean
}

export type AlbumData = {
    id: string
    medias: string[]
    name: string
    cover: string
    primaryColor: string
    secondaryColor: string
    owner: string
    showOnTimeline: boolean
}

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
    albumsMap: Map<string, AlbumData>
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

// File Browser Types
export type TPDispatchT = Dispatch<TasksProgressAction>
export type GalleryDispatchT = (action: GalleryAction) => void

export type ScanMeta = {
    taskId: string
    taskType: string
    target: string
    mostRecent: string
    note: string

    progress: number
    tasksComplete: number
    tasksTotal: number
    time: number

    complete: boolean
}

export interface FbViewOptsT {
    dirViewMode: string
    sortDirection: number // 1 or -1
    sortFunc: string
}

export type SizeT = {
    height: number
    width: number
}

export type ApiKeyInfo = {
    id: string
    key: string
    remoteUsing: string
}
