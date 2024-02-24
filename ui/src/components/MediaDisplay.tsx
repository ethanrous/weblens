import { memo, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { RawOn, Image, Folder, Theaters } from '@mui/icons-material'
import { VariableSizeList as List } from 'react-window'

import { Box, Loader, MantineStyleProp, Menu, MenuTarget, Text, Tooltip } from '@mantine/core'

import { MediaData, MediaWrapperProps, fileData } from '../types/Types'
import { ColumnBox, RowBox } from '../Pages/FileBrowser/FilebrowserStyles'
import { GetFileInfo } from '../api/FileBrowserApi'
import { useWindowSize } from './ItemScroller'
import { MediaImage } from './PhotoContainer'
import { userContext } from '../Context'
import { IconHome, IconTrash } from '@tabler/icons-react'
import { StyledLoaf } from './Crumbs'
import './galleryStyle.css'

const MultiFileMenu = ({ filesInfo, loading, menuOpen, setMenuOpen }: { filesInfo: fileData[], loading, menuOpen, setMenuOpen }) => {
    const [showLoader, setShowLoader] = useState(false)
    if (!menuOpen) {
        return null
    }

    if (loading) {
        setTimeout(() => setShowLoader(true), 150)
    }

    const FileRows = filesInfo.map(v => {
        const parts: any[] = v.pathFromHome.split("/")
        parts[0] = parts[0] === "HOME" ? <IconHome /> : <IconTrash />
        return StyledLoaf({ crumbs: parts })
    })

    return (
        <Menu opened={menuOpen && (showLoader || !loading)} onClose={() => setMenuOpen(false)}>
            <MenuTarget>
                <Box style={{ height: 0, width: 0 }} />
            </MenuTarget>

            <Menu.Dropdown style={{ minHeight: 80 }} onClick={e => e.stopPropagation()}>
                <Menu.Label>Multiple Files</Menu.Label>
                {loading && showLoader && (
                    <ColumnBox style={{ justifyContent: 'center', height: 40 }}>
                        <Loader color='white' size={20} />
                    </ColumnBox>
                )}
                {!loading && filesInfo.map((f, i) => {
                    return (
                        <Menu.Item key={f.id} onClick={e => { e.stopPropagation(); window.open(`/files/${f.parentFolderId}?jumpTo=${f.id}`, '_blank') }}>
                            {FileRows[i]}
                        </Menu.Item>
                    )
                })}
            </Menu.Dropdown>

        </Menu>
    )
}

const goToFolder = async (e, fileIds: string[], filesInfo, setLoading, setMenuOpen, setFileInfo, authHeader) => {
    e.stopPropagation()
    if (fileIds.length === 1) {
        const fileInfo: fileData = await GetFileInfo(fileIds[0], authHeader)
        window.open(`/files/${fileInfo.parentFolderId}?jumpTo=${fileInfo.id}`, '_blank')
        return
    }

    setMenuOpen(true)
    if (filesInfo.length === 0) {
        setLoading(true)
        const fileInfos = await Promise.all(fileIds.map(async v => await GetFileInfo(v, authHeader)))
        setFileInfo(fileInfos)
        setLoading(false)
    }

}

const PreviewCardContainer = ({ reff, setPresentation, setHover, onContextMenu, children, style }: { reff, setPresentation, setHover, onContextMenu?, children, style: MantineStyleProp }) => {
    return (
        <Box
            className='preview-card-container'
            ref={reff}
            children={children}
            onClick={setPresentation}
            onMouseOver={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
            onContextMenu={onContextMenu}
            style={style}
        />
    )
}

const TypeIcon = (mediaData: MediaData) => {
    let icon
    const name = mediaData.mediaType.FriendlyName
    if (mediaData.mediaType.IsRaw) {
        icon = RawOn
        // name = "RAW"
    } else if (mediaData.mediaType.IsVideo) {
        icon = Theaters
        // name = "Video"
    } else {
        icon = Image
        // name = "Image"
    }
    return [icon, name]
}

type mediaTypeProps = {
    Icon: any
    filesCount?: number
    onClick?: React.MouseEventHandler<HTMLDivElement>
    innerRef?: React.ForwardedRef<HTMLDivElement>
}

const StyledIcon = ({ Icon, filesCount, onClick, innerRef }: mediaTypeProps) => {
    return (

        <ColumnBox
            reff={innerRef}
            style={{ width: 'max-content', justifyContent: 'center', height: 'max-content' }}
            onClick={e => { e.stopPropagation(); if (onClick) { onClick(e) } }}
        >
            {Boolean(filesCount) && filesCount > 1 && (
                <Text c={'black'} size={'10px'} fw={700} style={{ position: 'absolute', userSelect: 'none' }}>{filesCount}</Text>
            )}
            <Icon
                className="meta-icon"
                style={{
                    cursor: onClick ? "pointer" : "default"
                }}
            />
        </ColumnBox>
    )
}

function MediaInfoDisplay({ mediaData, hovering }: { mediaData: MediaData, hovering: boolean }) {
    const { authHeader } = useContext(userContext)
    const [icon, name] = TypeIcon(mediaData)
    const [menuOpen, setMenuOpen] = useState(false)
    const [filesInfo, setFilesInfo] = useState([])
    const [loading, setLoading] = useState(false)

    if (!hovering && !menuOpen) {
        return null
    }

    return (
        <Box className='media-meta-preview'>
            <Tooltip label={name} refProp="innerRef">
                <StyledIcon Icon={icon} />
            </Tooltip>
            <RowBox style={{ height: 32 }}>
                <Tooltip label={mediaData.fileIds.length === 1 ? "Visit File" : "Visit Files"} refProp="innerRef">
                    <StyledIcon Icon={Folder} filesCount={mediaData.fileIds.length} onClick={e => goToFolder(e, mediaData.fileIds, filesInfo, setLoading, setMenuOpen, setFilesInfo, authHeader)} />
                </Tooltip>
                <MultiFileMenu filesInfo={filesInfo} loading={loading} menuOpen={menuOpen} setMenuOpen={setMenuOpen} />
            </RowBox>
        </Box>
    )
}

const MediaWrapper = memo(function MediaWrapper({ mediaData, scale, menu, dispatch }: MediaWrapperProps) {
    const ref = useRef()
    const [hovering, setHovering] = useState(false)
    const [menuOpen, setMenuOpen] = useState(false)
    const [menuPos, setMenuPos] = useState({ x: null, y: null })

    const width = useMemo(() => {
        mediaData.ImgRef = ref
        return mediaData.mediaWidth * (scale / mediaData.mediaHeight)
    }, [scale, mediaData])

    const filledMenu = useMemo(() => {
        if (menuOpen && menu) {
            return menu(mediaData.fileHash, menuOpen, setMenuOpen)
        } else {
            return null
        }
    }, [menuOpen, mediaData.fileHash, menu])

    return (
        <Box
            className='preview-card-container'
            ref={ref}
            onClick={() => { dispatch({ type: 'set_presentation', media: mediaData }) }}
            onMouseOver={() => setHovering(true)}
            onMouseLeave={() => setHovering(false)}
            onContextMenu={(e) => { e.stopPropagation(); e.preventDefault(); setMenuPos({ x: e.clientX, y: e.clientY }); setMenuOpen(true) }}
            style={{
                height: scale,
                width: width,
            }}
        >
            <MediaImage
                media={mediaData}
                quality={"thumbnail"}
                lazy={true}
                containerStyle={{ height: scale, width: '100%' }}
            />
            <MediaInfoDisplay mediaData={mediaData} hovering={hovering} />
            <Box style={{ position: 'fixed', top: menuPos.y, left: menuPos.x }}>
                {filledMenu}
            </Box>
        </Box >
    )
}, (prev: MediaWrapperProps, next: MediaWrapperProps) => {
    if (prev.scale !== next.scale) {
        return false
    }
    return (prev.mediaData.fileHash === next.mediaData.fileHash)
})
export const BucketCards = ({ medias, scale, dispatch, menu }: { medias, scale, dispatch, menu?: (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => JSX.Element }) => {
    if (!medias) {
        medias = []
    }

    const mediaCards = useMemo(() => {
        return medias.map((mediaData: MediaData) => {
            return (
                <MediaWrapper
                    key={mediaData.fileHash}
                    mediaData={mediaData}
                    scale={scale}
                    dispatch={dispatch}
                    menu={menu}
                />
            )
        }
        )
    }, [medias, scale, dispatch, menu])

    return (
        <RowBox style={{ justifyContent: 'center' }}>
            <RowBox style={{ height: scale + 4, width: '98%' }}>
                {mediaCards}
            </RowBox>
        </RowBox>
    )
}

const TitleWrapper = ({ bucketTitle }) => {
    if (bucketTitle === "") {
        return null
    }
    return (
        <Text style={{ fontSize: 20, fontWeight: 600 }} c={'white'} mt={1} pl={0.5}>
            {bucketTitle}
        </Text>
    )
}

export function PhotoGallery({ medias, imageBaseScale, title, dispatch }) {
    const listRef = useRef(null)
    const [, setWindowSize] = useState(null)
    const [boxNode, setBoxNode] = useState(null)
    const boxHeight = boxNode ? boxNode.clientHeight : 0
    const boxWidth = boxNode ? boxNode.clientWidth : 0
    useWindowSize(setWindowSize)

    const rows = useMemo(() => {
        const MARGIN_SIZE = 4

        if (medias.length === 0 || !boxWidth) {
            return []
        }

        const innerMedias = [...medias]

        const rows: { rowScale: number, items: MediaData[], element?: JSX.Element }[] = []
        let currentRowWidth = 0
        let currentRow = []

        while (true) {
            if (innerMedias.length === 0) {
                if (currentRow.length !== 0) {
                    rows.push({ rowScale: imageBaseScale, items: currentRow })
                }
                break
            }
            const m: MediaData = innerMedias.pop()
            // Calculate width given height "imageBaseScale", keeping aspect ratio
            const newWidth = Math.floor((imageBaseScale / m.mediaHeight) * m.mediaWidth) + MARGIN_SIZE

            // If we are out of media, and the image does not overflow this row, add it and break
            if (innerMedias.length === 0 && !(currentRowWidth + newWidth > boxWidth)) {
                currentRow.push(m)
                rows.push({ rowScale: imageBaseScale, items: currentRow })
                break
            }

            // If the image would overflow the window
            else if (currentRowWidth + newWidth > boxWidth) {
                const leftover = boxWidth - currentRowWidth
                let consuming = false
                if (newWidth / 2 < leftover || currentRow.length === 0) {
                    currentRow.push(m)
                    currentRowWidth += newWidth
                    consuming = true
                }
                const marginTotal = (currentRow.length * MARGIN_SIZE)
                rows.push({ rowScale: (boxWidth - marginTotal) / (currentRowWidth - marginTotal) * imageBaseScale, items: currentRow })
                currentRow = []
                currentRowWidth = 0

                if (consuming) {
                    continue
                }
            }
            currentRow.push(m)
            currentRowWidth += newWidth

        }
        rows.unshift({ rowScale: 10, items: [] })
        if (title) {
            rows.unshift({ rowScale: 75, items: [], element: title })
        }
        rows.push({ rowScale: 40, items: [] })
        return rows
    }, [medias, imageBaseScale, boxWidth])

    useEffect(() => {
        listRef.current?.resetAfterIndex(0)
    }, [boxWidth, imageBaseScale, medias.length])

    const Cell = useCallback(({ data, index, style }) => (
        <Box style={{ ...style, height: data[index].rowScale + 4 }}>
            {data[index].items.length > 0 && (
                <BucketCards key={data[index].items[0].fileHash} medias={data[index].items} scale={data[index].rowScale} dispatch={dispatch} />
            )}
            {data[index].element}
        </Box>
    ), [dispatch])

    const updateBox = useCallback(node => { if (!node) { return }; setBoxNode(node) }, [])

    return (
        <ColumnBox reff={updateBox}>
            <List
                className="no-scrollbars"
                ref={listRef}

                height={boxHeight}
                overscanRowCount={150}
                width={boxWidth}

                itemCount={rows.length}
                estimatedItemSize={imageBaseScale}

                itemSize={i => rows[i].rowScale + 4}
                itemData={rows}
            >
                {Cell}
            </List>
        </ColumnBox>
    )
}