import { useEffect, useReducer, useMemo, memo } from 'react'
import { useNavigate } from "react-router-dom"
import InfiniteScroll from 'react-infinite-scroll-component'

import Container from '@mui/material/Container'

import HeaderBar from "./HeaderBar"
import PhotoContainer from './PhotoContainer'
import Presentation from './Presentation'
import GetWebsocket from './Websocket'

import styled from '@emotion/styled'

import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Box from '@mui/material/Box'
import AppBar from '@mui/material/AppBar'
import Toolbar from '@mui/material/Toolbar'
import CircularProgress from '@mui/material/CircularProgress'
import { LinearProgress } from '@mui/material'
import ToggleButton from '@mui/material/ToggleButton'
import RawOnIcon from '@mui/icons-material/RawOn'
import InfoIcon from '@mui/icons-material/Info'
import { useSnackbar } from 'notistack';
import Tooltip from '@mui/material/Tooltip';


const computeDateString = (dateTime) => {
    const dateObj = new Date(dateTime)
    const dateStr = dateObj.toUTCString().split(" 00:00:00 GMT")[0]
    return dateStr
}

const DateWrapper = ({ dateTime }) => {
    const dateString = useMemo(() => computeDateString(dateTime), [dateTime])
    return (
        <Box
            key={`${dateString} title`}
            component="h3"
            fontSize={25}
            pl={1}
            fontFamily={"Roboto,RobotoDraft,Helvetica,Arial,sans-serif"}
            marginBottom={1}
        >
            {dateString}
        </Box>
    )
}

const BlankCard = styled("div")({
    height: '250px',
    flexGrow: 999999
})


const BucketCards = ({ medias, showIcons, dispatch }) => {
    const mediaCards = medias.map((mediaData) => (
        <PhotoContainer
            key={mediaData.FileHash}
            mediaData={mediaData}
            showIcons={showIcons}
            dispatch={dispatch}
        />
    ))

    return (
        <Grid container
            display="flex"
            flexDirection="row"
            flexWrap="wrap"
            justifyContent="flex-start"
        >
            {mediaCards}
            <BlankCard />
        </Grid>
    )
}

type GalleryBucketProps = {
    date: string
    bucketData: []
    showIcons: boolean
    dispatch: React.Dispatch<any>
}

const GalleryBucket = memo(function GalleryBucket({
    date,
    bucketData,
    showIcons,
    dispatch
}: GalleryBucketProps) {
    return (
        <Grid item >
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} showIcons={showIcons} dispatch={dispatch} />
        </Grid >

    )
})

const mediaReducer = (state, action) => {
    switch (action.type) {
        case 'add_media': {

            const datemap = {}
            for (const item of action.mediaList) {
                const [date, _] = item.CreateDate.split("T")
                if (datemap[date] == null) {
                    datemap[date] = [item]
                } else {
                    datemap[date].push(item)
                }
            }

            return {
                ...state,
                mediaList: action.mediaList,
                mediaIdMap: action.mediaIdMap,
                mediaCount: action.mediaCount,
                hasMoreMedia: action.hasMoreMedia,
                dateMap: datemap
            }
        }
        case 'set_loading': {
            return {
                ...state,
                loading: action.loading
            }
        }
        case 'toggle_info': {
            return {
                ...state,
                showIcons: !state.showIcons
            }
        }
        case 'toggle_raw': {
            window.scrollTo({
                top: 0,
                behavior: "smooth"
            })

            return {
                ...state,
                mediaList: [],
                mediaIdMap: {},
                datemap: {},
                mediaCount: 0,
                includeRaw: !state.includeRaw
            }
        }
        case 'set_presentation': {
            document.documentElement.style.overflow = "hidden"

            return {
                ...state,
                presentingHash: action.presentingHash
            }
        }
        case 'presentation_next': {
            return {
                ...state,
                presentingHash: state.mediaIdMap[state.presentingHash].next ? state.mediaIdMap[state.presentingHash].next : state.presentingHash
            }
        }
        case 'presentation_previous': {
            return {
                ...state,
                presentingHash: state.mediaIdMap[state.presentingHash].previous ? state.mediaIdMap[state.presentingHash].previous : state.presentingHash
            }
        }
        case 'stop_presenting': {
            document.documentElement.style.overflow = "visible"
            return {
                ...state,
                presentingHash: ""
            }
        }
    }
}

