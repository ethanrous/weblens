import { Blurhash, BlurhashCanvas } from "react-blurhash";
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import { useState } from "react";
import Box from '@mui/material/Box';

import { styled } from "@mui/system";
import Grid from '@mui/material/Grid';

const ImageComponent = ({
    src,
    blurhash,
    ctxMenu,
    width,
    ...props
}) => {
    const [imageLoaded, setImageLoaded] = useState(false)

    var url = new URL("http:localhost:3000/api/thumbnail");
    url.searchParams.append('filehash', src)

    return (
        <Box
            height={250}
            width="100%"
            justifyItems="center"
            alignItems="center"
            alignContent="center"
            justifyContent="center"
            display="grid"
            overflow="hidden"
        >
            <img
                {...props}
                src={url.toString()}
                //loading="lazy" // WHY DONT THIS WORK
                style={{ display: imageLoaded ? "block" : "none" }}
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

const PhotoContainer = ({ mediaData }) => {

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

    let height = 250
    let width = mediaData.ThumbWidth * (height / mediaData.ThumbHeight)
    let minWidth = (mediaData.ThumbWidth / mediaData.ThumbHeight) * 250


    const StyledLazyThumb = styled(ImageComponent)({
        width: "100%",
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

    return (
        <Grid item
            flexGrow={1}
            flexBasis={0}
            margin={0.3}
            minWidth={`clamp(124px, ${minWidth}px, 100% - 8px)`}

            height={height}
            width={width}
        >
            <StyledLazyThumb src={mediaData.FileHash} blurhash={mediaData.BlurHash} ctxMenu={handleContextMenu} width={width} />
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