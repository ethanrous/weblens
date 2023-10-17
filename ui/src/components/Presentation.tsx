import { useEffect, useRef, useState, useMemo } from 'react'

import Box from '@mui/material/Box'
import IconButton from '@mui/material/IconButton'
import CloseIcon from '@mui/icons-material/Close'
import styled from '@emotion/styled'
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'
import Typography from '@mui/material/Typography'
import Crumbs from './Crumbs'

import { MediaData } from '../types/Generic'
import { MediaImage } from './PhotoContainer'

const PresentationContainer = styled(Box)({
    position: "fixed",
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    color: "white",
    top: 0,
    left: 0,
    padding: "25px",
    height: "calc(100vh - 50px)",
    width: "calc(100vw - 50px)",
    zIndex: 3,
    backgroundColor: "rgb(0, 0, 0, 0.92)",
})

const StyledMediaImage = styled(MediaImage)({
    height: "calc(100% - 40px)",
    width: "calc(100% - 40px)",
    objectFit: "contain",
    margin: "20px"
})

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
            // <Box display={"flex"} flexDirection={"column"} alignItems={"center"}>
            <InsertDriveFileIcon style={{ width: "80%", height: "80%" }} onDragOver={() => { }} />
            // </Box>
        )
    } else {
        return (
            <div style={{ height: "100%", width: "100%" }}>
                <StyledMediaImage key={`${mediaData.FileHash} thumbnail`} mediaData={mediaData} quality={"thumbnail"} lazy={false} />
                <StyledMediaImage key={`${mediaData.FileHash} fullres`} mediaData={mediaData} quality={"fullres"} lazy={false} />

            </div>
        )
    }
}

function startKeyDownHandler(dispatch) {
    console.log("AHH")
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
}

const Presentation = ({ mediaData, dispatch }: { mediaData: MediaData, dispatch: React.Dispatch<any> }) => {
    useEffect(() => {
        return startKeyDownHandler(dispatch)
    }, [])

    return (
        <PresentationContainer>
            <PresentationVisual mediaData={mediaData} />

            <IconButton
                onClick={() => dispatch({ type: 'stop_presenting' })}
                color={"inherit"}
                sx={{ display: "block", position: "absolute", top: 15, left: 15, cursor: "pointer", zIndex: 100 }}
            >
                <CloseIcon />
            </IconButton>
            <Box position={"absolute"} top={10}>
                <Crumbs path={mediaData?.Filepath} includeHome={false} navigate={() => { }} />
            </Box>
        </PresentationContainer>
    )
}

export default Presentation