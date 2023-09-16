// React
import { useState, useEffect, useReducer, useMemo, memo, useRef } from 'react'
import { NavigateFunction, useNavigate, useParams } from 'react-router-dom'

// MUI
import Box from '@mui/material/Box'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import Chip from '@mui/material/Chip'
import { emphasize, styled } from '@mui/material/styles'
import DeleteIcon from '@mui/icons-material/Delete'
import { IconButton, LinearProgress } from '@mui/material'
import Menu from '@mui/material/Menu'
import MenuItem from '@mui/material/MenuItem'
import ListItemIcon from '@mui/material/ListItemIcon'
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder'

// Local
import { GetDirectoryData, CreateDirectory, RenameFile, DeleteFile } from '../../api/FileBrowserApi'
import Item from './FileItem'
import Presentation from '../../components/Presentation'
import HeaderBar from "../../components/HeaderBar"
import HandleFileUpload from "../../api/Upload"
import GetWebsocket, { dispatchSync } from '../../api/Websocket'
import { itemData } from '../../types/FileBrowserTypes'

// Other
import { useSnackbar } from 'notistack'

const mapToList = (dirMap) => {
    const newList = Object.keys(dirMap).map((key) => {
        return dirMap[key]
    })

    newList.sort((a, b) => {
        if (a.modTime > b.modTime) {
            return -1
        } else if (a.modTime < b.modTime) {
            return 1
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
            for (const item of action.items) {
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
            for (const item of action.files) {
                newMap[state.path + item.name] = { filepath: state.path + item.name, modTime: new Date().toString() }
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
                presentingHash: action.presentingHash
            }
        }

        case 'presentation_next': {
            return { ...state }
            let changeTo
            if (state.mediaIdMap[state.presentingHash].next) {
                changeTo = state.mediaIdMap[state.presentingHash].next
            } else {
                changeTo = state.presentingHash
            }
            return {
                ...state,
                presentingHash: changeTo
            }
        }

        case 'presentation_previous': {
            return { ...state }
            let changeTo
            if (state.mediaIdMap[state.presentingHash].previous) {
                changeTo = state.mediaIdMap[state.presentingHash].previous
            } else {
                changeTo = state.presentingHash
            }
            return {
                ...state,
                presentingHash: changeTo
            }
        }

        case 'stop_presenting': {
            document.documentElement.style.overflow = "visible"
            return {
                ...state,
                presentingHash: ""
            }
        }

        default: {
            console.error("Got unexpected dispatch type: ", action.type)
            return { ...state }
        }
    }
}

const StyledBreadcrumb = styled(Chip)(({ theme }) => {
    const backgroundColor =
        theme.palette.mode === 'light'
            ? theme.palette.grey[100]
            : theme.palette.grey[800]
    return {
        backgroundColor,
        height: theme.spacing(3),
        color: theme.palette.text.primary,
        fontWeight: theme.typography.fontWeightRegular,
        '&:hover, &:focus': {
            backgroundColor: emphasize(backgroundColor, 0.06),
        },
        '&:active': {
            boxShadow: theme.shadows[1],
            backgroundColor: emphasize(backgroundColor, 0.12),
        },
    }
}) as typeof Chip

const Crumbs = (path: string, navigate) => {
    path = path.slice(1)
    let parts = path.split('/')
    while (parts[parts.length - 1] == '') {
        parts.pop()
    }

    parts.unshift('/')
    const current = parts.pop()

    let crumbPaths = []
    for (let [index, val] of parts.entries()) {
        if (index == 0) {
            crumbPaths.push("/")
            continue
        } else {
            crumbPaths.push(crumbPaths[index - 1] + "/" + val)
        }
    }
    const crumbs = parts.map((part, i) => (
        <StyledBreadcrumb key={part} label={part == "/" ? "Home" : part} onClick={() => { navigate(`/files/${crumbPaths[i]}`.replace(/\/\/+/g, '/')) }} />)
    )
    crumbs.push(
        <StyledBreadcrumb key={current} label={current == "/" ? "Home" : current} />
    )
    return crumbs

}

const HandleDrag = (event, dispatch, dragging) => {
    event.preventDefault()
    if (event.type === "dragenter" || event.type === "dragover") {
        !dragging && dispatch({ type: "set_dragging", dragging: true })

    } else {
        dispatch({ type: "set_dragging", dragging: false })
    }
}

const HandleDrop = (event, path, dirMap, dispatch, sendMessage, enqueueSnackbar) => {
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
        HandleFileUpload(filteredFiles, path, sendMessage)
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

const FileBrowser = () => {

    const path = ("/" + useParams()["*"] + "/").replace(/\/\/+/g, '/')

    const [filebrowserState, dispatch] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, itemData>(),
        path: path,
        dragging: false,
        loading: false,
        presentingHash: "",
        numSelected: 0,
        holdingShift: false,
        lastSelected: "",
        editing: ""
    })

    const { enqueueSnackbar } = useSnackbar()
    const { sendMessage, lastMessage, readyState } = GetWebsocket(enqueueSnackbar)
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
        setScanProgress(0)
        GetDirectoryData(path, dispatch)
    }, [path])

    const crumbs = useMemo(() => {
        const crumbs = Crumbs(filebrowserState.path, navigate)
        return crumbs
    }, [filebrowserState.path])

    const dirItems = useMemo(() => {
        const itemsList = mapToList(filebrowserState.dirMap)
        const anyChecked = filebrowserState.numSelected > 0 ? true : false
        let scanRequired = false
        const items = itemsList.map((entry: itemData) => {
            if (!entry.imported && !entry.isDir) {
                scanRequired = true
            }
            return (
                <Box key={entry.filepath} >
                    <Item itemData={entry} editing={filebrowserState.editing == entry.filepath} dispatch={dispatch} anyChecked={anyChecked} navigate={navigate} />
                </Box>
            )
        })
        if (scanRequired && !alreadyScanned) { setAlreadyScanned(true); dispatch({ type: "set_loading", loading: true }); dispatchSync(path, sendMessage, false) }
        return items
    }, [filebrowserState.dirMap, filebrowserState.editing])

    return (
        <Box
            sx={{
                display: "flex",
                flexDirection: 'column',
            }}
        >
            <HeaderBar dispatch={dispatch} sendMessage={sendMessage} page={"files"} />
            <Box
                onDragOver={(event => HandleDrag(event, dispatch, filebrowserState.dragging))}
                onDragLeave={event => HandleDrag(event, dispatch, filebrowserState.dragging)}
                onDrop={(event => HandleDrop(event, path, filebrowserState.dirMap, dispatch, sendMessage, enqueueSnackbar))}
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
                {filebrowserState.presentingHash != "" && (
                    <Presentation fileHash={filebrowserState.presentingHash} dispatch={dispatch} />
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
                    <IconButton style={{ position: 'absolute', left: "5vw" }} onClick={() => { dispatch({ type: "delete_selected" }) }}>
                        <DeleteIcon />
                    </IconButton>
                )}
                <Box marginTop={2} marginBottom={2} width={"max-content"}>
                    <Breadcrumbs separator={"›"} >
                        {crumbs}
                    </Breadcrumbs>
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
                    <MenuItem onClick={() => { handleClose(); CreateDirectory(path, dispatch) }}>
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