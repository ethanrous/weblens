// React
import { useState, useEffect, useReducer, useMemo, useRef, useContext, useCallback } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

// Icons
import { IconCloud, IconDownload, IconExternalLink, IconFile, IconFileZip, IconFolder, IconFolderPlus, IconHome, IconPhotoPlus, IconPlus, IconShare, IconSpiral, IconTrash, IconUpload, IconUsers, IconUsersGroup } from "@tabler/icons-react"

// Mantine
import { Box, Button, Text, Space, FileButton, Paper, Card, Divider, Popover, ScrollArea, Loader, TextInput, Tooltip, LoadingOverlay, TooltipFloating, Progress, Menu, Center, Skeleton } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'

// Weblens
import Presentation, { PresentationContainer } from '../../components/Presentation'
import HeaderBar from "../../components/HeaderBar"
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import { DeleteFiles, GetFolderData, NewWormhole, ShareFiles } from '../../api/FileBrowserApi'
import { fileData, FileBrowserStateType, AlbumData, MediaData, FileBrowserAction, getBlankFile, FileBrowserDispatch } from '../../types/Types'
import useWeblensSocket, { dispatchSync } from '../../api/Websocket'
import { deleteSelected, HandleDrop, HandleWebsocketMessage, downloadSelected, fileBrowserReducer, useKeyDown, useMousePosition, moveSelected, HandleUploadButton, usePaste, UploadViaUrl, MapToList, HandleRename } from './FileBrowserLogic'
import { DirViewWrapper, ColumnBox, RowBox, ScanFolderButton } from './FilebrowserStyles'
import { userContext } from '../../Context'
import UploadStatus, { useUploadStatus } from '../../components/UploadStatus'
import { ItemScroller } from '../../components/ItemScroller'
import { AddMediaToAlbum, CreateAlbum, GetAlbums } from '../../api/GalleryApi'
import { MediaImage } from '../../components/PhotoContainer'
import { ShareInput } from '../../components/Share'
import { humanFileSize } from '../../util'
import NotFound from '../../components/NotFound'
import { getMedia } from '../../api/ApiFetch'
import { GlobalContextType, ItemProps } from '../../components/ItemDisplay'

function NewAlbum({ refreshAlbums }: { refreshAlbums: (doLoading) => Promise<void> }) {
    const { authHeader } = useContext(userContext)
    const [hovered, setHovered] = useState(false)
    const [newAlbumName, setNewAlbumName] = useState(null)
    const [loading, setLoading] = useState(false)

    if (newAlbumName == null) {
        return (
            <ColumnBox
                style={{ height: '100%', width: '100%', minHeight: '260px', cursor: 'pointer', padding: '5px', borderRadius: '5px', backgroundColor: hovered ? '#3333ee' : "", justifyContent: 'center' }}
                onClick={_ => setNewAlbumName("")}
                onMouseOver={() => { setHovered(true) }}
                onMouseLeave={() => { setHovered(false) }}
            >
                <IconPlus />
                <Text>New Album</Text>
            </ColumnBox>
        )
    } else {
        return (
            <ColumnBox style={{ justifyContent: 'center', height: '100%', minHeight: '260px' }}>
                <ColumnBox style={{ height: '85px', width: '100%', padding: '5px' }}>
                    <TextInput autoFocus onBlur={() => { if (!newAlbumName) { setHovered(false); setNewAlbumName(null) } }} placeholder="New Album Name" value={newAlbumName} onChange={e => setNewAlbumName(e.target.value)} style={{ width: '100%' }} />
                    <Space h={10} />
                    <Button fullWidth color='#4444ff' onClick={() => { setHovered(false); setLoading(true); CreateAlbum(newAlbumName, authHeader).then(() => { refreshAlbums(false).then(() => setNewAlbumName(null)); setLoading(false) }) }}>Create</Button>
                </ColumnBox>
                <LoadingOverlay visible={loading} />
            </ColumnBox>
        )
    }
}

