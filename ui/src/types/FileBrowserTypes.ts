import { MediaData } from "./Generic"

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
