import { itemData, FileBrowserStateType } from '../../types/Types'
import HandleFileUpload from "../../api/Upload"
import { Dispatch, DragEvent, useEffect, useState } from 'react'
import { ChangeOwner, DeleteFile, GetDirectoryData } from '../../api/FileBrowserApi'
import API_ENDPOINT from '../../api/ApiEndpoint'
import Item from './FileItem'
import { BlankCard } from '../../types/Styles'

const handleSelect = (state: FileBrowserStateType, action) => {
    let numSelected = state.selected.size
    if (state.holdingShift && numSelected > 0 && state.lastSelected != "") {
        const dirList = MapToList(state.dirMap)
        let startIndex = dirList.findIndex((val) => val.filepath == state.lastSelected)
        let endIndex = dirList.findIndex((val) => val.filepath == action.itempath)

        if (endIndex < startIndex) {
            [startIndex, endIndex] = [endIndex, startIndex]
        }

        for (const val of dirList.slice(startIndex, endIndex + 1)) {
            state.selected.set(val.filepath, true)
        }

    } else {
        // If action.selected is undefined, i.e. not passed to the request,
        // we treat that as a request to toggle the selection
        if (action.selected == undefined) {
            if (state.selected.get(action.itempath)) {
                state.selected.delete(action.itempath)
            } else {
                state.selected.set(action.itempath, true)
                return { ...state, lastSelected: action.itempath }
            }
        }
        // state.selected.get returns undefined if not selected,
        // so we not (!) it to make boolean, and not the other to match... yay javascript :/
        else if (!state.selected.get(action.itempath) === !action.selected) {
            // If the item is already in the correct state, we do nothing.
            // Specifically, we do not overwrite lastSelected
        } else {
            if (action.selected) {
                state.selected.set(action.itempath, true)
            } else {
                state.selected.delete(action.itempath)
            }
        }
    }
    return { ...state }
}

