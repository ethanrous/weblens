import {
    IconChevronLeft,
    IconChevronRight,
    IconDownload,
    IconFolder,
    IconHeart,
    IconPhoto,
    IconUser,
    IconX,
} from '@tabler/icons-react'
import { FileApi } from '@weblens/api/FileBrowserApi'
import MediaApi from '@weblens/api/MediaApi'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import {
    useKeyDown,
    useResize,
    useResizeDrag,
    useTimeout,
} from '@weblens/lib/hooks'
import {
    calculateShareId,
    downloadSelected,
} from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import WeblensFile from '@weblens/types/files/File'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import {
    MouseEventHandler,
    ReactNode,
    RefObject,
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react'
import { useNavigate } from 'react-router-dom'

import { humanFileSize } from '../util'
import { useSessionStore } from './UserInfo'

export const PresentationContainer = ({
    onMouseMove,
    onClick,
    children,
}: {
    onMouseMove?: MouseEventHandler<HTMLDivElement>
    onClick?: MouseEventHandler<HTMLDivElement>
    children?: ReactNode
}) => {
    return (
        <div
            className="bg-background-primary/85 absolute top-0 left-0 z-50 flex h-full w-full items-center justify-center gap-6 p-6 backdrop-blur-sm"
            onMouseMove={onMouseMove}
            onClick={onClick}
        >
            {children}
        </div>
    )
}

export const ContainerMedia = ({
    mediaData,
    containerRef,
}: {
    mediaData: WeblensMedia
    containerRef: RefObject<HTMLDivElement | null>
}) => {
    const [boxSize, setBoxSize] = useState({
        height: 0,
        width: 0,
    })
    const { width: containerWidth, height: containerHeight } =
        useResize(containerRef)

    const height = mediaData.GetHeight()
    const width = mediaData.GetWidth()

    useEffect(() => {
        let newWidth: number
        if (!containerRef.current) {
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
        const mediaRatio = width / height
        const windowRatio = boxSize.width / boxSize.height
        let absHeight = 0
        let absWidth = 0
        if (mediaRatio > windowRatio) {
            absWidth = boxSize.width
            absHeight = (absWidth / width) * height
        } else {
            absHeight = boxSize.height
            absWidth = (absHeight / height) * width
        }

        return { height: absHeight, width: absWidth }
    }, [mediaData, height, width, boxSize])

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
            <div className="no-scrollbar flex h-full flex-col gap-1">
                {pages.map((p) => p)}
            </div>
        )
    } else {
        return (
            <MediaImage
                key={mediaData.Id()}
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

function TextDisplay({
    file,
    shareId,
}: {
    file: WeblensFile
    shareId: string
}) {
    const setBlockFocus = useFileBrowserStore((state) => state.setBlockFocus)
    const [content, setContent] = useState('')

    useEffect(() => {
        setBlockFocus(true)
        FileApi.getFileText(file.Id(), shareId)
            .then((r) => {
                setContent(r.data)
            })
            .catch(ErrorHandler)

        return () => setBlockFocus(false)
    }, [])

    if (!file) {
        return null
    }

    if (content.length == 0) {
        return null
    }

    return (
        <div
            className="bg-background-secondary rounded-sm p-8"
            onClick={(e) => e.stopPropagation()}
        >
            {/* <ReactCodeMirror */}
            {/*     value={content} */}
            {/*     theme={'dark'} */}
            {/*     basicSetup={{ lineNumbers: false, foldGutter: false }} */}
            {/*     minHeight={'100%'} */}
            {/*     minWidth={'100%'} */}
            {/*     editable={false} */}
            {/* /> */}
        </div>
    )
}

function MediaHeart({ mediaData }: { mediaData: WeblensMedia }) {
    const user = useSessionStore((state) => state.user)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const setMediaLiked = useMediaStore((state) => state.setLiked)
    const [likedHover, setLikedHover] = useState(false)

    const { isLiked, otherLikes } = useMemo(() => {
        if (!mediaData) {
            return { isLiked: false, otherLikes: null }
        }

        const isLiked = mediaData.GetLikedBy()?.includes(user.username)

        const otherLikes =
            (!isLiked && mediaData.GetLikedBy()?.length > 0) ||
            (isLiked && mediaData.GetLikedBy()?.length > 1)

        return { isLiked, otherLikes }
    }, [mediaData, user.username])

    return (
        <div
            className="cursor-pointer"
            data-shown={true}
            onClick={(e) => {
                e.stopPropagation()
                MediaApi.setMediaLiked(mediaData.Id(), !isLiked, shareId)
                    .then(() => {
                        setMediaLiked(mediaData.Id(), user.username)
                    })
                    .catch(ErrorHandler)
            }}
            onMouseOver={() => {
                setLikedHover(true)
            }}
            onMouseLeave={() => {
                setLikedHover(false)
            }}
        >
            <div className="relative flex h-max flex-row items-center justify-center">
                <IconHeart
                    size={30}
                    fill={isLiked ? 'red' : ''}
                    color={isLiked ? 'red' : 'white'}
                />
                {mediaData.GetLikedBy()?.length !== 0 && (
                    <p className="mt-auto text-xs">
                        {mediaData.GetLikedBy()?.length}
                    </p>
                )}
            </div>
            {likedHover && otherLikes && (
                <div className="bg-background-secondary/75 absolute right-0 bottom-7 flex w-max flex-col items-center rounded-sm p-2">
                    <p>Likes</p>
                    <div className="bg-raised-grey m-1 h-[1px] w-full" />
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
    )
}

export const FileInfo = ({ file }: { file: WeblensFile }) => {
    const mediaData = useMediaStore((state) =>
        state.mediaMap.get(file.GetContentId())
    )
    const user = useSessionStore((state) => state.user)

    const removeLoading = useFileBrowserStore((state) => state.removeLoading)
    const nav = useNavigate()

    if (!file) {
        return null
    }
    const [size, units] = humanFileSize(file.GetSize())
    return (
        <div
            className="text-color-text-primary relative mx-auto flex w-max"
            onClick={(e) => e.stopPropagation()}
        >
            <div className="flex h-max max-w-full flex-col justify-center gap-2">
                {file.IsFolder() && <IconFolder size={'1em'} />}
                <h3 className="truncate font-bold">{file.GetFilename()}</h3>
                <div className="flex flex-row items-center text-white">
                    <h4>{size}</h4>
                    <h4>{units}</h4>
                </div>
                <div className="flex gap-1">
                    <h4>
                        {file.GetModified().toLocaleDateString('en-us', {
                            year: 'numeric',
                            month: 'short',
                            day: 'numeric',
                        })}
                    </h4>
                </div>
                <div className="flex flex-row gap-2">
                    <WeblensButton
                        label={'Download'}
                        Left={IconDownload}
                        onClick={() => {
                            const shareId = calculateShareId([file])
                            downloadSelected(
                                [file],
                                removeLoading,
                                shareId
                            ).catch(ErrorHandler)
                        }}
                    />
                    {mediaData?.mediaType?.IsDisplayable && (
                        <WeblensButton
                            label={'Jpeg'}
                            Left={IconDownload}
                            onClick={() => {
                                const shareId = calculateShareId([file])
                                downloadSelected(
                                    [file],
                                    removeLoading,
                                    shareId,
                                    'jpeg'
                                ).catch(ErrorHandler)
                            }}
                        />
                    )}
                </div>
                {mediaData && (
                    <div>
                        {!user.isLoggedIn && (
                            <div className="flex items-center">
                                <WeblensButton
                                    squareSize={40}
                                    label="Login"
                                    subtle
                                    Left={IconUser}
                                    onClick={() => {
                                        const path = window.location.pathname
                                        nav('/login', {
                                            state: { returnTo: path },
                                        })
                                    }}
                                />
                                <p>To Like and Edit Media</p>
                            </div>
                        )}
                        {user.isLoggedIn && (
                            <div className="flex items-center gap-1 text-white">
                                <IconPhoto className="shrink-0" />
                                <p className="mr-4 text-xl text-nowrap">
                                    {mediaData
                                        .GetCreateDate()
                                        .toLocaleDateString('en-us', {
                                            year: 'numeric',
                                            month: 'short',
                                            day: 'numeric',
                                        })}
                                </p>
                                <MediaHeart mediaData={mediaData} />
                            </div>
                        )}
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
    Element?: () => ReactNode
}) => {
    const screenRef = useRef<HTMLDivElement>(null)
    const containerRef = useRef<HTMLDivElement>(null)
    const [splitSize, setSplitSize] = useState(-1)
    const [dragging, setDragging] = useState(false)
    const screenSize = useResize(screenRef)
    const splitCalc = useCallback(
        (o: number) => {
            if (screenSize.width === -1 || !screenRef.current) {
                return
            }
            setSplitSize(
                (o - screenRef.current.getBoundingClientRect().left - 56) /
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
        <div ref={screenRef} className="flex h-full w-full items-center">
            {mediaData && (
                <div
                    className="flex h-full items-center justify-center"
                    style={imgStyle}
                    ref={containerRef}
                >
                    <ContainerMedia
                        mediaData={mediaData}
                        containerRef={containerRef}
                    />
                </div>
            )}
            {mediaData && Element && (
                <div
                    className="m-12 flex h-1/6 w-4 cursor-pointer justify-center"
                    onClick={(e) => e.stopPropagation()}
                    onMouseDown={() => setDragging(true)}
                >
                    <div className="h-full w-1/12 rounded-sm bg-[#666666]" />
                </div>
            )}
            {Element && <Element />}
        </div>
    )
}

function useKeyDownPresentation(
    contentId: string,
    setTarget: (targetId: string) => void
) {
    const mediaData = useMediaStore((state) => state.mediaMap.get(contentId))

    const keyDownHandler = useCallback(
        (event: KeyboardEvent) => {
            if (!contentId || !mediaData) {
                return
            } else if (event.key === 'Escape') {
                event.preventDefault()
                event.stopPropagation()
                setTarget('')
            } else if (event.key === 'ArrowLeft') {
                event.preventDefault()
                if (!mediaData.Prev()) {
                    return
                }

                const prevId = mediaData.Prev()?.Id()
                if (prevId) {
                    setTarget(prevId)
                }
            } else if (event.key === 'ArrowRight') {
                event.preventDefault()
                if (!mediaData.Next()) {
                    return
                }

                const nextId = mediaData.Next()?.Id()
                if (nextId) {
                    setTarget(nextId)
                }
            } else if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                event.preventDefault()
            }
        },
        [contentId, mediaData]
    )
    useEffect(() => {
        window.addEventListener('keydown', keyDownHandler)
        return () => {
            window.removeEventListener('keydown', keyDownHandler)
        }
    }, [keyDownHandler])
}

export function PresentationFile({ file }: { file: WeblensFile }) {
    const [guiShown, setGuiShown] = useState(false)

    const closeGui = useCallback(() => {
        setGuiShown(false)
    }, [])

    const { reset } = useTimeout(closeGui, 2000, {
        onStart: () => setGuiShown(true),
    })

    const [fileInfoOpen, setFileInfoOpen] = useState(true)
    const containerRef = useRef<HTMLDivElement>(null)

    const mediaMap = useMediaStore((state) => state.mediaMap)
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setPresTarget = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )

    useKeyDown('Escape', () => {
        if (file) {
            setPresTarget('')
        }
    })

    const contentId = file?.GetContentId()
    const mediaData = mediaMap.get(contentId)

    let Visual = null
    if (mediaData && mediaData.Id() !== '') {
        Visual = (
            <ContainerMedia mediaData={mediaData} containerRef={containerRef} />
        )
    } else if (file.IsFolder()) {
        Visual = <IconFolder className="h-[50%] w-[50%]" />
    } else {
        Visual = <TextDisplay file={file} shareId={shareId} />
    }

    const ToggleInfoIcon = fileInfoOpen ? IconChevronRight : IconChevronLeft

    return (
        <PresentationContainer
            onMouseMove={() => {
                reset()
            }}
            onClick={() => setPresTarget('')}
        >
            <div
                className="absolute top-4 left-4 opacity-0 transition hover:opacity-100 data-[shown=true]:opacity-100"
                data-shown={guiShown}
            >
                <WeblensButton
                    subtle
                    Left={IconX}
                    onClick={() => setPresTarget('')}
                />
            </div>

            <div
                ref={containerRef}
                className="mx-auto flex h-full max-w-full min-w-0 grow items-center justify-center"
            >
                {Visual}
            </div>
            <WeblensButton
                flavor="outline"
                Left={ToggleInfoIcon}
                className="max-4-[4%] shrink-0"
                onClick={(e) => {
                    e.stopPropagation()
                    setFileInfoOpen(!fileInfoOpen)
                }}
            />
            {/* <ToggleInfoIcon */}
            {/*     className="max-4-[4%] shrink-0 cursor-pointer text-white" */}
            {/* /> */}
            <div
                className="relative flex w-max grow items-center transition-[max-width]"
                style={{
                    maxWidth: fileInfoOpen ? '40%' : 0,
                    opacity: fileInfoOpen ? '100%' : 0,
                }}
            >
                <FileInfo file={file} />
            </div>
        </PresentationContainer>
    )
}

interface PresentationProps {
    mediaId: string
    setTarget: (targetId: string) => void
    element?: () => ReactNode
}

function Presentation({ mediaId, element, setTarget }: PresentationProps) {
    useKeyDownPresentation(mediaId, setTarget)

    const [guiShown, setGuiShown] = useState(false)
    const { reset } = useTimeout(() => setGuiShown(false), 1000, {
        onStart: () => setGuiShown(true),
    })
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
                reset()
            }}
            onClick={() => setTarget('')}
        >
            <PresentationVisual
                key={mediaId}
                mediaData={mediaData}
                Element={element}
            />

            <div className="absolute top-4 left-4" data-shown={guiShown}>
                <WeblensButton
                    subtle
                    Left={IconX}
                    onClick={() => setTarget('')}
                />
            </div>
            <div
                className="presentation-icon right-4 bottom-4"
                data-shown={guiShown || isLiked}
                onClick={(e) => {
                    e.stopPropagation()
                    MediaApi.setMediaLiked(mediaData.Id(), !isLiked)
                        .then(() => {
                            setMediaLiked(mediaData.Id(), user.username)
                        })
                        .catch(ErrorHandler)
                }}
                onMouseOver={() => {
                    setLikedHover(true)
                }}
                onMouseLeave={() => {
                    setLikedHover(false)
                }}
            >
                <div className="flex h-max flex-col items-center justify-center">
                    {mediaData.GetLikedBy()?.length !== 0 && (
                        <p className="absolute right-0 -bottom-1 text-xs">
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
                    <div className="bg-background-secondary/75 absolute right-0 bottom-7 flex w-max flex-col items-center rounded-sm p-2">
                        <p>Likes</p>
                        <div className="bg-raised-grey m-1 h-[1px] w-full" />
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
}

export default Presentation
