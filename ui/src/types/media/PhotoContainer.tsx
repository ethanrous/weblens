import { IconExclamationCircle, IconPhoto } from '@tabler/icons-react'
import WeblensLoader from '@weblens/components/Loading.tsx'
import { useVideo } from '@weblens/lib/hooks'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import React, {
    CSSProperties,
    MouseEvent,
    Suspense,
    useCallback,
    useEffect,
    useState,
} from 'react'

const VideoWrapper = React.lazy(async () => {
    const videoPlayer = await import('@weblens/components/media/VideoPlayer')
    return { default: videoPlayer.VideoWrapper }
})

export function MediaImage({
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
}) {
    if (!media) {
        media = new WeblensMedia({ contentId: '' })
    }

    const [loadError, setLoadErr] = useState('')
    const [src, setUrl] = useState({ url: '', id: media.Id() })
    const [videoRef, setVideoRef] = useState<HTMLVideoElement>()
    const { playtime, isPlaying, isWaiting } = useVideo(videoRef!)

    useEffect(() => {
        if (
            media.GetMediaType() &&
            doFetch &&
            media.Id() &&
            !media.HasQualityLoaded(quality)
        ) {
            media
                .LoadBytes(
                    quality,
                    pageNumber,
                    () => {
                        setUrl({
                            url: media.GetObjectUrl(quality, pageNumber),
                            id: media.Id(),
                        })
                        setLoadErr(media.HasLoadError())
                    },
                    () => {
                        setUrl({
                            url: media.GetObjectUrl(quality, pageNumber),
                            id: media.Id(),
                        })
                        setLoadErr(media.HasLoadError())
                    }
                )
                .catch((e) => {
                    console.error('Failed to get media bytes', e)
                })
        }

        if (!doFetch) {
            media.CancelLoad()
        } else if (
            (media.HasQualityLoaded(quality) && src.url === '') ||
            src.id !== media.Id()
        ) {
            setUrl({
                url: media.GetObjectUrl(quality, pageNumber),
                id: media.Id(),
            })
        } else if (
            media.HighestQualityLoaded() !== '' &&
            src.url === '' &&
            pageNumber === 0
        ) {
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
        media.GetMediaType().IsVideo! && quality === PhotoQuality.HighRes

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
                    <div className="absolute right-10 bottom-10 w-8">
                        <WeblensLoader />
                    </div>
                )}
            {src.url !== '' && (
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
            )}

            {shouldShowVideo && (
                <Suspense
                    fallback={
                        <div className="absolute right-10 bottom-10 w-8">
                            <WeblensLoader />
                        </div>
                    }
                >
                    <VideoWrapper
                        url={src.url}
                        shouldShowVideo={shouldShowVideo}
                        media={media}
                        fitLogic={fitLogic}
                        imgStyle={imgStyle ?? {}}
                        videoRef={videoRef!}
                        setVideoRef={setVideoRef}
                        isPlaying={isPlaying}
                        playtime={playtime}
                    />
                </Suspense>
            )}
        </div>
    )
}
