import { useEffect, useReducer, useMemo, useRef, useContext, useState } from 'react'

import HeaderBar from "../../components/HeaderBar"
import Presentation from '../../components/Presentation'
import { GalleryBucket } from './MediaDisplay'
import { mediaReducer, useKeyDown, handleWebsocket } from './GalleryLogic'
import { CreateAlbum, FetchData, GetAlbums } from '../../api/GalleryApi'
import { AlbumData, MediaData, MediaStateType } from '../../types/Types'
import useWeblensSocket from '../../api/Websocket'
import { userContext } from '../../Context'

import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { Box, Button, Center, Divider, Loader, Modal, Paper, ScrollArea, Space, Switch, Tabs, Text, TextInput } from '@mantine/core'
import { FlexColumnBox, FlexRowBox } from '../FileBrowser/FilebrowserStyles'
import { MediaImage, useIsVisible } from '../../components/PhotoContainer'
import { ItemsWrapper } from '../../types/Styles'
import { IconPhoto, IconPlus } from '@tabler/icons-react'
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

// funcs

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

const TimelineControls = ({ rawSelected, dispatch }) => {
    return (
        <Box>
            <Switch checked={rawSelected} label={'RAWs'} onChange={() => dispatch({ type: 'toggle_raw' })} />
        </Box>
    )
}

const AlbumsControls = ({ dispatch }) => {
    const [newAlbumModal, setNewAlbumModal] = useState(false)
    const [newAlbumName, setNewAlbumName] = useState("")
    const { authHeader } = useContext(userContext)
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

}

const GalleryControls = ({ page, albumId, rawSelected, dispatch }) => {

    return (
        <FlexRowBox style={{ position: 'sticky', alignSelf: 'flex-stat', alignItems: 'center', width: '100%', marginLeft: 225, height: 65, zIndex: 10 }}>
            <Divider my={10} size={1} orientation='vertical' />
            <Space w={30} />
            <Box
                style={{
                    width: '100%',
                    borderRadius: "6px",
                }}
            >
                {page == "timeline" && (
                    <TimelineControls rawSelected={rawSelected} dispatch={dispatch} />
                )}

                {page == "albums" && !albumId && (
                    <AlbumsControls dispatch={dispatch} />
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
            <Tabs.List>
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

function InfiniteScroll({ mediaState, page, loading, dispatch }: { mediaState: MediaStateType, page: string, loading: boolean, dispatch: (value) => void }) {
    const { authHeader } = useContext(userContext)
    useEffect(() => { FetchData(mediaState, dispatch, authHeader).then(() => dispatch({ type: 'set_loading', loading: false })) }, [mediaState.includeRaw, authHeader, page])
    const scrollerRef = useRef(null)

    const dateMap: Map<string, Array<MediaData>> = useMemo(() => {
        let dateMap = new Map<string, Array<MediaData>>()

        if (mediaState.mediaMap.size === 0) {
            return dateMap
        }

        for (let value of mediaState.mediaMap.values()) {

            const [date, _] = value.createDate.split("T")
            if (dateMap.get(date) == null) {
                dateMap.set(date, [value])
            } else {
                dateMap.get(date).push(value)
            }
        }
        return dateMap
    }, [mediaState.mediaMap.size])

    const [dateGroups, numShownItems]: [JSX.Element[], number] = useMemo(() => {
        if (!dateMap || !scrollerRef?.current) { return [[], 0] }
        let counter = 0
        const buckets = Array.from(dateMap.keys()).map((date) => {
            const items = dateMap.get(date).filter((item) => { return mediaState.searchContent ? item.filename.toLowerCase().includes(mediaState.searchContent.toLowerCase()) : true })
            if (items.length == 0) { return null }
            counter += items.length
            return (<GalleryBucket key={date} date={date} bucketData={items} scrollerRef={scrollerRef} dispatch={dispatch} />)
        })
        return [buckets, counter]
    }, [dateMap, scrollerRef.current, mediaState.searchContent])

    return (
        <ScrollArea ref={scrollerRef} type='never' style={{ height: "100%", width: '100%', overflow: 'scroll', borderRadius: "8px" }}>
            {dateGroups}
            {!loading && dateGroups.length === 0 && (
                <NoMediaDisplay />
            )}
            {loading && (
                <Center style={{ height: '100px', width: "100%" }}>
                    <Loader color='white' bottom={10} />
                </Center>
            )}
            {!loading && (
                <Text c='white' style={{ textAlign: "center", paddingTop: "80px", paddingBottom: "10px" }}> Wow, you scrolled this whole way? </Text>
            )}
        </ScrollArea>
    )
}

const Gallery = () => {
    const [mediaState, dispatch]: [MediaStateType, React.Dispatch<any>] = useReducer(mediaReducer, {
        mediaMap: new Map<string, MediaData>(),
        albumsMap: new Map<string, AlbumData>(),
        mediaCount: 0,
        presentingHash: "",
        includeRaw: false,
        loading: true,
        scanProgress: 0,
        searchContent: "",
        blockSearchFocus: false,
        newAlbumDialogue: false
    })

    const loc = useLocation()
    const page = loc.pathname === "/" || loc.pathname === "/timeline" ? 'timeline' : "albums"
    const albumId = useParams()["*"]
    const { wsSend, lastMessage } = useWeblensSocket()

    const searchRef = useRef()
    useKeyDown(mediaState.blockSearchFocus, searchRef)
    useEffect(() => { handleWebsocket(lastMessage, dispatch) }, [lastMessage])

    return (
        <Box>
            <HeaderBar folderId={"home"} searchContent={mediaState.searchContent} dispatch={dispatch} wsSend={wsSend} page={"gallery"} searchRef={searchRef} loading={mediaState.loading} progress={mediaState.scanProgress} />
            <Presentation mediaData={mediaState.mediaMap.get(mediaState.presentingHash)} parents={null} dispatch={dispatch} />
            <FlexRowBox style={{ width: "100%", height: "100vh", paddingTop: "70px" }}>
                <GalleryControls page={page} albumId={albumId} rawSelected={mediaState.includeRaw} dispatch={dispatch} />
                <Box style={{ height: "calc(100% - 80px)", width: '99%', paddingLeft: '25px', paddingTop: "15px", position: 'absolute' }}>
                    <ViewSwitch
                        page={page}
                        timeline={
                            <InfiniteScroll mediaState={mediaState} page={page} loading={mediaState.loading} dispatch={dispatch} />
                        }
                        albums={
                            <Albums albumsMap={mediaState.albumsMap} selectedAlbum={albumId} dispatch={dispatch} />
                        }
                        albumId={albumId}
                    />
                </Box>
            </FlexRowBox>
        </Box>
    )
}

export default Gallery
