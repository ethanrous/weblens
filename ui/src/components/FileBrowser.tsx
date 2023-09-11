import { useState, useEffect, useReducer, useMemo, memo, useRef } from 'react'
import FolderIcon from '@mui/icons-material/Folder'
import Box from '@mui/material/Box'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import CachedIcon from '@mui/icons-material/Cached'
import Chip from '@mui/material/Chip'
import { emphasize, styled } from '@mui/material/styles'
import Checkbox from '@mui/material/Checkbox'
import DeleteIcon from '@mui/icons-material/Delete';

import { NavigateFunction, useNavigate, useParams } from 'react-router-dom'

import { StyledLazyThumb } from './PhotoContainer'
import Presentation from './Presentation'
import HeaderBar from "./HeaderBar"
import HandleFileUpload from "./Upload"
import GetWebsocket, { dispatchSync } from './Websocket'
import { MediaData } from './Types'

import { IconButton, LinearProgress } from '@mui/material'
import Skeleton from '@mui/material/Skeleton'
import { useSnackbar } from 'notistack'

type dirMap = {
    [key: string]: itemData
}

type fileBrowserAction =
    | { type: 'set_path'; path: string }
    | { type: 'update_items'; item: [{}] }
    | { type: 'set_selected'; itempath: string, selected: boolean }
    | { type: 'clear_selected'; }
    | { type: 'holding_shift'; shift: boolean }
    | { type: 'set_loading'; loading: boolean }
    | { type: 'set_dragging'; dragging: boolean }
    | { type: 'set_presentation'; presentingHash: string }
    | { type: 'stop_presenting' }

type itemData = {
    filepath: string
    isDir: boolean
    imported: boolean
    modTime: string
    selected: boolean
    mediaData: MediaData
}

type ItemProps = {
    itemData: itemData
    dispatch: React.Dispatch<fileBrowserAction>
    navigate: NavigateFunction
}

