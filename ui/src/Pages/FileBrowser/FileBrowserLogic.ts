import { DragEvent, useCallback, useEffect, useState } from 'react'

import {
    CreateFolder,
    DeleteFiles,
    downloadSingleFile,
    moveFiles,
    RenameFile,
    requestZipCreate,
    SubToTask,
} from '../../api/FileBrowserApi'

import Upload, { fileUploadMetadata } from '../../api/Upload'
import { WeblensFile, WeblensFileParams } from '../../Files/File'
import { DraggingStateT } from '../../Files/FBTypes'
import {
    AuthHeaderT,
    MediaDispatchT,
    TPDispatchT,
    UserInfoT,
} from '../../types/Types'

import WeblensMedia from '../../Media/Media'
import {
    FBDispatchT,
    FbModeT,
    FileBrowserAction,
    useFileBrowserStore,
} from './FBStateControl'
import { WsSendT } from '../../api/Websocket'

// const handleSelect = (state: FbStateT, action: FileBrowserAction): FbStateT => {
//     if (action.fileIds) {
//         for (const fId of action.fileIds) {
//             state.dirMap.get(fId).SetSelected(SelectedState.Selected);
//             state.selected.set(fId, true);
//         }
//         state.lastSelected = action.fileIds[action.fileIds.length - 1];
//     } else {
//         const file = state.dirMap.get(action.fileId);
//         if (!file) {
//             console.error('Failed to handle select: file does not exist:  ', action.fileId);
//             return { ...state };
//         }
//         // If action.selected is undefined, i.e. not passed to the request,
//         // we treat that as a request to toggle the selection
//         if (action.selected === undefined) {
//             if (state.selected.get(action.fileId)) {
//                 file.UnsetSelected(SelectedState.Selected);
//                 state.selected.delete(action.fileId);
//             } else {
//                 file.SetSelected(SelectedState.Selected);
//                 state.selected.set(action.fileId, true);
//                 return {
//                     ...state,
//                     lastSelected: action.fileId,
//                     selected: new Map(state.selected),
//                 };
//             }
//         }
//
//         // state.selected.get returns undefined if not selected,
//         // so we not (!) it to make boolean, and again to match... yay javascript :/
//         else if (!!state.selected.get(action.fileId) === action.selected) {
//             // If the file is already in the correct state, we do nothing.
//             // Specifically, we do not overwrite lastSelected
//         } else {
//             if (action.selected) {
//                 state.lastSelected = action.fileId;
//                 file.SetSelected(SelectedState.Selected);
//                 state.selected.set(action.fileId, true);
//             } else {
//                 file.UnsetSelected(SelectedState.Selected);
//                 state.selected.delete(action.fileId);
//             }
//         }
//
//         if (state.selected.size === 0) {
//             state.lastSelected = '';
//         }
//     }
//
//     return { ...state, selected: new Map(state.selected) };
// };

export const getRealId = async (
    contentId: string,
    mode: FbModeT,
    usr: UserInfoT
) => {
    if (mode === FbModeT.stats && contentId === 'external') {
        return 'EXTERNAL'
    }

    if (contentId === 'home') {
        return usr.homeId
    } else if (contentId === 'trash') {
        return usr.trashId
    } else if (!contentId) {
        return ''
    } else {
        return contentId
    }
}

export const handleDragOver = (
    event: DragEvent,
    setDragging: (dragging: DraggingStateT) => void,
    dragging: number
) => {
    event.preventDefault()
    event.stopPropagation()

    if (event.type === 'dragenter' || event.type === 'dragover') {
        if (!dragging) {
            setDragging(DraggingStateT.ExternalDrag)
        }
    } else {
        setDragging(DraggingStateT.NoDrag)
    }
}

export const handleRename = (
    itemId: string,
    newName: string,
    addLoading: (loading: string) => void,
    removeLoading: (loading: string) => void,
    authHeader: AuthHeaderT
) => {
    addLoading('renameFile')
    RenameFile(itemId, newName, authHeader).then(() =>
        removeLoading('renameFile')
    )
}

async function getFile(file): Promise<File> {
    try {
        return await file.getAsFile()
    } catch (err) {
        console.error(file, err)
        return await new Promise((resolve, reject) =>
            file.file(resolve, reject)
        )
    }
}

