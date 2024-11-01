import { DraggingStateT } from '@weblens/types/files/FBTypes'
import {
    FbMenuModeT,
    SelectedState,
    WeblensFile,
    WeblensFileParams,
} from '@weblens/types/files/File'
import WeblensMedia, { MediaDataT } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { FbViewOptsT, UserInfoT } from '@weblens/types/Types'
import { create, StateCreator } from 'zustand'
import { useSessionStore } from '@weblens/components/UserInfo'
import { NavigateFunction, NavigateOptions, To } from 'react-router-dom'
import { DirViewModeT } from './FileBrowserTypes'

export enum FbModeT {
    unset,
    default,
    share,
    external,
    stats,
    search,
}

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
    dirViewMode?: DirViewModeT
}) => void

export interface FileBrowserStateT {
    filesMap: Map<string, WeblensFile>
    filesLists: Map<string, WeblensFile[]>
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

    jumpTo: string
    moveDest: string

    menuMode: FbMenuModeT

    // fileInfoMenu: boolean;

    shareId: string
    contentId: string
    pastTime: Date
    pasteImgBytes: ArrayBuffer

    nav: (to: To, options?: NavigateOptions) => void
    navTimer: NodeJS.Timeout
    setNav: (nav: NavigateFunction) => void

    addToFilesMap: (file: WeblensFileParams) => void
    updateFile: (fileParams: WeblensFileParams, user: UserInfoT) => void
    replaceFile: (oldId: string, newParams: WeblensFileParams) => void
    deleteFile: (fileId: string) => void

    addLoading: (loading: string) => void
    removeLoading: (loading: string) => void
    setHoldingShift: (holdingShift: boolean) => void
    setLocationState: ({
        contentId,
        mode,
        shareId,
        pastTime,
        jumpTo,
    }: {
        contentId: string
        mode?: FbModeT
        shareId?: string
        pastTime?: Date
        jumpTo?: string
    }) => void
    clearFiles: () => void
    setSearch: (searchContent: string) => void
    setFilesData: ({
        selfInfo,
        childrenInfo,
        parentsInfo,
        mediaData,
        user,
    }: {
        selfInfo?: WeblensFileParams | WeblensFile
        childrenInfo?: WeblensFileParams[]
        parentsInfo?: WeblensFileParams[]
        mediaData?: MediaDataT[]
        user?: UserInfoT
    }) => void
    setScrollTarget: (scrollTarget: string) => void
    setSelected: (selected: string[], exclusive?: boolean) => void
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
            dirViewMode: DirViewModeT.Grid,
            sortDirection: 1,
            sortFunc: 'Name',
        }
    }
}

const nameSortFunc = (sortDirection: number) => {
    return (a: WeblensFile, b: WeblensFile) =>
        a.GetFilename().localeCompare(b.GetFilename(), 'en-US', {
            numeric: true,
        }) *
        sortDirection *
        -1
}

