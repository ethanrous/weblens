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
import { FbMenuModeT, WeblensFile } from '../../Files/File'
import { DraggingStateT } from '../../Files/FBTypes'
import { AuthHeaderT, TPDispatchT, UserInfoT } from '../../types/Types'

import { FbModeT, useFileBrowserStore } from './FBStateControl'

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
    } catch {
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
    authHeader: AuthHeaderT
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
        Upload(files, isPublic, shareId, rootFolderId, authHeader)
    }
}

export function HandleUploadButton(
    files: File[],
    parentFolderId: string,
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT
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
        Upload(uploads, isPublic, shareId, parentFolderId, authHeader)
    }
}

export async function downloadSelected(
    files: WeblensFile[],
    removeLoading: (loading: string) => void,
    taskProgDispatch: TPDispatchT,
    wsSend: (action: string, content) => void,
    authHeader: AuthHeaderT,
    shareId?: string
) {
    if (files.length === 1 && !files[0].IsFolder()) {
        return downloadSingleFile(
            files[0].Id(),
            authHeader,
            taskProgDispatch,
            files[0].GetFilename(),
            false,
            shareId
        )
    }

    return requestZipCreate(
        files.map((f) => f.Id()),
        shareId,
        authHeader
    ).then(({ json, status }) => {
        if (status === 200) {
            downloadSingleFile(
                json.takeoutId,
                authHeader,
                taskProgDispatch,
                json.filename,
                true,
                shareId
            )
        } else if (status === 202) {
            SubToTask(json.taskId, ['takeoutId'], wsSend)
        } else if (status !== 0) {
            console.error(json.error)
        }
        removeLoading('zipCreate')
    })
}

export const useKeyDownFileBrowser = () => {
    const blockFocus = useFileBrowserStore((state) => state.blockFocus)
    const presentingId = useFileBrowserStore((state) => state.presentingId)
    const lastSelected = useFileBrowserStore((state) => state.lastSelectedId)
    const searchContent = useFileBrowserStore((state) => state.searchContent)
    const isSearching = useFileBrowserStore((state) => state.isSearching)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    const selectAll = useFileBrowserStore((state) => state.selectAll)
    const setIsSearching = useFileBrowserStore((state) => state.setIsSearching)
    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const setHoldingShift = useFileBrowserStore(
        (state) => state.setHoldingShift
    )
    const setPresentation = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )

    useEffect(() => {
        const onKeyDown = (event) => {
            if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
                event.preventDefault()
                event.stopPropagation()
                setIsSearching(!isSearching)
            }
            if (!blockFocus) {
                if ((event.metaKey || event.ctrlKey) && event.key === 'a') {
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
                } else if (
                    event.key === 'Escape' &&
                    menuMode === FbMenuModeT.Closed &&
                    presentingId === ''
                ) {
                    event.preventDefault()
                    clearSelected()
                } else if (event.key === 'Shift') {
                    setHoldingShift(true)
                } else if (event.key === 'Enter') {
                    console.log('NOT ME RIGHT?')
                    if (!folderInfo.IsModifiable()) {
                        console.error(
                            'This folder does not allow paste-to-upload'
                        )
                        return
                    }
                    // uploadViaUrl(
                    //     fbState.pasteImg,
                    //     folderInfo.Id(),
                    //     filesList,
                    //     auth,
                    //     wsSend
                    // )
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
        presentingId,
        lastSelected,
        isSearching,
        menuMode,
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
    const setPaste = useFileBrowserStore((state) => state.setPasteImgBytes)

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
                        const img: ArrayBuffer = await (
                            await item.getType(mime)
                        ).arrayBuffer()
                        setPaste(img)
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
    authHeader: AuthHeaderT
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
    await Upload([meta], false, '', folderId, authHeader)
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
