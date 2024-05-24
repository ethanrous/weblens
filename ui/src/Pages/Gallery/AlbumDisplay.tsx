import { Box, Text } from "@mantine/core";
import { MediaImage } from "../../components/PhotoContainer";
import { AlbumData, UserContextT } from "../../types/Types";

import { useContext, useMemo, useState } from "react";
import WeblensMedia, { PhotoQuality } from "../../classes/Media";
import { useNavigate } from "react-router-dom";
import { getAlbumPreview } from "../../api/ApiFetch";
import { UserContext } from "../../Context";
import { FixedSizeList } from "react-window";

import "./albumStyle.scss";
import { useResize } from "../../components/hooks";
import { GalleryContext } from "./Gallery";

export function AlbumPreview({ album }: { album: AlbumData }) {
    const { galleryDispatch } = useContext(GalleryContext);
    const nav = useNavigate();
    const { authHeader }: UserContextT = useContext(UserContext);
    const [coverM, setCoverM] = useState(album.CoverMedia);
    const [coverQuality, setCoverQ]: [
        coverQuality: PhotoQuality,
        setCoverQ: (v) => void
    ] = useState<PhotoQuality>("thumbnail");
    const [previewMedia, setPreviewMedia] = useState<WeblensMedia[]>(null);
    const fontSize = Math.floor(Math.pow(0.975, album.Name.length) * 40);

    return (
        <Box
            className="album-preview"
            onMouseOver={() => {
                setCoverQ("fullres");
                if (previewMedia === null && album.Medias.length !== 0) {
                    getAlbumPreview(album.Id, authHeader).then((r) =>
                        setPreviewMedia(
                            r.mediaIds.map((m) => {
                                return new WeblensMedia({ mediaId: m });
                            })
                        )
                    );
                }
            }}
            onMouseLeave={() => {
                setCoverQ("thumbnail");
            }}
            onClick={() => {
                nav(album.Id);
            }}
        >
            <Box
                className="cover-box"
                mod={{
                    "faux-album": album.Id === "",
                    "no-cover": coverM?.Id() === "",
                }}
                onContextMenu={(e) => {
                    e.stopPropagation();
                    e.preventDefault();
                    galleryDispatch({
                        type: "set_menu_pos",
                        pos: { x: e.clientX, y: e.clientY },
                    });
                    galleryDispatch({
                        type: "set_menu_target",
                        targetId: album.Id,
                    });
                    galleryDispatch({
                        type: "set_menu_open",
                        open: true,
                    });
                }}
            >
                <MediaImage
                    media={coverM}
                    quality={coverQuality}
                    imgStyle={{ zIndex: -1 }}
                    containerClass="cover-image"
                />
                <Text
                    truncate="end"
                    className="album-title-text"
                    size={`${fontSize}px`}
                >
                    {album.Name}
                </Text>
            </Box>
            <Box
                className="content-peek-wrapper"
                onMouseLeave={() => {
                    setCoverM(album.CoverMedia);
                }}
            >
                <Box className="content-peek-box">
                    {previewMedia?.map((m) => {
                        return (
                            <Box
                                key={m.Id()}
                                className="peek-image-container"
                                onMouseOver={() => {
                                    setCoverM(m);
                                }}
                            >
                                <MediaImage
                                    media={m}
                                    quality="thumbnail"
                                    containerClass="peek-image-container"
                                />
                            </Box>
                        );
                    })}
                </Box>
            </Box>
        </Box>
    );
}

function AlbumWrapper({
    data,
    index,
    style,
}: {
    data: { albums: AlbumData[]; colCount: number };
    index: number;
    style;
}) {
    const thisData = useMemo(() => {
        const thisData = data.albums.slice(
            index * data.colCount,
            index * data.colCount + data.colCount
        );

        while (thisData.length < data.colCount) {
            thisData.push({ Id: "", Name: "" } as AlbumData);
        }

        return thisData;
    }, [data, index]);

    return (
        <Box className="albums-row" style={style}>
            {thisData.map((a, i) => {
                if (a.Id !== "") {
                    return <AlbumPreview key={a.Id} album={a} />;
                } else {
                    return (
                        <Box key={`fake-album-${i}`} className="faux-album" />
                    );
                }
            })}
        </Box>
    );
}

const ALBUM_WIDTH = 350;

export function AlbumScroller({ albums }: { albums: AlbumData[] }) {
    const [containerRef, setContainerRef] = useState(null);
    const containerSize = useResize(containerRef);

    const colCount = Math.floor(containerSize.width / ALBUM_WIDTH);

    return (
        <Box ref={setContainerRef} className="albums-container">
            <FixedSizeList
                className="no-scrollbar"
                height={
                    containerSize.height >= 21 ? containerSize.height - 21 : 0
                }
                width={containerSize.width}
                itemSize={400}
                itemCount={
                    colCount !== 0 ? Math.ceil(albums.length / colCount) : 0
                }
                itemData={{
                    albums: albums,
                    colCount: colCount,
                }}
                overscanRowCount={5}
            >
                {AlbumWrapper}
            </FixedSizeList>
        </Box>
    );
}
