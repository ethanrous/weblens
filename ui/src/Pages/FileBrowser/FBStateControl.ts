import {
    FbMenuModeT,
    SelectedState,
    WeblensFile,
    WeblensFileParams,
} from '../../Files/File'
import { FbViewOptsT, MediaDispatchT, UserInfoT } from '../../types/Types'
import { DraggingStateT } from '../../Files/FBTypes'
import { Dispatch } from 'react'
import { create, StateCreator } from 'zustand'
import WeblensMedia from '../../Media/Media'

export enum FbModeT {
    unset,
    default,
    share,
    external,
    stats,
    search,
}

export type FileBrowserAction = {
    type: string

    loading?: string
    fileId?: string
    fileName?: string
    search?: string
    presentingId?: string
    hoveringId?: string
    direction?: string
    realId?: string
    shareId?: string
    sortType?: string
    taskType?: string
    target?: string
    note?: string
    dirViewMode?: string

    mode?: FbModeT
    menuMode?: FbMenuModeT

    fileIds?: string[]

    dragging?: DraggingStateT
    selected?: boolean
    external?: boolean
    blockFocus?: boolean
    shift?: boolean
    open?: boolean
    isSearching?: boolean

    numCols?: number
    sortDirection?: number
    time?: number

    user?: UserInfoT

    img?: ArrayBuffer
    pos?: { x: number; y: number }

    file?: WeblensFile
    fileInfo?: WeblensFileParams
    files?: WeblensFileParams[]

    past?: Date
}

export type FBDispatchT = Dispatch<FileBrowserAction>

function fileIsInView(
    newFileInfo: WeblensFileParams,
    mode: FbModeT,
    viewingId: string,
    shareId: string,
    searchContent: string
) {
    if (mode === FbModeT.default && newFileInfo.parentFolderId !== viewingId) {
        return false
    } else if (
        mode === FbModeT.search &&
        !newFileInfo?.filename?.includes(searchContent)
    ) {
        return false
    } else if (mode === FbModeT.share) {
        // if (shareId === '' && newFileInfo.owner === usr.username) {
        //     return false
        // }
        if (viewingId === newFileInfo.id) {
            return false
        }
    }

    return true
}

