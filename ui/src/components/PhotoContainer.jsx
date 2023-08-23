import { Blurhash } from "react-blurhash";
import Box from '@mui/material/Box';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import React, { useState, useEffect } from "react";


const ImageComponent = ({
    src,
    width,
    height,
    blurhash
}) => {
    const [imageLoaded, setImageLoaded] = useState(false)
    const loadFinished = () => setImageLoaded(true)

    return (
        <>
            <Box
                component="img"
                position="relative"

                src={`/api/thumbnail?filehash=${src}`}
                loading="lazy"
                onLoad={loadFinished}
                sx={{
                    width: "100%",
                    borderRadius: '7px',
                    objectFit: "cover",
                    overflow: 'hidden',
                }}
            />
            {!imageLoaded && (
                <Blurhash
                    hash={blurhash}
                    height={height}
                    width={width}
                    resolutionX={32}
                    resolutionY={32}
                />
            )}
        </>
    )
}

const PhotoContainer = ({ imageData }) => {

    const [contextMenu, setContextMenu] = React.useState(null);

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

    let newHeight = 350
    let newWidth = Math.floor(imageData.Thumbnail.Width * (newHeight / imageData.Thumbnail.Height))

    return (
        <Box
            onContextMenu={handleContextMenu}
            style={{ cursor: 'context-menu' }}
            sx={{
                display: "flex",
                flexGrow: 0.1,
                height: newHeight,
                position: "relative",
                margin: 1,
                zIndex: 1,
                transitionDuration: "200ms",
                transitionProperty: "transform",
                transform: "scale3d(1.00, 1.00, 1)",
                "&:hover": {
                    zIndex: 10,
                    transitionProperty: "transform",
                    transitionDuration: "200ms",
                    transform: "scale3d(1.03, 1.03, 1)",

                },
            }}
        >
            <ImageComponent src={imageData.FileHash} width={newWidth} height={newHeight} blurhash={imageData.BlurHash} />
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
        </Box>
    )
}

export default PhotoContainer