// React
import { useState, useEffect, useReducer, useMemo, useRef, useContext } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

// MUI
import { Typography, Card, CardContent, Divider } from '@mui/joy'

// Icons
import { IconDownload, IconFolderPlus, IconHome, IconShare, IconTrash, IconUpload, IconUsers } from "@tabler/icons-react"

// Mantine
import { Box, Button, Text, Space, FileButton } from '@mantine/core'

// Other
import { useSnackbar } from 'notistack'

// Local
import Presentation from '../../components/Presentation'
import HeaderBar from "../../components/HeaderBar"
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import { GetFolderData } from '../../api/FileBrowserApi'
import { itemData, FileBrowserStateType } from '../../types/Types'
import useWeblensSocket, { dispatchSync } from '../../api/Websocket'
import { deleteSelected, GetDirItems, HandleDrop, HandleWebsocketMessage, downloadSelected, fileBrowserReducer, useKeyDown, useMousePosition, moveSelected } from './FileBrowserLogic'
import { DirItemsWrapper, DirViewWrapper, FlexColumnBox, FlexRowBox } from './FilebrowserStyles'
import { userContext } from '../../Context'
import UploadStatus, { useUploadStatus } from '../../components/UploadStatus'
import ShareDialogue from './Share'
import { useDebouncedValue } from '@mantine/hooks'

function GlobalActions({ folderId, selectedMap, dirMap, dragging, dispatch, wsSend, uploadDispatch, authHeader }) {
    const nav = useNavigate()
    const { userInfo } = useContext(userContext)
    const amHome = folderId === userInfo?.homeFolderId
    const inShared = folderId === "shared"
    const numFilesIOwn = Array.from(selectedMap.keys()).filter((key) => dirMap.get(key)?.owner === userInfo.username).length
    return (
        <Box pr={0} top={150} h={'max-content'} display={'flex'} pos={'sticky'} style={{ marginLeft: "16px", flexDirection: 'column' }} >
            <Button m={3} disabled={dragging || amHome} justify='space-between' rightSection={<IconHome />} onClick={() => { nav('/files/home') }} >
                My Files
            </Button>
            <Button m={3} disabled={dragging || inShared} justify='space-between' rightSection={<IconUsers />} onClick={() => { nav('/files/shared') }} >
                Shared With Me
            </Button>
            <Space h={"md"} />
            <Button disabled={dragging || inShared} m={3} justify='space-between' rightSection={<IconFolderPlus />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'new_dir' }) }}>
                New Folder
            </Button>
            <FileButton onChange={(files) => { HandleDrop(files, folderId, dirMap, authHeader, uploadDispatch, dispatch, wsSend) }} accept="file" multiple>
                {(props) => {
                    return (
                        <Button disabled={dragging || inShared} m={3} justify='space-between' rightSection={<IconUpload />} onClick={() => props.onClick()}>
                            Upload
                        </Button>

                    )
                }}

            </FileButton>
            <Space h={"md"} />
            <Button m={3} disabled={dragging || selectedMap.size === 0} justify='space-between' leftSection={<Text>{selectedMap.size}</Text>} rightSection={<IconDownload />} onClick={(e) => { e.stopPropagation(); downloadSelected(selectedMap, dirMap, folderId, dispatch, wsSend, authHeader) }} >
                Download
            </Button>
            <Button m={3} disabled={dragging || numFilesIOwn === 0} justify='space-between' leftSection={<Text>{numFilesIOwn}</Text>} rightSection={<IconShare />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'share_selected' }) }} >
                Share
            </Button>
            <Space h={"md"} />
            <Button m={3} color='red' disabled={dragging || numFilesIOwn === 0} justify='space-between' leftSection={<Text>{numFilesIOwn}</Text>} rightSection={<IconTrash />} onClick={(e) => { e.stopPropagation(); deleteSelected(selectedMap, dirMap, folderId, dispatch, authHeader) }} >
                Delete
            </Button>
        </Box>
    )
}

