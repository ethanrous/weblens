import { DragEvent, useCallback, useEffect, useState } from 'react'

import { notifications } from '@mantine/notifications'

import Upload, { fileUploadMetadata } from "../../api/Upload"
import { fileData, FileBrowserStateType, getBlankFile, FileBrowserAction, FileBrowserDispatch } from '../../types/Types'
import { CreateFolder, DeleteFiles, MoveFiles, RenameFile, downloadSingleFile, requestZipCreate } from '../../api/FileBrowserApi'
import { humanFileSize } from '../../util'


const handleSelect = (state: FileBrowserStateType, action: FileBrowserAction): FileBrowserStateType => {
    let numSelected = state.selected.size
    if (state.holdingShift && numSelected > 0 && state.lastSelected !== "") {
        const dirList = MapToList(state.dirMap)
        let startIndex = dirList.findIndex((val) => val.id === state.lastSelected)
        let endIndex = dirList.findIndex((val) => val.id === action.fileId)

        if (endIndex < startIndex) {
            [startIndex, endIndex] = [endIndex, startIndex]
        }

        for (const val of dirList.slice(startIndex, endIndex + 1)) {
            state.selected.set(val.id, true)
        }

        return { ...state, lastSelected: action.fileId }

    } else {
        // If action.selected is undefined, i.e. not passed to the request,
        // we treat that as a request to toggle the selection
        if (action.selected === undefined) {
            if (state.selected.get(action.fileId)) {
                state.selected.delete(action.fileId)
            } else {
                state.selected.set(action.fileId, true)
                return { ...state, lastSelected: action.fileId }
            }
        }
        // state.selected.get returns undefined if not selected,
        // so we not (!) it to make boolean, and not the other to match... yay javascript :/
        else if (!state.selected.get(action.fileId) === !action.selected) {
            // If the file is already in the correct state, we do nothing.
            // Specifically, we do not overwrite lastSelected
        } else {
            if (action.selected) {
                state.selected.set(action.fileId, true)
            } else {
                state.selected.delete(action.fileId)
            }
        }

        if (state.selected.size === 0) {
            return { ...state, lastSelected: "" }
        }
    }
    return { ...state }
}

function createFile(dirMap: Map<string, fileData>, user, newData: fileData, currentFolderId) {
    if (dirMap.get(newData.id)) {
        console.warn("Taking no action creating file that already exists", newData.id)
        return
    }

    if (newData.id === user.trashFolderId) {
        return
    }

    if (newData.parentFolderId === currentFolderId || user.username !== newData.owner) {
        dirMap.set(newData.id, newData)
    }
}

function updateFile(state: FileBrowserStateType, user, existingId: string, newData: fileData) {
    let existingFile: fileData = state.dirMap.get(existingId)
    if (!newData) {
        return
    }

    if (newData.id === user.trashFolderId) {
        if (state.folderInfo.id === user.trashFolderId) {
            return { ...state, folderInfo: newData, trashDirSize: newData.size }
        }

        return { ...state, trashDirSize: newData.size }
    }

    if (newData.id === user.homeFolderId) {
        if (newData.id === state.folderInfo.id) {
            return { ...state, folderInfo: newData, homeDirSize: newData.size }
        }
        return { ...state, homeDirSize: newData.size }
    }

    if (newData.id === state.folderInfo.id) {
        return { ...state, folderInfo: newData }
    }

    if (!existingFile) {
        console.warn("Not upserting file", existingId)
        return { ...state }
    }

    if (existingFile && existingFile.id !== newData.id) {
        state.dirMap.delete(existingFile.id)
    }

    if (newData.parentFolderId === state.folderInfo.id || user.username !== newData.owner) {
        state.dirMap.set(newData.id, newData)
    }

    return { ...state }
}

