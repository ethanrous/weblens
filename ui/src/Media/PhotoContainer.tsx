import { memo, useCallback, useContext, useEffect, useState } from 'react'
import { UserContext } from '../Context'
import {
    IconExclamationCircle,
    IconPhoto,
    IconPlayerPauseFilled,
    IconPlayerPlayFilled,
} from '@tabler/icons-react'
import { CSSProperties, Loader } from '@mantine/core'
import { UserContextT } from '../types/Types'
import WeblensMedia, { PhotoQuality } from './Media'

import '../components/style.scss'
import { WeblensProgress } from '../components/WeblensProgress'
import { useResize, useVideo } from '../components/hooks'

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
}) {
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const size = useResize(containerRef)
    const { authHeader } = useContext(UserContext)

    if (!shouldShowVideo) {
        return null
    }
    console.log(media.GetVideoLength())

    return (
        <div ref={setContainerRef} className="flex items-center justify-center">
            <div className="flex shrink-0 w-[24px] h-[24px] absolute z-50 cursor-pointer">
                {isPlaying && (
                    <IconPlayerPauseFilled onClick={() => videoRef.pause()} />
                )}
                {!isPlaying && (
                    <IconPlayerPlayFilled onClick={() => videoRef.play()} />
                )}
            </div>
            <video
                ref={setVideoRef}
                autoPlay
                className="media-image animate-fade"
                poster={media.GetObjectUrl('thumbnail')}
                data-fit-logic={fitLogic}
                data-hide={
                    url === '' || media.HasLoadError() || !shouldShowVideo
                }
                style={imgStyle}
                onClick={() => {
                    if (isPlaying) {
                        videoRef.pause()
                    } else {
                        videoRef.play()
                    }
                }}
            >
                <source
                    src={media.StreamVideoUrl(authHeader)}
                    // type="video/mp4"
                />
            </video>
            <div
                className="flex absolute justify-center items-end p-3 pointer-events-none"
                style={{ width: size.width, height: size.height }}
            >
                <div className="flex h-2 w-9/12">
                    <WeblensProgress
                        value={
                            (playtime * 100) / (media.GetVideoLength() / 1000)
                        }
                        seekCallback={(v) => {
                            videoRef.currentTime =
                                (media.GetVideoLength() / 1000) * v
                        }}
                    />
                </div>
            </div>
        </div>
    )
}