// export const fileBrowserReducer = (state: FBState, action: FileBrowserAction): FBState => {
//     switch (action.type) {
//         case 'create_file': {
//             if (action.files === undefined) {
//                 return state
//             }
//
//             for (const newFileInfo of action.files) {
//                 if (!fileIsInView(newFileInfo, state.fbMode, state.contentId, state.shareId, state.searchContent)) {
//                     continue;
//                 }
//                 const file = new WeblensFile(newFileInfo);
//                 state.filesMap.set(file.Id(), file);
//             }
//             return { ...state, filesMap: new Map(state.filesMap), filesList: Array.from(state.filesMap.values()) };
//         }
//
//         case 'replace_file': {
//             state.filesMap.delete(action.fileId);
//
//             // save if it was previously selected
//             const sel = state.selected.delete(action.fileId);
//
//             if (action.fileInfo.parentFolderId !== state.folderInfo.Id()) {
//                 return {
//                     ...state,
//                     filesMap: new Map(state.filesMap),
//                     selected: new Map(state.selected),
//                     filesList: Array.from(state.filesMap.values()),
//                 };
//             }
//
//             const newFile = new WeblensFile(action.fileInfo);
//             state.filesMap.set(newFile.Id(), newFile);
//             if (sel) {
//                 state.selected.set(newFile.Id(), true);
//             }
//
//             return {
//                 ...state,
//                 filesMap: new Map(state.filesMap),
//                 selected: new Map(state.selected),
//                 filesList: Array.from(state.filesMap.values()),
//             };
//         }
//
//         case 'update_many': {
//             if (!action.files) {
//                 return state;
//             }
//             for (const newFileInfo of action.files) {
//                 if (newFileInfo.id === state.contentId && state.folderInfo !== null) {
//                     state.folderInfo.SetSize(newFileInfo.size);
//                 }
//                 if (newFileInfo.id === action.user.homeId) {
//                     state.homeDirSize = newFileInfo.size;
//                     continue;
//                 }
//                 if (newFileInfo.id === action.user.trashId) {
//                     state.trashDirSize = newFileInfo.size;
//                     continue;
//                 }
//
//                 if (!fileIsInView(newFileInfo, state.fbMode, state.contentId, state.shareId, state.searchContent)) {
//                     continue;
//                 }
//
//                 const file = new WeblensFile(newFileInfo);
//
//                 state.filesMap.set(file.Id(), file);
//             }
//             return { ...state, filesMap: new Map(state.filesMap), filesList: Array.from(state.filesMap.values()) };
//         }
//
//         case 'set_folder_info': {
//             if (!action.file || !action.file.Id() || !action.user) {
//                 console.error('Trying to set undefined file info or user');
//                 return { ...state };
//             }
//
//             return { ...state, folderInfo: action.file };
//         }
//
//         case 'add_loading': {
//             const newLoading = state.loading.filter(v => v !== action.loading);
//             newLoading.push(action.loading);
//             return {
//                 ...state,
//                 loading: newLoading,
//             };
//         }
//
//         case 'remove_loading': {
//             const newLoading = state.loading.filter(v => v !== action.loading);
//             return {
//                 ...state,
//                 loading: newLoading,
//             };
//         }
//
//         case 'set_search': {
//             return {
//                 ...state,
//                 searchContent: action.search,
//             };
//         }
//
//         case 'set_is_searching': {
//             return {
//                 ...state,
//                 isSearching: action.isSearching,
//             };
//         }
//
//         case 'set_dragging': {
//             if (!state.folderInfo || (!state.folderInfo.IsModifiable() && !action.external)) {
//                 return {
//                     ...state,
//                     draggingState: DraggingStateT.NoDrag,
//                 };
//             }
//
//             let dragging: number;
//
//             if (!action.dragging) {
//                 dragging = 0;
//             } else if (action.dragging && !action.external) {
//                 dragging = 1;
//             } else if (action.dragging && action.external) {
//                 dragging = 2;
//             }
//
//             return {
//                 ...state,
//                 draggingState: dragging,
//             };
//         }
//
//         case 'set_hovering': {
//             if (state.hoveringId === action.hoveringId) {
//                 return state;
//             }
//             state.hoveringId = action.hoveringId;
//         }
//
//         case 'set_selected': {
//             // state = handleSelect(state, action);
//             state.filesMap.get(state.hoveringId).SetSelected(SelectedState.Selected);
//             state.selected.set(state.hoveringId, true);
//             return { ...state };
//         }
//
//         case 'select_all': {
//             for (const file of state.filesList) {
//                 state.selected.set(file.Id(), true);
//             }
//             return {
//                 ...state,
//                 menuMode: FbMenuModeT.Closed,
//                 selected: new Map(state.selected),
//             };
//         }
//
//         case 'select_ids': {
//             for (const id of action.fileIds) {
//                 state.selected.set(id, true);
//             }
//             return { ...state, selected: new Map(state.selected) };
//         }
//
//         case 'set_block_focus': {
//             state.blockFocus = action.blockFocus;
//             break;
//         }
//
//         case 'clear_files': {
//             state.filesMap.clear();
//             state.selected.clear();
//             state.folderInfo = null;
//             state.parents = [];
//             state.lastSelectedId = '';
//             state.hoveringId = '';
//             state.presentingId = '';
//
//             break;
//         }
//
//         case 'clear_selected': {
//             if (state.selected.size === 0) {
//                 return state;
//             }
//
//             state.lastSelectedId = '';
//             state.selected = new Map(state.selected);
//
//             break;
//         }
//
//         case 'delete_from_map': {
//             for (const fileId of action.fileIds) {
//                 state.filesMap.delete(fileId);
//                 state.selected.delete(fileId);
//             }
//             state.filesMap = new Map(state.filesMap);
//             state.selected = new Map(state.selected);
//             break;
//         }
//
//         case 'holding_shift': {
//             state.holdingShift = action.shift;
//             break;
//         }
//
//         case 'stop_presenting':
//         case 'set_presentation': {
//             if (action.presentingId) {
//                 state.selected.clear();
//                 state.selected.set(action.presentingId, true);
//             }
//             state.presentingId = action.presentingId;
//             break;
//         }
//
//         case 'set_col_count': {
//             state.numCols = action.numCols;
//             break;
//         }
//
//         case 'set_menu_open': {
//             return {
//                 ...state,
//                 menuMode: action.menuMode,
//             };
//         }
//
//         case 'set_menu_target': {
//             return { ...state, menuTargetId: action.fileId };
//         }
//
//         case 'set_menu_pos': {
//             return { ...state, menuPos: action.pos };
//         }
//
//         case 'set_selected_moved': {
//             state.selected.forEach((_, k) => state.filesMap.get(k).SetSelected(SelectedState.Moved));
//             return { ...state };
//         }
//
//         case 'presentation_next': {
//             const index = state.filesList.findIndex(f => f.Id() === state.lastSelectedId);
//             let lastSelected = state.lastSelectedId;
//             if (index + 1 < state.filesList.length) {
//                 state.selected.clear();
//                 lastSelected = state.filesList[index + 1].Id();
//                 state.selected.set(lastSelected, true);
//             }
//             return {
//                 ...state,
//                 lastSelectedId: lastSelected,
//                 presentingId: lastSelected,
//             };
//         }
//
//         case 'presentation_previous': {
//             const index = state.filesList.findIndex(f => f.Id() === state.lastSelectedId);
//             let lastSelected = state.lastSelectedId;
//             if (index - 1 >= 0) {
//                 state.selected.clear();
//                 lastSelected = state.filesList[index - 1].Id();
//                 state.selected.set(lastSelected, true);
//             }
//             return {
//                 ...state,
//                 lastSelectedId: lastSelected,
//                 presentingId: lastSelected,
//             };
//         }
//
//         case 'move_selection': {
//             if (state.presentingId) {
//                 const presentingIndex = state.filesMap.get(state.presentingId).GetIndex();
//                 let newId;
//                 if (action.direction === 'ArrowLeft' && presentingIndex !== 0) {
//                     newId = state.filesList[presentingIndex - 1];
//                 } else if (action.direction === 'ArrowRight' && presentingIndex !== state.filesList.length - 1) {
//                     newId = state.filesList[presentingIndex + 1];
//                 } else {
//                     return state;
//                 }
//
//                 return { ...state, presentingId: newId };
//             }
//
//             let lastSelected = state.lastSelectedId;
//             const prevIndex = state.lastSelectedId
//                 ? state.filesList.findIndex(f => f.Id() === state.lastSelectedId)
//                 : -1;
//             let finalIndex = -1;
//             if (action.direction === 'ArrowDown') {
//                 if (prevIndex === -1) {
//                     finalIndex = 0;
//                 } else if (prevIndex + state.numCols < state.filesList.length) {
//                     finalIndex = prevIndex + state.numCols;
//                 }
//             } else if (action.direction === 'ArrowUp') {
//                 if (prevIndex === -1) {
//                     finalIndex = state.filesList.length - 1;
//                 } else if (prevIndex - state.numCols >= 0) {
//                     finalIndex = prevIndex - state.numCols;
//                 }
//             } else if (action.direction === 'ArrowLeft') {
//                 if (prevIndex === -1) {
//                     finalIndex = state.filesList.length - 1;
//                 }
//                 if (prevIndex - 1 >= 0 && prevIndex % state.numCols !== 0) {
//                     finalIndex = prevIndex - 1;
//                 }
//             } else if (action.direction === 'ArrowRight') {
//                 if (prevIndex === -1) {
//                     finalIndex = 0;
//                 } else if (prevIndex + 1 < state.filesList.length && prevIndex % state.numCols !== state.numCols -
// 1) { finalIndex = prevIndex + 1; } }  if (finalIndex !== -1) { if (!state.holdingShift) { state.selected.clear();
// state.selected.set(state.filesList[finalIndex].Id(), true); } else { if (prevIndex < finalIndex) { for (const file
// of state.filesList.slice(prevIndex, finalIndex + 1)) { state.selected.set(file.Id(), true); } } else { for (const
// file of state.filesList.slice(finalIndex, prevIndex + 1)) { state.selected.set(file.Id(), true); } } } lastSelected
// = state.filesList[finalIndex].Id(); }  return { ...state, lastSelectedId: lastSelected, presentingId:
// state.presentingId ? lastSelected : '', selected: new Map(state.selected), }; }  case 'paste_image': { return {
// ...state, pasteImgBytes: action.img }; }  case 'set_scroll_to': { return { ...state, scrollTo: action.fileId }; }
// case 'set_move_dest': { return { ...state, moveDest: action.fileName }; }  case 'set_location_state': { if
// (action.realId !== undefined) { state.contentId = action.realId; } if (action.mode !== undefined) { state.fbMode =
// action.mode; } if (action.shareId !== undefined) { state.shareId = action.shareId; } return { ...state, }; }  case 'set_sort': { if (action.sortType) { state.viewOpts = { ...state.viewOpts, sortFunc: action.sortType }; return { ...state }; } else if (action.sortDirection) { state.viewOpts = { ...state.viewOpts, sortDirection: action.sortDirection }; return { ...state }; } else { return { ...state }; } }  case 'set_file_view': { state.viewOpts = { ...state.viewOpts, dirViewMode: action.dirViewMode }; return { ...state }; }  case 'set_past_time': { return { ...state, viewingPast: action.past }; }  default: { console.error('Got unexpected dispatch type: ', action.type); return state; } }  return new FBState(state); };

