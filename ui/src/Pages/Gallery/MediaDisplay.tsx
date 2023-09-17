import { memo, useEffect, useMemo, useRef, useState } from 'react'

import Box from '@mui/material/Box'
import Grid from '@mui/material/Grid'

import RawOnIcon from '@mui/icons-material/RawOn';
import ImageIcon from '@mui/icons-material/Image';
import TheatersIcon from '@mui/icons-material/Theaters';
import styled from '@emotion/styled'

import { MediaWrapperProps, GalleryBucketProps } from '../../types/GalleryTypes'
import { MediaImage } from '../../components/PhotoContainer'
import { Typography } from '@mui/material';

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
    flexGrow: 1,
    flexBasis: 0,
    margin: 0.3,
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

// COMPONENTS //

const MediaInfoDisplay = ({ showIcons, mediaData }) => {
    if (!showIcons) {
        return
    }
    if (mediaData.MediaType.IsRaw) {
        return (
            <Box display="flex" justifyContent="flex-end" position="absolute" zIndex={1} sx={{ transform: "translate(1px, -20px)" }}>
                <RawOnIcon />
            </Box>
        )
    } else if (mediaData.MediaType.IsVideo) {
        return (

            <Box display="flex" justifyContent="flex-end" position="absolute" zIndex={1} sx={{ transform: "translate(2px, -26px)" }}>
                <TheatersIcon />
            </Box>
        )
    } else {
        return (
            <Box display="flex" justifyContent="flex-end" position="absolute" zIndex={1} sx={{ transform: "translate(2px, -26px)" }}>
                <ImageIcon />
            </Box>
        )
    }
}

const MediaWrapper = memo(function MediaWrapper({ mediaData, showIcons, dispatch }: MediaWrapperProps) {
    let ref = useRef()
    mediaData.ImgRef = ref

    let height = 250
    let width = mediaData.ThumbWidth * (height / mediaData.ThumbHeight)

    return (
        <PreviewCardContainer
            ref={ref}
            minWidth={`clamp(150px, ${width}px, 100% - 8px)`}
            onClick={() => dispatch({ type: 'set_presentation', presentingHash: mediaData.FileHash })}
        >
            <PreviewCardImg
                mediaData={mediaData}
                quality={"thumbnail"}
                lazy={true}
            />
            <MediaInfoDisplay showIcons={showIcons} mediaData={mediaData} />

        </PreviewCardContainer>
    )
}, (prev: MediaWrapperProps, next: MediaWrapperProps) => {
    return (prev.mediaData.FileHash == next.mediaData.FileHash)
})
const BucketCards = ({ medias, showIcons, dispatch }) => {
    const mediaCards = useMemo(() => {
        return medias.map((mediaData) => {
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
    }, [[...medias]])

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
    return (
        <Grid item >
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} showIcons={showIcons} dispatch={dispatch} />
        </Grid >

    )
}

