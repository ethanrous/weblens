import { IconFile, IconFolder, IconHome, IconTrash, IconUser } from '@tabler/icons-react';
import { ShareDataT, WeblensShare } from '../Share/Share';
import { MediaDataT } from '../Media/Media';
import { FBDispatchT } from '../types/Types';
import { humanFileSize } from '../util';
import { DraggingStateT, FbModeT } from './filesContext';
import API_ENDPOINT from '../api/ApiEndpoint';

function getIcon(folderName: string): (p) => JSX.Element {
    if (folderName === 'HOME') {
        return IconHome;
    } else if (folderName === 'TRASH') {
        return IconTrash;
    } else if (folderName === 'SHARE') {
        return IconUser;
    } else {
        return null;
    }
}

export interface WeblensFileInfo {
    id?: string;
    owner?: string;
    modTime?: string;
    filename?: string;
    pathFromHome?: string;
    parentFolderId?: string;
    fileFriendlyName?: string;

    children?: string[];

    isDir?: boolean;
    pastFile?: boolean;
    imported?: boolean;
    modifiable?: boolean;
    displayable?: boolean;

    size?: number;
    mediaData?: MediaDataT;
    shareId?: string;
}

export class WeblensFile {
    id?: string;
    owner?: string;
    filename?: string;
    pathFromHome?: string;
    parentFolderId?: string;
    fileFriendlyName?: string;

    modifyDate?: Date;

    children?: string[];

    isDir?: boolean;
    pastFile?: boolean;
    imported?: boolean;
    modifiable?: boolean;
    displayable?: boolean;

    size?: number;
    shareId?: string;

    // Non-api props
    parents?: WeblensFile[];
    visible?: boolean;
    selected?: boolean;
    hovering?: boolean;

    private mediaId: string;
    private share: WeblensShare;

    constructor(init: WeblensFileInfo) {
        Object.assign(this, init);
        this.hovering = false;
        this.modifyDate = new Date(init.modTime);

        if (init.mediaData) {
            this.mediaId = init.mediaData.contentId;
        }

        this.shareId = init.shareId;
        if (init.shareId) {
        }
    }

    Id(): string {
        return this.id;
    }

    Update(newInfo: WeblensFileInfo) {
        Object.assign(this, newInfo);

        if (newInfo.mediaData && newInfo.mediaData.contentId !== this.mediaId) {
            this.mediaId = newInfo.mediaData.contentId;
        }
    }

    ParentId(): string {
        return this.parentFolderId;
    }

    SetParents(parents: WeblensFile[]) {
        const index = parents.findIndex(v => {
            return v.IsTrash();
        });

        if (index !== -1) {
            parents = parents.slice(index);
        }
        this.parents = parents;
    }

    FormatParents(): WeblensFile[] {
        if (!this.parents) {
            return [];
        }
        return this.parents;
    }

    GetPathParts(replaceIcons?: boolean): (string | ((p) => JSX.Element))[] {
        const parts: (string | ((p) => JSX.Element))[] = this.pathFromHome.split('/');
        if (replaceIcons) {
            const icon = getIcon(String(parts[0]));
            if (icon !== null) {
                parts[0] = icon;
            }
        }
        return parts;
    }

    IsModifiable(): boolean {
        return this.modifiable;
    }

    GetFilename(): string {
        if (this.pathFromHome === 'HOME') {
            return 'Home';
        }
        if (this.filename === '.user_trash') {
            return 'Trash';
        }
        return this.filename;
    }

    GetModified(): Date {
        if (!this.modifyDate) {
            return new Date();
        }
        return this.modifyDate;
    }

    SetSize(size: number) {
        this.size = size;
    }

    GetSize(): number {
        return this.size;
    }

    FormatSize(): string {
        const [value, units] = humanFileSize(this.size);
        return value + units;
    }

    IsFolder(): boolean {
        return this.isDir;
    }

    IsImage(): boolean {
        if (this.mediaId) {
            return true;
        }
        const ext = this.filename.split('.').pop();
        console.log(ext);
        return true;
        // if (ext)
    }

    GetMediaId(): string {
        return this.mediaId;
    }

    IsTrash(): boolean {
        return this.filename === '.user_trash';
    }

