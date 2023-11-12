import { forwardRef, memo, useMemo, useRef, useState } from 'react'
import { Box, Tooltip, Typography, useTheme } from '@mui/joy'
import { RawOn, Image, Folder, Theaters } from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'

import { MediaImage } from '../../components/PhotoContainer'
import { StyledBreadcrumb } from '../../components/Crumbs'
import { MediaData, MediaWrapperProps, GalleryBucketProps } from '../../types/Types'
import { BlankCard } from '../../types/Styles'

// STYLES //

const Gallery = ({ ...props }) => {
    return (
        <Box
            {...props}
            sx={{
                display: "flex",
                flexWrap: "wrap",
                alignItems: "center",
                minHeight: "250px",
                position: "relative",
            }}
        />
    )
}

const PreviewCardContainer = ({ reff, ...props }) => {
    return (
        <Box
            ref={reff}
            {...props}
            sx={{
                height: "250px",
                borderRadius: "2px",
                flexGrow: 1,
                flexBasis: 0,
                margin: 0.2,
                position: "relative",
                overflow: "hidden",
                cursor: "pointer"
            }}
        />
    )
}

const PreviewCardImg = ({ mediaData, quality, lazy, ...props }) => {
    return (
        <MediaImage
            mediaData={mediaData}
            quality={quality}
            lazy={lazy}
            sx={{
                // height: "250px",
                minWidth: "100%",
                position: "absolute",
                objectFit: "cover",

                transitionDuration: "300ms",
                transform: "scale3d(1.00, 1.00, 1)",
                "&:hover": {
                    transitionDuration: "200ms",
                    transform: "scale3d(1.03, 1.03, 1)",
                }
            }}
        />
    )
}


const StyledIconBox = ({ onClick, ...props }) => {
    return (
        <Box
            {...props}
            onClick={onClick}
            sx={{
                height: "24px",
                width: "24px",
                padding: "5px"
            }}
        />
    )
}


// const StyledIcon = forwardRef((props: { Element: any, onClick: any }, ref) => {
//     let theme = useTheme()

//     return (
//         <props.Element
//             ref={ref}
//             {...props}
//             onClick={props.onClick}
//             sx={{
//                 position: "static",
//                 color: theme.palette.primary.mainChannel,
//                 backgroundColor: theme.palette.background.surface,
//                 backdropFilter: "blur(8px)",
//                 borderRadius: "3px"
//             }}

//         />
//     )
// })

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
    const theme = useTheme()
    return (
        <Tooltip title={props.ttText} disableInteractive>
            {/* <StyledIconBox onClick={onClick}> */}
            <props.Icon
                onClick={props.onClick}
                sx={{
                    position: "static",
                    margin: '8px',
                    color: theme.palette.primary.mainChannel,
                    backgroundColor: theme.palette.background.surface,
                    backdropFilter: "blur(8px)",
                    borderRadius: "3px"
                }}
            />
            {/* <StyledIcon Element={icon} onClick={onClick} /> */}
            {/* </StyledIconBox> */}
        </Tooltip>
    )
})

const MediaInfoDisplay = ({ mediaData }: { mediaData: MediaData }) => {
    const nav = useNavigate()
    const filename = mediaData.Filepath.substring(mediaData.Filepath.lastIndexOf('/') + 1)
    const [icon, name] = TypeIcon(mediaData)

    return (
        <Box width={"100%"} height={"100%"} display={"flex"} flexDirection={'column'} justifyContent={"space-between"}>
            <StyledIcon Icon={icon} ttText={name} onClick={(e) => { e.stopPropagation() }} />
            <Box display={"flex"} flexDirection={"row"} alignItems={"center"} justifyContent={"space-between"} width={"auto"} margin={"5px"} height={"max-content"} bottom={0}>
                <StyledBreadcrumb label={filename} doCopy />
                <StyledIcon Icon={Folder} ttText={"Go To Folder"} onClick={(e) => {
                    e.stopPropagation();
                    nav(`/files/${mediaData.Filepath.slice(0, mediaData.Filepath.lastIndexOf('/'))}`)
                }}
                />
            </Box>
        </Box>
    )
}

const MediaWrapper = memo(function MediaWrapper({ mediaData, dispatch }: MediaWrapperProps) {
    const ref = useRef()
    const [hovering, setHovering] = useState(false)
    mediaData.ImgRef = ref

    let height = 250
    let width = mediaData.ThumbWidth * (height / mediaData.ThumbHeight)

    return (
        <PreviewCardContainer
            reff={ref}
            minWidth={`clamp(100px, ${width}px, 100% - 8px)`}
            maxWidth={`${width * 1.5}px`}
            onMouseEnter={() => setHovering(true)}
            onMouseLeave={() => setHovering(false)}
            onClick={() => dispatch({ type: 'set_presentation', presentingHash: mediaData.FileHash })}
        >
            <PreviewCardImg
                mediaData={mediaData}
                quality={"thumbnail"}
                lazy={true}
            />
            {hovering && (
                <MediaInfoDisplay mediaData={mediaData} />
            )}

        </PreviewCardContainer>
    )
}, (prev: MediaWrapperProps, next: MediaWrapperProps) => {
    return (prev.mediaData.FileHash == next.mediaData.FileHash)
})
const BucketCards = ({ medias, dispatch }) => {
    const mediaCards = useMemo(() => {
        return medias.map((mediaData: MediaData) => {
            return (
                <MediaWrapper
                    key={mediaData.FileHash}
                    mediaData={mediaData}
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
        <Typography fontSize={20} color={'neutral'} fontWeight={'bold'} mt={1} pl={0.5}>
            {dateString}
        </Typography>
    )
}

export const GalleryBucket = ({
    date,
    bucketData,
    dispatch
}: GalleryBucketProps) => {
    return (
        <Box>
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} dispatch={dispatch} />
        </Box>
    )
}