export interface FileBrowserStateT {
    filesMap: Map<string, WeblensFile>
    filesList: WeblensFile[]
    selected: Map<string, boolean>
    // uploadMap: Map<string, boolean>;
    menuPos: { x: number; y: number }
    viewOpts: FbViewOptsT
    fbMode: FbModeT
    //
    folderInfo: WeblensFile
    // parents: WeblensFile[];
    //
    draggingState: DraggingStateT
    loading: string[]
    // numCols: number;
    //
    menuTargetId: string
    presentingId: string
    hoveringId: string
    lastSelectedId: string
    //
    searchContent: string
    isSearching: boolean
    holdingShift: boolean
    //
    homeDirSize: number
    trashDirSize: number
    //
    blockFocus: boolean
    //
    scrollTo: string
    moveDest: string
    //
    menuMode: FbMenuModeT
    //
    // fileInfoMenu: boolean;
    //
    shareId: string
    contentId: string
    viewingPast: Date
    // pasteImgBytes: ArrayBuffer;

    addToFilesMap: (file: WeblensFileParams) => void
    updateFile: (fileParams: WeblensFileParams, user: UserInfoT) => void
    replaceFile: (oldId: string, newParams: WeblensFileParams) => void
    deleteFile: (fileId: string) => void

    addLoading: (loading: string) => void
    removeLoading: (loading: string) => void
    setHoldingShift: (holdingShift: boolean) => void
    setLocationState: (
        contentId: string,
        mode: FbModeT,
        shareId: string
    ) => void
    clearFiles: () => void
    setSearch: (searchContent: string) => void
    setFilesData: (
        self: WeblensFileParams,
        children: WeblensFileParams[],
        parents: WeblensFileParams[],
        mediaDispatch: MediaDispatchT,
        user: UserInfoT
    ) => void
    setScrollTarget: (scrollTarget: string) => void
    setSelected: (selected: string[]) => void
    selectAll: () => void
    clearSelected: () => void
    setPresentationTarget: (presentingId: string) => void
    setDragging: (drag: DraggingStateT) => void
    setBlockFocus: (block: boolean) => void
    setPastTime: (pastTime: Date) => void
    setMoveDest: (dest: string) => void
    setHovering: (hovering: string) => void
    setSelectedMoved: (movedIds?: string[]) => void
    setMenu: SetMenuT
    setViewOptions: SetViewOptionsT
}

