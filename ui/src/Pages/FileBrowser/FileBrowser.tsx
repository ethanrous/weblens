// React
import { useState, useEffect, useReducer, useMemo, useRef, useContext, useCallback } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

// Icons
import { IconChevronRight, IconCloud, IconDownload, IconFile, IconFileZip, IconFolder, IconFolderPlus, IconHome, IconLogin, IconPhotoPlus, IconShare, IconSpiral, IconTrash, IconUpload, IconUsers } from "@tabler/icons-react"

// Mantine
import { Box, Button, Text, Space, FileButton, Divider, Popover, Progress, Menu, Center, Skeleton } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useDebouncedValue } from '@mantine/hooks'

// Weblens
import { HandleDrop, HandleWebsocketMessage, downloadSelected, fileBrowserReducer, useKeyDown, useMousePosition, moveSelected, HandleUploadButton, usePaste, UploadViaUrl, MapToList, HandleRename, selectedMediaIds, selectedFolderIds, useSubscribe, SetFileData, getRealId } from './FileBrowserLogic'
import { fileData, FileBrowserStateType, MediaData, FileBrowserAction, getBlankFile, FileBrowserDispatch } from '../../types/Types'
import { DirViewWrapper, ColumnBox, RowBox, TransferCard, PresentationFile, FolderIcon, GetStartedCard, IconDisplay, FileInfoDisplay } from './FilebrowserStyles'
import { DeleteFiles, DeleteWormhole, GetFileShare, GetFolderData, NewWormhole, ShareFiles } from '../../api/FileBrowserApi'
import Presentation, { PresentationContainer } from '../../components/Presentation'
import UploadStatus, { useUploadStatus } from '../../components/UploadStatus'
import { GlobalContextType, ItemProps } from '../../components/ItemDisplay'
import useWeblensSocket, { dispatchSync } from '../../api/Websocket'
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import { ItemScroller } from '../../components/ItemScroller'
import { MediaImage } from '../../components/PhotoContainer'
import { ShareInput } from '../../components/Share'
import HeaderBar from "../../components/HeaderBar"
import { AlbumScoller } from './FileBrowserAlbums'
import NotFound from '../../components/NotFound'
import '../../components/filebrowserStyle.css'
import { userContext } from '../../Context'
import { humanFileSize } from '../../util'
import '../../components/style.css'
import { ShareBox } from './FilebrowserShareMenu'

function PasteImageDialogue({ img, dirMap, folderId, authHeader, dispatch, wsSend }: { img: ArrayBuffer, dirMap: Map<string, fileData>, folderId, authHeader, dispatch, wsSend }) {
    if (!img) {
        return null
    }
    const media: MediaData = {} as MediaData
    media.fileHash = "paste"
    media.thumbnail = img

    return (
        <PresentationContainer shadeOpacity={"0.25"} onClick={() => dispatch({ type: "paste_image", img: null })} >
            <ColumnBox style={{ position: 'absolute', justifyContent: 'center', alignItems: 'center', zIndex: 2 }}>
                <Text fw={700} size='40px' style={{ paddingBottom: '50px' }}>Upload from clipboard?</Text>
                <ColumnBox onClick={(e) => { e.stopPropagation() }} style={{ height: "50%", width: "max-content", backgroundColor: '#222277ee', padding: '10px', borderRadius: '8px', overflow: 'hidden' }}>
                    <MediaImage media={media} quality='thumbnail' />
                </ColumnBox>
                <RowBox style={{ justifyContent: 'space-between', width: '300px', height: '150px' }}>
                    <Button size='xl' variant='default' onClick={(e) => { e.stopPropagation(); dispatch({ type: "paste_image", img: null }) }}>Cancel</Button>
                    <Button size='xl' color='#4444ff' onClick={e => { e.stopPropagation(); UploadViaUrl(img, folderId, dirMap, authHeader, dispatch, wsSend) }}>Upload</Button>
                </RowBox>
            </ColumnBox>
        </PresentationContainer>
    )
}