async function addDir(
    fsEntry,
    parentFolderId: string,
    topFolderKey: string,
    rootFolderId: string,
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT
) {
    return await new Promise(
        // eslint-disable-next-line no-async-promise-executor
        async (
            resolve: (value: fileUploadMetadata[]) => void,
            reject
        ): Promise<fileUploadMetadata[]> => {
            if (fsEntry.isDirectory === true) {
                const folderId = await CreateFolder(
                    parentFolderId,
                    fsEntry.name,
                    [],
                    isPublic,
                    shareId,
                    authHeader
                )
                if (!folderId) {
                    reject()
                }
                let e: fileUploadMetadata = null
                if (!topFolderKey) {
                    topFolderKey = folderId
                    e = {
                        file: fsEntry,
                        isDir: true,
                        folderId: folderId,
                        parentId: rootFolderId,
                        isTopLevel: true,
                        topLevelParentKey: null,
                    }
                }

                const dirReader = fsEntry.createReader()
                // addDir(entry, parentFolderId, topFolderKey, rootFolderId, authHeader)
                const entriesPromise = new Promise(
                    (resolve: (value) => void) => {
                        const allEntries = []

                        const reader = (callback) => (entries) => {
                            if (entries.length === 0) {
                                resolve(allEntries)
                                return
                            }

                            for (const entry of entries) {
                                allEntries.push(entry)
                            }

                            if (entries.length !== 100) {
                                resolve(allEntries)
                                return
                            }
                            dirReader.readEntries(callback(callback))
                        }

                        dirReader.readEntries(reader(reader))
                    }
                )

                const allResults = []
                if (e !== null) {
                    allResults.push(e)
                }
                for (const entry of await entriesPromise) {
                    allResults.push(
                        ...(await addDir(
                            entry,
                            folderId,
                            topFolderKey,
                            rootFolderId,
                            isPublic,
                            shareId,
                            authHeader
                        ))
                    )
                }
                resolve(allResults)
            } else {
                if (fsEntry.name === '.DS_Store') {
                    resolve([])
                    return
                }
                const f = await getFile(fsEntry)
                const e: fileUploadMetadata = {
                    file: f,
                    parentId: parentFolderId,
                    isDir: false,
                    isTopLevel: parentFolderId === rootFolderId,
                    topLevelParentKey: topFolderKey,
                }
                resolve([e])
            }
        }
    )
}

export async function HandleDrop(
    entries,
    rootFolderId: string,
    conflictNames: string[],
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT,
    uploadDispatch,
    wsSend: WsSendT
) {
    const files: fileUploadMetadata[] = []
    const topLevels = []
    if (entries) {
        // Handle Directory
        for (const entry of entries) {
            if (!entry) {
                console.error('Upload entry does not exist or is not a file')
                continue
            }
            const file = entry.webkitGetAsEntry()
            if (!file) {
                console.error('Drop is not a file')
                continue
            }
            if (conflictNames.includes(file.name)) {
                continue
            }
            topLevels.push(
                addDir(
                    file,
                    rootFolderId,
                    null,
                    rootFolderId,
                    isPublic,
                    shareId,
                    authHeader
                )
                    .then((newFiles) => {
                        files.push(...newFiles)
                    })
                    .catch((r) => {
                        console.error(r)
                    })
            )
        }
    }

    await Promise.all(topLevels)

    if (files.length !== 0) {
        Upload(
            files,
            isPublic,
            shareId,
            rootFolderId,
            authHeader,
            uploadDispatch,
            wsSend
        )
    }
}

export function HandleUploadButton(
    files: File[],
    parentFolderId: string,
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT,
    uploadDispatch,
    wsSend: (action: string, content) => void
) {
    const uploads: fileUploadMetadata[] = []
    for (const f of files) {
        uploads.push({
            file: f,
            parentId: parentFolderId,
            isDir: false,
            isTopLevel: true,
            topLevelParentKey: parentFolderId,
        })
    }

    if (uploads.length !== 0) {
        Upload(
            uploads,
            isPublic,
            shareId,
            parentFolderId,
            authHeader,
            uploadDispatch,
            wsSend
        )
    }
}

