import { Divider } from '@mantine/core'
import {
    IconDownload,
    IconFolder,
    IconHeart,
    IconPhoto,
    IconX,
} from '@tabler/icons-react'
import ReactCodeMirror from '@uiw/react-codemirror'
import { GetFileText } from '@weblens/api/FileBrowserApi'
import WeblensButton from '@weblens/lib/WeblensButton'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { downloadSelected } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { TaskProgContext } from '@weblens/types/files/FBTypes'
import { WeblensFile } from '@weblens/types/files/File'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { likeMedia } from '@weblens/types/media/MediaQuery'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'

import { MediaImage } from '@weblens/types/media/PhotoContainer'
import React, {
    memo,
    ReactNode,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
} from 'react'
import { WebsocketContext } from '../Context'
import { humanFileSize } from '../util'
import { useResize, useResizeDrag } from './hooks'
import { useSessionStore } from './UserInfo'

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
                        w-full z-50 bg-bottom-grey bg-opacity-90 backdrop-blur absolute"
            onMouseMove={onMouseMove}
            onClick={onClick}
            children={children}
        />
    )
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
        const pages: ReactNode[] = []
        for (let i = 0; i < mediaData.GetPageCount(); i++) {
            pages.push(
                <MediaImage
                    key={mediaData.Id() + i}
                    media={mediaData}
                    quality={PhotoQuality.HighRes}
                    pageNumber={i}
                    containerStyle={style}
                    preventClick
                />
            )
        }
        return (
            <div className="flex flex-col no-scrollbar gap-1 h-full">
                {pages.map((p) => p)}
            </div>
        )
    } else {
        return (
            <MediaImage
                media={mediaData}
                quality={PhotoQuality.HighRes}
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

function TextDisplay({ file }: { file: WeblensFile }) {
    const setBlockFocus = useFileBrowserStore((state) => state.setBlockFocus)
    const [content, setContent] = useState('')

    if (!file) {
        return null
    }

    useEffect(() => {
        setBlockFocus(true)
        GetFileText(file.Id())
            .then((r) => {
                setContent(r)
            })
            .catch((err) => console.error(err))

        return () => setBlockFocus(false)
    }, [])

    if (content.length == 0) {
        return null
    }

    return (
        <div
            className="p-8 bg-[#282c34] rounded"
            onClick={(e) => e.stopPropagation()}
        >
            <ReactCodeMirror
                value={content}
                theme={'dark'}
                basicSetup={{ lineNumbers: false, foldGutter: false }}
                minHeight={'100%'}
                minWidth={'100%'}
                editable={false}
            />
        </div>
    )
}

export const FileInfo = ({ file }: { file: WeblensFile }) => {
    const { progDispatch } = useContext(TaskProgContext)
    const mediaData = useMediaStore((state) =>
        state.mediaMap.get(file.GetMediaId())
    )

    const wsSend = useContext(WebsocketContext)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)

    if (!file) {
        return null
    }
    const [size, units] = humanFileSize(file.GetSize())
    return (
        <div
            className="flex grow w-[10%] justify-center"
            onClick={(e) => e.stopPropagation()}
        >
            <div className="flex flex-col justify-center h-max max-w-full gap-2">
                <p className="font-semibold text-3xl truncate">
                    {file.GetFilename()}
                </p>
                <div className="flex flex-row items-center gap-3">
                    <p className="text-2xl">
                        {size}
                        {units}
                    </p>
                    {file.IsFolder() && (
                        <p>{file.GetChildren().length} Item(s)</p>
                    )}
                </div>
                <div className="flex gap-1">
                    <p className="text-xl">
                        {file.GetModified().toLocaleDateString('en-us', {
                            year: 'numeric',
                            month: 'short',
                            day: 'numeric',
                        })}
                    </p>
                </div>
                <WeblensButton
                    label={'Download'}
                    Left={IconDownload}
                    onClick={() => {
                        downloadSelected(
                            [file],
                            removeLoading,
                            progDispatch,
                            wsSend
                        )
                    }}
                />
                {mediaData && (
                    <div>
                        <Divider />
                        <div className="flex gap-1">
                            <IconPhoto />
                            <p className="text-xl">
                                {mediaData
                                    .GetCreateDate()
                                    .toLocaleDateString('en-us', {
                                        year: 'numeric',
                                        month: 'short',
                                        day: 'numeric',
                                    })}
                            </p>
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
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

export function PresentationFile({ file }: { file: WeblensFile }) {
    // useKeyDownPresentation(mediaId, dispatch)

    const [to, setTo] = useState(null)
    const [guiShown, setGuiShown] = useState(false)
    const [likedHover, setLikedHover] = useState(false)
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const { user } = useSessionStore()

    const mediaData = useMediaStore((state) =>
        state.mediaMap.get(file?.GetMediaId())
    )
    const { isLiked, otherLikes } = useMemo(() => {
        if (!mediaData) {
            return { isLiked: false, otherLikes: null }
        }

        const isLiked = mediaData.GetLikedBy()?.includes(user.username)

        const otherLikes =
            (!isLiked && mediaData.GetLikedBy()?.length > 0) ||
            (isLiked && mediaData.GetLikedBy()?.length > 1)

        return { isLiked, otherLikes }
    }, [mediaData])

    const setMediaLiked = useMediaStore((state) => state.setLiked)
    const setPresTarget = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )

    if (!file) {
        return null
    }

    let Visual = null
    if (mediaData) {
        Visual = (
            <ContainerMedia mediaData={mediaData} containerRef={containerRef} />
        )
    } else if (file.IsFolder()) {
        Visual = <IconFolder className="w-[50%] h-[50%]" />
    } else {
        Visual = <TextDisplay file={file} />
    }

    return (
        <PresentationContainer
            onMouseMove={() => {
                setGuiShown(true)
                handleTimeout(to, setTo, setGuiShown)
            }}
            onClick={() => setPresTarget('')}
        >
            <div
                className="presentation-icon top-4 left-4"
                data-shown={guiShown}
            >
                <WeblensButton
                    subtle
                    Left={IconX}
                    onClick={() => setPresTarget('')}
                />
            </div>

            <div
                ref={setContainerRef}
                className="flex justify-center items-center w-[50%] h-full"
            >
                {Visual}
            </div>
            <FileInfo file={file} />

            {mediaData && (
                <div
                    className="presentation-icon bottom-4 right-4"
                    data-shown={guiShown || isLiked}
                    onClick={(e) => {
                        e.stopPropagation()
                        likeMedia(mediaData.Id(), !isLiked).then(() => {
                            setMediaLiked(mediaData.Id(), user.username)
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
            )}
        </PresentationContainer>
    )
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
        const { user } = useSessionStore()

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
                        likeMedia(mediaId, !isLiked).then(() => {
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
