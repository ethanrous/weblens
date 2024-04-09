// Global Types

import { GalleryAction } from "../Pages/Gallery/GalleryLogic";
import { ItemProps } from "../components/ItemDisplay";

export type AuthHeaderT = {
    Authorization: string;
};

export type UserInfoT = {
    homeId: string;
    trashId: string;
    username: string;
    admin: boolean;
    owner: boolean;
    activated: boolean;
    isLoggedIn: boolean;
};

export type UserContextT = {
    authHeader: AuthHeaderT;
    usr: UserInfoT;
    setCookie;
    clear;
    serverInfo;
}

export type MediaDataT = {
    mediaId: string;
    fileIds: string[];
    thumbnailCacheId: string;
    fullresCacheIds: string;
    blurHash: string;
    owner: string;
    mediaWidth: number;
    mediaHeight: number;
    createDate: string;
    mimeType: string;
    recognitionTags: string[];
    pageCount: number;

    // Non-api props
    thumbnail: ArrayBuffer;
    fullres: ArrayBuffer;
    Previous: MediaDataT;
    Next: MediaDataT;
    selected: boolean;
    mediaType;
    // Display: boolean
    ImgRef: React.MutableRefObject<any>;
};

export type mediaType = {
    FileExtension: [];
    FriendlyName: string;
    IsRaw: boolean;
    IsVideo: boolean;
    IsDisplayable: boolean;
};

export type AlbumData = {
    Id: string;
    Medias: string[];
    SharedWith: string[];
    Name: string;
    Cover: string;
    CoverMedia: MediaDataT;
    PrimaryColor: string;
    SecondaryColor: string;
    Owner: string;
    ShowOnTimeline: boolean;
};

// Gallery Types

export type GalleryBucketProps = {
    bucketTitle: string;
    bucketData: MediaDataT[];
    scale: number;
    dispatch: React.Dispatch<any>;
};

export type MediaWrapperProps = {
    mediaData: MediaDataT;
    selected: boolean;
    selecting: boolean;
    scale: number;
    dispatch: GalleryDispatch;
    menu?: (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => JSX.Element;
};

export type MediaStateT = {
    mediaMap: Map<string, MediaDataT>;
    selected: Map<string, boolean>;
    mediaMapUpdated: number;
    albumsMap: Map<string, AlbumData>;
    albumsFilter: string[];
    loading: string[];
    includeRaw: boolean;
    newAlbumDialogue: boolean;
    blockSearchFocus: boolean;
    menuOpen: boolean;
    selecting: boolean;
    menuTargetId: string;
    imageSize: number;
    scanProgress: number;
    showingCount: number;
    searchContent: string;
    presentingMedia: MediaDataT;
    menuPos: { x: number; y: number };
};

// File Browser Types

export type FileBrowserAction = {
    type: string;

    loading?: string;
    taskId?: string;
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
    mode?: string;

    fileIds?: string[];

    dragging?: boolean;
    selected?: boolean;
    external?: boolean;
    block?: boolean;
    shift?: boolean;
    open?: boolean;

    progress?: number;
    tasksComplete?: number;
    tasksTotal?: number;
    numCols?: number;
    sortDirection?: number;
    time?: number;

    user?: UserInfoT;

    img?: ArrayBuffer;
    pos?: { x: number; y: number };

    fileInfo?: FileInfoT;
    itemInfo?: ItemProps;
    fileInfos?: FileInfoT[];
    files?: { fileId: string; updateInfo: FileInfoT }[];

    parents?: any;
};

export type FBDispatchT = (action: FileBrowserAction) => void;
export type GalleryDispatch = (action: GalleryAction) => void;

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
    dirMap: Map<string, FileInfoT>;
    selected: Map<string, boolean>;
    uploadMap: Map<string, boolean>;
    menuPos: { x: number; y: number };
    searchResults: FileInfoT[];
    folderInfo: FileInfoT;
    parents: FileInfoT[];
    filesList: string[];
    draggingState: number;
    loading: string[];
    waitingForNewName: string;
    menuTargetId: string;
    presentingId: string;
    searchContent: string;
    scanProgress: ScanMeta[];
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
    menuOpen: boolean;

    fbMode: string;
    // isShare: boolean;
    // isExternal: boolean;

    shareId: string;
    contentId: string;

    sortDirection: number; // 1 or -1
    sortFunc: string;
};

export type FileInfoT = {
    id: string;
    imported: boolean;
    displayable: boolean;
    isDir: boolean;
    modifiable: boolean;
    size: number;
    modTime: string;
    filename: string;
    parentFolderId: string;
    mediaData: MediaDataT;
    fileFriendlyName: string;
    owner: string;
    pathFromHome: string;
    shares: shareData[];
    children: string[];

    // Non-api props
    visible: boolean;
};

export type shareData = {
    Accessors: string[];
    Expires: string;
    Public: boolean;
    shareId: string;
    fileId: string;
    ShareName: string;
    Wormhole: boolean;
};

export const getBlankFile = () => {
    const blank: FileInfoT = {
        id: "",
        imported: false,
        displayable: false,
        isDir: false,
        modifiable: false,
        size: 0,
        modTime: new Date().toString(),
        filename: "",
        parentFolderId: "",
        mediaData: null,
        fileFriendlyName: "",
        owner: "",
        pathFromHome: "",
        shares: [],
        visible: false,
        children: [],
    };
    return blank;
};

export const getBlankMedia = () => {
    const blank: MediaDataT = {
        mediaId: "",
        fileIds: [""],
        thumbnailCacheId: "",
        fullresCacheIds: "",
        blurHash: "",
        owner: "",
        mediaWidth: 0,
        mediaHeight: 0,
        createDate: "",
        mimeType: "",
        recognitionTags: [],
        pageCount: 0,

        selected: false,
        thumbnail: null,
        fullres: null,
        Previous: null,
        mediaType: null,
        Next: null,
        ImgRef: null,
    };
    return blank;
};
