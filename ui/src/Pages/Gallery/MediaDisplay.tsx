import { memo, useEffect, useRef, useState } from 'react'

import Box from '@mui/material/Box'
import Grid from '@mui/material/Grid'

import RawOnIcon from '@mui/icons-material/RawOn';
import ImageIcon from '@mui/icons-material/Image';
import TheatersIcon from '@mui/icons-material/Theaters';
import styled from '@emotion/styled'

import { MediaWrapperProps, GalleryBucketProps } from '../../types/GalleryTypes'
import { MediaThumbComponent } from '../../components/PhotoContainer'
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

const PreviewCardImg = styled(MediaThumbComponent)({
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

// const RecentlyViewedDisplay = ({ visible }) => {
//     if (visible) {
//         return (
//             <Box display={"absolute"}>
//                 Recently Viewed
//             </Box>
//         )
//     }
// }

const MediaWrapper = memo(function MediaWrapper({ mediaData, showIcons, presentingHash, dispatch }: MediaWrapperProps) {
    const [matching, setMatching] = useState(false)
    let ref = useRef()

    useEffect(() => {
        if (presentingHash == mediaData.FileHash) {
            dispatch({ type: 'set_presentation_ref', ref: ref })
            setMatching(true)
        } else if (presentingHash != mediaData.FileHash && presentingHash != "") {
            setMatching(false)
        } else if (matching && presentingHash == "") {
            setTimeout(() => setMatching(false), 3000)
        }
    }, [presentingHash])

    let height = 250
    let width = mediaData.ThumbWidth * (height / mediaData.ThumbHeight)

    return (
        <PreviewCardContainer
            ref={ref}
            minWidth={`clamp(150px, ${width}px, 100% - 8px)`}
        //width={width}
        >
            <RecentlyViewedDisplay key={"Naw"} display={matching ? "flex" : "none"} onClick={() => { }} >
                <Typography key={"Yah"} boxShadow={3}>
                    Just Viewed
                </Typography>
            </RecentlyViewedDisplay>
            <PreviewCardImg
                fileHash={mediaData.FileHash}
                blurhash={mediaData.BlurHash}
                onClick={() => dispatch({ type: 'set_presentation', presentingHash: mediaData.FileHash })}
            />
            <MediaInfoDisplay showIcons={showIcons} mediaData={mediaData} />

        </PreviewCardContainer>
    )
})

const BucketCards = ({ medias, showIcons, presentingHash, dispatch }) => {
    const mediaCards = medias.map((mediaData) => {
        return (
            <MediaWrapper
                key={mediaData.FileHash}
                mediaData={mediaData}
                showIcons={showIcons}
                presentingHash={presentingHash}
                dispatch={dispatch}
            />
        )
    }
    )

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

export const GalleryBucket = memo(function GalleryBucket({
    date,
    bucketData,
    showIcons,
    presentingHash,
    dispatch
}: GalleryBucketProps) {
    return (
        <Grid item >
            <DateWrapper dateTime={date} />
            <BucketCards medias={bucketData} showIcons={showIcons} presentingHash={presentingHash} dispatch={dispatch} />
        </Grid >

    )
})