function GlobalActions({ fbState, dispatch, wsSend, uploadDispatch }: { fbState: FileBrowserStateType, dispatch: FileBrowserDispatch, wsSend: (action: string, content: any) => void, uploadDispatch }) {
    const nav = useNavigate()
    const { userInfo, authHeader } = useContext(userContext)
    const amHome = fbState.folderInfo.id === userInfo?.homeFolderId

    const BUTTON_WIDTH = 190

    return (
        <Box style={{ position: 'relative', paddingLeft: 15, paddingRight: 15, paddingTop: 60, width: 230, }}>
            {userInfo.username === "" && (
                <ColumnBox style={{ position: 'absolute', backdropFilter: 'blur(3px)', WebkitBackdropFilter: 'blur(4px)', borderRadius: 10, margin: -10, zIndex: 1, justifyContent: 'center' }}>
                    <RowBox onClick={() => nav("/login")} style={{ backgroundColor: '#4444ff', boxShadow: '5px 5px 10px 2px black', borderRadius: 4, padding: 10, height: 'max-content', width: 'max-content', cursor: 'pointer' }}>
                        <IconLogin />
                        <Text c='white' style={{ paddingLeft: 10 }}>Login</Text>
                    </RowBox>
                </ColumnBox>
            )}

            <ColumnBox >
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


                <Divider w={'100%'} my='lg' size={1.5} />

                <UsageInfo inHome={fbState.folderInfo.id === userInfo.homeFolderId} homeDirSize={fbState.homeDirSize} currentFolderSize={fbState.folderInfo.size} displayCurrent={fbState.folderInfo.id !== "shared"} trashSize={fbState.trashDirSize} />
            </ColumnBox>
        </Box>
    )
}

function UsageInfo({ inHome, homeDirSize, currentFolderSize, displayCurrent, trashSize }) {
    if (!displayCurrent) {
        currentFolderSize = homeDirSize
    }
    if (homeDirSize === 0) {
        homeDirSize = currentFolderSize
    }
    if (!trashSize) {
        trashSize = 0
    }
    if (inHome) {
        currentFolderSize = currentFolderSize - trashSize
    }

    return (
        <ColumnBox style={{ height: 'max-content', width: '100%', alignItems: 'flex-start' }}>
            <Text fw={600} style={{ userSelect: 'none' }}>Usage</Text>
            <Space h={10} />
            <Progress radius={'xs'} color='#4444ff' value={(currentFolderSize / homeDirSize) * 100} style={{ height: 20, width: '100%', marginBottom: 10 }} />
            <RowBox>
                {displayCurrent && (
                    <RowBox>
                        <IconFolder size={20} />
                        <Text style={{ userSelect: 'none' }} size='14px' pl={3}>{humanFileSize(currentFolderSize)}</Text>
                    </RowBox>
                )}
                <RowBox style={{ justifyContent: 'right' }}>
                    <Text style={{ userSelect: 'none' }} size='14px' pr={3}>{humanFileSize(homeDirSize)}</Text>
                    <IconCloud size={20} />
                </RowBox>
            </RowBox>
        </ColumnBox>
    )
}

function DraggingCounter({ dragging, dirMap, selected, dispatch }) {
    const position = useMousePosition()
    const selectedKeys = Array.from(selected.keys())
    const { files, folders } = useMemo(() => {
        let files = 0
        let folders = 0

        selectedKeys.forEach((k: string) => {
            if (dirMap.get(k)?.isDir) {
                folders++
            } else {
                files++
            }
        })
        return { files, folders }
    }, [JSON.stringify(selectedKeys)])

    if (dragging !== 1) {
        return null
    }

    return (
        <Box
            style={{
                display: 'flex',
                flexDirection: 'column',
                position: 'fixed',
                top: position.y + 8,
                left: position.x + 8,
                zIndex: 10
            }}
            onMouseUp={() => { dispatch({ type: 'set_dragging', dragging: false }) }}
        >
            {Boolean(files) && (
                <RowBox style={{ height: 'max-content' }}>
                    <IconFile size={30} />
                    <Space w={10} />
                    <StyledBreadcrumb label={files.toString()} fontSize={30} compact />
                </RowBox>
            )}
            {Boolean(folders) && (
                <RowBox style={{ height: 'max-content' }}>
                    <IconFolder size={30} />
                    <Space w={10} />
                    <StyledBreadcrumb label={folders.toString()} fontSize={30} compact />
                </RowBox>
            )}
        </Box>
    )
}

