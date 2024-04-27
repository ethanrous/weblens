import {
    memo,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useRef,
    useState,
} from "react";
import { VariableSizeList as List } from "react-window";

import { Box, Loader, Menu, MenuTarget, Text, Tooltip } from "@mantine/core";

import {
    AlbumData,
    GalleryDispatchT,
    MediaWrapperProps,
    UserContextT,
} from "../types/Types";
import { WeblensFile } from "../classes/File";
import WeblensMedia from "../classes/Media";
import { ColumnBox, RowBox } from "../Pages/FileBrowser/FileBrowserStyles";
import { GetFileInfo } from "../api/FileBrowserApi";
import { MediaImage } from "./PhotoContainer";
import { MediaTypeContext, UserContext } from "../Context";
import {
    IconHome,
    IconTrash,
    IconFolder,
    IconPhoto,
    IconPhotoScan,
    IconUser,
    IconTheater,
} from "@tabler/icons-react";
import { StyledLoaf } from "./Crumbs";
import { useResize } from "./hooks";
import { GalleryAction } from "../Pages/Gallery/GalleryLogic";
import { GalleryMenu } from "../Pages/Gallery/GalleryMenu";

import "../Pages/Gallery/galleryStyle.scss";

const MultiFileMenu = ({
    filesInfo,
    loading,
    menuOpen,
    setMenuOpen,
}: {
    filesInfo: WeblensFile[];
    loading;
    menuOpen;
    setMenuOpen;
}) => {
    const [showLoader, setShowLoader] = useState(false);
    if (!menuOpen) {
        return null;
    }

    if (loading) {
        setTimeout(() => setShowLoader(true), 150);
    }

    const FileRows = filesInfo.map((v) => {
        return StyledLoaf({ crumbs: v.GetPathParts(), postText: "" });
    });

    return (
        <Menu
            opened={menuOpen && (showLoader || !loading)}
            onClose={() => setMenuOpen(false)}
        >
            <MenuTarget>
                <Box style={{ height: 0, width: 0 }} />
            </MenuTarget>

            <Menu.Dropdown
                style={{ minHeight: 80 }}
                onClick={(e) => e.stopPropagation()}
            >
                <Menu.Label>Multiple Files</Menu.Label>
                {loading && showLoader && (
                    <ColumnBox style={{ justifyContent: "center", height: 40 }}>
                        <Loader color="white" size={20} />
                    </ColumnBox>
                )}
                {!loading &&
                    filesInfo.map((f, i) => {
                        return (
                            <Menu.Item
                                key={f.Id()}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    window.open(
                                        `/files/${f.ParentId()}?jumpTo=${f.Id()}`,
                                        "_blank"
                                    );
                                }}
                            >
                                {FileRows[i]}
                            </Menu.Item>
                        );
                    })}
            </Menu.Dropdown>
        </Menu>
    );
};

const goToFolder = async (
    e,
    fileIds: string[],
    filesInfo,
    setLoading,
    setMenuOpen,
    setFileInfo,
    authHeader
) => {
    e.stopPropagation();
    if (fileIds.length === 1) {
        const fileInfo: WeblensFile = await GetFileInfo(
            fileIds[0],
            "",
            authHeader
        );

        const newUrl = `/files/${fileInfo.ParentId()}?jumpTo=${fileIds[0]}`;

        window.open(newUrl, "_blank");
        return;
    }

    setMenuOpen(true);
    if (filesInfo.length === 0) {
        setLoading(true);
        const fileInfos = await Promise.all(
            fileIds.map(async (v) => await GetFileInfo(v, "", authHeader))
        );
        setFileInfo(fileInfos);
        setLoading(false);
    }
};

const TypeIcon = (mediaData: WeblensMedia) => {
    const typeMap = useContext(MediaTypeContext);
    let icon;

    if (mediaData.GetMediaType(typeMap).IsRaw) {
        icon = IconPhotoScan;
    } else if (mediaData.GetMediaType(typeMap).IsVideo) {
        <IconTheater />;
    } else {
        icon = IconPhoto;
    }
    return [icon, mediaData.GetMediaType(typeMap).FriendlyName];
};