function SingleAlbum({ album, loading, PartialApiCall }: { album: AlbumData, loading: boolean, PartialApiCall: (albumId: string) => void }) {
    const [hovered, setHovered] = useState(false)
    const { userInfo, authHeader } = useContext(userContext)
    const [coverMedia, setCoverMedia]: [MediaData, any] = useState(null)

    useEffect(() => {
        if (album.Cover) {
            getMedia(album.Cover, authHeader).then(c => setCoverMedia(c))
        }
    }, [album.Cover])

    return (
        <Box pos={'relative'}>
            <ColumnBox
                style={{ cursor: 'pointer', padding: '5px', borderRadius: '5px', backgroundColor: hovered ? '#3333ee' : "#ffffff21", justifyContent: 'space-between' }}
                onClick={() => { PartialApiCall(album.Id) }}
                onMouseOver={() => { setHovered(true) }}
                onMouseLeave={() => { setHovered(false) }}
            >
                {album.Owner !== userInfo.username && (
                    <Tooltip label={`Shared by ${album.Owner}`}>
                        <IconUsersGroup color="white" style={{ position: "absolute", alignSelf: 'flex-start', margin: 10, zIndex: 1 }} />
                    </Tooltip>
                )}
                <MediaImage media={coverMedia} quality='thumbnail' expectFailure={album.Cover === ""} containerStyle={{ borderRadius: '6px', overflow: 'hidden', width: '200px', height: '200px' }} />
                <RowBox style={{ height: '50px', justifyContent: 'space-between', paddingLeft: 4 }}>
                    <ColumnBox style={{ height: 'max-content', alignItems: 'flex-start', width: '175px' }}>
                        <TooltipFloating position='top' label={album.Name}>
                            <Text c='white' fw={500} truncate='end' w={'100%'}>{album.Name}</Text>
                        </TooltipFloating>
                        <Text size='12px'>{album.Medias.length}</Text>
                    </ColumnBox>
                    <TooltipFloating position='right' label='Open Album'>
                        <IconExternalLink onClick={(e) => { e.stopPropagation(); window.open(`/albums/${album.Id}`, '_blank') }} onMouseOver={(e) => { e.stopPropagation(); setHovered(false) }} />
                    </TooltipFloating>
                </RowBox>
            </ColumnBox>
            <LoadingOverlay visible={loading} overlayProps={{ radius: 'sm' }} />
        </Box>
    )
}

function AlbumScoller({ getSelected, authHeader }: {
    getSelected: () => { media: string[], folders: string[] },
    authHeader
}) {
    const [albums, setAlbums]: [albums: AlbumData[], setAlbums: any] = useState(null)
    // This is for the state if we are waiting for the list of albums
    const [loading, setLoading] = useState(false)

    // This is for tracking which album(s) are waiting
    // for results of adding media... naming is hard
    const [loadingAlbums, setLoadingAlbums] = useState([])

    const fetchAlbums = useCallback((doLoading) => {
        if (authHeader.Authorization === "") {
            return
        }
        if (doLoading) {
            setLoading(true)
        }
        return GetAlbums(authHeader).then(ret => { setAlbums(ret); setLoading(false) })
    }, [authHeader])

    const addMediaApiCall = useCallback((albumId) => {
        const { media, folders } = getSelected()
        setLoadingAlbums(cur => [...cur, albumId])
        AddMediaToAlbum(albumId, media, folders, authHeader)
            .then((res) => {
                if (res.errors.length === 0) {
                    setLoadingAlbums(cur => cur.filter(v => v !== albumId))
                    fetchAlbums(false)
                    if (res.addedCount === 0) {
                        notifications.show({ message: `No new media to add to album`, color: 'orange' })
                    } else {
                        notifications.show({ message: `Added ${res.addedCount} medias to album`, color: 'green' })
                    }
                } else {
                    Promise.reject(res.errors)
                }
            })
            .catch((r) => { notifications.show({ title: "Could not add media to album", message: String(r), color: 'red' }) })
    }, [getSelected, authHeader, fetchAlbums])

    useEffect(() => {
        fetchAlbums(true)
    }, [fetchAlbums])

    const albumElements = useMemo(() => {
        if (!albums || !addMediaApiCall) {
            return []
        }
        const albumElements = albums.map((val) => {
            return (
                <SingleAlbum key={val.Name} album={val} loading={loadingAlbums.includes(val.Id)} PartialApiCall={addMediaApiCall} />
            )
        })
        return albumElements
    }, [albums, addMediaApiCall, loadingAlbums])

    const columns = useMemo(() => { return Math.min(albumElements.length + 1, 4) }, [albumElements.length])

    if (loading || !albumElements) {
        return (
            <Loader />
        )
    }
    return (
        <ScrollArea.Autosize type='never' mah={1000} maw={1000}>
            <Box style={{ display: 'grid', gap: 16, gridTemplateColumns: `repeat(${columns}, 210px)` }}>
                {albumElements}
                <NewAlbum refreshAlbums={fetchAlbums} />
            </Box>
            {albumElements.length === 0 && (
                <Text style={{ textAlign: 'center' }}>No Albums</Text>
            )}
        </ScrollArea.Autosize>
    )
}

