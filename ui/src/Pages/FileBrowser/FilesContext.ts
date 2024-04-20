import { GlobalContextType, ItemProps } from "../../components/ItemDisplay";
import { AuthHeaderT, FBDispatchT, FbStateT, UserInfoT } from "../../types/Types";
import { humanFileSize } from "../../util";
import { getSortFunc, getVisitRoute, handleRename, MoveSelected } from "./FileBrowserLogic";
import { IconDisplay } from "./FileBrowserStyles";

export function GetFilesContext(fb: FbStateT, itemsList: ItemProps[], nav, authHeader: AuthHeaderT, dispatch: FBDispatchT) {
    let scrollToIndex: number = null;
    if (fb.scrollTo) {
        scrollToIndex = itemsList.findIndex(
            (v) => v.itemId === fb.scrollTo
        );
    }
    const hoveringIndex = fb.holdingShift
        ? itemsList.findIndex((v) => v.itemId === fb.hovering)
        : -1;

    const lastSelectedIndex = fb.holdingShift
        ? itemsList.findIndex((v) => v.itemId === fb.lastSelected)
        : -1;
    const context: GlobalContextType = {
        setSelected: (itemId: string, selected?: boolean) =>
            dispatch({
                type: "set_selected",
                fileId: itemId,
                selected: selected,
            }),
        visitItem: (itemId: string) => {
            const item = fb.dirMap.get(itemId);
            const jumpTo = getVisitRoute(
                fb.fbMode,
                itemId,
                fb.shareId || item.shares[0]?.shareId,
                item.isDir,
                item.displayable,
                dispatch
            );
            if (jumpTo) {
                nav(jumpTo);
            }
        },
        setHovering: (i: string) =>
            dispatch({ type: "set_hovering", hovering: i }),
        setDragging: (d: boolean) =>
            dispatch({ type: "set_dragging", dragging: d }),
        doSelectMany: () => {
            if (lastSelectedIndex === -1) {
                dispatch({
                    type: "set_selected",
                    fileId: fb.hovering,
                });
                return;
            }

            dispatch({
                type: "select_ids",
                fileIds:
                    hoveringIndex > lastSelectedIndex
                        ? itemsList
                                .slice(lastSelectedIndex, hoveringIndex + 1)
                                .map((v) => v.itemId)
                        : itemsList
                                .slice(hoveringIndex, lastSelectedIndex + 1)
                                .map((v) => v.itemId),
            });
        },
        moveSelected: (entryId: string) => {
            if (fb.dirMap.get(entryId).isDir) {
                MoveSelected(fb.selected, entryId, authHeader).then(() =>
                    dispatch({ type: "clear_selected" })
                );
            }
        },
        blockFocus: (b: boolean) =>
            dispatch({ type: "set_block_focus", block: b }),
        rename: (itemId: string, newName: string) =>
            handleRename(
                itemId,
                newName,
                fb.contentId,
                fb.selected.size,
                dispatch,
                authHeader
            ),

        setMenuOpen: (o: boolean) =>
            dispatch({ type: "set_menu_open", open: o }),
        setMenuPos: (pos: { x: number; y: number }) =>
            dispatch({ type: "set_menu_pos", pos: pos }),
        setMenuTarget: (target: string) =>
            dispatch({ type: "set_menu_target", fileId: target }),

        iconDisplay: ({ itemInfo, size }) => {
            const file = fb.dirMap.get(itemInfo.itemId);

            return IconDisplay({file: file, size: size, allowMedia: true})
        },
        setMoveDest: (itemName) =>
            dispatch({ type: "set_move_dest", fileName: itemName }),

        dragging: fb.draggingState,
        initialScrollIndex: scrollToIndex,
        allowEditing: fb.folderInfo.modifiable,
        hoveringIndex: hoveringIndex,
        lastSelectedIndex: lastSelectedIndex,
    };
    return context;
}

export function GetItemsList(fb: FbStateT, usr: UserInfoT, debouncedSearch: string) {
    if (usr.isLoggedIn === undefined) {
        return [];
    }
    let filesList = Array.from(fb.dirMap.values()).filter((val) => {
        if (!val.filename) {
            return false;
        }
        return (
            val.filename
                .toLowerCase()
                .includes(debouncedSearch.toLowerCase()) &&
            val.id !== usr.trashId
        );
    });
    const itemsList: ItemProps[] = filesList.map((v) => {
        const selected = Boolean(fb.selected.get(v.id));
        const [size, units] = humanFileSize(v.size);
        const item: ItemProps = {
            itemId: v.id,
            itemTitle: v.filename,
            itemSize: size,
            itemSizeBytes: v.size,
            itemSizeUnits: units,
            modifyDate: new Date(v.modTime),
            selected:
                (selected ? 0x1 : 0x0) |
                (fb.lastSelected === v.id ? 0x10 : 0x0),
            mediaData: v.mediaData,
            droppable: v.isDir && !selected,
            isDir: v.isDir,
            imported: v.imported,
            displayable: v.displayable,
            shares: v.shares,
        };

        return item;
    });

    try {
        itemsList.sort(getSortFunc(fb.sortFunc, fb.sortDirection));
    } catch (e) {
        console.error(e);
    }
    return itemsList;
}