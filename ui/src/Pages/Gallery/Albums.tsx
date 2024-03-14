import { useCallback, useContext, useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"

import { notifications } from "@mantine/notifications"
import { Box, Button, Divider, Menu, Popover, Space, Text } from "@mantine/core"
import { IconClearAll, IconExclamationCircle, IconPhoto, IconTrash, IconUsersGroup } from "@tabler/icons-react"

import { ColumnBox, RowBox } from "../FileBrowser/FilebrowserStyles"
import { CleanAlbum, DeleteAlbum, GetAlbumMedia, GetAlbums, RemoveMediaFromAlbum, RenameAlbum, SetAlbumCover, ShareAlbum } from "../../api/GalleryApi"
import { AlbumData, MediaData, MediaStateType, getBlankMedia } from "../../types/Types"
import { userContext } from "../../Context"
import { PhotoGallery } from "../../components/MediaDisplay"
import NotFound from "../../components/NotFound"
import { GlobalContextType, ItemProps } from "../../components/ItemDisplay"
import { ItemScroller } from "../../components/ItemScroller"


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
                {/* <ShareInput valueSetCallback={setValue} initValues={sharedWith} /> */}
                <Space h={10} />
                <Button fullWidth disabled={JSON.stringify(value) === JSON.stringify(sharedWith)} color="#4444ff" onClick={() => { ShareAlbum(albumId, authHeader, value.filter((v) => !sharedWith.includes(v)), sharedWith.filter((v) => !value.includes(v))).then(() => fetchAlbums()); setOpen(false) }}>
                    Update
                </Button>
            </Popover.Dropdown>
        </Popover>
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
            dispatch({ type: 'set_media', medias: m.media })
            dispatch({ type: 'set_loading', albums: false })
            setAlbumData(m)
        })
            .catch(r => {
                dispatch({ type: 'set_loading', loading: false })
                if (r === 404) {
                    setNotFound(true)
                    return
                }
                notifications.show({ title: "Failed to load album", message: String(r), color: 'red' });
            })
    }, [albumId, includeRaw])

    useEffect(() => {
        fetchAlbum()
    }, [fetchAlbum])

    const media = useMemo(() => {
        if (!albumData) {
            return []
        }

        const media = albumData.media.filter(v => {
            if (searchContent === "") {
                return true
            }
            if (!v.recognitionTags) {
                return false
            }
            for (const tag of v.recognitionTags) {
                if (tag.includes(searchContent)) {
                    return true
                }
            }
            return false
        }).reverse()
        media.unshift()

        return media
    }, [albumData?.media, searchContent])

    if (notFound) {
        return (
            <NotFound resourceType="Album" link="/albums" setNotFound={setNotFound} />
        )
    }

    console.log(albumData)
    if (!albumData) {
        return null
        return (
            <RowBox style={{ height: 'max-content', justifyContent: 'center', paddingTop: '25vh' }}>
                <IconExclamationCircle color="red" />
                <Space w={8} />
                <Text c='red' fw={600} size={'30px'}>Album failed to load</Text>
            </RowBox>
        )
    }

    if (media.length === 0) {
        return (
            <ColumnBox>
                <Text size={'75px'} fw={900} variant="gradient" style={{ display: 'flex', justifyContent: 'center', userSelect: 'none', lineHeight: 1.1 }}>
                    {albumData.albumMeta.Name}
                </Text>
                <ColumnBox style={{ paddingTop: '150px', width: 'max-content' }}>
                    <Text fw={800} size="30px">This album has no media
                    </Text>
                    <Space h={5} />
                    <Text size="23px">
                        You can add some in the filebrowser
                    </Text>
                    <Space h={20} />
                    <Button fullWidth color="#4444ff" onClick={() => nav('/files/home')}>Filebrowser</Button>
                </ColumnBox>
            </ColumnBox>
        )
    }
    const startColor = albumData.albumMeta.PrimaryColor ? `#${albumData.albumMeta.PrimaryColor}` : '#ffffff'
    const endColor = albumData.albumMeta.SecondaryColor ? `#${albumData.albumMeta.SecondaryColor}` : '#ffffff'

    return (
        <ColumnBox >
            <Space h={10} />
            <PhotoGallery medias={media} imageBaseScale={imageSize} title={(
                <ColumnBox style={{ height: 'max-content' }}>
                    <Text size={'75px'} fw={900} variant="gradient" gradient={{ from: startColor, to: endColor, deg: 45 }} style={{ display: 'flex', justifyContent: 'center', userSelect: 'none', lineHeight: 1.1 }}>
                        {albumData.albumMeta.Name}
                    </Text>
                </ColumnBox>
            )} dispatch={dispatch} />
        </ColumnBox>
    )
}

