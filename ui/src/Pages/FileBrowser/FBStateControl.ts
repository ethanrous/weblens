import {
    FbMenuModeT,
    SelectedState,
    WeblensFile,
    WeblensFileParams,
} from '../../Files/File'
import { FbViewOptsT, UserInfoT } from '../../types/Types'
import { DraggingStateT } from '../../Files/FBTypes'
import { create, StateCreator } from 'zustand'
import WeblensMedia from '../../Media/Media'
import { useMediaStore } from '../../Media/MediaStateControl'

export enum FbModeT {
    unset,
    default,
    share,
    external,
    stats,
    search,
}

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

export interface FileBrowserStateT {
    filesMap: Map<string, WeblensFile>
    filesList: WeblensFile[]
    selected: Map<string, boolean>

    menuPos: { x: number; y: number }
    viewOpts: FbViewOptsT
    fbMode: FbModeT

    folderInfo: WeblensFile

    draggingState: DraggingStateT
    loading: string[]
    numCols: number

    menuTargetId: string
    presentingId: string
    hoveringId: string
    lastSelectedId: string

    searchContent: string
    isSearching: boolean
    holdingShift: boolean

    homeDirSize: number
    trashDirSize: number

    blockFocus: boolean

    scrollTo: string
    moveDest: string

    menuMode: FbMenuModeT

    // fileInfoMenu: boolean;

    shareId: string
    contentId: string
    viewingPast: Date
    pasteImgBytes: ArrayBuffer

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
    setIsSearching: (isSearching: boolean) => void
    setNumCols: (cols: number) => void
    setPasteImgBytes: (bytes: ArrayBuffer) => void

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
    let sorterBase: (a: WeblensFile, b: WeblensFile) => number
    const timeCoeff = 60000
    switch (sortType) {
        case 'Name':
            sorterBase = (a: WeblensFile, b: WeblensFile) =>
                a.GetFilename().localeCompare(b.GetFilename(), 'en-US', {
                    numeric: true,
                }) * sortDirection
            break
        case 'Date Modified':
            sorterBase = (a: WeblensFile, b: WeblensFile) => {
                // Round each to the nearest minute since that is what is displayed
                // in the UI. This allows sorting alpabetically when many files have
                // seemingly the same time values, and would appear in random order
                // otherwise
                return (
                    (Math.floor(b.GetModified().getTime() / timeCoeff) -
                        Math.floor(a.GetModified().getTime() / timeCoeff)) *
                    sortDirection
                )
            }
            break
        case 'Size':
            sorterBase = (a: WeblensFile, b: WeblensFile) =>
                (b.GetSize() - a.GetSize()) * sortDirection
            break
        default:
            console.error('Unknown file sort type:', sortType)
            return
    }

    return (a, b) => {
        // Get comparison of selected sort type
        const cmp = sorterBase(a, b)

        if (cmp !== 0) {
            return cmp
        }

        // If the selected sort function claims the 2 are the same,
        // fall back to sorting alphabetically
        return (
            a.GetFilename().localeCompare(b.GetFilename(), 'en-US', {
                numeric: true,
            }) * sortDirection
        )
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
    numCols: 0,
    fbMode: FbModeT.unset,
    viewOpts: loadViewOptions(),
    draggingState: DraggingStateT.NoDrag,
    menuPos: { x: 0, y: 0 },
    pasteImgBytes: null,

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
                if (state.lastSelectedId === oldId) {
                    state.lastSelectedId = newF.Id()
                }
                state.filesMap.set(newParams.id, newF)
            } else if (state.lastSelectedId === oldId) {
                state.lastSelectedId = ''
            }

            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(state.selected),
                lastSelectedId: state.lastSelectedId,
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

        useMediaStore.getState().addMedias(medias)

        const self = new WeblensFile(selfInfo)
        if (parents) {
            self.SetParents(parents)
        }

        set({
            folderInfo: self,
        })

        if (!self.IsFolder() && selfInfo.mediaData) {
            useMediaStore
                .getState()
                .addMedias([new WeblensMedia(selfInfo.mediaData)])
        }

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
                loading: state.loading.filter((l) => l !== 'files'),
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
                    const f = state.filesMap.get(fId)
                    if (f) {
                        if (f.GetSelectedState() & SelectedState.Selected) {
                            state.selected.delete(fId)
                            f.UnsetSelected(SelectedState.Selected)
                        } else {
                            state.selected.set(fId, true)
                            f.SetSelected(SelectedState.Selected)
                        }
                    } else {
                        return
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

    setNumCols: (cols: number) => {
        set({ numCols: cols })
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
    setIsSearching: (isSearching: boolean) => set({ isSearching }),
    setPasteImgBytes: (bytes: ArrayBuffer) => {
        set({
            pasteImgBytes: bytes,
        })
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
