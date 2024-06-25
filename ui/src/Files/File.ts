import {
    IconFile,
    IconFolder,
    IconHome,
    IconTrash,
    IconUser,
} from '@tabler/icons-react'
import { ShareDataT, WeblensShare } from '../classes/Share'
import { MediaDataT } from '../Media/Media'
import { FBDispatchT } from '../types/Types'
import { humanFileSize } from '../util'
import { SelectedState } from './FileDisplay'
import { FbModeT } from './filesContext'

function getIcon(folderName: string): (p) => JSX.Element {
    if (folderName === 'HOME') {
        return IconHome
    } else if (folderName === 'TRASH') {
        return IconTrash
    } else if (folderName === 'SHARE') {
        return IconUser
    } else {
        return null
    }
}

export interface FileInitT {
    id?: string
    owner?: string
    modTime?: string
    filename?: string
    pathFromHome?: string
    parentFolderId?: string
    fileFriendlyName?: string

    children?: string[]

    isDir?: boolean
    pastFile?: boolean
    imported?: boolean
    modifiable?: boolean
    displayable?: boolean

    size?: number
    mediaData?: MediaDataT
    share?: string
}

interface fileData {
    id?: string
    owner?: string
    filename?: string
    pathFromHome?: string
    parentFolderId?: string
    fileFriendlyName?: string

    modifyDate?: Date

    children?: string[]

    isDir?: boolean
    pastFile?: boolean
    imported?: boolean
    modifiable?: boolean
    displayable?: boolean

    size?: number

    // Non-api props
    parents?: WeblensFile[]
    visible?: boolean
    selected?: boolean
    hovering?: boolean
}

export class WeblensFile {
    private data: fileData
    private mediaId: string
    private share: WeblensShare

    constructor(init: FileInitT) {
        this.data = {
            ...init,
            modifyDate: new Date(init.modTime),
            hovering: false,
        } as fileData

        if (init.mediaData) {
            this.mediaId = init.mediaData.contentId
        }

        if (init.share) {
            this.share = new WeblensShare({ shareId: init.share } as ShareDataT)
        }
    }

    Id(): string {
        return this.data.id
    }

    Update(newInfo: FileInitT) {
        this.data = newInfo

        if (newInfo.mediaData && newInfo.mediaData.contentId !== this.mediaId) {
            this.mediaId = newInfo.mediaData.contentId
        }
    }

    ParentId(): string {
        return this.data.parentFolderId
    }

    SetParents(parents: WeblensFile[]) {
        const index = parents.findIndex((v) => {
            return v.IsTrash()
        })

        if (index !== -1) {
            parents = parents.slice(index)
        }
        this.data.parents = parents
    }

    FormatParents(): WeblensFile[] {
        if (!this.data.parents) {
            return []
        }
        return this.data.parents
    }

    GetPathParts(replaceIcons?: boolean): (string | ((p) => JSX.Element))[] {
        const parts: (string | ((p) => JSX.Element))[] =
            this.data.pathFromHome.split('/')
        if (replaceIcons) {
            const icon = getIcon(String(parts[0]))
            if (icon !== null) {
                parts[0] = icon
            }
        }
        return parts
    }

    IsModifiable(): boolean {
        return this.data.modifiable
    }

    GetFilename(): string {
        if (this.data.pathFromHome === 'HOME') {
            return 'Home'
        }
        if (this.data.filename === '.user_trash') {
            return 'Trash'
        }
        return this.data.filename
    }

    GetModified(): Date {
        if (!this.data.modifyDate) {
            return new Date()
        }
        return this.data.modifyDate
    }

    SetSize(size: number) {
        this.data.size = size
    }

    GetSize(): number {
        return this.data.size
    }

    FormatSize(): string {
        const [value, units] = humanFileSize(this.data.size)
        return value + units
    }

    IsFolder(): boolean {
        return this.data.isDir
    }

    IsImage(): boolean {
        if (this.mediaId) {
            return true
        }
        const ext = this.data.filename.split('.').pop()
        console.log(ext)
        return true
        // if (ext)
    }

    GetMediaId(): string {
        return this.mediaId
    }

    IsTrash(): boolean {
        return this.data.filename === '.user_trash'
    }

    GetOwner(): string {
        return this.data.owner
    }

    SetSelected(): void
    SetSelected(selected: boolean): void

    SetSelected(selected?: boolean): void {
        if (selected === undefined) {
            this.data.selected = !this.data.selected
            return
        }
        this.data.selected = selected
    }

    IsSelected(): boolean {
        return this.data.selected
    }

    IsImported(): boolean {
        return this.data.imported
    }

    IsPastFile(): boolean {
        return this.data.pastFile
    }

    GetChildren(): string[] {
        return this.data.children
    }

    SetHovering(hovering: boolean) {
        this.data.hovering = hovering
    }

    IsHovering(): boolean {
        return this.data.hovering
    }

    GetBaseIcon(mustBeRoot?: boolean): (p) => JSX.Element {
        if (!this.data.pathFromHome) {
            return null
        }
        const parts: any[] = this.data.pathFromHome.split('/')
        if (mustBeRoot && parts.length > 1) {
            return null
        }

        if (parts[0] === 'HOME') {
            return IconHome
        } else if (parts[0] === 'TRASH') {
            return IconTrash
        } else if (parts[0] === 'SHARE') {
            return IconUser
        } else {
            console.error('Unknown filepath base type')
            return null
        }
    }

    GetFileIcon() {
        if (!this.data.pathFromHome) {
            return null
        }

        if (this.data.pathFromHome === 'HOME') {
            return IconHome
        } else if (this.data.pathFromHome === 'TRASH') {
            return IconTrash
        } else if (this.data.pathFromHome === 'SHARE') {
            return IconUser
        } else if (this.data.isDir) {
            return IconFolder
        } else {
            return IconFile
        }
    }

    GetShare(): WeblensShare {
        return this.share
    }

    GetVisitRoute(mode: FbModeT, shareId: string, dispatch: FBDispatchT) {
        if (this.data.isDir) {
            if (mode === FbModeT.share && shareId === '') {
                return `/files/share/${this.share[0].Id()}/${this.data.id}`
            } else if (mode === FbModeT.share) {
                return `/files/share/${shareId}/${this.data.id}`
            } else if (mode === FbModeT.external) {
                return `/files/external/${this.data.id}`
            } else if (mode === FbModeT.default) {
                return `/files/${this.data.id}`
            }
        } else if (this.data.displayable) {
            dispatch({
                type: 'set_presentation',
                presentingId: this.data.id,
            })
            return
        }
        console.error('Did not find location to visit for', this.data.filename)
    }
}

export type FileContextT = {
    file: WeblensFile
    selected: SelectedState
}

export enum FbMenuModeT {
    Closed,
    Default,
    Sharing,
    NameFolder,
    AddToAlbum,
}
