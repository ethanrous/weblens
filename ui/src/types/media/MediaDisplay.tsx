import {
    Icon,
    IconExclamationCircle,
    IconFolder,
    IconHeart,
    IconMovie,
    IconPhoto,
    IconPhotoScan,
} from '@tabler/icons-react'
import MediaApi from '@weblens/api/MediaApi'
import WeblensLoader from '@weblens/components/Loading.tsx'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import { useResize } from '@weblens/lib/hooks'
import { useGalleryStore } from '@weblens/pages/Gallery/GalleryLogic'
import { GalleryMenu } from '@weblens/pages/Gallery/GalleryMenu'
import '@weblens/pages/Gallery/galleryStyle.scss'
import {
    ErrorHandler,
    MediaWrapperProps,
    PresentType,
} from '@weblens/types/Types'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import React, {
    CSSProperties,
    MouseEvent,
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react'
import { useNavigate } from 'react-router-dom'
import { VariableSizeList } from 'react-window'

import { GalleryRowInfo, GalleryRowItem, GetMediaRows } from './MediaRows'

const goToMediaFile = async (mediaId: string) => {
    return MediaApi.getMediaFile(mediaId).then((r) => {
        const fileInfo = r.data
        const newUrl = `/files/${fileInfo.parentId}#${fileInfo.id}`
        window.open(newUrl, '_blank')
    })
}

const TypeIcon = (mediaData: WeblensMedia) => {
    let icon: Icon
    // let icon: ForwardRefExoticComponent<IconProps & RefAttributes<Icon>>

    if (mediaData.GetMediaType()?.IsRaw) {
        icon = IconPhotoScan
    } else if (mediaData.GetMediaType()?.IsVideo) {
        icon = IconMovie
    } else {
        icon = IconPhoto
    }
    return { icon, name: mediaData.GetMediaType()?.FriendlyName }
}

type mediaTypeProps = {
    Icon: Icon
    label: string
    visible: boolean

    onClick?: React.MouseEventHandler<Element>
}

function StyledIcon({ Icon, visible, onClick, label }: mediaTypeProps) {
    const [hover, setHover] = useState(false)
    const textRef = useRef<HTMLParagraphElement>(null)
    const textSize = useResize(textRef)

    const style = useMemo(() => {
        return {
            width: hover ? textSize.width + 33 : 28,
            cursor: onClick ? 'pointer' : 'default',
        }
    }, [hover, visible])

    const stopProp = useCallback((e: MouseEvent) => {
        e.stopPropagation()
        if (onClick) {
            onClick(e)
        }
    }, [])

    return (
        <div
            className="hover-icon"
            onMouseOver={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
            onClick={stopProp}
            style={style}
        >
            <Icon className="shrink-0" />
            <p
                className="pl-1 font-semibold text-nowrap text-white select-none"
                ref={textRef}
            >
                {label}
            </p>
        </div>
    )
}

const MediaInfoDisplay = ({
    mediaData,
    mediaMenuOpen,
    tooSmall,
}: {
    mediaData: WeblensMedia
    mediaMenuOpen: boolean
    tooSmall: boolean
}) => {
    const { user } = useSessionStore()
    // const mediaData = useMediaStore((state) => state.mediaMap.get(mediaId))
    const { icon, name } = useMemo(() => {
        return TypeIcon(mediaData)
    }, [])

    const setLiked = useMediaStore((state) => state.setLiked)
    const likedArray = mediaData.GetLikedBy()

    const visible = Boolean(icon) && !mediaMenuOpen && !tooSmall

    const liked = mediaData ? likedArray.includes(user.username) : false

    const othersLiked = likedArray.length - Number(liked) > 0
    let heartFill: string
    if (liked) {
        heartFill = 'red'
    } else if (othersLiked) {
        heartFill = 'white'
    } else {
        heartFill = 'transparent'
    }

    const goto = useCallback((e: MouseEvent) => {
        e.stopPropagation()
        goToMediaFile(mediaData.Id()).catch((e) => {
            console.error('Failed to go to media file', e)
        })
    }, [])

    return (
        <div className="media-meta-preview">
            <StyledIcon Icon={icon} label={name ?? ''} visible={visible} />

            {user.username === mediaData.GetOwner() && (
                <StyledIcon
                    Icon={IconFolder}
                    label="Visit File"
                    visible={visible}
                    onClick={goto}
                />
            )}
            <div
                className="hover-icon absolute right-0 bottom-0"
                data-show-anyway={liked || othersLiked}
                onClick={(e) => {
                    e.stopPropagation()
                    MediaApi.setMediaLiked(mediaData.Id(), !liked)
                        .then(() => {
                            setLiked(mediaData.Id(), user.username)
                        })
                        .catch(ErrorHandler)
                }}
            >
                <IconHeart
                    className="shrink-0"
                    fill={heartFill}
                    color={liked ? 'red' : 'white'}
                />
            </div>
        </div>
    )
}

const MARGIN_SIZE = 4

function MediaWrapper({
    mediaData,
    scale,
    width,
    showMedia,
}: MediaWrapperProps) {
    const ref = useRef<HTMLDivElement>(null)

    const presentingId = useGalleryStore((state) => state.presentingMediaId)
    const presentingMode = useGalleryStore((state) => state.presentingMode)
    const menuTargetId = useGalleryStore((state) => state.menuTargetId)
    const imageSize = useGalleryStore((state) => state.imageSize)
    const selecting = useGalleryStore((state) => state.selecting)
    const holdingShift = useGalleryStore((state) => state.holdingShift)
    const setMenuTarget = useGalleryStore((state) => state.setMenuTarget)
    const setPresentationTarget = useGalleryStore(
        (state) => state.setPresentationTarget
    )

    const hover = useMediaStore((state) => state.mediaMap.get(state.hoverId))
    const lastSelected = useMediaStore((state) =>
        state.mediaMap.get(state.lastSelectedId)
    )

    const user = useSessionStore((state) => state.user)
    const setHovering = useMediaStore((state) => state.setHovering)
    const setSelected = useMediaStore((state) => state.setSelected)

    const style = useMemo(() => {
        return {
            height: scale,
            width: width - MARGIN_SIZE,
        }
    }, [scale, mediaData, presentingId, presentingMode])

    const click = (selecting: boolean, holdingShift: boolean) => {
        if (selecting) {
            if (holdingShift) {
                console.error('media multi select not impl')
            }
            setSelected(mediaData.Id(), !mediaData.IsSelected())
            return
        }
        setPresentationTarget(mediaData.Id(), PresentType.Fullscreen)
    }

    const contextMenu = useCallback(
        (e: MouseEvent) => {
            e.stopPropagation()
            e.preventDefault()
            if (
                menuTargetId === mediaData.Id() ||
                mediaData.GetOwner() !== user.username
            ) {
                return
            }
            setMenuTarget(mediaData.Id())
        },
        [menuTargetId, style.width]
    )

    const choosing = useMemo(() => {
        return (
            selecting &&
            hover !== undefined &&
            lastSelected !== undefined &&
            holdingShift &&
            mediaData.GetAbsIndex() >= 0 &&
            (mediaData.GetAbsIndex() - lastSelected.GetAbsIndex()) *
                (mediaData.GetAbsIndex() - hover.GetAbsIndex()) <=
                0
        )
    }, [hover, lastSelected])

    const presenting =
        presentingId === mediaData.Id() && presentingMode === PresentType.InLine

    return (
        <div
            key={mediaData.Id()}
            className="preview-card-container animate-fade"
            data-selecting={selecting}
            data-selected={mediaData.IsSelected()}
            data-choosing={choosing}
            data-presenting={presenting}
            data-menu-open={menuTargetId === mediaData.Id()}
            ref={ref}
            onClick={() => click(selecting, holdingShift)}
            onMouseOver={() => {
                if (selecting) {
                    setHovering(mediaData.Id())
                }
            }}
            onMouseLeave={() => {
                if (selecting) {
                    setHovering('')
                }
            }}
            onContextMenu={contextMenu}
            style={style}
        >
            <MediaImage
                media={mediaData}
                quality={
                    presenting ? PhotoQuality.HighRes : PhotoQuality.LowRes
                }
                doFetch={showMedia}
                containerStyle={style}
            />

            {showMedia && mediaData && (
                <MediaInfoDisplay
                    mediaData={mediaData}
                    mediaMenuOpen={menuTargetId === mediaData.Id()}
                    tooSmall={imageSize < 200}
                />
            )}

            {menuTargetId === mediaData.Id() && (
                <GalleryMenu
                    media={mediaData}
                    open={menuTargetId === mediaData.Id()}
                    setOpen={(o: boolean) => {
                        if (o) {
                            setMenuTarget(mediaData.Id())
                        } else {
                            setMenuTarget('')
                        }
                    }}
                />
            )}
        </div>
    )
}

type bucketCardsProps = {
    medias: WeblensMedia[]
    widths: number[]
    index: number
    scale: number
    showMedia: boolean
}

export function BucketCards({
    medias,
    widths,
    index,
    scale,
    showMedia,
}: bucketCardsProps) {
    if (!medias) {
        medias = []
    }

    const placeholders = useMemo(() => {
        return medias.map((_, i) => {
            return (
                <div
                    key={`media-placeholder-${index}-${i}`}
                    className="m-[2px] bg-gray-900"
                    style={{ height: scale, width: widths[i] }}
                />
            )
        })
    }, [medias])

    const mediaCards = useMemo(() => {
        return medias.map((media: WeblensMedia, i: number) => {
            if (!showMedia) {
                media.CancelLoad()
                return placeholders[i]
            }

            return (
                <MediaWrapper
                    key={media.Id()}
                    mediaData={media}
                    scale={scale}
                    width={widths[i]}
                    showMedia={showMedia}
                    rowIndex={index}
                    colIndex={i}
                />
            )
        })
    }, [medias, showMedia])

    const style = useMemo(() => {
        return { height: scale + 4 }
    }, [scale])

    return (
        <div className="flex w-full" style={style}>
            {mediaCards}
        </div>
    )
}

function GalleryRow({
    data,
    index,
    style,
}: {
    data: GalleryRowInfo[]
    index: number
    style: CSSProperties
}) {
    const { medias, widths } = useMemo(() => {
        const medias: WeblensMedia[] = []
        const widths: number[] = []
        data[index].items.map((v: GalleryRowItem) => {
            medias.push(v.m)
            widths.push(v.w)
        })
        return { medias, widths }
    }, [data, index])

    return (
        <div className="flex justify-center pr-4 pl-4" style={style}>
            <div style={{ width: data[index].rowWidth }}>
                {data[index].items.length !== 0 && (
                    <BucketCards
                        key={data[index].items[0].m.Id()}
                        index={index}
                        medias={medias}
                        widths={widths}
                        scale={data[index].rowHeight}
                        showMedia={true}
                    />
                )}
            </div>
        </div>
    )
}

const NoMediaDisplay = () => {
    const nav = useNavigate()

    return (
        <div className="flex w-full flex-col items-center">
            <div className="mt-20 flex w-[300px] flex-col items-center gap-2">
                <h2 className="text-3xl font-bold select-none">
                    No media to display
                </h2>
                <p className="select-none">
                    Upload files or adjust the filters
                </p>
                <div className="h-max w-full gap-2">
                    <WeblensButton
                        squareSize={48}
                        fillWidth
                        label="FileBrowser"
                        Left={IconFolder}
                        onClick={() => nav('/files')}
                    />
                </div>
            </div>
        </div>
    )
}

export function PhotoGallery({
    medias,
    loading,
    error,
}: {
    medias: WeblensMedia[]
    loading: boolean
    error: Error
}) {
    const viewRef = useRef<HTMLDivElement>(null)
    const windowRef = useRef<VariableSizeList>(null)
    const viewSize = useResize(viewRef)

    const rowHeight = useGalleryStore((state) => state.imageSize)

    const showHidden = useMediaStore((state) => state.showHidden)

    const filteredMedias = useMemo(() => {
        if (!showHidden) {
            return medias.filter((m) => !m.IsHidden())
        }

        return medias
    }, [showHidden, medias])

    const rows: GalleryRowInfo[] = useMemo(() => {
        return GetMediaRows(
            filteredMedias,
            rowHeight,
            viewSize.width,
            MARGIN_SIZE
        )
    }, [filteredMedias, rowHeight, viewSize])

    useEffect(() => {
        if (windowRef.current) {
            windowRef.current.resetAfterIndex(0, true)
        }
    }, [rows])

    return (
        <div className="gallery-wrapper no-scrollbar" ref={viewRef}>
            {rows.length === 0 && !loading && !error && <NoMediaDisplay />}
            {loading && <WeblensLoader />}
            {error && (
                <div className="m-auto flex flex-row items-center gap-1 p-2 pb-40">
                    <IconExclamationCircle />
                    <h3>Failed to fetch media</h3>
                </div>
            )}
            {rows.length !== 0 && viewSize.width !== -1 && (
                <VariableSizeList
                    ref={windowRef}
                    height={viewSize.height}
                    width={viewSize.width}
                    itemSize={(i) => rows[i].rowHeight + MARGIN_SIZE}
                    itemCount={rows.length}
                    itemData={rows}
                >
                    {GalleryRow}
                </VariableSizeList>
            )}
        </div>
    )
}
