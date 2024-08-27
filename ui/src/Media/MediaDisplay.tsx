import React, {
    ReactElement,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react'

import { AlbumData, MediaWrapperProps, PresentType } from '../types/Types'
import { WeblensFile, WeblensFileParams } from '../Files/File'
import WeblensMedia from './Media'
import { GetFileInfo } from '../api/FileBrowserApi'
import { MediaImage } from './PhotoContainer'
import {
    IconFolder,
    IconHeart,
    IconPhoto,
    IconPhotoScan,
    IconTheater,
} from '@tabler/icons-react'
import { useResize } from '../components/hooks'
import { GalleryMenu } from '../Pages/Gallery/GalleryMenu'
import { VariableSizeList as WindowList } from 'react-window'

import '../Pages/Gallery/galleryStyle.scss'

import { GalleryContext } from '../Pages/Gallery/GalleryLogic'
import { useSessionStore } from '../components/UserInfo'
import { useMediaStore } from './MediaStateControl'
import { likeMedia } from './MediaQuery'
import { AlbumNoContent } from '../Albums/Albums'

const goToFolder = async (
    e,
    fileIds: string[],
    filesInfo,
    setLoading,
    setMenuOpen,
    setFileInfo,
    authHeader
) => {
    e.stopPropagation()
    if (fileIds.length === 1) {
        const fileInfo: WeblensFileParams = await GetFileInfo(
            fileIds[0],
            '',
            authHeader
        )

        const newFile = new WeblensFile(fileInfo)

        const newUrl = `/files/${newFile.ParentId()}?jumpTo=${fileIds[0]}`

        window.open(newUrl, '_blank')
        return
    }

    setMenuOpen(true)
    if (filesInfo.length === 0) {
        setLoading(true)
        const fileInfos = await Promise.all(
            fileIds.map(async (v) => await GetFileInfo(v, '', authHeader))
        )
        setFileInfo(fileInfos)
        setLoading(false)
    }
}

const TypeIcon = (mediaData: WeblensMedia) => {
    let icon

    if (mediaData.GetMediaType()?.IsRaw) {
        icon = IconPhotoScan
    } else if (mediaData.GetMediaType()?.IsVideo) {
        icon = IconTheater
    } else {
        icon = IconPhoto
    }
    return [icon, mediaData.GetMediaType()?.FriendlyName]
}

type mediaTypeProps = {
    Icon: (p) => ReactElement
    label: string
    visible: boolean

    onClick?: React.MouseEventHandler<HTMLDivElement>
}

function StyledIcon({ Icon, visible, onClick, label }: mediaTypeProps) {
    const [hover, setHover] = useState(false)
    const [textRef, setTextRef] = useState(null)
    const textSize = useResize(textRef)

    const style = useMemo(() => {
        return {
            width: hover ? textSize.width + 33 : 28,
            cursor: onClick ? 'pointer' : 'default',
        }
    }, [hover, visible])

    const stopProp = useCallback((e) => {
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
                className="font-semibold pl-1 text-nowrap select-none"
                ref={setTextRef}
            >
                {label}
            </p>
        </div>
    )
}

const MediaInfoDisplay = ({
    mediaId,
    mediaMenuOpen,
    tooSmall,
}: {
    mediaId: string
    mediaMenuOpen: boolean
    tooSmall: boolean
}) => {
    const { auth, user } = useSessionStore()
    const mediaData = useMediaStore((state) => state.mediaMap.get(mediaId))
    const [icon, name] = useMemo(() => {
        return TypeIcon(mediaData)
    }, [])

    const setLiked = useMediaStore((state) => state.setLiked)
    const likedArray = useMediaStore((state) =>
        state.mediaMap.get(mediaData.Id())?.GetLikedBy()
    )

    const [menuOpen, setMenuOpen] = useState(false)
    const [filesInfo, setFilesInfo] = useState([])

    const visible = Boolean(icon) && !mediaMenuOpen && !tooSmall

    const liked = useMediaStore((state) => {
        const m = state.mediaMap.get(mediaId)
        return m ? m.GetLikedBy().includes(user.username) : false
    })

    const othersLiked = likedArray.length - Number(liked) > 0
    let heartFill
    if (liked) {
        heartFill = 'red'
    } else if (othersLiked) {
        heartFill = 'white'
    } else {
        heartFill = 'transparent'
    }

    const goto = useCallback(
        (e) =>
            goToFolder(
                e,
                mediaData.GetFileIds(),
                filesInfo,
                () => {},
                setMenuOpen,
                setFilesInfo,
                auth
            ),
        []
    )

    return (
        <div className="media-meta-preview">
            <StyledIcon
                Icon={icon}
                label={name}
                visible={visible}
                onClick={null}
            />

            {user.username === mediaData.GetOwner() && (
                <StyledIcon
                    Icon={IconFolder}
                    label="Visit File"
                    visible={visible}
                    onClick={goto}
                />
            )}
            <div
                className="hover-icon absolute bottom-0 right-0"
                data-show-anyway={liked || othersLiked}
                onClick={(e) => {
                    e.stopPropagation()
                    likeMedia(mediaId, !liked, auth).then(() => {
                        console.log('HERE???')
                        setLiked(mediaId, user.username)
                    })
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

const MediaWrapper = ({
    mediaData,
    scale,
    width,
    showMedia,
    fetchAlbum,
}: MediaWrapperProps) => {
    const ref = useRef()

    const { galleryState, galleryDispatch } = useContext(GalleryContext)

    const hover = useMediaStore((state) => state.mediaMap.get(state.hoverId))
    const lastSelected = useMediaStore((state) =>
        state.mediaMap.get(state.lastSelectedId)
    )

    const user = useSessionStore((state) => state.user)
    const setHovering = useMediaStore((state) => state.setHovering)
    const setSelected = useMediaStore((state) => state.setSelected)

    const menuSwitch = useCallback(
        (o: boolean) => {
            if (o) {
                galleryDispatch({
                    type: 'set_menu_target',
                    targetId: mediaData.Id(),
                })
            } else {
                galleryDispatch({ type: 'set_menu_target', targetId: '' })
            }
        },
        [mediaData, galleryDispatch]
    )

    const style = useMemo(() => {
        // mediaData.SetImgRef(ref);
        return {
            height: scale,
            width: width - MARGIN_SIZE,
        }
    }, [
        scale,
        mediaData,
        galleryState.presentingMediaId,
        galleryState.presentingMode,
    ])

    const click = useCallback(
        (selecting: boolean, holdingShift: boolean) => {
            if (selecting) {
                if (holdingShift) {
                    console.error('Media multi select not impl')
                }
                setSelected(mediaData.Id(), !mediaData.IsSelected())
                return
            }
            galleryDispatch({
                type: 'set_presentation',
                mediaId: mediaData.Id(),
                presentMode: PresentType.Fullscreen,
            })
        },
        [mediaData.Id(), galleryState.presentingMediaId === mediaData.Id()]
    )

    const mouseOver = useCallback(() => {
        if (galleryState.selecting) {
            setHovering(mediaData.Id())
        }
    }, [galleryState.selecting])

    const mouseLeave = useCallback(() => {
        if (galleryState.selecting) {
            setHovering('')
        }
    }, [galleryState.selecting])

    const contextMenu = useCallback(
        (e) => {
            e.stopPropagation()
            e.preventDefault()
            if (
                galleryState.menuTargetId === mediaData.Id() ||
                mediaData.GetOwner() !== user.username
            ) {
                return
            }
            galleryDispatch({
                type: 'set_menu_target',
                targetId: mediaData.Id(),
            })
            menuSwitch(true)
        },
        [galleryState.menuTargetId, style.width]
    )

    const choosing = useMemo(() => {
        return (
            galleryState.selecting &&
            hover !== undefined &&
            lastSelected !== undefined &&
            galleryState.holdingShift &&
            mediaData.GetAbsIndex() >= 0 &&
            (mediaData.GetAbsIndex() - lastSelected.GetAbsIndex()) *
                (mediaData.GetAbsIndex() - hover.GetAbsIndex()) <=
                0
        )
    }, [galleryState, hover, lastSelected])

    const presenting =
        galleryState.presentingMediaId === mediaData.Id() &&
        galleryState.presentingMode === PresentType.InLine

    return (
        <div
            key={mediaData.Id()}
            className="preview-card-container"
            data-selecting={galleryState.selecting}
            data-selected={mediaData.IsSelected()}
            data-choosing={choosing}
            data-presenting={presenting}
            data-menu-open={galleryState.menuTargetId === mediaData.Id()}
            ref={ref}
            onClick={() =>
                click(galleryState.selecting, galleryState.holdingShift)
            }
            onMouseOver={mouseOver}
            onMouseLeave={mouseLeave}
            onContextMenu={contextMenu}
            style={style}
        >
            <MediaImage
                media={mediaData}
                quality={presenting ? 'fullres' : 'thumbnail'}
                doFetch={showMedia}
                containerStyle={style}
            />

            {showMedia && (
                <MediaInfoDisplay
                    mediaId={mediaData.Id()}
                    mediaMenuOpen={galleryState.menuTargetId === mediaData.Id()}
                    tooSmall={galleryState.imageSize < 200}
                />
            )}

            {galleryState.menuTargetId === mediaData.Id() && (
                <GalleryMenu
                    media={mediaData}
                    open={galleryState.menuTargetId === mediaData.Id()}
                    setOpen={menuSwitch}
                    updateAlbum={fetchAlbum}
                />
            )}
        </div>
    )
}
export const BucketCards = ({
    medias,
    widths,
    index,
    scale,
    showMedia,
}: {
    medias: WeblensMedia[]
    widths: number[]
    index: number
    scale: number
    showMedia: boolean
}) => {
    if (!medias) {
        medias = []
    }

    const placeholders = useMemo(() => {
        return medias.map((m, i) => {
            return (
                <div
                    key={`media-placeholder-${index}-${i}`}
                    className="bg-gray-900 m-[2px]"
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

type GalleryRowItem = { m: WeblensMedia; w: number }

type GalleryRow = {
    rowScale: number
    rowWidth: number
    items: GalleryRowItem[]
    element?: JSX.Element
}

function GalleryRow({ data, index, style }) {
    const { medias, widths } = useMemo(() => {
        const medias = []
        const widths = []
        data[index].items.map((v: GalleryRowItem) => {
            medias.push(v.m)
            widths.push(v.w)
        })
        return { medias, widths }
    }, [data])

    return (
        <div className="flex justify-center pl-4 pr-4" style={style}>
            <div style={{ width: data[index].rowWidth }}>
                {data[index].items.length !== 0 && (
                    <BucketCards
                        key={data[index].items[0].m.Id()}
                        index={index}
                        medias={medias}
                        widths={widths}
                        scale={data[index].rowScale}
                        showMedia={true}
                    />
                )}
            </div>
        </div>
    )
}

export function PhotoGallery({
    medias,
    album,
}: {
    medias: WeblensMedia[]
    album?: AlbumData
}) {
    const [viewRef, setViewRef] = useState(null)
    const [windowRef, setWindowRef] = useState(null)
    const viewSize = useResize(viewRef)
    const { galleryState } = useContext(GalleryContext)

    const showHidden = useMediaStore((state) => state.showHidden)

    const rows: GalleryRow[] = useMemo(() => {
        if (medias.length === 0 || viewSize.width === -1) {
            return []
        }

        const ROW_WIDTH = viewSize.width - 32

        let innerMedias = [...medias]

        let sortDirection = 1
        if (galleryState.albumId) {
            sortDirection = -1
        }
        if (!showHidden) {
            innerMedias = innerMedias.filter((m) => !m.IsHidden())
        }
        innerMedias.sort((m1, m2) => {
            const val =
                (m2.GetCreateTimestampUnix() - m1.GetCreateTimestampUnix()) *
                sortDirection
            // console.log(val)
            return val
        })
        innerMedias.forEach((m, i) => {
            if (i !== 0) {
                m.SetPrevLink(innerMedias[i - 1])
            } else {
                m.SetPrevLink(null)
            }
            if (i !== innerMedias.length - 1) {
                m.SetNextLink(innerMedias[i + 1])
            } else {
                m.SetNextLink(null)
            }
        })

        const rows: GalleryRow[] = []
        let currentRowWidth = 0
        let currentRow: GalleryRowItem[] = []

        let absIndex = 0

        while (true) {
            if (innerMedias.length === 0) {
                if (currentRow.length !== 0) {
                    rows.push({
                        rowScale: galleryState.imageSize,
                        rowWidth: ROW_WIDTH,
                        items: currentRow,
                    })
                }
                break
            }
            const m: WeblensMedia = innerMedias.shift()

            if (m.GetHeight() === 0) {
                console.error('Attempt to display media with 0 height:', m.Id())
                continue
            }

            m.SetAbsIndex(absIndex)
            absIndex++

            // Calculate width given height "imageBaseScale", keeping aspect ratio
            const newWidth =
                Math.floor(
                    (galleryState.imageSize / m.GetHeight()) * m.GetWidth()
                ) + MARGIN_SIZE

            // If we are out of media, and the image does not overflow this row, add it and break
            if (
                innerMedias.length === 0 &&
                !(currentRowWidth + newWidth > ROW_WIDTH)
            ) {
                currentRow.push({ m: m, w: newWidth })

                rows.push({
                    rowScale: galleryState.imageSize,
                    rowWidth: ROW_WIDTH,
                    items: currentRow,
                })
                break
            }

            // If the image will overflow the window
            else if (currentRowWidth + newWidth > ROW_WIDTH) {
                const leftover = ROW_WIDTH - currentRowWidth
                let consuming = false
                if (newWidth / 2 < leftover || currentRow.length === 0) {
                    currentRow.push({ m: m, w: newWidth })
                    currentRowWidth += newWidth
                    consuming = true
                }
                const marginTotal = currentRow.length * MARGIN_SIZE
                const rowScale =
                    ((ROW_WIDTH - marginTotal) /
                        (currentRowWidth - marginTotal)) *
                    galleryState.imageSize

                currentRow = currentRow.map((v) => {
                    v.w = v.w * (rowScale / galleryState.imageSize)
                    return v
                })

                rows.push({
                    rowScale: rowScale,
                    rowWidth: ROW_WIDTH,
                    items: currentRow,
                })
                currentRow = []
                currentRowWidth = 0

                if (consuming) {
                    continue
                }
            }
            currentRow.push({ m: m, w: newWidth })
            currentRowWidth += newWidth
        }
        rows.unshift({ rowScale: 20, rowWidth: ROW_WIDTH, items: [] })
        rows.push({ rowScale: 20, rowWidth: ROW_WIDTH, items: [] })
        return rows
    }, [medias, galleryState.imageSize, viewSize, album, showHidden])

    useEffect(() => {
        if (windowRef) {
            windowRef.resetAfterIndex(0, true)
        }
    }, [rows])

    return (
        <div className="gallery-wrapper no-scrollbar" ref={setViewRef}>
            <div className="gallery-scroll-fade" />
            {rows.length === 0 && (
                <AlbumNoContent hasContent={medias.length === 0} />
            )}
            {rows.length !== 0 && viewSize.width !== -1 && (
                <WindowList
                    ref={setWindowRef}
                    height={viewSize.height}
                    width={viewSize.width}
                    itemSize={(i) => rows[i].rowScale + MARGIN_SIZE}
                    itemCount={rows.length}
                    itemData={rows}
                    overscan={20}
                >
                    {GalleryRow}
                </WindowList>
            )}
        </div>
    )
}
