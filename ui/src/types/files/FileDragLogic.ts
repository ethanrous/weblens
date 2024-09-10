import { FbModeT, SetMenuT } from '@weblens/pages/FileBrowser/FBStateControl'
import { MoveSelected } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import {
    FbMenuModeT,
    SelectedState,
    WeblensFile,
} from '@weblens/types/files/File'
import { Dispatch, MouseEvent } from 'react'

export function mouseMove(
    e,
    file: WeblensFile,
    draggingState: DraggingStateT,
    mouseDown: { x: number; y: number },
    setSelected: (fileIds: string[]) => void,
    setDragging: (dragging: DraggingStateT) => void
) {
    if (
        mouseDown &&
        !draggingState &&
        (Math.abs(mouseDown.x - e.clientX) > 20 ||
            Math.abs(mouseDown.y - e.clientY) > 20)
    ) {
        if (!(file.GetSelectedState() & SelectedState.Selected)) {
            setSelected([file.Id()])
        }
        setDragging(DraggingStateT.InternalDrag)
    }
}

export function visitFile(
    e,
    mode: FbModeT,
    shareId: string,
    file: WeblensFile,
    inTrash: boolean,
    nav,
    setPresentation: (presentingId: string) => void
) {
    if (inTrash && file.IsFolder()) {
        return
    }

    e.stopPropagation()
    const jump = file.GetVisitRoute(mode, shareId, setPresentation)
    if (jump) {
        nav(jump)
    }
}

export function fileHandleContextMenu(
    e,
    menuMode: FbMenuModeT,
    setMenu: SetMenuT,
    file: WeblensFile
) {
    e.stopPropagation()
    e.preventDefault()

    setMenu({
        menuState: FbMenuModeT.Default,
        menuTarget: file.Id(),
        menuPos: { x: e.clientX, y: e.clientY },
    })
}

export function handleMouseUp(
    file: WeblensFile,
    draggingState: DraggingStateT,
    selected: string[],
    setSelectedMoved: () => void,
    clearSelected: () => void,
    setMoveDest: (dest: string) => void,
    setDragging: (dragging: DraggingStateT) => void,
    setMouseDown: Dispatch<any>
) {
    if (draggingState !== DraggingStateT.NoDrag) {
        if (
            !(file.GetSelectedState() & SelectedState.Selected) &&
            file.IsFolder()
        ) {
            setSelectedMoved()
            MoveSelected(selected, file.Id()).then(() => clearSelected())
        }
        setMoveDest('')
        setDragging(DraggingStateT.NoDrag)
    }
    setMouseDown(null)
}

export function handleMouseLeave(
    file: WeblensFile,
    draggingState: DraggingStateT,
    setMoveDest: (dest: string) => void,
    setHovering: (hovering: string) => void,
    mouseDown: boolean,
    setMouseDown
) {
    file.UnsetSelected(SelectedState.Hovering)
    file.UnsetSelected(SelectedState.Droppable)
    setHovering('')
    if (draggingState && file.IsFolder()) {
        setMoveDest('')
    }
    if (mouseDown) {
        setMouseDown(null)
    }
}

export function handleMouseOver(
    e: MouseEvent<HTMLDivElement>,
    file: WeblensFile,
    draggingState: DraggingStateT,
    setHovering: (hovering: string) => void,
    setMoveDest: (dest: string) => void
) {
    e.stopPropagation()
    if (draggingState === DraggingStateT.NoDrag || file.IsFolder()) {
        file.SetSelected(SelectedState.Hovering)
        setHovering(file.Id())
    }

    if (
        draggingState &&
        !(file.GetSelectedState() & SelectedState.Selected) &&
        file.IsFolder()
    ) {
        file.SetSelected(SelectedState.Droppable)
        setMoveDest(file.GetFilename())
    }
}
