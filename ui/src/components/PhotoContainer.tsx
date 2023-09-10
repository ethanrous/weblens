import { Blurhash, BlurhashCanvas } from "react-blurhash";
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import { useState } from "react";
import Box from '@mui/material/Box';

import { styled } from "@mui/system";
import Grid from '@mui/material/Grid';
import RawOnIcon from '@mui/icons-material/RawOn';
import ImageIcon from '@mui/icons-material/Image';
import TheatersIcon from '@mui/icons-material/Theaters';

export const MediaComponent = ({
    fileHash,
    blurhash,
    ...props
}) => {
    const [imageLoaded, setImageLoaded] = useState(false)
    var url = new URL(`http:localhost:3000/api/item/${fileHash}`)
    url.searchParams.append('thumbnail', 'true')

    return (
        <Box
            height="100%"
            width="100%"
            display="grid"
            justifyItems="center"
            alignItems="center"
            alignContent="center"
            justifyContent="center"
            overflow="hidden"
        >
            <img
                {...props}
                src={url.toString()}
                //loading="lazy" // WHY DONT THIS WORK
                style={{ display: imageLoaded ? "block" : "none", opacity: 1 }}
                onLoad={() => setImageLoaded(true)}
            />
            {blurhash && !imageLoaded && (
                <Blurhash
                    height={250}
                    width={475}
                    hash={blurhash}
                />

            )}
        </Box>

    )
}

export const StyledLazyThumb = styled(MediaComponent)({
    position: "relative",

    objectFit: "cover",
    cursor: "pointer",
    overflow: "hidden",


    transitionDuration: "200ms",
    transitionProperty: "transform, box-shadow",
    transform: "scale3d(1.00, 1.00, 1)",
    "&:hover": {
        transitionProperty: "transform, box-shadow",
        transitionDuration: "200ms",
        transform: "scale3d(1.03, 1.03, 1)",
    }
});

const PhotoContainer = ({ mediaData, showIcons, dispatch }) => {

    const [contextMenu, setContextMenu] = useState<any>(null);

    const handleContextMenu = (event) => {
        event.preventDefault();
        setContextMenu(
            contextMenu === null
                ? {
                    mouseX: event.clientX + 2,
                    mouseY: event.clientY - 6,
                }
                : null,
        );
    };

    const handleClose = () => {
        setContextMenu(null);
    };

    const MediaInfoDisplay = () => {
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

    let height = 250
    let width = mediaData.ThumbWidth * (height / mediaData.ThumbHeight)
    let minWidth = (mediaData.ThumbWidth / mediaData.ThumbHeight) * height
    let maxWidth = mediaData.ThumbWidth - 33

    return (
        <Grid item
            flexGrow={1}
            flexBasis={0}
            margin={0.3}
            minWidth={`clamp(150px, ${minWidth}px, 100% - 8px)`}
            maxWidth={maxWidth}

            height={height}
            width={width}
        >
            <StyledLazyThumb fileHash={mediaData.FileHash} blurhash={mediaData.BlurHash} onClick={() => dispatch({ type: 'set_presentation', presentingHash: mediaData.FileHash })} width={"100%"} />
            <MediaInfoDisplay />
            <Menu
                open={contextMenu !== null}
                onClose={handleClose}
                anchorReference="anchorPosition"
                anchorPosition={
                    contextMenu !== null
                        ? { top: contextMenu.mouseY, left: contextMenu.mouseX }
                        : undefined
                }
            >
                <MenuItem onClick={handleClose}>Share</MenuItem>
                <MenuItem onClick={handleClose}>Hide</MenuItem>
                <MenuItem onClick={handleClose}>Favorite</MenuItem>
            </Menu>
        </Grid>
    )
}

export default PhotoContainer