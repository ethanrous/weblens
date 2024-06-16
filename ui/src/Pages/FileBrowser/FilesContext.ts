import { GlobalContextType, SelectedState } from '../../Files/FileDisplay'
import { FileContextT } from '../../components/FileScroller'
import {
    AuthHeaderT,
    FBDispatchT,
    FbStateT,
    UserInfoT,
} from '../../types/Types'
import { binarySearch } from '../../util'
import { DraggingState } from './FileBrowser'
import { getSortFunc, handleRename, MoveSelected } from './FileBrowserLogic'
import { FbMenuModeT } from './FileBrowserStyles'

export function GetFilesContext(
    fb: FbStateT,
    files: FileContextT[],
    hoveringIndex: number,
    lastSelectedIndex: number,
    authHeader: AuthHeaderT,
    dispatch: FBDispatchT
) {
    let scrollToIndex: number = null
    if (fb.scrollTo) {
        scrollToIndex = files.findIndex((v) => v.file.Id() === fb.scrollTo)
    }

    const context: GlobalContextType = {
        setSelected: (itemId: string, selected?: boolean) =>
            dispatch({
                type: 'set_selected',
                fileId: itemId,
                selected: selected,
            }),
        setHovering: (i: string) =>
            dispatch({ type: 'set_hovering', hovering: i }),
        setDragging: (d: DraggingState) =>
            dispatch({ type: 'set_dragging', dragging: d }),
        doSelectMany: () => {
            if (lastSelectedIndex === -1) {
                dispatch({
                    type: 'set_selected',
                    fileId: fb.hovering,
                })
                return
            }

            dispatch({
                type: 'select_ids',
                fileIds:
                    hoveringIndex > lastSelectedIndex
                        ? files
                              .slice(lastSelectedIndex, hoveringIndex + 1)
                              .map((v) => v.file.Id())
                        : files
                              .slice(hoveringIndex, lastSelectedIndex + 1)
                              .map((v) => v.file.Id()),
            })
        },
        moveSelected: (entryId: string) => {
            if (fb.dirMap.get(entryId).IsFolder()) {
                MoveSelected(fb.selected, entryId, authHeader).then(() =>
                    dispatch({ type: 'clear_selected' })
                )
            }
        },
        blockFocus: (b: boolean) =>
            dispatch({ type: 'set_block_focus', block: b }),
        rename: (itemId: string, newName: string) =>
            handleRename(itemId, newName, dispatch, authHeader),
        setMenuOpen: (m: FbMenuModeT) =>
            dispatch({ type: 'set_menu_open', menuMode: m }),
        setMenuPos: (pos: { x: number; y: number }) =>
            dispatch({ type: 'set_menu_pos', pos: pos }),
        setMenuTarget: (target: string) =>
            dispatch({ type: 'set_menu_target', fileId: target }),

        setMoveDest: (itemName) =>
            dispatch({ type: 'set_move_dest', fileName: itemName }),

        dragging: fb.draggingState,
        initialScrollIndex: scrollToIndex,
        allowEditing: fb.folderInfo.IsModifiable(),
        hoveringIndex: hoveringIndex,
        lastSelectedIndex: lastSelectedIndex,
    }
    return context
}

export function GetItemsList(
    fb: FbStateT,
    usr: UserInfoT,
    debouncedSearch: string
): { files: FileContextT[]; hoveringIndex: number; lastSelectedIndex: number } {
    if (usr.isLoggedIn === undefined) {
        return { files: [], hoveringIndex: -1, lastSelectedIndex: -1 }
    }
    let filesList = Array.from(fb.dirMap.values()).filter((val) => {
        if (!val.GetFilename()) {
            return false
        }
        return (
            val
                .GetFilename()
                .toLowerCase()
                .includes(debouncedSearch.toLowerCase()) && !val.IsTrash()
        )
    })

    const sortFunc = getSortFunc(fb.sortFunc, fb.sortDirection)
    filesList.sort(sortFunc)

    let hoveringIndex = -1
    let lastSelectedIndex = -1

    if (fb.holdingShift && fb.lastSelected && fb.hovering) {
        const hovering = fb.dirMap.get(fb.hovering)
        const lastSelected = fb.dirMap.get(fb.lastSelected)
        if (hovering) {
            hoveringIndex = binarySearch(filesList, hovering, sortFunc)
        }
        if (lastSelected) {
            lastSelectedIndex = binarySearch(filesList, lastSelected, sortFunc)
        }
    }

    const ctxS = filesList.map((v, i) => {
        const itemId = v.Id()
        const ctx: FileContextT = {
            file: v,
            selected:
                (fb.selected.get(itemId) && SelectedState.Selected) |
                (fb.hovering === itemId && SelectedState.Hovering) |
                (fb.holdingShift &&
                    (i - hoveringIndex) * (i - lastSelectedIndex) < 1 &&
                    SelectedState.InRange) |
                (v.Id() === fb.lastSelected && SelectedState.LastSelected) |
                (fb.draggingState !== DraggingState.NoDrag &&
                    v.IsFolder() &&
                    !v.IsSelected() &&
                    fb.hovering === itemId &&
                    SelectedState.Droppable),
        }

        return ctx
    })

    return { files: ctxS, hoveringIndex, lastSelectedIndex }
}
