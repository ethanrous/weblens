import { useEffect } from 'react'

import { Close } from '@mui/icons-material'
import { styled, Button, Box } from '@mui/joy'
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'
import Crumbs from './Crumbs'

import { MediaData } from '../types/Types'
import { MediaImage } from './PhotoContainer'
import { useNavigate } from 'react-router-dom'

const PresentationContainer = styled(Box)({
    position: "fixed",
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    top: 0,
    left: 0,
    padding: "25px",
    height: "calc(100vh - 50px)",
    width: "calc(100vw - 50px)",
    zIndex: 100,
    backgroundColor: "rgb(0, 0, 0, 0.92)",
    backdropFilter: "blur(4px)"
})

const StyledMediaImage = styled(MediaImage)({
    height: "calc(100% - 10px)",
    width: "calc(100% - 10px)",
    position: 'absolute',
    objectFit: "contain",
    margin: "5px"
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
            <InsertDriveFileIcon style={{ width: "80%", height: "80%" }} onDragOver={() => { }} />
        )
    } else {
        return (
            <div style={{ height: "100%", width: "100%", display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                <StyledMediaImage key={`${mediaData.FileHash} thumbnail`} mediaData={mediaData} quality={"thumbnail"} lazy={false} />
                <StyledMediaImage key={`${mediaData.FileHash} fullres`} mediaData={mediaData} quality={"fullres"} lazy={false} />
            </div>
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

const Presentation = ({ mediaData, dispatch }: { mediaData: MediaData, dispatch: React.Dispatch<any> }) => {
    useKeyDown(mediaData, dispatch)

    const navigate = useNavigate()

    if (!mediaData) {
        return null
    }

    return (
        <PresentationContainer>
            <PresentationVisual mediaData={mediaData} />
            <Button
                onClick={() => dispatch({ type: 'stop_presenting' })}
                sx={{ display: "flex", justifyContent: 'center', position: "absolute", top: 15, left: 15, cursor: "pointer", zIndex: 100, height: '10px', width: '10px', padding: 2 }}
            >
                <Close />
            </Button>
            <Box position={"absolute"} top={10}>
                <Crumbs path={mediaData?.Filepath} dispatch={dispatch} includeHome={false} navOnLast={false} navigate={navigate} />
            </Box>
        </PresentationContainer>
    )
}

export default Presentation