import { useCallback, useContext, useEffect, useState } from "react"
import { FlexColumnBox, FlexRowBox } from "../FileBrowser/FilebrowserStyles"
import { MediaImage } from "../../components/PhotoContainer"
import { Box, Button, Menu, Popover, ScrollArea, Space, Text } from "@mantine/core"
import { GetAlbumMedia, GetAlbums, SetAlbumCover, ShareAlbum } from "../../api/GalleryApi"
import { ItemsWrapper } from "../../types/Styles"
import { IconPencil, IconPhoto } from "@tabler/icons-react"
import { AlbumData, MediaData } from "../../types/Types"
import { userContext } from "../../Context"
import { BucketCards } from "./MediaDisplay"
import { useNavigate } from "react-router-dom"
import { notifications } from "@mantine/notifications"
import { IconUsersGroup } from "@tabler/icons-react"
import { useMousePosition } from "../FileBrowser/FileBrowserLogic"
import { ShareInput } from "../FileBrowser/Share"

function ShareBox({ open, setOpen, pos, albumId }: { open: boolean, setOpen, pos: { x: number, y: number }, albumId }) {
    const { authHeader } = useContext(userContext)
    const [value, setValue] = useState(null)

    // if (!open) {
    //     return null
    // }

    return (
        <Popover opened={open} onClose={() => setOpen(false)} closeOnClickOutside>
            <Popover.Target>
                <Box style={{ position: 'fixed', top: pos.y, left: pos.x }} />
            </Popover.Target>
            <Popover.Dropdown>
                <ShareInput valueSetCallback={setValue} />
                <Space h={10} />
                <Button fullWidth color="#4444ff" onClick={() => { ShareAlbum(albumId, value, authHeader); setOpen(false) }}>Share</Button>
            </Popover.Dropdown>
        </Popover>
    )
}

function AlbumPreviewCard({ albumData, dispatch }: { albumData: AlbumData, dispatch }) {
    const [hovering, setHovering] = useState(false)
    const [menuOpen, setMenuOpen] = useState(false)
    const [shareOpen, setShareOpen] = useState(false)
    const nav = useNavigate()
    const { x, y } = useMousePosition()
    const [menuPos, setMenuPos] = useState({ x: 0, y: 0 })

    useEffect(() => {
        dispatch({ type: 'set_block_search_focus', block: shareOpen })
    }, [shareOpen])

    return (
        <FlexColumnBox
            key={albumData.Name}
            onContextMenu={(e) => { e.preventDefault(); setMenuPos({ x, y }); setMenuOpen(true) }}
            style={{ height: '280px', width: '250px', cursor: 'pointer', borderRadius: '6px', backgroundColor: hovering ? '#333333' : '#222222', alignItems: 'center' }}
        >

            <FlexColumnBox
                style={{ width: '90%', height: '225px', margin: 10, justifyContent: 'center' }}
                onMouseOver={() => setHovering(true)}
                onMouseLeave={() => setHovering(false)}
                onClick={() => { nav(`/albums/${albumData.Id}`) }}
            >
                {albumData.Cover && (
                    <MediaImage mediaId={albumData.Cover} quality='thumbnail' lazy={true} containerStyle={{ overflow: 'hidden', borderRadius: '6px' }} />
                )}
                {!albumData.Cover && (
                    <IconPhoto />
                )}
            </FlexColumnBox>

            <FlexRowBox style={{ justifyContent: 'space-between', width: '85%' }}>
                <Text c={'white'} style={{ userSelect: 'none' }}>{albumData.Name}</Text>
                <Text c={'white'} style={{ userSelect: 'none' }}>{albumData.Medias.length}</Text>
            </FlexRowBox>

            <Menu opened={menuOpen} onChange={setMenuOpen} transitionProps={{ transition: 'pop' }}>
                <Menu.Target>
                    <Box pos={'fixed'} top={menuPos.y} left={menuPos.x} />
                </Menu.Target>
                <Menu.Dropdown>
                    <Menu.Label>Album Actions</Menu.Label>
                    <Menu.Item leftSection={<IconPencil />}>
                        Rename
                    </Menu.Item>

                    <Menu.Item leftSection={<IconUsersGroup />} onClick={(e) => { e.stopPropagation(); setShareOpen(true) }}>
                        Share
                    </Menu.Item>
                </Menu.Dropdown>
            </Menu>

            <ShareBox open={shareOpen} setOpen={setShareOpen} pos={menuPos} albumId={albumData.Id} />

        </FlexColumnBox>
    )
}

