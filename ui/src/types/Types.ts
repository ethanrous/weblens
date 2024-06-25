// Global Types

import React from 'react'
import { FbMenuModeT, FileInitT, WeblensFile } from '../Files/File'
import { DraggingStateT, FbModeT } from '../Files/filesContext'
import WeblensMedia from '../Media/Media'
import { TaskProgress } from '../Pages/FileBrowser/TaskProgress'
import { GalleryAction } from '../Pages/Gallery/GalleryLogic'

export type AuthHeaderT = {
    Authorization: string
}

export type UserInfoT = {
    homeId: string;
    trashId: string;
    username: string;
    admin: boolean;
    owner: boolean;
    activated: boolean;
    isLoggedIn: boolean;
};

export type ServerInfoT = {
    id: string;
    name: string;
    role: string;
    coreAddress: string;
    userCount: number;
};

export type UserContextT = {
    authHeader: AuthHeaderT;
    usr: UserInfoT;
    setCookie;
    clear;
    serverInfo: ServerInfoT;
};

// export type WeblensMedia = {
//     mediaId: string;
//     fileIds: string[];
//     thumbnailCacheId: string;
//     fullresCacheIds: string;
//     blurHash: string;
//     owner: string;
//     mediaWidth: number;
//     mediaHeight: number;
//     createDate: string;
//     mimeType: string;
//     recognitionTags: string[];
//     pageCount: number;

//     // Non-api props
//     thumbnail: ArrayBuffer;
//     fullres: ArrayBuffer;
//     Previous: WeblensMedia;
//     Next: WeblensMedia;
//     selected: boolean;
//     mediaType;
//     // Display: boolean
//     ImgRef: React.MutableRefObject<any>;
// };

export type WsMessageT = {
    subscribeKey: string;
    eventTag: string;
    taskType: string;
    error: string;
    content: any[];
};

export type mediaType = {
    FileExtension: [];
    FriendlyName: string;
    IsRaw: boolean;
    IsVideo: boolean;
    IsDisplayable: boolean;
};

export type AlbumData = {
    id: string;
    medias: string[];
    sharedWith: string[];
    name: string;
    cover: string;
    // CoverMedia: WeblensMedia;
    primaryColor: string;
    secondaryColor: string;
    owner: string;
    showOnTimeline: boolean;
};

// Gallery Types

export type GalleryBucketProps = {
    bucketTitle: string;
    bucketData: WeblensMedia[];
    scale: number;
    dispatch: React.Dispatch<any>;
};

export type MediaWrapperProps = {
    mediaData: WeblensMedia;
    scale: number;
    width: number;
    showMedia: boolean;
    viewSize: SizeT;
    rowIndex?: number;
    colIndex?: number;
    hoverIndex?: { row: number; col: number };
    albumId?: string;
    fetchAlbum?: () => void;
    menu?: (
        mediaId: string,
        open: boolean,
        setOpen: (open: boolean) => void,
    ) => JSX.Element;
};


export enum PresentType {
    None = 1,
    InLine,
    Fullscreen,
}


export type TimeOffset = {
    second: 0;
    minute: 0;
    hour: 0;
    day: 0;
    month: 0;
    year: 0;
};

export const newTimeOffset = (): TimeOffset => {
    return {
        second: 0,
        minute: 0,
        hour:   0,
        day:    0,
        month:  0,
        year:   0,
    }
}

export type GalleryStateT = {
    selected: Map<string, boolean>;
    albumsMap: Map<string, AlbumData>;
    albumsFilter: string[];
    loading: string[];
    includeRaw: boolean;
    newAlbumDialogue: boolean;
    blockSearchFocus: boolean;
    selecting: boolean;
    menuTargetId: string;
    imageSize: number;
    searchContent: string;
    presentingMediaId: string;
    presentingMode: PresentType;
    timeAdjustOffset: TimeOffset;
    hoverIndex: number;
    lastSelIndex: number;
    holdingShift: boolean;
};

// File Browser Types

export type FileBrowserAction = {
    type: string;

    loading?: string;
    serverId?: string;
    poolId?: string;
    fileId?: string;
    fileName?: string;
    search?: string;
    presentingId?: string;
    hovering?: string;
    direction?: string;
    realId?: string;
    shareId?: string;
    sortType?: string;
    taskType?: string;
    target?: string;
    note?: string;

    mode?: FbModeT;
    menuMode?: FbMenuModeT;

    fileIds?: string[];

    dragging?: DraggingStateT;
    selected?: boolean;
    external?: boolean;
    block?: boolean;
    shift?: boolean;
    open?: boolean;
    isSearching?: boolean;

    progress?: number;
    tasksComplete?: number | string;
    tasksTotal?: number | string;
    numCols?: number;
    sortDirection?: number;
    time?: number;

    user?: UserInfoT;

    img?: ArrayBuffer;
    pos?: { x: number; y: number };

    file?: WeblensFile;
    fileInfo?: FileInitT;
    files?: FileInitT[];

    past?: Date;
};

export type FBDispatchT = (action: FileBrowserAction) => void;
export type GalleryDispatchT = (action: GalleryAction) => void;

export type ScanMeta = {
    taskId: string;
    taskType: string;
    target: string;
    mostRecent: string;
    note: string;

    progress: number;
    tasksComplete: number;
    tasksTotal: number;
    time: number;

    complete: boolean;
};

export type FbStateT = {
    dirMap: Map<string, WeblensFile>;
    selected: Map<string, boolean>;
    uploadMap: Map<string, boolean>;
    menuPos: { x: number; y: number };
    folderInfo: WeblensFile;
    parents: WeblensFile[];
    filesList: string[];
    draggingState: number;
    loading: string[];
    menuTargetId: string;
    presentingId: string;
    searchContent: string;
    isSearching: boolean;
    scanProgress: TaskProgress[];
    homeDirSize: number;
    trashDirSize: number;
    numCols: number;
    holdingShift: boolean;
    blockFocus: boolean;
    lastSelected: string;
    hovering: string;
    pasteImg: ArrayBuffer;
    scrollTo: string;
    moveDest: string;
    menuMode: FbMenuModeT;
    fileInfoMenu: boolean;

    fbMode: FbModeT;

    shareId: string;
    contentId: string;

    sortDirection: number; // 1 or -1
    sortFunc: string;
    viewingPast: Date;
};

export type SizeT = {
    height: number;
    width: number;
};

// export const getBlankFile = () => {
//     const blank: WeblensFile = {
//         id: "",
//         imported: false,
//         displayable: false,
//         isDir: false,
//         modifiable: false,
//         size: 0,
//         modTime: new Date().toString(),
//         filename: "",
//         parentFolderId: "",
//         mediaData: null,
//         fileFriendlyName: "",
//         owner: "",
//         pathFromHome: "",
//         shares: [],
//         children: [],
//         pastFile: false,

//         visible: false,
//         selected: false,
//     };
//     return blank;
// };

// export const getBlankMedia = () => {
//     const blank: WeblensMedia = {
//         mediaId: "",
//         fileIds: [""],
//         thumbnailCacheId: "",
//         fullresCacheIds: "",
//         blurHash: "",
//         owner: "",
//         mediaWidth: 0,
//         mediaHeight: 0,
//         createDate: "",
//         mimeType: "",
//         recognitionTags: [],
//         pageCount: 0,

//         selected: false,
//         thumbnail: null,
//         fullres: null,
//         Previous: null,
//         mediaType: null,
//         Next: null,
//         ImgRef: null,
//     };
//     return blank;
// };
