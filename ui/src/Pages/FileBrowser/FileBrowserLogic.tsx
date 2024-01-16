import { fileData, FileBrowserStateType, getBlankFile } from '../../types/Types'
import Upload, { fileUploadMetadata } from "../../api/Upload"
import { Dispatch, DragEvent, useEffect, useState } from 'react'
import { CreateFolder, DeleteFile, MoveFiles, downloadSingleItem, requestZipCreate } from '../../api/FileBrowserApi'
import Item from './FileItem'

import { notifications } from '@mantine/notifications'

const handleSelect = (state: FileBrowserStateType, action): FileBrowserStateType => {
    let numSelected = state.selected.size
    if (state.holdingShift && numSelected > 0 && state.lastSelected !== "") {
        const dirList = MapToList(state.dirMap)
        let startIndex = dirList.findIndex((val) => val.id === state.lastSelected)
        let endIndex = dirList.findIndex((val) => val.id === action.itemId)

        if (endIndex < startIndex) {
            [startIndex, endIndex] = [endIndex, startIndex]
        }

        for (const val of dirList.slice(startIndex, endIndex + 1)) {
            state.selected.set(val.id, true)
        }

    } else {
        // If action.selected is undefined, i.e. not passed to the request,
        // we treat that as a request to toggle the selection
        if (action.selected === undefined) {
            if (state.selected.get(action.itemId)) {
                state.selected.delete(action.itemId)
            } else {
                state.selected.set(action.itemId, true)
                return { ...state, lastSelected: action.itemId }
            }
        }
        // state.selected.get returns undefined if not selected,
        // so we not (!) it to make boolean, and not the other to match... yay javascript :/
        else if (!state.selected.get(action.itemId) === !action.selected) {
            // If the item is already in the correct state, we do nothing.
            // Specifically, we do not overwrite lastSelected
        } else {
            if (action.selected) {
                state.selected.set(action.itemId, true)
            } else {
                state.selected.delete(action.itemId)
            }
        }
    }
    return { ...state }
}

function createFile(dirMap: Map<string, fileData>, user, newData: fileData, currentFolderId) {
    console.log("NEW DATA", newData)
    if (dirMap.get(newData.id)) {
        console.error("Trying to create file with Id that already exists", newData.id)
        return
    }

    if (newData.parentFolderId === currentFolderId || user !== newData.owner) {
        dirMap.set(newData.id, newData)
    }
}

function updateFile(dirMap: Map<string, fileData>, user, existingId: string, newData: fileData, currentFolderId) {
    let existingFile: fileData = dirMap.get(existingId)
    if (!newData) {
        console.error("No newData")
        return
    }

    if (!existingFile) {
        console.warn("Not upserting file", existingId)
        return
    }

    let possibleTemplateItem: fileData = dirMap.get("TEMPLATE_NEW_FOLDER")
    if (existingFile && existingFile.id !== newData.id) {
        dirMap.delete(existingFile.id)
    }

    if (possibleTemplateItem) {
        dirMap.delete(possibleTemplateItem.id)
    }

    if (newData.parentFolderId === currentFolderId || user !== newData.owner) {
        dirMap.set(newData.id, newData)
    }
}

export const fileBrowserReducer = (state: FileBrowserStateType, action) => {
    switch (action.type) {
        case 'create_file': {
            createFile(state.dirMap, action.user, action.fileInfo, state.folderInfo.id)
            return { ...state }
        }

        case 'update_file': {
            updateFile(state.dirMap, action.user, action.fileInfo.id, action.fileInfo, state.folderInfo.id)
            return { ...state }
        }

        case 'update_many': {
            const itemsMeta: { itemId: string, updateInfo: fileData }[] = action.items
            for (const itemMeta of itemsMeta) {
                createFile(state.dirMap, action.user, itemMeta.updateInfo, state.folderInfo.id)
            }
            return { ...state }
        }

        case 'set_folder_info': {
            return { ...state, folderInfo: action.folderInfo }
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
            newDir.id = "TEMPLATE_NEW_FOLDER"
            newDir.isDir = true
            state.dirMap.set(newDir.id, newDir)
            return { ...state }
        }

        case 'set_selected': {
            return handleSelect(state, action)
        }

        case 'select_all': {
            for (const item of state.dirMap.values()) {
                if (item.filename.toLowerCase().includes(state.searchContent.toLowerCase())) {
                    state.selected.set(item.id, true)
                }
            }
            return {
                ...state,
                lastSelected: ""
            }
        }

        case 'set_block_focus': {
            return { ...state, blockFocus: action.block }
        }

        case 'set_hovering': {
            if (state.draggingState === 0 || state.hovering === action.itemId) {
                return { ...state }
            }

            return {
                ...state,
                hovering: action.itemId
            }
        }

        case 'clear_items': {
            state.dirMap.clear()
            state.selected.clear()

            return {
                ...state,
                folderInfo: {},
                parents: [],
                lastSelected: ""
            }
        }

        case 'clear_selected': {

            state.selected.clear()

            return {
                ...state,
                lastSelected: ""
            }
        }

        case 'delete_from_map': {
            state.dirMap.delete(action.itemId)
            state.selected.delete(action.itemId)

            return { ...state }
        }

        case 'holding_shift': {
            return {
                ...state,
                holdingShift: action.shift,
            }
        }

        case 'set_presentation': {
            return {
                ...state,
                presentingId: action.presentingId
            }
        }

        default: {
            console.error("Got unexpected dispatch type: ", action.type)
            notifications.show({ title: "Unexpected filebrowser dispatch", message: action.type, color: 'red' })
            return { ...state }
        }
    }
}

