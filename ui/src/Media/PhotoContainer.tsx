import {
    memo,
    useCallback,
    useContext,
    useEffect,
    useRef,
    useState,
} from 'react'
import { UserContext } from '../Context'
import {
    IconExclamationCircle,
    IconPhoto,
    IconPlayerPauseFilled,
    IconPlayerPlayFilled,
    IconVolume3,
} from '@tabler/icons-react'
import { CSSProperties, Loader } from '@mantine/core'
import { UserContextT } from '../types/Types'
import WeblensMedia, { PhotoQuality } from './Media'
import Hls from 'hls.js'

import '../components/style.scss'
import { useKeyDown, useResize, useVideo } from '../components/hooks'
import { WeblensProgress } from '../components/WeblensProgress'

export const MediaImage = memo(
    ({
        media,
        quality,
        fitLogic = 'cover',
        pageNumber = undefined,
        expectFailure = false,
        preventClick = false,
        doFetch = true,
        imgStyle,
        imageClass = '',
        containerStyle,
        containerClass = '',

        disabled = false,
    }: {
        media: WeblensMedia
        quality: PhotoQuality
        fitLogic?: 'contain' | 'cover'
        pageNumber?: number
        expectFailure?: boolean
        preventClick?: boolean
        doFetch?: boolean
        imgStyle?: CSSProperties
        imageClass?: string
        containerStyle?: CSSProperties
        containerClass?: string

        disabled?: boolean
    }) => {
        if (!media) {
            media = new WeblensMedia({ contentId: '' })
        }

        const [loadError, setLoadErr] = useState('')
        const [src, setUrl] = useState({ url: '', id: media.Id() })
        const [videoRef, setVideoRef] = useState<HTMLVideoElement>()
        const { playtime, isPlaying, isWaiting } = useVideo(videoRef)
        const { authHeader }: UserContextT = useContext(UserContext)
        const playerRef = useRef(null)

        useEffect(() => {
            if (doFetch && media.Id() && !media.HasQualityLoaded(quality)) {
                media.LoadBytes(
                    quality,
                    authHeader,
                    pageNumber,
                    () => {
                        setUrl({
                            url: media.GetObjectUrl(quality),
                            id: media.Id(),
                        })
                        setLoadErr(media.HasLoadError())
                    },
                    () => {
                        setUrl({
                            url: media.GetObjectUrl(quality),
                            id: media.Id(),
                        })
                        setLoadErr(media.HasLoadError())
                    }
                )
            }

            if (!doFetch) {
                media.CancelLoad()
            } else if (
                (media.HasQualityLoaded(quality) && src.url === '') ||
                src.id !== media.Id()
            ) {
                setUrl({ url: media.GetObjectUrl(quality), id: media.Id() })
            } else if (media.HighestQualityLoaded() !== '' && src.url === '') {
                setUrl({ url: media.GetObjectUrl('thumbnail'), id: media.Id() })
            }
        }, [media, quality, doFetch])

        const containerClick = useCallback(
            (e) => {
                preventClick && e.stopPropagation()
            },
            [preventClick]
        )

        const shouldShowVideo =
            media.GetMediaType()?.IsVideo && quality === 'fullres'
        // media.HighestQualityLoaded() === 'fullres'

        return (
            <div
                className={`photo-container ${containerClass}`}
                style={containerStyle}
                onClick={containerClick}
            >
                {loadError && !expectFailure && (
                    <IconExclamationCircle color="red" />
                )}
                {((loadError && expectFailure) ||
                    !media.Id() ||
                    !media.HighestQualityLoaded()) && <IconPhoto />}
                {media.Id() !== '' &&
                    quality === 'fullres' &&
                    media.HighestQualityLoaded() !== 'fullres' &&
                    !loadError &&
                    (!media.GetMediaType().IsVideo || isWaiting) && (
                        <Loader
                            color="white"
                            bottom={40}
                            right={40}
                            size={20}
                            style={{ position: 'absolute' }}
                        />
                    )}

                <img
                    className="media-image animate-fade"
                    data-fit-logic={fitLogic}
                    data-disabled={disabled}
                    data-hide={
                        src.url === '' ||
                        media.HasLoadError() ||
                        shouldShowVideo
                    }
                    draggable={false}
                    src={src.url}
                    style={imgStyle}
                />

                <VideoWrapper
                    url={src.url}
                    shouldShowVideo={shouldShowVideo}
                    media={media}
                    fitLogic={fitLogic}
                    imgStyle={imgStyle}
                    videoRef={videoRef}
                    setVideoRef={setVideoRef}
                    isPlaying={isPlaying}
                    playtime={playtime}
                />

                {/* {media?.blurHash && lazy && !imgData && (
                 <Blurhash
                 style={{ position: "absolute" }}
                 height={visibleRef?.current?.clientHeight ? visibleRef.current.clientHeight : 0}
                 width={visibleRef?.current?.clientWidth ? visibleRef.current.clientWidth : 0}
                 hash={media.blurHash}
                 />
                 )} */}
            </div>
        )
    },
    (last, next) => {
        if (last.doFetch !== next.doFetch) {
            return false
        } else if (last.disabled !== next.disabled) {
            return false
        } else if (last.media?.Id() !== next.media?.Id()) {
            return false
        } else if (last.containerStyle !== next.containerStyle) {
            return false
        } else if (
            last.media?.HighestQualityLoaded() !==
            next.media?.HighestQualityLoaded()
        ) {
            return false
        } else if (last.quality !== next.quality) {
            return false
        }
        return true
    }
)