export const fileBrowserReducer = (state: FileBrowserStateType, action: FileBrowserAction): FileBrowserStateType => {
    switch (action.type) {
        case 'create_file': {
            for (const fileInfo of action.fileInfos) {
                if (fileInfo.filename === state.waitingForNewName) {
                    state.dirMap.delete("NEW_DIR")
                    state.dirMap.set(fileInfo.id, fileInfo)
                    return { ...state, waitingForNewName: "" }
                }
                createFile(state.dirMap, action.user, fileInfo, state.folderInfo.id)
            }
            return { ...state }
        }

        case 'update_file': {
            let update
            for (const fileInfo of action.fileInfos) {
                update = updateFile(state, action.user, fileInfo.id, fileInfo)
            }

            return update
        }

        case 'update_many': {
            const filesMeta: { fileId: string, updateInfo: fileData }[] = action.files
            for (const fileMeta of filesMeta) {
                createFile(state.dirMap, action.user, fileMeta.updateInfo, state.folderInfo.id)
            }
            return { ...state }
        }

        case 'set_folder_info': {
            if (!action.fileInfo) {
                console.error("Trying to set undefined file info")
                return { ...state }
            }
            return { ...state, folderInfo: action.fileInfo }
        }

        case 'set_parents_info': {
            return { ...state, parents: action.parents }
        }

        case 'set_loading': {
            return {
                ...state,
                loading: action.loading
            }
        }

        case 'set_scan_progress': {
            return {
                ...state,
                scanProgress: action.progress
            }
        }

        case 'set_search': {
            return {
                ...state,
                searchContent: action.search,
            }
        }

        case 'set_dragging': {
            let dragging: number

            if (!action.dragging) {
                dragging = 0
            } else if (action.dragging && !action.external) {
                dragging = 1
            } else if (action.dragging && action.external) {
                dragging = 2
            }

            return {
                ...state,
                draggingState: dragging
            }
        }

        case 'new_dir': {
            let newDir: fileData = getBlankFile()
            newDir.id = "NEW_DIR"
            newDir.isDir = true
            newDir.parentFolderId = state.folderInfo.id
            state.dirMap.set(newDir.id, newDir)
            return { ...state }
        }

        case 'set_selected': {
            return handleSelect(state, action)
        }

        case 'select_all': {
            for (const fileId of state.filesList) {
                state.selected.set(fileId, true)
            }
            return { ...state, menuOpen: false }
        }

        case 'set_block_focus': {
            return { ...state, blockFocus: action.block }
        }

        case 'clear_files': {
            state.dirMap.clear()
            state.selected.clear()

            return {
                ...state,
                parents: [],
                lastSelected: ""
            }
        }

        case 'clear_selected': {
            if (state.selected.size === 0) {
                return state
            }

            state.selected.clear()

            return {
                ...state,
                lastSelected: ""
            }
        }

        case 'delete_from_map': {
            for (const fileId of action.fileIds) {
                state.dirMap.delete(fileId)
                state.selected.delete(fileId)
            }

            return { ...state }
        }

        case 'holding_shift': {
            return {
                ...state,
                holdingShift: action.shift,
            }
        }

        case 'stop_presenting':
        case 'set_presentation': {
            if (action.presentingId) {
                state.selected.clear()
                state.selected.set(action.presentingId, true)
            }
            return {
                ...state,
                presentingId: action.presentingId
            }
        }

        case 'set_col_count': {
            return { ...state, numCols: action.numCols }
        }

        case 'set_files_list': {
            return { ...state, filesList: action.fileIds }
        }

        case 'set_menu_open': {
            return { ...state, menuOpen: action.open }
        }

        case 'set_menu_target': {
            return { ...state, menuTargetId: action.fileId }
        }

        case 'set_menu_pos': {
            return { ...state, menuPos: action.pos }
        }

        case 'presentation_next': {
            const index = state.filesList.indexOf(state.lastSelected)
            let lastSelected = state.lastSelected
            if (index + 1 < state.filesList.length) {
                state.selected.clear()
                lastSelected = state.filesList[index + 1]
                state.selected.set(lastSelected, true)
            }
            return { ...state, lastSelected: lastSelected, presentingId: lastSelected }
        }

        case 'presentation_previous': {
            const index = state.filesList.indexOf(state.lastSelected)
            let lastSelected = state.lastSelected
            if (index - 1 >= 0) {
                state.selected.clear()
                lastSelected = state.filesList[index - 1]
                state.selected.set(lastSelected, true)
            }
            return { ...state, lastSelected: lastSelected, presentingId: lastSelected }
        }

        case 'move_selection': {
            if (state.presentingId) {
                return { ...state }
            }
            let lastSelected = state.lastSelected
            const prevIndex = state.lastSelected ? state.filesList.indexOf(state.lastSelected) : -1
            let finalIndex = -1
            if (action.direction === 'ArrowDown') {
                if (prevIndex === -1) {
                    finalIndex = 0
                } else if (prevIndex + state.numCols < state.filesList.length) {
                    finalIndex = prevIndex + state.numCols
                }
            }

            else if (action.direction === 'ArrowUp') {
                if (prevIndex === -1) {
                    finalIndex = state.filesList.length - 1
                } else if (prevIndex - state.numCols >= 0) {
                    finalIndex = prevIndex - state.numCols
                }
            }

            else if (action.direction === 'ArrowLeft') {
                if (prevIndex === -1) {
                    finalIndex = state.filesList.length - 1
                }
                if (prevIndex - 1 >= 0 && prevIndex % state.numCols !== 0) {
                    finalIndex = prevIndex - 1
                }
            }

            else if (action.direction === 'ArrowRight') {
                if (prevIndex === -1) {
                    finalIndex = 0
                }
                else if (prevIndex + 1 < state.filesList.length && prevIndex % state.numCols !== state.numCols - 1) {
                    finalIndex = prevIndex + 1
                }
            }

            if (finalIndex !== -1) {
                if (!state.holdingShift) {
                    state.selected.clear()
                    state.selected.set(state.filesList[finalIndex], true)
                } else {
                    if (prevIndex < finalIndex) {
                        for (const file of state.filesList.slice(prevIndex, finalIndex + 1)) {
                            state.selected.set(file, true)
                        }
                    } else {
                        for (const file of state.filesList.slice(finalIndex, prevIndex + 1)) {
                            state.selected.set(file, true)
                        }
                    }
                }
                lastSelected = state.filesList[finalIndex]
            }

            return { ...state, lastSelected: lastSelected, presentingId: state.presentingId ? lastSelected : "" }
        }

        case 'paste_image': {
            return { ...state, pasteImg: action.img }
        }

        case 'set_scroll_to': {
            return { ...state, scrollTo: action.fileId }
        }

        // When we are waiting for a new file to be created, we don't know the id
        // so we wait to see the file with the right name to be created
        case 'set_waiting_for': {
            return { ...state, waitingForNewName: action.fileName }
        }

        case 'set_move_dest': {
            return { ...state, moveDest: action.fileName }
        }

        default: {
            console.error("Got unexpected dispatch type: ", action.type)
            notifications.show({ title: "Unexpected filebrowser dispatch", message: action.type, color: 'red' })
            return { ...state }
        }
    }
}

