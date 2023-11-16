
// Global Types

export type MediaData = {
    BlurHash: string
    CreateDate: string
    FileHash: string
    Filepath: string
    MediaType: {
        FileExtension: []
        FriendlyName: string
        IsRaw: boolean
        IsVideo: boolean
    }
    ThumbFilepath: string
    MediaWidth: number
    MediaHeight: number
    ThumbWidth: number
    ThumbHeight: number
    Thumbnail64: string

    // Local things
    Previous: MediaData
    Next: MediaData
    Display: boolean
    ImgRef: React.MutableRefObject<any>
}

// Gallery Types

export type GalleryBucketProps = {
    date: string
    bucketData: MediaData[]
    dispatch: React.Dispatch<any>
}

export type MediaWrapperProps = {
    mediaData: MediaData
    dispatch
}

export type MediaStateType = {
    mediaMap: Map<string, MediaData>
    mediaCount: number
    maxMediaCount: number
    hasMoreMedia: boolean
    presentingHash: string
    previousLast: string
    includeRaw: boolean
    loading: boolean
    scanProgress: number
    searchContent: string
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
    path: string
    draggingState: number
    loading: boolean
    presentingPath: string
    searchContent: string
    scanProgress: number
    holdingShift: boolean
    lastSelected: string
    editing: string
    hovering: string
}

export type itemData = {
    filepath: string
    updatePath: string
    isDir: boolean
    imported: boolean
    modTime: string
    visible: boolean
    mediaData: MediaData
}

export type fileBrowserAction =
    | { type: 'set_path'; path: string }
    | { type: 'update_item'; item: itemData }
    | { type: 'set_selected'; itempath: string, selected: boolean }
    | { type: 'clear_selected'; }
    | { type: 'holding_shift'; shift: boolean }
    | { type: 'set_loading'; loading: boolean }
    | { type: 'set_dragging'; dragging: boolean }
    | { type: 'set_presentation'; presentingHash: string }
    | { type: 'stop_presenting' }