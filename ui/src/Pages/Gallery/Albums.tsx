import { useCallback, useContext, useEffect, useMemo, useRef, useState } from "react"
import { FlexColumnBox, FlexRowBox, ItemVisualComponentWrapper } from "../FileBrowser/FilebrowserStyles"
import { MediaImage } from "../../components/PhotoContainer"
import { Box, Button, Card, Menu, Popover, ScrollArea, Space, Text, TextInput, Tooltip, TooltipFloating } from "@mantine/core"
import { DeleteAlbum, GetAlbumMedia, GetAlbums, RemoveMediaFromAlbum, RenameAlbum, SetAlbumCover, ShareAlbum } from "../../api/GalleryApi"
import { ItemsWrapper } from "../../types/Styles"
import { IconPencil, IconPhoto, IconTrash } from "@tabler/icons-react"
import { AlbumData, MediaData, MediaStateType } from "../../types/Types"
import { userContext } from "../../Context"
import { BucketCards } from "./MediaDisplay"
import { useNavigate } from "react-router-dom"
import { notifications } from "@mantine/notifications"
import { IconUsersGroup } from "@tabler/icons-react"
import { ShareInput } from "../../components/Share"
import NotFound from "../../components/NotFound"

function ShareBox({ open, setOpen, pos, albumId, sharedWith, fetchAlbums }: { open: boolean, setOpen, pos: { x: number, y: number }, albumId, sharedWith, fetchAlbums }) {
    const { authHeader } = useContext(userContext)
    const [value, setValue] = useState(sharedWith)

    useEffect(() => {
        setValue(sharedWith)
    }, [sharedWith])

    return (
        <Popover opened={open} onClose={() => setOpen(false)} closeOnClickOutside>
            <Popover.Target>
                <Box style={{ position: 'fixed', top: pos.y, left: pos.x }} />
            </Popover.Target>
            <Popover.Dropdown>
                <ShareInput valueSetCallback={setValue} initValues={sharedWith} />
                <Space h={10} />
                <Button fullWidth disabled={JSON.stringify(value) === JSON.stringify(sharedWith)} color="#4444ff" onClick={() => { ShareAlbum(albumId, authHeader, value.filter((v) => !sharedWith.includes(v)), sharedWith.filter((v) => !value.includes(v))).then(() => fetchAlbums()); setOpen(false) }}>
                    Update
                </Button>
            </Popover.Dropdown>
        </Popover>
    )
}

export const useKeyDown = (albumId, newName, editing, setEditing, setShareOpen, fetchAlbums, authHeader) => {

    const onKeyDown = useCallback((event) => {
        if (event.key === "Enter" && editing) {
            RenameAlbum(albumId, newName, authHeader).then(() => { setEditing(false); fetchAlbums() })
        } else if (event.key === "Escape") {
            setEditing(false)
            setShareOpen(false)
        }
    }, [editing, albumId, newName, authHeader, setEditing, fetchAlbums, setShareOpen])

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        }
    }, [onKeyDown])
}

