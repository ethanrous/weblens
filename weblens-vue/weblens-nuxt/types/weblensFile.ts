import { IconFile, IconFolder, IconHome, IconTrash, IconUser, type Icon } from '@tabler/icons-vue'

import { useUserStore } from '~/stores/user'
import { humanBytes } from '~/util/humanBytes'
import WeblensShare from '~/types/weblensShare'
import { useWeblensAPI } from '~/api/AllApi'
import type { FileActionInfo, FileInfo } from '@ethanrous/weblens-api'

export class SelectedState {
    public static NotSelected = new SelectedState(0b0)
    public static Hovering = new SelectedState(0b1)
    public static InRange = new SelectedState(0b10)
    public static Selected = new SelectedState(0b100)
    public static LastSelected = new SelectedState(0b1000)
    public static Droppable = new SelectedState(0b10000)
    public static Moved = new SelectedState(0b100000)

    public static ALL = new SelectedState(0b111111)

    private selected: number

    constructor(selected: number | SelectedState = SelectedState.NotSelected) {
        if (selected instanceof SelectedState) {
            this.selected = selected.selected
        } else {
            this.selected = selected
        }
    }

    public Is(selected: SelectedState): boolean {
        return this.selected === selected.selected
    }

    public Has(...selected: SelectedState[]): boolean {
        return selected.every((s) => (this.selected & s.selected) !== 0)
    }

    public Any(...selected: SelectedState[]): boolean {
        return selected.find((s) => (this.selected & s.selected) !== 0) !== undefined
    }

    public Add(selected: SelectedState): SelectedState {
        return new SelectedState(this.selected | selected.selected)
    }

