import { useEffect, useState } from 'react'

import { Close } from '@mui/icons-material'
import { styled, Button } from '@mui/joy'
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'

import { MediaData, itemData } from '../types/Types'
import { MediaImage } from './PhotoContainer'
import { useNavigate } from 'react-router-dom'
import { FlexColumnBox } from '../Pages/FileBrowser/FilebrowserStyles'
import { Box } from '@mantine/core'

const PresentationContainer = ({ ...extra }) => {
    return (
        <Box
            {...extra}
            style={{
                position: "fixed",
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                top: 0,
                left: 0,
                padding: "25px",
                height: "100%",
                width: "100%",
                zIndex: 100,
                backgroundColor: "rgb(0, 0, 0, 0.92)",
                backdropFilter: "blur(4px)"
            }}
        />
    )
}

const StyledMediaImage = ({ mediaData, quality, lazy }) => {

    return (
        <Box pos={'absolute'} style={{ height: "100%", width: "100%" }} onClick={(e) => e.stopPropagation()}>
            <MediaImage mediaData={mediaData} quality={quality} lazy={lazy} imgStyle={{ height: "calc(100% - 10px)", width: "calc(100% - 10px)", position: 'absolute', objectFit: "contain" }} />
        </Box>
    )
}

const PresentationVisual = ({ mediaData }) => {
    if (!mediaData) {
        return
    }
    else if (mediaData.MediaType.IsVideo) {
        return (
            <StyledMediaImage key={`${mediaData.FileHash} thumbnail`} mediaData={mediaData} quality={"thumbnail"} lazy={false} />
        )
    } else if (mediaData.MediaType?.FriendlyName == "File") {
        return (
            <InsertDriveFileIcon style={{ width: "80%", height: "80%" }} onDragOver={() => { }} />
        )
    } else {
        return (
            <FlexColumnBox style={{ height: "100%", width: "100%" }}>
                {/* <MediaImage mediaData={mediaData} quality={"thumbnail"} lazy={false} imgStyle={{ objectFit: "contain" }} /> */}
                <MediaImage mediaData={mediaData} quality={"fullres"} lazy={false} imgStyle={{ objectFit: "contain" }} />
                {/* <StyledMediaImage key={`${mediaData.FileHash} thumbnail`} mediaData={mediaData} quality={"thumbnail"} lazy={false} /> */}
                {/* <StyledMediaImage key={`${mediaData.FileHash} fullres`} mediaData={mediaData} quality={"fullres"} lazy={false} /> */}
            </FlexColumnBox>
        )
    }
}

function useKeyDown(mediaData, dispatch) {
    const keyDownHandler = event => {
        if (!mediaData) {
            return
        }
        else if (event.key === 'Escape') {
            event.preventDefault()
            dispatch({ type: 'stop_presenting' })
        }
        else if (event.key === 'ArrowLeft') {
            event.preventDefault()
            dispatch({ type: 'presentation_previous' })
        }
        else if (event.key === 'ArrowRight') {
            event.preventDefault()
            dispatch({ type: 'presentation_next' })
        }
        else if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
            event.preventDefault()
        }
    }
    useEffect(() => {
        window.addEventListener('keydown', keyDownHandler)
        return () => {
            window.removeEventListener('keydown', keyDownHandler)
        }
    }, [keyDownHandler])
}

function handleTimeout(to, setTo, setGuiShown) {
    if (to) {
        clearTimeout(to)
    }
    setTo(setTimeout(() => setGuiShown(false), 1000))
}

const Presentation = ({ mediaData, parents, dispatch }: { mediaData: MediaData, parents: itemData[], dispatch: React.Dispatch<any> }) => {
    useKeyDown(mediaData, dispatch)

    const [to, setTo] = useState(null)
    const [guiShown, setGuiShown] = useState(false)

    if (!mediaData || !mediaData.MediaType.IsDisplayable) {
        return null
    }

    return (
        <PresentationContainer onMouseMove={(_) => { setGuiShown(true); handleTimeout(to, setTo, setGuiShown) }}>
            <PresentationVisual mediaData={mediaData} />
            <Button
                onClick={() => dispatch({ type: 'stop_presenting' })}
                sx={{ display: "flex", justifyContent: 'center', position: "absolute", top: guiShown ? 15 : -100, left: 15, cursor: "pointer", zIndex: 100, height: '10px', width: '10px', padding: 2 }}
            >
                <Close />
            </Button>
        </PresentationContainer>
    )
}

export default Presentation