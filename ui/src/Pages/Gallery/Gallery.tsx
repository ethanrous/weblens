import { useEffect, useReducer, useMemo, useRef, useContext } from 'react'

import HeaderBar from "../../components/HeaderBar"
import Presentation from '../../components/Presentation'
import { GalleryBucket } from './MediaDisplay'
import { mediaReducer, useScroll, useKeyDown } from './GalleryLogic'
import { FetchData } from '../../api/GalleryApi'
import { MediaData, MediaStateType } from '../../types/Types'
import useWeblensSocket from '../../api/Websocket'
import { StyledBreadcrumb } from '../../components/Crumbs'
import { userContext } from '../../Context'

import { Divider, Sheet, Button, Box, ToggleButtonGroup, Typography, useTheme, styled, IconButton } from '@mui/joy'
import { useNavigate } from 'react-router-dom'
import { useSnackbar } from 'notistack'
import { RawOn } from '@mui/icons-material'

// styles

const NoMediaContainer = styled(Box)({
    display: "flex",
    flexWrap: "wrap",
    flexDirection: "column",
    marginTop: "50px",
    gap: "25px",
    alignContent: "center"
})

// funcs

const NoMediaDisplay = ({ loading }: { loading: boolean }) => {
    const nav = useNavigate()
    if (loading) {
        return null
    }
    return (
        <NoMediaContainer>
            <Typography color={'primary'}>No media to display</Typography>
            <Button sx={{ border: (theme) => `1px solid ${theme.palette.divider}` }} onClick={() => nav('/files')}>
                Upload Media
            </Button>
        </NoMediaContainer>
    )
}

const Sidebar = ({ rawSelected, itemCount, dispatch }) => {
    const theme = useTheme()
    return (
        <Box display={"flex"} flexDirection={"column"} alignItems={"center"} position={"sticky"} alignSelf={'flex-start'} top={'100px'} padding={'15px'} >
            <Sheet
                sx={{
                    width: "max-content",
                    display: 'flex',
                    flexWrap: 'wrap',
                    flexDirection: 'column',
                    alignItems: 'center',
                    padding: "15px",
                    borderRadius: "8px",
                    backgroundColor: theme.colorSchemes.dark.palette.primary.solidBg,
                }}
            >
                <StyledBreadcrumb label={itemCount} tooltipText={`Showing ${itemCount} images`} />
                <IconButton
                    variant='solid'
                    onClick={() => { dispatch({ type: 'toggle_raw' }) }}
                    sx={{
                        margin: '8px'
                    }}
                >
                    <RawOn sx={{ color: theme.colorSchemes.dark.palette.common.white }} />
                </IconButton>
                <Divider orientation="horizontal" />
            </Sheet>
        </Box>
    )
}

const InfiniteScroll = ({ items, itemCount, loading, moreItems }: { items: any, itemCount: number, loading: boolean, moreItems: boolean }) => {
    return (
        <Box flexDirection={"column"} width={"90vw"} >
            {items}
            {itemCount == 0 && (
                <NoMediaDisplay loading={loading} />
            )}
            {!moreItems && (
                <Typography color={'neutral'} style={{ textAlign: "center", paddingTop: "80px", paddingBottom: "10px" }}> Wow, you scrolled this whole way? </Typography>
            )}
        </Box>
    )
}

const handleWebsocket = (lastMessage, dispatch, enqueueSnackbar) => {
    if (lastMessage) {
        const msgData = JSON.parse(lastMessage.data)
        switch (msgData["type"]) {
            case "item_update": {
                return
            }
            case "item_deleted": {
                dispatch({ type: "delete_from_map", item: msgData["content"].hash })
                return
            }
            case "scan_directory_progress": {
                dispatch({ type: "set_scan_progress", progress: ((1 - (msgData["remainingTasks"] / msgData["totalTasks"])) * 100) })
                return
            }
            case "finished": {
                dispatch({ type: "set_loading", loading: false })
                return
            }
            case "error": {
                enqueueSnackbar(msgData["error"], { variant: "error" })
                return
            }
            default: {
                console.error("Got unexpected websocket message: ", msgData)
                return
            }
        }
    }
}

const Gallery = () => {
    document.documentElement.style.overflow = "visible"
    const { authHeader, userInfo } = useContext(userContext)

    const [mediaState, dispatch]: [MediaStateType, React.Dispatch<any>] = useReducer(mediaReducer, {
        mediaMap: new Map<string, MediaData>(),
        mediaCount: 0,
        maxMediaCount: 100,
        hasMoreMedia: true,
        presentingHash: "",
        previousLast: "",
        includeRaw: false,
        loading: false,
        scanProgress: 0,
        searchContent: ""
    })

    const nav = useNavigate()
    const { enqueueSnackbar } = useSnackbar()
    const { wsSend, lastMessage, readyState } = useWeblensSocket()

    const searchRef = useRef()
    useKeyDown(searchRef)
    useScroll(mediaState.hasMoreMedia, dispatch)
    useEffect(() => { FetchData(mediaState, dispatch, nav, authHeader) }, [mediaState.maxMediaCount, mediaState.includeRaw, authHeader])
    useEffect(() => { handleWebsocket(lastMessage, dispatch, enqueueSnackbar) }, [lastMessage])

    const dateMap: Map<string, Array<MediaData>> = useMemo(() => {
        let dateMap = new Map<string, Array<MediaData>>()

        if (mediaState.mediaMap.size === 0) {
            return dateMap
        }

        for (let value of mediaState.mediaMap.values()) {

            const [date, _] = value.CreateDate.split("T")
            if (dateMap.get(date) == null) {
                dateMap.set(date, [value])
            } else {
                dateMap.get(date).push(value)
            }
        }
        return dateMap
    }, [mediaState.mediaMap.size])

    const [dateGroups, numShownItems]: [JSX.Element[], number] = useMemo(() => {
        if (!dateMap) { return [[], 0] }
        let counter = 0
        const buckets = Array.from(dateMap.keys()).map((date) => {
            const items = dateMap.get(date).filter((item) => { return mediaState.searchContent ? item.Filepath.toLowerCase().includes(mediaState.searchContent.toLowerCase()) : true })
            if (items.length == 0) { return null }
            counter += items.length
            return (<GalleryBucket key={date} date={date} bucketData={items} dispatch={dispatch} />)
        })

        return [buckets, counter]
    }, [dateMap, mediaState.searchContent])

    useEffect(() => {
        wsSend(JSON.stringify({ type: "subscribe", content: { path: "/", recursive: true }, error: null }))
    }, [])

    return (
        <Box>
            <HeaderBar path={"/"} searchContent={mediaState.searchContent} dispatch={dispatch} wsSend={wsSend} page={"gallery"} searchRef={searchRef} loading={mediaState.loading} progress={mediaState.scanProgress} />

            <Box display={'flex'} flexDirection={'row'} width={"100%"} minHeight={"50vh"} justifyContent={'space-evenly'} pt={"70px"}>
                <Sidebar rawSelected={mediaState.includeRaw} itemCount={numShownItems} dispatch={dispatch} />
                <Box width={'90%'} style={{ justifyContent: "center", paddingTop: "25px", paddingRight: mediaState.presentingHash == "" ? "30px" : "46px" }}>
                    <InfiniteScroll items={dateGroups} itemCount={mediaState.mediaCount} loading={mediaState.loading} moreItems={mediaState.hasMoreMedia} />
                    <Presentation mediaData={mediaState.mediaMap.get(mediaState.presentingHash)} dispatch={dispatch} />
                </Box>
            </Box>
        </Box>
    )
}

export default Gallery
