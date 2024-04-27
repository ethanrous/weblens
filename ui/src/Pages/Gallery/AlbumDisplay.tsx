import { Box, Text } from "@mantine/core";
import { MediaImage } from "../../components/PhotoContainer";
import { AlbumData } from "../../types/Types";

import "./albumStyle.scss";

export function AlbumPreview({ album }: { album: AlbumData }) {
    const fontSize = Math.floor(Math.pow(0.975, album.Name.length) * 35);
    console.log(fontSize);
    return (
        <Box className="album-preview">
            <MediaImage
                media={album.CoverMedia}
                quality="thumbnail"
                imgStyle={{ zIndex: -1 }}
            />
            <Text
                truncate="end"
                className="album-title-text"
                size={`${fontSize}px`}
            >
                {album.Name}
            </Text>
        </Box>
    );
}
