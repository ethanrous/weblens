// React
import { useState, useEffect, useReducer, useMemo, useRef, useContext } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

// MUI
import { Box, Button, Tooltip, Badge, Typography, Card, CardContent, Divider } from '@mui/joy'
import { Delete, Download, CreateNewFolder, CheckBox, Person, Backup } from '@mui/icons-material'
// Other
import { useSnackbar } from 'notistack';

// Local
import Presentation from '../../components/Presentation'
import HeaderBar from "../../components/HeaderBar"
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import { GetDirectoryData, CreateDirectory } from '../../api/FileBrowserApi'
import { itemData, FileBrowserStateType } from '../../types/Types'
import useWeblensSocket, { dispatchSync } from '../../api/Websocket'
import { deleteSelected, GetDirItems, HandleDrop, HandleWebsocketMessage, downloadSelected, fileBrowserReducer, useKeyDown, changeOwner, useMousePosition } from './FileBrowserLogic';
import { DirItemsWrapper, DirViewWrapper, FlexColumnBox } from './FilebrowserStyles';
import { userContext } from '../../Context';

function SelectedActions({ dirMap, selectedMap, path, dispatch }) {
    const { authHeader, userInfo } = useContext(userContext)

    if (selectedMap.size == 0) {
        return null
    }

    return (
        <Box position={"absolute"} display={"flex"} left={"6vw"} paddingTop={'30px'}>
            <Tooltip title={"Unselect All (esc)"} disableInteractive>
                <Button style={{ padding: "15px" }} color='primary' onClick={() => { dispatch({ type: "clear_selected" }) }}>
                    <Badge badgeContent={selectedMap.size} max={999} color='neutral'>
                        <CheckBox />
                    </Badge>
                </Button>
            </Tooltip>

            <Tooltip title={"Download Selected"} disableInteractive >
                <Button style={{ padding: "15px" }} color='primary' onClick={() => { downloadSelected(selectedMap, path, dispatch, authHeader) }}>
                    <Download />
                </Button>
            </Tooltip>

            <Tooltip title={"Change Owner"} disableInteractive >
                <Button style={{ padding: "15px" }} color='primary' onClick={() => changeOwner(dirMap, dispatch, authHeader)}>
                    <Person />
                </Button>
            </Tooltip>

            <Tooltip title={"Delete Selected"} disableInteractive >
                <Button style={{ padding: "15px" }} color='primary' onClick={() => deleteSelected(selectedMap, dispatch, authHeader)}>
                    <Delete />
                </Button>
            </Tooltip>
        </Box>
    )
}

function UnselectedActions({ numSelected, dispatch, navigate, authHeader, path }) {
    if (numSelected !== 0) {
        return null
    }

    return (
        <Box position={"absolute"} display={"flex"} left={"6vw"} paddingTop={'30px'}>
            <Tooltip title={"New Folder"} disableInteractive>
                <Button style={{ padding: "15px" }} color='primary' onClick={() => { CreateDirectory(path, authHeader).then(newDir => { GetDirectoryData(path, dispatch, navigate, authHeader); dispatch({ type: 'start_editing', file: newDir }) }) }}>
                    <CreateNewFolder />
                </Button>
            </Tooltip>
        </Box>
    )
}

function DraggingCounter({ dragging, numSelected, dispatch }) {
    const position = useMousePosition()
    if (dragging != 1) {
        return null
    }
    return (
        <Box
            sx={{
                position: 'fixed',
                top: position.y + 8,
                left: position.x + 8,
                zIndex: 10
            }}
            onMouseUp={() => { dispatch({ type: 'set_dragging', dragging: false }) }}
        >

            <StyledBreadcrumb label={numSelected.toString()} />
        </Box>
    )
}

function Files({ filebrowserState, alreadyScanned, setAlreadyScanned, path, navigate, dispatch, wsSend, authHeader }) {
    const { items, scanRequired } = GetDirItems(filebrowserState, dispatch, authHeader)
    useEffect(() => {
        if (scanRequired && !alreadyScanned) { setAlreadyScanned(true); dispatch({ type: "set_loading", loading: true }); dispatchSync(filebrowserState.path, wsSend, false) }
    }, [scanRequired])
    if (items.length != 0) {
        return (
            <DirItemsWrapper>
                {items}
            </DirItemsWrapper>
        )
    } else if (!filebrowserState.loading) {
        return (
            <Card variant="solid" sx={{ height: 'max-content', top: '40vh', position: 'fixed' }}>
                <CardContent sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
                    <Typography level="title-md" display={'flex'} flexDirection={'column'} alignItems={'center'} padding={2} onClick={() => { }} sx={{ cursor: "pointer" }} >
                        <Backup sx={{ height: 100, width: 100 }} />
                        Upload
                        <Typography level="body-sm" display={'flex'} position={'absolute'} variant='plain' width={"100px"} justifyContent={'center'} paddingTop={'125px'}>
                            Click or Drop
                        </Typography>
                    </Typography>
                    <Divider orientation='vertical' >Or</Divider>
                    <Typography level="title-md" display={'flex'} flexDirection={'column'} alignItems={'center'} padding={2} onClick={() => { CreateDirectory(path, authHeader).then(newDir => { GetDirectoryData(path, dispatch, navigate, authHeader); dispatch({ type: 'start_editing', file: newDir }) }) }} sx={{ cursor: "pointer" }}>
                        <CreateNewFolder sx={{ height: 100, width: 100 }} />
                        New Folder
                    </Typography>
                </CardContent>
            </Card>
        )
    } else {
        return null
    }
}