function loadViewOptions(): FbViewOptsT {
    try {
        const viewOptsString = localStorage.getItem('fbViewOpts')
        if (viewOptsString) {
            const opts = JSON.parse(viewOptsString)
            if (!opts) {
                throw new Error()
            }
            return opts
        }
        throw new Error('Could not get view opts')
    } catch {
        return {
            dirViewMode: 'Grid',
            sortDirection: 1,
            sortFunc: 'Name',
        }
    }
}

function getSortFunc(sortType: string, sortDirection: number) {
    switch (sortType) {
        case 'Name':
            return (a: WeblensFile, b: WeblensFile) =>
                a.GetFilename().localeCompare(b.GetFilename(), 'en-US', {
                    numeric: true,
                }) * sortDirection
        case 'Date Modified':
            return (a: WeblensFile, b: WeblensFile) => {
                return (
                    (b.GetModified().getTime() - a.GetModified().getTime()) *
                    sortDirection
                )
            }
        case 'Size':
            return (a: WeblensFile, b: WeblensFile) =>
                (b.GetSize() - a.GetSize()) * sortDirection
    }
}

function getSortedFilesList(
    filesMap: Map<string, WeblensFile>,
    sortKey: string,
    sortDirection: number
): WeblensFile[] {
    const sortFunc = getSortFunc(sortKey, sortDirection)
    const files = Array.from(filesMap.values())
    files.sort(sortFunc)

    files.forEach((f, i) => {
        f.SetIndex(i)
    })

    return files
}

