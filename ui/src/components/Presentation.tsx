import { useCallback, useEffect, useMemo, useState } from 'react'

import { MediaData } from '../types/Types'
import { MediaImage } from './PhotoContainer'
import { ColumnBox } from '../Pages/FileBrowser/FilebrowserStyles'
import { Box, CloseButton, MantineStyleProp } from '@mantine/core'
import { IconFile } from '@tabler/icons-react'
import { useWindowSize } from './ItemScroller'

export const PresentationContainer = ({ shadeOpacity, onMouseMove, onClick, children }: { shadeOpacity?, onMouseMove?, onClick?, children }) => {
    if (!shadeOpacity) {
        shadeOpacity = "0.90"
    }
    return (
        <Box
            onMouseMove={onMouseMove}
            onClick={onClick}
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
                backgroundColor: `rgb(0, 0, 0, ${shadeOpacity})`,
                backdropFilter: "blur(4px)"
            }}
            children={children}
        />
    )
}

const StyledMediaImage = ({ mediaData, quality, lazy }: { mediaData: MediaData, quality: "thumbnail" | "fullres", lazy: boolean }) => {

    return (
        <Box pos={'absolute'} style={{ height: "100%", width: "100%" }} onClick={(e) => e.stopPropagation()}>
            <MediaImage media={mediaData} quality={quality} lazy={lazy} imgStyle={{ height: "calc(100% - 10px)", width: "calc(100% - 10px)", position: 'absolute', objectFit: "contain" }} />
        </Box>
    )
}

const PresentationVisual = ({ mediaData }: { mediaData: MediaData }) => {
    const [windowSize, setWindowSize] = useState({ height: window.innerHeight, width: window.innerWidth })
    useWindowSize(() => { setWindowSize({ height: window.innerHeight, width: window.innerWidth }) })

    const [absHeight, absWidth] = useMemo(() => {
        if (mediaData.mediaHeight === 0 || mediaData.mediaWidth === 0 || windowSize.height === 0 || windowSize.width === 0) {
            return [0, 0]
        }
        const mediaRatio = mediaData.mediaWidth / mediaData.mediaHeight
        const windowRatio = windowSize.width / windowSize.height
        let absHeight = 0
        let absWidth = 0
        if (mediaRatio > windowRatio) {
            absWidth = windowSize.width * 0.95
            absHeight = (absWidth / mediaData.mediaWidth) * mediaData.mediaHeight
        } else {
            absHeight = windowSize.height * 0.95
            absWidth = (absHeight / mediaData.mediaHeight) * mediaData.mediaWidth
        }
        return [absHeight, absWidth]
    }, [mediaData.mediaHeight, mediaData.mediaWidth, windowSize])

    if (!mediaData) {
        return
    }
    else if (mediaData.mediaType.IsVideo) {
        return (
            <StyledMediaImage key={`${mediaData.fileHash} thumbnail`} mediaData={mediaData} quality={"thumbnail"} lazy={false} />
        )
    } else if (mediaData.mediaType?.FriendlyName === "File") {
        return (
            <IconFile />
        )
    } else {

        return (
            <ColumnBox style={{ height: absHeight, width: absWidth }} onClick={e => { e.stopPropagation(); console.log(mediaData.recognitionTags) }}>
                <MediaImage media={mediaData} quality={"fullres"} lazy={false} imgStyle={{ height: absHeight, width: absWidth, objectFit: 'contain' }} />
            </ColumnBox>
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

const Presentation = ({ mediaData, dispatch }: { mediaData: MediaData, dispatch: React.Dispatch<any> }) => {
    useKeyDown(mediaData, dispatch)

    const [to, setTo] = useState(null)
    const [guiShown, setGuiShown] = useState(false)

    if (!mediaData || !mediaData.mediaType.IsDisplayable) {
        return null
    }

    return (
        <PresentationContainer onMouseMove={(_) => { setGuiShown(true); handleTimeout(to, setTo, setGuiShown) }} onClick={() => dispatch({ type: 'set_presentation', media: null })}>
            <PresentationVisual mediaData={mediaData} />
            {/* <Text style={{ position: 'absolute', bottom: guiShown ? 15 : -100, left: '50vw' }} >{}</Text> */}
            <CloseButton c={'white'} style={{ position: 'absolute', top: guiShown ? 15 : -100, left: 15 }} onClick={() => dispatch({ type: 'set_presentation', presentingId: null })} />
        </PresentationContainer>
    )
}

export default Presentation