    GetOwner(): string {
        return this.owner;
    }

    SetSelected(): void;
    SetSelected(selected: boolean): void;

    SetSelected(selected?: boolean): void {
        if (selected === undefined) {
            this.selected = !this.selected;
            return;
        }
        this.selected = selected;
    }

    IsSelected(): boolean {
        return this.selected;
    }

    IsImported(): boolean {
        return this.imported;
    }

    IsPastFile(): boolean {
        return this.pastFile;
    }

    GetChildren(): string[] {
        return this.children;
    }

    SetHovering(hovering: boolean) {
        this.hovering = hovering;
    }

    IsHovering(): boolean {
        return this.hovering;
    }

    GetBaseIcon(mustBeRoot?: boolean): (p) => JSX.Element {
        if (!this.pathFromHome) {
            return null;
        }
        const parts: any[] = this.pathFromHome.split('/');
        if (mustBeRoot && parts.length > 1) {
            return null;
        }

        if (parts[0] === 'HOME') {
            return IconHome;
        } else if (parts[0] === 'TRASH') {
            return IconTrash;
        } else if (parts[0] === 'SHARE') {
            return IconUser;
        } else {
            console.error('Unknown filepath base type');
            return null;
        }
    }

    GetFileIcon() {
        if (!this.pathFromHome) {
            return null;
        }

        if (this.pathFromHome === 'HOME') {
            return IconHome;
        } else if (this.pathFromHome === 'TRASH') {
            return IconTrash;
        } else if (this.pathFromHome === 'SHARE') {
            return IconUser;
        } else if (this.isDir) {
            return IconFolder;
        } else {
            return IconFile;
        }
    }

    public async LoadShare(authHeader) {
        if (this.share) {
            return this.share;
        } else if (!this.shareId) {
            return null;
        }

        const url = new URL(`${API_ENDPOINT}/file/share/${this.shareId}`);
        return fetch(url.toString(), {
            headers: authHeader,
        })
            .then(r => r.json())
            .then(j => {
                console.log(j);
                this.share = new WeblensShare(j as ShareDataT);
                return this.share;
            });
    }

    GetShare(): WeblensShare {
        return this.share;
    }

    GetVisitRoute(mode: FbModeT, shareId: string, dispatch: FBDispatchT) {
        if (this.isDir) {
            if (mode === FbModeT.share && shareId === '') {
                return `/files/share/${this.shareId}/${this.id}`;
            } else if (mode === FbModeT.share) {
                return `/files/share/${shareId}/${this.id}`;
            } else if (mode === FbModeT.external) {
                return `/files/external/${this.id}`;
            } else if (mode === FbModeT.default) {
                return `/files/${this.id}`;
            }
        } else if (this.displayable) {
            dispatch({
                type: 'set_presentation',
                presentingId: this.id,
            });
            return;
        }
        console.error('Did not find location to visit for', this.filename);
    }
}

export enum SelectedState {
    NotSelected = 0x0,
    Hovering = 0x1,
    InRange = 0x10,
    Selected = 0x100,
    LastSelected = 0x1000,
    Droppable = 0x10000,
}

export type FileContextT = {
    file: WeblensFile;
    selected: SelectedState;
};

export enum FbMenuModeT {
    Closed,
    Default,
    Sharing,
    NameFolder,
    AddToAlbum,
}

export type GlobalContextType = {
    setDragging: (d: DraggingStateT) => void;
    blockFocus: (b: boolean) => void;
    rename: (itemId: string, newName: string) => void;

    setMenuOpen: (m: FbMenuModeT) => void;
    setMenuPos: ({ x, y }: { x: number; y: number }) => void;
    setMenuTarget: (itemId: string) => void;

    setHovering?: (itemId: string) => void;
    setSelected?: (itemId: string, selected?: boolean) => void;
    selectAll?: (itemId: string, selected?: boolean) => void;
    moveSelected?: (itemId: string) => void;
    doSelectMany?: () => string[];
    setMoveDest?: (itemName) => void;

    dragging?: number;
    numCols?: number;
    itemWidth?: number;
    initialScrollIndex?: number;
    hoveringIndex?: number;
    lastSelectedIndex?: number;
    doMediaFetch?: boolean;
    allowEditing?: boolean;
};