function calculateMultiSelectHint(
    state: FileBrowserStateT,
    hoveringId: string,
    select: boolean
): FileBrowserStateT {
    if (!state.lastSelectedId) {
        return state
    }

    let lastSelectedIndex = state.filesMap.get(state.lastSelectedId).GetIndex()
    if (hoveringId && state.holdingShift) {
        let hoveringIndex = state.filesMap.get(hoveringId).GetIndex()
        if (hoveringIndex < lastSelectedIndex) {
            const swap = hoveringIndex
            hoveringIndex = lastSelectedIndex
            lastSelectedIndex = swap
        }

        for (let index = lastSelectedIndex; index <= hoveringIndex; index++) {
            if (select) {
                state.selected.set(state.filesList[index].Id(), true)
                state.filesList[index].UnsetSelected(SelectedState.InRange)
                state.filesList[index].SetSelected(SelectedState.Selected)
            } else {
                state.filesList[index].SetSelected(SelectedState.InRange)
            }
        }
    } else if (state.hoveringId) {
        let hoveringIndex = state.filesMap.get(state.hoveringId).GetIndex()

        if (hoveringIndex < lastSelectedIndex) {
            const swap = hoveringIndex
            hoveringIndex = lastSelectedIndex
            lastSelectedIndex = swap
        }

        for (let index = lastSelectedIndex; index <= hoveringIndex; index++) {
            state.filesList[index].UnsetSelected(SelectedState.InRange)
        }
    }

    return state
}

