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
                minHeight: "250px",
                position: "relative",
            }}
        />
    )
}

const PreviewCardContainer = ({ reff, setPresentation, setHover, children, style }: { reff, setPresentation, setHover, children, style: MantineStyleProp }) => {
    return (
        <Box
            ref={reff}
            children={children}
            onClick={setPresentation}
            onMouseOver={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
            style={{
                ...style,
                height: "250px",
                borderRadius: "2px",
                flexGrow: 1,
                flexBasis: 0,
                margin: 1,
                position: "relative",
                overflow: "hidden",
                cursor: "pointer"
            }}
        />
    )
}

// Functions


const TypeIcon = (mediaData) => {
    let icon
    let name
    if (mediaData.MediaType.IsRaw) {
        icon = RawOn
        name = "RAW"
    } else if (mediaData.MediaType.IsVideo) {
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
            <FlexRowBox style={{ height: 'max-content', justifyContent: 'space-between', alignItems: 'center', margin: "5px", bottom: 0 }}>
                <StyledBreadcrumb label={mediaData.Filename} fontSize={15} alwaysOn={true} doCopy />
                <StyledIcon Icon={Folder} ttText={"Go To Folder"} onClick={(e) => {
                    e.stopPropagation()
                    nav(`/files/${mediaData.ParentFolder}`)
                }}
                />
            </FlexRowBox>
        </FlexColumnBox>
    )
}

const MediaWrapper = memo(function MediaWrapper({ mediaData, scrollerRef, dispatch }: MediaWrapperProps) {
    const ref = useRef()
    const [hovering, setHovering] = useState(false)
    mediaData.ImgRef = ref

    let height = 250
    let width = mediaData.ThumbWidth * (height / mediaData.ThumbHeight)

    return (
        <PreviewCardContainer
            reff={ref}
            style={{
                minWidth: `clamp(100px, ${width}px, 100% - 8px)`,
                maxWidth: `${width * 1.5}px`
            }}
            setPresentation={() => { dispatch({ type: 'set_presentation', presentingHash: mediaData.FileHash }) }}
            setHover={(input: boolean) => { console.log("HERE!", input); setHovering(input) }}
        >
            <MediaImage
                mediaData={mediaData}
                quality={"thumbnail"}
                lazy={true}
                root={scrollerRef}
                containerStyle={{ objectFit: "cover" }}
            />
            {hovering && (
                <MediaInfoDisplay mediaData={mediaData} />
            )}

        </PreviewCardContainer>
    )
}, (prev: MediaWrapperProps, next: MediaWrapperProps) => {
    return (prev.mediaData.FileHash == next.mediaData.FileHash)
})
const BucketCards = ({ medias, scrollerRef, dispatch }) => {
    const mediaCards = useMemo(() => {
        return medias.map((mediaData: MediaData) => {
            return (
                <MediaWrapper
                    key={mediaData.FileHash}
                    mediaData={mediaData}
                    scrollerRef={scrollerRef}
                    dispatch={dispatch}
                />
            )
        }
        )
    }, [[...medias]])

    return (
        <Gallery>
            {mediaCards}
            <BlankCard />
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
        <Box ml={"-1.6px"}>
            <Space h={"md"} />
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} scrollerRef={scrollerRef} dispatch={dispatch} />
        </Box>
    )
}

