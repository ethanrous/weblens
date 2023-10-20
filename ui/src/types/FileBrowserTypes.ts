import { EnqueueSnackbar } from "notistack"
import { MediaData } from "./Generic"


export type FileBrowserTypes = {
    wsSend: (msg: string) => void
    lastMessage: MessageEvent<any> | null
    readyState: ReadyState
    enqueueSnackbar: EnqueueSnackbar
}

export type FileBrowserStateType = {
    dirMap: Map<string, itemData>
    path: string
    dragging: boolean
    loading: boolean
    presentingPath: string
    numSelected: number
    scanProgress: number
    holdingShift: boolean
    lastSelected: string
    editing: string
}

export type itemData = {
    filepath: string
    isDir: boolean
    imported: boolean
    modTime: string
    selected: boolean
    mediaData: MediaData
}

export type fileBrowserAction =
    | { type: 'set_path'; path: string }
    | { type: 'update_items'; item: [{}] }
    | { type: 'set_selected'; itempath: string, selected: boolean }
    | { type: 'clear_selected'; }
    | { type: 'holding_shift'; shift: boolean }
    | { type: 'set_loading'; loading: boolean }
    | { type: 'set_dragging'; dragging: boolean }
    | { type: 'set_presentation'; presentingHash: string }
    | { type: 'stop_presenting' }
