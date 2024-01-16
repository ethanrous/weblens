import { useCallback, useEffect, useState } from 'react'

import { MediaData, fileData } from '../types/Types'
import { MediaImage } from './PhotoContainer'
import { FlexColumnBox } from '../Pages/FileBrowser/FilebrowserStyles'
import { Box, CloseButton } from '@mantine/core'
import { IconFile } from '@tabler/icons-react'

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

const StyledMediaImage = ({ mediaData, quality, lazy }: { mediaData: MediaData, quality: "thumbnail" | "fullres", lazy: boolean }) => {

    return (
        <Box pos={'absolute'} style={{ height: "100%", width: "100%" }} onClick={(e) => e.stopPropagation()}>
            <MediaImage mediaId={mediaData.fileHash} blurhash={mediaData.blurHash} quality={quality} lazy={lazy} imgStyle={{ height: "calc(100% - 10px)", width: "calc(100% - 10px)", position: 'absolute', objectFit: "contain" }} />
        </Box>
    )
}

const PresentationVisual = ({ mediaData }: { mediaData: MediaData }) => {
    if (!mediaData) {
        return
    }
    else if (mediaData.mediaType.IsVideo) {
        return (
            <StyledMediaImage key={`${mediaData.fileHash} thumbnail`} mediaData={mediaData} quality={"thumbnail"} lazy={false} />
        )
    } else if (mediaData.mediaType?.FriendlyName == "File") {
        return (
            <IconFile />
        )
    } else {
        return (
            <FlexColumnBox style={{ height: "100%", width: "max-content" }} onClick={e => e.stopPropagation()}>
                <MediaImage mediaId={mediaData.fileHash} blurhash={mediaData.blurHash} quality={"fullres"} lazy={false} imgStyle={{ objectFit: "contain", maxHeight: "100%", height: "100%" }} />
            </FlexColumnBox>
        )
    }
}

function useKeyDown(mediaData, dispatch) {
    const keyDownHandler = useCallback(event => {
        if (!mediaData) {
            return
        }
        else if (event.key === 'Escape') {
            event.preventDefault()
            dispatch({ type: 'set_presentation', presentingId: '' })
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
    }, [mediaData, dispatch])
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

const Presentation = ({ mediaData, parents, dispatch }: { mediaData: MediaData, parents: fileData[], dispatch: React.Dispatch<any> }) => {
    useKeyDown(mediaData, dispatch)

    const [to, setTo] = useState(null)
    const [guiShown, setGuiShown] = useState(false)

    if (!mediaData || !mediaData.mediaType.IsDisplayable) {
        return null
    }

    return (
        <PresentationContainer onMouseMove={(_) => { setGuiShown(true); handleTimeout(to, setTo, setGuiShown) }} onClick={() => dispatch({ type: 'set_presentation', presentingId: '' })}>
            <PresentationVisual mediaData={mediaData} />
            <CloseButton c={'white'} style={{ position: 'absolute', top: guiShown ? 15 : -100, left: 15 }} onClick={() => dispatch({ type: 'set_presentation', presentingId: '' })} />
        </PresentationContainer>
    )
}

export default Presentation