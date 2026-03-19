import { useUserStore } from '~/stores/user'
import { humanBytes } from '~/util/humanBytes'
import WeblensShare from '~/types/weblensShare'
import { useWeblensAPI } from '~/api/AllApi'
import type { FileActionInfo, FileInfo, PermissionsInfo } from '@ethanrous/weblens-api'
import useLocationStore from '~/stores/location'
import { PortablePath } from './portablePath'

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

class WeblensFile implements FileInfo {
    id: string
    owner: string = ''
    private filename: string = ''

    portablePath: string = ''
    private _filepath: PortablePath = PortablePath.empty()

    parentID: string = ''

    modifyDate?: Date
    contentCreationDate?: Date

    childrenIds: string[] = []

    isDir?: boolean
    pastFile: boolean = false
    hasRestoreData?: boolean
    modifiable: boolean = false
    displayable?: boolean

    hasMedia: boolean = false

    size: number = -1
    shareID?: string

    // Non-api props
    parents: WeblensFile[] = []
    hovering?: boolean
    index: number = -1
    visible: boolean = true
    rewindTimestamp: number = 0

    permissions?: PermissionsInfo

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

        if (this.portablePath) {
            this._filepath = new PortablePath(this.portablePath)
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

    IsModifiable(): boolean {
        return this.modifiable
    }

    GetFilename(): string {
        if (!this.filename) {
            if (!this.portablePath) {
                return ''
            }

            const root = this.portablePath.split(':')[0]
            const filename = this.portablePath.slice(this.portablePath.indexOf(':') + 1)
            const parts = filename.split('/')
            let name = parts.pop()

            // If the path is a directory, the portable path will end with a slash, so we need to pop again
            if (this.isDir) {
                name = parts.pop()
            }

            if (root === 'SHARED' && parts.length === 0) {
                name = 'Shared'
            } else if (this.parentID === 'USERS' || parts.length === 0) {
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

    GetFilepath(): PortablePath {
        return this._filepath
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
        return WeblensFile.IsShareRoot(this.portablePath)
    }

    public static IsShareRoot(portablePath: string): boolean {
        return portablePath === 'SHARED:'
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

    IsSharedWithMe(): boolean {
        return this.owner !== useUserStore().user.username
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

    public CanDelete(): boolean {
        if (!this.modifiable) {
            return false
        }

        if (this.IsHome() || this.IsTrash() || this.IsShareRoot()) {
            return false
        }

        return this.permissions?.canDelete === true
    }

    public CanEdit(): boolean {
        return this.permissions?.canEdit === true
    }

    public CanDownload(): boolean {
        return this.permissions?.canDownload === true
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

    public URLID(): string {
        if (this.id === useUserStore().user.homeID) {
            return 'home'
        }

        if (WeblensFile.IsShareRoot(this.id)) {
            return 'share'
        }

        if (this.id === useUserStore().user.trashID) {
            return 'trash'
        }

        if (!this.isDir) {
            return this.parentID
        }

        return this.id
    }

    public URLHash(): string | undefined {
        if (!this.isDir) {
            return '#file-' + this.id
        }

        return undefined
    }

    public MediaUrl(): string {
        if (!this.displayable || !this.contentID) {
            return ''
        }

        return `${window.location.origin}/media/${this.contentID}`
    }

    public FileURL(opts?: { forcePresent?: boolean }): string {
        const locationStore = useLocationStore()
        let path = '/files/' + this.URLID()

        if (!this.IsHome() && locationStore.isInShare) {
            if (this.IsShareRoot()) {
                path = `/files/share`
            } else if ((this.shareID ? this.shareID : locationStore.activeShareID) === undefined) {
                console.error('No active share ID to navigate to shared file')

                return ''
            } else {
                path = `/files/share/${this.shareID ? this.shareID : locationStore.activeShareID}/${this.URLID()}`
            }
        }

        console.debug('Navigating to', path, 'with hash', this.URLHash())

        // If this file is a past file, we want to include the rewind timestamp in the query params so that the file browser can rewind to the correct time
        const tsString = opts?.forcePresent
            ? null
            : this.rewindTimestamp > 0
              ? new Date(this.rewindTimestamp).toISOString()
              : undefined

        return `${path}${this.URLHash() ?? ''}${tsString ? '?rewindTo=' + encodeURIComponent(tsString) : ''}`
    }

    public async GoTo(opts?: { replace?: boolean; newTab?: boolean }): Promise<void> {
        if (!this.id) {
            console.error('File has no ID, refusing to navigate', this)

            return
        }

        const navPath = this.FileURL()

        await navigateTo(navPath, {
            replace: opts?.replace ?? false,
            open: opts?.newTab
                ? {
                      target: '_blank',
                  }
                : undefined,
        })
    }

    public static Home(): WeblensFile {
        const user = useUserStore().user
        if (!user.homeID) {
            throw new Error('User has no home ID, cannot create home file')
        }

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
            id: 'share',
            owner: 'WEBLENS',
            portablePath: `SHARED:`,
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
        const newF = new WeblensFile({
            id: action.fileID,
            portablePath: action.filepath,
            parentID: action.liveParentID,
            isDir: action.filepath?.endsWith('/') ?? false,
        })

        newF.pastFile = true
        newF.rewindTimestamp = action.timestamp

        return newF
    }
}

export default WeblensFile