export const fileBrowserReducer = (state: FileBrowserStateType, action) => {
    switch (action.type) {
        case 'update_item': {
            let items: itemData[]
            if (action.item) {
                items = [action.item]
            } else if (action.items) {
                items = action.items
            } else {
                return { ...state }
            }

            for (const item of items) {
                item.updatePath = ""
                state.dirMap.set(item.filepath, item)
            }

            return { ...state, loading: false }
        }

        case 'add_template_items': {
            let newMap = state.dirMap
            for (const tmpItem of action.files) {

                let item: itemData = {
                    filepath: state.path + tmpItem.name,
                    updatePath: "",
                    isDir: false,
                    imported: false,
                    modTime: new Date().toString(),
                    mediaData: null
                }
                newMap.set(item.filepath, item)
            }
            return {
                ...state,
            }
        }

        case 'set_path': {
            let newPath = action.path.replace(/\/\/+/g, '/')
            newPath = action.path.replace("/files", '')
            if (newPath != state.path) {
                state.dirMap.clear()
                state.selected.clear()
            }
            if (!newPath.endsWith('/')) {
                newPath = newPath + '/'
            }
            return {
                ...state,
                path: newPath,
            }
        }

        case 'set_loading': {
            return {
                ...state,
                loading: action.loading
            }
        }

        case 'set_scan_progress': {
            let loading = true
            if (action.progress == 0 || action.progress == 1) {
                loading = false
            }
            return {
                ...state,
                loading: loading,
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

        case 'start_editing': {
            return {
                ...state,
                editing: action.file
            }
        }

        case 'reject_edit': {
            return {
                ...state,
                editing: ""
            }
        }

        case 'confirm_edit': {
            if (state.dirMap.has(action.newPath)) {
                return {
                    ...state,
                    editing: ""
                }
            }

            state.dirMap.set(action.newPath, state.dirMap.get(action.file))
            state.dirMap.get(action.newPath).filepath = action.newPath

            if (state.selected.get(action.file)) {
                state.selected.delete(action.file)
                state.selected.set(action.newPath, true)
            }

            state.dirMap.delete(action.file)
            return {
                ...state,
                editing: ""
            }
        }

        case 'set_selected': {
            return handleSelect(state, action)
        }

        case 'select_all': {
            for (const key of state.dirMap.keys()) {
                if (key.toLowerCase().includes(state.searchContent.toLowerCase())) {
                    state.selected.set(key, true)
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

            // If its already selected, the only action that will ever be taken is to unselect
            if (state.selected.get(action.itempath)) {
                state.selected.delete(item.filepath)

                return { ...state }
            }

            // If there are other things selected, we have some decisions to make
            if (state.selected.size != 0) {
                return handleSelect(state, action)
            } else {
                if (item.isDir) {
                    return { ...state, path: (item.filepath).replace(/\/\/+/g, '/') }
                }
                document.documentElement.style.overflow = "hidden"
                return {
                    ...state,
                    presentingPath: item.filepath
                }
            }
        }

        case 'set_hovering': {
            if (state.draggingState === 0) {
                return { ...state }
            }

            if (state.hovering === action.itempath) {
                return { ...state }
            }

            return {
                ...state,
                hovering: action.itempath
            }
        }

        case 'clear_items': {
            state.dirMap.clear()
            state.selected.clear()

            return {
                ...state,
                lastSelected: ""
            }
        }

        case 'clear_selected': {
            state.selected.clear()
            state.selected.values()

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
            if (!action.ignoreMissingItem && (!targetItem || state.draggingState === 0 || !targetItem.isDir || state.selected.get(targetPath))) {
                return { ...state, draggingState: 0 }
            }

            for (const itemPath of state.selected.keys()) {
                const item = state.dirMap.get(itemPath)
                item.updatePath = (targetPath + itemPath.slice(itemPath.lastIndexOf('/'))).replace('//', '/')
            }
            state.selected.clear()
            return { ...state, draggingState: 0 }
        }

        case 'delete_selected': {
            for (const key of state.selected.keys()) {
                state.dirMap.delete(key)
                state.selected.delete(key)
            }
            return { ...state }
        }

        case 'delete_from_map': {
            state.dirMap.delete(action.item)
            state.selected.delete(action.item)

            return {
                ...state,
            }
        }

        case 'holding_shift': {
            return {
                ...state,
                holdingShift: action.shift,
            }
        }

        case 'set_presentation': {
            document.documentElement.style.overflow = "hidden"
            return {
                ...state,
                presentingPath: action.presentingPath
            }
        }

        // case 'presentation_next': {
        //     return { ...state }
        //     if (state.mediaIdMap[state.presentingPath].next) {
        //         changeTo = state.mediaIdMap[state.presentingPath].next
        //     } else {
        //         changeTo = state.presentingPath
        //     }
        //     return {
        //         ...state,
        //         presentingPath: changeTo
        //     }
        // }

        // case 'presentation_previous': {
        //     return { ...state }
        //     let changeTo
        //     if (state.mediaIdMap[state.presentingPath].previous) {
        //         changeTo = state.mediaIdMap[state.presentingPath].previous
        //     } else {
        //         changeTo = state.presentingPath
        //     }
        //     return {
        //         ...state,
        //         presentingPath: changeTo
        //     }
        // }

        case 'stop_presenting': {
            document.documentElement.style.overflow = "visible"
            return {
                ...state,
                presentingPath: ""
            }
        }

        default: {
            console.error("Got unexpected dispatch type: ", action.type)
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

        if (a.filepath > b.filepath) {
            return 1
        } else if (a.filepath < b.filepath) {
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

export const HandleDrop = (event, path, dirMap, dispatch, wsSend, enqueueSnackbar, authHeader) => {
    event.preventDefault()
    event.stopPropagation()
    dispatch({ type: "set_dragging", dragging: false })
    dispatch({ type: "set_loading", loading: true })

    let uploads: Promise<any>[] = []

    for (const file of event.dataTransfer.files) {
        if (dirMap.has(path + file.name)) {
            enqueueSnackbar(file.name + " already exists in this directory", { variant: "error" })
        } else {
            dispatch({ type: "add_template_items", files: [file] })
            let p = HandleFileUpload(file, path, wsSend, authHeader)
            uploads.push(p)
        }
    }
    Promise.all(uploads).then(() => dispatch({ type: "set_loading", loading: false }))
}

export function downloadSelected(selectedMap: Map<string, boolean>, path, dispatch, authHeader) {
    dispatch({ type: "set_loading", loading: true })

    var url = new URL(`${API_ENDPOINT}/takeout`)
    var filename: string

    fetch(url.toString(), { headers: authHeader, method: "POST", body: JSON.stringify({ items: selectedMap.keys(), path: path }) })
        .then((res) => {
            filename = res.headers.get("Content-Disposition").split(';')[1].split('=')[1].replaceAll("\"", "")
            return res.blob()
        })
        .then((res) => {
            const aElement = document.createElement("a");
            aElement.setAttribute("download", filename);
            const href = URL.createObjectURL(res);
            aElement.href = href;
            aElement.setAttribute("target", "_blank");
            aElement.click();
            URL.revokeObjectURL(href);
            dispatch({ type: "set_loading", loading: false })
        });
}

export function HandleWebsocketMessage(lastMessage, path, dispatch, navigate, enqueueSnackbar, authHeader) {
    if (lastMessage) {
        let msgData = JSON.parse(lastMessage.data)

        switch (msgData["type"]) {
            case "item_update": {
                dispatch({ type: "update_item", item: msgData["content"] })
                return
            }
            case "item_deleted": {
                dispatch({ type: "delete_from_map", item: msgData["content"].path })
                return
            }
            case "finished": {
                dispatch({ type: "set_scan_progress", progress: 0 })
                return
            }
            // case "refresh": {
            //     GetDirectoryData(path, dispatch, navigate, authHeader)
            //     return
            // }
            case "scan_directory_progress": {
                dispatch({ type: "set_scan_progress", progress: (1 - (msgData["remainingTasks"] / msgData["totalTasks"])) * 100 })
                return
            }
            case "error": {
                if (msgData["error"] == "upload_error") {
                    dispatch({ type: "delete_from_map", item: msgData["content"]["File"] })
                }
                enqueueSnackbar(msgData["content"]["Message"], { variant: "error" })
                return
            }
            default: {
                console.error("Got unexpected websocket message type: ", msgData["type"])
                return
            }
        }
    }
}

export const useKeyDown = (editing, dispatch, searchRef) => {

    const onKeyDown = (event) => {
        if (!editing) {
            if (document.activeElement !== searchRef.current?.children[0] && event.metaKey && event.key === 'a') {
                event.preventDefault();
                dispatch({ type: 'select_all' })

            } else if (!event.metaKey && ((event.which >= 65 && event.which <= 90) || event.key == "Backspace")) {
                searchRef.current.children[0].focus()

            } else if (document.activeElement === searchRef.current.children[0] && event.key === 'Escape') {
                searchRef.current.children[0].blur()

            } else if (event.key === 'Escape') {
                event.preventDefault()
                dispatch({ type: 'clear_selected' })

            } else if (event.key === 'Shift') {
                dispatch({ type: 'holding_shift', shift: true })

            }
        } else {
            if (event.metaKey && event.key === 'a') {
                event.stopPropagation();
            } else if (event.key === 'Escape') {
                event.preventDefault()
                dispatch({ type: 'reject_edit' })
                return
            }
        }
    };

    const onKeyUp = (event) => {
        if (!editing) {
            if (event.key === 'Shift') {
                dispatch({ type: 'holding_shift', shift: false })
            }
        }
    }
    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        document.addEventListener('keyup', onKeyUp)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
            document.removeEventListener('keyup', onKeyUp)
        };
    }, [onKeyDown, onKeyUp])
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
    return mousePosition;
}

export function deleteSelected(selectedMap: Map<string, boolean>, dispatch, authHeader) {
    for (const key of selectedMap.keys()) {
        DeleteFile(key, authHeader)
    }
    dispatch({ type: "delete_selected" })
}

export function changeOwner(dirMap, dispatch, authHeader) {
    let hashesToUpdate = []
    for (const key of dirMap.keys()) {
        if (dirMap.get(key).selected) {
            hashesToUpdate.push(dirMap.get(key).mediaData.FileHash)
        }
    }
    ChangeOwner(hashesToUpdate, "ethan", authHeader)
}

export function GetDirItems(filebrowserState: FileBrowserStateType, dispatch, authHeader) {
    let itemsList = MapToList(filebrowserState.dirMap)
    let scanRequired = false

    itemsList = itemsList.filter((val) => { return val.filepath.toLowerCase().includes(filebrowserState.searchContent.toLowerCase()) })

    // const filteredItems = itemsList.filter((val) => {
    //     // If item is not in the directory we are looking at... don't show it
    //     return (filebrowserState.path + val.filepath.slice(val.filepath.lastIndexOf('/'))).replace('//', '/') == val.filepath
    // })

    const items = itemsList.map((entry: itemData) => {
        if (entry.mediaData && !entry.imported && !entry.isDir) {
            scanRequired = true
        }
        return (
            <Item
                key={entry.filepath}
                itemData={entry}
                selected={filebrowserState.selected.get(entry.filepath)}
                editing={filebrowserState.editing == entry.filepath}
                dragging={filebrowserState.draggingState}
                dispatch={dispatch}
                authHeader={authHeader}
            />
        )
    })
    return { items, scanRequired }
}

