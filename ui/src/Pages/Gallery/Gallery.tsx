import { useEffect, useReducer, useMemo, memo, useRef } from 'react'

import HeaderBar from "../../components/HeaderBar"
import Presentation from '../../components/Presentation'
import GetWebsocket from '../../api/Websocket'
import { GalleryBucket } from './MediaDisplay'
import { mediaReducer, startKeybaordListener, handleScroll, moreData } from './GalleryLogic'

import Container from '@mui/material/Container'
import Button from '@mui/material/Button'
import Box from '@mui/material/Box'
import AppBar from '@mui/material/AppBar'
import Toolbar from '@mui/material/Toolbar'
import { LinearProgress } from '@mui/material'
import ToggleButton from '@mui/material/ToggleButton'
import RawOnIcon from '@mui/icons-material/RawOn'
import InfoIcon from '@mui/icons-material/Info'
import { useSnackbar } from 'notistack';
import Tooltip from '@mui/material/Tooltip';
import styled from '@emotion/styled'

const NoMediaContainer = styled(Box)({
    display: "flex",
    flexWrap: "wrap",
    flexDirection: "column",
    marginTop: "50px",
    gap: "25px",
    alignContent: "center"
})

const NoMediaDisplay = () => {
    return (
        <NoMediaContainer>
            No media to display
            <Button >
                Upload Media
            </Button>
        </NoMediaContainer>
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

const InfiniteScroll = ({ items, itemCount, loading, moreItems }: { items: any, itemCount: number, loading: boolean, moreItems: boolean }) => {
    return (
        <Box flexDirection={"column"} pb={10} width={"90vw"}>
            {items}
            {itemCount == 0 && !loading && (
                <NoMediaDisplay />
            )}
            {!moreItems && (
                <p style={{ textAlign: "center", paddingTop: "90px" }}>Wow, you scrolled this whole way?</p>
            )}
        </Box>
    )
}

const Gallery = () => {
    const { enqueueSnackbar } = useSnackbar()

    const { sendMessage, lastMessage, readyState } = GetWebsocket(enqueueSnackbar)

    const presentingRef = useRef()

    const [mediaState, dispatch] = useReducer(mediaReducer, {
        mediaMap: {},
        dateMap: {},
        mediaCount: 0,
        maxMediaCount: 100,
        hasMoreMedia: true,
        presentingHash: "",
        previousLast: "",
        presentingRef: presentingRef,
        includeRaw: false,
        showIcons: false,
        loading: false
    })

    useEffect(() => {
        moreData(mediaState, dispatch)
    }, [mediaState.maxMediaCount, mediaState.includeRaw])

    useEffect(() => {
        window.addEventListener('scroll', (e) => { handleScroll(e, dispatch) }, false)
        return startKeybaordListener(dispatch)
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
                    console.error("Got unexpected websocket message: ", msgData)
                    return
                }
            }
        }
    }, [lastMessage])

    const dateGroups = useMemo(() => {
        if (!mediaState.dateMap) {
            return []
        }
        const buckets = Object.keys(mediaState.dateMap).map((value, i) => (
            <GalleryBucket date={value} bucketData={mediaState.dateMap[value]} showIcons={mediaState.showIcons} dispatch={dispatch} key={value} />
        ))
        return buckets
    }, [{ ...mediaState.mediaMap }, mediaState.dateMap, mediaState.showIcons])

    return (
        <Box>
            <HeaderBar dispatch={dispatch} sendMessage={sendMessage} page={"gallery"} />
            {mediaState.loading && (
                <LinearProgress style={{ width: "100%", position: "absolute" }} />
            )}
            <Container maxWidth={false} style={{ display: 'flex', flexDirection: 'row', justifyContent: "center", paddingLeft: "0px", paddingRight: mediaState.presentingHash == "" ? "55px" : "71px" }}>
                {mediaState.presentingHash != "" && (
                    <Presentation mediaData={mediaState.mediaMap[mediaState.presentingHash]} dispatch={dispatch} />
                )}

                <GalleryOptions rawSelected={mediaState.includeRaw} showIcons={mediaState.showIcons} dispatch={dispatch} />
                <InfiniteScroll items={dateGroups} itemCount={mediaState.mediaCount} loading={mediaState.loading} moreItems={mediaState.hasMoreMedia} />
            </Container>
        </Box>
    )
}

export default Gallery