const mapToList = (dirMap: dirMap) => {
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

const fileBrowserReducer = (state, action) => {
    switch (action.type) {
        case 'update_items': {
            const existingMap = { ...state.dirMap }
            //let existingMap: dirMap = { ...state.dirMap }
            for (const item of action.items) {
                item.selected = false
                existingMap[item.filepath] = item
            }

            return {
                ...state,
                dirMap: existingMap,
            }
        }

        case 'add_template_items': {
            let existingMap = { ...state.dirMap }
            for (const item of action.files) {
                existingMap[state.path + item.name] = { filepath: state.path + item.name, modTime: new Date().toString() }
            }
            return {
                ...state,
                dirMap: existingMap,
            }
        }

        case 'set_path': {
            const newPath = action.path.replace(/\/\/+/g, '/')
            let newMap
            if (newPath != state.path) {
                newMap = {}
            } else {
                newMap = { ...state.dirMap }
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

        case 'set_selected': {
            let map = { ...state.dirMap }
            let numSelected = state.numSelected

            if (state.holdingShift && state.numSelected > 0 && state.lastSelected != "") {
                const dirList = mapToList(state.dirMap)
                let startIndex = dirList.findIndex((val) => val.filepath == state.lastSelected)
                let endIndex = dirList.findIndex((val) => val.filepath == action.itempath)
                if (endIndex < startIndex) {
                    [startIndex, endIndex] = [endIndex, startIndex]
                }
                let changedCounter = 0
                console.log(startIndex, endIndex)
                for (const val of dirList.slice(startIndex, endIndex + 1)) {
                    map[val.filepath].selected = true
                    changedCounter += 1
                }
                numSelected += changedCounter
            } else {
                map[action.itempath].selected = action.selected
                if (action.selected) {
                    numSelected += 1
                } else {
                    numSelected -= 1
                }
            }
            return {
                ...state,
                dirMap: map,
                numSelected: numSelected,
                lastSelected: action.itempath
            }
        }

        case 'clear_selected': {
            let existingMap: dirMap = { ...state.dirMap }
            let newMap = new Map<string, itemData>()
            for (const item of Object.values(existingMap)) {
                item.selected = false
                newMap[item.filepath] = item
            }

            return {
                ...state,
                dirMap: newMap,
                numSelected: 0,
                lastSelected: ""
            }
        }

        case 'delete_selected': {
            let existingMap: dirMap = { ...state.dirMap }
            let newMap = new Map<string, itemData>()
            for (const val of Object.values(existingMap)) {
                if (!val.selected) {
                    newMap[val.filepath] = val
                }
            }
            return {
                ...state,
                dirMap: newMap,
                numSelected: 0
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

const GetDirectoryData = (path, dispatch) => {
    var url = new URL("http:localhost:3000/api/dirinfo")
    url.searchParams.append('path', ('/' + path).replace(/\/\/+/g, '/'))
    fetch(url.toString()).then((res) => res.json()).then((data) => {
        dispatch({
            type: 'update_items', items: data == null ? [] : data
        })
    })
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

const boxSX = {
    outline: "1px solid #00F0FF",
    color: 'gray',
    backgroundColor: 'lightblue'
}

const Item = ({ itemData, dispatch, anyChecked, navigate }) => {
    const hasInfo = itemData.mediaData ? true : false
    const [hovering, setHovering] = useState(false)
//console.log(itemData.selected)

    let itemVisual
    let unselectedAction: () => any
    if (itemData.isDir) {
        itemVisual = <FolderIcon style={{ width: "80%", height: "80%" }} onDragOver={() => { }} sx={{ cursor: "pointer" }} />
        unselectedAction = () => navigate(("/files/" + itemData.filepath).replace(/\/\/+/g, '/'))
    } else if (itemData.imported) {
        itemVisual = <StyledLazyThumb draggable={false} fileHash={itemData.mediaData.FileHash} blurhash={itemData.mediaData.BlurHash} width={"200px"} height={"200px"} sx={{ cursor: "pointer" }} />
        unselectedAction = () => dispatch({ type: 'set_presentation', presentingHash: itemData.mediaData.FileHash })
    } else {
        itemVisual = <Skeleton animation="wave" height={"100%"} width={"100%"} variant="rectangular" />
    }

    let textBox
    if (hasInfo) {
        textBox = <p style={{ overflowWrap: "break-word", maxWidth: "90%", margin: 0, color: "white" }}>{itemData.filepath.substring(itemData.filepath.lastIndexOf('/') + 1)}</p>
    } else {
        textBox = <Skeleton animation="wave" height={10} width="40%" />
    }

    const select = (e) => { dispatch({ type: 'set_selected', itempath: itemData.filepath, selected: !itemData.selected }) }

    return (
        <Box
            alignItems={"center"}
            height={"200px"}
            width={"200px"}
            overflow={"hidden"} borderRadius={"10px"}
            margin={"10px"}
            sx={itemData.selected ? boxSX : {}}
            onMouseOver={() => setHovering(true)}
            onMouseLeave={() => setHovering(false)}
            onClick={anyChecked ? select : unselectedAction}
        >
            {(hovering || itemData.selected) && hasInfo && (
                <Checkbox
                    checked={itemData.selected}
                    style={{ position: "absolute", zIndex: 2, boxShadow: "10px" }}
                    onChange={select}
                    onClick={(e) => { e.stopPropagation() }}
                />

            )}
            <Box display={"flex"} justifyContent={"center"} alignItems={"center"} position={"relative"} height={"80%"} width={"100%"} >
                {itemVisual}
            </Box>
            <Box position={"relative"} display={"flex"} justifyContent={"center"} alignItems={"center"} bgcolor={"rgb(0, 0, 0, 0.50)"} width={"100%"} height={"20%"} textAlign={"center"} >
                {textBox}
            </Box>

        </Box>
    )
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

    const path = (useParams()["*"] + "/").replace(/\/\/+/g, '/')

    const [filebrowserState, dispatch] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, itemData>(),
        path: path,
        dragging: false,
        loading: false,
        presentingHash: "",
        numSelected: 0,
        holdingShift: false,
        lastSelected: ""
    })

    const { enqueueSnackbar } = useSnackbar()
    const { sendMessage, lastMessage, readyState } = GetWebsocket(enqueueSnackbar)
    const [scanProgress, setScanProgress] = useState(0)
    const navigate = useNavigate()
    const [alreadyScanned, setAlreadyScanned] = useState(false)

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
                    enqueueSnackbar(msgData["error"], { variant: "error" })
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
                <Box key={entry.filepath} draggable>
                    <Item itemData={entry} dispatch={dispatch} anyChecked={anyChecked} navigate={navigate} />
                </Box>
            )
        })
        if (scanRequired && !alreadyScanned) { setAlreadyScanned(true); dispatch({ type: "set_loading", loading: true }); dispatchSync(path, sendMessage, false) }
        return items
    }, [filebrowserState.dirMap])

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
                                <p style={{ position: "absolute", left: "6vw" }}>Files are loading into datbase...</p>
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
            </Box>
        </Box >
    )
}

export default FileBrowser