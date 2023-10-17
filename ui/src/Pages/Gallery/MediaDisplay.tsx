import { memo, useEffect, useMemo, useRef, useState } from 'react'

import Box from '@mui/material/Box'
import Grid from '@mui/material/Grid'

import RawOnIcon from '@mui/icons-material/RawOn'
import ImageIcon from '@mui/icons-material/Image'
import TheatersIcon from '@mui/icons-material/Theaters'
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import styled from '@emotion/styled'

import { MediaWrapperProps, GalleryBucketProps } from '../../types/GalleryTypes'
import { MediaImage } from '../../components/PhotoContainer'
import { Tooltip } from '@mui/material'
import { MediaData } from '../../types/Generic'

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

const RecentlyViewedDisplay = styled(Box)({
    position: "absolute",
    height: "100%",
    width: "100%",
    outlineWidth: "5px",
    outlineStyle: "solid",
    outlineOffset: "-5px",
    outlineColor: "rgb(78 0 179 / 50%)",
    zIndex: 1,
    alignItems: "center",
    justifyContent: "center",
    pointerEvents: "none"
})

const StyledIconBox = styled(Box)({
    position: "absolute",
    top: 10,
    left: 10,
})

// Functions

function handleCopy(setCopied) {
    setCopied(true)
    setTimeout(() => setCopied(false), 1000)
}

// COMPONENTS //

const MediaInfoDisplay = ({ mediaData }) => {
    const [copied, setCopied] = useState(false)

    const TypeIcon = () => {
        if (mediaData.MediaType.IsRaw) {
            return (
                <StyledIconBox>
                    <RawOnIcon sx={{ bgcolor: "rgb(255, 255, 255, 0.40)", borderRadius: "3px" }} />
                </StyledIconBox>
            )
        } else if (mediaData.MediaType.IsVideo) {
            return (
                <StyledIconBox>
                    <TheatersIcon sx={{ bgcolor: "rgb(255, 255, 255, 0.40)", borderRadius: "3px" }} />
                </StyledIconBox>
            )
        } else {
            return (
                <StyledIconBox>
                    <ImageIcon sx={{ bgcolor: "rgb(255, 255, 255, 0.40)", borderRadius: "3px" }} />
                </StyledIconBox>
            )
        }
    }

    const filename = mediaData.Filepath.substring(mediaData.Filepath.lastIndexOf('/') + 1)

    return (
        <Box width={"100%"} height={"100%"}>
            <TypeIcon />
            <Box width={"95%"} height={"max-content"} bottom={10} left={10} position={"absolute"}>
                <StyledBreadcrumb label={copied ? "Copied" : filename} success={copied} onClick={e => { e.stopPropagation(); navigator.clipboard.writeText(filename); handleCopy(setCopied) }} />

                {/* <Crumbs path={mediaData.Filepath} includeHome={false} navigate={() => { }} /> */}
                {/* <Typography position={"absolute"} top={4} left={50}>{mediaData.Filepath}</Typography> */}
            </Box>
        </Box>
    )
}

const MediaWrapper = memo(function MediaWrapper({ mediaData, showIcons, dispatch }: MediaWrapperProps) {
    let ref = useRef()
    mediaData.ImgRef = ref

    let height = 250
    let width = mediaData.ThumbWidth * (height / mediaData.ThumbHeight)

    return (
        <PreviewCardContainer
            ref={ref}
            minWidth={`clamp(100px, ${width}px, 100% - 8px)`}
            maxWidth={`${width * 1.5}px`}
            onClick={() => dispatch({ type: 'set_presentation', presentingHash: mediaData.FileHash })}
        >
            <PreviewCardImg
                mediaData={mediaData}
                quality={"thumbnail"}
                lazy={true}
            />
            {showIcons && (
                <MediaInfoDisplay mediaData={mediaData} />
            )}

        </PreviewCardContainer>
    )
}, (prev: MediaWrapperProps, next: MediaWrapperProps) => {
    return (prev.mediaData.FileHash == next.mediaData.FileHash) && (prev.showIcons == next.showIcons)
})
const BucketCards = ({ medias, showIcons, dispatch }) => {
    const mediaCards = useMemo(() => {
        return medias.map((mediaData: MediaData) => {
            if (!mediaData.Display == true) {
                return null
            }
            return (
                <MediaWrapper
                    key={mediaData.FileHash}
                    mediaData={mediaData}
                    showIcons={showIcons}
                    dispatch={dispatch}
                />
            )
        }
        )
    }, [[...medias], showIcons])

    return (
        <Gallery>
            {mediaCards}
            <BlankCard />
        </Gallery>
    )
}

const DateWrapper = ({ dateTime }) => {
    const dateObj = new Date(dateTime)
    const dateString = dateObj.toUTCString().split(" 00:00:00 GMT")[0]

    return (
        <Box
            key={`${dateString} title`}
            component="h3"
            fontSize={25}
            pl={1}
            fontFamily={"Roboto,RobotoDraft,Helvetica,Arial,sans-serif"}
            marginBottom={1}
        >
            {dateString}
        </Box>
    )
}

//export const GalleryBucket = memo(function GalleryBucket({
export const GalleryBucket = ({
    date,
    bucketData,
    showIcons,
    dispatch
}: GalleryBucketProps) => {
    const anyDisplayed = () => {
        for (const val of bucketData) {
            if (val.Display) {
                return true
            }
        }
        return false
    }

    if (!anyDisplayed()) {
        return null
    }

    return (
        <Grid item >
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} showIcons={showIcons} dispatch={dispatch} />
        </Grid >
    )
}