function VideoWrapper({
    url,
    shouldShowVideo,
    fitLogic,
    media,
    imgStyle,
    videoRef,
    setVideoRef,
    isPlaying,
    playtime,
}: {
    url
    shouldShowVideo
    fitLogic
    media
    imgStyle
    videoRef: HTMLVideoElement
    setVideoRef
    isPlaying
    playtime
}) {
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const size = useResize(containerRef)
    const { authHeader } = useContext(UserContext)
    const [showUi, setShowUi] = useState<NodeJS.Timeout>()
    const [volume, setVolume] = useState(0)

    useEffect(() => {
        if (!videoRef) {
            return
        }

        if (videoRef.canPlayType('application/vnd.apple.mpegurl')) {
            console.log('Using native HLS')
            videoRef.src = media.StreamVideoUrl(authHeader)
        } else if (Hls.isSupported()) {
            console.log('Using package HLS')
            var hls = new Hls()
            hls.loadSource(media.StreamVideoUrl(authHeader))
            hls.attachMedia(videoRef)
        }
    }, [videoRef])

    const togglePlayState = useCallback(() => {
        if (!videoRef) {
            return
        }
        if (isPlaying) {
            videoRef.pause()
        } else {
            videoRef.play()
        }
    }, [isPlaying, videoRef])

    useKeyDown(' ', togglePlayState)

    console.log('HEA')

    if (!shouldShowVideo) {
        return null
    }

    if (videoRef) {
        videoRef.volume = volume / 100
    }

    return (
        <div
            ref={setContainerRef}
            className="flex relative items-center justify-center"
            onMouseMove={() => {
                setShowUi((prev) => {
                    if (prev) {
                        clearTimeout(prev)
                    }
                    return setTimeout(() => setShowUi(null), 2000)
                })
            }}
        >
            <div
                className="flex shrink-0 w-[24px] h-[24px] absolute z-50 cursor-pointer transition-opacity duration-300 drop-shadow-xl"
                style={{ opacity: showUi || !isPlaying ? 1 : 0 }}
            >
                {isPlaying && (
                    <IconPlayerPauseFilled onClick={() => videoRef.pause()} />
                )}
                {!isPlaying && (
                    <IconPlayerPlayFilled onClick={() => videoRef.play()} />
                )}
            </div>
            <div
                className="flex justify-end shrink-0 w-[98%] h-[98%] absolute z-50 transition-opacity duration-300 pointer-events-none"
                style={{
                    opacity: (showUi || !isPlaying) && volume === 0 ? 1 : 0,
                }}
            >
                <IconVolume3
                    className="w-4 h-4 pointer-events-auto cursor-pointer"
                    onClick={() => {
                        setVolume(20)
                    }}
                    style={{
                        pointerEvents:
                            (showUi || !isPlaying) && volume === 0
                                ? 'all'
                                : 'none',
                    }}
                />
            </div>
            <video
                ref={setVideoRef}
                autoPlay
                muted={volume === 0}
                className="media-image animate-fade"
                poster={media.GetObjectUrl('thumbnail')}
                data-fit-logic={fitLogic}
                data-hide={
                    url === '' || media.HasLoadError() || !shouldShowVideo
                }
                style={imgStyle}
                onClick={togglePlayState}
            />
            <div
                className="flex absolute justify-center items-end p-3 pointer-events-none transition-opacity duration-300"
                style={{
                    width: size.width,
                    height: size.height,
                    opacity: showUi || !isPlaying ? 1 : 0,
                }}
            >
                <div className="flex flex-row h-2 w-[98%] justify-around absolute">
                    <div className="relative w-[80%]">
                        <WeblensProgress
                            value={
                                (playtime * 100) /
                                (media.GetVideoLength() / 1000)
                            }
                            secondaryValue={
                                videoRef && videoRef.buffered.length
                                    ? videoRef.buffered.end(
                                          videoRef.buffered.length - 1
                                      )
                                    : 0
                            }
                            seekCallback={(v) => {
                                videoRef.currentTime =
                                    (media.GetVideoLength() / 1000) * v
                            }}
                        />
                    </div>
                    <div className="relative w-[10%]">
                        <WeblensProgress
                            value={volume}
                            seekCallback={(v) => {
                                console.log(v)
                                setVolume(v * 100)
                            }}
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}
