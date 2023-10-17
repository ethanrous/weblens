import { useEffect, useReducer, useMemo, memo, useRef, useState } from 'react'

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
import { MediaData } from '../../types/Generic'

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
    const { wsSend, lastMessage, readyState } = GetWebsocket(enqueueSnackbar)

    const [mediaState, dispatch] = useReducer(mediaReducer, {
        mediaMap: {},
        mediaCount: 0,
        maxMediaCount: 100,
        hasMoreMedia: true,
        presentingHash: "",
        previousLast: "",
        includeRaw: false,
        showIcons: false,
        loading: false,
        scanProgress: 0,
        searchContent: ""
    })

    useEffect(() => {
        moreData(mediaState, dispatch)
    }, [mediaState.maxMediaCount, mediaState.includeRaw])

    useEffect(() => {
        window.addEventListener('scroll', (_) => { handleScroll(dispatch) }, false)
        //return startKeybaordListener(dispatch)
    }, [])

    useEffect(() => {
        if (lastMessage) {
            const msgData = JSON.parse(lastMessage.data)
            switch (msgData["type"]) {
                case "new_items": {
                    dispatch({ type: "append_items", items: msgData["content"] })
                    return
                }
                case "scan_directory_progress": {
                    dispatch({ type: "set_scan_progress", progress: ((1 - (msgData["remainingTasks"] / msgData["totalTasks"])) * 100) })
                    console.log(msgData["remainingTasks"], "/", msgData["totalTasks"])
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

    const dateMap = useMemo(() => {
        let dateMap = {}

        for (let value of Object.values<MediaData>(mediaState.mediaMap)) {

            const [date, _] = value.CreateDate.split("T")
            if (dateMap[date] == null) {
                dateMap[date] = [value]
            } else {
                dateMap[date].push(value)
            }
        }

        return dateMap
    }, [mediaState.mediaMap])

    const c = useMemo(() => {
        handleScroll(dispatch)
        for (let value of Object.values<MediaData>(mediaState.mediaMap)) {
            var count = 0
            if (!value.Filepath.includes(mediaState.searchContent)) {
                value.Display = false
            } else {
                count += 1
                value.Display = true
            }
        }
        return count
    }, [dateMap, mediaState.searchContent])

    const dateGroups = useMemo(() => {
        if (!dateMap) {
            return []
        }
        const buckets = Object.keys(dateMap).map((value, i) => (
            <GalleryBucket key={value} date={value} bucketData={dateMap[value]} showIcons={mediaState.showIcons} dispatch={dispatch} />
        ))
        return buckets
    }, [dateMap, mediaState.showIcons, mediaState.searchContent, c])

    return (
        <Box>
            <HeaderBar dispatch={dispatch} wsSend={wsSend} page={"gallery"} />
            {mediaState.loading && (
                <Box sx={{ width: '100%' }}>
                    {mediaState.scanProgress == 0 && (
                        <LinearProgress />
                    )}
                    {mediaState.scanProgress != 0 && (
                        <Box sx={{ width: '100%' }}>
                            <LinearProgress variant="determinate" value={mediaState.scanProgress} />
                            <p style={{ position: "absolute", left: "6vw" }}>Syncing filesystem with database...</p>
                        </Box>
                    )}
                </Box>
            )}
            <Container maxWidth={false} style={{ display: 'flex', flexDirection: 'row', justifyContent: "center", paddingLeft: "0px", paddingTop: "25px", paddingRight: mediaState.presentingHash == "" ? "55px" : "71px" }}>
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