const fetchData = async (mediaList, mediaIdMap, mediaCount, includeRaw) => {
    try {
        //let mediaList = [...mediaState.mediaList]
        //let mediaIdMap = { ...mediaState.mediaIdMap }

        const url = new URL("http:localhost:3000/api/media")
        url.searchParams.append('limit', '100')
        url.searchParams.append('skip', mediaCount)
        url.searchParams.append('raw', includeRaw.toString())
        const response = await fetch(url.toString())
        const data = await response.json()

        let moreMedia: boolean
        if (data.Media != null) {
            moreMedia = data.MoreMedia

            const prevousLast = mediaList.length > 0 ? mediaList[mediaList.length - 1] : null
            mediaList.push(...data.Media)
            for (const [index, value] of data.Media.entries()) {

                // This is the first media in this fetch, and no prior media exists
                if (index === 0 && prevousLast == null) {
                    mediaIdMap[value.FileHash] = {
                        previous: null,
                        next: null
                    }

                    // This is the first media in this fetch, but prior media does exist
                } else if (index === 0) {
                    mediaIdMap[value.FileHash] = {
                        previous: prevousLast.FileHash,
                        next: null
                    }
                    mediaIdMap[prevousLast.FileHash].next = value.FileHash

                    // Not the first media in this fetch
                } else {
                    mediaIdMap[value.FileHash] = {
                        previous: data.Media[index - 1].FileHash,
                        next: null
                    }
                    mediaIdMap[data.Media[index - 1].FileHash].next = value.FileHash
                }

            }
        } else {
            moreMedia = false
        }

        return [mediaList, mediaIdMap, moreMedia]

    } catch (error) {
        console.log(error)
        throw new Error("Error fetching data")
    }
}

const NoMediaDisplay = () => {
    return (
        <Box
            display="flex"
            flexWrap="wrap"
            flexDirection="column"
            mt="50px"
            alignContent="center"
            gap="25px"
        >
            {"No media to display"}
            <Button >
                Upload Media
            </Button>
        </Box>
    )
}

const GalleryOptions = ({ rawSelected, showIcons, dispatch }) => {

    return (
        <AppBar
            position="sticky"
            color='transparent'
            sx={{
                width: 'fit-content',
                height: 'fit-content',
                padding: "25px",
                boxShadow: 0,
                zIndex: 2
            }}
        >
            <Toolbar style={{ padding: 0, flexDirection: 'column' }}>
                <Tooltip title={"Toggle RAW Images"} placement={"right"}>
                <ToggleButton value="RAW" selected={rawSelected} onChange={() => { dispatch({ type: 'toggle_raw' }) }} style={{ backgroundColor: "white" }}>
                    <RawOnIcon />
                </ToggleButton>
                </Tooltip>
                <Tooltip title={"Toggle Media Info"} placement={"right"}>

                <ToggleButton value="INFO" selected={showIcons} onChange={() => { dispatch({ type: 'toggle_info' }) }} style={{ backgroundColor: "white" }}>
                    <InfoIcon />
                </ToggleButton>
                </Tooltip>
            </Toolbar>
        </AppBar>
    )
}

function StartKeybaordListener(dispatch) {

    const keyDownHandler = event => {
        if (event.key === 'i') {
            event.preventDefault()
            dispatch({
                type: 'toggle_info'
            })
        }
    }

    document.addEventListener('keydown', keyDownHandler)
    //return () => {
    //    document.removeEventListener('keydown', keyDownHandler)
    //}
}