export const MapToList = (dirMap: Map<string, fileData>, limit?: number) => {
    const newList = Array.from(dirMap.values())

    newList.sort((a, b) => {
        // if (a.mediaData && !b.mediaData) {
        //     return -1
        // } else if (!a.mediaData && b.mediaData) {
        //     return 1
        // }

        if (a.isDir && !a.imported) {
            return -1
        } else if (b.isDir && !b.imported) {
            return 1
        }

        if (a.filename > b.filename) {
            return 1
        } else if (a.filename < b.filename) {
            return -1
        } else {
            return 0
        }
    })

    if (limit) {
        return newList.slice(0, limit)
    } else {
        return newList
    }
}

export const FilebrowserDragOver = (event: DragEvent, dispatch: FileBrowserDispatch, dragging: number) => {
    event.preventDefault()
    event.stopPropagation()

    if (event.type === "dragenter" || event.type === "dragover") {
        !dragging && dispatch({ type: "set_dragging", dragging: true, external: Boolean(event.dataTransfer.types.length) })
    } else {
        dispatch({ type: "set_dragging", dragging: false })
    }
}

export const HandleRename = (itemId: string, newName: string, folderId: string, selectedCount: number, dispatch: FileBrowserDispatch, authHeader) => {
    // When we are creating a new folder, the id is initially ""
    if (itemId === "NEW_DIR") {
        // If we do not get a new name, the rename is canceled
        if (newName === "") {
            dispatch({ type: "delete_from_map", fileIds: ["NEW_DIR"] })
        } else {
            dispatch({ type: "set_loading", loading: true })
            dispatch({ type: "set_waiting_for", fileName: newName })

            CreateFolder(folderId, newName, false, "", authHeader).then(d => {
                if (selectedCount === 0) {
                    dispatch({ type: "set_selected", fileId: d })
                }
                dispatch({ type: "set_loading", loading: false })
            })
        }
    } else {
        RenameFile(itemId, newName, authHeader)
    }
}