function selectedMediaIds(dirMap: Map<string, fileData>, selectedMap: Map<string, boolean>): string[] {
    let selectedObjs = Array.from(selectedMap.keys()).map((key) => {
        console.log(key, dirMap)
        const file: fileData = dirMap.get(key)
        return file.mediaData?.fileHash
    })
        .filter((val) => {
            return Boolean(val)
        })
    return selectedObjs
}

function selectedFolderIds(dirMap: Map<string, fileData>, selectedMap: Map<string, boolean>): string[] {
    let selectedObjs = Array.from(selectedMap.keys()).map((key) => {
        const file: fileData = dirMap.get(key)
        if (file.isDir) {
            return file.id
        }
        return ""
    })
    selectedObjs = selectedObjs.filter((val) => {
        return val !== ""
    })
    return selectedObjs
}

function ShareBox({ open, setOpen, fileIds, fetchFiles, dragging, numFilesIOwn }: { open: boolean, setOpen, fileIds: string[], fetchFiles, dragging, numFilesIOwn }) {
    const { authHeader } = useContext(userContext)
    const [value, setValue] = useState([])

    return (
        <Popover opened={open} position='right' onClose={() => setOpen(false)} closeOnClickOutside>
            <Popover.Target>
                <Button style={{ width: 190 }} variant='subtle' color='#eeeeee' m={3} disabled={dragging || numFilesIOwn === 0} justify='space-between' rightSection={<Text>{numFilesIOwn}</Text>} leftSection={<IconShare />} onClick={(e) => { e.stopPropagation(); setOpen(true) }} >
                    Share
                </Button>
            </Popover.Target>
            <Popover.Dropdown>
                <ShareInput valueSetCallback={setValue} initValues={[]} />
                <Space h={10} />
                <Button fullWidth disabled={JSON.stringify(value) === JSON.stringify([])} color="#4444ff" onClick={() => { ShareFiles(fileIds, value, authHeader).then(() => { notifications.show({ message: "File(s) shared", color: 'green' }); setOpen(false) }).catch((r) => notifications.show({ title: "Failed to share files", message: String(r), color: 'red' })) }}>
                    Update
                </Button>
            </Popover.Dropdown>
        </Popover>
    )
}

function PasteImageDialogue({ img, dirMap, folderId, authHeader, dispatch, wsSend }: { img: ArrayBuffer, dirMap: Map<string, fileData>, folderId, authHeader, dispatch, wsSend }) {
    if (!img) {
        return null
    }
    const media: MediaData = {} as MediaData
    media.thumbnail = img

    return (
        <PresentationContainer shadeOpacity={"0.25"} onClick={() => dispatch({ type: "paste_image", img: null })} >
            <ColumnBox style={{ position: 'absolute', justifyContent: 'center', alignItems: 'center', zIndex: 2 }}>
                <Text fw={700} size='40px' style={{ paddingBottom: '50px' }}>Upload from clipboard?</Text>
                <ColumnBox onClick={(e) => { e.stopPropagation() }} style={{ height: "50%", width: "max-content", backgroundColor: '#222277ee', padding: '10px', borderRadius: '8px' }}>
                    <MediaImage media={media} quality='thumbnail' imgStyle={{ objectFit: "contain", maxHeight: "100%", height: "100%" }} />
                </ColumnBox>
                <RowBox style={{ justifyContent: 'space-between', width: '300px', height: '150px' }}>
                    <Button size='xl' variant='default' onClick={(e) => { e.stopPropagation(); dispatch({ type: "paste_image", img: "" }) }}>Cancel</Button>
                    <Button size='xl' color='#4444ff' onClick={e => { e.stopPropagation(); UploadViaUrl(img, folderId, dirMap, authHeader, dispatch, wsSend) }}>Upload</Button>
                </RowBox>
            </ColumnBox>
        </PresentationContainer>
    )
}