function AlbumCoverMenu({ albumData, open, setMenuOpen, fetchAlbums, menuPos, dispatch }: { albumData: AlbumData, open, setMenuOpen: (o: boolean) => void, fetchAlbums, menuPos, dispatch }) {
    const { userInfo, authHeader } = useContext(userContext)
    const [shareOpen, setShareOpen] = useState(false)

    if (!albumData) {
        return
    }

    return (
        <Box>
            <Menu opened={open} position="right-start" onChange={setMenuOpen} transitionProps={{ transition: 'pop' }}>
                <Menu.Target>
                    <Box style={{ position: 'absolute', top: menuPos.y - 60, left: menuPos.x }} />
                </Menu.Target>
                <Menu.Dropdown>
                    <Menu.Label>Album Actions</Menu.Label>

                    <Menu.Item leftSection={<IconUsersGroup />} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'set_block_focus', block: true }); setShareOpen(true) }}>
                        Share
                    </Menu.Item>

                    <ColumnBox style={{ height: 'max-content', padding: '3px' }}>
                        <Divider w={'90%'} />
                    </ColumnBox>

                    {albumData.Owner === userInfo.username && (
                        <Menu.Item c={'red'} leftSection={<IconClearAll />} onClick={(e) => { e.stopPropagation(); CleanAlbum(albumData.Id, authHeader).then(() => fetchAlbums()) }}>
                            Clean Missing
                        </Menu.Item>
                    )}
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

            <ShareBox open={shareOpen} setOpen={o => { setShareOpen(o); dispatch({ type: 'set_block_focus', block: o }) }} sharedWith={albumData.SharedWith} pos={menuPos} albumId={albumData.Id} fetchAlbums={fetchAlbums} />
        </Box>
    )
}

function AlbumsHomeView({ albumsMap, menuOpen, menuTarget, menuPos, searchContent, dispatch }: { albumsMap: Map<string, AlbumData>, menuOpen, menuTarget, menuPos, searchContent: string, dispatch }) {
    const { authHeader, userInfo } = useContext(userContext)
    const nav = useNavigate()

    const fetchAlbums = useCallback(() => {
        GetAlbums(authHeader)
            .then((val) => {
                dispatch({ type: 'set_albums', albums: val })
                dispatch({ type: 'set_loading', loading: false })
            })
    }, [authHeader, dispatch])

    useEffect(() => {
        if (albumsMap.size === 0) {
            fetchAlbums()
        }
    }, [fetchAlbums, albumsMap.size])

    const albums = useMemo(() => {
        const albums = Array.from(albumsMap.values()).filter(val => val.Name.toLowerCase().includes(searchContent.toLowerCase())).map(v => {
            if (!v.Cover || !v.CoverMedia) {
                v.CoverMedia = getBlankMedia()
                v.CoverMedia.fileHash = v.Cover
            }
            const item: ItemProps = {
                itemId: v.Id,
                itemTitle: v.Name,
                secondaryInfo: String(v.Medias.length),
                selected: 0,
                mediaData: v.CoverMedia,
                droppable: false,
                isDir: false,
                imported: true,
                displayable: true
            }
            if (v.Owner !== userInfo.username) {
                item.extraIcons = [IconUsersGroup]
            }
            return item
        })
        return albums
    }, [JSON.stringify(Array.from(albumsMap.values())), searchContent])

    const albumsContext = useMemo(() => {
        const ctx: GlobalContextType = {
            visitItem: (itemId: string) => nav(itemId),
            setDragging: () => { },
            blockFocus: (b: boolean) => dispatch({ type: "set_block_focus", block: b }),
            setSelected: () => { },
            setMenuOpen: (o: boolean) => dispatch({ type: 'set_menu_open', open: o }),
            setMenuPos: (pos) => { dispatch({ type: 'set_menu_pos', pos: pos }) },
            setMenuTarget: (target: string) => { dispatch({ type: 'set_menu_target', targetId: target }) },
            rename: (itemId: string, newName: string) => { RenameAlbum(itemId, newName, authHeader).then(v => fetchAlbums()) },
            doMediaFetch: true
        }
        return ctx
    }, [albumsMap.size, nav, dispatch])

    if (albums.length === 0) {
        return (
            <ColumnBox>
                <Space h={200} />
                <Text> You have no albums </Text>
            </ColumnBox>
        )
    } else {
        return (
            <Box style={{ width: '100%', height: '100%', padding: 10 }}>
                <AlbumCoverMenu albumData={albumsMap.get(menuTarget)} open={menuOpen} setMenuOpen={o => dispatch({ type: 'set_menu_open', open: o })} fetchAlbums={fetchAlbums} menuPos={menuPos} dispatch={dispatch} />
                <ItemScroller itemsContext={albums} globalContext={albumsContext} />
            </Box>
        )
    }
}

export function Albums({ mediaState, selectedAlbum, dispatch }: { mediaState: MediaStateType, selectedAlbum: string, dispatch }) {
    if (selectedAlbum === "") {
        return (
            <AlbumsHomeView albumsMap={mediaState.albumsMap} menuOpen={mediaState.menuOpen} menuTarget={mediaState.menuTargetId} menuPos={mediaState.menuPos} searchContent={mediaState.searchContent} dispatch={dispatch} />
        )
    } else {
        return (
            <Album albumId={selectedAlbum} includeRaw={mediaState.includeRaw} imageSize={mediaState.imageSize} searchContent={mediaState.searchContent} dispatch={dispatch} />
        )
    }
}