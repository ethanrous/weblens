
// Global Types

export type MediaData = {
    fileHash: string
    parentFolder: string
    filename: string
    mediaType: {
        FileExtension: []
        FriendlyName: string
        IsRaw: boolean
        IsVideo: boolean
        IsDisplayable: boolean
    }
    blurHash: string
    thumbnail64: string
    mediaWidth: number
    mediaHeight: number
    thumbWidth: number
    thumbHeight: number
    createDate: string
    owner: string

    // Local things
    Previous: MediaData
    Next: MediaData
    // Display: boolean
    ImgRef: React.MutableRefObject<any>
}

export type AlbumData = {
    Id: string
    Medias: string[]
    Name: string
    Cover: string
    PrimaryColor: string
    SecondaryColor: string
    Owner: string
    ShowOnTimeline: boolean
}

// Gallery Types

export type GalleryBucketProps = {
    date: string
    bucketData: MediaData[]
    scrollerRef
    dispatch: React.Dispatch<any>
}

export type MediaWrapperProps = {
    mediaData: MediaData
    scale: number
    scrollerRef
    dispatch
    menu?: (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => JSX.Element
}

export type MediaStateType = {
    mediaMap: Map<string, MediaData>
    albumsMap: Map<string, AlbumData>
    mediaCount: number
    // maxMediaCount: number
    // hasMoreMedia: boolean
    presentingHash: string
    // previousLast: string
    includeRaw: boolean
    loading: boolean
    scanProgress: number
    searchContent: string
    blockSearchFocus: boolean
    newAlbumDialogue: boolean
}

// File Browser Types

export type FileBrowserTypes = {
    wsSend: (msg: string) => void
    lastMessage: MessageEvent<any> | null
    readyState: ReadyState
}

export type FileBrowserStateType = {
    dirMap: Map<string, itemData>
    selected: Map<string, boolean>
    uploadMap: Map<string, boolean>
    folderInfo: itemData,
    parents: itemData[],
    draggingState: number
    loading: boolean
    presentingId: string
    searchContent: string
    scanProgress: number
    holdingShift: boolean
    sharing: boolean
    albuming: boolean
    lastSelected: string
    editing: string
    hovering: string
}

export type itemData = {
    id: string
    parentFolderId: string
    filename: string
    isDir: boolean
    imported: boolean
    modTime: string
    owner: string
    size: number
    visible: boolean
    mediaData: MediaData
}