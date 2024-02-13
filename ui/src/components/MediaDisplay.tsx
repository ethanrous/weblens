import { forwardRef, memo, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { RawOn, Image, Folder, Theaters } from '@mui/icons-material'
import { VariableSizeList as List } from 'react-window';

import { MediaImage } from './PhotoContainer'
import { MediaData, MediaWrapperProps, GalleryBucketProps, fileData } from '../types/Types'
import { Box, MantineStyleProp, Space, Text, Tooltip } from '@mantine/core'
import { ColumnBox, RowBox } from '../Pages/FileBrowser/FilebrowserStyles'
import { notifications } from '@mantine/notifications'
import { GetFileInfo } from '../api/FileBrowserApi'
import { userContext } from '../Context'
import { useWindowSize } from './ItemScroller';

const goToFolder = async (e, fileIds, authHeader) => {
    e.stopPropagation()
    if (fileIds.length > 1) {
        notifications.show({ title: "Failed to jump to folder", message: "This media has more than 1 file associated", color: 'red' })
        return
    }
    const fileInfo: fileData = await GetFileInfo(fileIds[0], authHeader)
    console.log(fileInfo)

    if (!fileInfo.parentFolderId) {
        console.error("File info:", fileInfo)
        notifications.show({ title: "Failed to jump to folder", message: "Parent folder could not be identified", color: 'red' })
        return
    }

    window.open(`/files/${fileInfo.parentFolderId}?jumpTo=${fileInfo.id}`, '_blank')
}

const PreviewCardContainer = ({ reff, setPresentation, setHover, onContextMenu, children, style }: { reff, setPresentation, setHover, onContextMenu?, children, style: MantineStyleProp }) => {
    return (
        <Box
            ref={reff}
            children={children}
            onClick={setPresentation}
            onMouseOver={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
            onContextMenu={onContextMenu}
            style={{
                borderRadius: "2px",
                margin: 2,
                position: "relative",
                overflow: "hidden",
                cursor: "pointer",
                ...style,
            }}
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
    ttText: string
    onClick?: React.MouseEventHandler<HTMLDivElement>
}

const StyledIcon = forwardRef((props: mediaTypeProps, ref) => {
    return (
        <Tooltip label={props.ttText} >
            <props.Icon
                className="meta-icon"
                onClick={e => { e.stopPropagation(); if (props.onClick) { props.onClick(e) } }}
                style={{
                    cursor: props.onClick ? "pointer" : "default"
                }}
            />
        </Tooltip>
    )
})

function MediaInfoDisplay({ mediaData }: { mediaData: MediaData }) {
    const { authHeader } = useContext(userContext)
    const [icon, name] = TypeIcon(mediaData)

    return (
        <Box className='media-meta-preview'>
            <StyledIcon Icon={icon} ttText={name} />
            <StyledIcon Icon={Folder} ttText={"Go To Folder"} onClick={e => goToFolder(e, mediaData.fileIds, authHeader)} />
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
        <PreviewCardContainer
            reff={ref}
            style={{
                height: scale,
                width: width,
            }}
            setPresentation={() => { dispatch({ type: 'set_presentation', media: mediaData }) }}
            setHover={setHovering}
            onContextMenu={(e) => { e.stopPropagation(); e.preventDefault(); setMenuPos({ x: e.clientX, y: e.clientY }); setMenuOpen(true) }}
        >
            <MediaImage
                media={mediaData}
                quality={"thumbnail"}
                lazy={true}
                containerStyle={{ height: scale, width: '100%' }}
            />
            {hovering && (
                <MediaInfoDisplay mediaData={mediaData} />
            )}
            <Box style={{ position: 'fixed', top: menuPos.y, left: menuPos.x }}>
                {filledMenu}
            </Box>
        </PreviewCardContainer>
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
        <RowBox style={{ height: scale + 4 }}>
            {mediaCards}
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

export function PhotoGallery({ medias, imageBaseScale, dispatch }) {
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

        const rows: { rowScale: number, items: MediaData[] }[] = []
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
        return rows
    }, [medias, imageBaseScale, boxWidth])

    useEffect(() => {
        listRef.current?.resetAfterIndex(0)
    }, [boxWidth, imageBaseScale, medias.length])

    const Cell = useCallback(({ data, index, style }) => (
        <Box style={{ ...style, height: data[index].rowScale + 4 }}>
            <BucketCards key={data[index].items[0].fileHash} medias={data[index].items} scale={data[index].rowScale} dispatch={dispatch} />
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


export const GalleryBucket = ({
    bucketTitle,
    bucketData,
    scale,
    dispatch
}: GalleryBucketProps) => {
    return (
        <Box style={{ width: '100%' }}>
            <Space h={"md"} />
            <TitleWrapper bucketTitle={bucketTitle} />
            <BucketCards medias={bucketData} scale={scale} dispatch={dispatch} />
        </Box>
    )
}

