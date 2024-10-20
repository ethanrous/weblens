import {
    IconFile,
    IconFolder,
    IconHome,
    IconTrash,
    IconUser,
} from '@tabler/icons-react'
import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import { fetchJson } from '@weblens/api/ApiFetch'
import {
    FbModeT,
    useFileBrowserStore,
} from '@weblens/pages/FileBrowser/FBStateControl'
import { MediaDataT } from '@weblens/types/media/Media'
import { WeblensShare } from '@weblens/types/share/share'
import { humanFileSize } from '@weblens/util'

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

export interface WeblensFileInfo {
    filename: string
    id: string
    isDir: boolean
    modifyTimestamp: number
    ownerName: string
    parentId: string
    portablePath: string
    shareId: string
    size: number
}

export interface WeblensFileParams {
    id?: string
    owner?: string
    modifyTimestamp?: string
    filename?: string
    portablePath?: string
    parentId?: string
    contentId?: string

    children?: string[]

    isDir?: boolean
    pastFile?: boolean
    imported?: boolean
    modifiable?: boolean
    displayable?: boolean

    size?: number
    mediaData?: MediaDataT
    shareId?: string
}

export class WeblensFile {
    id?: string
    owner?: string
    filename?: string
    portablePath?: string
    parentId?: string

    modifyDate?: Date

    children?: string[]

    isDir?: boolean
    pastFile?: boolean
    modifiable?: boolean
    displayable?: boolean

    size?: number
    shareId?: string

    // Non-api props
    parents?: WeblensFile[]
    hovering?: boolean
    index?: number
    visible?: boolean

    private selected: SelectedState
    private contentId: string
    private share: WeblensShare

    constructor(init: WeblensFileParams) {
        if (!init || !init.id) {
            throw new Error('trying to construct WeblensFile with no id')
        }
        Object.assign(this, init)
        this.hovering = false
        this.modifyDate = new Date(init.modifyTimestamp)
        this.shareId = init.shareId
        this.selected = SelectedState.NotSelected
    }

    Id(): string {
        return this.id
    }

    SetIndex(index: number): void {
        this.index = index
    }

    GetIndex(): number {
        return this.index
    }

    Update(newInfo: WeblensFileParams) {
        Object.assign(this, newInfo)
        // this.share = undefined;

        if (
            newInfo.mediaData &&
            newInfo.mediaData.contentId !== this.contentId
        ) {
            this.contentId = newInfo.mediaData.contentId
        }
    }

    ParentId(): string {
        return this.parentId
    }

    SetParents(parents: WeblensFile[]) {
        const index = parents.findIndex((v) => {
            return v.IsTrash()
        })

        if (index !== -1) {
            parents = parents.slice(index)
        }
        this.parents = parents
    }

    FormatParents(): WeblensFile[] {
        if (!this.parents) {
            return []
        }
        return this.parents
    }

    GetPathParts(replaceIcons?: boolean): (string | ((p) => JSX.Element))[] {
        const parts: (string | ((p) => JSX.Element))[] =
            this.portablePath.split('/')
        if (replaceIcons) {
            const icon = getIcon(String(parts[0]))
            if (icon !== null) {
                parts[0] = icon
            }
        }
        return parts
    }

    IsModifiable(): boolean {
        return this.modifiable
    }

    GetFilename(): string {
        if (this.portablePath === 'HOME') {
            return 'Home'
        }
        if (this.filename === '.user_trash') {
            return 'Trash'
        }
        return this.filename
    }

    GetModified(): Date {
        if (!this.modifyDate) {
            return new Date()
        }
        return this.modifyDate
    }

    SetSize(size: number) {
        this.size = size
    }

    GetSize(): number {
        return this.size
    }

    FormatSize(): string {
        const [value, units] = humanFileSize(this.size)
        return value + units
    }

    IsFolder(): boolean {
        return this.isDir
    }

    GetContentId(): string {
        return this.contentId
    }

    IsTrash(): boolean {
        return this.filename === '.user_trash'
    }

    GetOwner(): string {
        return this.owner
    }

    SetSelected(selected: SelectedState): void {
        this.selected = this.selected | selected
    }

    UnsetSelected(selected: SelectedState): void {
        // console.trace('Unset selected', selected)
        let mask = SelectedState.ALL - 1
        while (selected !== 0) {
            selected = selected >> 1
            mask = (mask << 1) + 1
        }
        mask = mask >> 1
        this.selected = this.selected & mask
    }

    GetSelectedState(): SelectedState {
        return this.selected
    }

    IsPastFile(): boolean {
        return this.pastFile
    }

    GetChildren(): string[] {
        return this.children
    }

    IsHovering(): boolean {
        return (this.selected & SelectedState.Hovering) !== 0
    }

    GetBaseIcon(mustBeRoot?: boolean): (p) => JSX.Element {
        if (!this.portablePath) {
            return null
        }
        const parts = this.portablePath.split('/')
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
        if (!this.portablePath) {
            return null
        }

        if (this.portablePath === 'HOME') {
            return IconHome
        } else if (this.portablePath === 'TRASH') {
            return IconTrash
        } else if (this.portablePath === 'SHARE') {
            return IconUser
        } else if (this.isDir) {
            return IconFolder
        } else {
            return IconFile
        }
    }

    public SetShare(share: WeblensShare) {
        if (this.shareId && this.shareId !== share.Id()) {
            console.error(
                'Trying to set share with mismatched id, expected',
                this.shareId,
                'but got',
                share.Id()
            )
            return
        } else if (!this.shareId) {
            this.shareId = share.Id()
        }
        this.share = share
    }

    public async GetShare(): Promise<WeblensShare> {
        if (this.share) {
            return this.share
        } else if (!this.shareId) {
            return null
        }

        const url = `${API_ENDPOINT}/file/share/${this.shareId}`
        return fetchJson(url)
    }

    GetVisitRoute(
        mode: FbModeT,
        shareId: string,
        setPresentation: (presentationId: string) => void
    ) {
        let timestampQuery = ''
        const pastTime = useFileBrowserStore.getState().pastTime
        if (pastTime) {
            timestampQuery = `?at=${pastTime.getTime().toString()}`
        }

        if (this.isDir) {
            if (mode === FbModeT.share && shareId === '') {
                return `/files/share/${this.shareId}/${this.id}`
            } else if (mode === FbModeT.share) {
                return `/files/share/${shareId}/${this.id}`
            } else if (mode === FbModeT.external) {
                return `/files/external/${this.id}`
            } else if (mode === FbModeT.default) {
                return `/files/${this.id}${timestampQuery}`
            }
        } else if (this.displayable || !this.displayable) {
            setPresentation(this.id)
            return
        }
        console.error('Did not find location to visit for', this.filename)
    }
}

export enum SelectedState {
    NotSelected = 0b0,
    Hovering = 0b1,
    InRange = 0b10,
    Selected = 0b100,
    LastSelected = 0b1000,
    Droppable = 0b10000,
    Moved = 0b100000,

    ALL = 0b111111,
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
    RenameFile,
    SearchForFile,
}