export function downloadSelected(
    files: WeblensFile[],
    removeLoading: (loading: string) => void,
    taskProgDispatch: TPDispatchT,
    wsSend: (action: string, content) => void,
    authHeader: AuthHeaderT,
    shareId?: string
) {
    if (files.length === 1 && !files[0].IsFolder()) {
        downloadSingleFile(
            files[0].Id(),
            authHeader,
            taskProgDispatch,
            files[0].GetFilename(),
            shareId
        )
        return
    }

    requestZipCreate(
        files.map((f) => f.Id()),
        shareId,
        authHeader
    )
        .then(({ json, status }) => {
            if (status === 200) {
                downloadSingleFile(
                    json.takeoutId,
                    authHeader,
                    taskProgDispatch,
                    json.filename,
                    shareId
                )
            } else if (status === 202) {
                SubToTask(json.taskId, ['takeoutId'], wsSend)
            } else if (status !== 0) {
                console.error(json.error)
            }
            removeLoading('zipCreate')
        })
        .catch((r) => console.error(r))
}

export const useKeyDownFileBrowser = () => {
    const isSearching = useFileBrowserStore((state) => state.isSearching)
    const blockFocus = useFileBrowserStore((state) => state.blockFocus)
    const presentingId = useFileBrowserStore((state) => state.presentingId)
    const lastSelected = useFileBrowserStore((state) => state.lastSelectedId)
    const searchContent = useFileBrowserStore((state) => state.searchContent)
    const selectAll = useFileBrowserStore((state) => state.selectAll)
    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const setHoldingShift = useFileBrowserStore(
        (state) => state.setHoldingShift
    )
    const setPresentation = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )

    useEffect(() => {
        const onKeyDown = (event) => {
            if (isSearching) {
                return
            }
            if (!blockFocus) {
                if (event.metaKey && event.key === 'a') {
                    event.preventDefault()
                    selectAll()
                } else if (
                    !event.metaKey &&
                    (event.key === 'ArrowUp' ||
                        event.key === 'ArrowDown' ||
                        event.key === 'ArrowLeft' ||
                        event.key === 'ArrowRight')
                ) {
                    event.preventDefault()
                    console.error('move selected not impl')
                    // dispatch({
                    //     type: 'move_selection',
                    //     direction: event.key,
                    // });
                } else if (event.key === 'Escape') {
                    event.preventDefault()
                    clearSelected()
                    // if (fbState.pasteImg) {
                    //     dispatch({ type: 'paste_image', img: null });
                    // } else {
                    // }
                } else if (event.key === 'Shift') {
                    setHoldingShift(true)
                    // } else if (event.key === 'Enter' && fbState.pasteImg) {
                    //     if (fbState.folderInfo.Id() === 'shared' || fbState.folderInfo.Id() === usr.trashId) {
                    //         console.error('This folder does not allow paste-to-upload');
                    //         return;
                    //     }
                    //     uploadViaUrl(
                    //         fbState.pasteImg,
                    //         folderInfo.Id(),
                    //         fbState.dirMap,
                    //         authHeader,
                    //         dispatch,
                    //         wsSend,
                    //     );
                } else if (event.key === ' ') {
                    event.preventDefault()
                    if (lastSelected && !presentingId) {
                        setPresentation(lastSelected)
                    } else if (presentingId) {
                        setPresentation('')
                    }
                }
            }
        }

        const onKeyUp = (event) => {
            if (!blockFocus) {
                if (event.key === 'Shift') {
                    setHoldingShift(false)
                }
            }
        }

        document.addEventListener('keydown', onKeyDown)
        document.addEventListener('keyup', onKeyUp)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
            document.removeEventListener('keyup', onKeyUp)
        }
    }, [
        blockFocus,
        searchContent,
        // fbState.pasteImg,
        presentingId,
        lastSelected,
    ])
}

export const useMousePosition = () => {
    const [mousePosition, setMousePosition] = useState({ x: null, y: null })

    useEffect(() => {
        const updateMousePosition = (ev) => {
            setMousePosition({ x: ev.clientX, y: ev.clientY })
        }
        window.addEventListener('mousemove', updateMousePosition)
        return () => {
            window.removeEventListener('mousemove', updateMousePosition)
        }
    }, [])
    return mousePosition
}