function DraggingCounter({ dragging, numSelected, dispatch }) {
    const position = useMousePosition()
    if (dragging !== 1) {
        return null
    }
    return (
        <Box
            style={{
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

function Files({ filebrowserState, folderId, alreadyScanned, setAlreadyScanned, dispatch, wsSend, uploadDispatch, authHeader }) {
    const gridRef = useRef()
    const nav = useNavigate()
    const [debouncedSearch] = useDebouncedValue(filebrowserState.searchContent, 200)

    const { items, scanRequired } = useMemo(() => {
        return GetDirItems(filebrowserState, dispatch, authHeader, gridRef)
    }, [debouncedSearch, filebrowserState, dispatch, authHeader, gridRef])

    useEffect(() => {
        if (scanRequired && !alreadyScanned) { setAlreadyScanned(true); dispatchSync(filebrowserState.folderInfo.id, wsSend, false) }
    }, [scanRequired])


    if (items.length !== 0) {
        return (
            <DirItemsWrapper reff={gridRef}>
                {items}
            </DirItemsWrapper>
        )
    } else if (!filebrowserState.loading && folderId !== "shared") {
        return (
            <Box display={'flex'} style={{ justifyContent: 'center' }}>
                <Card variant="solid" sx={{ height: 'max-content', top: '40vh', position: 'fixed' }}>
                    <Text ta={'center'} fw={600} mt={"sm"}>
                        This folder is empty
                    </Text>
                    <CardContent sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
                        <FileButton onChange={(files) => { HandleDrop(files, folderId, filebrowserState.dirMap, authHeader, uploadDispatch, dispatch, wsSend) }} accept="file" multiple>
                            {(props) => {
                                return (
                                    <Typography level="title-md" display={'flex'} flexDirection={'column'} alignItems={'center'} padding={2} sx={{ cursor: "pointer" }} onClick={() => { props.onClick() }}>
                                        <IconUpload size={100} style={{ padding: "10px" }} />
                                        Upload
                                        <Typography level="body-sm" display={'flex'} position={'absolute'} variant='plain' width={"100px"} justifyContent={'center'} paddingTop={'125px'}>
                                            Click or Drop
                                        </Typography>
                                    </Typography>
                                )
                            }}
                        </FileButton>
                        <Divider orientation='vertical' >Or</Divider>
                        <Typography level="title-md" display={'flex'} flexDirection={'column'} alignItems={'center'} padding={2} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'new_dir' }) }} sx={{ cursor: "pointer" }}>
                            <IconFolderPlus size={100} style={{ padding: "10px" }} />
                            New Folder
                        </Typography>
                    </CardContent>
                </Card>
            </Box>
        )
    } else if (!filebrowserState.loading && folderId === "shared") {
        return (
            <Card variant="solid" sx={{ height: 'max-content', top: '40vh', position: 'fixed' }}>
                <CardContent sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                    <Typography level="title-md" display={'flex'} flexDirection={'column'} alignItems={'center'} padding={2} >
                        You have no files shared with you
                    </Typography>
                    <Button onClick={() => nav('/files/home')}>Return Home</Button>
                </CardContent>
            </Card>
        )
    } else {
        return null
    }
}

