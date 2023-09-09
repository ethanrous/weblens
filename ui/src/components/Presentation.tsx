import { useEffect, useRef, useState } from 'react'

import Box from '@mui/material/Box'
import IconButton from '@mui/material/IconButton'
import CloseIcon from '@mui/icons-material/Close'
import styled from '@emotion/styled'
import CircularProgress from '@mui/material/CircularProgress'
import PlayArrowIcon from '@mui/icons-material/PlayArrow';

const StyledPhoto = styled("img")({
    width: "calc(100% - 50px)",
    height: "calc(100% - 50px)",
    position: "inherit",
    objectFit: "contain",
    objectPosition: "center",
    zIndex: 100,
})

const StyledVideo = styled("video")({
    width: "calc(100% - 50px)",
    height: "calc(100% - 50px)",
    position: "inherit",
    objectFit: "contain",
    objectPosition: "center",
    zIndex: 100,
})

type MediaData = {
    BlurHash: string
    CreateDate: string
    FileHash: string
    Filepath: string
    MediaType: {
        FileExtension: []
        FriendlyName: string
        IsRaw: boolean
        IsVideo: boolean
    }
    ThumbFilepath: string
    MediaWidth: number
    MediaHeight: number
    ThumbWidth: number
    ThumbHeight: number
    Thumbnail64: string
}

const fetchMetadata = (fileHash, setMediaData) => {
    var url = new URL(`http:localhost:3000/api/item/${fileHash}`)
    url.searchParams.append('meta', 'true')
    // url.searchParams.append('thumbnail', 'true') // Dont include thumbnail because chances are it's already cached in the browser

    fetch(url.toString()).then((res) => res.json()).then((data) => setMediaData(data))
}

const Presentation = ({ fileHash, dispatch }) => {
    const [fullResLoaded, setFullResLoaded] = useState(false)
    const [mediaData, setMediaData] = useState({} as MediaData)
    const [videoPlaying, setVideoPlaying] = useState(false)
    const vidRef = useRef(null)

    useEffect(() => {
        setFullResLoaded(false)
        fetchMetadata(fileHash, setMediaData)
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

    const handleVideoClick = () => {
        if (videoPlaying) {
            setVideoPlaying(false)
            vidRef.current.pause()
        } else {
            setVideoPlaying(true)
            vidRef.current.play()
        }
    }

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
            padding={"25px"}
            height={"calc(100vh - 50px)"}
            width={"calc(100vw - 50px)"}
            zIndex={3}
            bgcolor={"rgb(0, 0, 0, 0.92)"}
        >

            {!fullResLoaded && (
                <StyledPhoto
                    src={thumburl.toString()}
                    height={mediaData.MediaHeight}
                    width={mediaData.MediaWidth}
                />
            )}
            {!fullResLoaded && mediaData.MediaType?.IsVideo && (
                <CircularProgress style={{ position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)", zIndex: 101 }} />
            )}
            {fullResLoaded && mediaData.MediaType?.IsVideo && !videoPlaying && (
                <PlayArrowIcon style={{ width: "150px", height: "150px", position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)", zIndex: 101 }} />
            )}
            {mediaData.FileHash && !mediaData.MediaType.IsVideo && (
                <StyledPhoto
                    src={fullresurl.toString()}
                    onLoad={() => { setFullResLoaded(true) }}
                    style={{ opacity: fullResLoaded ? "100%" : "0%" }}
                />
            )}
            {mediaData.FileHash && mediaData.MediaType.IsVideo && (
                <StyledVideo
                    ref={vidRef}
                    autoPlay={true}
                    playsInline={true}
                    muted={true}
                    loop={true}
                    src={fullresurl.toString()}
                    onPlay={() => { setVideoPlaying(true); setFullResLoaded(true) }}
                    onClick={() => handleVideoClick()}
                    width={"inherit"}
                //style={{ opacity: fullResLoaded ? "100%" : "0%" }}
                />
            )}

            <IconButton
                onClick={() => dispatch({ type: 'stop_presenting' })}
                color={"inherit"}
                sx={{ display: "block", position: "absolute", top: "2.5em", left: "2.5em", cursor: "pointer", zIndex: 100 }}
            >
                <CloseIcon />
            </IconButton>
        </Box>
    )
}

export default Presentation