function getSortFunc(sortType: string, sortDirection: number) {
    let sorterBase: (a: WeblensFile, b: WeblensFile) => number
    const timeCoeff = 60000
    switch (sortType) {
        case 'Name':
            sorterBase = nameSortFunc(sortDirection)
            break
        case 'Date Modified':
            sorterBase = (a: WeblensFile, b: WeblensFile) => {
                // Round each to the nearest minute since that is what is displayed
                // in the UI. This allows sorting alpabetically when many files have
                // seemingly the same time values, and would appear in random order
                // otherwise
                const mediaMap = useMediaStore.getState().mediaMap
                const aMedia = mediaMap.get(a.GetContentId())
                const bMedia = mediaMap.get(b.GetContentId())

                let aModified = a.GetModified()
                let bModified = b.GetModified()

                if (aMedia) {
                    aModified = aMedia.GetCreateDate()
                }
                if (bMedia) {
                    bModified = bMedia.GetCreateDate()
                }

                return (
                    (Math.floor(bModified.getTime() / timeCoeff) -
                        Math.floor(aModified.getTime() / timeCoeff)) *
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

    return (a: WeblensFile, b: WeblensFile) => {
        // Get comparison of selected sort type
        const cmp = sorterBase(a, b)

        if (cmp !== 0) {
            return cmp
        }

        // If the selected sort function claims the 2 are the same,
        // fall back to sorting alphabetically
        return nameSortFunc(sortDirection)(a, b)
    }
}

function getSortedFilesLists(
    filesMap: Map<string, WeblensFile>,
    sortKey: string,
    sortDirection: number
): Map<string, WeblensFile[]> {
    const lists: Map<string, WeblensFile[]> = new Map()

    for (const file of filesMap.values()) {
        const fs = lists.get(file.ParentId()) ?? []
        fs.push(file)
        lists.set(file.ParentId(), fs)
    }

    const sortFunc = getSortFunc(sortKey, sortDirection)
    for (const pId of lists.keys()) {
        const files = lists.get(pId)
        files.sort(sortFunc)
        for (let i = 0; i < files.length; i++) {
            files[i].SetIndex(i)
        }
        lists.set(pId, files)
    }

    return lists
}

function calculateMultiSelectHint(
    state: FileBrowserStateT,
    hoveringId: string,
    select: boolean
): FileBrowserStateT {
    if (!state.lastSelectedId) {
        return state
    }

    const lastSelected = state.filesMap.get(state.lastSelectedId)

    let activeList = state.filesLists.get(lastSelected.ParentId())
    if (!activeList) {
        activeList = state.filesLists.get(state.contentId)
    }

    const hovering = state.filesMap.get(hoveringId)

    if (
        hovering &&
        lastSelected &&
        hovering.ParentId() !== lastSelected.ParentId()
    ) {
        return state
    }

    let lastSelectedIndex = lastSelected?.GetIndex()
    if (hoveringId && state.holdingShift) {
        let hoveringIndex = hovering.GetIndex()
        if (hoveringIndex < lastSelectedIndex) {
            const swap = hoveringIndex
            hoveringIndex = lastSelectedIndex
            lastSelectedIndex = swap
        }

        for (let index = lastSelectedIndex; index <= hoveringIndex; index++) {
            if (select) {
                state.selected.set(activeList[index].Id(), true)
                activeList[index].UnsetSelected(SelectedState.InRange)
                activeList[index].SetSelected(SelectedState.Selected)
            } else {
                activeList[index].SetSelected(SelectedState.InRange)
            }
        }
    } else if (hovering) {
        let hoveringIndex = hovering.GetIndex()

        if (hoveringIndex < lastSelectedIndex) {
            const swap = hoveringIndex
            hoveringIndex = lastSelectedIndex
            lastSelectedIndex = swap
        }

        for (let index = lastSelectedIndex; index <= hoveringIndex; index++) {
            activeList[index].UnsetSelected(SelectedState.InRange)
        }
    }

    return state
}

const FBStateControl: StateCreator<FileBrowserStateT, [], []> = (set) => ({
    filesMap: new Map<string, WeblensFile>(),
    selected: new Map<string, boolean>(),
    folderInfo: null,
    filesLists: new Map<string, WeblensFile[]>(),
    loading: [],
    shareId: '',
    jumpTo: '',
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
    pastTime: null,
    menuMode: 0,
    homeDirSize: 0,
    trashDirSize: 0,
    numCols: 0,
    fbMode: FbModeT.unset,
    viewOpts: loadViewOptions(),
    draggingState: DraggingStateT.NoDrag,
    menuPos: { x: 0, y: 0 },
    pasteImgBytes: null,

    nav: null,
    navTimer: null,
    setNav: (nav: NavigateFunction) =>
        set({
            nav: (to: To, options?: NavigateOptions) => {
                nav(to, options)
            },
        }),

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
                filesLists: getSortedFilesLists(
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

            const file = new WeblensFile(fileParams)
            state.filesMap.set(file.Id(), file)

            return {
                filesLists: getSortedFilesLists(
                    state.filesMap,
                    state.viewOpts.sortFunc,
                    state.viewOpts.sortDirection
                ),
                filesMap: new Map<string, WeblensFile>(state.filesMap),
            }
        })
    },

    replaceFile: (oldId: string, newInfo: WeblensFileParams) => {
        set((state) => {
            state.filesMap.delete(oldId)
            state.selected.delete(oldId)

            const newF = new WeblensFile(newInfo)

            if (state.folderInfo.Id() === newInfo.parentId) {
                if (state.lastSelectedId === oldId) {
                    state.lastSelectedId = newF.Id()
                }
                state.filesMap.set(newInfo.id, newF)
            } else if (state.lastSelectedId === oldId) {
                state.lastSelectedId = ''
                return {
                    filesMap: new Map<string, WeblensFile>(state.filesMap),
                    selected: new Map<string, boolean>(state.selected),
                    lastSelectedId: '',
                }
            }

            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(state.selected),
                lastSelectedId: state.lastSelectedId,
                filesLists: getSortedFilesLists(
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
                filesLists: getSortedFilesLists(
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

            return {
                loading: [...state.loading],
            }
        }),

    setHoldingShift: (holdingShift: boolean) =>
        set((state) => {
            state.holdingShift = holdingShift
            state = calculateMultiSelectHint(state, state.hoveringId, false)

            return {
                holdingShift: holdingShift,
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
            }
        }),

    setLocationState: ({
        contentId,
        mode,
        shareId,
        pastTime,
        jumpTo,
    }: {
        contentId: string
        mode?: FbModeT
        shareId?: string
        pastTime?: Date
        jumpTo?: string
    }) => {
        set((state) => {
            const homeId: string = useSessionStore.getState().user.homeId
            const trashId: string = useSessionStore.getState().user.trashId

            state.fbMode = mode ? mode : state.fbMode
            state.shareId = shareId ? shareId : state.shareId

            // Doing a lot of navigation (like if the user is standing on the "next" key)
            // can cause the browser to lag or hang, so we debounce the navigation ~400ms
            if (state.navTimer) {
                clearTimeout(state.navTimer)
            }
            state.navTimer = setTimeout(() => {
                const path =
                    window.location.pathname + (window.location.hash ?? '')

                let shouldBe = `/files/${contentId}`
                if (state.fbMode === FbModeT.share) {
                    shouldBe = '/files/share'
                    if (shareId) {
                        shouldBe += shareId + '/' + contentId
                    }
                }

                if (jumpTo) {
                    shouldBe += `#${jumpTo}`
                }

                if (
                    shouldBe === '/files' ||
                    shouldBe.startsWith(`/files/${homeId}`)
                ) {
                    shouldBe = shouldBe.replace(`/files/${homeId}`, '/files/home')
                } else if (shouldBe.startsWith(`/files/${trashId}`)) {
                    shouldBe = shouldBe.replace(`/files/${trashId}`, '/files/trash')
                }

                if (path !== shouldBe) {
                    state.nav(shouldBe)
                }
            }, 200)

            // If we are moving out of a folder, and no longer need the children,
            // clear the list for that folder
            if (
                contentId !== state.contentId &&
                state.filesMap.get(state.contentId)?.parentId !== contentId &&
                state.filesMap.get(contentId)?.parentId !== state.contentId
            ) {
                state.filesLists.delete(state.contentId)
            }

            return {
                contentId: contentId,
                fbMode: state.fbMode,
                shareId: state.shareId,
                lastSelectedId: jumpTo ? jumpTo : contentId,
                pastTime: pastTime ? pastTime : state.pastTime,
                jumpTo: jumpTo ? jumpTo : state.jumpTo,
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
                navTimer: state.navTimer,
            }
        })
    },

    clearFiles: () =>
        set({
            folderInfo: null,
            filesMap: new Map<string, WeblensFile>(),
            selected: new Map<string, boolean>(),
            filesLists: new Map<string, WeblensFile[]>(),
        }),

    setSearch: (search) => set({ searchContent: search }),

    setFilesData: ({
        selfInfo,
        childrenInfo,
        parentsInfo,
        mediaData,
        user,
    }: {
        selfInfo: WeblensFileParams | WeblensFile
        childrenInfo: WeblensFileParams[]
        parentsInfo: WeblensFileParams[]
        mediaData: MediaDataT[]
        user: UserInfoT
    }) => {
        const parents = parentsInfo?.map((f) => new WeblensFile(f))
        if (parents?.length > 1) {
            parents.reverse()
        }

        if (mediaData) {
            const medias = mediaData.map((m) => new WeblensMedia(m))
            useMediaStore.getState().addMedias(medias)
        }

        let selfFile: WeblensFile
        if (selfInfo && !(selfInfo instanceof WeblensFile)) {
            selfFile = new WeblensFile(selfInfo)
            if (!selfFile.IsFolder() && selfInfo.mediaData) {
                useMediaStore
                    .getState()
                    .addMedias([new WeblensMedia(selfInfo.mediaData)])
            }
        } else if (selfInfo && selfInfo instanceof WeblensFile) {
            selfFile = selfInfo
        }

        if (selfFile) {
            if (parents) {
                selfFile.SetParents(parents)
            }

            if (
                selfFile.parents.length !==
                selfFile.portablePath.split('/').length - 2
            ) {
                set({
                    folderInfo: null,
                    filesMap: new Map<string, WeblensFile>(),
                    filesLists: new Map<string, WeblensFile[]>(),
                    lastSelectedId: '',
                })
                return
            }
        }

        set((state) => {
            if (selfFile) {
                state.filesMap.set(selfFile.Id(), selfFile)
            }

            if (parents) {
                for (const p of parents) {
                    if (!state.filesMap.has(p.Id())) {
                        state.filesMap.set(p.Id(), p)
                    }
                }
            }

            for (const newFileInfo of childrenInfo ?? []) {
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
                    if (!newFileInfo.pastFile) {
                        state.trashDirSize = newFileInfo.size
                    }
                    continue
                }

                const file = new WeblensFile(newFileInfo)
                state.filesMap.set(file.Id(), file)
            }

            state.filesLists = getSortedFilesLists(
                state.filesMap,
                state.viewOpts.sortFunc,
                state.viewOpts.sortDirection
            )
            return {
                folderInfo: selfFile ?? state.folderInfo,
                loading: state.loading.filter((l) => l !== 'files'),
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
            }
        })
    },

    setScrollTarget: (scrollTarget: string) => {
        set({
            jumpTo: scrollTarget,
        })
    },

    setSelected: (selected: string[], exclusive?: boolean) => {
        set((state) => {
            if (selected.length === 0 || state.filesMap.size === 0) {
                return state
            }
            if (selected[0] === '') {
                console.error('Empty selected')
                return state
            }

            if (!state.holdingShift) {
                if (exclusive) {
                    for (const file of Array.from(state.selected.keys())) {
                        state.filesMap
                            .get(file)
                            .UnsetSelected(SelectedState.Selected)
                    }
                    state.selected = new Map<string, boolean>()
                    state.lastSelectedId = ''
                }

                for (const fId of selected) {
                    let f: WeblensFile
                    if (state.folderInfo?.Id() === fId) {
                        f = state.folderInfo
                    } else {
                        f = state.filesMap.get(fId)
                    }

                    if (f) {
                        if (f.GetSelectedState() & SelectedState.Selected) {
                            state.selected.delete(fId)
                            f.UnsetSelected(SelectedState.Selected)
                        } else {
                            state.selected.set(fId, true)
                            f.SetSelected(SelectedState.Selected)
                        }
                    } else {
                        console.error(
                            `No file in set selected: [${fId}] ${selected}`
                        )
                        return state
                    }
                }
            } else {
                state = calculateMultiSelectHint(state, state.hoveringId, true)
            }

            return {
                selected: new Map(state.selected),
                filesMap: new Map(state.filesMap),
                filesLists: new Map(state.filesLists),
                lastSelectedId: selected[selected.length - 1],
            }
        })
    },

    selectAll: () => {
        set((state) => {
            const toSelect = state.filesLists.get(state.folderInfo.Id())
            for (const file of toSelect) {
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

            if (state.viewOpts.dirViewMode === DirViewModeT.Columns) {
                const init: [string, boolean][] = state.folderInfo
                    ? [[state.folderInfo.Id(), true]]
                    : []
                return {
                    selected: new Map<string, boolean>(init),
                    lastSelectedId: state.folderInfo?.Id() ?? '',
                }
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
        set({ pastTime: pastTime })
    },

    setMoveDest: (dest: string) => {
        set({ moveDest: dest })
    },

    setHovering: (hoveringId: string) => {
        set((state) => {
            if (state.lastSelectedId) {
                state = calculateMultiSelectHint(state, hoveringId, false)
                return {
                    hoveringId: hoveringId,
                    filesLists: new Map(state.filesLists),
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
        dirViewMode: DirViewModeT
    }) => {
        set((state) => ({
            viewOpts: {
                sortFunc:
                    sortKey !== undefined ? sortKey : state.viewOpts.sortFunc,
                sortDirection:
                    sortDirection !== undefined
                        ? sortDirection
                        : state.viewOpts.sortDirection,
                dirViewMode:
                    dirViewMode !== undefined
                        ? dirViewMode
                        : state.viewOpts.dirViewMode,
            },
            filesLists: getSortedFilesLists(
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

export const useFileBrowserStore = create<FileBrowserStateT>()(FBStateControl)
