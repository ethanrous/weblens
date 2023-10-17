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
import GetWebsocket, { dispatchSync } from '../../api/Websocket'
import { itemData } from '../../types/FileBrowserTypes'
import Crumbs from '../../components/Crumbs'

// Other
import { useSnackbar } from 'notistack'
import API_ENDPOINT from '../../api/ApiEndpoint';

const mapToList = (dirMap) => {
    const newList = Object.keys(dirMap).map((key) => {
        return dirMap[key]
    })

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

function copyObject(map) {
    var newMap = {};
    for (var i in map) {
        newMap[i] = map[i]
    }
    return { ...map }

}

const fileBrowserReducer = (state, action) => {
    switch (action.type) {
        case 'update_items': {
            let newMap = copyObject(state.dirMap)
            let items: itemData[] = action.items
            for (const item of items) {
                item.selected = false
                newMap[item.filepath] = item
            }

            return {
                ...state,
                dirMap: newMap,
            }
        }

        case 'add_template_items': {
            let newMap = copyObject(state.dirMap)
            for (const tmpItem of action.files) {

                let item: itemData = {
                    filepath: state.path + tmpItem.name,
                    isDir: false,
                    imported: false,
                    modTime: new Date().toString(),
                    selected: false,
                    mediaData: null
                }
                newMap[item.filepath] = item
            }
            return {
                ...state,
                dirMap: newMap,
            }
        }

        case 'set_path': {
            let newMap = copyObject(state.dirMap)
            const newPath = action.path.replace(/\/\/+/g, '/')
            if (newPath != state.path) {
                newMap = {}
            }
            return {
                ...state,
                path: newPath,
                dirMap: newMap,
            }
        }

        case 'set_loading': {
            return {
                ...state,
                loading: action.loading
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
            let newMap = copyObject(state.dirMap)

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

            if (newMap[newPath]) {
                return {
                    ...state,
                    editing: ""
                }
            }

            RenameFile(action.file, newPath)

            newMap[newPath] = newMap[action.file]
            newMap[newPath].filepath = newPath

            delete newMap[action.file]
            return {
                ...state,
                dirMap: newMap,
                editing: ""
            }
        }

        case 'set_selected': {
            let newMap = copyObject(state.dirMap)
            let numSelected = state.numSelected

            if (state.holdingShift && state.numSelected > 0 && state.lastSelected != "") {
                const dirList = mapToList(state.dirMap)
                let startIndex = dirList.findIndex((val) => val.filepath == state.lastSelected)
                let endIndex = dirList.findIndex((val) => val.filepath == action.itempath)
                if (endIndex < startIndex) {
                    [startIndex, endIndex] = [endIndex, startIndex]
                }
                let changedCounter = 0
                for (const val of dirList.slice(startIndex, endIndex + 1)) {
                    if (newMap[val.filepath].selected == true) {
                        continue
                    }
                    newMap[val.filepath].selected = true
                    changedCounter += 1
                }
                numSelected += changedCounter
            } else {
                newMap[action.itempath].selected = action.selected
                if (action.selected) {
                    numSelected += 1
                } else {
                    numSelected -= 1
                }
            }
            return {
                ...state,
                dirMap: newMap,
                numSelected: numSelected,
                lastSelected: action.itempath
            }
        }

        case 'clear_selected': {
            let newMap = copyObject(state.dirMap)
            for (const key of Object.keys(newMap)) {
                newMap[key].selected = false
            }

            return {
                ...state,
                dirMap: newMap,
                numSelected: 0,
                lastSelected: ""
            }
        }

        case 'delete_selected': {
            let newMap = copyObject(state.dirMap)
            for (const key of Object.keys(newMap)) {
                if (newMap[key].selected) {
                    DeleteFile(key)
                    delete newMap[key]
                }
            }
            return {
                ...state,
                dirMap: newMap,
                numSelected: 0
            }
        }

        case 'delete_from_map': {
            let newMap = copyObject(state.dirMap)

            delete newMap[action.file]

            return {
                ...state,
                dirMap: newMap
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

        case 'presentation_next': {
            return { ...state }
            let changeTo
            if (state.mediaIdMap[state.presentingPath].next) {
                changeTo = state.mediaIdMap[state.presentingPath].next
            } else {
                changeTo = state.presentingPath
            }
            return {
                ...state,
                presentingPath: changeTo
            }
        }

        case 'presentation_previous': {
            return { ...state }
            let changeTo
            if (state.mediaIdMap[state.presentingPath].previous) {
                changeTo = state.mediaIdMap[state.presentingPath].previous
            } else {
                changeTo = state.presentingPath
            }
            return {
                ...state,
                presentingPath: changeTo
            }
        }

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
        if (dirMap[path + file.name]) {
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
    for (const item of Object.values(dirMap)) {
        if (item.selected) {
            itemsToDownload.push(item.filepath)
        }
    }

    dispatch({ type: "set_loading", loading: true })

    var url = new URL(`${API_ENDPOINT}/takeout`)
    var filename: string

    fetch(url.toString(), { method: "POST", body: JSON.stringify({ items: itemsToDownload, path: path }) })
        .then((res) => {
            console.log(res.headers.get("Content-Disposition"))
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

const FileBrowser = () => {

    const path = ("/" + useParams()["*"] + "/").replace(/\/\/+/g, '/')

    const [filebrowserState, dispatch] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, itemData>(),
        path: path,
        dragging: false,
        loading: false,
        presentingPath: "",
        numSelected: 0,
        holdingShift: false,
        lastSelected: "",
        editing: ""
    })

    const { enqueueSnackbar } = useSnackbar()
    const { wsSend, lastMessage, readyState } = GetWebsocket(enqueueSnackbar)
    const [scanProgress, setScanProgress] = useState(0)
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
                : // repeated contextmenu when it is already open closes it with Chrome 84 on Ubuntu
                // Other native context menus might behave different.
                // With this behavior we prevent contextmenu from the backdrop to re-locale existing context menus.
                null,
        )
    }

    const handleClose = () => {
        setContextMenu(null)
    }

    useEffect(() => {
        if (readyState == 1) {
            GetDirectoryData(path, dispatch)
        }
    }, [readyState])

    useEffect(() => {
        StartKeybaordListener(dispatch)
    }, [])

    useEffect(() => {
        if (lastMessage) {
            let msgData = JSON.parse(lastMessage.data)
            //console.log("Got server msg: ", msgData)

            switch (msgData["type"]) {
                case "new_items": {
                    dispatch({ type: "update_items", items: msgData["content"] })
                    return
                }
                case "finished": {
                    setScanProgress(0)
                    dispatch({ type: "set_loading", loading: false })
                    return
                }
                case "refresh": {
                    GetDirectoryData(path, dispatch)
                    return
                }
                case "scan_directory_progress": {
                    setScanProgress((1 - (msgData["remainingTasks"] / msgData["totalTasks"])) * 100)
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
    }, [lastMessage])

    useEffect(() => {
        setAlreadyScanned(false)
        dispatch({ type: 'set_path', path: path })
        dispatch({ type: "set_loading", loading: false })
        dispatch({ type: "clear_selected" })
        setScanProgress(0)
        GetDirectoryData(path, dispatch)
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
                <Box key={entry.filepath} >
                    <Item itemData={entry} editing={filebrowserState.editing == entry.filepath} dispatch={dispatch} anyChecked={anyChecked} navigate={navigate} />
                </Box>
            )
        })
        if (scanRequired && !alreadyScanned) { setAlreadyScanned(true); dispatch({ type: "set_loading", loading: true }); dispatchSync(path, wsSend, false) }
        return items
    }, [filebrowserState.dirMap, filebrowserState.editing])

    return (
        <Box
            sx={{
                display: "flex",
                flexDirection: 'column',
            }}
        >
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
                    <Presentation mediaData={filebrowserState.dirMap[filebrowserState.presentingPath].mediaData} dispatch={dispatch} />
                )}
                {filebrowserState.loading && (
                    <Box sx={{ width: '100%' }}>
                        {scanProgress == 0 && (
                            <LinearProgress />
                        )}
                        {scanProgress != 0 && (
                            <Box sx={{ width: '100%' }}>
                                <LinearProgress variant="determinate" value={scanProgress} />
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
                    <Crumbs path={filebrowserState.path} includeHome={true} navigate={navigate} />
                </Box>

                <Box display={"flex"} justifyContent={"center"} width={"100%"} >
                    <Box display={"flex"} flexDirection={"row"} flexWrap={"wrap"} width={"90%"} height={"100%"} justifyContent={"flex-start"} >
                        {dirItems}
                    </Box>
                </Box>
                <Menu
                    open={contextMenu !== null}
                    onClose={handleClose}
                    anchorReference="anchorPosition"
                    anchorPosition={
                        contextMenu !== null
                            ? { top: contextMenu.mouseY, left: contextMenu.mouseX }
                            : undefined
                    }
                >
                    <MenuItem onClick={() => { handleClose(); CreateDirectory(path, dispatch); console.log("ESCAPE"); dispatch({ type: 'start_editing', file: path }) }}>
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