async function getFile(file): Promise<File> {
    try {
        const f = await file.getAsFile()
        return f
        // return new Promise((resolve, reject) => file.file(resolve, reject));
    } catch (err) {
        return await new Promise((resolve, reject) => file.file(resolve, reject));

        // return new Promise((resolve, reject) => file)
    }
}

async function addDir(fsEntry, parentFolderId: string, topFolderKey: string, rootFolderId: string, isPublic: boolean, authHeader): Promise<any> {
    return await new Promise(async (resolve: (value: fileUploadMetadata[]) => void, reject): Promise<fileUploadMetadata[]> => {
        if (fsEntry.isDirectory === true) {
            const folderId = await CreateFolder(parentFolderId, fsEntry.name, isPublic, rootFolderId, authHeader)
            if (!folderId) {
                reject()
            }
            let e: fileUploadMetadata = null
            if (!topFolderKey) {
                topFolderKey = folderId
                e = { file: fsEntry, isDir: true, folderId: folderId, parentId: rootFolderId, isTopLevel: true, topLevelParentKey: null }
            }

            let dirReader = fsEntry.createReader()
            // addDir(entry, parentFolderId, topFolderKey, rootFolderId, authHeader)
            const entriesPromise = new Promise((resolve: (value: any[]) => void, reject) => {
                let allEntries = []

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
            })

            let allResults = []
            if (e !== null) {
                allResults.push(e)
            }
            for (const entry of await entriesPromise) {
                allResults.push(...(await addDir(entry, folderId, topFolderKey, rootFolderId, isPublic, authHeader)))
            }
            resolve(allResults)

        } else {
            if (fsEntry.name === ".DS_Store") {
                resolve([])
                return
            }
            const f = await getFile(fsEntry)
            let e: fileUploadMetadata = { file: f, parentId: parentFolderId, isDir: false, isTopLevel: parentFolderId === rootFolderId, topLevelParentKey: topFolderKey }
            resolve([e])
        }
    })
}

export async function HandleDrop(entries, rootFolderId, conflictNames: string[], isPublic: boolean, shareId: string, authHeader, uploadDispatch, wsSend: (action: string, content: any) => void) {
    let files: fileUploadMetadata[] = []
    let topLevels = []
    if (entries) { // Handle Directory
        for (const entry of entries) {
            if (!entry) {
                console.error("Upload entry does not exist or is not a file")
                continue
            }
            const file = entry.webkitGetAsEntry()
            if (conflictNames.includes(file.name)) {
                notifications.show({ title: `Cannot upload "${file.name}"`, message: "A file or folder with that name already exists in this folder", autoClose: 10000, color: "red" })
                continue
            }
            topLevels.push(
                addDir(file, rootFolderId, null, rootFolderId, isPublic, authHeader).then(newFiles => { files.push(...newFiles) }).catch(r => { notifications.show({ message: String(r), color: "red" }) })
            )
        }
    }

    await Promise.all(topLevels)

    if (files.length !== 0) {
        Upload(files, isPublic, shareId, rootFolderId, authHeader, uploadDispatch, wsSend)
    }
}

export function HandleUploadButton(files: File[], parentFolderId, isPublic: boolean, shareId: string, authHeader, uploadDispatch, wsSend: (action: string, content: any) => void) {
    let uploads: fileUploadMetadata[] = []
    for (const f of files) {
        uploads.push({ file: f, parentId: parentFolderId, isDir: false, isTopLevel: true, topLevelParentKey: parentFolderId })
    }

    if (uploads.length !== 0) {
        Upload(uploads, isPublic, shareId, parentFolderId, authHeader, uploadDispatch, wsSend)
    }
}