function FileContextMenu({ itemId, fbState, open, setOpen, menuPos, dispatch, wsSend, authHeader }: { itemId: string, fbState: FileBrowserStateType, open, setOpen, menuPos, dispatch, wsSend, authHeader }) {
    const [shareMenu, setShareMenu] = useState(false)
    const [addToAlbumMenu, setAddToAlbumMenu] = useState(false)
    const itemInfo: fileData = fbState.dirMap.get(itemId) || {} as fileData
    const selected: boolean = Boolean(fbState.selected.get(itemId))

    const { items, mediaCount } = useMemo(() => {
        if (fbState.dirMap.size === 0) {
            return { items: [], anyDisplayable: false }
        }
        const itemIds = selected ? Array.from(fbState.selected.keys()) : [itemId]
        let mediaCount = 0
        const items = itemIds.map(i => {
            const item = fbState.dirMap.get(i)
            if (!item) {
                return null
            }
            if (item.displayable || item.isDir) {
                mediaCount++
            }
            return item
        })

        return { items: items.filter(i => Boolean(i)), mediaCount }
    }, [itemId, selected, fbState.selected.size])

    let extraString
    if (selected && fbState.selected.size > 1) {
        extraString = ` +${fbState.selected.size - 1} more`
    }

    return (
        <Menu opened={open || shareMenu} onClose={() => setOpen(false)} closeOnClickOutside={!addToAlbumMenu} position='right-start' closeOnItemClick={false} styles={{ dropdown: { boxShadow: '0px 0px 20px -5px black', width: 'max-content', padding: 10, border: 0 } }}>
            <Menu.Target>
                <Box style={{ position: 'absolute', top: menuPos.y, left: menuPos.x }} />
            </Menu.Target>

            <Menu.Dropdown onClick={e => e.stopPropagation()} onDoubleClick={e => e.stopPropagation()}>
                <Menu.Label>
                    <RowBox style={{ gap: 10 }}>
                        <Text truncate='end'>
                            {itemInfo.filename}
                        </Text>
                        {extraString}
                    </RowBox>
                </Menu.Label>

                <Menu opened={addToAlbumMenu} trigger='hover' offset={0} position="right-start" onOpen={() => { setAddToAlbumMenu(true); dispatch({ type: 'set_block_focus', block: true }) }} onClose={() => { setAddToAlbumMenu(false); dispatch({ type: 'set_block_focus', block: false }) }} styles={{ dropdown: { boxShadow: '0px 0px 20px -5px black', width: 'max-content', padding: 10, border: 0 } }}>
                    <Menu.Target>
                        <Menu.Item leftSection={<IconPhotoPlus />} rightSection={<IconChevronRight />} disabled={false}>
                            <Text>Add to Album</Text>
                        </Menu.Item>
                    </Menu.Target>
                    <Menu.Dropdown onMouseOver={e => e.stopPropagation()}>
                        <AlbumScoller candidates={{ media: items.filter(i => i.displayable).map(i => i.id), folders: items.filter(i => i.isDir).map(i => i.id) }} authHeader={authHeader} />
                    </Menu.Dropdown>
                </Menu>

                {/* Wormhole menu */}
                {itemInfo.isDir && (
                    <Menu.Item styles={{ itemLabel: { flex: 0, width: '90%' } }} disabled={(fbState.selected.size > 1 && selected)} leftSection={<IconSpiral />} onClick={(e) => { e.stopPropagation(); if (itemInfo.shares?.length === 0) { NewWormhole(itemId, authHeader) } else { navigator.clipboard.writeText(`${window.location.origin}/wormhole/${itemInfo.shares[0].ShareId}`); setOpen(false); notifications.show({ message: 'Link to wormhole copied', color: 'green' }) } }}>
                        <Text truncate='end'>{itemInfo.shares?.length === 0 ? "Attach" : "Copy"} Wormhole</Text>
                    </Menu.Item>
                )}

                {/* Share menu */}
                <Menu opened={shareMenu} trigger='hover' offset={0} position="right-start" onOpen={() => { setShareMenu(true); dispatch({ type: 'set_block_focus', block: true }) }} onClose={() => { setShareMenu(false); dispatch({ type: 'set_block_focus', block: false }) }} styles={{ dropdown: { boxShadow: '0px 0px 20px -5px black', width: 'max-content', padding: 10, border: 0 } }}>
                    <Menu.Target>
                        <Menu.Item styles={{ itemLabel: { flex: 0, width: '90%' } }} leftSection={<IconShare />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'set_block_focus', block: true }); setShareMenu(true) }}>
                            <Text>Share</Text>
                        </Menu.Item>
                    </Menu.Target>
                    <Menu.Dropdown onMouseOver={e => e.stopPropagation()}>
                        <ShareBox candidates={items.map(i => i.id)} authHeader={authHeader} />
                    </Menu.Dropdown>
                </Menu>

                <Menu.Item styles={{ itemLabel: { flex: 0, width: '90%' } }} leftSection={<IconDownload />} onClick={(e) => { e.stopPropagation(); downloadSelected(selected ? Array.from(fbState.selected.keys()).map(fId => fbState.dirMap.get(fId)) : [fbState.dirMap.get(itemId)], dispatch, wsSend, authHeader) }} >
                    <Text>Download</Text>
                </Menu.Item>

                <Divider w={'100%'} my='sm' />

                {itemInfo.shares && itemInfo.shares.length !== 0 && (
                    <Menu.Item styles={{ itemLabel: { flex: 0, width: '90%' } }} color={'red'} leftSection={<IconSpiral />} onClick={(e) => { e.stopPropagation(); DeleteWormhole(itemInfo.shares[0].ShareId, authHeader) }}>
                        <Text truncate='end'>Remove Wormhole</Text>
                    </Menu.Item>
                )}

                <Menu.Item styles={{ itemLabel: { flex: 0, width: '90%' } }} leftSection={<IconTrash />} color='red' onClick={(e) => { e.stopPropagation(); DeleteFiles(items.map(i => i.id), authHeader); setOpen(false) }}>
                    <Text>Delete</Text>
                </Menu.Item>

            </Menu.Dropdown>
        </Menu>
    )
}

