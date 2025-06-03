import {
    Icon,
    IconFile,
    IconFolder,
    IconHome,
    IconTrash,
    IconUser,
} from '@tabler/icons-react'
import SharesApi from '@weblens/api/SharesApi'
import { FileInfo } from '@weblens/api/swag'
import { useSessionStore } from '@weblens/components/UserInfo'
import { WeblensShare } from '@weblens/types/share/share'
import { humanFileSize } from '@weblens/util'

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

function getIcon(folderName: string): Icon | null {
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

class WeblensFile {
    id: string
    owner?: string
    private filename?: string
    portablePath?: string = ''
    parentId?: string

    modifyDate?: Date

    childrenIds: string[] = []

    isDir?: boolean
    pastFile?: boolean
    hasRestoreMedia?: boolean
    modifiable?: boolean
    displayable?: boolean

    size?: number
    shareId?: string

    // Non-api props
    parents?: WeblensFile[] = []
    hovering?: boolean
    index?: number
    visible?: boolean

    private fetching: boolean = false
    public fromAPI: boolean = false

    private selected: SelectedState
    private contentId: string = ''
    private share: WeblensShare

    constructor(init: FileInfo) {
        this.id = init.id ?? ''

        Object.assign(this, init)
        this.hovering = false
        this.modifyDate = new Date(init.modifyTimestamp ?? 0)
        this.selected = SelectedState.NotSelected
        if (!this.parents) {
            this.parents = []
        }
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

    Update(newInfo: FileInfo) {
        Object.assign(this, newInfo)
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
        if (this.filename === '.user_trash') {
            return []
        }
        return this.parents.filter((parent) => Boolean(parent))
    }

    GetPathParts(replaceIcons?: boolean): (string | Icon)[] {
        const parts: (string | Icon)[] = this.portablePath.split('/')
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
        if (!this.filename) {
            if (!this.portablePath) {
                return ''
            }

            const filename = this.portablePath.slice(
                this.portablePath.indexOf(':') + 1
            )
            const parts = filename.split('/')
            let name = parts.pop()

            // If the path is a directory, the portable path will end with a slash, so we need to pop again
            if (this.isDir) {
                name = parts.pop()
            }

            if (this.parentId === 'ROOT') {
                name = 'Home'
            } else if (name === '.user_trash') {
                name = 'Trash'
            }

            this.filename = name
        }

        if (!this.filename) {
            console.error('Filename is null', this)
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
        return this.id && this.id === useSessionStore.getState()?.user?.trashId
    }

    IsInTrash(): boolean {
        const trashId = useSessionStore.getState()?.user?.trashId
        if (this.id === trashId) {
            return true
        }
        return this.parents.map((parent) => parent?.Id()).includes(trashId)
    }

    GetOwner(): string {
        return this.owner
    }

    SetSelected(selected: SelectedState, override: boolean = false): void {
        if (override) {
            this.selected = selected
            return
        }
        this.selected = this.selected | selected
    }

    UnsetSelected(selected: SelectedState): void {
        let mask = SelectedState.ALL - 1
        while (selected !== SelectedState.NotSelected) {
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

    SetFetching(fetching: boolean): void {
        this.fetching = fetching
    }

    GetFetching(): boolean {
        return this.fetching
    }

    GetChildren(): string[] {
        if (!this.childrenIds) {
            return []
        }
        const trashId = useSessionStore.getState()?.user?.trashId
        return this.childrenIds.filter((child) => {
            return child !== trashId
        })
    }

    IsHovering(): boolean {
        return (this.selected & SelectedState.Hovering) !== 0
    }

    GetBaseIcon(mustBeRoot?: boolean): Icon {
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

    public async GetShare(refetch?: boolean): Promise<WeblensShare> {
        if (this.share && !refetch) {
            return this.share
        } else if (!this.shareId) {
            return new WeblensShare({ fileId: this.id, owner: this.owner })
        }

        const res = await SharesApi.getFileShare(this.shareId)
        if (res.status !== 200) {
            return Promise.reject(new Error('Failed to get share info'))
        }

        this.share = new WeblensShare(res.data)
        return this.share
    }

    public static Home(): WeblensFile {
        const user = useSessionStore.getState().user

        return new WeblensFile({
            id: user.homeId,
            owner: user.username,
            portablePath: `USERS:${user.username}/`,
            isDir: true,
            modifiable: true,
            pastFile: false,
        })
    }
}

export default WeblensFile