function AlbumMediaContextMenu({ albumId, fetchAlbum, mediaId, open, setOpen, authHeader }) {
    const { x, y } = useMousePosition()
    const [menuPos, setMenuPos] = useState({ x: 0, y: 0 })
    useEffect(() => {
        if (open) {
            setMenuPos({ x, y })
        }
    }, [open])

    return (
        <Menu opened={open} onClose={() => setOpen(false)} closeOnClickOutside>
            <Menu.Target>
                <Box style={{ position: 'fixed', top: menuPos.y, left: menuPos.x }} />
            </Menu.Target>

            <Menu.Dropdown>
                <Menu.Label>Media Actions</Menu.Label>
                <Menu.Item leftSection={<IconPhoto />} onClick={(e) => { e.stopPropagation(); SetAlbumCover(albumId, mediaId, authHeader).then(fetchAlbum) }}>
                    Make cover photo
                </Menu.Item>
            </Menu.Dropdown>
        </Menu>
    )
}

function MenuFactory(albumId, fetchAlbum, authHeader) {
    const partialMenu = (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => <AlbumMediaContextMenu albumId={albumId} fetchAlbum={fetchAlbum} mediaId={mediaId} open={open} setOpen={setOpen} authHeader={authHeader} />
    return partialMenu
}

function Album({ albumId, dispatch }) {
    const { authHeader } = useContext(userContext)
    const [albumData, setAlbumData]: [albumData: { albumMeta: AlbumData, media: MediaData[] }, setAlbumData: any] = useState(null)
    const nav = useNavigate()

    const fetchAlbum = useCallback(() => {
        GetAlbumMedia(albumId, dispatch, authHeader).then(m => { dispatch({ type: 'set_media', media: m.media }); dispatch({ type: 'set_loading', albums: false }); setAlbumData(m) }).catch(r => { notifications.show({ title: "Failed to load album", message: r, color: 'red' }); dispatch({ type: 'set_loading', loading: false }) })
    }, [albumId])

    useEffect(() => {
        fetchAlbum()
    }, [fetchAlbum])

    if (!albumData) {
        return null
    }

    if (albumData.media.length == 0) {
        return (
            <FlexColumnBox>
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

    return (
        <ScrollArea h={'100%'} type='never'>
            <Text size={'75px'} fw={900} variant="gradient" gradient={{ from: `#${albumData.albumMeta.PrimaryColor}`, to: `#${albumData.albumMeta.SecondaryColor}`, deg: 45 }} style={{ display: 'flex', justifyContent: 'center', userSelect: 'none' }}>
                {albumData.albumMeta.Name}
            </Text>
            <Space h={25} />
            <BucketCards medias={albumData.media} scrollerRef={null} dispatch={dispatch} menu={MenuFactory(albumData.albumMeta.Id, fetchAlbum, authHeader)} />
            <Space h={25} />
        </ScrollArea>
    )
}

function AlbumsHomeView({ albumsMap, dispatch }: { albumsMap: Map<string, AlbumData>, dispatch }) {
    const { authHeader } = useContext(userContext)

    useEffect(() => {
        GetAlbums(authHeader).then((val) => { dispatch({ type: 'set_albums', albums: val }); dispatch({ type: 'set_loading', albums: false }) })
    }, [])

    let albumItems = []
    if (albumsMap.size) {
        const albums = Array.from(albumsMap.values())
        albumItems = albums.map((val) => {

            return (
                <AlbumPreviewCard key={val.Name} albumData={val} dispatch={dispatch} />
            )
        })
    }

    if (albumItems.length !== 0) {
        return (
            <ItemsWrapper size={250}>
                {albumItems}
            </ItemsWrapper>
        )
    } else {
        return (
            <FlexColumnBox>
                <Space h={200} />
                <Text> You have no albums, create one on the left </Text>
                {/* <Button onClick={() => dispatch({ type: 'set_new_album_open', open: true })}> New Album </Button> */}
            </FlexColumnBox>
        )
    }
}
export function Albums({ albumsMap, selectedAlbum, dispatch }: { albumsMap: Map<string, AlbumData>, selectedAlbum: string, dispatch }) {
    if (selectedAlbum == "") {
        return (
            <AlbumsHomeView albumsMap={albumsMap} dispatch={dispatch} />
        )
    } else {
        return (
            <Album albumId={selectedAlbum} dispatch={dispatch} />
        )
    }

}