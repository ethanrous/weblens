import { forwardRef, memo, useContext, useMemo, useRef, useState } from 'react'
import { RawOn, Image, Folder, Theaters } from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'

import { MediaImage } from '../../components/PhotoContainer'
import { StyledBreadcrumb } from '../../components/Crumbs'
import { MediaData, MediaWrapperProps, GalleryBucketProps, fileData } from '../../types/Types'
import { Box, MantineStyleProp, Space, Text, Tooltip } from '@mantine/core'
import { FlexColumnBox, FlexRowBox } from '../FileBrowser/FilebrowserStyles'
import { notifications } from '@mantine/notifications'
import { GetFileInfo } from '../../api/FileBrowserApi'
import { userContext } from '../../Context'

// STYLES //

const Gallery = ({ ...props }) => {
    return (
        <Box
            {...props}
            style={{
                display: "flex",
                flexWrap: "wrap",
                alignItems: "center",
                position: "relative",
            }}
        />
    )
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
                flexGrow: 1,
                flexBasis: 0,
                margin: 2,
                position: "relative",
                overflow: "hidden",
                cursor: "pointer",
                ...style,
            }}
        />
    )
}

// Functions


const TypeIcon = (mediaData: MediaData) => {
    let icon
    let name
    if (mediaData.mediaType.IsRaw) {
        icon = RawOn
        name = "RAW"
    } else if (mediaData.mediaType.IsVideo) {
        icon = Theaters
        name = "Video"
    } else {
        icon = Image
        name = "Image"
    }
    return [icon, name]
}

// COMPONENTS //

type mediaTypeProps = {
    Icon: any
    ttText: string
    onClick?: React.MouseEventHandler<HTMLDivElement>
}

const StyledIcon = forwardRef((props: mediaTypeProps, ref) => {
    return (
        <Tooltip label={props.ttText} >
            <props.Icon
                onClick={props.onClick}
                sx={{
                    position: "static",
                    margin: '8px',
                    backgroundColor: "rgba(30, 30, 30, 0.5)",
                    borderRadius: "3px"
                }}
            />
        </Tooltip>
    )
})

const MediaInfoDisplay = ({ mediaData }: { mediaData: MediaData }) => {
    const nav = useNavigate()
    const {authHeader} = useContext(userContext)
    const [icon, name] = TypeIcon(mediaData)

    return (
        <FlexColumnBox style={{ height: "100%", width: "100%", justifyContent: 'space-between', alignItems: 'flex-start', position: 'absolute', transform: 'translateY(-100%)' }}>
            <StyledIcon Icon={icon} ttText={name} onClick={(e) => { e.stopPropagation() }} />
            <FlexRowBox style={{ height: 'max-content', justifyContent: 'space-between', alignItems: 'center', margin: "5px", bottom: 0, width: '95%', alignSelf: 'center' }}>
                {/* <StyledBreadcrumb label={mediaData.filename} fontSize={15} alwaysOn={true} doCopy /> */}
                <StyledIcon Icon={Folder} ttText={"Go To Folder"} onClick={async (e) => {
                    e.stopPropagation()
                    const fileInfo: fileData = await GetFileInfo(mediaData.fileId, authHeader)

                    if (!fileInfo.parentFolderId) {
                        console.error("File info:", fileInfo)
                        notifications.show({ title: "Failed to jump to folder", message: "Parent folder could not be identified", color: 'red' })
                        return
                    }

                    nav(`/files/${fileInfo.parentFolderId}`)
                }}
                />
            </FlexRowBox>
        </FlexColumnBox>
    )
}

const MediaWrapper = memo(function MediaWrapper({ mediaData, scale, scrollerRef, menu, dispatch }: MediaWrapperProps) {
    const ref = useRef()
    const [hovering, setHovering] = useState(false)
    const [menuOpen, setMenuOpen] = useState(false)
    const [menuPos, setMenuPos] = useState({ x: null, y: null })

    const width = useMemo(() => {
        mediaData.ImgRef = ref
        return mediaData.thumbWidth * (scale / mediaData.thumbHeight)
    }, [scale])

    const filledMenu = useMemo(() => {
        if (menuOpen) {
            return menu(mediaData.fileHash, menuOpen, setMenuOpen)
        } else {
            return null
        }
    }, [menuOpen])

    return (
        <PreviewCardContainer
            reff={ref}
            style={{
                height: scale,
                // minHeight: scale,
                width: width,
                minWidth: width * 0.85,
                maxWidth: width * 1.2
            }}
            setPresentation={() => { dispatch({ type: 'set_presentation', itemId: mediaData.fileHash }) }}
            setHover={setHovering}
            onContextMenu={(e) => { e.stopPropagation(); e.preventDefault(); setMenuPos({ x: e.clientX, y: e.clientY }); setMenuOpen(true) }}
        >
            <MediaImage
                mediaId={mediaData.fileHash}
                blurhash={mediaData.blurHash}
                quality={"thumbnail"}
                lazy={true}
                root={scrollerRef}
                imgStyle={{ objectFit: "cover", flexGrow: 1 }}
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
    if (prev.scale != next.scale) {
        return false
    }
    return (prev.mediaData.fileHash === next.mediaData.fileHash)
})
export const BucketCards = ({ medias, scale = 300, scrollerRef, dispatch, menu }: { medias, scale?, scrollerRef, dispatch, menu?: (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => JSX.Element }) => {
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
                    scrollerRef={scrollerRef}
                    dispatch={dispatch}
                    menu={menu}
                />
            )
        }
        )
    }, [medias, scale])

    return (
        <Gallery>
            {mediaCards}
        </Gallery >
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

export const GalleryBucket = ({
    bucketTitle,
    bucketData,
    scrollerRef,
    scale,
    dispatch
}: GalleryBucketProps) => {
    return (
        <Box style={{ width: '100%' }}>
            <Space h={"md"} />
            <TitleWrapper bucketTitle={bucketTitle} />
            <BucketCards medias={bucketData} scale={scale} scrollerRef={scrollerRef} dispatch={dispatch} />
        </Box>
    )
}