const FileBrowser = () => {
    const folderId = useParams()["*"]
    const navigate = useNavigate()
    const { authHeader, userInfo } = useContext(userContext)
    const searchRef = useRef()
    const { enqueueSnackbar } = useSnackbar()
    const { wsSend, lastMessage, readyState } = useWeblensSocket()
    const [alreadyScanned, setAlreadyScanned] = useState(false)
    const { uploadState, uploadDispatch } = useUploadStatus()

    const [filebrowserState, dispatch]: [FileBrowserStateType, React.Dispatch<any>] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, itemData>(),
        selected: new Map<string, boolean>(),
        uploadMap: new Map<string, boolean>(),
        folderInfo: {},
        parents: [],
        draggingState: 0,
        loading: true,
        presentingId: "",
        searchContent: "",
        scanProgress: 0,
        holdingShift: false,
        sharing: false,
        lastSelected: "",
        editing: null,
        hovering: "",
    })

    useKeyDown(filebrowserState.editing, filebrowserState.sharing, dispatch, searchRef)

    const realId = useMemo(() => {
        let realId
        if (userInfo) {
            realId = (folderId === "home") ? userInfo.homeFolderId : folderId
        }
        return realId
    }, [folderId, userInfo])

    useEffect(() => {
        if (readyState === 1 && realId != null) {
            wsSend(JSON.stringify({ req: "subscribe", content: { subType: "folder", folderId: realId, recursive: false }, error: null }))
        }
    }, [readyState])

    useEffect(() => {
        if (!userInfo) {
            return
        }
        HandleWebsocketMessage(lastMessage, userInfo.username, dispatch, authHeader)
    }, [lastMessage, userInfo])

    useEffect(() => {
        if (!userInfo) {
            return
        }
        if (!folderId || folderId === userInfo.homeFolderId || folderId == "undefined") {
            navigate('/files/home')
        }
        // Kinda just reset everything...
        setAlreadyScanned(false)
        dispatch({ type: "clear_items" })
        dispatch({ type: "set_search", search: "" })
        dispatch({ type: "set_scan_progress", progress: 0 })
        wsSend(JSON.stringify({ req: "subscribe", content: { subType: "folder", folderId: realId, recursive: false }, error: null }))
        GetFolderData(realId, userInfo.username, dispatch, navigate, authHeader)
    }, [folderId, userInfo])

    return (
        <FlexColumnBox style={{ backgroundColor: "#111418" }} >
            <HeaderBar
                folderId={folderId}
                searchContent={filebrowserState.searchContent}
                dispatch={dispatch}
                wsSend={wsSend}
                page={"files"}
                searchRef={searchRef}
                loading={filebrowserState.loading}
                progress={filebrowserState.scanProgress}
            />
            <DraggingCounter dragging={filebrowserState.draggingState} numSelected={filebrowserState.selected.size} dispatch={dispatch} />
            <Presentation mediaData={filebrowserState.dirMap.get(filebrowserState.presentingId)?.mediaData} parents={[...filebrowserState.parents, filebrowserState.folderInfo]} dispatch={dispatch} />
            <UploadStatus uploadState={uploadState} uploadDispatch={uploadDispatch} count={uploadState.uploadsMap.size} />
            <ShareDialogue sharing={filebrowserState.sharing} selectedMap={filebrowserState.selected} dirMap={filebrowserState.dirMap} dispatch={dispatch} authHeader={authHeader} />
            <FlexRowBox style={{ height: "calc(100vh - 70px)" }}>
                <GlobalActions folderId={filebrowserState.folderInfo.id} selectedMap={filebrowserState.selected} dirMap={filebrowserState.dirMap} dragging={filebrowserState.draggingState} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} authHeader={authHeader} />
                <DirViewWrapper
                    folderName={filebrowserState.folderInfo?.filename}
                    dragging={filebrowserState.draggingState}
                    hoverTarget={filebrowserState.hovering}
                    onDrop={(e => { e.preventDefault(); e.stopPropagation(); dispatch({ type: "set_dragging", dragging: false }); HandleDrop(e.dataTransfer.items, realId, filebrowserState.dirMap, authHeader, uploadDispatch, dispatch, wsSend) })}
                    dispatch={dispatch}
                    onMouseOver={() => dispatch({ type: 'set_hovering', itempath: "" })}
                >
                    <Crumbs finalItem={filebrowserState.folderInfo} parents={filebrowserState.parents} navOnLast={false} dragging={filebrowserState.draggingState} moveSelectedTo={(folderId) => moveSelected(filebrowserState.selected, filebrowserState.dirMap, folderId, authHeader)} />
                    <Files filebrowserState={filebrowserState} folderId={realId} alreadyScanned={alreadyScanned} setAlreadyScanned={setAlreadyScanned} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} authHeader={authHeader} />
            </DirViewWrapper>
            </FlexRowBox>
        </FlexColumnBox>
    )
}

export default FileBrowser