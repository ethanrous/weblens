import { itemData, FileBrowserStateType } from '../../types/Types'
import Upload, { fileUploadMetadata } from "../../api/Upload"
import { Dispatch, DragEvent, useEffect, useState } from 'react'
import { ChangeOwner, CreateFolder, DeleteFile, MoveFiles, downloadSingleItem, downloadTakeout, requestZipCreate } from '../../api/FileBrowserApi'
import Item from './FileItem'

import { notifications } from '@mantine/notifications'

const handleSelect = (state: FileBrowserStateType, action) => {
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

function updateItem(dirMap: Map<string, itemData>, user, existingId: string, newData: itemData, currentFolderId) {

    let exitingItem: itemData = dirMap.get(existingId)
    let possibleTemplateItem: itemData = dirMap.get(newData.filename)
    if (exitingItem && exitingItem.id !== newData.id) {
        dirMap.delete(exitingItem.id)
    }
    if (possibleTemplateItem) {
        dirMap.delete(possibleTemplateItem.id)
    }

    if (newData.parentFolderId === currentFolderId || user !== newData.owner) {
        if (!newData.imported) {
            newData.id = newData.filename
        }

        dirMap.set(newData.id, newData)
    }
}

export const fileBrowserReducer = (state: FileBrowserStateType, action) => {
    switch (action.type) {
        case 'update_item': {
            updateItem(state.dirMap, action.user, action.itemId, action.updateInfo, state.folderInfo.id)
            return { ...state }
        }

        case 'update_many': {
            const itemsMeta: { itemId: string, updateInfo: itemData }[] = action.items
            for (const itemMeta of itemsMeta) {
                updateItem(state.dirMap, action.user, itemMeta.itemId, itemMeta.updateInfo, state.folderInfo.id)
            }
            return { ...state }
        }

        case 'add_skeleton': {
            let item: itemData = {
                id: action.filename,
                filename: action.filename,
                parentFolderId: state.folderInfo.id,
                owner: "",
                isDir: false,
                imported: false,
                modTime: new Date().toString(),
                size: 0,
                visible: true,
                mediaData: null
            }

            state.dirMap.set(action.filename, item)
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
            let item: itemData = {
                id: "TEMPLATE_NEW_FOLDER",
                filename: "",
                parentFolderId: state.folderInfo.id,
                owner: "",
                isDir: true,
                imported: false,
                modTime: new Date().toString(),
                size: 0,
                visible: true,
                mediaData: null
            }
            state.dirMap.set(item.id, item)
            return { ...state, editing: item.id }
        }

        case 'start_editing': {
            return {
                ...state,
                editing: action.fileId
            }
        }

        case 'reject_edit': {
            const item = state.dirMap.get(state.editing)
            if (!item) {
                return state
            }
            if (state.editing === "TEMPLATE_NEW_FOLDER") {
                state.dirMap.delete("TEMPLATE_NEW_FOLDER")
            }
            return {
                ...state,
                editing: null
            }
        }

        case 'confirm_edit': {
            if (state.editing === "TEMPLATE_NEW_FOLDER") {
                state.dirMap.delete("TEMPLATE_NEW_FOLDER")
                state.selected.delete("TEMPLATE_NEW_FOLDER")
            }
            if (action.itemId !== action.newItemId) {
                state.selected.delete(action.itemId)
            }
            return {
                ...state,
                editing: null
            }
        }

        case 'set_selected': {
            return handleSelect(state, action)
        }

        case 'set_visible': {
            const item = state.dirMap.get(action.item)
            if (item) {
                item.visible = action.visible
                state.dirMap.set(action.item, item)
            }

            return { ...state }
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

        case 'handle_click': {
            const item = state.dirMap.get(action.itempath)
            if (!item) {
                console.error("Failed to handle click on file item")
                return { ...state }
            }

            // If an item is already selected, we only ever unselect
            if (state.selected.get(action.itempath)) {
                state.selected.delete(item.id)

                return { ...state }
            }

            // If there are other things selected, there are many options, see `handleSelect`
            if (state.selected.size !== 0) {
                return handleSelect(state, action)
            }
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

        case 'move_selected': {
            // Move selected items into directory at action.targetItemPath
            let targetPath: string = action.targetItemPath
            targetPath = targetPath.replace('files/', '')
            const targetItem = state.dirMap.get(targetPath)

            // If (1) the item does not exist, (2) we are not dragging (3) the target is not a directory, or (4) the target is selected, we bail
            if (state.draggingState === 0 || !action.ignoreMissingItem && (!targetItem || !targetItem.isDir || state.selected.get(targetPath))) {
                return { ...state, draggingState: 0 }
            }

            state.selected.clear()
            return { ...state, draggingState: 0 }
        }

        case 'share_selected': {
            return { ...state, sharing: true }
        }

        case 'add_selected_to_album': {
            return { ...state, albuming: true }
        }

        case 'close_share': {
            return { ...state, sharing: false }
        }

        case 'close_add_to': {
            return { ...state, albuming: false }
        }

        case 'delete_from_map': {
            state.dirMap.delete(action.itemId)
            state.selected.delete(action.itemId)

            return { ...state }
        }

        case 'add_to_upload_map': {
            state.uploadMap.set(action.uploadName, true)
            return { ...state }
        }

        case 'remove_from_upload_map': {
            state.uploadMap.delete(action.uploadName)
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

        case 'stop_presenting': {
            return {
                ...state,
                presentingId: ""
            }
        }

        default: {
            console.error("Got unexpected dispatch type: ", action.type)
            notifications.show({ title: "Unexpected filebrowser dispatch", message: action.type, color: 'red' })
            return { ...state }
        }
    }
}

export const MapToList = (dirMap: Map<string, itemData>) => {
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
    if (event.type === "dragenter" || event.type === "dragover") {
        !dragging && dispatch({ type: "set_dragging", dragging: true, external: true })

    } else {
        dispatch({ type: "set_dragging", dragging: false })
    }
}

async function getFile(file): Promise<File> {
    try {
        return await file.getAsFile()
        // return new Promise((resolve, reject) => file.file(resolve, reject));
    } catch (err) {
        return file
        // return new Promise((resolve, reject) => file)
    }
}

async function addDir(fsEntry, parentFolderId: string, topFolderKey: string, rootFolderId: string, authHeader): Promise<any> {
    return await new Promise(async (resolve: (value: fileUploadMetadata[]) => void, reject): Promise<fileUploadMetadata[]> => {
        if (fsEntry.isDirectory) {
            const { folderId, alreadyExisted } = await CreateFolder(parentFolderId, fsEntry.name, authHeader)
            if (!folderId || alreadyExisted) {
                reject("Could not create folder")
            }
            let e: fileUploadMetadata = null
            if (!topFolderKey) {
                topFolderKey = folderId
                e = { file: fsEntry, isDir: true, folderId: folderId, parentId: rootFolderId, isTopLevel: true, topLevelParentKey: null }
            }
            var dirReader = fsEntry.createReader()
            dirReader.readEntries(async (entries: FileSystemEntry[]) => {
                let resolvedEntries = await Promise.all(entries.map((entry) => addDir(entry, folderId, topFolderKey, rootFolderId, authHeader)))
                let re = []
                if (e !== null) {
                    re.push(e)
                }
                for (let r of resolvedEntries) {
                    if (r.length != undefined) {
                        re.push(...r)
                    } else {
                        re.push(r)
                    }
                }
                resolve(re)
            })
        } else {
            if (fsEntry.name === ".DS_Store") {
                resolve([])
                return
            }
            const f = await getFile(fsEntry)
            let e: fileUploadMetadata = { file: f, parentId: parentFolderId, isDir: f.type === "", isTopLevel: parentFolderId === rootFolderId, topLevelParentKey: topFolderKey }
            resolve([e])
        }
    })
}

export async function HandleDrop(items, rootFolderId, dirMap, authHeader, uploadDispatch, dispatch, wsSend) {
    let files: fileUploadMetadata[] = []
    let topLevels = []

    if (items) { // Handle Directory
        const names = new Map<string, boolean>()
        Array.from(dirMap.values()).map((value: itemData) => names.set(value.filename, true))

        for (const item of items) {
            let entry = item
            if (!entry) {
                continue
            }
            if (names.get(entry.name)) {
                notifications.show({ message: `A file or folder named "${entry.name}" already exists in this folder`, autoClose: 5000, color: "red" })
                continue
            }
            topLevels.push(
                addDir(entry, rootFolderId, null, rootFolderId, authHeader).then(newItems => files.push(...newItems)).catch((r) => notifications.show({ message: r, color: "red" }))
            )
        }
    }

    await Promise.all(topLevels)
    console.log(files)
    if (files.length !== 0) {
        Upload(files, rootFolderId, authHeader, uploadDispatch, dispatch, wsSend)
    }
}

export function downloadSelected(selectedMap: Map<string, boolean>, dirMap: Map<string, itemData>, folderId: string, dispatch, wsSend, authHeader) {
    const items = Array.from(selectedMap.keys())

    let taskId: string = ""

    if (items.length == 1 && !dirMap.get(items[0]).isDir) {
        const item = dirMap.get(items[0])
        if (!item.isDir) {
            downloadSingleItem(item, authHeader, dispatch)
            return
        }
    }

    const body = { items: items.map((val) => { let item: itemData = dirMap.get(val); return { parentFolderId: item.parentFolderId, filename: item.filename } }) }
    requestZipCreate(body, authHeader).then(({ json, status }) => {
        if (status === 200) {
            downloadTakeout(json.takeoutId, authHeader, dispatch)
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
        switch (msgData["messageStatus"]) {
            case "item_update": {
                dispatch({ type: "update_item", itemId: msgData["content"].itemId, updateInfo: msgData["content"].updateInfo, user: username })
                return
            }
            case "item_deleted": {
                dispatch({ type: "delete_from_map", itemId: msgData["content"].itemId })
                return
            }
            case "finished": {
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
                downloadTakeout(msgData.content["takeoutId"], authHeader, dispatch)
                return
            }
            case "error": {
                if (msgData["error"] === "upload_error") {
                    dispatch({ type: "delete_from_map", item: msgData["content"]["File"] })
                }
                return
            }
            default: {
                console.error("Could not parse websocket message type: ", msgData)
                return
            }
        }
    }
}

export const useKeyDown = (editing, blockFocus, dispatch, searchRef) => {
    useEffect(() => {
        const onKeyDown = (event) => {
            if (!editing && !blockFocus) {
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
            } else if (editing) {
                if (event.metaKey && event.key === 'a') {
                    event.stopPropagation()
                } else if (event.key === 'Escape') {
                    event.preventDefault()
                    dispatch({ type: 'reject_edit' })
                    return
                }
            }
        }

        const onKeyUp = (event) => {
            if (!editing) {
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
    }, [editing, blockFocus, dispatch, searchRef, document.activeElement])
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

export function deleteSelected(selectedMap: Map<string, boolean>, dirMap: Map<string, itemData>, folderId: string, dispatch, authHeader) {
    for (const itemKey of selectedMap.keys()) {
        const item = dirMap.get(itemKey)
        if (!item?.filename) {
            console.error("Could not get filename to delete", item)
            continue
        }
        DeleteFile(folderId, item.filename, authHeader)
    }
// dispatch({ type: "delete_selected" })
}

export function moveSelected(selectedMap: Map<string, boolean>, dirMap: Map<string, itemData>, destinationId: string, authHeader) {
    let files = []
    for (const itemKey of selectedMap.keys()) {
        const item = dirMap.get(itemKey)
        if (item) {
            files.push({ parentFolderId: item.parentFolderId, filename: item.filename })
        }
    }
    MoveFiles(files, destinationId, authHeader)
}

export function GetDirItems(filebrowserState: FileBrowserStateType, dispatch, authHeader, gridRef) {
    let itemsList = MapToList(filebrowserState.dirMap).filter((val) => { return val.filename.toLowerCase().includes(filebrowserState.searchContent.toLowerCase()) })
    let scanRequired = false

    const items = itemsList.map((entry: itemData) => {
        if (entry.mediaData && !entry.imported && !entry.isDir) {
            scanRequired = true
        }
        let move: () => void
        if (!entry.isDir) {
            move = () => { }
        } else {
            move = () => { moveSelected(filebrowserState.selected, filebrowserState.dirMap, entry.id, authHeader) }
        }

        return (
            <Item
                key={entry.id}
                itemData={entry}
                selected={filebrowserState.selected.get(entry.id)}
                moveSelected={move}
                editing={filebrowserState.editing === entry.id}
                dragging={filebrowserState.draggingState}
                dispatch={dispatch}
                authHeader={authHeader}
                root={gridRef}
            />
        )
    })
    return { items, scanRequired }
}