function AlbumPreviewCard({ albumData, fetchAlbums, dispatch }: { albumData: AlbumData, fetchAlbums: () => void, dispatch }) {
    const { userInfo, authHeader } = useContext(userContext)

    const [editing, setEditing] = useState(false)
    const [hovering, setHovering] = useState(false)
    const [menuOpen, setMenuOpen] = useState(false)
    const [shareOpen, setShareOpen] = useState(false)

    const nav = useNavigate()
    const [menuPos, setMenuPos] = useState({ x: 0, y: 0 })
    const [renameVal, setRenameVal] = useState(albumData.Name)

    useKeyDown(albumData.Id, renameVal, editing, setEditing, setShareOpen, fetchAlbums, authHeader)

    useEffect(() => {
        dispatch({ type: 'set_block_search_focus', block: shareOpen })
    }, [shareOpen])

    const editRef: React.Ref<HTMLInputElement> = useRef()

    useEffect(() => {
        if (editing) {
            dispatch({ type: "set_block_search_focus", block: true })
            editRef.current.select()
        } else {
            dispatch({ type: "set_block_search_focus", block: false })
            setRenameVal(albumData.Name)
        }
    }, [editing])

    return (
        <Box>
            <Card
                onContextMenu={(e) => { e.preventDefault(); setMenuPos({ x: e.clientX, y: e.clientY }); setMenuOpen(true) }}
                onClick={() => { if (!menuOpen && !shareOpen) { nav(`/albums/${albumData.Id}`) } }}
                onMouseOver={() => setHovering(true)}
                onMouseLeave={() => setHovering(false)}
                style={{ display: 'flex', height: '100%', width: '100%', position: "relative", padding: 0, cursor: 'pointer', backgroundColor: hovering ? '#333333' : '#222222', alignItems: 'center' }}
            >
                <ItemVisualComponentWrapper >
                    <MediaImage mediaId={albumData.Cover} quality='thumbnail' lazy={true} containerStyle={{ overflow: 'hidden', borderRadius: '6px' }} />
                </ItemVisualComponentWrapper>

                {/* <FlexRowBox onClick={(e) => { e.stopPropagation(); setEditing(true) }} style={{ justifyContent: 'space-between', width: '85%', cursor: 'text', height: '36px' }}> */}
                {!editing && (
                    <FlexRowBox style={{height: '50px', justifyContent: 'space-between', alignItems: 'flex-start', paddingLeft: 10}}>
                        <FlexColumnBox style={{height: 'max-content', alignItems: 'flex-start', width: '175px', cursor: 'text'}} onClick={(e) => { e.stopPropagation(); setEditing(true) }}>
                            <TooltipFloating position='top' label={albumData.Name}>
                                <Text c='white' fw={500} truncate='end' w={'100%'}>{albumData.Name}</Text>
                            </TooltipFloating>
                            <Text size='12px'>{albumData.Medias.length}</Text>
                        </FlexColumnBox>
                        {albumData.Owner !== userInfo.username && (
                            <Tooltip label={`Shared by ${albumData.Owner}`}>
                                <IconUsersGroup color="white" style={{ margin: 10, zIndex: 1 }} />
                            </Tooltip>
                        )}
                    </FlexRowBox>
                )}

                {editing && (
                    <FlexRowBox onBlur={() => setEditing(false)} style={{ width: '85%' }}>
                        <TextInput
                            color={'white'}
                            autoFocus
                            ref={editRef}
                            variant='unstyled'
                            defaultValue={albumData.Name}
                            onClick={(e) => { e.stopPropagation() }}
                            onChange={(e) => { setRenameVal(e.target.value) }}
                            style={{ width: "100%" }}
                        />
                    </FlexRowBox>
                )}

                <Menu opened={menuOpen} onChange={setMenuOpen} transitionProps={{ transition: 'pop' }}>
                    <Menu.Target>
                        <Box pos={'fixed'} top={menuPos.y} left={menuPos.x} />
                    </Menu.Target>
                    <Menu.Dropdown>
                        <Menu.Label>Album Actions</Menu.Label>
                        <Menu.Item leftSection={<IconPencil />} onClick={() => setEditing(true)}>
                            Rename
                        </Menu.Item>

                        <Menu.Item leftSection={<IconUsersGroup />} onClick={(e) => { e.stopPropagation(); setShareOpen(true) }}>
                            Share
                        </Menu.Item>

                        {albumData.Owner === userInfo.username && (
                            <Menu.Item c={'red'} leftSection={<IconTrash />} onClick={(e) => { e.stopPropagation(); DeleteAlbum(albumData.Id, authHeader).then(() => fetchAlbums()) }}>
                                Delete
                            </Menu.Item>
                        )}
                        {albumData.Owner !== userInfo.username && (
                            <Menu.Item c={'red'} leftSection={<IconTrash />} onClick={(e) => { e.stopPropagation(); DeleteAlbum(albumData.Id, authHeader).then(() => fetchAlbums()) }}>
                                Leave
                            </Menu.Item>
                        )}

                    </Menu.Dropdown>
                </Menu>

                <ShareBox open={shareOpen} setOpen={setShareOpen} sharedWith={albumData.SharedWith} pos={menuPos} albumId={albumData.Id} fetchAlbums={fetchAlbums} />

            </Card>
        </Box>
    )
}

function AlbumMediaContextMenu({ albumId, fetchAlbum, mediaId, open, setOpen, authHeader }) {
    return (

        <Menu opened={open} onClose={() => setOpen(false)} closeOnClickOutside>
            <Menu.Target>
                <Box style={{ position: 'absolute' }} />
            </Menu.Target>

            <Menu.Dropdown>
                <Menu.Label>Media Actions</Menu.Label>

                <Menu.Item leftSection={<IconPhoto />} onClick={(e) => { e.stopPropagation(); SetAlbumCover(albumId, mediaId, authHeader).then(fetchAlbum) }}>
                    Make Cover Photo
                </Menu.Item>

                <Menu.Item leftSection={<IconTrash />} color='red' onClick={(e) => { e.stopPropagation(); RemoveMediaFromAlbum(albumId, [mediaId], authHeader).then(fetchAlbum) }}>
                    Remove From Album
                </Menu.Item>
            </Menu.Dropdown>
        </Menu>
    )
}

