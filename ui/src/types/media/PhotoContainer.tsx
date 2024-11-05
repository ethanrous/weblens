import { CSSProperties, Loader } from '@mantine/core'
import {
    IconExclamationCircle,
    IconPhoto,
    IconPlayerPauseFilled,
    IconPlayerPlayFilled,
    IconVolume3,
} from '@tabler/icons-react'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { useKeyDown, useResize, useVideo } from 'components/hooks'
import Hls from 'hls.js'

import 'components/style.scss'
import { memo, MouseEvent, useCallback, useEffect, useState } from 'react'

export const MediaImage = memo(
    ({
        media,
        quality,
        fitLogic = 'cover',
        pageNumber = 0,
        expectFailure = false,
        preventClick = false,
        doFetch = true,
        imgStyle,
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

        useEffect(() => {
            if (
                media.GetMediaType() &&
                doFetch &&
                media.Id() &&
                !media.HasQualityLoaded(quality)
            ) {
                media.LoadBytes(
                    quality,
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
                setUrl({
                    url: media.GetObjectUrl(PhotoQuality.LowRes),
                    id: media.Id(),
                })
            }
            return () => {
                media.CancelLoad()
            }
        }, [media, quality, doFetch, media.GetMediaType()])

        const containerClick = useCallback(
            (e: MouseEvent) => {
                if (preventClick) {
                    e.stopPropagation()
                }
            },
            [preventClick]
        )

        const shouldShowVideo =
            media.GetMediaType()?.IsVideo && quality === PhotoQuality.HighRes

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
                    media.GetMediaType() &&
                    quality === PhotoQuality.HighRes &&
                    media.HighestQualityLoaded() !== PhotoQuality.HighRes &&
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
                    className="media-image"
                    data-fit-logic={fitLogic}
                    data-disabled={disabled}
                    data-hide={
                        src.url === '' ||
                        media.HasLoadError() ||
                        shouldShowVideo ||
                        !media.HighestQualityLoaded()
                    }
                    draggable={false}
                    src={src.url}
                    style={imgStyle}
                    data-id={media.Id()}
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
    url: string
    shouldShowVideo: boolean
    fitLogic: string
    media: WeblensMedia
    imgStyle: CSSProperties
    videoRef: HTMLVideoElement
    setVideoRef: (r: HTMLVideoElement) => void
    isPlaying: boolean
    playtime: number
}) {
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const size = useResize(containerRef)

    const [showUi, setShowUi] = useState<NodeJS.Timeout>()
    const [volume, setVolume] = useState(0)
    const [playtimeInternal, setPlaytime] = useState(0)

    useEffect(() => {
        setPlaytime(playtime)
    }, [playtime])

    useEffect(() => {
        if (!videoRef) {
            return
        }

        if (videoRef.canPlayType('application/vnd.apple.mpegurl')) {
            videoRef.src = media.StreamVideoUrl()
        } else if (Hls.isSupported()) {
            const hls = new Hls()
            hls.loadSource(media.StreamVideoUrl())
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

    useKeyDown(' ', togglePlayState, !shouldShowVideo)

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
                className="flex shrink-0 w-[24px] h-[24px] absolute z-50 cursor-pointer
                            transition-opacity duration-300 drop-shadow-xl"
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
                className="flex justify-end shrink-0 w-[98%] h-[98%] absolute z-50
                            transition-opacity duration-300 pointer-events-none"
                style={{
                    opacity: (showUi || !isPlaying) && volume === 0 ? 1 : 0,
                }}
            >
                <IconVolume3
                    className="w-5 h-5 pointer-events-auto cursor-pointer"
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
                poster={media.GetObjectUrl(PhotoQuality.LowRes)}
                data-fit-logic={fitLogic}
                data-hide={
                    url === '' || media.HasLoadError() || !shouldShowVideo
                }
                style={imgStyle}
                onClick={togglePlayState}
            />
            <div
                className="flex absolute justify-center items-end p-3 pointer-events-none
                            transition-opacity duration-300"
                style={{
                    width: size.width,
                    height: size.height,
                    opacity: showUi || !isPlaying ? 1 : 0,
                }}
            >
                <div className="flex flex-row h-2 w-[98%] justify-around absolute">
                    <div className="relative w-[80%]">
                        <WeblensProgress
                            height={12}
                            value={Math.round(
                                (playtimeInternal * 100000) /
                                    media.GetVideoLength()
                            )}
                            secondaryValue={
                                videoRef && videoRef.buffered.length
                                    ? videoRef.buffered.end(
                                          videoRef.buffered.length - 1
                                      )
                                    : 0
                            }
                            seekCallback={(v) => {
                                if (videoRef) {
                                    const newTime =
                                        media.GetVideoLength() * (v / 100000)
                                    videoRef.currentTime = newTime
                                    setPlaytime(newTime)
                                }
                            }}
                        />
                    </div>
                    <div className="relative w-[10%]">
                        <WeblensProgress
                            height={12}
                            value={volume}
                            seekCallback={(v) => {
                                setVolume(v)
                            }}
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}