export const MapToList = (dirMap: Map<string, fileData>) => {
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

    return newList
}

export const HandleDrag = (event: DragEvent, dispatch: Dispatch<any>, dragging: number) => {
    event.preventDefault()
    event.stopPropagation()
    if (event.type === "dragenter" || event.type === "dragover") {
        !dragging && dispatch({ type: "set_dragging", dragging: true, external: true })

    } else {
        dispatch({ type: "set_dragging", dragging: false })
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

async function addDir(fsEntry, parentFolderId: string, topFolderKey: string, rootFolderId: string, authHeader): Promise<any> {
    return await new Promise(async (resolve: (value: fileUploadMetadata[]) => void, reject): Promise<fileUploadMetadata[]> => {
        if (fsEntry.isDirectory === true) {
            const folderId = await CreateFolder(parentFolderId, fsEntry.name, authHeader)
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
                allResults.push(...(await addDir(entry, folderId, topFolderKey, rootFolderId, authHeader)))
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

export async function HandleDrop(items, rootFolderId, dirMap, authHeader, uploadDispatch, dispatch, wsSend) {
    let files: fileUploadMetadata[] = []
    let topLevels = []
    if (items) { // Handle Directory
        const names = new Map<string, boolean>()
        Array.from(dirMap.values()).map((value: fileData) => names.set(value.filename, true))

        for (const item of items) {
            if (!item) {
                console.error("Upload item does not exist or is not a file")
                continue
            }
            const file = item.webkitGetAsEntry()
            if (names.get(file.name)) {
                notifications.show({ title: `Cannot upload "${file.name}"`, message: "A file or folder with that name already exists in this folder", autoClose: 10000, color: "red" })
                continue
            }
            topLevels.push(
                addDir(file, rootFolderId, null, rootFolderId, authHeader).then(newItems => { files.push(...newItems) }).catch((r) => { if (r) { notifications.show({ message: r, color: "red" }) } })
            )
        }
    }

    await Promise.all(topLevels)

    if (files.length !== 0) {
        Upload(files, rootFolderId, authHeader, uploadDispatch, dispatch, wsSend)
    }
}

export function HandleUploadButton(files: File[], parentFolderId, authHeader, uploadDispatch, dispatch, wsSend) {
    let uploads: fileUploadMetadata[] = []
    for (const f of files) {
        uploads.push({ file: f, parentId: parentFolderId, isDir: false, isTopLevel: true, topLevelParentKey: parentFolderId })
    }

    if (uploads.length !== 0) {
        Upload(uploads, parentFolderId, authHeader, uploadDispatch, dispatch, wsSend)
    }
}

export function downloadSelected(selectedMap: Map<string, boolean>, dirMap: Map<string, fileData>, folderId: string, dispatch, wsSend, authHeader) {
    const items = Array.from(selectedMap.keys())

    let taskId: string = ""

    if (items.length === 1 && !dirMap.get(items[0]).isDir) {
        const item = dirMap.get(items[0])
        if (!item.isDir) {
            downloadSingleItem(item.id, authHeader, dispatch, item.filename)
            return
        }
    }

    requestZipCreate(items, authHeader).then(({ json, status }) => {
        if (status === 200) {
            downloadSingleItem(json.takeoutId, authHeader, dispatch, undefined, "zip")
            } else {
            taskId = json.taskId
            notifications.show({ id: `zip_create_${taskId}`, message: `Requesting zip...`, autoClose: false })
            wsSend(JSON.stringify({ req: "subscribe", content: { subType: "task", lookingFor: ["takeoutId"], taskId: taskId }, error: null }))
            }
            dispatch({ type: "set_loading", loading: false })
        })
        .catch((r) => console.error(r))
}


export function HandleWebsocketMessage(lastMessage, username, dispatch, authHeader) {
    if (lastMessage) {
        let msgData = JSON.parse(lastMessage.data)
        console.log("WSmsg", msgData)
        switch (msgData.messageStatus) {
            case "file_created": {
                dispatch({ type: "create_file", fileInfo: msgData.content.fileInfo, user: username })
                return
            }
            case "file_updated": {
                dispatch({ type: "update_file", fileInfo: msgData.content.fileInfo, updateInfo: msgData.content.updateInfo, user: username })
                return
            }
            case "file_deleted": {
                dispatch({ type: "delete_from_map", itemId: msgData.content.itemId })
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
                notifications.update({ id: `zip_create_${msgData.subscribeKey}`, message: `Compressing ${msgData.content["totalFiles"]} files: ${msgData.content["completedFiles"]} / ${msgData.content["totalFiles"]}`, autoClose: false, loading: true })
                dispatch({ type: "set_scan_progress", progress: (msgData.content["completedFiles"] / msgData.content["totalFiles"]) * 100 })
                return
            }
            case "zip_complete": {
                notifications.hide(`zip_create_${msgData.subscribeKey}`)
                downloadSingleItem(msgData.content["takeoutId"], authHeader, dispatch, undefined, "zip")
                return
            }
            case "error": {
                notifications.show({title: "Websocket error", message: msgData.error, color: 'red'})
                return
            }
            default: {
                console.error("Could not parse websocket message type: ", msgData)
                return
            }
        }
    }
}

export const useKeyDown = (blockFocus, dispatch, searchRef) => {
    useEffect(() => {
        const onKeyDown = (event) => {
            if (!blockFocus) {
                if (document.activeElement !== searchRef.current && event.metaKey && event.key === 'a') {
                    event.preventDefault()
                    dispatch({ type: 'select_all' })
                } else if (!event.metaKey && ((event.which >= 49 && event.which <= 90) || event.key === "Backspace")) {
                    searchRef.current.focus()
                } else if (document.activeElement === searchRef.current && event.key === 'Escape') {
                    searchRef.current.blur()
                } else if (event.key === 'Escape') {
                    event.preventDefault()
                    dispatch({ type: 'clear_selected' })
                } else if (event.key === 'Shift') {
                    dispatch({ type: 'holding_shift', shift: true })
                }
            } else {
                if (event.metaKey && event.key === 'a') {
                    event.stopPropagation()
                }
            }
        }

        const onKeyUp = (event) => {
            if (!blockFocus) {
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
    }, [blockFocus, dispatch, searchRef])
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

export function deleteSelected(selectedMap: Map<string, boolean>, dirMap: Map<string, fileData>, authHeader) {
    for (const itemKey of selectedMap.keys()) {
        const item = dirMap.get(itemKey)
        if (!item?.id) {
            console.error("Could not get id of file to delete", item)
            continue
        }
        DeleteFile(item.id, authHeader)
    }
// dispatch({ type: "delete_selected" })
}

export function moveSelected(selectedMap: Map<string, boolean>, destinationId: string, authHeader) {
    return MoveFiles(Array.from(selectedMap.keys()), destinationId, authHeader).catch(r => notifications.show({title: "Failed to move files", message: String(r), color: 'red'}))
}

export function GetDirItems(filebrowserState: FileBrowserStateType, userInfo, dispatch, authHeader, gridRef) {
    let itemsList = MapToList(filebrowserState.dirMap).filter((val) => { return val.filename.toLowerCase().includes(filebrowserState.searchContent.toLowerCase()) })
    let scanRequired = false

    const items = itemsList.map((entry: fileData) => {
        if (entry.id === userInfo.trashFolderId) {return null}
        if (entry.mediaData?.mediaType.IsDisplayable && !entry.imported && !entry.isDir) {
            scanRequired = true
        }
        let move: () => void
        if (!entry.isDir) {
            move = () => { }
        } else {
            move = () => { moveSelected(filebrowserState.selected, entry.id, authHeader).then(() => dispatch({type: 'clear_selected'})) }
        }

        return (
            <Item
                key={entry.id}
                itemData={entry}
                selected={filebrowserState.selected.get(entry.id)}
                moveSelected={move}
                dragging={filebrowserState.draggingState}
                dispatch={dispatch}
                authHeader={authHeader}
                root={gridRef}
            />
        )
    })
    return { items, scanRequired }
}