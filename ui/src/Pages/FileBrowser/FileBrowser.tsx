// React
import { useState, useEffect, useReducer, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

// MUI
import Box from '@mui/material/Box'
import DeleteIcon from '@mui/icons-material/Delete'
import DownloadIcon from '@mui/icons-material/Download';
import { IconButton, LinearProgress, Tooltip } from '@mui/material'
import Menu from '@mui/material/Menu'
import MenuItem from '@mui/material/MenuItem'
import ListItemIcon from '@mui/material/ListItemIcon'
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder'
import Badge from '@mui/material/Badge';
import CheckBoxIcon from '@mui/icons-material/CheckBox';

// Local
import { GetDirectoryData, CreateDirectory, RenameFile, DeleteFile } from '../../api/FileBrowserApi'
import Item from './FileItem'
import Presentation from '../../components/Presentation'
import HeaderBar from "../../components/HeaderBar"
import HandleFileUpload from "../../api/Upload"
import { itemData, FileBrowserTypes, FileBrowserStateType } from '../../types/FileBrowserTypes'
import Crumbs from '../../components/Crumbs'
import API_ENDPOINT from '../../api/ApiEndpoint'
import { dispatchSync } from '../../api/Websocket'


const mapToList = (dirMap: Map<string, itemData>) => {
    const newList = Array.from(dirMap.values())

    newList.sort((a, b) => {
        if (a.mediaData && !b.mediaData) {
            return -1
        } else if (!a.mediaData && b.mediaData) {
            return 1
        }

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

const fileBrowserReducer = (state: FileBrowserStateType, action) => {
    switch (action.type) {
        case 'update_items': {
            let items: itemData[] = action.items
            for (const item of items) {
                item.selected = false
                state.dirMap.set(item.filepath, item)
            }

            return {
                ...state
            }
        }

        case 'add_template_items': {
            let newMap = state.dirMap
            for (const tmpItem of action.files) {

                let item: itemData = {
                    filepath: state.path + tmpItem.name,
                    isDir: false,
                    imported: false,
                    modTime: new Date().toString(),
                    selected: false,
                    mediaData: null
                }
                newMap.set(item.filepath, item)
            }
            return {
                ...state,
            }
        }

        case 'set_path': {
            let newMap = state.dirMap
            const newPath = action.path.replace(/\/\/+/g, '/')
            if (newPath != state.path) {
                newMap.clear()
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
            return {
                ...state,
                scanProgress: action.progress
            }
        }

        case 'set_dragging': {
            return {
                ...state,
                dragging: action.dragging
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
            // Call api to change filename
            let newMap = state.dirMap

            let oldName = action.file.replace(/.*?([^/]*)$/, '$1')
            let newName = action.newName
            const parentDir = action.file.replace(/(.*?)[^/]*$/, '$1')

            const [_, oldExt] = oldName.split('.')
            if (oldExt) {
                const [_, newExt] = newName.split('.')
                if (!newExt || newExt != oldExt) {
                    newName += "." + oldExt
                }
            }
            let newPath = (parentDir + newName)

            if (newMap.has(newPath)) {
                return {
                    ...state,
                    editing: ""
                }
            }

            RenameFile(action.file, newPath)

            newMap.set(newPath, newMap.get(action.file))
            newMap.get(newPath).filepath = newPath

            newMap.delete(action.file)
            return {
                ...state,
                editing: ""
            }
        }

        case 'set_selected': {
            let newMap = state.dirMap
            let numSelected = state.numSelected
            if (state.holdingShift && state.numSelected > 0 && state.lastSelected != "") {
                const dirList = mapToList(newMap)
                let startIndex = dirList.findIndex((val) => val.filepath == state.lastSelected)
                let endIndex = dirList.findIndex((val) => val.filepath == action.itempath)
                if (endIndex < startIndex) {
                    [startIndex, endIndex] = [endIndex, startIndex]
                }
                let changedCounter = 0
                for (const val of dirList.slice(startIndex, endIndex + 1)) {
                    if (newMap.get(val.filepath).selected == true) {
                        continue
                    }
                    newMap.get(val.filepath).selected = true
                    changedCounter += 1
                }
                numSelected += changedCounter
            } else {
                newMap.get(action.itempath).selected = action.selected
                if (action.selected) {
                    numSelected += 1
                } else {
                    numSelected -= 1
                }
            }
            return {
                ...state,
                numSelected: numSelected,
                lastSelected: action.itempath
            }
        }

        case 'clear_selected': {
            for (const value of state.dirMap.values()) {
                value.selected = false
            }

            return {
                ...state,
                numSelected: 0,
                lastSelected: ""
            }
        }

        case 'delete_selected': {
            let newMap = state.dirMap
            for (const key of newMap.keys()) {
                if (newMap.get(key).selected) {
                    DeleteFile(key)
                    newMap.delete(key)
                }
            }
            return {
                ...state,
                numSelected: 0
            }
        }

        case 'delete_from_map': {
            let newMap = state.dirMap

            newMap.delete(action.file)

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

const HandleDrag = (event, dispatch, dragging) => {
    event.preventDefault()
    if (event.type === "dragenter" || event.type === "dragover") {
        !dragging && dispatch({ type: "set_dragging", dragging: true })

    } else {
        dispatch({ type: "set_dragging", dragging: false })
    }
}

const HandleDrop = (event, path, dirMap, dispatch, wsSend, enqueueSnackbar) => {
    event.preventDefault()
    event.stopPropagation()
    dispatch({ type: "set_dragging", dragging: false })

    let filteredFiles = []
    for (const file of event.dataTransfer.files) {
        if (dirMap.has(path + file.name)) {
            enqueueSnackbar(file.name + " already exists in this directory", { variant: "error" })
        } else {
            filteredFiles.push(file)
        }
    }

    dispatch({ type: "add_template_items", files: filteredFiles })
    if (filteredFiles.length != 0) {
        dispatch({ type: "set_loading", loading: true })
        HandleFileUpload(filteredFiles, path, wsSend)
    }
}

function downloadSelected(dirMap: Map<string, itemData>, path, dispatch) {
    let itemsToDownload = []
    for (const item of dirMap.values()) {
        if (item.selected) {
            itemsToDownload.push(item.filepath)
        }
    }

    dispatch({ type: "set_loading", loading: true })

    var url = new URL(`${API_ENDPOINT}/takeout`)
    var filename: string

    fetch(url.toString(), { method: "POST", body: JSON.stringify({ items: itemsToDownload, path: path }) })
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

function HandleWebsocketMessage(lastMessage, path, dispatch, navigate, enqueueSnackbar) {
    if (lastMessage) {
        let msgData = JSON.parse(lastMessage.data)

        switch (msgData["type"]) {
            case "new_items": {
                dispatch({ type: "update_items", items: msgData["content"] })
                return
            }
            case "finished": {
                dispatch({ type: "set_scan_progress", progress: 0 })
                dispatch({ type: "set_loading", loading: false })
                return
            }
            case "refresh": {
                GetDirectoryData(path, dispatch, navigate)
                return
            }
            case "scan_directory_progress": {
                dispatch({ type: "set_scan_progress", progress: (1 - (msgData["remainingTasks"] / msgData["totalTasks"])) * 100 })
                return
            }
            case "error": {
                if (msgData["error"] == "upload_error") {
                    dispatch({ type: "delete_from_map", file: msgData["content"]["File"] })
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

function StartKeybaordListener(dispatch) {

    const keyDownHandler = event => {
        if (event.key === 'Escape') {
            event.preventDefault()
            dispatch({
                type: 'clear_selected'
            })
        } else if (event.key === 'Shift') {
            dispatch({ type: 'holding_shift', shift: true })
        }
        else {
            //console.log("Uncaught keypress: ", event.key)
        }
    }

    const keyUpHandler = event => {
        if (event.key === 'Shift') {
            dispatch({ type: 'holding_shift', shift: false })
        }
    }

    window.addEventListener('keydown', keyDownHandler)
    window.addEventListener('keyup', keyUpHandler)

    //return () => {
    //    document.removeEventListener('keydown', keyDownHandler)
    //    document.removeEventListener('keyup', keyUpHandler)
    //}
}

const FileBrowser = ({ wsSend, lastMessage, readyState, enqueueSnackbar }: FileBrowserTypes) => {

    const path = ("/" + useParams()["*"] + "/").replace(/\/\/+/g, '/')

    const [filebrowserState, dispatch]: [FileBrowserStateType, React.Dispatch<any>] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, itemData>(),
        path: path,
        dragging: false,
        loading: false,
        presentingPath: "",
        numSelected: 0,
        scanProgress: 0,
        holdingShift: false,
        lastSelected: "",
        editing: ""
    })

    const navigate = useNavigate()
    const [alreadyScanned, setAlreadyScanned] = useState(false)

    const [contextMenu, setContextMenu] = useState<{
        mouseX: number
        mouseY: number
    } | null>(null)

    const handleContextMenu = (event: React.MouseEvent) => {
        event.preventDefault()
        setContextMenu(
            contextMenu === null
                ? {
                    mouseX: event.clientX + 2,
                    mouseY: event.clientY - 6,
                }
                : null,
        )
    }

    useEffect(() => {
        if (readyState == "open") {
            GetDirectoryData(path, dispatch, navigate)
        }
    }, [readyState])

    useEffect(() => {
        StartKeybaordListener(dispatch)
    }, [])

    useEffect(() => {
        HandleWebsocketMessage(lastMessage, path, dispatch, navigate, enqueueSnackbar)
    }, [lastMessage])

    useEffect(() => {
        setAlreadyScanned(false)
        dispatch({ type: 'set_path', path: path })
        dispatch({ type: "clear_selected" })
        dispatch({ type: "set_scan_progress", progress: 0 })
        GetDirectoryData(path, dispatch, navigate)
    }, [path])

    const dirItems = useMemo(() => {
        const itemsList = mapToList(filebrowserState.dirMap)
        const anyChecked = filebrowserState.numSelected > 0 ? true : false
        let scanRequired = false
        const items = itemsList.map((entry: itemData) => {
            if (entry.mediaData && !entry.imported && !entry.isDir) {
                scanRequired = true
            }
            return (
                // <Box key={entry.filepath} >
                <Item key={entry.filepath} itemData={entry} editing={filebrowserState.editing == entry.filepath} dispatch={dispatch} anyChecked={anyChecked} navigate={navigate} />
                // </Box>
            )
        })
        if (scanRequired && !alreadyScanned) { setAlreadyScanned(true); dispatch({ type: "set_loading", loading: true }); dispatchSync(path, wsSend, false) }
        return items
    }, [filebrowserState.dirMap.values(), filebrowserState.editing, filebrowserState.numSelected])

    return (
        <Box display={"flex"} flexDirection={"column"}>
            <HeaderBar dispatch={dispatch} wsSend={wsSend} page={"files"} />
            <Box
                onDragOver={(event => HandleDrag(event, dispatch, filebrowserState.dragging))}
                onDragLeave={event => HandleDrag(event, dispatch, filebrowserState.dragging)}
                onDrop={(event => HandleDrop(event, path, filebrowserState.dirMap, dispatch, wsSend, enqueueSnackbar))}
                bgcolor={filebrowserState.dragging ? "rgb(200, 200, 200)" : "white"}
                display={"flex"}
                flexDirection={"column"}
                alignItems={"center"}
                paddingLeft={0}
                paddingRight={0}
                minHeight={"max-content"}
                height={"93vh"}
                sx={{ outline: filebrowserState.dragging ? "rgb(54, 147, 255) solid 2px" : "", outlineOffset: "-10px" }}
                onContextMenu={handleContextMenu}
            >

                {filebrowserState.presentingPath != "" && (
                    <Presentation mediaData={filebrowserState.dirMap.get(filebrowserState.presentingPath)?.mediaData} dispatch={dispatch} />
                )}


                {filebrowserState.loading && (
                    <Box width={"100%"} position={"absolute"} zIndex={10}>
                        <LinearProgress style={{ position: "absolute", width: "100%" }} />
                        {/* {!scanProgress && (
                        )} */}
                        {filebrowserState.scanProgress != 0 && (
                            <Box sx={{ width: '100%' }}>
                                <LinearProgress variant="determinate" value={filebrowserState.scanProgress} />
                                <p style={{ position: "absolute", left: "6vw" }}>Syncing filesystem with database...</p>
                            </Box>
                        )}
                    </Box>
                )}
                {filebrowserState.numSelected > 0 && (
                    <Box position={"absolute"} display={"flex"} left={"5vw"} >

                        <IconButton style={{ padding: "15px" }} onClick={() => { dispatch({ type: "clear_selected" }) }}>
                            <Tooltip title={"Unselect All (esc)"}>
                                <Badge badgeContent={filebrowserState.numSelected} color="secondary">
                                    <CheckBoxIcon color="action" />
                                </Badge>
                            </Tooltip>
                        </IconButton>

                        <IconButton style={{ padding: "15px" }} onClick={() => { downloadSelected(filebrowserState.dirMap, path, dispatch) }}>
                            <Tooltip title={"Download Selected"}>
                                <DownloadIcon />
                            </Tooltip>
                        </IconButton>

                        <IconButton style={{ padding: "15px" }} onClick={() => { dispatch({ type: "delete_selected" }) }}>
                            <Tooltip title={"Delete Selected"}>
                                <DeleteIcon />
                            </Tooltip>
                        </IconButton>

                    </Box>
                )}
                <Box marginTop={2} marginBottom={2} width={"max-content"}>
                    <Crumbs path={filebrowserState.path} includeHome={true} navOnLast={true} navigate={navigate} />
                </Box>

                <Box display={"flex"} justifyContent={"center"} width={"100%"} >
                    <Box display={"flex"} flexDirection={"row"} flexWrap={"wrap"} width={"90%"} height={"100%"} justifyContent={"flex-start"} >
                        {dirItems}
                    </Box>
                </Box>
                <Menu
                    open={contextMenu !== null}
                    onClose={() => setContextMenu(null)}
                    anchorReference="anchorPosition"
                    anchorPosition={
                        contextMenu !== null
                            ? { top: contextMenu.mouseY, left: contextMenu.mouseX }
                            : undefined
                    }
                >
                    <MenuItem onClick={() => { setContextMenu(null); CreateDirectory(path).then(newDir => { GetDirectoryData(path, dispatch, navigate); dispatch({ type: 'start_editing', file: newDir }) }) }}>
                        <ListItemIcon>
                            <CreateNewFolderIcon />
                        </ListItemIcon>
                        New Folder
                    </MenuItem>
                </Menu>
            </Box>
        </Box >
    )
}

export default FileBrowser