import {
    TaskProgressState,
    TasksProgressAction,
} from '@weblens/pages/FileBrowser/TaskProgress'
import { createContext, Dispatch } from 'react'

export type TaskProgContextT = {
    progState: TaskProgressState
    progDispatch: Dispatch<TasksProgressAction>
}

export const TaskProgContext = createContext<TaskProgContextT>({
    progState: null,
    progDispatch: null,
})

export enum DraggingStateT {
    NoDrag, // No dragging is taking place
    InternalDrag, // Dragging is of only internal elements

    // Dragging is from external source, such as
    // dragging files from your computer over the browser
    ExternalDrag,
}
