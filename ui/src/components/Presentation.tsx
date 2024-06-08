import React, {
    memo,
    ReactNode,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react'

import { MediaImage } from './PhotoContainer'
import WeblensMedia from '../classes/Media'
import { IconX } from '@tabler/icons-react'
import { SizeT } from '../types/Types'
import { useResize } from './hooks'

export const PresentationContainer = ({
    onMouseMove,
    onClick,
    children,
}: {
    onMouseMove?
    onClick?
    children
}) => {
    return (
        <div
            className="flex justify-center items-center top-0 left-0 p-6 h-full w-full z-50 fixed bg-bottom-grey bg-opacity-90 backdrop-blur"
            onMouseMove={onMouseMove}
            onClick={onClick}
            children={children}
        />
    )
}

export function GetMediaFullscreenSize(
    mediaData: WeblensMedia,
    containerSize: SizeT
): SizeT {
    let newWidth
    if (!containerSize) {
        newWidth = 0
    } else if (containerSize.width < 150 && mediaData.GetPageCount() > 1) {
        newWidth = 150
    } else {
        newWidth = containerSize.width
    }

    if (
        !mediaData ||
        !mediaData.GetHeight() ||
        !mediaData.GetWidth() ||
        !containerSize.height ||
        !containerSize.width
    ) {
        return { height: 0, width: 0 }
    }
    const mediaRatio = mediaData.GetWidth() / mediaData.GetHeight()
    const windowRatio = containerSize.width / containerSize.height
    let absHeight = 0
    let absWidth = 0
    if (mediaRatio > windowRatio) {
        absWidth = containerSize.width
        absHeight = (absWidth / mediaData.GetWidth()) * mediaData.GetHeight()
    } else {
        absHeight = containerSize.height
        absWidth = (absHeight / mediaData.GetHeight()) * mediaData.GetWidth()
    }
    return { height: absHeight, width: absWidth }
}

export const ContainerMedia = ({
    mediaData,
    containerRef,
}: {
    mediaData: WeblensMedia
    containerRef
}) => {
    const [boxSize, setBoxSize] = useState({
        height: 0,
        width: 0,
    })
    const { width: containerWidth, height: containerHeight } =
        useResize(containerRef)

    useEffect(() => {
        let newWidth: number
        if (!containerRef) {
            newWidth = 0
        } else if (containerWidth < 150 && mediaData.GetPageCount() > 1) {
            newWidth = 150
        } else {
            newWidth = containerWidth
        }
        setBoxSize({ height: containerHeight, width: newWidth })
    }, [containerWidth, containerHeight])

    const style = useMemo(() => {
        if (
            !mediaData ||
            !mediaData.GetHeight() ||
            !mediaData.GetWidth() ||
            !boxSize.height ||
            !boxSize.width
        ) {
            return { height: 0, width: 0 }
        }
        const mediaRatio = mediaData.GetWidth() / mediaData.GetHeight()
        const windowRatio = boxSize.width / boxSize.height
        let absHeight = 0
        let absWidth = 0
        if (mediaRatio > windowRatio) {
            absWidth = boxSize.width
            absHeight =
                (absWidth / mediaData.GetWidth()) * mediaData.GetHeight()
        } else {
            absHeight = boxSize.height
            absWidth =
                (absHeight / mediaData.GetHeight()) * mediaData.GetWidth()
        }
        return { height: absHeight, width: absWidth }
    }, [mediaData, mediaData.GetHeight(), mediaData.GetWidth(), boxSize])

    if (!mediaData || !containerRef) {
        return <></>
    }

    if (mediaData.GetPageCount() > 1) {
        return (
            <div className="no-scrollbar gap-1">
                {[...Array(mediaData.GetPageCount()).keys()].map((p) => (
                    <MediaImage
                        key={p}
                        media={mediaData}
                        quality={'fullres'}
                        pageNumber={p}
                        containerStyle={style}
                        preventClick
                    />
                ))}
            </div>
        )
    } else {
        return (
            <MediaImage
                media={mediaData}
                quality={'fullres'}
                containerStyle={style}
                preventClick
            />
        )
    }
}

const PresentationVisual = ({
    mediaData,
    Element,
}: {
    mediaData: WeblensMedia
    Element: () => ReactNode
}) => {
    const [containerRef, setContainerRef] = useState(null)

    const imgStyle = useMemo(() => {
        return { width: Element ? '50%' : '100%' }
    }, [Element])

    return (
        <div className="flex items-center justify-around h-full w-full">
            {mediaData && (
                <div
                    className="flex items-center justify-center h-full"
                    style={imgStyle}
                    ref={setContainerRef}
                >
                    <ContainerMedia
                        mediaData={mediaData}
                        containerRef={containerRef}
                    />
                </div>
            )}
            {Element && <Element />}
        </div>
    )
}

function useKeyDownPresentation(itemId: string, dispatch) {
    const keyDownHandler = useCallback(
        (event) => {
            if (!itemId) {
                return
            } else if (event.key === 'Escape') {
                event.preventDefault()
                dispatch({ type: 'stop_presenting' })
            } else if (event.key === 'ArrowLeft') {
                event.preventDefault()
                dispatch({ type: 'presentation_previous' })
            } else if (event.key === 'ArrowRight') {
                event.preventDefault()
                dispatch({ type: 'presentation_next' })
            } else if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                event.preventDefault()
            }
        },
        [itemId, dispatch]
    )
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

const Presentation = memo(
    ({
        itemId,
        mediaData,
        element,
        dispatch,
    }: {
        itemId: string
        mediaData: WeblensMedia
        dispatch: React.Dispatch<any>
        element?
    }) => {
        useKeyDownPresentation(itemId, dispatch)

        const [to, setTo] = useState(null)
        const [guiShown, setGuiShown] = useState(false)

        if (!mediaData) {
            return null
        }

        return (
            <PresentationContainer
                onMouseMove={() => {
                    setGuiShown(true)
                    handleTimeout(to, setTo, setGuiShown)
                }}
                onClick={() =>
                    dispatch({ type: 'set_presentation', media: null })
                }
            >
                <PresentationVisual
                    key={mediaData.Id()}
                    mediaData={mediaData}
                    Element={element}
                />
                {/* <Text style={{ position: 'absolute', bottom: guiShown ? 15 : -100, left: '50vw' }} >{}</Text> */}

                <div
                    className="close-icon"
                    data-shown={guiShown.toString()}
                    onClick={() =>
                        dispatch({
                            type: 'set_presentation',
                            presentingId: null,
                        })
                    }
                >
                    <IconX />
                </div>
            </PresentationContainer>
        )
    },
    (prev, next) => {
        if (prev.itemId !== next.itemId) {
            return false
        }

        return true
    }
)

export default Presentation