function SingleFile({ file, doDownload }: { file: fileData, doDownload: (file: fileData) => void }) {
    if (file.id === "") {
        return null
    }

    return (
        <RowBox style={{ height: "90vh" }}>
            <Box style={{ display: 'flex', width: "55%", height: "100%", padding: 30 }}>
                <IconDisplay file={file} quality='fullres' />
            </Box>
            <ColumnBox style={{ width: 'max-content', padding: 30 }}>
                <FileInfoDisplay file={file} />
                <Box style={{ minHeight: '40%' }}>
                    <RowBox onClick={() => doDownload(file)} style={{ backgroundColor: '#4444ff', borderRadius: 4, padding: 10, height: 'max-content', cursor: 'pointer' }}>
                        <IconDownload />
                        <Text c='white' style={{ paddingLeft: 10 }}>Download {file.filename}</Text>
                    </RowBox>

                </Box>
            </ColumnBox>
        </RowBox>
    )
}

function Files({ filebrowserState, folderId, notFound, setNotFound, alreadyScanned, setAlreadyScanned, dispatch, wsSend, uploadDispatch, authHeader }:
    { filebrowserState: FileBrowserStateType, folderId, notFound, setNotFound, alreadyScanned, setAlreadyScanned, dispatch: (action: FileBrowserAction) => void, wsSend: (action: string, content: any) => void, uploadDispatch, authHeader }) {
    const { userInfo } = useContext(userContext)
    const nav = useNavigate()
    const [debouncedSearch] = useDebouncedValue(filebrowserState.searchContent, 200)
    const [scanNeeded, setScanNeeded] = useState(false)
    const boundRef = useRef(null)
    const moveSelectedTo = useCallback(folderId => { moveSelected(filebrowserState.selected, folderId, authHeader); dispatch({ type: 'clear_selected' }) }, [filebrowserState.selected.size, folderId, authHeader])

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
                selected: (selected ? 0x1 : 0x0) | (filebrowserState.lastSelected === v.id ? 0x10 : 0x0),
                mediaData: v.mediaData,
                droppable: v.isDir && !selected,
                isDir: v.isDir,
                imported: v.imported,
                displayable: v.displayable,
                shares: v.shares
            }
            if (!scanNeeded && !v.imported && v.displayable) {
                setScanNeeded(true)
            }
            return item
        })
        return itemsList
    }, [JSON.stringify(Array.from(filebrowserState.dirMap.values())), JSON.stringify(Array.from(filebrowserState.selected.keys())), debouncedSearch, userInfo, filebrowserState.lastSelected])

    useEffect(() => {
        dispatch({ type: "set_loading", loading: true })
    }, [debouncedSearch])

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
            visitItem: (itemId: string) => { const item = filebrowserState.dirMap.get(itemId); if (item.isDir) { nav(itemId) } else if (item.displayable) { dispatch({ type: 'set_presentation', presentingId: itemId }) } },
            setDragging: (d: boolean) => dispatch({ type: 'set_dragging', dragging: d }),
            moveSelected: (entryId: string) => { if (filebrowserState.dirMap.get(entryId).isDir) { moveSelected(filebrowserState.selected, entryId, authHeader).then(() => dispatch({ type: 'clear_selected' })) } },
            blockFocus: (b: boolean) => dispatch({ type: 'set_block_focus', block: b }),
            rename: (itemId: string, newName: string) => HandleRename(itemId, newName, folderId, filebrowserState.selected.size, dispatch, authHeader),

            setMenuOpen: (o: boolean) => dispatch({ type: 'set_menu_open', open: o }),
            setMenuPos: (pos: { x: number, y: number }) => dispatch({ type: 'set_menu_pos', pos: pos }),
            setMenuTarget: (target: string) => dispatch({ type: 'set_menu_target', fileId: target }),

            iconDisplay: ({ itemInfo }) => { const file = filebrowserState.dirMap.get(itemInfo.itemId); return (<IconDisplay file={file} />) },
            setMoveDest: (itemName) => dispatch({ type: 'set_move_dest', fileName: itemName }),

            dragging: filebrowserState.draggingState,
            initialScrollIndex: scrollToIndex,
        }

        return context
    }, [itemsList, filebrowserState.selected.size, filebrowserState.draggingState, dispatch, folderId])

    useEffect(() => {
        dispatch({ type: 'set_files_list', fileIds: itemsList.map(v => v.itemId) })
        dispatch({ type: 'set_loading', loading: false })
    }, [itemsList, dispatch])

    if (notFound) {
        return (
            <NotFound resourceType='Folder' link='/files/home' setNotFound={setNotFound} />
        )
    }

    return (
        <Box ref={boundRef} style={{ width: '100%', height: '100%' }}>
            <TransferCard action='Move' destination={filebrowserState.moveDest} boundRef={boundRef} />
            <Crumbs finalFile={filebrowserState.folderInfo} parents={filebrowserState.parents} navOnLast={false} dragging={filebrowserState.draggingState} moveSelectedTo={moveSelectedTo} setMoveDest={(itemName) => dispatch({ type: "set_move_dest", fileName: itemName })} />

            {
                (itemsList.length !== 0 && (
                    <ItemScroller itemsContext={itemsList} globalContext={itemsCtx} dispatch={dispatch} />
                ))
                || ((!filebrowserState.loading && folderId !== "shared" && filebrowserState.searchContent === "" && filebrowserState.searchContent === debouncedSearch) && (
                    <GetStartedCard filebrowserState={filebrowserState} moveSelectedTo={moveSelectedTo} dispatch={dispatch} uploadDispatch={uploadDispatch} authHeader={authHeader} wsSend={wsSend} />
                ))
                || ((!filebrowserState.loading && filebrowserState.folderInfo.id === "shared") && (
                    <ColumnBox>
                        <ColumnBox style={{ alignItems: 'center', marginTop: '20vh' }}>
                            <Text size='28px'>
                                No files are shared with you
                            </Text>
                        </ColumnBox>
                    </ColumnBox>
                ))
                || (!filebrowserState.loading && filebrowserState.searchContent !== "" && (
                    <ColumnBox style={{ justifyContent: 'flex-end', height: '20%' }}>
                        <Text size='20px'>No items match your search</Text>
                    </ColumnBox>
                ))
            }
        </Box>
    )
}