const moreData = (mediaState, dispatch) => {
    //if (mediaState.loading) {
    //    return
    //}
    dispatch({ type: "set_loading", loading: true })
    fetchData(mediaState.mediaList, mediaState.mediaIdMap, mediaState.mediaCount, mediaState.includeRaw)
        .then(
            ([mediaList, mediaIdMap, hasMoreMedia]) => {
                dispatch({
                    type: 'add_media',
                    mediaList: mediaList,
                    mediaIdMap: mediaIdMap,
                    mediaCount: mediaList.length,
                    hasMoreMedia: hasMoreMedia
                })
                dispatch({ type: "set_loading", loading: false })
            },
            () => { }
        )
}

const Gallery = () => {
    const { enqueueSnackbar } = useSnackbar()

    const { sendMessage, lastMessage, readyState } = GetWebsocket(enqueueSnackbar)

    const [mediaState, dispatch] = useReducer(mediaReducer, {
        mediaList: [],
        mediaIdMap: {},
        mediaCount: 0,
        presentingHash: "",
        includeRaw: false,
        showIcons: false,
        hasMoreMedia: true,
        loading: false
    })

    useEffect(() => {
        StartKeybaordListener(dispatch)
        window.addEventListener('scroll', handleScroll, true)
    }, [])

    useEffect(() => {
        if (lastMessage) {
            const msgData = JSON.parse(lastMessage.data)
            switch (msgData["type"]) {
                case "new_items": {
                    dispatch({ type: "append_items", items: msgData["content"] })
                    return
                }
                case "finished": {
                    dispatch({ type: "set_loading", loading: false })
                    return
                }
                case "refresh": {
                    //GetDirectoryData(path, dispatch)
                    return
                }
                case "error": {
                    enqueueSnackbar(msgData["error"], { variant: "error" })
                    return
                }
                default: {
                    console.log("I dunno man")
                    console.log(msgData)
                    return
                }
            }
        }
    }, [lastMessage])

    useEffect(() => { console.log("HERE"); moreData(mediaState, dispatch) }, [mediaState.includeRaw])

    const dateGroups = useMemo(() => {
        if (!mediaState.dateMap) {
            return []
        }
        const buckets = Object.keys(mediaState.dateMap).map((value, i) => (
            <GalleryBucket date={value} bucketData={mediaState.dateMap[value]} showIcons={mediaState.showIcons} dispatch={dispatch} key={value} />
        ))
        return buckets
    }, [mediaState.dateMap, mediaState.showIcons])

    console.log(mediaState.mediaCount)

    const handleScroll = (e) => {

        if (!mediaState.loading && document.documentElement.scrollHeight - (document.documentElement.scrollTop + window.innerHeight) < 300 && mediaState.hasMoreMedia) {
            console.log("ERE")
            moreData(mediaState, dispatch)
        }
    }

    return (
        <Box
            sx={{
                display: "flex",
                flexDirection: 'column',
            }}
        >
            <HeaderBar dispatch={dispatch} sendMessage={sendMessage} page={"gallery"} />
            {mediaState.loading && (
                <Box sx={{ width: '100%' }} position={"absolute"}>
                    <LinearProgress />
                </Box>
            )}
            <Container maxWidth={false} style={{ display: 'flex', flexDirection: 'row', justifyContent: "center", paddingLeft: "0px", paddingRight: mediaState.presentingHash == "" ? "55px" : "71px" }}>
                {mediaState.presentingHash != "" && (
                    <Presentation fileHash={mediaState.presentingHash} dispatch={dispatch} />
                )}

                <GalleryOptions rawSelected={mediaState.includeRaw} showIcons={mediaState.showIcons} dispatch={dispatch} />

                <Box flexDirection={"column"} pb={10} width={"90vw"}>
                    {dateGroups}
                    {mediaState.mediaCount == 0 && !mediaState.loading && (
                        <NoMediaDisplay />
                    )}
                    {!mediaState.loading && mediaState.mediaCount != 0 && (
                        <p style={{ textAlign: "center", paddingTop: "90px" }}>Wow, you scrolled this whole way?</p>
                    )}
                </Box>

            </Container>
        </Box>
    )
}

export default Gallery
