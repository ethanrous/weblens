import { IconHeart, IconX } from '@tabler/icons-react'
import React, {
    memo,
    ReactNode,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react'
import WeblensMedia from '../Media/Media'

import { MediaImage } from '../Media/PhotoContainer'
import { SizeT } from '../types/Types'
import { useResize, useResizeDrag } from './hooks'
import WeblensButton from './WeblensButton'
import { useSessionStore } from './UserInfo'
import { likeMedia } from '../Media/MediaQuery'
import { useMediaStore } from '../Media/MediaStateControl'

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
            className="flex justify-center items-center top-0 left-0 p-6 h-full 
                        w-full z-50 fixed bg-bottom-grey bg-opacity-90 backdrop-blur absolute"
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
    // let newWidth
    // if (!containerSize) {
    //     newWidth = 0
    // } else if (containerSize.width < 150 && mediaData.GetPageCount() > 1) {
    //     newWidth = 150
    // } else {
    //     newWidth = containerSize.width
    // }

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
            <div className="flex flex-col no-scrollbar gap-1 h-full">
                {[...Array(mediaData.GetPageCount())].map((p) => (
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
                containerStyle={{
                    ...style,
                    borderRadius: 8,
                    overflow: 'hidden',
                }}
                preventClick
            />
        )
    }
}

export const PresentationVisual = ({
    mediaData,
    Element,
}: {
    mediaData: WeblensMedia
    Element: () => ReactNode
}) => {
    const [screenRef, setScreenRef] = useState(null)
    const [containerRef, setContainerRef] = useState(null)
    const [splitSize, setSplitSize] = useState(-1)
    const [dragging, setDragging] = useState(false)
    const screenSize = useResize(screenRef)
    const splitCalc = useCallback(
        (o) => {
            if (screenSize.width === -1) {
                return
            }
            setSplitSize(
                (o - screenRef.getBoundingClientRect().left - 56) /
                    screenSize.width
            )
        },
        [screenSize.width]
    )

    useResizeDrag(dragging, setDragging, splitCalc)

    const imgStyle = useMemo(() => {
        if (splitSize === -1) {
            return { width: Element ? '50%' : '100%' }
        } else {
            return { width: splitSize * screenSize.width }
        }
    }, [Element, splitSize, screenSize])

    return (
        <div ref={setScreenRef} className="flex items-center h-full w-full">
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
            {mediaData && Element && (
                <div
                    className="flex h-1/6 w-4 cursor-pointer justify-center m-12"
                    onClick={(e) => e.stopPropagation()}
                    onMouseDown={() => setDragging(true)}
                >
                    <div className="h-full w-1/12 bg-[#666666] rounded" />
                </div>
            )}
            {Element && <Element />}
        </div>
    )
}

function useKeyDownPresentation(
    contentId: string,
    dispatch: PresentationDispatchT
) {
    const mediaData = useMediaStore((state) => state.mediaMap.get(contentId))

    const keyDownHandler = useCallback(
        (event) => {
            if (!contentId) {
                return
            } else if (event.key === 'Escape') {
                event.preventDefault()
                event.stopPropagation()
                dispatch.setPresentationTarget('')
            } else if (event.key === 'ArrowLeft') {
                event.preventDefault()
                if (!mediaData.Prev()) {
                    return
                }
                dispatch.setPresentationTarget(mediaData.Prev()?.Id())
            } else if (event.key === 'ArrowRight') {
                event.preventDefault()
                if (!mediaData.Next()) {
                    return
                }
                dispatch.setPresentationTarget(mediaData.Next()?.Id())
            } else if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                event.preventDefault()
            }
        },
        [contentId, dispatch, mediaData]
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

interface PresentationDispatchT {
    setPresentationTarget(targetId: string)
}

const Presentation = memo(
    ({
        mediaId,
        element,
        dispatch,
    }: {
        mediaId: string
        dispatch: PresentationDispatchT
        element?
    }) => {
        useKeyDownPresentation(mediaId, dispatch)

        const [to, setTo] = useState(null)
        const [guiShown, setGuiShown] = useState(false)
        const [likedHover, setLikedHover] = useState(false)
        const { user, auth } = useSessionStore()

        const mediaData = useMediaStore((state) => state.mediaMap.get(mediaId))
        const isLiked = useMediaStore((state) => {
            const m = state.mediaMap.get(mediaId)
            return m ? m.GetLikedBy().includes(user.username) : false
        })
        const setMediaLiked = useMediaStore((state) => state.setLiked)

        if (!mediaId || !mediaData) {
            return null
        }

        const otherLikes =
            (!isLiked && mediaData.GetLikedBy()?.length > 0) ||
            (isLiked && mediaData.GetLikedBy()?.length > 1)

        return (
            <PresentationContainer
                onMouseMove={() => {
                    setGuiShown(true)
                    handleTimeout(to, setTo, setGuiShown)
                }}
                onClick={() => dispatch.setPresentationTarget('')}
            >
                <PresentationVisual
                    key={mediaId}
                    mediaData={mediaData}
                    Element={element}
                />

                <div
                    className="presentation-icon top-4 left-4"
                    data-shown={guiShown}
                >
                    <WeblensButton
                        subtle
                        Left={IconX}
                        onClick={() => dispatch.setPresentationTarget('')}
                    />
                </div>
                <div
                    className="presentation-icon bottom-4 right-4"
                    data-shown={guiShown || isLiked}
                    onClick={(e) => {
                        e.stopPropagation()
                        likeMedia(mediaId, !isLiked, auth).then(() => {
                            setMediaLiked(mediaId, user.username)
                        })
                    }}
                    onMouseOver={() => {
                        setLikedHover(true)
                    }}
                    onMouseLeave={() => {
                        setLikedHover(false)
                    }}
                >
                    <div className="flex flex-col h-max items-center justify-center">
                        {mediaData.GetLikedBy()?.length !== 0 && (
                            <p className="absolute text-xs right-0 -bottom-1">
                                {mediaData.GetLikedBy()?.length}
                            </p>
                        )}
                        <IconHeart
                            size={30}
                            fill={isLiked ? 'red' : ''}
                            color={isLiked ? 'red' : 'white'}
                        />
                    </div>
                    {likedHover && otherLikes && (
                        <div className="flex flex-col bg-bottom-grey p-2 rounded items-center absolute bottom-7 right-0 w-max">
                            <p>Likes</p>
                            <div className="bg-raised-grey h-[1px] w-full m-1" />
                            {mediaData.GetLikedBy().map((username: string) => {
                                return (
                                    <p className="text-lg" key={username}>
                                        {username}
                                    </p>
                                )
                            })}
                        </div>
                    )}
                </div>
            </PresentationContainer>
        )
    },
    (prev, next) => {
        if (prev.mediaId !== next.mediaId) {
            return false
        } else if (prev.element !== next.element) {
            return false
        } else if (prev.dispatch !== next.dispatch) {
            return false
        }

        return true
    }
)

export default Presentation