export const usePaste = (
    folderId: string,
    usr: UserInfoT,
    blockFocus: boolean
) => {
    const setSearch = useFileBrowserStore((state) => state.setSearch)

    const handlePaste = useCallback(
        async (e) => {
            if (blockFocus) {
                return
            }
            e.preventDefault()
            e.stopPropagation()

            const clipboardItems =
                typeof navigator?.clipboard?.read === 'function'
                    ? await navigator.clipboard.read().catch((v) => {
                          console.error(v)
                      })
                    : e.clipboardData?.files
            if (!clipboardItems) {
                return
            }
            for (const item of clipboardItems) {
                for (const mime of item.types) {
                    if (mime.startsWith('image/')) {
                        if (folderId === 'shared' || folderId === usr.trashId) {
                            console.error(
                                'This folder does not allow paste-to-upload'
                            )
                            return
                        }
                        const img = await item.getType(mime)
                        console.error('Paste img not impl', img)
                        // dispatch({ type: 'paste_image', img: img });
                    } else if (mime === 'text/plain') {
                        const text = await (
                            await item.getType('text/plain')
                        )?.text()
                        if (!text) {
                            continue
                        }
                        setSearch(text)
                    } else {
                        console.error('Unknown mime', mime)
                    }
                }
            }
        },
        [folderId, blockFocus]
    )

    useEffect(() => {
        window.addEventListener('paste', handlePaste)
        return () => {
            window.removeEventListener('paste', handlePaste)
        }
    }, [handlePaste])
}

export function deleteSelected(
    selectedMap: Map<string, boolean>,
    dirMap: Map<string, WeblensFile>,
    authHeader: AuthHeaderT
) {
    const fileIds = Array.from(selectedMap.keys())
    DeleteFiles(fileIds, authHeader)
}

export function MoveSelected(
    selected: string[],
    destinationId: string,
    authHeader: AuthHeaderT
) {
    return moveFiles(selected, destinationId, authHeader).catch((r) =>
        console.error(r)
    )
}

export async function uploadViaUrl(
    img: ArrayBuffer,
    folderId: string,
    dirMap: Map<string, WeblensFile>,
    authHeader: AuthHeaderT,
    dispatch: (Action: FileBrowserAction) => void,
    wsSend
) {
    const names = Array.from(dirMap.values()).map((v) => v.GetFilename())
    let imgNumber = 1
    let imgName = `image${imgNumber}.jpg`
    while (names.includes(imgName)) {
        imgNumber++
        imgName = `image${imgNumber}.jpg`
    }

    const meta: fileUploadMetadata = {
        file: new File([img], imgName),
        isDir: false,
        parentId: folderId,
        topLevelParentKey: '',
        isTopLevel: true,
    }
    await Upload([meta], false, '', folderId, authHeader, () => {}, wsSend)
    dispatch({ type: 'paste_image', img: null })
}

export function selectedMediaIds(
    dirMap: Map<string, WeblensFile>,
    selectedIds: string[]
): string[] {
    return selectedIds
        .map((id) => dirMap.get(id)?.GetMediaId())
        .filter((v) => Boolean(v))
}

export function selectedFolderIds(
    dirMap: Map<string, WeblensFile>,
    selectedIds: string[]
): string[] {
    return selectedIds.filter((id) => dirMap.get(id).IsFolder())
}

export function SetFileData(
    data: {
        self?: WeblensFileParams
        children?: WeblensFileParams[]
        parents?: WeblensFileParams[]
        error?
    },
    fbDispatch: FBDispatchT,
    mediaDispatch: MediaDispatchT,
    usr: UserInfoT
) {
    if (!data) {
        console.error('Trying to set null file data')
        return
    }
    if (data.error) {
        console.error(data.error)
        return
    }

    let parents: WeblensFile[]
    if (!data.parents) {
        parents = []
    } else {
        parents = data.parents.map((f) => new WeblensFile(f))
        parents.reverse()
    }

    const children = data.children ? data.children : []

    const medias: WeblensMedia[] = []
    for (const child of children) {
        if (child.mediaData) {
            medias.push(new WeblensMedia(child.mediaData))
        }
    }

    mediaDispatch({ type: 'add_medias', medias: medias })

    const self = new WeblensFile(data.self)
    self.SetParents(parents)

    fbDispatch({
        type: 'set_folder_info',
        file: self,
        user: usr,
    })

    fbDispatch({ type: 'update_many', files: children, user: usr })
}

export const historyDate = (timestamp: number) => {
    if (timestamp < 10000000000) {
        timestamp = timestamp * 1000
    }
    const dateObj = new Date(timestamp)
    const options: Intl.DateTimeFormatOptions = {
        month: 'long',
        day: 'numeric',
        minute: 'numeric',
        hour: 'numeric',
    }
    if (dateObj.getFullYear() !== new Date().getFullYear()) {
        options.year = 'numeric'
    }
    return dateObj.toLocaleDateString('en-US', options)
}
