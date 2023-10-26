import { memo, useMemo, useRef, useState } from 'react'

import Box from '@mui/material/Box'
import Grid from '@mui/material/Grid'
import RawOnIcon from '@mui/icons-material/RawOn'
import ImageIcon from '@mui/icons-material/Image'
import FolderIcon from '@mui/icons-material/Folder';
import TheatersIcon from '@mui/icons-material/Theaters'
import styled from '@emotion/styled'

import { MediaImage } from '../../components/PhotoContainer'
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import { MediaData, MediaWrapperProps, GalleryBucketProps } from '../../types/Types'
import { IconButton, SvgIconTypeMap, Tooltip, Typography, alpha, useTheme } from '@mui/material'
import { useNavigate } from 'react-router-dom'
import { OverridableComponent } from '@mui/material/OverridableComponent'

// STYLES //

const Gallery = styled(Box)({
    display: "flex",
    flexWrap: "wrap",
    alignItems: "center",
    minHeight: "250px",
    position: "relative",
})

const BlankCard = styled("div")({
    height: '250px',
    flexGrow: 999999
})

const PreviewCardContainer = styled(Box)({
    height: "250px",
    borderRadius: "2px",
    flexGrow: 1,
    flexBasis: 0,
    margin: 2,
    position: "relative",
    overflow: "hidden",
    cursor: "pointer"
})

const PreviewCardImg = styled(MediaImage)({
    height: "250px",
    minWidth: "100%",
    position: "absolute",
    objectFit: "cover",

    transitionDuration: "300ms",
    transform: "scale3d(1.00, 1.00, 1)",
    "&:hover": {
        transitionDuration: "200ms",
        transform: "scale3d(1.03, 1.03, 1)",
    }
})

const StyledIconBox = styled(Box)({
    // position: "absolute",
    height: "24px",
    width: "24px",
    padding: "5px"
})

const StyledIcon = ({ element }) => {
    let TempStyled = styled(element)(({ theme }) => ({
        position: "static",
        color: theme.palette.primary.main,
        backgroundColor: alpha(theme.palette.background.default, 0.60),
        backdropFilter: "blur(8px)",
        borderRadius: "3px"
    }))
    return (<TempStyled />)
}

// Functions

// COMPONENTS //

type mediaTypeProps = {
    name: string,
    icon: OverridableComponent<SvgIconTypeMap<{}, "svg">> & { muiName: string; }
    onClick?: React.MouseEventHandler<HTMLDivElement>
}

const MediaType = ({ name, icon, onClick }: mediaTypeProps) => {
    return (
        <Tooltip title={name} disableInteractive>
            <StyledIconBox onClick={onClick}>
                <StyledIcon element={icon} />
            </StyledIconBox>
        </Tooltip>
    )
}

const MediaInfoDisplay = ({ mediaData }: { mediaData: MediaData }) => {
    const nav = useNavigate()

    const TypeIcon = () => {
        if (mediaData.MediaType.IsRaw) {
            return <MediaType name={"RAW"} icon={RawOnIcon} />
        } else if (mediaData.MediaType.IsVideo) {
            return <MediaType name={"Video"} icon={TheatersIcon} />
        } else {
            return <MediaType name={"Image"} icon={ImageIcon} />
        }
    }

    const filename = mediaData.Filepath.substring(mediaData.Filepath.lastIndexOf('/') + 1)

    return (
        <Box width={"100%"} height={"100%"} display={"flex"} flexDirection={'column'} justifyContent={"space-between"}>
            <TypeIcon />
            <Box display={"flex"} flexDirection={"row"} alignItems={"center"} justifyContent={"space-between"} width={"auto"} margin={"5px"} height={"max-content"} bottom={0}>
                <StyledBreadcrumb label={filename} doCopy />
                <MediaType name={"Go To Folder"} icon={FolderIcon} onClick={(e) => {
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
            ref={ref}
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
        <Typography fontSize={20} color={'primary.contrastText'} fontWeight={'bold'} mt={1} pl={0.5}>
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
        <Grid item >
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} dispatch={dispatch} />
        </Grid >
    )
}

