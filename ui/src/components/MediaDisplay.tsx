import React, {
    memo,
    useCallback,
    useContext,
    useMemo,
    useRef,
    useState,
} from 'react'

import { Box, Loader, Menu, MenuTarget, Text } from '@mantine/core'

import {
    AlbumData,
    MediaWrapperProps,
    PresentType,
    SizeT,
    UserContextT,
} from '../types/Types'
import { FileInitT, WeblensFile } from '../classes/File'
import WeblensMedia from '../classes/Media'
import { GetFileInfo } from '../api/FileBrowserApi'
import { MediaImage } from './PhotoContainer'
import { UserContext } from '../Context'
import {
    IconFolder,
    IconPhoto,
    IconPhotoScan,
    IconTheater,
} from '@tabler/icons-react'
import { StyledLoaf } from './Crumbs'
import { useResize } from './hooks'
import { GalleryMenu } from '../Pages/Gallery/GalleryMenu'

import '../Pages/Gallery/galleryStyle.scss'
import { GalleryContext } from '../Pages/Gallery/Gallery'

const MultiFileMenu = ({
    filesInfo,
    loading,
    menuOpen,
    setMenuOpen,
}: {
    filesInfo: WeblensFile[]
    loading: boolean
    menuOpen: boolean
    setMenuOpen: (o: boolean) => void
}) => {
    const [showLoader, setShowLoader] = useState(false)
    if (!menuOpen) {
        return null
    }

    if (loading) {
        setTimeout(() => setShowLoader(true), 150)
    }

    const FileRows = filesInfo.map((v) => {
        return StyledLoaf({ crumbs: v.GetPathParts(), postText: '' })
    })

    return (
        <Menu
            opened={menuOpen && (showLoader || !loading)}
            onClose={() => setMenuOpen(false)}
        >
            <MenuTarget>
                <Box style={{ height: 0, width: 0 }} />
            </MenuTarget>

            <Menu.Dropdown
                style={{ minHeight: 80 }}
                onClick={(e) => e.stopPropagation()}
            >
                <Menu.Label>Multiple Files</Menu.Label>
                {loading && showLoader && (
                    <Box style={{ justifyContent: 'center', height: 40 }}>
                        <Loader color="white" size={20} />
                    </Box>
                )}
                {!loading &&
                    filesInfo.map((f, i) => {
                        return (
                            <Menu.Item
                                key={f.Id()}
                                onClick={(e) => {
                                    e.stopPropagation()
                                    window.open(
                                        `/files/${f.ParentId()}?jumpTo=${f.Id()}`,
                                        '_blank'
                                    )
                                }}
                            >
                                {FileRows[i]}
                            </Menu.Item>
                        )
                    })}
            </Menu.Dropdown>
        </Menu>
    )
}

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
        const fileInfo: FileInitT = await GetFileInfo(
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
    Icon: any
    label: string
    visible: boolean

    onClick?: React.MouseEventHandler<HTMLDivElement>
}

const StyledIcon = memo(
    ({ Icon, visible, onClick, label }: mediaTypeProps) => {
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
                <Text
                    fw={600}
                    ref={setTextRef}
                    style={{
                        paddingLeft: 5,
                        textWrap: 'nowrap',
                        userSelect: 'none',
                    }}
                >
                    {label}
                </Text>
            </div>
        )
    },
    (prev, next) => {
        return true
    }
)

