import { FileInfo, MediaInfo } from '@weblens/api/swag'
import { useSessionStore } from '@weblens/components/UserInfo'
import { Coordinates } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import WeblensFile, {
    FbMenuModeT,
    SelectedState,
} from '@weblens/types/files/File'
import WeblensMedia from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import User from '@weblens/types/user/User'
import { NavigateFunction, NavigateOptions, To } from 'react-router-dom'
import { StateCreator, create } from 'zustand'
import { devtools } from 'zustand/middleware'

import {
    DirViewModeT,
    FbViewOptsT,
} from '../pages/FileBrowser/FileBrowserTypes'

export enum FbModeT {
    unset,
    default,
    share,
    external,
    stats,
    search,
}

export interface MenuOptionsT {
    menuState?: FbMenuModeT
    menuPos?: Coordinates
    menuTarget?: string
}

export interface ViewOptionsT {
    sortKey?: string
    sortDirection?: number
    dirViewMode?: DirViewModeT
}

export interface SetFilesDataOpts {
    selfInfo?: FileInfo | WeblensFile
    childrenInfo?: FileInfo[]
    parentsInfo?: FileInfo[]
    mediaData?: MediaInfo[]
    overwriteContentId?: boolean
    shareId?: string
}

interface setLocationStateOptsT {
    contentId: string
    mode?: FbModeT
    shareId?: string
    pastTime?: Date
    jumpTo?: string
}

export interface FileBrowserStateT {
    filesMap: Map<string, WeblensFile>
    filesLists: Map<string, WeblensFile[]>
    selected: Map<string, boolean>

    menuPos: { x: number; y: number }
    viewOpts: FbViewOptsT
    fbMode: FbModeT
    sidebarCollapsed: boolean

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
    activeFileId: string
    pastTime: Date
    pasteImgBytes: ArrayBuffer

    nav: (to: To, options?: NavigateOptions) => void
    navTimer: NodeJS.Timeout
    setNav: (nav: NavigateFunction) => void

    addFile: (file: FileInfo) => void
    updateFile: (fileParams: FileInfo) => void
    updateFiles: (filesParams: FileInfo[]) => void
    deleteFile: (fileId: string) => void
    deleteFiles: (fileIds: string[]) => void
    sortLists: () => void
    addLoading: (loading: string) => void
    removeLoading: (loading: string) => void
    setHoldingShift: (holdingShift: boolean) => void
    setLocationState: ({
        contentId,
        mode,
        shareId,
        pastTime,
        jumpTo,
    }: setLocationStateOptsT) => void
    clearFiles: () => void
    setSearch: (searchContent: string) => void
    setFilesData: ({
        selfInfo,
        childrenInfo,
        parentsInfo,
        mediaData,
        overwriteContentId,
    }: SetFilesDataOpts) => void
    setSelected: (selected: string[], exclusive?: boolean) => void
    selectAll: () => void
    clearSelected: () => void
    setPresentationTarget: (presentingId: string) => void
    setDragging: (drag: DraggingStateT) => void
    setBlockFocus: (block: boolean) => void
    setPastTime: (pastTime: Date) => void
    setMoveDest: (dest: string) => void
    setHovering: (hovering: string) => void
    setLastSelected: (lastSelectedId: string) => void
    setSelectedMoved: (movedIds?: string[]) => void
    setIsSearching: (isSearching: boolean) => void
    setNumCols: (cols: number) => void
    setPasteImgBytes: (bytes: ArrayBuffer) => void
    setSidebarCollapsed: (c: boolean) => void

    setMenu: (opts: MenuOptionsT) => void
    setViewOptions: (opts: ViewOptionsT) => void

    reset: () => void
}

