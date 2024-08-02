import {
    IconFile,
    IconFolder,
    IconHome,
    IconTrash,
    IconUser,
} from '@tabler/icons-react'
import { ShareInfo, WeblensShare } from '../Share/Share'
import { MediaDataT } from '../Media/Media'
import { AuthHeaderT } from '../types/Types'
import { humanFileSize } from '../util'
import API_ENDPOINT from '../api/ApiEndpoint'
import { FbModeT } from '../Pages/FileBrowser/FBStateControl'

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

export interface WeblensFileParams {
    id?: string
    owner?: string
    modTime?: string
    filename?: string
    pathFromHome?: string
    parentFolderId?: string

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
    pathFromHome?: string
    parentFolderId?: string

    modifyDate?: Date

    children?: string[]

    isDir?: boolean
    pastFile?: boolean
    imported?: boolean
    modifiable?: boolean
    displayable?: boolean

    size?: number
    shareId?: string

    // Non-api props
    parents?: WeblensFile[]
    visible?: boolean
    private selected: SelectedState
    hovering?: boolean
    index?: number

    private mediaId: string
    private share: WeblensShare

    constructor(init: WeblensFileParams) {
        if (!init.id) {
            throw new Error('trying to construct WeblensFile with no id')
        }
        Object.assign(this, init)
        // this.share = undefined;
        this.hovering = false
        this.modifyDate = new Date(init.modTime)

        if (init.mediaData) {
            this.mediaId = init.mediaData.contentId
        }

        this.shareId = init.shareId
        this.selected = SelectedState.NotSelected
        // if (init.shareId) {
        //     new WeblensShare({ id: init.shareId });
        // }
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

        if (newInfo.mediaData && newInfo.mediaData.contentId !== this.mediaId) {
            this.mediaId = newInfo.mediaData.contentId
        }
    }

    ParentId(): string {
        return this.parentFolderId
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
            this.pathFromHome.split('/')
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
        if (this.pathFromHome === 'HOME') {
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

    GetMediaId(): string {
        return this.mediaId
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

    IsImported(): boolean {
        return this.imported
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
        if (!this.pathFromHome) {
            return null
        }
        const parts = this.pathFromHome.split('/')
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
        if (!this.pathFromHome) {
            return null
        }

        if (this.pathFromHome === 'HOME') {
            return IconHome
        } else if (this.pathFromHome === 'TRASH') {
            return IconTrash
        } else if (this.pathFromHome === 'SHARE') {
            return IconUser
        } else if (this.isDir) {
            return IconFolder
        } else {
            return IconFile
        }
    }

    public async LoadShare(authHeader: AuthHeaderT) {
        if (this.share) {
            return this.share
        } else if (!this.shareId) {
            return null
        }

        const url = new URL(`${API_ENDPOINT}/file/share/${this.shareId}`)
        return fetch(url.toString(), {
            headers: authHeader,
        })
            .then((r) => {
                if (r.status === 200) {
                    return r.json()
                } else {
                    return Promise.reject('Bad response: ' + r.statusText)
                }
            })
            .then((j) => {
                this.share = new WeblensShare(j as ShareInfo)
                return this.share
            })
            .catch((e: Error) => {
                console.error('Failed to load share:', e)
            })
    }

    GetShare(): WeblensShare {
        return this.share
    }

    GetVisitRoute(
        mode: FbModeT,
        shareId: string,
        setPresentation: (presentationId: string) => void
    ) {
        if (this.isDir) {
            if (mode === FbModeT.share && shareId === '') {
                return `/files/share/${this.shareId}/${this.id}`
            } else if (mode === FbModeT.share) {
                return `/files/share/${shareId}/${this.id}`
            } else if (mode === FbModeT.external) {
                return `/files/external/${this.id}`
            } else if (mode === FbModeT.default) {
                return `/files/${this.id}`
            }
        } else if (this.displayable) {
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
}
