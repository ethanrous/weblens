import { useState, useEffect, useReducer, memo } from 'react'
import FolderIcon from '@mui/icons-material/Folder'
import Box from '@mui/material/Box'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import CachedIcon from '@mui/icons-material/Cached'
import Chip from '@mui/material/Chip'
import { emphasize, styled } from '@mui/material/styles'

import { useNavigate, useParams } from 'react-router-dom'
import useWebSocket from 'react-use-websocket'

import { StyledLazyThumb } from './PhotoContainer'
import Presentation from './Presentation'
import HeaderBar from "./HeaderBar"
import HandleFileUpload from "./Upload"
import { LinearProgress } from '@mui/material'
import { useSnackbar } from 'notistack';


const fileBrowserReducer = (state, action) => {
    switch (action.type) {
        case 'set_dir_info': {
            return {
                ...state,
                dirInfo: action.data,
            }
        }
        case 'append_items': {
            let existing = [...state.dirInfo]

            existing.push(...action.items)

            existing.sort((a, b) => {
                if (a.modTime > b.modTime) {
                    return -1
                } else if (a.modTime < b.modTime) {
                    return 1
                } else {
                    return 0
                }
            })
            return {
                ...state,
                dirInfo: existing,
            }
        }
        case 'set_path': {
            return {
                ...state,
                path: action.path.replace(/\/\/+/g, '/'),
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
    }
}

const dispatchSync = (itemPath) => {
    var url = new URL("http://localhost:3000/api/scan")
    url.searchParams.append('path', itemPath)
    fetch(
        url.toString(),
        {
            method: 'POST',
        }
    )
}

const GetDirectoryData = (path, dispatch) => {
    var url = new URL("http:localhost:3000/api/dirinfo")
    url.searchParams.append('path', ('/' + path).replace(/\/\/+/g, '/'))
    fetch(url.toString()).then((res) => res.json()).then((data) => {
        dispatch({
            type: 'set_dir_info', data: data == null ? [] : data
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

const Crumbs = (path) => {
    let navigate = useNavigate()
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

export type fileBrowserAction =
    | { type: 'set_dir_info'; data: [] }
    | { type: 'set_path'; path: string }
    | { type: 'append_items'; item: [{}] }
    | { type: 'set_loading'; loading: boolean }
    | { type: 'set_dragging'; dragging: boolean }
    | { type: 'set_presentation'; presentingHash: string }
    | { type: 'stop_presenting' }

type ItemProps = {
    itemPath: string
    fileHash: string
    blurhash: string
    isDir: boolean
    isImported: boolean
    dispatch: React.Dispatch<fileBrowserAction>
}

const Item = memo(function Item({ itemPath, fileHash, blurhash, isDir, isImported, dispatch }: ItemProps) {
    let navigate = useNavigate()
    let itemVisual
    if (isDir) {
        itemVisual = <FolderIcon style={{ width: "80%", height: "80%" }} onDragOver={() => { }} onClick={() => navigate(("/files/" + itemPath).replace(/\/\/+/g, '/'))} sx={{ cursor: "pointer" }} />
    } else if (isImported) {
        itemVisual = <StyledLazyThumb draggable={false} fileHash={fileHash} blurhash={blurhash} width={"200px"} height={"200px"} onClick={() => dispatch({ type: 'set_presentation', presentingHash: fileHash })} sx={{ cursor: "pointer" }} />
    } else {
        itemVisual = <CachedIcon style={{ width: "50%", height: "50%" }} onClick={() => dispatchSync(itemPath)} sx={{ cursor: "pointer" }} />
    }

    return (
        <Box
            alignItems={"center"}
            height={"200px"}
            width={"200px"}
            overflow={"hidden"} borderRadius={"10px"}
            margin={"10px"}
        >
            <Box display={"flex"} justifyContent={"center"} alignItems={"center"} position={"relative"} height={"80%"} width={"100%"} >
                {itemVisual}
            </Box>
            <Box position={"relative"} display={"flex"} justifyContent={"center"} alignItems={"center"} bgcolor={"rgb(0, 0, 0, 0.50)"} width={"100%"} height={"20%"} textAlign={"center"} >
                <p style={{ overflowWrap: "break-word", maxWidth: "90%", margin: 0, color: "white" }}>{itemPath.substring(itemPath.lastIndexOf('/') + 1)}</p>
            </Box>

        </Box>
    )
}
)

const HandleDrag = (event, dispatch, dragging) => {
    event.preventDefault()
    if (event.type === "dragenter" || event.type === "dragover") {
        !dragging && dispatch({ type: "set_dragging", dragging: true })

    } else {
        dispatch({ type: "set_dragging", dragging: false })
    }
}

const HandleDrop = (event, path, dispatch, sendMessage) => {
    event.preventDefault()
    event.stopPropagation()
    dispatch({ type: "set_dragging", dragging: false })
    dispatch({ type: "set_loading", loading: true })
    if (event.dataTransfer.files.length != 0) {
        HandleFileUpload(event.dataTransfer.files, path, sendMessage)
    }
}


const FileBrowser = () => {
    const WS_URL = 'ws://localhost:4000/api/ws';
    const { sendMessage, lastMessage, readyState } = useWebSocket(WS_URL, {
        onOpen: () => {
            console.log('WebSocket connection established.')
        }
    })

    const { enqueueSnackbar } = useSnackbar()

    const [dcTimeout, setDcTimeout] = useState(null)

    useEffect(() => {
        if (readyState != 1) {
            setDcTimeout(setTimeout(() => {
                enqueueSnackbar("Websocket connection lost, refresh? ¯\\_(ツ)_/¯", { variant: "error", persist: true })
            }, 2000))
        } else {
            clearTimeout(dcTimeout)
        }

    }, [readyState])

    useEffect(() => {
        if (lastMessage) {
            let msgData = JSON.parse(lastMessage.data)

            switch (msgData["type"]) {
                case "new_items": {
                    dispatch({ type: "append_items", items: msgData["content"] })
                    return
                }
                case "finished": {
                    dispatch({ type: "set_loading", loading: false })
                    return
                }
                case "refresh": {
                    GetDirectoryData(path, dispatch)
                    return
                }
                case "error": {
                    enqueueSnackbar(msgData["error"], { variant: "error" })
                    return
                }
                default: {
                    console.log("I dunno man")
                    console.log(msgData)
                    return
                }
            }
        }
    }, [lastMessage])

    const path = (useParams()["*"] + "/").replace(/\/\/+/g, '/')

    const [filebrowserState, dispatch] = useReducer(fileBrowserReducer, {
        dirInfo: [],
        path: path,
        dragging: false,
        loading: false,
        presentingHash: "",
    })

    useEffect(() => {
        dispatch({ type: 'set_path', path: path })
        GetDirectoryData(path, dispatch)
    }, [path])

    const crumbs = Crumbs(filebrowserState.path)

    const dirItems = filebrowserState.dirInfo.map((entry) => {
        return (
            <Box key={entry.filepath} draggable>
                <Item itemPath={entry.filepath} fileHash={entry.mediaData.FileHash} blurhash={entry.mediaData.BlurHash} isDir={entry.isDir} isImported={entry.imported} dispatch={dispatch} />
            </Box>
        )
    })

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
                onDrop={(event => HandleDrop(event, path, dispatch, sendMessage))}
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
                        <LinearProgress />
                    </Box>
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