type mediaTypeProps = {
    Icon: any;
    label: string;
    visible: boolean;
    filesCount?: number;
    onClick?: React.MouseEventHandler<HTMLDivElement>;
    innerRef?: React.ForwardedRef<HTMLDivElement>;
};

const StyledIcon = ({
    Icon,
    filesCount,
    visible,
    onClick,
    innerRef,
    label,
}: mediaTypeProps) => {
    const [hover, setHover] = useState(false);
    const [textRef, setTextRef] = useState(null);
    const textSize = useResize(textRef);

    return (
        <Box
            className="hover-icon"
            onMouseOver={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
            onClick={onClick}
            mod={{ visible: visible }}
            style={{ width: hover ? textSize.width + 33 : 28 }}
        >
            <Icon style={{ flexShrink: 0 }} />
            <Text
                fw={600}
                ref={setTextRef}
                style={{
                    paddingLeft: 5,
                    textWrap: "nowrap",
                    userSelect: "none",
                }}
            >
                {label}
            </Text>
        </Box>
    );

    return (
        <ColumnBox
            reff={innerRef}
            style={{
                width: "max-content",
                justifyContent: "center",
                height: "max-content",
            }}
            onClick={(e) => {
                e.stopPropagation();
                if (onClick) {
                    onClick(e);
                }
            }}
        >
            {Boolean(filesCount) && filesCount > 1 && (
                <Text
                    c={"black"}
                    size={"10px"}
                    fw={700}
                    style={{ position: "absolute", userSelect: "none" }}
                >
                    {filesCount}
                </Text>
            )}
            <Icon
                className="meta-icon"
                style={{
                    cursor: onClick ? "pointer" : "default",
                }}
            />
        </ColumnBox>
    );
};

function MediaInfoDisplay({
    mediaData,
    mediaMenuOpen,
    tooSmall,
    hovering,
}: {
    mediaData: WeblensMedia;
    mediaMenuOpen: boolean;
    tooSmall: boolean;
    hovering: boolean;
}) {
    const { authHeader }: UserContextT = useContext(UserContext);
    const [icon, name] = TypeIcon(mediaData);
    const [menuOpen, setMenuOpen] = useState(false);
    const [filesInfo, setFilesInfo] = useState([]);

    const visible = hovering && Boolean(icon) && !mediaMenuOpen && !tooSmall;

    return (
        <Box className="media-meta-preview">
            <StyledIcon
                Icon={icon}
                label={name}
                visible={visible}
                onClick={(e) => e.stopPropagation()}
            />

            <StyledIcon
                Icon={IconFolder}
                label="Visit File"
                filesCount={mediaData.GetFileIds().length}
                visible={visible}
                onClick={(e) =>
                    goToFolder(
                        e,
                        mediaData.GetFileIds(),
                        filesInfo,
                        () => {},
                        setMenuOpen,
                        setFilesInfo,
                        authHeader
                    )
                }
            />
            {/* <RowBox style={{ height: 32 }}>
                <MultiFileMenu
                    filesInfo={filesInfo}
                    loading={loading}
                    menuOpen={menuOpen}
                    setMenuOpen={setMenuOpen}
                />
            </RowBox> */}
        </Box>
    );
}

const MediaWrapper = memo(
    ({
        mediaData,
        selected,
        selecting,
        scale,
        albumId,
        fetchAlbum,
        dispatch,
    }: MediaWrapperProps) => {
        const ref = useRef();
        const [hovering, setHovering] = useState(false);
        const [menuOpen, setMenuOpen] = useState(false);

        const width = useMemo(() => {
            mediaData.SetImgRef(ref);
            return mediaData.GetWidth() * (scale / mediaData.GetHeight());
        }, [scale, mediaData]);

        if (mediaData.IsSelected() === undefined) {
            mediaData.SetSelected(false);
        }

        return (
            <Box
                className="preview-card-container"
                mod={{
                    "data-selecting": selecting.toString(),
                    "data-selected": mediaData.IsSelected().toString(),
                    "data-menu-open": menuOpen.toString(),
                }}
                ref={ref}
                onClick={() => {
                    if (selecting) {
                        dispatch({
                            type: "set_selected",
                            mediaId: mediaData.Id(),
                            selected: !mediaData.IsSelected(),
                        });
                        return;
                    }
                    dispatch({ type: "set_presentation", media: mediaData });
                }}
                onMouseOver={() => setHovering(true)}
                onMouseLeave={() => setHovering(false)}
                onContextMenu={(e) => {
                    e.stopPropagation();
                    e.preventDefault();
                    if (menuOpen || scale < 200) {
                        return;
                    }
                    setMenuOpen(true);
                    dispatch({
                        type: "set_menu_target",
                        targetId: mediaData.Id(),
                    });
                }}
                style={{
                    height: scale,
                    width: width,
                }}
            >
                <MediaImage
                    media={mediaData}
                    quality={"thumbnail"}
                    lazy={true}
                    containerStyle={{
                        borderRadius: 4,
                        height: scale,
                        width: "100%",
                    }}
                />
                <MediaInfoDisplay
                    mediaData={mediaData}
                    mediaMenuOpen={menuOpen}
                    tooSmall={scale < 200}
                    hovering={hovering && !selecting}
                />
                <GalleryMenu
                    media={mediaData}
                    albumId={albumId}
                    open={menuOpen}
                    setOpen={setMenuOpen}
                    height={scale}
                    width={width}
                    updateAlbum={fetchAlbum}
                />
            </Box>
        );
    },
    (prev: MediaWrapperProps, next: MediaWrapperProps) => {
        if (prev.scale !== next.scale) {
            return false;
        }
        if (prev.selecting !== next.selecting) {
            return false;
        }
        if (prev.selected !== next.selected) {
            return false;
        }
        return prev.mediaData.Id() === next.mediaData.Id();
    }
);
export const BucketCards = ({
    medias,
    selecting = false,
    scale,
    albumId,
    fetchAlbum,
    dispatch,
}: {
    medias: WeblensMedia[];
    selecting?: boolean;
    scale: number;
    albumId: string;
    fetchAlbum: () => void;
    dispatch: GalleryDispatchT;
}) => {
    if (!medias) {
        medias = [];
    }

    const mediaCards = medias.map((media: WeblensMedia) => {
        return (
            <MediaWrapper
                key={media.Id()}
                mediaData={media}
                selected={media.IsSelected()}
                selecting={selecting}
                scale={scale}
                albumId={albumId}
                fetchAlbum={fetchAlbum}
                dispatch={dispatch}
            />
        );
    });

    return (
        <RowBox style={{ justifyContent: "center" }}>
            <RowBox style={{ height: scale + 4, width: "98%" }}>
                {mediaCards}
            </RowBox>
        </RowBox>
    );
};

type GalleryRow = {
    rowScale: number;
    items: WeblensMedia[];
    element?: JSX.Element;
};

const Cell = ({
    data,
    index,
    style,
}: {
    data: {
        rows: GalleryRow[];
        selecting: boolean;
        albumId: string;
        fetchAlbum: () => void;
        dispatch: (action: GalleryAction) => void;
    };
    index: number;
    style;
}) => {
    return (
        <Box style={{ ...style }}>
            {data.rows[index].items.length > 0 && (
                <BucketCards
                    key={data.rows[index].items[0].Id()}
                    selecting={data.selecting}
                    medias={data.rows[index].items}
                    scale={data.rows[index].rowScale}
                    albumId={data.albumId}
                    fetchAlbum={data.fetchAlbum}
                    dispatch={data.dispatch}
                />
            )}
            {data.rows[index].element}
        </Box>
    );
};

const AlbumTitle = ({ startColor, endColor, title }) => {
    const sc = startColor ? `#${startColor}` : "#ffffff";
    const ec = endColor ? `#${endColor}` : "#ffffff";
    return (
        <ColumnBox style={{ height: "max-content" }}>
            <Text
                size={"75px"}
                fw={900}
                variant="gradient"
                gradient={{
                    from: sc,
                    to: ec,
                    deg: 45,
                }}
                style={{
                    display: "flex",
                    justifyContent: "center",
                    userSelect: "none",
                    lineHeight: 1.1,
                }}
            >
                {title}
            </Text>
        </ColumnBox>
    );
};

export function PhotoGallery({
    medias,
    selecting,
    imageBaseScale,
    album,
    fetchAlbum,
    dispatch,
}: {
    medias: WeblensMedia[];
    selecting: boolean;
    imageBaseScale: number;
    album?: AlbumData;
    fetchAlbum?: () => void;
    dispatch;
}) {
    const listRef = useRef(null);
    const [boxNode, setBoxNode] = useState(null);
    const boxSize = useResize(boxNode);

    const rows = useMemo(() => {
        const MARGIN_SIZE = 4;

        if (medias.length === 0 || !boxSize.width) {
            return [];
        }

        const innerMedias = [...medias];

        const rows: {
            rowScale: number;
            items: WeblensMedia[];
            element?: JSX.Element;
        }[] = [];
        let currentRowWidth = 0;
        let currentRow = [];

        while (true) {
            if (innerMedias.length === 0) {
                if (currentRow.length !== 0) {
                    rows.push({ rowScale: imageBaseScale, items: currentRow });
                }
                break;
            }
            const m: WeblensMedia = innerMedias.pop();
            // Calculate width given height "imageBaseScale", keeping aspect ratio
            const newWidth =
                Math.floor((imageBaseScale / m.GetHeight()) * m.GetWidth()) +
                MARGIN_SIZE;

            // If we are out of media, and the image does not overflow this row, add it and break
            if (
                innerMedias.length === 0 &&
                !(currentRowWidth + newWidth > boxSize.width)
            ) {
                currentRow.push(m);
                rows.push({ rowScale: imageBaseScale, items: currentRow });
                break;
            }

            // If the image would overflow the window
            else if (currentRowWidth + newWidth > boxSize.width) {
                const leftover = boxSize.width - currentRowWidth;
                let consuming = false;
                if (newWidth / 2 < leftover || currentRow.length === 0) {
                    currentRow.push(m);
                    currentRowWidth += newWidth;
                    consuming = true;
                }
                const marginTotal = currentRow.length * MARGIN_SIZE;
                rows.push({
                    rowScale:
                        ((boxSize.width - marginTotal) /
                            (currentRowWidth - marginTotal)) *
                        imageBaseScale,
                    items: currentRow,
                });
                currentRow = [];
                currentRowWidth = 0;

                if (consuming) {
                    continue;
                }
            }
            currentRow.push(m);
            currentRowWidth += newWidth;
        }
        rows.unshift({ rowScale: 10, items: [] });
        if (album) {
            rows.unshift({
                rowScale: 75,
                items: [],
                element: (
                    <AlbumTitle
                        startColor={album.PrimaryColor}
                        endColor={album.SecondaryColor}
                        title={album.Name}
                    />
                ),
            });
        }
        rows.push({ rowScale: 40, items: [] });
        return rows;
    }, [medias, imageBaseScale, boxSize.width, album]);

    useEffect(() => {
        listRef.current?.resetAfterIndex(0);
    }, [boxSize.width, imageBaseScale, medias.length, selecting]);
    const itemSizeFunc = useCallback(
        (i: number) => rows[i].rowScale + 4,
        [rows]
    );

    return (
        <ColumnBox reff={setBoxNode}>
            <List
                className="no-scrollbars"
                ref={listRef}
                height={boxSize.height}
                overscanRowCount={150}
                width={boxSize.width}
                itemCount={rows.length}
                estimatedItemSize={imageBaseScale}
                itemSize={itemSizeFunc}
                itemData={{
                    rows: rows,
                    selecting: selecting,
                    dispatch: dispatch,
                    albumId: album?.Id,
                    fetchAlbum: fetchAlbum,
                }}
            >
                {Cell}
            </List>
        </ColumnBox>
    );
}
