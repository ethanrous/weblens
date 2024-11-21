import { FileApi } from '@weblens/api/FileBrowserApi'
import {
    MenuOptionsT,
    useFileBrowserStore,
} from '@weblens/pages/FileBrowser/FBStateControl'
import { DirViewModeT } from '@weblens/pages/FileBrowser/FileBrowserTypes'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import {
    FbMenuModeT,
    SelectedState,
    WeblensFile,
} from '@weblens/types/files/File'
import { Dispatch, MouseEvent } from 'react'

import { Coordinates } from '../Types'

export function mouseMove(
    e: MouseEvent,
    file: WeblensFile,
    draggingState: DraggingStateT,
    mouseDown: Coordinates,
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
    e: MouseEvent,
    file: WeblensFile,
    inTrash: boolean,
    setPresentation: (presentingId: string) => void
) {
    if (inTrash && file.IsFolder()) {
        return
    }
    e.stopPropagation()
    if (file.IsFolder()) {
        goToFile(file)
    } else {
        setPresentation(file.Id())
    }
}

export function fileHandleContextMenu(
    e: MouseEvent,
    setMenu: (opts: MenuOptionsT) => void,
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
    setMouseDown: Dispatch<Coordinates>,
    viewMode: DirViewModeT
) {
    if (draggingState !== DraggingStateT.NoDrag) {
        if (
            !(file.GetSelectedState() & SelectedState.Selected) &&
            file.IsFolder()
        ) {
            setSelectedMoved()
            FileApi.moveFiles({
                fileIds: selected,
                newParentId: file.Id(),
            })
                .then(() => {})
                .catch((err) => {
                    console.error('Failed to move files', err)
                })
        }
        setMoveDest('')
        setDragging(DraggingStateT.NoDrag)
        file.SetSelected(SelectedState.Hovering, true)
    }

    const state = useFileBrowserStore.getState()
    if (!state.holdingShift && viewMode === DirViewModeT.Columns) {
        goToFile(file, true)
    } else if (state.holdingShift) {
        const state = useFileBrowserStore.getState()
        if (
            viewMode === DirViewModeT.Columns &&
            file.parentId !== state.folderInfo?.Id()
        ) {
            const parent = state.filesMap.get(file.parentId)
            goToFile(parent)
        }
        state.setSelected([file.Id()], false)
    }

    setMouseDown(null)
}