const FileBrowser = () => {
    const realPath = ("/" + useParams()["*"] + "/").replace(/\/\/+/g, '/')
    const { authHeader } = useContext(userContext)
    const searchRef = useRef()
    const navigate = useNavigate()
    const { enqueueSnackbar } = useSnackbar()
    const { wsSend, lastMessage, readyState } = useWeblensSocket()
    const [alreadyScanned, setAlreadyScanned] = useState(false)

    const [filebrowserState, dispatch]: [FileBrowserStateType, React.Dispatch<any>] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, itemData>(),
        selected: new Map<string, boolean>(),
        path: realPath,
        draggingState: 0,
        loading: true,
        presentingPath: "",
        searchContent: "",
        scanProgress: 0,
        holdingShift: false,
        lastSelected: "",
        editing: "",
        hovering: "",
    })

    useKeyDown(filebrowserState.editing, dispatch, searchRef)

    useEffect(() => {
        if (readyState === 1) {
            wsSend(JSON.stringify({ type: "subscribe", content: { path: filebrowserState.path, recursive: false }, error: null }))
            GetDirectoryData(filebrowserState.path, dispatch, navigate, authHeader)
        }
    }, [readyState])

    useEffect(() => {
        HandleWebsocketMessage(lastMessage, filebrowserState.path, dispatch, navigate, enqueueSnackbar, authHeader)
    }, [lastMessage])

    useEffect(() => {
        if (realPath !== filebrowserState.path) {
            navigate("/files" + filebrowserState.path)
        }

        // Kinda just reset everything...
        document.documentElement.style.overflow = "visible"
        setAlreadyScanned(false)
        dispatch({ type: "clear_items" })
        dispatch({ type: "set_scan_progress", progress: 0 })
        wsSend(JSON.stringify({ type: "subscribe", content: { path: filebrowserState.path, recursive: false }, error: null }))
        GetDirectoryData(filebrowserState.path, dispatch, navigate, authHeader)
    }, [filebrowserState.path])

    return (
        <FlexColumnBox>
            <HeaderBar
                path={filebrowserState.path}
                dispatch={dispatch}
                wsSend={wsSend}
                page={"files"}
                searchRef={searchRef}
                loading={filebrowserState.loading}
                progress={filebrowserState.scanProgress}
            />
            <DraggingCounter dragging={filebrowserState.draggingState} numSelected={filebrowserState.selected.size} dispatch={dispatch} />
            <Presentation mediaData={filebrowserState.dirMap.get(filebrowserState.presentingPath)?.mediaData} dispatch={dispatch} />
            <DirViewWrapper
                path={filebrowserState.path}
                dragging={filebrowserState.draggingState}
                hoverTarget={filebrowserState.hovering}
                onDrop={(e => HandleDrop(e, filebrowserState.path, filebrowserState.dirMap, dispatch, wsSend, enqueueSnackbar, authHeader))}
                dispatch={dispatch}
                onMouseOver={e => dispatch({ type: 'set_hovering', itempath: "" })}
            >
                <SelectedActions dirMap={filebrowserState.dirMap} selectedMap={filebrowserState.selected} path={filebrowserState.path} dispatch={dispatch} />
                <UnselectedActions numSelected={filebrowserState.selected.size} dispatch={dispatch} navigate={navigate} authHeader={authHeader} path={filebrowserState.path} />
                <Crumbs path={filebrowserState.path} includeHome={true} dispatch={dispatch} navOnLast={true} navigate={(newPath) => dispatch({ type: 'set_path', path: newPath })} />

                <Files filebrowserState={filebrowserState} alreadyScanned={alreadyScanned} setAlreadyScanned={setAlreadyScanned} path={filebrowserState.path} navigate={navigate} dispatch={dispatch} wsSend={wsSend} authHeader={authHeader} />

            </DirViewWrapper>
        </FlexColumnBox>
    )
}

export default FileBrowser