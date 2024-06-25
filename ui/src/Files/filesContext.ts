import { FBDispatchT, FbStateT } from '../types/Types'
import { createContext } from 'react'

export type FbContextT = {
    fbState: FbStateT
    fbDispatch: FBDispatchT
}

export const FbContext = createContext<FbContextT>({
    fbState: null,
    fbDispatch: null,
})

export enum DraggingStateT {
    NoDrag, // No dragging is taking place
    InternalDrag, // Dragging is of only internal elements

    // Dragging is from external source, such as
    // dragging files from your computer over the browser
    ExternalDrag,
}

export enum FbModeT {
    unset,
    default,
    share,
    external,
    stats,
    search,
}
