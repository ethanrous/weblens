import { Box, Button, Combobox, Divider, Indicator, Modal, Slider, Space, Switch, Tabs, Text, TextInput, useCombobox } from '@mantine/core'
import { useEffect, useReducer, useMemo, useRef, useContext, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { IconCheck, IconFilter, IconPlus } from '@tabler/icons-react'

import HeaderBar from "../../components/HeaderBar"
import Presentation from '../../components/Presentation'
import { PhotoGallery } from '../../components/MediaDisplay'
import { mediaReducer, useKeyDown, handleWebsocket } from './GalleryLogic'
import { CreateAlbum, FetchData, GetAlbums } from '../../api/GalleryApi'
import { AlbumData, MediaData, MediaStateType } from '../../types/Types'
import useWeblensSocket from '../../api/Websocket'
import { userContext } from '../../Context'
import { ColumnBox, RowBox } from '../FileBrowser/FilebrowserStyles'
import { Albums } from './Albums'

const NoMediaDisplay = () => {
    const nav = useNavigate()
    return (
        <ColumnBox style={{ marginTop: 75, gap: 25, width: 'max-content' }}>
            <Text c='white' fw={700} size='31px'>No media to display</Text>
            <Text c='white' >Upload files then add them to an album</Text>
            <RowBox style={{ height: 'max-content', width: '100%', gap: 10 }}>
                <Button fullWidth color='#4444ff' onClick={() => nav('/files')}>
                    Upload Media
                </Button>
                <Button fullWidth color='#4444ff' onClick={() => nav('/albums')}>
                    View Albums
                </Button>
            </RowBox>
        </ColumnBox>
    )
}

const ImageSizeSlider = ({ imageSize, dispatch }) => {
    return (
        <Slider
            color='#4444ff'
            label={`Image Size`}
            defaultValue={300}
            value={imageSize}
            w={200}
            min={100}
            max={500}
            step={10}
            marks={[
                { value: 100, label: '10%' },
                { value: 300, label: '50%' },
                { value: 500, label: '100%' }
            ]}
            onChange={(e) => dispatch({ type: 'set_image_size', size: e })}
            onDoubleClick={() => dispatch({ type: 'set_image_size', size: 300 })}
        />
    )
}

const TimelineControls = ({ rawSelected, albumsFilter, imageSize, albumsMap, dispatch }: { rawSelected: boolean, albumsFilter: string[], imageSize: number, albumsMap: Map<string, AlbumData>, dispatch }) => {
    const albumNames = useMemo(() => Array.from(albumsMap.values()).map((v) => v.Name), [albumsMap.size])
    const combobox = useCombobox({
        onDropdownClose: () => {
            combobox.resetSelectedOption()
            dispatch({ type: 'set_albums_filter', albumNames: selectedAlbums })
            dispatch({ type: 'set_raw_toggle', raw: rawOn })
        }
    });
    const [selectedAlbums, setSelectedAlbums] = useState(albumsFilter)
    const [rawOn, setRawOn] = useState(rawSelected)

    const albumsOptions = useMemo(() => {
        const options = albumNames.map((name) => {
            return (
                <Combobox.Option value={name} key={name}>
                    <RowBox style={{ justifyContent: 'space-between' }}>
                        <Text>{name}</Text>
                        {selectedAlbums.includes(name) && (
                            <IconCheck />
                        )}
                    </RowBox>
                </Combobox.Option>
            )
        })
        return options
    }, [albumNames, selectedAlbums])

    return (
        <RowBox>
            <ImageSizeSlider imageSize={imageSize} dispatch={dispatch} />
            <Space w={20} />
            <Combobox
                store={combobox}
                width={200}
                position="bottom-start"
                withArrow
                withinPortal={false}
                positionDependencies={[selectedAlbums]}
                onOptionSubmit={(val) => {
                    setSelectedAlbums((current) => current.includes(val) ? current.filter((item) => item !== val) : [...current, val])
                }}
            >
                <Combobox.Target>
                    <Indicator color='#4444ff' disabled={!selectedAlbums.length && !rawSelected}>
                        <IconFilter onClick={() => combobox.toggleDropdown()} style={{ cursor: 'pointer' }} />
                    </Indicator>
                </Combobox.Target>

                <Combobox.Dropdown>
                    <Combobox.Header>
                        <Text>Timeline Filters</Text>
                    </Combobox.Header>
                    <Space h={10} />
                    <Switch color='#4444ff' checked={rawOn} label={'Show RAWs'} onChange={e => setRawOn(e.target.checked)} />
                    <Space h={10} />
                    <Combobox.Options>
                        <Combobox.Group label="Albums">
                            {albumsOptions}
                        </Combobox.Group>
                        <Combobox.Group label="Filetypes">
                            {/* {albumsOptions} */}
                        </Combobox.Group>
                    </Combobox.Options>
                </Combobox.Dropdown>

            </Combobox>
        </RowBox>
    )
}

const AlbumsControls = ({ albumId, imageSize, rawSelected, dispatch }) => {
    const [newAlbumModal, setNewAlbumModal] = useState(false)
    const [newAlbumName, setNewAlbumName] = useState("")
    const { authHeader } = useContext(userContext)

    if (!albumId) {
        return (
            <Box>
                <Button color='#4444ff' onClick={() => { dispatch({ type: 'set_block_search_focus', block: true }); setNewAlbumModal(true) }} leftSection={<IconPlus />}>
                    New Album
                </Button>

                <Modal opened={newAlbumModal} onClose={() => { dispatch({ type: 'set_block_search_focus', block: false }); setNewAlbumModal(false) }} title="New Album">
                    <TextInput value={newAlbumName} placeholder='Album name' onChange={(e) => setNewAlbumName(e.currentTarget.value)} />
                    <Space h={'md'} />
                    <Button onClick={() => { CreateAlbum(newAlbumName, authHeader).then(() => GetAlbums(authHeader).then((val) => dispatch({ type: 'set_albums', albums: val }))); dispatch({ type: 'set_block_search_focus', block: false }); setNewAlbumModal(false) }}>
                        Create
                    </Button>
                </Modal>
            </Box>
        )
    } else {
        return (
            <RowBox style={{ width: '100%' }}>
                <Switch color='#4444ff' checked={rawSelected} label={'RAWs'} onChange={(e) => dispatch({ type: 'set_raw_toggle', raw: e.target.checked })} />
                <Space w={20} />
                <ImageSizeSlider imageSize={imageSize} dispatch={dispatch} />
            </RowBox>
        )
    }

}

const GalleryControls = ({ mediaState, page, albumId, dispatch }: { mediaState: MediaStateType, page: string, albumId: string, dispatch }) => {
    return (
        <RowBox style={{ position: 'sticky', alignSelf: 'flex-stat', marginLeft: 225, height: 65, zIndex: 10 }}>
            <Divider my={10} size={1} orientation='vertical' />
            <Space w={30} />
            <Box
                style={{
                    width: '100%',
                    borderRadius: "6px",
                }}
            >
                {page === "timeline" && (
                    <TimelineControls rawSelected={mediaState.includeRaw} albumsFilter={mediaState.albumsFilter} imageSize={mediaState.imageSize} albumsMap={mediaState.albumsMap} dispatch={dispatch} />
                )}

                {page === "albums" && (
                    <AlbumsControls albumId={albumId} imageSize={mediaState.imageSize} rawSelected={mediaState.includeRaw} dispatch={dispatch} />
                )}
            </Box>
        </RowBox>
    )
}

function ViewSwitch({ page, timeline, albums, albumId }) {
    const nav = useNavigate()
    const [hovering, setHovering] = useState(false)
    let albumStyle = {}
    if (albumId && hovering) {
        albumStyle = { backgroundColor: '#2e2e2e', outline: '1px solid #4444ff' }
    }
    else if (albumId) {
        albumStyle = { backgroundColor: '#00000000', outline: '1px solid #4444ff' }
    }
    return (
        <Tabs value={page} keepMounted={false} onChange={(p) => nav(`/${p}`)} variant='pills' style={{ height: "100%" }}>
            <Tabs.List style={{ marginLeft: 20 }}>
                <Tabs.Tab value='timeline' color='#4444ff'>
                    Timeline
                </Tabs.Tab>
                <Tabs.Tab value='albums' color='#4444ff' onMouseOver={() => setHovering(true)} onMouseLeave={() => setHovering(false)} style={albumStyle}>
                    Albums
                </Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel value='timeline' style={{ height: "98%" }}>
                <ColumnBox>
                    {timeline}
                </ColumnBox>
            </Tabs.Panel>
            <Tabs.Panel value='albums' style={{ height: "98%" }}>
                <ColumnBox style={{ alignItems: 'center' }}>
                    {albums}
                </ColumnBox>
            </Tabs.Panel>
        </Tabs>
    )
}

export function Timeline({ mediaState, page, dispatch }: { mediaState: MediaStateType, page: string, dispatch: (value) => void }) {
    const { authHeader } = useContext(userContext)
    useEffect(() => { FetchData(mediaState, dispatch, authHeader).then(() => dispatch({ type: 'set_loading', loading: false })) }, [mediaState.includeRaw, mediaState.albumsFilter, page, authHeader])
    // useEffect(() => scrollerRef.current!.scrollTo({ top: 0, behavior: 'instant' }), [mediaState.includeRaw, mediaState.albumsFilter])
    const medias = useMemo(() => {
        return Array.from(mediaState.mediaMap.values()).filter(v => {
            if (mediaState.searchContent === "") {
                return true
            }
            if (!v.recognitionTags) {
                return false
            }
            for (const tag of v.recognitionTags) {
                if (tag.includes(mediaState.searchContent)) {
                    return true
                }
            }
            return false
        }).reverse()
    }, [mediaState.mediaMap.size, mediaState.searchContent])

    // let groupBy = "date"

    // const groupMap: Map<string, Array<MediaData>> = useMemo(() => {
    //     let groupMap = new Map<string, Array<MediaData>>()

    //     if (mediaState.mediaMap.size === 0) {
    //         return groupMap
    //     }
    //     const limit: number = 1000
    //     let counter = 0
    //     for (let value of mediaState.mediaMap.values()) {
    //         if (groupBy === "date") {
    //             const dateObj = new Date(value.createDate.split("T")[0])
    //             const dateString = dateObj.toUTCString().split(" 00:00:00 GMT")[0]
    //             if (groupMap.get(dateString) == null) {
    //                 groupMap.set(dateString, [value])
    //             } else {
    //                 groupMap.get(dateString).push(value)
    //             }
    //         } else if (groupBy === "filetype") {
    //             const ext = value.mediaType.FriendlyName
    //             if (groupMap.get(ext) == null) {
    //                 groupMap.set(ext, [value])
    //             } else {
    //                 groupMap.get(ext).push(value)
    //             }
    //         } else {
    //             if (groupMap.get("") == null) {
    //                 groupMap.set("", [])
    //             }
    //             groupMap.get("").push(value)
    //         }
    //         counter += 1
    //         if (counter >= limit) {
    //             break
    //         }
    //     }
    //     return groupMap
    // }, [mediaState.mediaMapUpdated, mediaState.mediaMap])

    if (mediaState.loading) {
        return null
    }
    if (medias.length === 0) {
        return (
            <NoMediaDisplay />
        )
    }
    return (
        <ColumnBox style={{ padding: 15 }}>
            <PhotoGallery medias={medias} imageBaseScale={mediaState.imageSize} dispatch={dispatch} />
        </ColumnBox>
    )
}

const Gallery = () => {
    const [mediaState, dispatch]: [MediaStateType, React.Dispatch<any>] = useReducer(mediaReducer, {
        mediaMap: new Map<string, MediaData>(),
        mediaMapUpdated: 0,
        albumsMap: new Map<string, AlbumData>(),
        presentingMedia: null,
        albumsFilter: [],
        loading: true,
        includeRaw: false,
        newAlbumDialogue: false,
        blockSearchFocus: false,
        imageSize: 300,
        showingCount: 300,
        scanProgress: 0,
        searchContent: "",
    })

    const { authHeader } = useContext(userContext)

    const loc = useLocation()
    const page = loc.pathname === "/" || loc.pathname === "/timeline" ? 'timeline' : "albums"
    const albumId = useParams()["*"]
    const { lastMessage } = useWeblensSocket()

    const viewportRef: React.Ref<HTMLDivElement> = useRef()

    const searchRef = useRef()
    useKeyDown(mediaState.blockSearchFocus, searchRef)
    useEffect(() => { handleWebsocket(lastMessage, dispatch) }, [lastMessage])

    useEffect(() => {
        if (authHeader.Authorization !== "" && page !== "albums") {
            GetAlbums(authHeader)
                .then((val) => {
                    dispatch({ type: 'set_albums', albums: val })
                    dispatch({ type: 'set_loading', loading: false })
                })
        }
    }, [authHeader, page])

    return (
        <Box>
            <HeaderBar searchContent={mediaState.searchContent} dispatch={dispatch} page={"gallery"} searchRef={searchRef} loading={mediaState.loading} progress={mediaState.scanProgress} />
            <Presentation mediaData={mediaState.presentingMedia} dispatch={dispatch} />
            <RowBox style={{ height: "100vh", alignItems: 'normal' }}>
                <GalleryControls mediaState={mediaState} page={page} albumId={albumId} dispatch={dispatch} />
                <Box ref={viewportRef} style={{ height: "calc(100% - 80px)", width: '100%', paddingTop: "15px", position: 'absolute' }}>
                    <ViewSwitch
                        page={page}
                        timeline={
                            <Timeline mediaState={mediaState} page={page} dispatch={dispatch} />
                        }
                        albums={
                            <Albums mediaState={mediaState} selectedAlbum={albumId} dispatch={dispatch} />
                        }
                        albumId={albumId}
                    />
                </Box>
            </RowBox>
        </Box>
    )
}

export default Gallery