function MenuFactory(albumId, fetchAlbum, authHeader) {
    const partialMenu = (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => <AlbumMediaContextMenu albumId={albumId} fetchAlbum={fetchAlbum} mediaId={mediaId} open={open} setOpen={setOpen} authHeader={authHeader} />
    return partialMenu
}

function Album({ albumId, includeRaw, imageSize, searchContent, dispatch }) {
    const { authHeader } = useContext(userContext)
    const [albumData, setAlbumData]: [albumData: { albumMeta: AlbumData, media: MediaData[] }, setAlbumData: any] = useState(null)
    const [notFound, setNotFound] = useState(false)
    const nav = useNavigate()

    const fetchAlbum = useCallback(() => {
        GetAlbumMedia(albumId, includeRaw, dispatch, authHeader).then(m => {
            dispatch({ type: 'set_media', media: m.media });
            dispatch({ type: 'set_loading', albums: false }); setAlbumData(m)
        })
            .catch(r => {
                dispatch({ type: 'set_loading', loading: false })
                if (r === 404) {
                    setNotFound(true)
                    return
                }
                notifications.show({ title: "Failed to load album", message: r, color: 'red' });
            })
    }, [albumId, includeRaw])

    useEffect(() => {
        fetchAlbum()
    }, [fetchAlbum])

    const media = useMemo(() => {
        if (!albumData) {
            return []
        }
        if (searchContent === "") {
        }
        return albumData.media
        // return albumData.media.filter(v => v?.filename.toLowerCase().includes(searchContent))
    }, [albumData?.media])

    if (notFound) {
        return (
            <NotFound resourceType="Album" link="/albums" setNotFound={setNotFound}/>
        )
    }

    if (!albumData) {
        return null
    }

    if (albumData.media.length === 0) {
        return (
            <FlexColumnBox>
                <Text size={'75px'} fw={900} variant="gradient" style={{ display: 'flex', justifyContent: 'center', userSelect: 'none', lineHeight: 1.1 }}>
                    {albumData.albumMeta.Name}
                </Text>
                <FlexColumnBox style={{ paddingTop: '150px', width: 'max-content' }}>
                    <Text fw={800} size="30px">This album has no media
                    </Text>
                    <Space h={5} />
                    <Text size="23px">
                        You can add some in the filebrowser
                    </Text>
                    <Space h={20} />
                    <Button fullWidth color="#4444ff" onClick={() => nav('/files/home')}>Filebrowser</Button>
                </FlexColumnBox>
            </FlexColumnBox>
        )
    }
    const startColor = albumData.albumMeta.PrimaryColor ? `#${albumData.albumMeta.PrimaryColor}` : '#ffffff'
    const endColor = albumData.albumMeta.SecondaryColor ? `#${albumData.albumMeta.SecondaryColor}` : '#ffffff'

    return (
        <ScrollArea type="never" mah={'100vh'} style={{ paddingLeft: 25, paddingRight: 35, height: '100%', width: '100%' }}>
            <FlexColumnBox style={{ height: 'max-content' }}>
                <Text size={'75px'} fw={900} variant="gradient" gradient={{ from: startColor, to: endColor, deg: 45 }} style={{ display: 'flex', justifyContent: 'center', userSelect: 'none', lineHeight: 1.1 }}>
                    {albumData.albumMeta.Name}
                </Text>
            </FlexColumnBox>
            <Space h={25} />
            <BucketCards medias={media} scale={imageSize} scrollerRef={null} dispatch={dispatch} menu={MenuFactory(albumData.albumMeta.Id, fetchAlbum, authHeader)} />
            <Space h={25} />
        </ScrollArea>
    )
}

function AlbumsHomeView({ albumsMap, searchContent, dispatch }: { albumsMap: Map<string, AlbumData>, searchContent: string, dispatch }) {
    const { authHeader } = useContext(userContext)

    const fetchAlbums = useCallback(() => {
        GetAlbums(authHeader)
            .then((val) => {
                dispatch({ type: 'set_albums', albums: val })
                dispatch({ type: 'set_loading', albums: false })
            })
    }, [authHeader, dispatch])

    useEffect(() => {
        fetchAlbums()
    }, [fetchAlbums])

    let albumItems = []
    if (albumsMap.size) {
        const albums = Array.from(albumsMap.values())
        albumItems = albums.filter(val => val.Name.toLowerCase().includes(searchContent.toLowerCase())).map(val => <AlbumPreviewCard key={val.Name} albumData={val} fetchAlbums={fetchAlbums} dispatch={dispatch} />)
    }

    if (albumItems.length === 0) {
        return (
            <FlexColumnBox>
                <Space h={200} />
                <Text> You have no albums, create one on the left </Text>
            </FlexColumnBox>
        )
    } else {
        return (
            <ItemsWrapper size={300} style={{ paddingLeft: 25, paddingRight: 25 }}>
                {albumItems}
            </ItemsWrapper>
        )
    }
}

export function Albums({ mediaState, selectedAlbum, dispatch }: { mediaState: MediaStateType, selectedAlbum: string, dispatch }) {
    if (selectedAlbum === "") {
        return (
            <AlbumsHomeView albumsMap={mediaState.albumsMap} searchContent={mediaState.searchContent} dispatch={dispatch} />
        )
    } else {
        return (
            <Album albumId={selectedAlbum} includeRaw={mediaState.includeRaw} imageSize={mediaState.imageSize} searchContent={mediaState.searchContent} dispatch={dispatch} />
        )
    }
}