const MediaInfoDisplay = memo(
    ({
        mediaData,
        mediaMenuOpen,
        tooSmall,
    }: {
        mediaData: WeblensMedia
        mediaMenuOpen: boolean
        tooSmall: boolean
    }) => {
        const { authHeader }: UserContextT = useContext(UserContext)
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
                    authHeader
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
    },
    (prev, next) => {
        return true
    }
)

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
        const [menuOpen, setMenuOpen] = useState(false)

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
                setMenuOpen(o)
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
            galleryState.presentingMedia,
            galleryState.presentingMode,
        ])

        const click = useCallback(() => {
            if (galleryState.selecting) {
                galleryDispatch({
                    type: 'set_selected',
                    mediaIndex: mediaData.GetAbsIndex(),
                    mediaId: mediaData.Id(),
                    selected: !mediaData.IsSelected(),
                })
                return
            }
            galleryDispatch({
                type: 'set_presentation',
                media: mediaData,
                presentMode: PresentType.Fullscreen,
            })
        }, [galleryState.selecting, galleryState.presentingMedia === mediaData])

        const mouseOver = useCallback(() => {
            if (galleryState.selecting) {
                galleryDispatch({
                    type: 'set_hover_target',
                    mediaIndex: mediaData.GetAbsIndex(),
                })
            }
        }, [galleryState.selecting])

        const mouseLeave = useCallback(() => {
            if (galleryState.selecting) {
                galleryDispatch({
                    type: 'set_hover_target',
                    mediaIndex: -1,
                })
            }
        }, [galleryState.selecting])

        const contextMenu = useCallback(
            (e) => {
                e.stopPropagation()
                e.preventDefault()
                if (menuOpen) {
                    return
                }
                galleryDispatch({
                    type: 'set_menu_target',
                    targetId: mediaData.Id(),
                })
                menuSwitch(true)
            },
            [menuOpen, style.width]
        )

        const mod = useMemo(() => {
            return {
                presenting:
                    galleryState.presentingMedia === mediaData &&
                    galleryState.presentingMode === PresentType.InLine,
                selecting: galleryState.selecting.toString(),
                choosing: (
                    galleryState.selecting &&
                    galleryState.hoverIndex !== -1 &&
                    galleryState.lastSelIndex !== -1 &&
                    galleryState.holdingShift &&
                    mediaData.GetAbsIndex() >= 0 &&
                    (mediaData.GetAbsIndex() - galleryState.lastSelIndex) *
                        (mediaData.GetAbsIndex() - galleryState.hoverIndex) <=
                        0
                ).toString(),
                selected: (!!mediaData.IsSelected()).toString(),
                'menu-open': menuOpen.toString(),
            }
        }, [
            galleryState.selecting,
            galleryState.presentingMedia === mediaData,
            galleryState.presentingMode,
            menuOpen,
            mediaData.IsSelected(),
            galleryState.hoverIndex,
            galleryState.holdingShift,
        ])

        return (
            <div
                key={`preview-card-container-${mediaData.Id()}`}
                className={`preview-card-container bg-gray-900 ${mediaData.Id()}`}
                data-selecting={mod.selecting}
                data-selected={mod.selected}
                data-choosing={mod.choosing}
                data-presenting={mod.presenting}
                ref={ref}
                onClick={click}
                onMouseOver={mouseOver}
                onMouseLeave={mouseLeave}
                onContextMenu={contextMenu}
                style={style}
            >
                <MediaImage
                    media={mediaData}
                    quality={mod.presenting ? 'fullres' : 'thumbnail'}
                    imageClass="gallery-image"
                    doFetch={showMedia}
                    containerStyle={style}
                />

                {showMedia && (
                    <MediaInfoDisplay
                        mediaData={mediaData}
                        mediaMenuOpen={menuOpen}
                        tooSmall={galleryState.imageSize < 200}
                    />
                )}

                <GalleryMenu
                    media={mediaData}
                    albumId={albumId}
                    open={menuOpen}
                    setOpen={menuSwitch}
                    updateAlbum={fetchAlbum}
                />
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
        if (prev.viewSize !== next.viewSize) {
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
    viewSize,
    albumId,
    fetchAlbum,
}: {
    medias: WeblensMedia[]
    widths: number[]
    index: number
    scale: number
    showMedia: boolean
    viewSize: SizeT
    albumId: string
    fetchAlbum: () => void
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
                    viewSize={viewSize}
                    albumId={albumId}
                    fetchAlbum={fetchAlbum}
                />
            )
        })
    }, [medias, showMedia])

    const style = useMemo(() => {
        return { height: scale + 4 }
    }, [scale])

    return (
        <Box className="flex justify-center">
            <Box className="flex w-full" style={style}>
                {mediaCards}
            </Box>
        </Box>
    )
}

type GalleryRow = {
    rowScale: number
    items: { m: WeblensMedia; w: number }[]
    element?: JSX.Element
}

const Cell = memo(
    ({
        data,
        showMedia,
        index,
    }: {
        data: {
            rows: GalleryRow[]
            selecting: boolean
            albumId: string
            viewSize: SizeT
            fetchAlbum: () => void
        }
        showMedia: boolean
        index: number
    }) => {
        return (
            <div className="z-1">
                {data.rows[index].items.length > 0 && (
                    <BucketCards
                        key={data.rows[index].items[0].m.Id()}
                        index={index}
                        medias={data.rows[index].items.map((v) => v.m)}
                        widths={data.rows[index].items.map((v) => v.w)}
                        scale={data.rows[index].rowScale}
                        showMedia={showMedia}
                        viewSize={data.viewSize}
                        albumId={data.albumId}
                        fetchAlbum={data.fetchAlbum}
                    />
                )}
                {data.rows[index].element}
            </div>
        )
    },
    (prev, next) => {
        if (prev.index !== next.index) {
            return false
        }
        if (prev.data !== next.data) {
            return false
        }
        if (prev.showMedia !== next.showMedia) {
            return false
        }
        return true
    }
)

const AlbumTitle = ({ startColor, endColor, title }) => {
    const sc = startColor ? `#${startColor}` : '#ffffff'
    const ec = endColor ? `#${endColor}` : '#ffffff'
    return (
        <Box style={{ height: 'max-content' }}>
            <Text
                size={'75px'}
                fw={900}
                variant="gradient"
                gradient={{
                    from: sc,
                    to: ec,
                    deg: 45,
                }}
                style={{
                    display: 'flex',
                    justifyContent: 'center',
                    userSelect: 'none',
                    lineHeight: 1.1,
                }}
            >
                {title}
            </Text>
        </Box>
    )
}

