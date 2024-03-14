import { memo, useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { MediaData } from '../types/Types'
import { MediaImage } from './PhotoContainer'
import { Box, CloseButton } from '@mantine/core'
import { useWindowSize } from './ItemScroller'
import { ColumnBox } from '../Pages/FileBrowser/FilebrowserStyles'

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

export const ContainerMedia = ({ mediaData, containerRef }: { mediaData: MediaData, containerRef }) => {
    const [boxSize, setBoxSize] = useState({ height: containerRef.current?.clientHeight || 0, width: containerRef.current?.clientWidth || 0 })
    useWindowSize(() => setBoxSize({ height: containerRef.current?.clientHeight, width: containerRef.current?.clientWidth }))

    useEffect(() => {
        setBoxSize({ height: containerRef.current?.clientHeight, width: containerRef.current?.clientWidth })
    }, [containerRef])

    const [absHeight, absWidth] = useMemo(() => {
        if (!mediaData || mediaData.mediaHeight === 0 || mediaData.mediaWidth === 0 || boxSize.height === 0 || boxSize.width === 0) {
            return [0, 0]
        }
        const mediaRatio = mediaData.mediaWidth / mediaData.mediaHeight
        const windowRatio = boxSize.width / boxSize.height
        let absHeight = 0
        let absWidth = 0
        if (mediaRatio > windowRatio) {
            absWidth = boxSize.width
            absHeight = (absWidth / mediaData.mediaWidth) * mediaData.mediaHeight
        } else {
            absHeight = boxSize.height
            absWidth = (absHeight / mediaData.mediaHeight) * mediaData.mediaWidth
        }
        return [absHeight, absWidth]
    }, [mediaData, mediaData?.mediaHeight, mediaData?.mediaWidth, boxSize])

    if (!mediaData || !containerRef.current) {
        return null
    }

    if (mediaData.pageCount > 1) {
        return (
            <ColumnBox style={{ overflow: 'scroll', gap: absHeight * 0.02 }}>
                {[...Array(mediaData.pageCount).keys()].map(p => <MediaImage media={mediaData} quality={"fullres"} pageNumber={p} lazy={false} containerStyle={{ height: absHeight * 0.85, width: absWidth }} preventClick />)}
            </ColumnBox>
        )
    } else {
        return (
            <MediaImage media={mediaData} quality={"fullres"} lazy={false} containerStyle={{ height: absHeight, width: absWidth }} preventClick />
        )
    }
}

const PresentationVisual = ({ mediaData, Element }: { mediaData: MediaData, Element }) => {
    const containerRef = useRef(null);
    return (
        <Box style={{ height: '100%', width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'space-around' }}>
            {mediaData && (
                <Box style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: Element ? '60%' : '100%', height: '100%', marginLeft: Element ? '5%' : '0px' }} ref={containerRef}>
                    <ContainerMedia mediaData={mediaData} containerRef={containerRef} />
                </Box>
            )}
            {Element && (
                <Element />
            )}
            {/* <Box style={{ width: mediaData ? '50%' : '100%', alignItems: 'center' }}>
            </Box> */}
        </Box>
    )
}

function useKeyDown(itemId: string, dispatch) {
    const keyDownHandler = useCallback(event => {
        if (!itemId) {
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
    }, [itemId, dispatch])
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

const Presentation = memo(({ itemId, mediaData, element, dispatch }: { itemId: string, mediaData: MediaData, dispatch: React.Dispatch<any>, element?}) => {
    useKeyDown(itemId, dispatch)

    const [to, setTo] = useState(null)
    const [guiShown, setGuiShown] = useState(false)

    console.log(mediaData)

    if (!mediaData) {
        return null
    }

    return (
        <PresentationContainer onMouseMove={() => { setGuiShown(true); handleTimeout(to, setTo, setGuiShown) }} onClick={() => dispatch({ type: 'set_presentation', media: null })}>
            <PresentationVisual mediaData={mediaData} Element={element} />
            {/* <Text style={{ position: 'absolute', bottom: guiShown ? 15 : -100, left: '50vw' }} >{}</Text> */}
            <CloseButton c={'white'} style={{ position: 'absolute', top: guiShown ? 15 : -100, left: 15 }} onClick={() => dispatch({ type: 'set_presentation', presentingId: null })} />
        </PresentationContainer>
    )
}, (prev, next) => {
    if (prev.itemId !== next.itemId) {
        return false
    }
    return true
})

export default Presentation