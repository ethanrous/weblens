import { ActionIcon, Box, Button, Combobox, Divider, Indicator, Modal, MultiSelect, ScrollArea, Slider, Space, Switch, Tabs, Text, TextInput, useCombobox } from '@mantine/core'
import { useEffect, useReducer, useMemo, useRef, useContext, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { IconCheck, IconFilter, IconPlus } from '@tabler/icons-react'

import HeaderBar from "../../components/HeaderBar"
import Presentation from '../../components/Presentation'
import { GalleryBucket } from './MediaDisplay'
import { mediaReducer, useKeyDown, handleWebsocket } from './GalleryLogic'
import { CreateAlbum, FetchData, GetAlbums } from '../../api/GalleryApi'
import { AlbumData, MediaData, MediaStateType } from '../../types/Types'
import useWeblensSocket from '../../api/Websocket'
import { userContext } from '../../Context'
import { FlexColumnBox, FlexRowBox } from '../FileBrowser/FilebrowserStyles'
import { Albums } from './Albums'

// styles

const NoMediaContainer = ({ ...children }) => {
    return (
        <Box
            style={{
                display: "flex",
                flexWrap: "wrap",
                flexDirection: "column",
                marginTop: "50px",
                gap: "25px",
                alignContent: "center"
            }}
            {...children}
        />
    )
}

const NoMediaDisplay = () => {
    const nav = useNavigate()
    return (
        <NoMediaContainer>
            <Text c='white' >No media to display</Text>
            <Button style={{ border: `1px solid white` }} onClick={() => nav('/files')}>
                Upload Media
            </Button>
        </NoMediaContainer>
    )
}

const TimelineControls = ({ rawSelected, albumsFilter, imageSize, albumsMap, dispatch }: { rawSelected: boolean, albumsFilter: string[], imageSize: number, albumsMap: Map<string, AlbumData>, dispatch }) => {
    const albumNames = useMemo(() => Array.from(albumsMap.values()).map((v) => v.Name), [albumsMap.size])
    const combobox = useCombobox({
        onDropdownClose: () => {
            combobox.resetSelectedOption()
            dispatch({ type: 'set_albums_filter', albumNames: selectedAlbums })
            dispatch({ type: 'set_raw_toggle', raw: rawOn})
        }
    });
    const [selectedAlbums, setSelectedAlbums] = useState(albumsFilter)
    const [rawOn, setRawOn] = useState(rawSelected)

    const albumsOptions = useMemo(() => {
        const options = albumNames.map((name) => {
            return (
                <Combobox.Option value={name} key={name}>
                    <FlexRowBox style={{justifyContent: 'space-between'}}>
                        <Text>{name}</Text>
                        {selectedAlbums.includes(name) && (
                            <IconCheck />
                        )}
                    </FlexRowBox>
                </Combobox.Option>
            )
        })
        return options
    }, [albumNames, selectedAlbums])

    return (
        <FlexRowBox>
            <Slider color='#4444ff' label={`Image Height: ${imageSize}px`} defaultValue={300} value={imageSize} w={200} min={100} max={500} step={10} onChange={(e) => dispatch({ type: 'set_image_size', size: e })} onDoubleClick={() => dispatch({ type: 'set_image_size', size: 300 })} />
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
                        <IconFilter onClick={() => combobox.toggleDropdown()} style={{cursor: 'pointer'}}/>
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
        </FlexRowBox>
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
            <FlexRowBox style={{ width: '100%' }}>
                <Switch color='#4444ff' checked={rawSelected} label={'RAWs'} onChange={(e) => dispatch({ type: 'set_raw_toggle', raw: e.target.checked })} />
                <Space w={20} />
                <Slider color='#4444ff' defaultValue={300} value={imageSize} w={200} min={100} max={500} step={10} onChange={(e) => dispatch({ type: 'set_image_size', size: e })} onDoubleClick={() => dispatch({ type: 'set_image_size', size: 300 })} />
            </FlexRowBox>
        )
    }

}

const GalleryControls = ({ mediaState, page, albumId, dispatch }: { mediaState: MediaStateType, page: string, albumId: string, dispatch }) => {
    return (
        <FlexRowBox style={{ position: 'sticky', alignSelf: 'flex-stat', marginLeft: 225, height: 65, zIndex: 10 }}>
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
        </FlexRowBox>
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

            <Tabs.Panel value='timeline' style={{ height: "96%" }}>
                <Space h={15} />
                {timeline}
            </Tabs.Panel>
            <Tabs.Panel value='albums' style={{ height: "96%" }}>
                <Space h={15} />
                {albums}
            </Tabs.Panel>
        </Tabs>
    )
}