export function PhotoGallery({
    medias,
    album,
    fetchAlbum,
}: {
    medias: WeblensMedia[]
    album?: AlbumData
    fetchAlbum?: () => void
}) {
    const [scrollRef, setScrollRef] = useState(null)
    const [viewRef, setViewRef] = useState(null)
    const [scroll, setScroll] = useState(0)
    const [resizing, setResizing] = useState(null)
    const scrollSize = useResize(scrollRef)
    const viewSize = useResize(viewRef)
    const { galleryState } = useContext(GalleryContext)

    // useEffect(() => {
    //     if (resizing) {
    //         clearTimeout(resizing);
    //     }
    //     setResizing(
    //         setTimeout(() => {
    //             setResizing(null);
    //         }, 200)
    //     );
    // }, [scrollSize.width, galleryState.imageSize]);

    const rows: GalleryRow[] = useMemo(() => {
        if (medias.length === 0 || !scrollSize.width) {
            return []
        }

        const innerMedias = [...medias]

        const rows: GalleryRow[] = []
        let currentRowWidth = 0
        let currentRow: {
            m: WeblensMedia
            w: number
        }[] = []

        let absIndex = 0

        while (true) {
            if (innerMedias.length === 0) {
                if (currentRow.length !== 0) {
                    rows.push({
                        rowScale: galleryState.imageSize,
                        items: currentRow,
                    })
                }
                break
            }
            const m: WeblensMedia = innerMedias.pop()

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
                !(currentRowWidth + newWidth > scrollSize.width)
            ) {
                currentRow.push({ m: m, w: newWidth })
                rows.push({
                    rowScale: galleryState.imageSize,
                    items: currentRow,
                })
                break
            }

            // If the image will overflow the window
            else if (currentRowWidth + newWidth > scrollSize.width) {
                const leftover = scrollSize.width - currentRowWidth
                let consuming = false
                if (newWidth / 2 < leftover || currentRow.length === 0) {
                    currentRow.push({ m: m, w: newWidth })
                    currentRowWidth += newWidth
                    consuming = true
                }
                const marginTotal = currentRow.length * MARGIN_SIZE
                let rowScale =
                    ((scrollSize.width - marginTotal) /
                        (currentRowWidth - marginTotal)) *
                    galleryState.imageSize

                currentRow = currentRow.map((v) => {
                    v.w = v.w * (rowScale / galleryState.imageSize)
                    return v
                })
                rows.push({
                    rowScale: rowScale,
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

        if (album) {
            rows.unshift({
                rowScale: 75,
                items: [],
                element: (
                    <AlbumTitle
                        startColor={album.PrimaryColor}
                        endColor={album.SecondaryColor}
                        title={album.Name}
                    />
                ),
            })
        }
        // rows.push({ rowScale: 40, items: [] });
        return rows
    }, [medias, galleryState.imageSize, scrollSize.width, album])

    const data = useMemo(() => {
        return {
            rows: rows,
            selecting: galleryState.selecting,
            albumId: album?.Id,
            viewSize: viewSize,
            fetchAlbum: fetchAlbum,
        }
    }, [rows, galleryState.selecting, album, fetchAlbum])

    const rendered = useMemo(() => {
        return rows.map((r, i) => {
            return (
                <div
                    key={`fake-media-row-${i}`}
                    style={{ height: r.rowScale, visibility: 'hidden' }}
                />
            )
        })
    }, [rows.length])

    const { beforeCount, inViewCount } = useMemo(() => {
        let space = 0
        let beforeCount = 0
        for (; space < scroll; beforeCount++) {
            if (!rows[beforeCount]) {
                break
            }
            space += rows[beforeCount].rowScale
        }
        space = 0
        let inViewCount = 0

        for (; space < viewSize.height; inViewCount++) {
            if (!rows[beforeCount + inViewCount]) {
                break
            }
            space += rows[beforeCount + inViewCount].rowScale
        }

        return {
            beforeCount: beforeCount - 2,
            inViewCount: inViewCount,
        }
    }, [scroll, rows, viewSize.height])

    // Controls how many rows before and after the viewport are fully rendered
    const overScan = 10

    // Controls how many rows before and after the viewport are *partially* rendered
    // i.e. only the background block is shown per image, and the media does not load
    // until you reach the regular over-scan range, as above
    const overScanLite = 50

    return (
        <div
            className="gallery-wrapper no-scrollbar"
            ref={setViewRef}
            onScroll={(e) => {
                if (
                    Math.abs((e.target as HTMLElement).scrollTop - scroll) >
                    galleryState.imageSize
                ) {
                    setScroll((e.target as HTMLElement).scrollTop)
                }
            }}
        >
            <div ref={setScrollRef} className="gallery-scroll-box">
                <div className="h-4" />
                {data.rows.map((r, i) => {
                    if (
                        i < beforeCount - overScanLite ||
                        i > beforeCount + inViewCount + overScanLite
                    ) {
                        return rendered[i]
                    }

                    const show =
                        beforeCount - overScan <= i &&
                        i <= beforeCount + inViewCount + overScan

                    return (
                        <Cell
                            key={`media-row-${i}`}
                            data={data}
                            index={i}
                            showMedia={show && !resizing}
                        />
                    )
                })}
            </div>
        </div>
    )
}
