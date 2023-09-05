import { Blurhash, BlurhashCanvas } from "react-blurhash";
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import { useState } from "react";
import Box from '@mui/material/Box';

import { styled } from "@mui/system";
import Grid from '@mui/material/Grid';
import RawOnIcon from '@mui/icons-material/RawOn';
import RawOffIcon from '@mui/icons-material/RawOff';

const ImageComponent = ({
    fileHash,
    blurhash,
    dispatch,
    ctxMenu,
    ...props
}) => {
    const [imageLoaded, setImageLoaded] = useState(false)


    var url = new URL(`http:localhost:3000/api/item/${fileHash}`)
    url.searchParams.append('thumbnail', 'true')
    //let img_data = `data:image/webp;base64,${src}`

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
                width="100%"

                src={url.toString()}
                //loading="lazy" // WHY DONT THIS WORK
                style={{ display: imageLoaded ? "block" : "none", opacity: 1 }}
                onLoad={() => setImageLoaded(true)}
                onClick={() => dispatch({ type: 'set_presentation', presentingHash: fileHash })}
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

const StyledLazyThumb = styled(ImageComponent)({
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

    const RawIcon = () => {
        if (!showIcons) {
            return
        }
        if (mediaData.MediaType.IsRaw) {
            return (
                <Box display="flex" justifyContent="flex-end" position="absolute" zIndex={1} sx={{ transform: "translate(1px, -20px)" }}>
                    <RawOnIcon />
                </Box>
            )
        } else {
            return (
                <Box display="flex" justifyContent="flex-end" position="absolute" zIndex={1} sx={{ transform: "translate(2px, -26px)" }}>
                    <RawOffIcon />
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
            //maxWidth={maxWidth}

            height={height}
            width={width}

        >
            <StyledLazyThumb fileHash={mediaData.FileHash} blurhash={mediaData.BlurHash} dispatch={dispatch} ctxMenu={handleContextMenu} />
            <RawIcon />
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