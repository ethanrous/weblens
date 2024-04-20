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
    MediaDataT,
    MediaWrapperProps,
    FileInfoT,
    UserContextT,
} from "../types/Types";
import { ColumnBox, RowBox } from "../Pages/FileBrowser/FileBrowserStyles";
import { GetFileInfo } from "../api/FileBrowserApi";
import { MediaImage } from "./PhotoContainer";
import { userContext } from "../Context";
import {
    IconHome,
    IconTrash,
    IconFolder,
    IconPhoto,
    IconPhotoScan,
    IconUser,
} from "@tabler/icons-react";
import { StyledLoaf } from "./Crumbs";
import "./galleryStyle.css";
import { useResize } from "./hooks";
import { GalleryAction } from "../Pages/Gallery/GalleryLogic";

const MultiFileMenu = ({
    filesInfo,
    loading,
    menuOpen,
    setMenuOpen,
}: {
    filesInfo: FileInfoT[];
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
        const parts: any[] = v.pathFromHome.split("/");
        if (parts[0] === "HOME") {
            parts[0] = <IconHome />;
        } else if (parts[0] === "TRASH") {
            parts[0] = <IconTrash />;
        } else if (parts[0] === "SHARE") {
            parts[0] = <IconUser />;
        } else {
            console.error("Unknown filepath base type");
            return;
        }

        return StyledLoaf({ crumbs: parts });
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
                                key={f.id}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    window.open(
                                        `/files/${f.parentFolderId}?jumpTo=${f.id}`,
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
        const fileInfo: FileInfoT = await GetFileInfo(
            fileIds[0],
            "",
            authHeader
        );
        window.open(
            `/files/${fileInfo.parentFolderId}?jumpTo=${fileInfo.id}`,
            "_blank"
        );
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

const TypeIcon = (mediaData: MediaDataT) => {
    let icon;

    if (mediaData.mediaType.IsRaw) {
        icon = IconPhotoScan;
    } else if (mediaData.mediaType.IsVideo) {
        // icon = Theaters
    } else {
        icon = IconPhoto;
        // name = "Image"
    }
    return [icon, mediaData.mediaType.FriendlyName];
};

type mediaTypeProps = {
    Icon: any;
    filesCount?: number;
    onClick?: React.MouseEventHandler<HTMLDivElement>;
    innerRef?: React.ForwardedRef<HTMLDivElement>;
};

const StyledIcon = ({
    Icon,
    filesCount,
    onClick,
    innerRef,
}: mediaTypeProps) => {
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
    hovering,
}: {
    mediaData: MediaDataT;
    hovering: boolean;
}) {
    const { authHeader }: UserContextT = useContext(userContext);
    const [icon, name] = TypeIcon(mediaData);
    const [menuOpen, setMenuOpen] = useState(false);
    const [filesInfo, setFilesInfo] = useState([]);
    const [loading, setLoading] = useState(false);

    if ((!hovering && !menuOpen) || !icon) {
        return null;
    }

    return (
        <Box className="media-meta-preview">
            <Tooltip label={name} refProp="innerRef">
                <StyledIcon Icon={icon} />
            </Tooltip>
            <RowBox style={{ height: 32 }}>
                <Tooltip
                    label={
                        mediaData.fileIds.length === 1
                            ? "Visit File"
                            : "Visit Files"
                    }
                    refProp="innerRef"
                >
                    <StyledIcon
                        Icon={IconFolder}
                        filesCount={mediaData.fileIds.length}
                        onClick={(e) =>
                            goToFolder(
                                e,
                                mediaData.fileIds,
                                filesInfo,
                                setLoading,
                                setMenuOpen,
                                setFilesInfo,
                                authHeader
                            )
                        }
                    />
                </Tooltip>
                <MultiFileMenu
                    filesInfo={filesInfo}
                    loading={loading}
                    menuOpen={menuOpen}
                    setMenuOpen={setMenuOpen}
                />
            </RowBox>
        </Box>
    );
}

const MediaWrapper = memo(
    ({
        mediaData,
        selected,
        selecting,
        scale,
        dispatch,
    }: MediaWrapperProps) => {
        const ref = useRef();
        const [hovering, setHovering] = useState(false);

        const width = useMemo(() => {
            mediaData.ImgRef = ref;
            return mediaData.mediaWidth * (scale / mediaData.mediaHeight);
        }, [scale, mediaData]);

        if (mediaData.selected === undefined) {
            mediaData.selected = false;
        }

        return (
            <Box
                className="preview-card-container"
                mod={{
                    "data-selecting": selecting.toString(),
                    "data-selected": mediaData.selected.toString(),
                }}
                ref={ref}
                onClick={() => {
                    if (selecting) {
                        dispatch({
                            type: "set_selected",
                            mediaId: mediaData.mediaId,
                            selected: !mediaData.selected,
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
                    dispatch({
                        type: "set_menu_pos",
                        pos: { x: e.clientX, y: e.clientY },
                    });
                    dispatch({
                        type: "set_menu_target",
                        targetId: mediaData.mediaId,
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
                        height: scale,
                        width: "100%",
                    }}
                />
                <MediaInfoDisplay
                    mediaData={mediaData}
                    hovering={hovering && !selecting}
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
        return prev.mediaData.mediaId === next.mediaData.mediaId;
    }
);
export const BucketCards = ({
    medias,
    selecting = false,
    scale,
    dispatch,
}: {
    medias: MediaDataT[];
    selecting?: boolean;
    scale: number;
    dispatch;
}) => {
    if (!medias) {
        medias = [];
    }

    const mediaCards = medias.map((mediaData: MediaDataT) => {
        return (
            <MediaWrapper
                key={mediaData.mediaId}
                mediaData={mediaData}
                selected={mediaData.selected}
                selecting={selecting}
                scale={scale}
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

const TitleWrapper = ({ bucketTitle }) => {
    if (bucketTitle === "") {
        return null;
    }
    return (
        <Text
            style={{ fontSize: 20, fontWeight: 600 }}
            c={"white"}
            mt={1}
            pl={0.5}
        >
            {bucketTitle}
        </Text>
    );
};

type GalleryRow = {
    rowScale: number;
    items: MediaDataT[];
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
        dispatch: (action: GalleryAction) => void;
    };
    index: number;
    style;
}) => {
    return (
        <Box style={{ ...style }}>
            {data.rows[index].items.length > 0 && (
                <BucketCards
                    key={data.rows[index].items[0].mediaId}
                    selecting={data.selecting}
                    medias={data.rows[index].items}
                    scale={data.rows[index].rowScale}
                    dispatch={data.dispatch}
                />
            )}
            {data.rows[index].element}
        </Box>
    );
};

export function PhotoGallery({
    medias,
    selecting,
    imageBaseScale,
    title,
    dispatch,
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
            items: MediaDataT[];
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
            const m: MediaDataT = innerMedias.pop();
            // Calculate width given height "imageBaseScale", keeping aspect ratio
            const newWidth =
                Math.floor((imageBaseScale / m.mediaHeight) * m.mediaWidth) +
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
        if (title) {
            rows.unshift({ rowScale: 75, items: [], element: title });
        }
        rows.push({ rowScale: 40, items: [] });
        return rows;
    }, [medias, imageBaseScale, boxSize.width, title]);

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
                }}
            >
                {Cell}
            </List>
        </ColumnBox>
    );
}