function DirView({ filebrowserState, folderId, notFound, setNotFound, alreadyScanned, setAlreadyScanned, dispatch, wsSend, uploadDispatch, authHeader }:
    { filebrowserState: FileBrowserStateType, folderId, notFound, setNotFound, alreadyScanned, setAlreadyScanned, dispatch: (action: FileBrowserAction) => void, wsSend: (action: string, content: any) => void, uploadDispatch, authHeader }) {

    const download = useCallback((file: fileData) => downloadSelected([file], dispatch, wsSend, authHeader, filebrowserState.isShare ? filebrowserState.realId : undefined), [authHeader, wsSend, dispatch, filebrowserState.isShare, filebrowserState.realId])

    if (filebrowserState.folderInfo.isDir) {
        return (
            <Files filebrowserState={filebrowserState} folderId={folderId} notFound={notFound} setNotFound={setNotFound} alreadyScanned={alreadyScanned} setAlreadyScanned={setAlreadyScanned} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} authHeader={authHeader} />
        )
    } else {
        return (
            <SingleFile file={filebrowserState.folderInfo} doDownload={download} />
        )
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

    const [alreadyScanned, setAlreadyScanned] = useState(false)
    const [notFound, setNotFound] = useState(false)
    const { uploadState, uploadDispatch } = useUploadStatus()

    const [filebrowserState, dispatch]: [FileBrowserStateType, (action: FileBrowserAction) => void] = useReducer(fileBrowserReducer, {
        dirMap: new Map<string, fileData>(),
        selected: new Map<string, boolean>(),
        uploadMap: new Map<string, boolean>(),
        menuPos: { x: 0, y: 0 },
        pasteImg: null,
        folderInfo: getBlankFile(),
        parents: [],
        filesList: [],
        draggingState: 0,
        scanProgress: 0,
        homeDirSize: 0,
        trashDirSize: 0,
        waitingForNewName: "",
        menuTargetId: "",
        presentingId: "",
        searchContent: "",
        lastSelected: "",
        scrollTo: "",
        moveDest: "",
        loading: true,
        holdingShift: false,
        blockFocus: false,
        menuOpen: false,
        numCols: 0,
        isShare: false,
        realId: "",
    })

    useEffect(() => {
        dispatch({ type: 'set_is_share', isShare: window.location.pathname.startsWith("/share") })
        dispatch({ type: 'set_real_id', realId: getRealId(folderId, userInfo, authHeader) })
    }, [folderId, dispatch, authHeader, userInfo])

    const wsSend = useSubscribe(filebrowserState.realId, filebrowserState.folderInfo.id, userInfo, dispatch, authHeader)
    useKeyDown(filebrowserState, userInfo, dispatch, authHeader, wsSend, searchRef)

    // Hook to handle uploading images from the clipboard
    usePaste(filebrowserState.realId, userInfo, searchRef, dispatch)

    // Reset most of the state when we change folders
    useEffect(() => {
        const syncState = async () => {
            if (!folderId || folderId === userInfo?.homeFolderId || folderId === "undefined") {
                navigate('/files/home', { replace: true })
            }

            if ((!filebrowserState.isShare && (!userInfo.username || authHeader.Authorization === '')) || !filebrowserState.realId) {
                return
            }
            // Kinda just reset everything...
            setNotFound(false)
            setAlreadyScanned(false)
            dispatch({ type: "clear_files" })
            dispatch({ type: "set_search", search: "" })
            dispatch({ type: "set_scan_progress", progress: 0 })
            dispatch({ type: "set_loading", loading: true })

            let fileData
            if (filebrowserState.isShare) {
                fileData = await GetFileShare(filebrowserState.realId, authHeader)
            } else {
                fileData = await GetFolderData(filebrowserState.realId, userInfo, dispatch, authHeader)
                    .catch(r => { if (r === 400) { setNotFound(true) } else { notifications.show({ title: "Could not get folder info", message: String(r), color: 'red', autoClose: 5000 }) }; return -1 })
            }

            SetFileData(fileData, dispatch, userInfo)

            const jumpItem = query.get("jumpTo")
            if (jumpItem) {
                dispatch({ type: 'set_scroll_to', fileId: jumpItem })
                dispatch({ type: 'set_selected', fileId: jumpItem, selected: true })
            }

            dispatch({ type: "set_loading", loading: false })
        }
        syncState()
    }, [userInfo.username, authHeader, filebrowserState.realId])

    // const doScan = useCallback(() => { dispatch({ type: 'set_loading', loading: true }); dispatchSync(realId, wsSend, filebrowserState.holdingShift, filebrowserState.holdingShift) }, [realId, wsSend, filebrowserState.holdingShift])

    // if (userInfo.username === "") {
    //     console.log("HERE!")
    //     return null
    // }

    return (
        <ColumnBox style={{ height: '100vh', backgroundColor: "#111418" }} >
            <HeaderBar searchContent={filebrowserState.searchContent} dispatch={dispatch} page={"files"} searchRef={searchRef} loading={filebrowserState.loading} progress={filebrowserState.scanProgress} />
            <DraggingCounter dragging={filebrowserState.draggingState} dirMap={filebrowserState.dirMap} selected={filebrowserState.selected} dispatch={dispatch} />
            <Presentation itemId={filebrowserState.presentingId} mediaData={filebrowserState.dirMap.get(filebrowserState.presentingId)?.mediaData} element={() => PresentationFile({ file: filebrowserState.dirMap.get(filebrowserState.presentingId) })} dispatch={dispatch} />
            <UploadStatus uploadState={uploadState} uploadDispatch={uploadDispatch} />
            <PasteImageDialogue img={filebrowserState.pasteImg} folderId={filebrowserState.realId} dirMap={filebrowserState.dirMap} authHeader={authHeader} dispatch={dispatch} wsSend={wsSend} />
            <FileContextMenu itemId={filebrowserState.menuTargetId} fbState={filebrowserState} open={filebrowserState.menuOpen} setOpen={o => dispatch({ type: 'set_menu_open', open: o })} menuPos={filebrowserState.menuPos} dispatch={dispatch} wsSend={wsSend} authHeader={authHeader} />
            <RowBox style={{ alignItems: 'flex-start' }}>
                <GlobalActions fbState={filebrowserState} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} />
                <DirViewWrapper
                    folderId={filebrowserState.realId}
                    folderName={filebrowserState.folderInfo?.filename}
                    dragging={filebrowserState.draggingState}
                    onDrop={(e => HandleDrop(e.dataTransfer.items, filebrowserState.realId, Array.from(filebrowserState.dirMap.values()).map((value: fileData) => value.filename), false, "", authHeader, uploadDispatch, wsSend))}
                    dispatch={dispatch}
                >
                    <Space h={10} />
                    <DirView filebrowserState={filebrowserState} folderId={filebrowserState.realId} notFound={notFound} setNotFound={setNotFound} alreadyScanned={alreadyScanned} setAlreadyScanned={setAlreadyScanned} dispatch={dispatch} wsSend={wsSend} uploadDispatch={uploadDispatch} authHeader={authHeader} />
                </DirViewWrapper>
            </RowBox>
        </ColumnBox>
    )
}

export default FileBrowser