export function downloadSelected(files: string[], dirMap: Map<string, fileData>, dispatch: FileBrowserDispatch, wsSend: (action: string, content: any) => void, authHeader) {
    let taskId: string = ""

    if (files.length === 1 && !dirMap.get(files[0]).isDir) {
        const file = dirMap.get(files[0])
        if (!file.isDir) {
            downloadSingleFile(file.id, authHeader, dispatch, file.filename)
            return
        }
    }

    requestZipCreate(files, authHeader).then(({ json, status }) => {
        if (status === 200) {
            downloadSingleFile(json.takeoutId, authHeader, dispatch, undefined, "zip")
        } else if (status === 202) {
            taskId = json.taskId
            notifications.show({ id: `zip_create_${taskId}`, message: `Requesting zip...`, autoClose: false })
            wsSend("subscribe", { subscribeType: "task", subscribeMeta: JSON.stringify({ lookingFor: ["takeoutId"] }), subscribeKey: taskId })
        } else if (status !== 0) {
            notifications.show({ title: "Failed to request takeout", message: String(json.error), color: 'red' })
        }
        dispatch({ type: "set_loading", loading: false })
    })
        .catch((r) => console.error(r))
}


export function HandleWebsocketMessage(lastMessage, userInfo, dispatch: FileBrowserDispatch, authHeader) {
    if (lastMessage) {
        let msgData = JSON.parse(lastMessage.data)
        console.log("WSRecv", msgData)
        switch (msgData.messageStatus) {
            case "file_created": {
                dispatch({ type: "create_file", fileInfos: msgData.content.map(v => v.fileInfo), user: userInfo })
                return
            }
            case "file_updated": {
                dispatch({ type: "update_file", fileInfos: msgData.content.map(v => v.fileInfo), user: userInfo })
                return
            }
            case "file_deleted": {
                dispatch({ type: "delete_from_map", fileIds: msgData.content.map(v => v.fileId) })
                return
            }
            case "scan_complete": {
                dispatch({ type: "set_loading", loading: false })
                dispatch({ type: "set_scan_progress", progress: 0 })
                return
            }
            case "scan_directory_progress": {
                dispatch({ type: "set_scan_progress", progress: (1 - (msgData.content["remainingTasks"] / msgData.content["totalTasks"])) * 100 })
                return
            }
            case "create_zip_progress": {
                let content = msgData.content[0].result
                let [speed, units] = humanFileSize(content["speedBytes"])
                notifications.update({ id: `zip_create_${msgData.subscribeKey}`, message: `Compressing ${content["totalFiles"]} files: ${content["completedFiles"]} / ${content["totalFiles"]} (${speed}${units}/s)`, autoClose: false, loading: true })
                dispatch({ type: "set_scan_progress", progress: (content["completedFiles"] / content["totalFiles"]) * 100 })
                return
            }
            case "zip_complete": {
                notifications.hide(`zip_create_${msgData.subscribeKey}`)
                downloadSingleFile(msgData.content[0].result["takeoutId"], authHeader, dispatch, undefined, "zip")
                return
            }
            case "error": {
                notifications.show({ title: "Websocket error", message: msgData.error, color: 'red' })
                return
            }
            default: {
                console.error("Could not parse websocket message type: ", msgData)
                return
            }
        }
    }
}

export const useKeyDown = (fbState: FileBrowserStateType, userInfo, dispatch: (action: FileBrowserAction) => void, authHeader, wsSend, searchRef) => {
    useEffect(() => {
        const onKeyDown = (event) => {
            if (!fbState.blockFocus) {
                if (document.activeElement !== searchRef.current && event.metaKey && event.key === 'a') {
                    event.preventDefault()
                    dispatch({ type: 'select_all' })
                } else if (!event.metaKey && ((event.which >= 49 && event.which <= 90) || event.key === "Backspace")) {
                    searchRef.current.focus()
                } else if (document.activeElement === searchRef.current && event.key === 'Escape') {
                    searchRef.current.blur()
                } else if (event.key === 'Escape') {
                    event.preventDefault()
                    if (fbState.pasteImg) {
                        dispatch({ type: "paste_image", img: null })
                    } else {
                        dispatch({ type: 'clear_selected' })
                    }
                } else if (event.key === 'Shift') {
                    dispatch({ type: 'holding_shift', shift: true })
                } else if (event.key === 'Enter' && fbState.pasteImg) {
                    if (fbState.folderInfo.id === "shared" || fbState.folderInfo.id === userInfo.trashFolderId) {
                        notifications.show({ title: "Paste blocked", message: "This folder does not allow paste-to-upload", color: 'red' })
                        return
                    }
                    UploadViaUrl(fbState.pasteImg, fbState.folderInfo.id, fbState.dirMap, authHeader, dispatch, wsSend)
                } else if (event.key === 'ArrowUp' || event.key === 'ArrowDown' || event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
                    event.preventDefault()
                    dispatch({ type: "move_selection", direction: event.key })
                } else if (event.key === ' ' && fbState.lastSelected) {
                    if (!fbState.presentingId) {
                        dispatch({ type: "set_presentation", presentingId: fbState.lastSelected })
                    } else {
                        dispatch({ type: "stop_presenting" })
                    }
                }
            } else {
                if (event.metaKey && event.key === 'a') {
                    event.stopPropagation()
                }
            }
        }

        const onKeyUp = (event) => {
            if (!fbState.blockFocus) {
                if (event.key === 'Shift') {
                    dispatch({ type: 'holding_shift', shift: false })
                }
            }
        }

        document.addEventListener('keydown', onKeyDown)
        document.addEventListener('keyup', onKeyUp)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
            document.removeEventListener('keyup', onKeyUp)
        }
    }, [fbState.blockFocus, fbState.pasteImg, dispatch, searchRef, fbState.presentingId, fbState.lastSelected])
}

