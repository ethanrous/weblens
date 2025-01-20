import { IconFile, IconFolder, IconHome, IconTrash } from '@tabler/icons-react'
import {
    FileApi,
    SubToTask,
    downloadSingleFile,
} from '@weblens/api/FileBrowserApi'
import Upload, {
    FileUploadMetadata,
    UPLOAD_CHUNK_SIZE,
} from '@weblens/api/Upload'
import { WsSendT } from '@weblens/api/Websocket'
import { useSessionStore } from '@weblens/components/UserInfo'
import { FbModeT, useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import { FbMenuModeT, WeblensFile } from '@weblens/types/files/File'
import { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import User from '@weblens/types/user/User'
import { FC, useCallback, useEffect } from 'react'

import { DirViewModeT } from './FileBrowserTypes'

export function getRealId(contentId: string, mode: FbModeT, usr: User) {
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

export const handleRename = (
    itemId: string,
    newName: string,
    addLoading: (loading: string) => void,
    removeLoading: (loading: string) => void
) => {
    addLoading('renameFile')
    FileApi.updateFile(itemId, { newName: newName })
        .then(() => removeLoading('renameFile'))
        .catch(ErrorHandler)
}

function readAllFiles(
    reader: FileSystemDirectoryReader
): Promise<FileSystemEntry[]> {
    return new Promise((resolve) => {
        const allEntries = []

        function readEntriesRecursively() {
            reader.readEntries((entries) => {
                if (entries.length === 0) {
                    // No more entries, resolve the promise with all entries
                    resolve(allEntries)
                } else {
                    // Add entries to the array and call readEntriesRecursively again
                    allEntries.push(...entries)
                    readEntriesRecursively()
                }
            })
        }

        readEntriesRecursively()
    })
}

async function addDir(
    fsEntry: FileSystemEntry,
    uploadId: string,
    parentFolderId: string,
    topFolderKey: string,
    rootFolderId: string,
    isPublic: boolean,
    shareId: string
): Promise<FileUploadMetadata[]> {
    if (fsEntry.isDirectory) {
        const newDirRes = await FileApi.addFilesToUpload(uploadId, {
            newFiles: [
                {
                    isDir: true,
                    parentFolderId: parentFolderId,
                    newFileName: fsEntry.name,
                },
            ],
        }).catch((err) => {
            console.error('Failed to add files to upload', err)
        })

        if (!newDirRes) {
            throw new Error('Failed to add directory to upload')
        }

        const folderId = newDirRes.data.fileIds[0]
        if (!folderId) {
            return Promise.reject(
                new Error('Failed to create folder: no folderId')
            )
        }
        let e: FileUploadMetadata = null
        if (!topFolderKey) {
            topFolderKey = folderId
            e = {
                entry: fsEntry,
                isDir: true,
                folderId: folderId,
                parentId: rootFolderId,
                isTopLevel: true,
                topLevelParentKey: null,
            }
        }

        const allEntries = await readAllFiles(
            (fsEntry as FileSystemDirectoryEntry).createReader()
        )

        const allResults: FileUploadMetadata[] = []
        if (e !== null) {
            allResults.push(e)
        }
        for (const entry of allEntries) {
            allResults.push(
                ...(await addDir(
                    entry,
                    uploadId,
                    folderId,
                    topFolderKey,
                    rootFolderId,
                    isPublic,
                    shareId
                ))
            )
        }
        return allResults
    } else {
        if (fsEntry.name === '.DS_Store') {
            return []
        }
        const e: FileUploadMetadata = {
            entry: fsEntry,
            parentId: parentFolderId,
            isDir: false,
            isTopLevel: parentFolderId === rootFolderId,
            topLevelParentKey: topFolderKey,
        }
        return [e]
    }
}

export async function HandleDrop(
    items: DataTransferItemList,
    rootFolderId: string,
    isPublic: boolean,
    shareId: string
) {
    const files: FileUploadMetadata[] = []
    const topLevels = []
    if (!items || items.length === 0) {
        console.error('No items to upload')
        return
    }

    const entries = Array.from(items).map((i) => i.webkitGetAsEntry())

    const res = await FileApi.startUpload({
        rootFolderId: rootFolderId,
        chunkSize: UPLOAD_CHUNK_SIZE,
    }).catch((err) => {
        ErrorHandler(Error(String(err)))
    })

    if (!res) {
        return
    }

    const uploadId = res.data.uploadId

    if (entries) {
        // Handle Directory
        for (const entry of entries) {
            topLevels.push(
                addDir(
                    entry,
                    uploadId,
                    rootFolderId,
                    null,
                    rootFolderId,
                    isPublic,
                    shareId
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
        return Upload(files, isPublic, shareId, uploadId)
    }
}

export async function HandleUploadButton(
    files: File[],
    parentFolderId: string,
    isPublic: boolean,
    shareId: string
) {
    const uploads: FileUploadMetadata[] = []
    for (const f of files) {
        uploads.push({
            file: f,
            parentId: parentFolderId,
            isDir: false,
            isTopLevel: true,
            topLevelParentKey: parentFolderId,
        })
    }

    const res = await FileApi.startUpload({
        rootFolderId: parentFolderId,
        chunkSize: UPLOAD_CHUNK_SIZE,
    }).catch((err) => {
        ErrorHandler(Error(String(err)))
    })

    if (!res) {
        return
    }

    if (uploads.length !== 0) {
        Upload(uploads, isPublic, shareId, res.data.uploadId).catch(
            ErrorHandler
        )
    }
}

export async function downloadSelected(
    files: WeblensFile[],
    removeLoading: (loading: string) => void,
    wsSend: WsSendT,
    shareId?: string
) {
    if (files.length === 1 && !files[0].IsFolder()) {
        return downloadSingleFile(
            files[0].Id(),
            files[0].GetFilename(),
            false,
            shareId
        )
    }

    return FileApi.createTakeout(
        { fileIds: files.map((f) => f.Id()) },
        shareId
    ).then((res) => {
        if (res.status === 200) {
            downloadSingleFile(
                res.data.takeoutId,
                res.data.filename,
                true,
                shareId
            ).catch(ErrorHandler)
        } else if (res.status === 202) {
            SubToTask(res.data.taskId, ['takeoutId'], wsSend)
        }
        removeLoading('zipCreate')
    })
}

export const useKeyDownFileBrowser = () => {
    const blockFocus = useFileBrowserStore((state) => state.blockFocus)
    const presentingId = useFileBrowserStore((state) => state.presentingId)
    const setPresentationTarget = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )
    const lastSelected = useFileBrowserStore((state) => state.lastSelectedId)
    const searchContent = useFileBrowserStore((state) => state.searchContent)
    const isSearching = useFileBrowserStore((state) => state.isSearching)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const viewMode = useFileBrowserStore((state) => state.viewOpts.dirViewMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const filesLists = useFileBrowserStore((state) => state.filesLists)
    const mediaMap = useMediaStore((state) => state.mediaMap)

    const presentingTarget = filesMap.get(presentingId)

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
        const onKeyDown = (event: KeyboardEvent) => {
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
                    (event.key === 'ArrowLeft' || event.key === 'ArrowRight')
                ) {
                    event.preventDefault()
                    if (
                        viewMode === DirViewModeT.Columns ||
                        !presentingTarget
                    ) {
                        return
                    }
                    let direction = 0
                    if (event.key === 'ArrowLeft') {
                        direction = -1
                    } else if (event.key === 'ArrowRight') {
                        direction = 1
                    }
                    const newTarget = filesLists.get(folderInfo.Id())[
                        presentingTarget.GetIndex() + direction
                    ]
                    if (!newTarget) {
                        return
                    }
                    setPresentationTarget(newTarget.Id())

                    const onDeck = filesLists.get(folderInfo.Id())[
                        presentingTarget.GetIndex() + direction * 2
                    ]
                    if (onDeck) {
                        const m = mediaMap.get(onDeck.GetContentId())
                        if (m && !m.HasQualityLoaded(PhotoQuality.HighRes)) {
                            m.LoadBytes(PhotoQuality.HighRes).catch(
                                ErrorHandler
                            )
                        }
                    }
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

        const onKeyUp = (event: KeyboardEvent) => {
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

export const usePaste = (folderId: string, usr: User, blockFocus: boolean) => {
    const setSearch = useFileBrowserStore((state) => state.setSearch)
    const setPaste = useFileBrowserStore((state) => state.setPasteImgBytes)

    const handlePaste = useCallback(
        (e: ClipboardEvent) => {
            if (blockFocus) {
                return
            }
            e.preventDefault()
            e.stopPropagation()
            if (typeof navigator?.clipboard?.read === 'function') {
                navigator.clipboard
                    .read()
                    .then(async (items) => {
                        for (const item of items) {
                            for (const mime of item.types) {
                                if (mime.startsWith('image/')) {
                                    if (
                                        folderId === 'shared' ||
                                        folderId === usr.trashId
                                    ) {
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
                    })
                    .catch(ErrorHandler)
            } else {
                console.error('Unknown navigator clipboard type')
                // clipboardItems = e.clipboardData.files
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

export async function uploadViaUrl(
    img: ArrayBuffer,
    folderId: string,
    dirMap: Map<string, WeblensFile>
) {
    const names = Array.from(dirMap.values()).map((v) => v.GetFilename())
    let imgNumber = 1
    let imgName = `image${imgNumber}.jpg`
    while (names.includes(imgName)) {
        imgNumber++
        imgName = `image${imgNumber}.jpg`
    }

    const meta: FileUploadMetadata = {
        file: new File([img], imgName),
        isDir: false,
        parentId: folderId,
        topLevelParentKey: '',
        isTopLevel: true,
    }

    const res = await FileApi.startUpload({
        rootFolderId: folderId,
        chunkSize: UPLOAD_CHUNK_SIZE,
    }).catch((err) => {
        ErrorHandler(Error(String(err)))
    })

    if (!res) {
        return
    }

    await Upload([meta], false, '', res.data.uploadId)
}

export const historyDateTime = (timestamp: number, short: boolean = false) => {
    if (timestamp < 10000000000) {
        timestamp = timestamp * 1000
    }
    const dateObj = new Date(timestamp)
    let options: Intl.DateTimeFormatOptions
    if (short) {
        options = {
            month: 'short',
            day: 'numeric',
            minute: 'numeric',
            hour: 'numeric',
        }
    } else {
        options = {
            month: 'long',
            day: 'numeric',
            minute: 'numeric',
            hour: 'numeric',
        }
    }
    if (dateObj.getFullYear() !== new Date().getFullYear()) {
        options.year = 'numeric'
    }
    return dateObj.toLocaleDateString('en-US', options)
}

export const historyDate = (timestamp: number, short: boolean = false) => {
    if (timestamp < 10000000000) {
        timestamp = timestamp * 1000
    }
    const dateObj = new Date(timestamp)
    let options: Intl.DateTimeFormatOptions
    if (short) {
        options = {
            month: 'short',
            day: 'numeric',
        }
    } else {
        options = {
            month: 'long',
            day: 'numeric',
        }
    }
    if (dateObj.getFullYear() !== new Date().getFullYear()) {
        options.year = 'numeric'
    }
    return dateObj.toLocaleDateString('en-US', options)
}

export function filenameFromPath(pathName: string): {
    nameText: string
    StartIcon: FC<{ className: string }>
} {
    if (!pathName) {
        return { nameText: null, StartIcon: null }
    }

    pathName = pathName.slice(pathName.indexOf(':') + 1)
    const parts = pathName.split('/')

    let nameText: string = parts.pop()
    while (nameText === '' && parts.length) {
        nameText = parts.pop()
    }

    let StartIcon: FC<{ className: string }>
    if (
        nameText === useSessionStore.getState().user.username &&
        !parts.length
    ) {
        StartIcon = IconHome
        nameText = 'Home'
    } else if (nameText === '.user_trash') {
        StartIcon = IconTrash
        nameText = 'Trash'
    } else if (pathName.endsWith('/')) {
        StartIcon = IconFolder
    } else {
        StartIcon = IconFile
    }

    return { nameText, StartIcon }
}
