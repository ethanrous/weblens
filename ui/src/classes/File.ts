import { IconHome, IconTrash, IconUser } from '@tabler/icons-react'
import { shareData } from '../types/Types'
import WeblensMedia, { MediaDataT } from './Media'
import { humanFileSize } from '../util'

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

    shares?: shareData[]
    mediaData?: MediaDataT
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

    shares?: shareData[]

    // Non-api props
    parents?: WeblensFile[]
    visible?: boolean
    selected?: boolean
    hovering?: boolean
}

export class WeblensFile {
    private data: fileData
    private media: WeblensMedia

    constructor(init: FileInitT) {
        this.data = {
            ...init,
            modifyDate: new Date(init.modTime),
            hovering: false,
        } as fileData

        if (init.mediaData) {
            this.media = new WeblensMedia(init.mediaData)
        }
    }

    Id(): string {
        return this.data.id
    }

    Update(newInfo: FileInitT) {
        this.data = newInfo

        if (
            newInfo.mediaData &&
            newInfo.mediaData.mediaId !== this.media?.Id()
        ) {
            this.media = new WeblensMedia(newInfo.mediaData)
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
        if (this.media) {
            return true
        }
        const ext = this.data.filename.split('.').pop()
        console.log(ext)
        return true
        // if (ext)
    }

    GetMedia(): WeblensMedia {
        return this.media
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

    GetShares(): shareData[] {
        if (!this.data.shares) {
            return []
        }

        return this.data.shares
    }
}