export const useMousePosition = () => {
    const [
        mousePosition,
        setMousePosition
    ] = useState({ x: null, y: null })

    useEffect(() => {
        const updateMousePosition = ev => {
            setMousePosition({ x: ev.clientX, y: ev.clientY })
        }
        window.addEventListener('mousemove', updateMousePosition)
        return () => {
            window.removeEventListener('mousemove', updateMousePosition)
        }
    }, [])
    return mousePosition
}

export const usePaste = (folderId: string, userInfo, searchRef, dispatch) => {
    const handlePaste = useCallback(async (e) => {
        e.preventDefault()
        e.stopPropagation()

        const clipboardItems = typeof navigator?.clipboard?.read === 'function'
            ? await navigator.clipboard.read().catch(v => { console.error(v); notifications.show({ title: "Could not paste", message: "Does your browser block clipboard for Weblens?", color: 'red' }) })
            : e.clipboardData?.files
        if (!clipboardItems) {
            return
        }
        for (const item of clipboardItems) {
            for (const mime of item.types) {
                if (mime.startsWith("image/")) {
                    if (folderId === "shared" || folderId === userInfo.trashFolderId) {
                        notifications.show({ title: "Paste blocked", message: "This folder does not allow paste-to-upload", color: 'red' })
                        return
                    }
                    const img = await item.getType(mime)
                    dispatch({ type: 'paste_image', img: img })
                } else if (mime === "text/plain") {
                    const text = await (await item.getType("text/plain"))?.text()
                    if (!text) {
                        continue
                    }
                    searchRef.current.focus()
                    dispatch({ type: "set_search", search: text })
                } else {
                    console.error("Unknown mime", mime)
                }
            }
        }

    }, [folderId])

    useEffect(() => {
        window.addEventListener('paste', handlePaste)
        return () => {
            window.removeEventListener('paste', handlePaste)
        }
    }, [handlePaste])
}

export function deleteSelected(selectedMap: Map<string, boolean>, dirMap: Map<string, fileData>, authHeader) {
    const fileIds = Array.from(selectedMap.keys())
    DeleteFiles(fileIds, authHeader)

    // dispatch({ type: "delete_selected" })
}

export function moveSelected(selectedMap: Map<string, boolean>, destinationId: string, authHeader) {
    return MoveFiles(Array.from(selectedMap.keys()), destinationId, authHeader).catch(r => notifications.show({ title: "Failed to move files", message: String(r), color: 'red' }))
}

export async function UploadViaUrl(img: ArrayBuffer, folderId, dirMap: Map<string, fileData>, authHeader, dispatch, wsSend) {
    const names = Array.from(dirMap.values()).map((v) => v.filename)
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
        topLevelParentKey: "",
        isTopLevel: true
    }
    await Upload([meta], false, "", folderId, authHeader, () => { }, wsSend)
    dispatch({ type: "paste_image", img: null })
}

export function selectedMediaIds(dirMap: Map<string, fileData>, selectedIds: string[]): string[] {
    return selectedIds
        .map(id => dirMap.get(id)?.mediaData?.fileHash)
        .filter(v => Boolean(v))
}

export function selectedFolderIds(dirMap: Map<string, fileData>, selectedIds: string[]): string[] {
    return selectedIds.filter(id => dirMap.get(id).isDir)
}