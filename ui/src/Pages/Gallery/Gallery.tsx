import { useEffect, useReducer, useMemo, memo, useRef, useState } from 'react'

import HeaderBar from "../../components/HeaderBar"
import Presentation from '../../components/Presentation'
import { GalleryBucket } from './MediaDisplay'
import { mediaReducer, handleScroll } from './GalleryLogic'
import { fetchData } from '../../api/GalleryApi'

import FormatBoldIcon from '@mui/icons-material/FormatBold';
import FormatItalicIcon from '@mui/icons-material/FormatItalic';
import Button from '@mui/material/Button'
import Box from '@mui/material/Box'
import ToggleButton from '@mui/material/ToggleButton'
import RawOnIcon from '@mui/icons-material/RawOn'

import Tooltip from '@mui/material/Tooltip';
import { styled } from '@mui/material/styles';
import { MediaData, MediaStateType } from '../../types/Types'
import { Divider, Paper, ToggleButtonGroup, Typography, useTheme } from '@mui/material'
import { useCookies } from 'react-cookie'
import { useNavigate } from 'react-router-dom'
import { useSnackbar } from 'notistack'
import useWeblensSocket from '../../api/Websocket'
import RawOn from '@mui/icons-material/RawOn'
import { StyledBreadcrumb } from '../../components/Crumbs'
import { ThemeContext } from '@emotion/react'

// styles

const NoMediaContainer = styled(Box)({
    display: "flex",
    flexWrap: "wrap",
    flexDirection: "column",
    marginTop: "50px",
    gap: "25px",
    alignContent: "center"
})

const StyledToggleButtonGroup = styled(ToggleButtonGroup)(({ theme }) => ({
    '& .MuiToggleButtonGroup-grouped': {
        margin: theme.spacing(0.5),

        '&.Mui-disabled': {
            border: 0,
        },
    },
}));

// funcs

const NoMediaDisplay = ({ loading }: { loading: boolean }) => {
    const theme = useTheme()
    const nav = useNavigate()
    if (loading) {
        return null
    }
    return (
        <NoMediaContainer>
            <Typography color={theme.palette.primary.main}>No media to display</Typography>
            <Button sx={{ border: (theme) => `1px solid ${theme.palette.divider}` }} onClick={() => nav('/files')}>
                Upload Media
            </Button>
        </NoMediaContainer>
    )
}

const Sidebar = ({ rawSelected, itemCount, dispatch }) => {
    return (
        <Box display={"flex"} flexDirection={"column"} alignItems={"center"} position={"sticky"} alignSelf={'flex-start'} top={'75px'} padding={'25px'} >
            <Paper
                elevation={3}
                sx={{
                    width: "max-content",
                    display: 'flex',
                    flexWrap: 'wrap',
                    flexDirection: 'column',
                    padding: "4px",
                    backgroundColor: (theme) => { return theme.palette.background.default },
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                }}
            >
                <StyledBreadcrumb label={itemCount} tooltipText={`Showing ${itemCount} images`} />
                <ToggleButton value="RAW" selected={rawSelected} onChange={(e, val) => { console.log(val); dispatch({ type: 'toggle_raw' }) }} sx={{ m: 4 }}>
                    <RawOn sx={{ color: (theme) => { return theme.palette.primary.main } }} />
                </ToggleButton>
                <Divider flexItem orientation="horizontal" variant='middle' />
                {/* <StyledToggleButtonGroup orientation="vertical" value={"bold"} onChange={() => { }} sx={{ m: 4 }}>
                    <ToggleButton value="bold" aria-label="bold">
                        <FormatBoldIcon />
                    </ToggleButton>
                    <ToggleButton value="italic" aria-label="italic">
                        <FormatItalicIcon />
                    </ToggleButton>
                </StyledToggleButtonGroup> */}
            </Paper>
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
                <Typography color={'primary'} style={{ textAlign: "center", paddingTop: "90px" }}> Wow, you scrolled this whole way? </Typography>
            )}
        </Box>
    )
}

const handleWebsocket = (lastMessage, dispatch, enqueueSnackbar) => {
    if (lastMessage) {
        const msgData = JSON.parse(lastMessage.data)
        switch (msgData["type"]) {
            case "scan_directory_progress": {
                dispatch({ type: "set_scan_progress", progress: ((1 - (msgData["remainingTasks"] / msgData["totalTasks"])) * 100) })
                console.log(msgData["remainingTasks"], "/", msgData["totalTasks"])
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

const useKeyDown = (dispatch, searchRef) => {

    const onKeyDown = (event) => {
        if (!event.metaKey && event.which >= 65 && event.which <= 90) {
            searchRef.current.children[0].focus()
        }
    };
    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        };
    }, [onKeyDown])
}

const useScroll = (hasMoreMedia, dispatch) => {
    const onScrollEvent = (_) => {
        if (hasMoreMedia) { handleScroll(dispatch) }
    }
    useEffect(() => {
        window.addEventListener('scroll', onScrollEvent)
        return () => {
            window.removeEventListener('scroll', onScrollEvent)
        }
    }, [onScrollEvent])
}

const Gallery = () => {
    document.documentElement.style.overflow = "visible"

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
    const [cookies, setCookie, removeCookie] = useCookies(['weblens-username', 'weblens-login-token']);
    const { wsSend, lastMessage, readyState } = useWeblensSocket()

    const searchRef = useRef()
    useKeyDown(dispatch, searchRef)
    useScroll(mediaState.hasMoreMedia, dispatch)
    useEffect(() => { fetchData(mediaState, dispatch, nav, cookies['weblens-username'], cookies["weblens-login-token"]) }, [mediaState.maxMediaCount, mediaState.includeRaw])
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

    return (
        <Box>
            <HeaderBar dispatch={dispatch} wsSend={wsSend} page={"gallery"} searchRef={searchRef} loading={mediaState.loading} progress={mediaState.scanProgress} />

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
