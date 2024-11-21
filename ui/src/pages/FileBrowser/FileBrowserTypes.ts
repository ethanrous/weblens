import { Dispatch } from 'react'

import { GalleryAction } from '../Gallery/GalleryLogic'
import { TasksProgressAction } from './TaskStateControl'

export type TPDispatchT = Dispatch<TasksProgressAction>
export type GalleryDispatchT = (action: GalleryAction) => void

export interface FbViewOptsT {
    dirViewMode: DirViewModeT
    sortDirection: number // 1 or -1
    sortFunc: string
}

export type FileEventT = {
    Action: string
    // At: number;
    FileId: string
    Path: string
    FromFileId: string
    FromPath: string
    // Size: number
    // SnapshotId: string;
    Timestamp: string

    millis: number

    // Non-api fields
    fileName: string
    // oldPath: string
}

// export type FileAction = {
//     actionType: string
//     destinationId: string
//     destinationPath: string
//     eventId: string
//     lifeId: string
//     originId: string
//     originPath: string
//     timestamp: number
// }

export enum DirViewModeT {
    Grid,
    List,
    Columns,
}