function GlobalActions({ fbState, dispatch, wsSend, uploadDispatch }: { fbState: FileBrowserStateType, dispatch: FileBrowserDispatch, wsSend: (action: string, content: any) => void, uploadDispatch }) {
    const nav = useNavigate()
    const { userInfo, authHeader } = useContext(userContext)
    const [sharing, setSharing] = useState(false)
    const [adding, setAdding] = useState(false)
    const amHome = fbState.folderInfo.id === userInfo?.homeFolderId
    const numFilesIOwn = useMemo(() => Array.from(fbState.selected.keys()).filter((key) => fbState.dirMap.get(key)?.owner === userInfo.username).length, [fbState.selected.size, fbState.dirMap, fbState.selected, userInfo?.username])

    const BUTTON_WIDTH = 190

    return (
        <ColumnBox style={{ paddingLeft: 15, paddingRight: 15, paddingTop: 60, width: 230, }}>
            <Button style={{ width: BUTTON_WIDTH }} variant={amHome ? 'light' : 'subtle'} color={'#4444ff'} m={3} disabled={fbState.draggingState !== 0} justify='flex-start' leftSection={<IconHome />} onClick={() => { nav('/files/home') }} >
                My Files
            </Button>
            <Button style={{ width: BUTTON_WIDTH }} variant={fbState.folderInfo.id === "shared" ? 'light' : 'subtle'} color={'#4444ff'} m={3} disabled={fbState.draggingState !== 0} justify='flex-start' leftSection={<IconUsers />} onClick={() => { nav('/files/shared') }} >
                Shared With Me
            </Button>
            <Button style={{ width: BUTTON_WIDTH }} variant={fbState.folderInfo.id === userInfo?.trashFolderId ? 'light' : 'subtle'} color={'#4444ff'} m={3} disabled={fbState.draggingState !== 0} justify='flex-start' leftSection={<IconTrash />} onClick={() => { nav('/files/trash') }} >
                Trash
            </Button>
            <Space h={"md"} />
            <Button style={{ width: BUTTON_WIDTH }} variant='subtle' color='#eeeeee' disabled={fbState.draggingState !== 0 || !fbState.folderInfo.modifiable} m={3} justify='flex-start' leftSection={<IconFolderPlus />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'new_dir' }) }}>
                New Folder
            </Button>
            <FileButton onChange={(files) => { HandleUploadButton(files, fbState.folderInfo.id, false, "", authHeader, uploadDispatch, wsSend) }} accept="file" multiple>
                {(props) => {
                    return (
                        <Button style={{ width: BUTTON_WIDTH }} variant='subtle' color='#eeeeee' disabled={fbState.draggingState !== 0 || !fbState.folderInfo.modifiable} m={3} justify='flex-start' leftSection={<IconUpload />} onClick={() => props.onClick()}>
                            Upload
                        </Button>
                    )
                }}
            </FileButton>
            <Space h={"md"} />
            <ShareBox open={sharing} setOpen={(open) => { setSharing(open); dispatch({ type: 'set_block_focus', block: open }) }} fileIds={Array.from(fbState.selected.keys())} fetchFiles={() => { }} dragging={fbState.draggingState} numFilesIOwn={numFilesIOwn} />
            <Popover opened={adding} trapFocus position="right-end" onClose={() => { setAdding(false); dispatch({ type: 'set_block_focus', block: false }) }} styles={{ dropdown: { marginTop: 100 } }}>
                <Popover.Target>
                    <Button style={{ width: BUTTON_WIDTH }} variant='subtle' color='#eeeeee' m={3} disabled={fbState.draggingState !== 0 || fbState.selected.size === 0} justify='space-between' rightSection={<Text>{fbState.selected.size}</Text>} leftSection={<IconPhotoPlus />} onClick={(e) => { e.stopPropagation(); setAdding(true); dispatch({ type: 'set_block_focus', block: true }) }}>
                        Add to
                    </Button>
                </Popover.Target>
                <Popover.Dropdown style={{ width: 'max-content' }}>
                    <AlbumScoller getSelected={() => { return { media: selectedMediaIds(fbState.dirMap, fbState.selected), folders: selectedFolderIds(fbState.dirMap, fbState.selected) } }} authHeader={authHeader} />
                </Popover.Dropdown>
            </Popover>

            <Button style={{ width: BUTTON_WIDTH }} variant='subtle' color='#eeeeee' m={3} disabled={fbState.draggingState !== 0 || fbState.selected.size === 0} justify='space-between' rightSection={<Text>{fbState.selected.size}</Text>} leftSection={<IconDownload />} onClick={(e) => { e.stopPropagation(); downloadSelected(fbState.selected, fbState.dirMap, dispatch, wsSend, authHeader) }} >
                Download
            </Button>
            <Space h={"md"} />
            <Button style={{ width: BUTTON_WIDTH }} variant='subtle' color='red' m={3} disabled={fbState.draggingState !== 0 || numFilesIOwn === 0} justify='space-between' rightSection={<Text>{numFilesIOwn}</Text>} leftSection={<IconTrash />} onClick={(e) => { e.stopPropagation(); deleteSelected(fbState.selected, fbState.dirMap, authHeader) }} >
                {(fbState.folderInfo.id === userInfo.trashFolderId || fbState.parents.map(v => v.id).includes(userInfo?.trashFolderId)) ? "Delete" : "Trash"}
            </Button>

            <Divider w={'100%'} my='lg' size={1.5} />

            <UsageInfo inHome={fbState.folderInfo.id === userInfo.homeFolderId} homeDirSize={fbState.homeDirSize} currentFolderSize={fbState.folderInfo.size} displayCurrent={fbState.folderInfo.id !== "shared"} trashSize={fbState.trashDirSize} />
        </ColumnBox>
    )
}

