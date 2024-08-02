import React, {
    memo,
    ReactElement,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react'

import { AlbumData, MediaWrapperProps, PresentType } from '../types/Types'
import { WeblensFileParams, WeblensFile } from '../Files/File'
import WeblensMedia, { MediaAction } from './Media'
import { GetFileInfo } from '../api/FileBrowserApi'
import { MediaImage } from './PhotoContainer'
import { MediaContext } from '../Context'
import {
    IconFolder,
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

    if (mediaData.GetMediaType().IsRaw) {
        icon = IconPhotoScan
    } else if (mediaData.GetMediaType().IsVideo) {
        icon = IconTheater
    } else {
        icon = IconPhoto
    }
    return [icon, mediaData.GetMediaType().FriendlyName]
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
            <Icon style={{ flexShrink: 0 }} />
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
    mediaData,
    mediaMenuOpen,
    tooSmall,
}: {
    mediaData: WeblensMedia
    mediaMenuOpen: boolean
    tooSmall: boolean
}) => {
    const auth = useSessionStore((state) => state.auth)
    const [icon, name] = useMemo(() => {
        return TypeIcon(mediaData)
    }, [])

    const [menuOpen, setMenuOpen] = useState(false)
    const [filesInfo, setFilesInfo] = useState([])

    const visible = Boolean(icon) && !mediaMenuOpen && !tooSmall

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

            <StyledIcon
                Icon={IconFolder}
                label="Visit File"
                visible={visible}
                onClick={goto}
            />
            {/* <Box style={{ height: 32 }}>
                 <MultiFileMenu
                 filesInfo={filesInfo}
                 loading={false}
                 menuOpen={menuOpen}
                 setMenuOpen={setMenuOpen}
                 />
                 </Box> */}
        </div>
    )
}

const MARGIN_SIZE = 4

const MediaWrapper = memo(
    ({
        mediaData,
        scale,
        width,
        showMedia,
        albumId,
        fetchAlbum,
    }: MediaWrapperProps) => {
        const ref = useRef()

        const { galleryState, galleryDispatch } = useContext(GalleryContext)
        const { mediaState, mediaDispatch } = useContext(MediaContext)
        const [hovering, setHovering] = useState(false)

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
                    const action: MediaAction = {
                        type: 'set_selected',
                        mediaId: mediaData.Id(),
                        selectMany: holdingShift,
                    }
                    mediaDispatch(action)
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
            setHovering(true)
            if (galleryState.selecting) {
                mediaDispatch({
                    type: 'set_hovering',
                    mediaId: mediaData.Id(),
                })
            }
        }, [galleryState.selecting])

        const mouseLeave = useCallback(() => {
            setHovering(false)
            if (galleryState.selecting) {
                mediaDispatch({
                    type: 'set_hovering',
                    mediaId: '',
                })
            }
        }, [galleryState.selecting])

        const contextMenu = useCallback(
            (e) => {
                e.stopPropagation()
                e.preventDefault()
                if (galleryState.menuTargetId === mediaData.Id()) {
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
                mediaState.getHovering() !== undefined &&
                mediaState.getLastSelected() !== undefined &&
                galleryState.holdingShift &&
                mediaData.GetAbsIndex() >= 0 &&
                (mediaData.GetAbsIndex() -
                    mediaState.getLastSelected().GetAbsIndex()) *
                    (mediaData.GetAbsIndex() -
                        mediaState.getHovering().GetAbsIndex()) <=
                    0
            )
        }, [galleryState, mediaState])

        const presenting =
            galleryState.presentingMediaId === mediaData.Id() &&
            galleryState.presentingMode === PresentType.InLine

        return (
            <div
                key={`preview-card-container-${mediaData.Id()}`}
                className={`preview-card-container ${mediaData.Id()}`}
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

                {hovering && showMedia && (
                    <MediaInfoDisplay
                        mediaData={mediaData}
                        mediaMenuOpen={
                            galleryState.menuTargetId === mediaData.Id()
                        }
                        tooSmall={galleryState.imageSize < 200}
                    />
                )}

                {galleryState.menuTargetId === mediaData.Id() && (
                    <GalleryMenu
                        media={mediaData}
                        albumId={albumId}
                        open={galleryState.menuTargetId === mediaData.Id()}
                        setOpen={menuSwitch}
                        updateAlbum={fetchAlbum}
                    />
                )}
            </div>
        )
    },
    (prev: MediaWrapperProps, next: MediaWrapperProps) => {
        if (prev.scale !== next.scale) {
            return false
        }
        if (prev.hoverIndex !== next.hoverIndex) {
            return false
        }
        if (prev.showMedia !== next.showMedia) {
            return false
        }
        if (prev.hoverIndex !== next.hoverIndex) {
            return false
        }

        return prev.mediaData.Id() === next.mediaData.Id()
    }
)
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
    }, [])
    return (
        <div className="flex justify-center" style={style}>
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

const AlbumTitle = ({ startColor, endColor, title }) => {
    const sc = startColor ? `#${startColor}` : '#447bff'
    const ec = endColor ? `#${endColor}` : '#6700ff'
    const style = {
        background: `linear-gradient(to right, ${sc}, ${ec}) text`,
    }
    return (
        <div className="flex h-max w-full justify-center">
            <h1
                className={`text-7xl font-extrabold select-none inline-block text-transparent `}
                style={style}
            >
                {title}
            </h1>
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

    const rows: GalleryRow[] = useMemo(() => {
        if (medias.length === 0 || viewSize.width === -1) {
            return []
        }

        const ROW_WIDTH = viewSize.width - 8

        const innerMedias = [...medias]

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
    }, [medias, galleryState.imageSize, viewSize, album])

    useEffect(() => {
        if (windowRef) {
            windowRef.resetAfterIndex(0)
        }
    }, [rows])

    // const onScroll = useCallback(
    //     (e: {
    //         scrollDirection: string
    //         scrollOffset: number
    //         scrollUpdateWasRequested: boolean
    //     }) => {
    //         if (e.scrollOffset) {
    //         }
    //     },
    //     []
    // )

    return (
        <div className="gallery-wrapper no-scrollbar" ref={setViewRef}>
            <div className="gallery-scroll-fade" />
            {viewSize.width !== -1 && (
                <WindowList
                    ref={setWindowRef}
                    height={viewSize.height}
                    width={viewSize.width}
                    itemSize={(i) => rows[i].rowScale + MARGIN_SIZE}
                    itemCount={rows.length}
                    itemData={rows}
                    overscan={20}
                    // onScroll={onScroll}
                >
                    {GalleryRow}
                </WindowList>
            )}
        </div>
    )
}
