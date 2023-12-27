// React
import { useState, useEffect, useReducer, useMemo, useRef, useContext } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

// Icons
import { IconArrowRight, IconDownload, IconFolderPlus, IconHome, IconPhotoPlus, IconShare, IconTrash, IconUpload, IconUsers } from "@tabler/icons-react"

// Mantine
import { Box, Button, Text, Space, FileButton, Paper, Card, Divider, Popover, ScrollArea, Loader } from '@mantine/core'

// Local
import Presentation from '../../components/Presentation'
import HeaderBar from "../../components/HeaderBar"
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import { AddToAlbum, GetFolderData } from '../../api/FileBrowserApi'
import { itemData, FileBrowserStateType, AlbumData } from '../../types/Types'
import useWeblensSocket, { dispatchSync } from '../../api/Websocket'
import { deleteSelected, GetDirItems, HandleDrop, HandleWebsocketMessage, downloadSelected, fileBrowserReducer, useKeyDown, useMousePosition, moveSelected } from './FileBrowserLogic'
import { DirViewWrapper, FlexColumnBox, FlexRowBox } from './FilebrowserStyles'
import { userContext } from '../../Context'
import UploadStatus, { useUploadStatus } from '../../components/UploadStatus'
import ShareDialogue from './Share'
import { useDebouncedValue } from '@mantine/hooks'
import { ItemsWrapper, StyledLazyThumb } from '../../types/Styles'
import { GetAlbums } from '../../api/GalleryApi'
import { MediaImage } from '../../components/PhotoContainer'
import { notifications } from '@mantine/notifications'

function SingleAlbum({ album, PartialApiCall }: { album: AlbumData, PartialApiCall: (albumId: string) => void }) {
    const [hovered, setHovered] = useState(false)
    console.log(album)
    return (
        <FlexColumnBox
            style={{ height: '165px', width: '150px', cursor: 'pointer', padding: '5px', borderRadius: '5px', backgroundColor: hovered ? '#3333ee' : "", justifyContent: 'space-between' }}
            onClick={() => { PartialApiCall(album.Id) }}
            onMouseOver={() => { setHovered(true) }}
            onMouseLeave={() => { setHovered(false) }}
        >
            <MediaImage mediaId={album.Cover} quality='thumbnail' expectFailure containerStyle={{ borderRadius: '6px', overflow: 'hidden', width: '125px', height: '125px' }} />
            <Text>{album.Name}</Text>
        </FlexColumnBox>
    )
}

function AlbumScoller({ getSelected, dispatch, authHeader }: {
    getSelected: () => string[],
    dispatch,
    authHeader
}) {
    const [albums, setAlbums]: [albums: AlbumData[], setAlbums: any] = useState(null)
    const [partialApiCall, setPartialApiCall] = useState(null)
    const nav = useNavigate()

    useEffect(() => {
        GetAlbums(authHeader).then(ret => setAlbums(ret))
        const func = (albumId) => {
            AddToAlbum(getSelected(), albumId, authHeader).catch((r) => { notifications.show({ title: "Could not add media to album", message: r, color: 'red' }) })
            dispatch({ type: 'close_add_to' })
        }
        setPartialApiCall((_) => func)
    }, [])

    const albumElements = useMemo(() => {
        if (!albums || !partialApiCall) {
            return []
        }
        const albumElements = albums.map((val) => {
            return (
                <SingleAlbum key={val.Name} album={val} PartialApiCall={partialApiCall} />
            )
        })
        return albumElements
    }, [albums, partialApiCall])

    console.log(albumElements)
    if (!albumElements) {
        return (
            <Loader />
        )
    }
    if (albumElements.length == 0) {
        return (
            <FlexColumnBox style={{ height: '100px', width: '200px', justifyContent: 'center' }}>
                <Text>You don't have any albums</Text>
                <Space h={'md'}></Space>
                <Button fullWidth variant='light' rightSection={<IconArrowRight />} onClick={() => { nav('/albums') }}>Albums</Button>
            </FlexColumnBox>
        )
    }
    return (
        <ScrollArea.Autosize mah={1000} maw={310}>
            {albumElements}
        </ScrollArea.Autosize>
    )
}