function loadViewOptions(): FbViewOptsT {
    try {
        const viewOptsString = localStorage.getItem('fbViewOpts')
        if (viewOptsString) {
            const opts = JSON.parse(viewOptsString) as FbViewOptsT
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
    state: FileBrowserStateT,
    hint: string[] = []
): FileBrowserStateT {
    const start = Date.now()
    const sortFunc = getSortFunc(
        state.viewOpts.sortFunc,
        state.viewOpts.sortDirection
    )

    if (hint?.length > 0) {
        for (const hintId of hint) {
            if (hintId === 'ROOT') {
                continue
            }
            let list = state.filesLists.get(hintId)
            if (!list) {
                console.error('Could not find list to sort:', hintId)
                continue
            }
            list = list.filter(
                (f1, i, a) => a.findIndex((f2) => f2.Id() === f1.Id()) === i
            )
            list.sort(sortFunc)
            state.filesLists.set(hintId, list)
        }
        return state
    }

    const lists: Map<string, WeblensFile[]> = new Map()

    const homeId = useSessionStore.getState().user.homeId
    for (const file of state.filesMap.values()) {
        if (file.ParentId() === homeId && file.IsTrash()) {
            continue
        }
        const fs = lists.get(file.ParentId()) ?? []
        fs.push(file)
        lists.set(file.ParentId(), fs)
    }

    for (const parentId of lists.keys()) {
        const files = lists.get(parentId)
        files.sort(sortFunc)
        for (let i = 0; i < files.length; i++) {
            files[i].SetIndex(i)
            state.filesMap.set(files[i].Id(), files[i])
        }
        lists.set(parentId, files)
    }

    state.filesLists = lists

    console.debug('Sorted files lists in ', Date.now() - start, 'ms')

    return state
}

function selectInRange(
    startFile: WeblensFile,
    endFile: WeblensFile,
    selectMode: SelectedState,
    files: WeblensFile[],
    selectedMap: Map<string, boolean>,
    remove: boolean = false
) {
    let startIndex = startFile.GetIndex()
    let endIndex = endFile.GetIndex()

    if (endIndex < startIndex) {
        // Swap the 2 if end index is before the start
        ;[startIndex, endIndex] = [endIndex, startIndex]
    }

    for (let index = startIndex; index <= endIndex; index++) {
        if (selectMode === SelectedState.Selected && !remove) {
            selectedMap.set(files[index].Id(), true)
        }

        if (remove) {
            files[index].UnsetSelected(selectMode)
        } else {
            files[index].SetSelected(selectMode)
        }
    }
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
    const lastHovering = state.filesMap.get(state.hoveringId)

    let hovering: WeblensFile
    if (hoveringId) {
        hovering = state.filesMap.get(hoveringId)
    }

    let activeList = state.filesLists.get(
        hovering ? hovering.ParentId() : lastHovering?.ParentId()
    )
    if (!activeList) {
        activeList = state.filesLists.get(state.folderInfo.Id())
    }

    if (
        hovering &&
        !select &&
        hovering?.ParentId() !== lastSelected?.ParentId()
    ) {
        return state
    }

    if (
        lastHovering &&
        lastSelected &&
        lastHovering.ParentId() === lastSelected.ParentId()
    ) {
        selectInRange(
            lastHovering,
            lastSelected,
            SelectedState.InRange,
            activeList,
            state.selected,
            true
        )
    }

    if (hovering) {
        hovering.SetSelected(SelectedState.Hovering)
    }

    if (hoveringId && state.holdingShift) {
        const selectMode = select
            ? SelectedState.Selected
            : SelectedState.InRange

        selectInRange(
            hovering,
            lastSelected,
            selectMode,
            activeList,
            state.selected
        )
    } else if (hovering) {
        // selectInRange(hovering, lastSelected, SelectedState.InRange, activeList)
    }

    return state
}

// Really slow, but can help make sure nothing is out of order
function _debug_sanity_check(state: FileBrowserStateT): FileBrowserStateT {
    return state
    // let hasError = false
    // for (const f of state.filesMap.values()) {
    //     const list = state.filesLists.get(f.parentId)
    //     if (!list) {
    //         console.trace('Sanity check failed')
    //         console.error(
    //             'Sanity check failed, could not find parent list of',
    //             f
    //         )
    //         continue
    //     }
    //     const thing = list[f.GetIndex()]
    //     if (
    //         list.length !==
    //         list.map((f) => f.Id()).filter((fId, i, a) => a.indexOf(fId) === i)
    //             .length
    //     ) {
    //         console.trace('Sanity check failed')
    //         console.error('Found duplicate in list', list)
    //
    //         continue
    //     }
    //     if (thing !== f) {
    //         hasError = true
    //         console.trace('Sanity check failed')
    //         console.error('BAD INDEX', thing, 'SECOND', f)
    //     }
    // }
    // if (!hasError) {
    //     console.trace('Sanity check passed')
    // } else {
    //     state = getSortedFilesLists(state)
    // }
    //
    // return state
}

function setLocation(
    { contentId, shareId, mode, pastTime, jumpTo }: setLocationStateOptsT,
    state: FileBrowserStateT
): FileBrowserStateT {
    console.debug(
        'SETTING LOCATION -- ',
        'contentId',
        contentId,
        'shareId',
        shareId,
        'mode',
        mode,
        'pastTime',
        pastTime,
        'jumpTo',
        jumpTo
    )

    const homeId: string = useSessionStore.getState().user.homeId
    const trashId: string = useSessionStore.getState().user.trashId

    state.fbMode = mode ?? state.fbMode
    state.shareId = shareId ?? state.shareId

    // A "0" timestamp for the pastTime means the present, so we want to write it on the state.
    // If it is undefined, we do not make an update. If it is the same as the current pastTime,
    // we do not want to make an update.
    let updatePastTime = false
    if (pastTime && state.pastTime?.getTime() !== pastTime.getTime()) {
        updatePastTime = true
        state.pastTime = pastTime
    }

    // Doing a lot of navigation (like if the user is standing on the "next" key)
    // can cause the browser to lag or hang, so we debounce the navigation ~200ms
    if (state.navTimer) {
        clearTimeout(state.navTimer)
    }
    state.navTimer = setTimeout(() => {
        const path =
            window.location.pathname +
            (window.location.hash ?? '') +
            (window.location.search ?? '')

        let shouldBe = `/files/${contentId}`
        if (state.fbMode === FbModeT.share) {
            shouldBe = '/files/share'
            if (state.shareId) {
                shouldBe += '/' + state.shareId + '/' + contentId
            } else {
                shouldBe += 'd'
            }
        }

        if (jumpTo) {
            shouldBe += `#${jumpTo}`
        }

        if (shouldBe === '/files' || shouldBe.startsWith(`/files/${homeId}`)) {
            shouldBe = shouldBe.replace(`/files/${homeId}`, '/files/home')
        } else if (shouldBe.startsWith(`/files/${trashId}`)) {
            shouldBe = shouldBe.replace(`/files/${trashId}`, '/files/trash')
        }

        if (state.pastTime && state.pastTime.getTime() !== 0) {
            shouldBe += `?past=${state.pastTime.toISOString()}`
        }

        console.debug('Should be:', shouldBe, 'Current:', path)
        if (path !== shouldBe && path !== '/login') {
            console.debug('Navigating to:', shouldBe)
            state.nav(shouldBe)
        }
    }, 200)

    if (updatePastTime) {
        // If we are updating the pastTime, do a hard reset, force
        // the filebrowser to re-build the list from the past files
        console.debug('Updating past time, clearing files')
        state = clearFiles(state)
    } else if (
        // If we are moving out of a folder, and no longer need the children,
        // clear the list for that folder
        contentId !== state.activeFileId &&
        state.filesMap.get(state.activeFileId)?.parentId !== contentId &&
        state.filesMap.get(contentId)?.parentId !== state.activeFileId
    ) {
        const list = state.filesLists.get(state.activeFileId)
        if (list) {
            for (const f of list) {
                state.filesMap.delete(f.Id())
            }
            state.filesLists.delete(state.activeFileId)
        }
    }

    state = _debug_sanity_check(state)

    console.debug('FINISH SETTING LOCATION -- ', state)

    return {
        ...state,
        activeFileId: contentId,
        // fbMode: state.fbMode,
        // shareId: state.shareId,
        // lastSelectedId: jumpTo ? jumpTo : contentId,
        jumpTo: jumpTo ? jumpTo : state.jumpTo,
        filesMap: new Map<string, WeblensFile>(state.filesMap),
        filesLists: new Map<string, WeblensFile[]>(state.filesLists),
        navTimer: state.navTimer,
    }
}

function addFile(
    state: FileBrowserStateT,
    newF: WeblensFile
): FileBrowserStateT {
    if (newF.portablePath === '' || !newF.parentId) {
        console.warn(
            'Tried to add file with no portable path or parentId',
            newF
        )
        return state
    }

    state.filesMap.set(newF.Id(), newF)

    const list = state.filesLists.get(newF.parentId) ?? []

    const index = list.findIndex((f) => f.Id() === newF.Id())
    if (index > -1) {
        list[index] = newF
    } else {
        list.push(newF)
    }

    state.filesLists.set(newF.parentId, list)

    return state
}

function purgeMovedFiles(
    state: FileBrowserStateT,
    newF: WeblensFile
): FileBrowserStateT {
    for (const child of state.folderInfo.childrenIds) {
        if (newF.childrenIds.includes(child)) {
            continue
        }

        state = deleteFile(child, state)
    }

    return state
}

function updateFileQuick(
    state: FileBrowserStateT,
    newF: WeblensFile,
    user: User
): FileBrowserStateT {
    if (newF.id === user.homeId) {
        if (state.trashDirSize === -1) {
            state.trashDirSize = user.trashSize
        }

        if (state.folderInfo && state.folderInfo?.Id() === user.homeId) {
            state.folderInfo.size = newF.size
            state = purgeMovedFiles(state, newF)
        }

        state.homeDirSize = newF.size ?? state.homeDirSize

        return state
    }

    if (newF.id === user.trashId) {
        state.trashDirSize = newF.size ?? state.homeDirSize

        if (newF.id !== state.activeFileId) {
            return state
        }
    }

    if (newF.id === state.activeFileId && state.folderInfo !== null) {
        state.folderInfo.SetSize(newF.size)

        state = purgeMovedFiles(state, newF)
    }

    const existing = state.filesMap.get(newF.Id())
    if (
        (existing && existing.parentId !== newF.parentId) ||
        newF.portablePath === ''
    ) {
        console.debug('(Re)moving file:', existing.portablePath)
        state = deleteFile(existing.Id(), state)
    } else if (existing) {
        newF.modifiable = existing.modifiable
        newF.SetSelected(existing.GetSelectedState(), true)
        newF.parents = existing.parents
        newF.index = existing.index
    }

    return addFile(state, newF)
}

function deleteFile(
    fileId: string,
    state: FileBrowserStateT
): FileBrowserStateT {
    const existing = state.filesMap.get(fileId)

    state.filesMap.delete(fileId)
    state.selected.delete(fileId)

    if (state.lastSelectedId === fileId) {
        state.lastSelectedId = ''
    }
    if (state.jumpTo === fileId) {
        state.jumpTo = ''
    }

    console.debug('Deleted file', existing?.portablePath)

    return state
}

function clearFiles(state: FileBrowserStateT): FileBrowserStateT {
    console.debug('Clearing files')
    return {
        ...state,
        activeFileId: null,
        lastSelectedId: '',
        shareId: '',
        folderInfo: null,
        filesMap: new Map<string, WeblensFile>(),
        selected: new Map<string, boolean>(),
        filesLists: new Map<string, WeblensFile[]>(),
        holdingShift: false,
    }
}

export const ShareRoot = new WeblensFile({
    id: 'shared',
    isDir: true,
})

const initState: Partial<FileBrowserStateT> = {
    filesMap: new Map<string, WeblensFile>(),
    selected: new Map<string, boolean>(),
    folderInfo: new WeblensFile({}),
    filesLists: new Map<string, WeblensFile[]>(),
    loading: [],
    shareId: undefined,
    jumpTo: '',
    activeFileId: '',
    searchContent: '',
    lastSelectedId: '',
    presentingId: '',
    moveDest: '',
    menuTargetId: '',
    hoveringId: '',
    blockFocus: false,
    isSearching: false,
    holdingShift: false,
    pastTime: new Date(0),
    menuMode: 0,
    homeDirSize: -1,
    trashDirSize: -1,
    numCols: 0,
    fbMode: FbModeT.unset,
    sidebarCollapsed: false,
    viewOpts: loadViewOptions(),
    draggingState: DraggingStateT.NoDrag,
    menuPos: { x: 0, y: 0 },
    pasteImgBytes: undefined,

    nav: undefined,
    navTimer: undefined,
} as Partial<FileBrowserStateT>

const FBStateControl: StateCreator<
    FileBrowserStateT,
    [],
    [['zustand/devtools', never]]
> = devtools((set) => ({
    ...initState,
    setNav: (nav: NavigateFunction) =>
        set({
            nav: (to: To, options?: NavigateOptions) => {
                nav(to, options)
            },
        }),

    addLoading: (loading: string) =>
        set((state: FileBrowserStateT) => {
            console.debug('Adding loading:', loading)
            state.loading.push(loading)
            return { loading: [...state.loading] }
        }),

    removeLoading: (loading: string) =>
        set((state: FileBrowserStateT) => {
            console.debug('Remove loading:', loading)
            state.loading = state.loading.filter((l) => l !== loading)

            return {
                loading: [...state.loading],
            }
        }),

    addFile: (fileParams: FileInfo) =>
        set((state: FileBrowserStateT) => {
            const newF = new WeblensFile(fileParams)
            state.filesMap.set(newF.Id(), newF)
            state = getSortedFilesLists(state)

            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
            }
        }),

    updateFile: (fileParams: FileInfo) => {
        set((state) => {
            const user = useSessionStore.getState().user
            state = updateFileQuick(state, new WeblensFile(fileParams), user)
            state = getSortedFilesLists(state)
            state = _debug_sanity_check(state)

            return {
                ...state,
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(state.selected),
            }
        })
    },

    updateFiles: (filesParams: FileInfo[]) => {
        set((state) => {
            const user = useSessionStore.getState().user
            for (const fileParams of filesParams) {
                state = updateFileQuick(
                    state,
                    new WeblensFile(fileParams),
                    user
                )
            }
            state = getSortedFilesLists(state)

            state = _debug_sanity_check(state)

            return {
                ...state,
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(state.selected),
            }
        })
    },

    deleteFile: (fileId: string) => {
        set((state) => {
            state = deleteFile(fileId, state)
            state = getSortedFilesLists(state)
            return {
                ...state,
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
                selected: new Map<string, boolean>(state.selected),
            }
        })
    },

    deleteFiles: (fileIds: string[]) => {
        set((state) => {
            for (const fileId of fileIds) {
                state = deleteFile(fileId, state)
            }
            state = getSortedFilesLists(state)
            return {
                ...state,
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
                selected: new Map<string, boolean>(state.selected),
            }
        })
    },

    setHoldingShift: (holdingShift: boolean) =>
        set((state) => {
            state.holdingShift = holdingShift
            state = calculateMultiSelectHint(state, state.hoveringId, false)

            return {
                holdingShift: holdingShift,
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
            }
        }),

    setLocationState: (opts: setLocationStateOptsT) => {
        set((state) => setLocation(opts, state))
    },

    clearFiles: () => {
        set((state) => clearFiles(state))
    },

    setSearch: (search) => set({ searchContent: search }),

    setFilesData: ({
        selfInfo,
        childrenInfo,
        parentsInfo,
        mediaData,
        overwriteContentId,
        shareId,
    }: SetFilesDataOpts) => {
        const user = useSessionStore.getState().user

        let parents = parentsInfo?.map((f) => new WeblensFile(f))
        if (parents && parents?.length > 1) {
            parents.reverse()
        }

        if (mediaData) {
            const medias = mediaData.map((m) => new WeblensMedia(m))
            useMediaStore.getState().addMedias(medias)
        }

        let selfFile: WeblensFile
        if (selfInfo && !(selfInfo instanceof WeblensFile)) {
            selfFile = new WeblensFile(selfInfo)
        } else if (selfInfo && selfInfo instanceof WeblensFile) {
            selfFile = selfInfo
        }

        set((state) => {
            if (selfFile) {
                if (
                    state.fbMode === FbModeT.share &&
                    selfFile.Id() !== ShareRoot.Id()
                ) {
                    parents = parents ?? selfFile.parents ?? []
                    if (
                        parents.length === 0 ||
                        parents[0].Id() !== ShareRoot.Id()
                    ) {
                        parents.unshift(ShareRoot)
                    }
                }

                if (parents) {
                    selfFile.SetParents(parents)
                }

                if (
                    selfFile.parents &&
                    selfFile.portablePath &&
                    selfFile.parents.length !==
                        selfFile.portablePath.split('/').length - 2
                ) {
                    console.error(
                        "Parent count doesn't match path length. Parents:",
                        selfFile.parents,
                        selfFile.parents.length,
                        'Path length:',
                        selfFile.portablePath.split('/'),
                        selfFile.portablePath.split('/').length - 2
                    )
                }
            }
            if (shareId !== undefined) {
                state.shareId = shareId
            }

            let changedFiles = false
            const prevParentId = state.folderInfo?.Id()
            if (selfFile) {
                if (selfFile.Id() === user.homeId) {
                    state.homeDirSize = selfFile.GetSize()
                }

                if (overwriteContentId) {
                    state = {
                        ...state,
                        ...setLocation(
                            {
                                contentId: selfFile.Id(),
                                jumpTo: selfFile.IsFolder()
                                    ? ''
                                    : selfFile.Id(),
                            },
                            state
                        ),
                    }
                }

                if (selfFile.Id() != 'shared') {
                    if (selfFile.Id() !== state.activeFileId) {
                        console.error(
                            `Content Id doesn't match new selfInfo, not updating state. Previous: ${state.folderInfo?.GetFilename()} (${state.activeFileId}) -- New: ${selfFile.GetFilename()} (${selfFile.Id()})`
                        )
                        return state
                    }
                    // changedFiles = true
                    state.filesMap.set(selfFile.Id(), selfFile)
                }
            }

            for (const newFileInfo of childrenInfo ?? []) {
                if (
                    newFileInfo.id === state.activeFileId &&
                    state.folderInfo !== null
                ) {
                    state.folderInfo.SetSize(newFileInfo.size)
                }

                if (newFileInfo.id === user.homeId) {
                    console.error('Got home as child, this should not happen')
                    continue
                }
                if (newFileInfo.id === user.trashId) {
                    console.debug('Got trash as child, skipping...')
                    continue
                }

                const file = new WeblensFile(newFileInfo)
                if (selfFile?.Id() === 'shared') {
                    file.parentId = selfFile.Id()
                }
                changedFiles = true
                state.filesMap.set(file.Id(), file)
            }

            state.folderInfo = selfFile ?? state.folderInfo

            if (!state.folderInfo) {
                return state
            }

            if (parents) {
                for (const p of parents) {
                    if (!state.filesMap.has(p.Id())) {
                        changedFiles = true
                        state.filesMap.set(p.Id(), p)
                    }
                }

                state.folderInfo.parents.map((p) => state.filesMap.get(p.Id()))
            }

            let filesCount = 0

            // Only count files if we are not already updating the list
            if (!changedFiles) {
                state.filesLists.forEach((list) => {
                    filesCount += list.length
                })
            }

            if (
                changedFiles ||
                prevParentId !== state.folderInfo.Id() ||
                state.filesMap.size !== filesCount
            ) {
                // const listsIds = selfFile.parents.map((p) => {
                //     return p.Id()
                // })
                // listsIds.push(selfFile.Id())
                // listsIds.push('ROOT')
                // for (const [listId, list] of state.filesLists.entries()) {
                //     if (!listsIds.includes(listId)) {
                //         for (const file of list) {
                //             state.filesMap.delete(file.Id())
                //         }
                //         state.filesLists.delete(listId)
                //     }
                // }

                state = getSortedFilesLists(state)
            }

            return {
                folderInfo: state.folderInfo,
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),

                activeFileId: state.activeFileId,
                jumpTo: state.jumpTo,
                shareId: state.shareId,
            }
        })
    },

    sortLists: () => {
        set((state) => getSortedFilesLists(state))
    },

    setSelected: (selected: string[], exclusive: boolean = false) => {
        set((state) => {
            if (selected.length === 0 || state.filesMap.size === 0) {
                console.trace('No files to select...')
                return state
            }
            if (selected[0] === '') {
                console.error('Empty selected')
                return state
            }

            if (!state.holdingShift) {
                if (state.lastSelectedId) {
                    const lastSelected = state.filesMap.get(
                        state.lastSelectedId
                    )

                    if (!lastSelected) {
                        state.lastSelectedId = ''
                    } else {
                        lastSelected.UnsetSelected(SelectedState.LastSelected)
                    }
                }

                if (exclusive) {
                    for (const fileId of Array.from(state.selected.keys())) {
                        state.filesMap
                            .get(fileId)
                            ?.SetSelected(SelectedState.NotSelected, true)
                    }
                    state.selected = new Map<string, boolean>()
                    state.lastSelectedId = ''

                    if (state.hoveringId) {
                        state.filesMap
                            .get(state.hoveringId)
                            ?.SetSelected(SelectedState.NotSelected, true)
                        state.hoveringId = ''
                    }
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
                        console.error(`No file in set selected: [${fId}]`)
                        return state
                    }
                }
            } else {
                if (state.lastSelectedId) {
                    state.filesMap
                        .get(state.lastSelectedId)
                        ?.UnsetSelected(SelectedState.LastSelected)
                }
                state = calculateMultiSelectHint(state, state.hoveringId, true)
            }

            const lastSelectedId = selected[selected.length - 1]

            state.filesMap
                .get(lastSelectedId)
                .SetSelected(SelectedState.LastSelected)

            state = _debug_sanity_check(state)

            console.debug('Last Selected:', lastSelectedId)

            return {
                selected: new Map(state.selected),
                filesMap: new Map(state.filesMap),
                filesLists: new Map(state.filesLists),
                lastSelectedId: lastSelectedId,
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
        console.debug('Clearing selected')
        set((state) => {
            for (const file of Array.from(state.selected.keys())) {
                state.filesMap.get(file).UnsetSelected(SelectedState.Selected)
            }

            if (state.lastSelectedId) {
                state.filesMap
                    .get(state.lastSelectedId)
                    ?.UnsetSelected(SelectedState.LastSelected)
            }

            if (state.viewOpts.dirViewMode === DirViewModeT.Columns) {
                const init: [string, boolean][] = state.folderInfo
                    ? [[state.folderInfo.Id(), true]]
                    : []
                return {
                    filesMap: new Map<string, WeblensFile>(state.filesMap),
                    selected: new Map<string, boolean>(init),
                    lastSelectedId: state.folderInfo?.Id() ?? '',
                }
            }

            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(),
                lastSelectedId: '',
            }
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
        set((state) => {
            const folderId = state.folderInfo?.Id()
            state = clearFiles(state)

            return {
                ...state,
                activeFileId: folderId,
                pastTime: pastTime,
            }
        })
    },

    setMoveDest: (dest: string) => {
        console.debug('Setting move dest:', dest)
        set({ moveDest: dest })
    },

    setHovering: (hoveringId: string) => {
        set((state) => {
            if (state.hoveringId === hoveringId) {
                return state
            }

            if (state.hoveringId) {
                state.filesMap
                    .get(state.hoveringId)
                    ?.UnsetSelected(SelectedState.Hovering)
            }
            if (state.lastSelectedId) {
                state = calculateMultiSelectHint(state, hoveringId, false)
                return {
                    hoveringId: hoveringId,
                    filesLists: new Map(state.filesLists),
                    selected: new Map(state.selected),
                }
            }

            return { hoveringId: hoveringId }
        })
    },

    setLastSelected: (lastSelectedId: string) => {
        set((state) => {
            if (state.lastSelectedId) {
                state.filesMap
                    .get(state.lastSelectedId)
                    .UnsetSelected(SelectedState.LastSelected)
            }

            state.filesMap
                .get(lastSelectedId)
                .SetSelected(SelectedState.LastSelected)

            console.debug('Last Selected:', lastSelectedId)

            return {
                filesMap: new Map(state.filesMap),
                selected: new Map(state.selected),
                lastSelectedId: lastSelectedId,
            }
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
                    state.filesMap
                        .get(k)
                        ?.SetSelected(SelectedState.Moved, true)
                )
                state.selected.clear()
            } else {
                movedIds.forEach((fId) => {
                    state.filesMap
                        .get(fId)
                        ?.SetSelected(SelectedState.Moved, true)
                    state.selected.delete(fId)
                })
            }

            return {
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                selected: new Map<string, boolean>(state.selected),
            }
        })
    },

    setNumCols: (cols: number) => {
        set({ numCols: cols })
    },

    setSidebarCollapsed: (c: boolean) => {
        set({ sidebarCollapsed: c })
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
        set((state) => {
            state.viewOpts.sortFunc = sortKey ?? state.viewOpts.sortFunc
            state.viewOpts.sortDirection =
                sortDirection ?? state.viewOpts.sortDirection

            state = getSortedFilesLists(state)

            return {
                viewOpts: {
                    sortFunc:
                        sortKey !== undefined
                            ? sortKey
                            : state.viewOpts.sortFunc,
                    sortDirection:
                        sortDirection !== undefined
                            ? sortDirection
                            : state.viewOpts.sortDirection,
                    dirViewMode:
                        dirViewMode !== undefined
                            ? dirViewMode
                            : state.viewOpts.dirViewMode,
                },
                filesMap: new Map<string, WeblensFile>(state.filesMap),
                filesLists: new Map<string, WeblensFile[]>(state.filesLists),
            }
        })
    },
    setIsSearching: (isSearching: boolean) => set({ isSearching }),
    setPasteImgBytes: (bytes: ArrayBuffer) => {
        set({
            pasteImgBytes: bytes,
        })
    },

    reset: () => set(initState),
}))

export const useFileBrowserStore = create<FileBrowserStateT>()(FBStateControl)