export function handleMouseLeave(
    e: MouseEvent<HTMLDivElement>,
    file: WeblensFile,
    draggingState: DraggingStateT,
    fileRef: HTMLDivElement,
    setMoveDest: (dest: string) => void,
    setHovering: (hovering: string) => void,
    mouseDown: Coordinates,
    setMouseDown: Dispatch<Coordinates>
) {
    if (draggingState === DraggingStateT.ExternalDrag) {
        if (
            e.relatedTarget &&
            e.relatedTarget instanceof Node &&
            !fileRef.contains(e.relatedTarget)
        ) {
            setHovering('')
            file.UnsetSelected(SelectedState.Droppable)
        }
        return
    }
    setHovering('')
    file.UnsetSelected(SelectedState.Hovering)
    file.UnsetSelected(SelectedState.Droppable)
    if (draggingState === DraggingStateT.InternalDrag && file.IsFolder()) {
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
    setHovering: (hoveringId: string) => void,
    setMoveDest: (dest: string) => void,
    setDragging: (dragging: DraggingStateT) => void
) {
    e.stopPropagation()
    e.preventDefault()
    if (draggingState === DraggingStateT.NoDrag && e.type === 'dragenter') {
        setDragging(DraggingStateT.ExternalDrag)
        draggingState = DraggingStateT.ExternalDrag
    }

    if (
        draggingState === DraggingStateT.InternalDrag &&
        file.IsFolder() &&
        !(file.GetSelectedState() & SelectedState.Selected)
    ) {
        file.SetSelected(SelectedState.Droppable)
        setMoveDest(file.Id())
        setHovering(file.Id())
    } else if (draggingState === DraggingStateT.NoDrag) {
        file.SetSelected(SelectedState.Hovering)
        setHovering(file.Id())
    } else if (
        draggingState === DraggingStateT.ExternalDrag &&
        !(file.GetSelectedState() & SelectedState.Selected) &&
        file.IsFolder()
    ) {
        setMoveDest(file.Id())
        setHovering(file.Id())
        file.SetSelected(SelectedState.Droppable)
    }
}

export function goToFile(next: WeblensFile, allowBlindHop: boolean = false) {
    if (!next) {
        console.error('goToFile called with no next file')
        return
    }

    const state = useFileBrowserStore.getState()

    // if (state.holdingShift) {
    //     return
    // }

    if (!state.folderInfo && allowBlindHop) {
        state.clearFiles()
        console.debug('GO TO FILE (blind):', next.Id())
        state.nav('/files/' + next.Id())
        return
    }

    // If the next file is a folder, we WILL be navigating to it.
    // We can do that with state change and a url update, and not
    // a full page reload.
    const parents = state.folderInfo ? [...state.folderInfo.parents] : []
    if (next.IsFolder()) {
        state.setPresentationTarget('')
        if (next.parentId && next.parentId === state.folderInfo?.Id()) {
            // If the next files parent is the currentFolder, we can set the parents
            // based on what we already have, and add the currentFolder to the list.
            parents.push(state.folderInfo)
            next.parents = parents
        } else if (
            next.parentId &&
            parents.map((p) => p.Id()).includes(next.Id())
        ) {
            // If the next file is the current folders parent of any distance (i.e. we are going up a level)
            while (
                parents.length &&
                parents[parents.length - 1].Id() !== next.parentId
            ) {
                parents.pop()
            }
            next.parents = parents
        } else if (next.Id() === state.folderInfo?.Id()) {
            // If we are in a folder and have selected a non-folder child, going up to the parent
            // is trivial, just just select the parent and nothing else
            next = state.folderInfo
        } else if (next.parentId === state.folderInfo?.parentId) {
            next.parents = [...state.folderInfo.parents]
        } else if (allowBlindHop) {
            // If we can't find a way to quickly navigate to the next file, we can just reload the page
            // at the new location. We need to clear the current files
            state.nav('/files/' + next.Id())
            state.clearFiles()

            console.debug('GO TO FILE (blind):', next)
            return
        } else {
            console.error(
                'BAD! goToFile did not find a valid state update rule'
            )
            return
        }
        next.modifiable = true
        next.parents = next.parents.map((p) => state.filesMap.get(p.Id()))

        console.debug('GO TO FILE:', next)
        state.setFilesData({ selfInfo: next, overwriteContentId: true })
    } else {
        // If the next is not a folder, we can set the location to the parent of the next file,
        // with the child as the "jumpTo" parameter. If the parent is the same as the current folder,
        // this is just a simple state change, and we don't need to fetch any new data

        let newParent = state.folderInfo

        if (next.parentId !== state.folderInfo?.Id()) {
            newParent = useFileBrowserStore
                .getState()
                .filesMap.get(next.parentId)
            if (newParent) {
                while (
                    parents.length &&
                    parents[parents.length - 1].Id() !== newParent.Id()
                ) {
                    parents.pop()
                }
                parents.pop()

                newParent.parents = parents
            } else if (allowBlindHop) {
                console.debug('Doing blind hop to', next.GetFilename())
                state.clearFiles()
                state.nav('/files/' + next.parentId + '#' + next.Id())
                return
            } else {
                console.error(
                    'BAD! goToFile did not find a valid state update rule'
                )
                return
            }
        } else if (state.presentingId) {
            state.setPresentationTarget(next.Id())
        }

        next.parents = next.parents.map((p) => state.filesMap.get(p.Id()))
        state.setFilesData({ selfInfo: newParent, overwriteContentId: true })
        state.setLocationState({ contentId: newParent.Id(), jumpTo: next.Id() })
    }
    state.setSelected([next.Id()], true)
}

export const activeItemsFromState = (
    filesMap: Map<string, WeblensFile>,
    selected: Map<string, boolean>,
    menuTargetId: string
): {
    items: WeblensFile[]
    anyDisplayable: boolean
    mediaCount: number
} => {
    if (filesMap.size === 0) {
        return { items: [], anyDisplayable: false, mediaCount: 0 }
    }
    const isSelected = Boolean(selected.get(menuTargetId))
    const itemIds = isSelected ? Array.from(selected.keys()) : [menuTargetId]
    let mediaCount = 0
    const items = itemIds.map((i) => {
        const item = filesMap.get(i)
        if (!item) {
            return null
        }
        if (item.GetContentId() || item.IsFolder()) {
            mediaCount++
        }
        return item
    })

    return {
        items: items.filter((i) => Boolean(i)),
        anyDisplayable: undefined,
        mediaCount,
    }
}