function formatSelected(dirMap: Map<string, itemData>, selectedMap: Map<string, boolean>): string[] {
    const selectedObjs = Array.from(selectedMap.keys()).map((key) => {
        const item: itemData = dirMap.get(key)
        return item.id
    })
    return selectedObjs
}

function GlobalActions({ folderId, selectedMap, dirMap, dragging, adding, dispatch, wsSend, uploadDispatch }: { folderId, selectedMap: Map<string, boolean>, dirMap, dragging, dispatch, adding, wsSend, uploadDispatch, authHeader }) {
    const nav = useNavigate()
    const { userInfo, authHeader } = useContext(userContext)
    const amHome = folderId === userInfo?.homeFolderId
    const inShared = folderId === "shared"
    const numFiles = selectedMap.size
    const numFilesIOwn = Array.from(selectedMap.keys()).filter((key) => dirMap.get(key)?.owner === userInfo.username).length
    return (
        <Box pr={0} top={150} h={'max-content'} display={'flex'} pos={'sticky'} style={{ marginLeft: "16px", flexDirection: 'column' }} >
            <Button variant={amHome ? 'light' : 'subtle'} m={3} disabled={dragging} justify='flex-start' leftSection={<IconHome />} onClick={() => { nav('/files/home') }} >
                My Files
            </Button>
            <Button variant={inShared ? 'light' : 'subtle'} m={3} disabled={dragging} justify='flex-start' leftSection={<IconUsers />} onClick={() => { nav('/files/shared') }} >
                Shared With Me
            </Button>
            <Space h={"md"} />
            <Button variant='subtle' color='#eeeeee' disabled={dragging || inShared} m={3} justify='flex-start' leftSection={<IconFolderPlus />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'new_dir' }) }}>
                New Folder
            </Button>
            <FileButton onChange={(files) => { HandleDrop(files, folderId, dirMap, authHeader, uploadDispatch, dispatch, wsSend) }} accept="file" multiple>
                {(props) => {
                    return (
                        <Button variant='subtle' color='#eeeeee' disabled={dragging || inShared} m={3} justify='flex-start' leftSection={<IconUpload />} onClick={() => props.onClick()}>
                            Upload
                        </Button>
                    )
                }}

            </FileButton>
            <Space h={"md"} />
            <Popover opened={false} >
                <Popover.Target>
                    <Button variant='subtle' color='#eeeeee' m={3} disabled={dragging || numFilesIOwn === 0} justify='space-between' rightSection={<Text>{numFilesIOwn}</Text>} leftSection={<IconShare />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'share_selected' }) }} >
                        Share
                    </Button>
                </Popover.Target>
            </Popover>

            <Popover opened={adding} trapFocus position="right" onClose={() => { dispatch({ type: 'close_add_to' }) }}>
                <Popover.Target>
                    <Button variant='subtle' color='#eeeeee' m={3} disabled={dragging || numFilesIOwn === 0} justify='space-between' rightSection={<Text>{numFiles}</Text>} leftSection={<IconPhotoPlus />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'add_selected_to_album' }) }} >
                        Add to
                    </Button>
                </Popover.Target>
                <Popover.Dropdown style={{ width: 'max-content' }}>
                    <AlbumScoller getSelected={() => formatSelected(dirMap, selectedMap)} dispatch={dispatch} authHeader={authHeader} />
                </Popover.Dropdown>
            </Popover>

            <Button variant='subtle' color='#eeeeee' m={3} disabled={dragging || selectedMap.size === 0} justify='space-between' rightSection={<Text>{selectedMap.size}</Text>} leftSection={<IconDownload />} onClick={(e) => { e.stopPropagation(); downloadSelected(selectedMap, dirMap, folderId, dispatch, wsSend, authHeader) }} >
                Download
            </Button>
            <Space h={"md"} />
            <Button variant='subtle' color='red' m={3} disabled={dragging || numFilesIOwn === 0} justify='space-between' rightSection={<Text>{numFilesIOwn}</Text>} leftSection={<IconTrash />} onClick={(e) => { e.stopPropagation(); deleteSelected(selectedMap, dirMap, folderId, dispatch, authHeader) }} >
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
            <ItemsWrapper reff={gridRef}>
                {items}
            </ItemsWrapper>
        )
    } else if (!filebrowserState.loading && folderId !== "shared") {
        return (
            <Box display={'flex'} style={{ justifyContent: 'center' }}>
                <Card variant="solid" style={{ height: 'max-content', top: '40vh', position: 'fixed' }}>
                    <Text ta={'center'} fw={600} mt={"sm"}>
                        This folder is empty
                    </Text>
                    <Card.Section style={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
                        <FileButton onChange={(files) => { HandleDrop(files, folderId, filebrowserState.dirMap, authHeader, uploadDispatch, dispatch, wsSend) }} accept="file" multiple>
                            {(props) => {
                                return (
                                    <Text display={'flex'} style={{ flexDirection: 'column', alignItems: 'center', padding: 2, cursor: 'pointer' }} onClick={() => { props.onClick() }}>
                                        <IconUpload size={100} style={{ padding: "10px" }} />
                                        Upload
                                        <Text display={'flex'} style={{ position: 'absolute', width: '100px', justifyContent: 'center', paddingTop: '125px' }}>
                                            Click or Drop
                                        </Text>
                                    </Text>
                                )
                            }}
                        </FileButton>
                        <Divider orientation='vertical' >Or</Divider>
                        <Text display={'flex'} style={{ flexDirection: 'column', alignItems: 'center', padding: 2, cursor: 'pointer' }} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'new_dir' }) }}>
                            <IconFolderPlus size={100} style={{ padding: "10px" }} />
                            New Folder
                        </Text>
                    </Card.Section>
                </Card>
            </Box>
        )
    } else if (!filebrowserState.loading && folderId === "shared") {
        return (
            <Paper variant="solid" style={{ height: 'max-content', top: '40vh', position: 'fixed', padding: 40 }}>
                <FlexColumnBox style={{ alignItems: 'center' }}>
                    <Text display={'flex'} style={{ flexDirection: 'column', alignItems: 'center', padding: 2 }}>
                        You have no files shared with you
                    </Text>
                    <Space h={'lg'} />
                    <Button variant='light' fullWidth onClick={() => nav('/files/home')}>Return Home</Button>
                </FlexColumnBox>
            </Paper>
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
        albuming: false,
        lastSelected: "",
        editing: null,
        hovering: "",
    })

    useKeyDown(filebrowserState.editing || filebrowserState.albuming, filebrowserState.sharing, dispatch, searchRef)

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
        <FlexColumnBox style={{ height: '100vh', backgroundColor: "#111418" }} >
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
            {/* <ShareDialogue sharing={filebrowserState.sharing} selectedMap={filebrowserState.selected} dirMap={filebrowserState.dirMap} dispatch={dispatch} /> */}
            {/* <AddToDialogue albuming={filebrowserState.albuming} selectedMap={filebrowserState.selected} dirMap={filebrowserState.dirMap} dispatch={dispatch} /> */}
            <FlexRowBox style={{ height: "calc(100vh - 70px)", width: '100%' }}>
                <GlobalActions folderId={filebrowserState.folderInfo.id} selectedMap={filebrowserState.selected} dirMap={filebrowserState.dirMap} dragging={filebrowserState.draggingState} adding={filebrowserState.albuming} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} authHeader={authHeader} />
                <DirViewWrapper
                    folderName={filebrowserState.folderInfo?.filename}
                    dragging={filebrowserState.draggingState}
                    hoverTarget={filebrowserState.hovering}
                    onDrop={(e => { e.preventDefault(); e.stopPropagation(); dispatch({ type: "set_dragging", dragging: false }); HandleDrop(e.dataTransfer.items, realId, filebrowserState.dirMap, authHeader, uploadDispatch, dispatch, wsSend) })}
                    dispatch={dispatch}
                    onMouseOver={() => dispatch({ type: 'set_hovering', itempath: "" })}
                >
                    <Box style={{ width: '100%' }}>
                        <Crumbs finalItem={filebrowserState.folderInfo} parents={filebrowserState.parents} navOnLast={false} dragging={filebrowserState.draggingState} moveSelectedTo={(folderId) => moveSelected(filebrowserState.selected, filebrowserState.dirMap, folderId, authHeader)} />
                    </Box>
                    <Files filebrowserState={filebrowserState} folderId={realId} alreadyScanned={alreadyScanned} setAlreadyScanned={setAlreadyScanned} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} authHeader={authHeader} />
            </DirViewWrapper>
            </FlexRowBox>
        </FlexColumnBox>
    )
}

export default FileBrowser