function InfiniteScroll({ mediaState, page, dispatch }: { mediaState: MediaStateType, page: string, dispatch: (value) => void }) {
    const { authHeader } = useContext(userContext)
    const scrollerRef = useRef(null)
    useEffect(() => scrollerRef.current!.scrollTo({ top: 0, behavior: 'instant' }), [mediaState.includeRaw])
    useEffect(() => { FetchData(mediaState, dispatch, authHeader).then(() => dispatch({ type: 'set_loading', loading: false })) }, [mediaState.includeRaw, mediaState.albumsFilter, page, authHeader])

    let groupBy = "date"

    const groupMap: Map<string, Array<MediaData>> = useMemo(() => {
        let groupMap = new Map<string, Array<MediaData>>()

        if (mediaState.mediaMap.size === 0) {
            return groupMap
        }

        for (let value of mediaState.mediaMap.values()) {
            if (groupBy === "date") {
                const dateObj = new Date(value.createDate.split("T")[0])
                const dateString = dateObj.toUTCString().split(" 00:00:00 GMT")[0]
                if (groupMap.get(dateString) == null) {
                    groupMap.set(dateString, [value])
                } else {
                    groupMap.get(dateString).push(value)
                }
            } else if (groupBy === "filetype") {
                const ext = value.mediaType.FriendlyName
                if (groupMap.get(ext) == null) {
                    groupMap.set(ext, [value])
                } else {
                    groupMap.get(ext).push(value)
                }
            } else {
                if (groupMap.get("") == null) {
                    groupMap.set("", [])
                }
                groupMap.get("").push(value)
            }
        }
        return groupMap
    }, [mediaState.mediaMapUpdated, mediaState.mediaMap])

    const [groups, numShownItems]: [JSX.Element[], number] = useMemo(() => {
        if (!groupMap || !scrollerRef?.current) { return [[], 0] }
        let counter = 0
        const buckets = Array.from(groupMap.keys()).map((title) => {
            const items = groupMap.get(title)//.filter((item) => { return mediaState.searchContent ? item.filename.toLowerCase().includes(mediaState.searchContent.toLowerCase()) : true })
            if (items.length === 0) { return null }
            counter += items.length
            return (<GalleryBucket key={title} bucketTitle={title} scale={mediaState.imageSize} bucketData={items} scrollerRef={scrollerRef} dispatch={dispatch} />)
        })
        return [buckets, counter]
    }, [groupMap, scrollerRef.current, mediaState.searchContent, mediaState.imageSize, dispatch])

    return (
        <ScrollArea viewportRef={scrollerRef} type='never' style={{ height: '100%', width: '100%', paddingLeft: 25, paddingRight: 25 }} onDoubleClick={() => console.log(`DEBUG: Showing ${numShownItems} images`)}>
            <FlexColumnBox style={{minHeight: '100vh'}}>
                {groups}
                {!mediaState.loading && groups.length === 0 && (
                    <NoMediaDisplay />
                )}
            </FlexColumnBox>
            {!mediaState.loading && (
                <Text c='white' style={{ textAlign: "center", paddingTop: "80px", paddingBottom: "10px", userSelect: 'none' }}> Wow, you scrolled this whole way? </Text>
            )}
        </ScrollArea>
    )
}

const Gallery = () => {
    const [mediaState, dispatch]: [MediaStateType, React.Dispatch<any>] = useReducer(mediaReducer, {
        mediaMap: new Map<string, MediaData>(),
        mediaMapUpdated: 0,
        albumsMap: new Map<string, AlbumData>(),
        albumsFilter: [],
        loading: true,
        includeRaw: false,
        newAlbumDialogue: false,
        blockSearchFocus: false,
        imageSize: 300,
        scanProgress: 0,
        searchContent: "",
        presentingHash: "",
    })

    const { authHeader } = useContext(userContext)

    const loc = useLocation()
    const page = loc.pathname === "/" || loc.pathname === "/timeline" ? 'timeline' : "albums"
    const albumId = useParams()["*"]
    const { wsSend, lastMessage } = useWeblensSocket()

    const searchRef = useRef()
    useKeyDown(mediaState.blockSearchFocus, searchRef)
    useEffect(() => { handleWebsocket(lastMessage, dispatch) }, [lastMessage])

    useEffect(() => {
        if (authHeader.Authorization !== "") {
            GetAlbums(authHeader)
                .then((val) => {
                    dispatch({ type: 'set_albums', albums: val })
                    dispatch({ type: 'set_loading', albums: false })
                })
        }
    }, [authHeader])

    return (
        <Box>
            <HeaderBar folderId={"home"} searchContent={mediaState.searchContent} dispatch={dispatch} wsSend={wsSend} page={"gallery"} searchRef={searchRef} loading={mediaState.loading} progress={mediaState.scanProgress} />
            <Presentation mediaData={mediaState.mediaMap.get(mediaState.presentingHash)} parents={null} dispatch={dispatch} />
            <FlexRowBox style={{ height: "100vh", paddingTop: "70px", alignItems: 'normal' }}>
                <GalleryControls mediaState={mediaState} page={page} albumId={albumId} dispatch={dispatch} />
                <Box style={{ height: "calc(100% - 80px)", width: '100%', paddingTop: "15px", position: 'absolute' }}>
                    <ViewSwitch
                        page={page}
                        timeline={
                            <InfiniteScroll mediaState={mediaState} page={page} dispatch={dispatch} />
                        }
                        albums={
                            <Albums mediaState={mediaState} selectedAlbum={albumId} dispatch={dispatch} />
                        }
                        albumId={albumId}
                    />
                </Box>
            </FlexRowBox>
        </Box>
    )
}

export default Gallery
