import { memo, useCallback, useContext, useEffect, useState } from 'react'
import { UserContext } from '../Context'
import { IconExclamationCircle, IconPhoto } from '@tabler/icons-react'
import { CSSProperties, Loader } from '@mantine/core'
import { UserContextT } from '../types/Types'
import WeblensMedia, { PhotoQuality } from '../classes/Media'

import './style.scss'

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
        const [loadError, setLoadErr] = useState('')
        const [src, setUrl] = useState({ url: '', id: media.Id() })
        const { authHeader }: UserContextT = useContext(UserContext)

        if (!media) {
            media = new WeblensMedia({ mediaId: '' })
        }

        useEffect(() => {
            if (doFetch && media.Id() && !media.HasQualityLoaded(quality)) {
                media.LoadBytes(
                    quality,
                    authHeader,
                    pageNumber,
                    () => {
                        setUrl({
                            url: media.GetImgUrl(quality),
                            id: media.Id(),
                        })
                        setLoadErr(media.HasLoadError())
                    },
                    () => {
                        setUrl({
                            url: media.GetImgUrl(quality),
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
                setUrl({ url: media.GetImgUrl(quality), id: media.Id() })
            }
        }, [media, quality, doFetch])

        const containerClick = useCallback(
            (e) => {
                preventClick && e.stopPropagation()
            },
            [preventClick]
        )

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
                    !loadError && (
                        <Loader
                            color="white"
                            bottom={40}
                            right={40}
                            size={20}
                            style={{ position: 'absolute' }}
                        />
                    )}

                <img
                    alt={'image'}
                    data-fit-logic={fitLogic}
                    data-disabled={disabled}
                    data-hide={src.url === '' || media.HasLoadError()}
                    className="media-image"
                    draggable={false}
                    src={src.url}
                    style={imgStyle}
                />

                {quality === 'fullres' && media.GetMediaType()?.IsVideo && (
                    <video src="" controls />
                )}

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