function UsageInfo({ inHome, homeDirSize, currentFolderSize, displayCurrent, trashSize }) {
    if (!displayCurrent) {
        currentFolderSize = homeDirSize
    }
    if (!trashSize) {
        trashSize = 0
    }

    if (inHome) {
        currentFolderSize = currentFolderSize - trashSize
    }

    return (
        <ColumnBox style={{ height: 'max-content', width: '100%', alignItems: 'flex-start' }}>
            <Text fw={600}>Usage</Text>
            <Space h={10} />
            <Progress radius={'xs'} color='#4444ff' value={(currentFolderSize / homeDirSize) * 100} style={{ height: 20, width: '100%' }} />
            <RowBox>
                {displayCurrent && (
                    <RowBox>
                        <IconFolder size={20} />
                        <Text size='14px' pl={3}>{humanFileSize(currentFolderSize)}</Text>
                    </RowBox>
                )}
                <RowBox style={{ justifyContent: 'right' }}>
                    <Text size='14px' pr={3}>{humanFileSize(homeDirSize)}</Text>
                    <IconCloud size={20} />
                </RowBox>
            </RowBox>
        </ColumnBox>
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

function FileContextMenu({ open, setOpen, itemInfo, menuPos, authHeader }: { open, setOpen, itemInfo: ItemProps, menuPos, authHeader }) {
    // const item = filebrowserState.dirMap.get(itemId)
    if (!itemInfo.isDir) {
        return null
    }

    return (
        <Menu opened={open} onClose={() => setOpen(false)} closeOnClickOutside>
            <Menu.Target>
                <Box style={{ position: 'fixed', top: menuPos.y - 140, left: menuPos.x - 230 }} />
            </Menu.Target>

            <Menu.Dropdown>
                <Menu.Label>{itemInfo.itemTitle}</Menu.Label>

                <Menu.Item disabled={itemInfo.extras.shares.length > 0} leftSection={<IconSpiral />} onClick={(e) => { e.stopPropagation(); NewWormhole(itemInfo.itemId, authHeader) }}>
                    Attach Wormhole
                </Menu.Item>

                <Menu.Item leftSection={<IconTrash />} color='red' onClick={(e) => { e.stopPropagation(); DeleteFiles([itemInfo.itemId], authHeader) }}>
                    Delete
                </Menu.Item>

            </Menu.Dropdown>
        </Menu>
    )
}

const FolderIcon = ({ shares }: { shares }) => {
    const [copied, setCopied] = useState(false)
    return (
        <RowBox style={{ justifyContent: 'center' }}>
            <IconFolder size={'75%'} />
            {shares.length !== 0 && (
                <Tooltip label={copied ? 'Copied' : 'Copy Wormhole'}>
                    <IconSpiral
                        color={copied ? '#4444ff' : 'white'}
                        style={{ position: 'absolute', right: 0, top: 0 }}
                        onClick={e => { e.stopPropagation(); navigator.clipboard.writeText(`${window.location.origin}/wormhole/${shares[0].ShareId}`); setCopied(true); setTimeout(() => setCopied(false), 1000) }}
                        onDoubleClick={e => e.stopPropagation()}
                    />
                </Tooltip>
            )}
        </RowBox>
    )
}

const IconDisplay = ({ itemInfo }: { itemInfo: ItemProps }) => {
    if (itemInfo.isDir) {
        return (<FolderIcon shares={itemInfo.extras.shares} />)
    }
    if (!itemInfo.imported) {
        return (
            <Center style={{ height: "100%", width: "100%" }}>
                <Skeleton height={"100%"} width={"100%"} />
                <Text pos={'absolute'} style={{ userSelect: 'none' }}>Processing...</Text>
            </Center>
        )
    }
    const ext = itemInfo.itemTitle.slice(itemInfo.itemTitle.indexOf('.'), itemInfo.itemTitle.length)

    switch (ext) {
        case "zip": return (<IconFileZip size={'75%'} />)
        default: return (<IconFile size={'75%'} />)
    }
}

function Files({ filebrowserState, folderId, notFound, setNotFound, alreadyScanned, setAlreadyScanned, dispatch, wsSend, uploadDispatch, authHeader }:
    { filebrowserState: FileBrowserStateType, folderId, notFound, setNotFound, alreadyScanned, setAlreadyScanned, dispatch, wsSend: (action: string, content: any) => void, uploadDispatch, authHeader }) {
    const { userInfo } = useContext(userContext)
    const nav = useNavigate()
    const [debouncedSearch] = useDebouncedValue(filebrowserState.searchContent, 200)
    const [scanNeeded, setScanNeeded] = useState(false)

    const itemsList: ItemProps[] = useMemo(() => {
        if (!userInfo) {
            return []
        }
        let filesList = MapToList(filebrowserState.dirMap).filter((val) => { return val.filename.toLowerCase().includes(debouncedSearch.toLowerCase()) && val.id !== userInfo.trashFolderId })
        const itemsList: ItemProps[] = filesList.map(v => {
            const selected = Boolean(filebrowserState.selected.get(v.id))
            const [size, uints] = humanFileSize(v.size)
            const item: ItemProps = {
                itemId: v.id,
                itemTitle: v.filename,
                secondaryInfo: String(size + " " + uints),
                selected: selected,
                mediaData: v.mediaData,
                droppable: v.isDir && !selected,
                isDir: v.isDir,
                imported: v.imported,

                extras: { shares: v.shares }
            }
            if (!scanNeeded && !v.imported && v.displayable) {
                setScanNeeded(true)
            }
            return item
        })
        return itemsList
    }, [JSON.stringify(Array.from(filebrowserState.dirMap.values())), filebrowserState.selected.size, debouncedSearch, userInfo])

    useEffect(() => {
        if (scanNeeded && !alreadyScanned) {
            setAlreadyScanned(true)
            dispatchSync(folderId, wsSend, false, false)
        }
    }, [scanNeeded, alreadyScanned, folderId, wsSend, setAlreadyScanned])

    const itemsCtx: GlobalContextType = useMemo(() => {
        let scrollToIndex: number
        if (filebrowserState.scrollTo) {
            scrollToIndex = itemsList.findIndex((v) => v.itemId === filebrowserState.scrollTo)
        }

        const context: GlobalContextType = {
            setSelected: (itemId: string, selected?: boolean) => dispatch({ type: 'set_selected', fileId: itemId, selected: selected }),
            visitItem: (itemId: string) => { if (filebrowserState.dirMap.get(itemId).isDir) { nav(itemId) } },
            setDragging: (d: boolean) => dispatch({ type: 'set_dragging', dragging: d }),
            moveSelected: (entryId: string) => { if (filebrowserState.dirMap.get(entryId).isDir) { moveSelected(filebrowserState.selected, entryId, authHeader).then(() => dispatch({ type: 'clear_selected' })) } },
            blockFocus: (b: boolean) => dispatch({ type: 'set_block_focus', block: b }),
            rename: (itemId: string, newName: string) => HandleRename(itemId, newName, folderId, filebrowserState.selected.size, dispatch, authHeader),

            menu: ({ open, setOpen, itemInfo, menuPos }) => FileContextMenu({ open, setOpen, itemInfo, menuPos, authHeader }),
            iconDisplay: IconDisplay,

            dragging: filebrowserState.draggingState,
            initialScrollIndex: scrollToIndex,
        }

        return context
    }, [itemsList, filebrowserState.selected.size, filebrowserState.draggingState, dispatch, folderId])

    if (notFound) {
        return (
            <NotFound resourceType='Folder' link='/files/home' setNotFound={setNotFound} />
        )
    }

    if (itemsList.length !== 0) {
        return (
            <ItemScroller itemsContext={itemsList} globalContext={itemsCtx} />
        )
    } else if (!filebrowserState.loading && folderId !== "shared") {
        return (
            <RowBox style={{ overflow: 'hidden', justifyContent: 'center' }}>
                <Card variant="solid" style={{ height: 'max-content', top: '40vh', position: 'fixed', padding: '50px' }}>
                    <Text size='25px' ta={'center'} fw={800}>
                        This folder is empty
                    </Text>

                    {filebrowserState.folderInfo.modifiable && (
                        <Card.Section style={{ display: 'flex', flexDirection: 'row', justifyContent: 'center', paddingTop: 15 }}>
                            <FileButton onChange={(files) => { HandleUploadButton(files, filebrowserState.folderInfo.id, false, "", authHeader, uploadDispatch, wsSend) }} accept="file" multiple>
                                {(props) => {
                                    return (
                                        <ColumnBox onClick={() => { props.onClick() }} style={{ cursor: 'pointer', marginTop: '0px' }}>
                                            <IconUpload size={100} style={{ padding: "10px" }} />
                                            <Text size='20px' fw={600}>
                                                Upload
                                            </Text>
                                            <Space h={4}></Space>
                                            <Text size='12px'>Click or Drop</Text>
                                        </ColumnBox>
                                    )
                                }}
                            </FileButton>
                            <Divider orientation='vertical' m={20} />

                            <ColumnBox onClick={(e) => { e.stopPropagation(); dispatch({ type: 'new_dir' }) }} style={{ cursor: 'pointer' }}>
                                <IconFolderPlus size={100} style={{ padding: "10px" }} />
                                <Text size='20px' fw={600}>
                                    New Folder
                                </Text>

                            </ColumnBox>
                        </Card.Section>
                    )}
                </Card>
            </RowBox>
        )
    } else if (!filebrowserState.loading && filebrowserState.folderInfo.id === "shared") {
        return (
            <ColumnBox>
                <Paper variant="solid" style={{ height: 'max-content', top: '40vh', position: 'fixed', padding: 40 }}>
                    <ColumnBox style={{ alignItems: 'center', justifyContent: 'center' }}>
                        <Text display={'flex'} style={{ flexDirection: 'column', alignItems: 'center', padding: 2 }}>
                            You have no files shared with you
                        </Text>
                        <Space h={'lg'} />
                        <Button color={'#4444ff'} fullWidth onClick={() => nav('/files/home')}>Return Home</Button>
                    </ColumnBox>
                </Paper>
            </ColumnBox>
        )
    } else {
        return null
    }
}

function useQuery() {
    const { search } = useLocation();

    return useMemo(() => new URLSearchParams(search), [search]);
}

const FileBrowser = () => {
    const folderId = useParams()["*"]
    const query = useQuery()
    const navigate = useNavigate()
    const { authHeader, userInfo } = useContext(userContext)
    const searchRef = useRef()
    const { wsSend, lastMessage, readyState } = useWeblensSocket()
    const [alreadyScanned, setAlreadyScanned] = useState(false)
    const [notFound, setNotFound] = useState(false)
    const { uploadState, uploadDispatch } = useUploadStatus()

    const [filebrowserState, dispatch]: [FileBrowserStateType, (action: FileBrowserAction) => void] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, fileData>(),
        selected: new Map<string, boolean>(),
        uploadMap: new Map<string, boolean>(),
        pasteImg: null,
        folderInfo: getBlankFile(),
        parents: [],
        draggingState: 0,
        scanProgress: 0,
        homeDirSize: 0,
        trashDirSize: 0,
        waitingForNewName: "",
        presentingId: "",
        searchContent: "",
        lastSelected: "",
        scrollTo: "",
        loading: true,
        holdingShift: false,
        blockFocus: false,
    })

    useKeyDown(filebrowserState, userInfo, dispatch, authHeader, wsSend, searchRef)

    const realId = useMemo(() => {
        let realId
        if (userInfo) {
            if (folderId === "home") {
                realId = userInfo.homeFolderId
            } else if (folderId === "trash") {
                realId = userInfo.trashFolderId
            } else {
                realId = folderId
            }
        }
        return realId
    }, [folderId, userInfo])

    // Hook to handle uploading images from the clipboard
    usePaste(realId, userInfo, searchRef, dispatch)

    // Subscribe to the current folder to get updates about size, children, etc.
    useEffect(() => {
        if (readyState === 1 && realId != null && realId !== "shared") {
            if (filebrowserState.folderInfo.id !== realId) {
                return
            }
            if (realId === userInfo.homeFolderId) {
                wsSend("subscribe", { subscribeType: "folder", subscribeKey: userInfo.trashFolderId, subscribeMeta: JSON.stringify({ recursive: false }) })
                return (
                    () => {
                        wsSend("unsubscribe", { subscribeKey: userInfo.trashFolderId })
                    }
                )
            }
            wsSend("subscribe", { subscribeType: "folder", subscribeKey: realId, subscribeMeta: JSON.stringify({ recursive: false }) })
            return (
                () => wsSend("unsubscribe", { subscribeKey: realId })
            )
        }
    }, [readyState, filebrowserState.folderInfo.id, realId, userInfo?.trashFolderId, wsSend])

    // Subscribe to the home folder if we aren't in it, to be able to update the total disk usage
    useEffect(() => {
        if (!userInfo || readyState !== 1) {
            return
        }

        wsSend("subscribe", { subscribeType: "folder", subscribeKey: userInfo.homeFolderId, subscribeMeta: JSON.stringify({ recursive: false }) })
        return (
            () => wsSend("unsubscribe", { subscribeKey: userInfo.homeFolderId })
        )
    }, [userInfo, readyState])

    // Listen for incoming websocket messages
    useEffect(() => {
        if (!userInfo) {
            return
        }
        HandleWebsocketMessage(lastMessage, userInfo, dispatch, authHeader)
    }, [lastMessage, userInfo, authHeader])

    // Reset most of the state when we change folders
    useEffect(() => {
        if (!folderId || folderId === userInfo?.homeFolderId || folderId === "undefined") {
            navigate('/files/home')
        }
        if (!userInfo || authHeader.Authorization === '' || !realId) {
            return
        }
        // Kinda just reset everything...
        setAlreadyScanned(false)
        dispatch({ type: "clear_files" })
        dispatch({ type: "set_search", search: "" })
        dispatch({ type: "set_scan_progress", progress: 0 })
        dispatch({ type: "set_loading", loading: true })

        GetFolderData(realId, userInfo, dispatch, authHeader)
            .then(() => {
                const jumpItem = query.get("jumpTo")
                if (jumpItem) {
                    dispatch({ type: 'set_scroll_to', fileId: jumpItem })
                    dispatch({ type: 'set_selected', fileId: jumpItem, selected: true })
                }
            })
            .catch((r) => {
                if (r === 404) {
                    setNotFound(true)
                    return
                }
                notifications.show({ title: "Could not get folder info", message: String(r), color: 'red', autoClose: 5000 })
            })
            .finally(() => dispatch({ type: "set_loading", loading: false }))
    }, [folderId, userInfo, authHeader, navigate, realId])

    const moveSelectedTo = useCallback(folderId => { moveSelected(filebrowserState.selected, folderId, authHeader); dispatch({ type: 'clear_selected' }) }, [filebrowserState.selected.size, folderId, authHeader])
    const doScan = useCallback(() => { dispatch({ type: 'set_loading', loading: true }); dispatchSync(realId, wsSend, filebrowserState.holdingShift, filebrowserState.holdingShift) }, [realId, wsSend, filebrowserState.holdingShift])

    if (!userInfo) {
        return null
    }

    return (
        <ColumnBox style={{ height: '100vh', backgroundColor: "#111418" }} >
            <HeaderBar
                searchContent={filebrowserState.searchContent}
                dispatch={dispatch}
                page={"files"}
                searchRef={searchRef}
                loading={filebrowserState.loading}
                progress={filebrowserState.scanProgress}
            />
            <DraggingCounter dragging={filebrowserState.draggingState} numSelected={filebrowserState.selected.size} dispatch={dispatch} />
            <Presentation mediaData={filebrowserState.dirMap.get(filebrowserState.presentingId)?.mediaData} dispatch={dispatch} />
            <UploadStatus uploadState={uploadState} uploadDispatch={uploadDispatch} />
            <PasteImageDialogue img={filebrowserState.pasteImg} folderId={realId} dirMap={filebrowserState.dirMap} authHeader={authHeader} dispatch={dispatch} wsSend={wsSend} />
            <RowBox style={{ alignItems: 'flex-start', paddingTop: 15 }}>
                <GlobalActions fbState={filebrowserState} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} />
                <DirViewWrapper
                    folderId={realId}
                    folderName={filebrowserState.folderInfo?.filename}
                    dragging={filebrowserState.draggingState}
                    onDrop={(e => HandleDrop(e.dataTransfer.items, realId, Array.from(filebrowserState.dirMap.values()).map((value: fileData) => value.filename), false, "", authHeader, uploadDispatch, wsSend))}
                    dispatch={dispatch}
                >
                    <RowBox style={{ height: 46 }}>
                        <ScanFolderButton folderId={realId} holdingShift={filebrowserState.holdingShift} doScan={doScan} />
                        <Space w={6} />
                        <Crumbs finalFile={filebrowserState.folderInfo} parents={filebrowserState.parents} navOnLast={false} dragging={filebrowserState.draggingState} moveSelectedTo={moveSelectedTo} />
                    </RowBox>

                    <Space h={10} />

                    <Files filebrowserState={filebrowserState} folderId={realId} notFound={notFound} setNotFound={setNotFound} alreadyScanned={alreadyScanned} setAlreadyScanned={setAlreadyScanned} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} authHeader={authHeader} />
                </DirViewWrapper>
            </RowBox>
        </ColumnBox>
    )
}

export default FileBrowser