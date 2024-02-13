
// Global Types

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
    imageSize: number
    scanProgress: number
    showingCount: number
    searchContent: string
    presentingMedia: MediaData
}

// File Browser Types

export type FileBrowserAction = {
    type: string

    selected?: boolean
    loading?: boolean
    external?: boolean
    block?: boolean
    shift?: boolean

    dragging?: boolean
    progress?: number

    user?: string
    fileId?: string
    fileIds?: string[]
    fileName?: string
    search?: string
    presentingId?: string

    img?: ArrayBuffer

    fileInfo?: fileData
    fileInfos?: fileData[]
    files?: { fileId: string, updateInfo: fileData }[]

    parents?: any
}

export type FileBrowserDispatch = (action: FileBrowserAction) => void

export type FileBrowserStateType = {
    dirMap: Map<string, fileData>
    selected: Map<string, boolean>
    uploadMap: Map<string, boolean>
    folderInfo: fileData
    parents: fileData[]
    draggingState: number
    loading: boolean
    waitingForNewName: string
    presentingId: string
    searchContent: string
    scanProgress: number
    homeDirSize: number
    trashDirSize: number
    holdingShift: boolean
    blockFocus: boolean
    lastSelected: string
    pasteImg: ArrayBuffer
    scrollTo: string
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
    shares: any[]
    visible: boolean
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
        shares: [],
        visible: false,
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
        createDate: "",
        owner: "",
        recognitionTags: [],

        Previous: null,
        Next: null,

        ImgRef: null
    }
    return blank
}