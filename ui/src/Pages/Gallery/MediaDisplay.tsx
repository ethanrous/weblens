import { forwardRef, memo, useMemo, useRef, useState } from 'react'
import { RawOn, Image, Folder, Theaters } from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'

import { MediaImage } from '../../components/PhotoContainer'
import { StyledBreadcrumb } from '../../components/Crumbs'
import { MediaData, MediaWrapperProps, GalleryBucketProps } from '../../types/Types'
import { BlankCard } from '../../types/Styles'
import { Box, MantineStyleProp, Space, Text, Tooltip } from '@mantine/core'
import { FlexColumnBox, FlexRowBox } from '../FileBrowser/FilebrowserStyles'

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

const PreviewCardContainer = ({ reff, setPresentation, setHover, onContextMenu, scale, children, style }: { reff, setPresentation, setHover, onContextMenu?, scale, children, style: MantineStyleProp }) => {
    return (
        <Box
            ref={reff}
            children={children}
            onClick={setPresentation}
            onMouseOver={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
            onContextMenu={onContextMenu}
            style={{
                height: scale,
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
    const [icon, name] = TypeIcon(mediaData)

    return (
        <FlexColumnBox alignOverride='flex-start' style={{ height: "100%", width: "100%", justifyContent: 'space-between', position: 'absolute', transform: 'translateY(-100%)' }}>
            <StyledIcon Icon={icon} ttText={name} onClick={(e) => { e.stopPropagation() }} />
            <FlexRowBox style={{ height: 'max-content', justifyContent: 'space-between', alignItems: 'center', margin: "5px", bottom: 0, width: '95%', alignSelf: 'center' }}>
                <StyledBreadcrumb label={mediaData.filename} fontSize={15} alwaysOn={true} doCopy />
                <StyledIcon Icon={Folder} ttText={"Go To Folder"} onClick={(e) => {
                    e.stopPropagation()
                    nav(`/files/${mediaData.parentFolder}`)
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
    mediaData.ImgRef = ref

    let width = mediaData.thumbWidth * (scale / mediaData.thumbHeight)

    const qualifiedMenu = useMemo(() => {
        if (menu) {
            return menu(mediaData.fileHash, menuOpen, setMenuOpen)
        }
        return null
    }, [menuOpen])

    return (
        <PreviewCardContainer
            reff={ref}
            style={{
                minWidth: `clamp(100px, ${width}px, 100% - 8px)`,
                maxWidth: `${width * 1.5}px`
            }}
            scale={scale}
            setPresentation={() => { dispatch({ type: 'set_presentation', itemId: mediaData.fileHash }) }}
            setHover={setHovering}
            onContextMenu={(e) => { e.preventDefault(); setMenuOpen(true) }}
        >
            <MediaImage
                mediaId={mediaData.fileHash}
                blurhash={mediaData.blurHash}
                quality={"thumbnail"}
                lazy={true}
                root={scrollerRef}
                imgStyle={{ objectFit: "cover" }}
                containerStyle={{ height: scale }}
            />
            {hovering && (
                <MediaInfoDisplay mediaData={mediaData} />
            )}
            {qualifiedMenu}
        </PreviewCardContainer>
    )
}, (prev: MediaWrapperProps, next: MediaWrapperProps) => {
    return (prev.mediaData.fileHash == next.mediaData.fileHash)
})
export const BucketCards = ({ medias, scrollerRef, dispatch, menu }: { medias, scrollerRef, dispatch, menu?: (mediaId: string, open: boolean, setOpen: (open: boolean) => void) => JSX.Element }) => {
    if (!medias) {
        medias = []
    }
    const scale = 250
    const mediaCards = useMemo(() => {
        // console.log(medias)
        // if (!medias) {
        //     return []
        // }
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
    }, [medias])

    return (
        <Gallery>
            {mediaCards}
            <BlankCard scale={scale} />
        </Gallery >
    )
}

const DateWrapper = ({ dateTime }) => {
    const dateObj = new Date(dateTime)
    const dateString = dateObj.toUTCString().split(" 00:00:00 GMT")[0]

    return (
        <Text style={{ fontSize: 20, fontWeight: 600 }} c={'white'} mt={1} pl={0.5}>
            {dateString}
        </Text>
    )
}

export const GalleryBucket = ({
    date,
    bucketData,
    scrollerRef,
    dispatch
}: GalleryBucketProps) => {
    return (
        <Box style={{ width: '100%' }}>
            <Space h={"md"} />
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} scrollerRef={scrollerRef} dispatch={dispatch} />
        </Box>
    )
}

