import { useEffect, useReducer, useState, useMemo } from 'react'
import { useNavigate } from "react-router-dom"
import InfiniteScroll from 'react-infinite-scroll-component'
import useWebSocket from 'react-use-websocket'

import Container from '@mui/material/Container'

import HeaderBar from "./HeaderBar"
import PhotoContainer from './PhotoContainer'
import Presentation from './Presentation'

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


const computeDateString = (dateTime) => {
    let dateObj = new Date(dateTime)
    let dateStr = dateObj.toUTCString().split(" 00:00:00 GMT")[0]
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
    let mediaCards = medias.map((mediaData) => (
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

const GalleryBucket = ({
    date,
    bucketData,
    showIcons,
    dispatch
}) => {
    return (
        <Grid item >
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} showIcons={showIcons} dispatch={dispatch} />
        </Grid >

    )
}

const mediaReducer = (state, action) => {
    switch (action.type) {
        case 'add_media': {
            return {
                ...state,
                mediaList: action.mediaList,
                mediaIdMap: action.mediaIdMap,
                mediaCount: action.mediaCount,
                hasMoreMedia: action.hasMoreMedia,
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
            let changeTo
            if (state.mediaIdMap[state.presentingHash].next) {
                changeTo = state.mediaIdMap[state.presentingHash].next
            } else {
                changeTo = state.presentingHash
            }
            return {
                ...state,
                presentingHash: changeTo
            }
        }
        case 'presentation_previous': {
            let changeTo
            if (state.mediaIdMap[state.presentingHash].previous) {
                changeTo = state.mediaIdMap[state.presentingHash].previous
            } else {
                changeTo = state.presentingHash
            }
            return {
                ...state,
                presentingHash: changeTo
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

const fetchData = async (mediaState) => {
    try {
        let mediaList = [...mediaState.mediaList]
        let mediaIdMap = { ...mediaState.mediaIdMap }

        var url = new URL("http:localhost:3000/api/media")
        url.searchParams.append('limit', '100')
        url.searchParams.append('skip', mediaState.mediaCount)
        url.searchParams.append('raw', mediaState.includeRaw.toString())
        const response = await fetch(url.toString())
        const data = await response.json()

        let moreMedia: boolean
        if (data.Media != null) {
            moreMedia = data.MoreMedia

            let prevousLast
            if (mediaList.length > 0) {
                prevousLast = mediaList[mediaList.length - 1]
            } else {
                prevousLast = null
            }
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
    let navigate = useNavigate()
    const routeChange = () => {
        let path = `/upload`
        navigate(path)
    }

    return (
        <Box
            display="flex"
            flexWrap="wrap"
            flexDirection="column"
            pt="50px"
            alignContent="center"
            gap="25px"
        >
            {"No media to display"}
            <Button onClick={routeChange}>
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
                <ToggleButton value="RAW" selected={rawSelected} onChange={() => { dispatch({ type: 'toggle_raw' }) }} style={{ backgroundColor: "white" }}>
                    <RawOnIcon />
                </ToggleButton>
                <ToggleButton value="INFO" selected={showIcons} onChange={() => { dispatch({ type: 'toggle_info' }) }} style={{ backgroundColor: "white" }}>
                    <InfoIcon />
                </ToggleButton>
            </Toolbar>
        </AppBar>
    )
}

type MediaData = {
    BlurHash: string
    CreateDate: string
    FileHash: string
    Filepath: string
    MediaType: {
        FileExtension: []
        FriendlyName: string
        IsRaw: boolean
    }
    ThumbFilepath: string
    ThumbWidth: number
    ThumbHeight: number
    Thumbnail64: string
}



function StartKeybaordListener(dispatch) {
    useEffect(() => {
        const keyDownHandler = event => {
            if (event.key === 'i') {
                event.preventDefault()
                dispatch({
                    type: 'toggle_info'
                })
            }
        }

        document.addEventListener('keydown', keyDownHandler)

        return () => {
            document.removeEventListener('keydown', keyDownHandler)
        }
    }, [])
}

const Gallery = () => {
    const WS_URL = 'ws://localhost:4000/api/ws';
    const { sendMessage, lastMessage, readyState } = useWebSocket(WS_URL, {
        onOpen: () => {
            console.log('WebSocket connection established.')
        }
    })

    const { enqueueSnackbar } = useSnackbar()

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

    StartKeybaordListener(dispatch)

    const moar_data = () => {
        dispatch({ type: "set_loading", loading: true })
        fetchData(mediaState)
            .then(
                ([mediaList, mediaIdMap, hasMoreMedia]) => {
                    dispatch({
                        type: 'add_media',
                        mediaList: mediaList,
                        mediaIdMap: mediaIdMap,
                        mediaCount: mediaList.length,
                        hasMoreMedia: hasMoreMedia
                    })
                },
                () => { }
            )
        dispatch({ type: "set_loading", loading: false })
    }

    useEffect(() => {
        if (lastMessage) {
            let msgData = JSON.parse(lastMessage.data)

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

    useEffect(() => moar_data(), [mediaState.includeRaw])

    let mediaBuckets = {}
    for (const item of mediaState.mediaList) {
        var [date, _] = item.CreateDate.split("T")
        if (mediaBuckets[date] == null) {
            mediaBuckets[date] = [item]
        } else {
            mediaBuckets[date].push(item)
        }
    }

    let dateGroups
    if (mediaBuckets) {
        dateGroups = Object.keys(mediaBuckets).map((value, i) => (
            <GalleryBucket date={value} bucketData={mediaBuckets[value]} showIcons={mediaState.showIcons} dispatch={dispatch} key={value} />
        ))
    }

    if ((mediaState.mediaCount == 0) && !mediaState.loading) {
        return (
            <Container maxWidth={false} >
                <GalleryOptions rawSelected={mediaState.includeRaw} showIcons={mediaState.showIcons} dispatch={dispatch} />
                <NoMediaDisplay />
            </Container>
        )
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
                <Box sx={{ width: '100%' }}>
                    <LinearProgress />
                </Box>
            )}
            <Container maxWidth={false} style={{ display: 'inherit', flexDirection: 'row', justifyContent: "center", paddingLeft: "0px", paddingRight: mediaState.presentingHash == "" ? "55px" : "71px" }}>
                {mediaState.presentingHash != "" && (
                    <Presentation fileHash={mediaState.presentingHash} dispatch={dispatch} />
                )}

                <GalleryOptions rawSelected={mediaState.includeRaw} showIcons={mediaState.showIcons} dispatch={dispatch} />

                <InfiniteScroll
                    dataLength={mediaState.mediaCount}
                    next={moar_data}
                    children={dateGroups}
                    hasMore={mediaState.hasMoreMedia}
                    loader={<Box justifyContent={"center"} display={"flex"} padding={"50px"}><CircularProgress /></Box>}
                    endMessage={<h1 style={{ fontSize: "10px", padding: "50px", textAlign: "center" }}>You did it, you reached the end :)</h1>}
                    style={{ width: "80vw" }}
                />
            </Container>
        </Box>
    )
}

export default Gallery
