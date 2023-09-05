import { useEffect, useReducer, useState, useMemo } from 'react'
import InfiniteScroll from 'react-infinite-scroll-component'

import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Box from '@mui/material/Box'
import Container from '@mui/material/Container'
import { useNavigate } from "react-router-dom"

import PhotoContainer from './PhotoContainer'
import styled from '@emotion/styled'

import RawOnIcon from '@mui/icons-material/RawOn'
import IconButton from '@mui/material/IconButton'
import Toolbar from '@mui/material/Toolbar'
import AppBar from '@mui/material/AppBar'
import CircularProgress from '@mui/material/CircularProgress'
import ToggleButton from '@mui/material/ToggleButton'
import InfoIcon from '@mui/icons-material/Info'
import CloseIcon from '@mui/icons-material/Close'

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

const fetchData = async (mediaState, setIsLoading, setError) => {
    setIsLoading(true)
    setError(null)

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

        setIsLoading(false)
        return [mediaList, mediaIdMap, moreMedia]

    } catch (error) {
        setIsLoading(false)
        console.log(error)
        setError(error)
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

const StyledPhoto = styled("img")({
    width: "calc(100% - 20px)",
    height: "calc(100% - 20px)",
    position: "inherit",
    objectFit: "contain",
    objectPosition: "center",
    zIndex: 100,
})

const Presentation = ({ fileHash, dispatch }) => {
    const [fullResLoaded, setFullResLoaded] = useState(false)
    useEffect(() => {
        setFullResLoaded(false)
    }, [fileHash])

    useEffect(() => {
        const keyDownHandler = event => {
            if (event.key === 'Escape') {
                event.preventDefault()
                dispatch({ type: 'stop_presenting' })
            }
            if (event.key === 'ArrowLeft') {
                event.preventDefault()
                dispatch({ type: 'presentation_previous' })
            }
            if (event.key === 'ArrowRight') {
                event.preventDefault()
                dispatch({ type: 'presentation_next' })
            }
        }
        document.addEventListener('keydown', keyDownHandler)
        return () => {
            document.removeEventListener('keydown', keyDownHandler)
        }
    }, [])

    //if (mediaMeta == null) {
    //    return
    //}

    var thumburl = new URL(`http:localhost:3000/api/item/${fileHash}`)
    thumburl.searchParams.append('thumbnail', 'true')

    var fullresurl = new URL(`http:localhost:3000/api/item/${fileHash}`)
    fullresurl.searchParams.append('fullres', 'true')

    return (
        <Box
            position={"fixed"}
            color={"white"}
            top={0}
            left={0}
            padding={"10px"}
            height={"calc(100vh - 20px)"}
            width={"calc(100vw - 20px)"}
            zIndex={3}
            onClick={() => dispatch({ type: 'stop_presenting' })}
            bgcolor={"rgb(0, 0, 0, 0.92)"}
        >

            {!fullResLoaded && (
                <StyledPhoto
                    src={thumburl.toString()}
                />
            )}
            <StyledPhoto
                src={fullresurl.toString()}
                onLoad={() => setFullResLoaded(true)}
                style={{ opacity: fullResLoaded ? "100%" : "0%" }}
            />

            <IconButton
                onClick={() => dispatch({ type: 'stop_presenting' })}
                color={"inherit"}
                sx={{ display: "block", position: "absolute", top: "1em", left: "1em", cursor: "pointer", zIndex: 100 }}
            >
                <CloseIcon />
            </IconButton>
        </Box>
    )
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
    document.documentElement.style.scrollBehavior = "smooth"

    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState(null)

    const [mediaState, dispatch] = useReducer(mediaReducer, {
        mediaList: [],
        mediaIdMap: {},
        mediaCount: 0,
        presentingHash: "",
        includeRaw: false,
        showIcons: false,
        hasMoreMedia: true
    })

    StartKeybaordListener(dispatch)

    const moar_data = () => {

        fetchData(mediaState, setIsLoading, setError)
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
    }

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

    if ((mediaState.mediaCount == 0) && !isLoading) {
        return (
            <Container maxWidth={false} >
                <GalleryOptions rawSelected={mediaState.includeRaw} showIcons={mediaState.showIcons} dispatch={dispatch} />
                <NoMediaDisplay />
            </Container>
        )
    }

    return (
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
    )
}

export default Gallery