const FBStateControl: StateCreator<FileBrowserStateT, [], []> = (set) => ({
    filesMap: new Map<string, WeblensFile>(),
    selected: new Map<string, boolean>(),
    folderInfo: null,
    filesList: [],
    loading: [],
    shareId: '',
    scrollTo: '',
    contentId: '',
    searchContent: '',
    lastSelectedId: '',
    presentingId: '',
    moveDest: '',
    menuTargetId: '',
    hoveringId: '',
    blockFocus: false,
    isSearching: false,
    holdingShift: false,
    viewingPast: null,
    menuMode: 0,
    homeDirSize: 0,
    trashDirSize: 0,
    fbMode: FbModeT.unset,
    viewOpts: loadViewOptions(),
    draggingState: DraggingStateT.NoDrag,
    menuPos: { x: 0, y: 0 },

    addLoading: (loading: string) =>
        set((state: FileBrowserStateT) => {
            state.loading.push(loading)
            return { loading: [...state.loading] }
        }),

    addToFilesMap: (fileParams: WeblensFileParams) =>
        set((state: FileBrowserStateT) => {
            const newF = new WeblensFile(fileParams)
            state.filesMap.set(newF.Id(), newF)
            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesList: getSortedFilesList(
                    state.filesMap,
                    state.viewOpts.sortFunc,
                    state.viewOpts.sortDirection
                ),
            }
        }),

    updateFile: (fileParams: WeblensFileParams, user: UserInfoT) => {
        set((state) => {
            if (
                fileParams.id === state.contentId &&
                state.folderInfo !== null
            ) {
                state.folderInfo.SetSize(fileParams.size)
            }

            if (fileParams.id === user.homeId) {
                state.homeDirSize = fileParams.size
                return { homeDirSize: fileParams.size }
            }
            if (fileParams.id === user.trashId) {
                state.trashDirSize = fileParams.size
                return { trashDirSize: fileParams.size }
            }

            if (
                !fileIsInView(
                    fileParams,
                    state.fbMode,
                    state.contentId,
                    state.shareId,
                    state.searchContent
                )
            ) {
                return state
            }

            const file = new WeblensFile(fileParams)
            state.filesMap.set(file.Id(), file)

            return {
                filesList: getSortedFilesList(
                    state.filesMap,
                    state.viewOpts.sortFunc,
                    state.viewOpts.sortDirection
                ),
                filesMap: new Map<string, WeblensFile>(state.filesMap),
            }
        })
    },

    replaceFile: (oldId: string, newParams: WeblensFileParams) => {
        set((state) => {
            state.filesMap.delete(oldId)
            state.selected.delete(oldId)

            const newF = new WeblensFile(newParams)

            if (
                fileIsInView(
                    newF,
                    state.fbMode,
                    state.contentId,
                    state.shareId,
                    state.searchContent
                )
            ) {
                state.filesMap.set(newParams.id, newF)
            }

            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(state.selected),
                filesList: getSortedFilesList(
                    state.filesMap,
                    state.viewOpts.sortFunc,
                    state.viewOpts.sortDirection
                ),
            }
        })
    },

    deleteFile: (fileId: string) => {
        set((state) => {
            state.filesMap.delete(fileId)
            state.selected.delete(fileId)

            if (state.lastSelectedId === fileId) {
                state.lastSelectedId = ''
            }

            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(state.selected),
                filesList: getSortedFilesList(
                    state.filesMap,
                    state.viewOpts.sortFunc,
                    state.viewOpts.sortDirection
                ),
            }
        })
    },

    removeLoading: (loading: string) =>
        set((state: FileBrowserStateT) => {
            const index = state.loading.indexOf(loading)
            if (index != -1) {
                state.loading.splice(index, 1)
            }

            return state
        }),

    setHoldingShift: (holdingShift: boolean) =>
        set((state) => {
            state.holdingShift = holdingShift
            state = calculateMultiSelectHint(state, state.hoveringId, false)

            return {
                holdingShift: holdingShift,
                filesList: [...state.filesList],
            }
        }),

    setLocationState: (contentId: string, mode: FbModeT, shareId: string) =>
        set({
            contentId: contentId,
            fbMode: mode,
            shareId: shareId,
            lastSelectedId: '',
        }),

    clearFiles: () =>
        set({
            filesMap: new Map<string, WeblensFile>(),
            selected: new Map<string, boolean>(),
            filesList: [],
        }),

    setSearch: (search) => set({ searchContent: search }),

    setFilesData: (
        selfInfo: WeblensFileParams,
        childrenInfo: WeblensFileParams[],
        parentsInfo: WeblensFileParams[],
        mediaDispatch: MediaDispatchT,
        user: UserInfoT
    ) => {
        const parents = parentsInfo?.map((f) => new WeblensFile(f))
        if (parents?.length > 1) {
            parents.reverse()
        }

        const medias: WeblensMedia[] = []
        for (const child of childrenInfo) {
            if (child.mediaData) {
                medias.push(new WeblensMedia(child.mediaData))
            }
        }

        const children = childrenInfo.map((c) => new WeblensFile(c))

        mediaDispatch({ type: 'add_medias', medias: medias })

        const self = new WeblensFile(selfInfo)
        if (parents) {
            self.SetParents(parents)
        }

        set({
            folderInfo: self,
        })

        set((state) => {
            for (const newFileInfo of children) {
                if (
                    newFileInfo.id === state.contentId &&
                    state.folderInfo !== null
                ) {
                    state.folderInfo.SetSize(newFileInfo.size)
                }

                if (newFileInfo.id === user.homeId) {
                    state.homeDirSize = newFileInfo.size
                    continue
                }
                if (newFileInfo.id === user.trashId) {
                    state.trashDirSize = newFileInfo.size
                    continue
                }

                if (
                    !fileIsInView(
                        newFileInfo,
                        state.fbMode,
                        state.contentId,
                        state.shareId,
                        state.searchContent
                    )
                ) {
                    continue
                }

                const file = new WeblensFile(newFileInfo)

                state.filesMap.set(file.Id(), file)
            }

            state.filesList = getSortedFilesList(
                state.filesMap,
                state.viewOpts.sortFunc,
                state.viewOpts.sortDirection
            )
            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesList: getSortedFilesList(
                    state.filesMap,
                    state.viewOpts.sortFunc,
                    state.viewOpts.sortDirection
                ),
            }
        })
    },

    setScrollTarget: (scrollTarget: string) => {
        set({
            scrollTo: scrollTarget,
        })
    },

    setSelected: (selected: string[]) => {
        set((state) => {
            if (!state.holdingShift) {
                for (const fId of selected) {
                    state.selected.set(fId, true)
                    const f = state.filesMap.get(fId)
                    if (f.GetSelectedState() & SelectedState.Selected) {
                        f.UnsetSelected(SelectedState.Selected)
                    } else {
                        f.SetSelected(SelectedState.Selected)
                    }
                }
            }
            state = calculateMultiSelectHint(state, state.hoveringId, true)
            // const newSelected = new Map<string, boolean>(state.selected)
            // for (const id of selected) {
            //     state.filesMap.get(id).SetSelected(SelectedState.Selected)
            //     newSelected.set(id, true)
            // }

            return {
                selected: new Map<string, boolean>(state.selected),
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesList: getSortedFilesList(
                    state.filesMap,
                    state.viewOpts.sortFunc,
                    state.viewOpts.sortDirection
                ),
                lastSelectedId: selected[selected.length - 1],
            }
        })
    },

    selectAll: () => {
        set((state) => {
            for (const file of state.filesList) {
                state.selected.set(file.Id(), true)
                file.SetSelected(SelectedState.Selected)
            }

            return {
                selected: new Map<string, boolean>(state.selected),
                lastSelectedId: '',
            }
        })
    },

    clearSelected: () => {
        set((state) => {
            for (const file of Array.from(state.selected.keys())) {
                state.filesMap.get(file).UnsetSelected(SelectedState.Selected)
            }

            return { selected: new Map<string, boolean>(), lastSelectedId: '' }
        })
    },

    setPresentationTarget: (presentingId: string) => {
        set({ presentingId: presentingId })
    },

    setDragging: (drag: DraggingStateT) => {
        set({ draggingState: drag })
    },

    setBlockFocus: (block: boolean) => {
        set({ blockFocus: block })
    },

    setPastTime: (pastTime: Date) => {
        set({ viewingPast: pastTime })
    },

    setMoveDest: (dest: string) => {
        set({ moveDest: dest })
    },

    setHovering: (hoveringId: string) => {
        set((state) => {
            if (state.lastSelectedId) {
                calculateMultiSelectHint(state, hoveringId, false)
                return {
                    hoveringId: hoveringId,
                    filesList: [...state.filesList],
                }
            }

            return { hoveringId: hoveringId }
        })
    },

    setMenu: ({
        menuState,
        menuPos,
        menuTarget,
    }: {
        menuState?: FbMenuModeT
        menuPos?: { x: number; y: number }
        menuTarget?: string
    }) => {
        // console.trace(menuState, menuPos, menuTarget)
        set((state) => ({
            menuMode: menuState !== undefined ? menuState : state.menuMode,
            menuPos: menuPos ? menuPos : state.menuPos,
            menuTargetId:
                menuTarget !== undefined ? menuTarget : state.menuTargetId,
        }))
    },

    setSelectedMoved: (movedIds?: string[]) => {
        set((state) => {
            if (movedIds === undefined) {
                state.selected.forEach((_, k) =>
                    state.filesMap.get(k)?.SetSelected(SelectedState.Moved)
                )
            } else {
                movedIds.forEach((fId) =>
                    state.filesMap.get(fId)?.SetSelected(SelectedState.Moved)
                )
            }

            return { filesMap: new Map<string, WeblensFile>(state.filesMap) }
        })
    },

    setViewOptions: ({
        sortKey,
        sortDirection,
        dirViewMode,
    }: {
        sortKey: string
        sortDirection: number
        dirViewMode: string
    }) => {
        set((state) => ({
            viewOpts: {
                sortFunc: sortKey ? sortKey : state.viewOpts.sortFunc,
                sortDirection: sortDirection
                    ? sortDirection
                    : state.viewOpts.sortDirection,
                dirViewMode: dirViewMode
                    ? dirViewMode
                    : state.viewOpts.dirViewMode,
            },
            filesList: getSortedFilesList(
                state.filesMap,
                sortKey ? sortKey : state.viewOpts.sortFunc,
                sortDirection ? sortDirection : state.viewOpts.sortDirection
            ),
        }))
    },
})

export type SetMenuT = ({
    menuState,
    menuPos,
    menuTarget,
}: {
    menuState?: FbMenuModeT
    menuPos?: { x: number; y: number }
    menuTarget?: string
}) => void

export type SetViewOptionsT = ({
    sortKey,
    sortDirection,
    dirViewMode,
}: {
    sortKey?: string
    sortDirection?: number
    dirViewMode?: string
}) => void

export const useFileBrowserStore = create<FileBrowserStateT>()(FBStateControl)
