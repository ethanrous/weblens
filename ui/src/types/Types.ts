
// Global Types

import { ItemProps } from "../components/ItemDisplay"

export type MediaData = {
    fileHash: string
    fileIds: string[]
    mediaType: {
        FileExtension: []
        FriendlyName: string
        IsRaw: boolean
        IsVideo: boolean
        IsDisplayable: boolean
    }
    blurHash: string
    thumbnail: ArrayBuffer
    fullres: ArrayBuffer
    mediaWidth: number
    mediaHeight: number
    thumbWidth: number
    thumbHeight: number
    pageCount: number
    createDate: string
    owner: string
    recognitionTags: string[]

    // Local things
    Previous: MediaData
    Next: MediaData
    // Display: boolean
    ImgRef: React.MutableRefObject<any>
}

export type AlbumData = {
    Id: string
    Medias: string[]
    SharedWith: string[]
    Name: string
    Cover: string
    CoverMedia: MediaData
    PrimaryColor: string
    SecondaryColor: string
    Owner: string
    ShowOnTimeline: boolean
}

// Gallery Types

export type GalleryBucketProps = {
    bucketTitle: string
    bucketData: MediaData[]
    scale: number
    dispatch: React.Dispatch<any>
}

export type MediaWrapperProps = {
    mediaData: MediaData
    scale: number
    dispatch: (galleryAction) => void
    menu?: (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => JSX.Element
}

export type MediaStateType = {
    mediaMap: Map<string, MediaData>
    mediaMapUpdated: number
    albumsMap: Map<string, AlbumData>
    albumsFilter: string[]
    loading: boolean
    includeRaw: boolean
    newAlbumDialogue: boolean
    blockSearchFocus: boolean
    menuOpen: boolean
    menuTargetId: string
    imageSize: number
    scanProgress: number
    showingCount: number
    searchContent: string
    presentingMedia: MediaData
    menuPos: { x: number, y: number }
}

// File Browser Types

export type FileBrowserAction = {
    type: string

    dragging?: boolean
    selected?: boolean
    external?: boolean
    loading?: boolean
    block?: boolean
    shift?: boolean
    open?: boolean
    isShare?: boolean

    progress?: number
    numCols?: number

    user?: string
    fileId?: string
    fileIds?: string[]
    fileName?: string
    search?: string
    presentingId?: string
    direction?: string
    realId?: string
    shareId?: string

    img?: ArrayBuffer
    pos?: {x: number, y: number}

    fileInfo?: fileData
    itemInfo?: ItemProps
    fileInfos?: fileData[]
    files?: { fileId: string, updateInfo: fileData }[]

    parents?: any
}

export type FileBrowserDispatch = (action: FileBrowserAction) => void

export type FileBrowserStateType = {
    dirMap: Map<string, fileData>
    selected: Map<string, boolean>
    uploadMap: Map<string, boolean>
    menuPos: { x: number, y: number },
    folderInfo: fileData
    parents: fileData[]
    filesList: string[]
    draggingState: number
    loading: boolean
    waitingForNewName: string
    menuTargetId: string
    presentingId: string
    searchContent: string
    scanProgress: number
    homeDirSize: number
    trashDirSize: number
    numCols: number
    holdingShift: boolean
    blockFocus: boolean
    lastSelected: string
    pasteImg: ArrayBuffer
    scrollTo: string
    moveDest: string
    menuOpen: boolean
    isShare: boolean
    shareId: string
    realId: string
}

export type fileData = {
    id: string
    imported: boolean
    displayable: boolean
    isDir: boolean
    modifiable: boolean
    size: number
    modTime: string
    filename: string
    parentFolderId: string
    mediaData: MediaData
    fileFriendlyName: string
    owner: string
    pathFromHome: string
    shares: shareData[]
    visible: boolean
    children: string[]
}

export type shareData = {
    Accessors: string[]
    Expires: string
    Public: boolean
    shareId: string
    fileId: string
    ShareName: string
    Wormhole: boolean
}

export const getBlankFile = () => {
    const blank: fileData = {
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
    }
    return blank
}

export const getBlankMedia = () => {
    const blank: MediaData = {
        fileHash: "",
        fileIds: [""],
        mediaType: null,
        blurHash: "",
        thumbnail: null,
        fullres: null,
        mediaWidth: 0,
        mediaHeight: 0,
        thumbWidth: 0,
        thumbHeight: 0,
        pageCount: 0,
        createDate: "",
        owner: "",
        recognitionTags: [],

        Previous: null,
        Next: null,

        ImgRef: null
    }
    return blank
}