    public Remove(selected: SelectedState): SelectedState {
        let mask = SelectedState.ALL.selected - 1
        while (selected.selected !== SelectedState.NotSelected.selected) {
            selected.selected = selected.selected >> 1
            mask = (mask << 1) + 1
        }
        mask = mask >> 1
        return new SelectedState(this.selected & mask)
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

class WeblensFile implements FileInfo {
    id: string
    owner: string = ''
    private filename: string = ''
    portablePath: string = ''
    parentID: string = ''

    modifyDate?: Date
    contentCreationDate?: Date

    childrenIds: string[] = []

    isDir?: boolean
    pastFile: boolean = false
    hasRestoreMedia?: boolean
    modifiable: boolean = false
    displayable?: boolean

    size: number = -1
    shareID?: string

    // Non-api props
    parents: WeblensFile[] = []
    hovering?: boolean
    index: number = -1
    visible?: boolean

    private fetching: boolean = false
    public fromAPI: boolean = false

    private selected: SelectedState
    private share: WeblensShare

    contentID: string = ''

    constructor(init: FileInfo) {
        this.id = init.id ?? ''

        Object.assign(this, init)
        this.hovering = false
        this.modifyDate = new Date(init.modifyTimestamp ?? 0)
        this.selected = SelectedState.NotSelected
        if (!this.parents) {
            this.parents = []
        }
        this.share = new WeblensShare({ fileID: this.id, owner: this.owner })

        if (!this.filename) {
            this.GetFilename()
        }
    }

    ID(): string {
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

    ParentID(): string {
        return this.parentID
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

            const filename = this.portablePath.slice(this.portablePath.indexOf(':') + 1)
            const parts = filename.split('/')
            let name = parts.pop()

            // If the path is a directory, the portable path will end with a slash, so we need to pop again
            if (this.isDir) {
                name = parts.pop()
            }

            if (this.parentID === 'USERS' || parts.length === 0) {
                name = 'Home'
                if (!this.id) {
                    this.id = useUserStore().user.homeID
                }
            } else if (name === '.user_trash') {
                name = 'Trash'
            }

            if (!name) {
                console.error('Filename could not be found')
            } else {
                this.filename = name
            }
        }

        if (!this.filename) {
            console.error('Filename is null', this)
        }

        return this.filename
    }

    GetModified(): Date {
        if (this.contentCreationDate) {
            return this.contentCreationDate
        }

        if (this.modifyDate) {
            return this.modifyDate
        }

        return new Date()
    }

    FormatModified(): string {
        return this.GetModified().toDateString()
    }

    SetSize(size: number) {
        this.size = size
    }

    GetSize(): number {
        return this.size
    }

    FormatSize(): string {
        const [value, units] = humanBytes(this.size)
        return value + units
    }

    IsFolder(): boolean {
        return this.isDir === true
    }

    GetContentID(): string {
        return this.contentID
    }

    public IsHome(): boolean {
        return WeblensFile.IsHome(this.id)
    }

    public static IsHome(id: string): boolean {
        return id === useUserStore().user.homeID
    }

    public IsShareRoot(): boolean {
        return WeblensFile.IsShareRoot(this.id)
    }

    public static IsShareRoot(id: string): boolean {
        return id === 'SHARE'
    }

    public IsTrash(): boolean {
        return WeblensFile.IsTrash(this.id)
    }

    public static IsTrash(id: string): boolean {
        return id === useUserStore().user.trashID
    }

    IsInTrash(): boolean {
        const trashID = useUserStore().user.trashID

        if (this.id === trashID) {
            return true
        }
        return this.parents.map((parent) => parent?.ID()).includes(trashID)
    }

    SetSelected(selected: SelectedState, override: boolean = false): void {
        if (override) {
            this.selected = selected
            return
        }
        this.selected = this.selected.Add(selected)
    }

    UnsetSelected(selected: SelectedState): void {
        this.selected = this.selected.Remove(selected)
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
        const trashID = useUserStore().user.trashID

        return this.childrenIds.filter((child) => {
            return child !== trashID
        })
    }

    IsHovering(): boolean {
        return this.selected.Has(SelectedState.Hovering)
    }

    GetBaseIcon(mustBeRoot?: boolean): Icon | undefined {
        if (!this.portablePath) {
            return
        }
        const parts = this.portablePath.split('/')
        if (mustBeRoot && parts.length > 1) {
            return
        }

        if (parts[0] === 'HOME') {
            return IconHome
        } else if (parts[0] === 'TRASH') {
            return IconTrash
        } else if (parts[0] === 'SHARE') {
            return IconUser
        } else {
            console.error('Unknown filepath base type')
            return
        }
    }

    GetFileIcon(): Icon | undefined {
        if (!this.portablePath) {
            return
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
        if (this.shareID && this.shareID !== share.ID()) {
            console.error('Trying to set share with mismatched id, expected', this.shareID, 'but got', share.ID())
            return
        } else if (!this.shareID) {
            this.shareID = share.ID()
        }
        this.share = share
    }

    public async GetShare(refetch?: boolean): Promise<WeblensShare> {
        if (this.share.shareID && !this.shareID) {
            this.shareID = this.share.shareID
        }

        if (this.share.shareID && !refetch) {
            return this.share
        } else if (!this.shareID) {
            return this.share
        }

        const res = await useWeblensAPI().SharesAPI.getFileShare(this.shareID)
        if (res.status !== 200) {
            return Promise.reject(new Error('Failed to get share info'))
        }

        this.share = new WeblensShare(res.data)
        return this.share
    }

    public UrlID(): string {
        if (this.id === useUserStore().user.homeID) {
            return 'home'
        }

        if (WeblensFile.IsShareRoot(this.id)) {
            return 'share'
        }

        if (this.id === useUserStore().user.trashID) {
            return 'trash'
        }

        return this.id
    }

    public MediaUrl(): string {
        if (!this.displayable || !this.contentID) {
            return ''
        }

        return `${window.location.origin}/media/${this.contentID}`
    }

    public async GoTo(replace: boolean = false): Promise<void> {
        await navigateTo(
            {
                path: '/files/' + this.UrlID(),
            },
            { replace: replace },
        )
    }

    public static Home(): WeblensFile {
        const user = useUserStore().user

        return new WeblensFile({
            id: user.homeID,
            owner: user.username,
            portablePath: `USERS:${user.username}/`,
            isDir: true,
            modifiable: true,
            pastFile: false,
        })
    }

    public static ShareRoot(): WeblensFile {
        return new WeblensFile({
            id: 'SHARE',
            owner: 'WEBLENS',
            portablePath: `SHARED:Shared/`,
            isDir: true,
            modifiable: false,
            pastFile: false,
        })
    }

    public static Trash(): WeblensFile {
        const user = useUserStore().user

        return new WeblensFile({
            id: user.trashID,
            owner: user.username,
            portablePath: `USERS:${user.username}/.user_trash/`,
            isDir: true,
            modifiable: false,
            pastFile: false,
        })
    }

    public static FromAction(action: FileActionInfo): WeblensFile {
        return new WeblensFile({
            id: action.fileID,
            portablePath: action.filepath,
            parentID: action.parentID,
            isDir: action.filepath?.endsWith('/'),
        })
    }
}

export default WeblensFile
