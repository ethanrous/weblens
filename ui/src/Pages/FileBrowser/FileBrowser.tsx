// React
import {
    memo,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useReducer,
    useRef,
    useState,
} from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";

// Icons
import {
    IconArrowBackUp,
    IconArrowLeft,
    IconCalendar,
    IconChevronRight,
    IconDownload,
    IconFile,
    IconFileAnalytics,
    IconFolder,
    IconFolderPlus,
    IconHistory,
    IconHome,
    IconInfoCircle,
    IconLogin,
    IconPhotoPlus,
    IconScan,
    IconSearch,
    IconServer,
    IconShare,
    IconSpiral,
    IconTrash,
    IconUpload,
    IconUserMinus,
    IconUsers,
    IconX,
} from "@tabler/icons-react";

// Mantine
import {
    Box,
    Button,
    Divider,
    FileButton,
    Input,
    Menu,
    Space,
    Text,
} from "@mantine/core";
import { useDebouncedValue, useMouse } from "@mantine/hooks";
import { notifications } from "@mantine/notifications";

// Weblens
import { userContext } from "../../Context";
import {
    DeleteFiles,
    DeleteShare,
    GetFileInfo,
    GetFolderData,
    NewWormhole,
    SearchFolder,
    TrashFiles,
    UnTrashFiles,
    UpdateFileShare,
    getPastFolderInfo,
    moveFiles,
} from "../../api/FileBrowserApi";
import { dispatchSync } from "../../api/Websocket";
import Crumbs, { StyledBreadcrumb } from "../../components/Crumbs";
import HeaderBar from "../../components/HeaderBar";
import { GlobalContextType, ItemProps } from "../../components/ItemDisplay";
import { ItemScroller } from "../../components/ItemScroller";
import NotFound from "../../components/NotFound";
import { MediaImage } from "../../components/PhotoContainer";
import Presentation, {
    PresentationContainer,
} from "../../components/Presentation";
import UploadStatus, { useUploadStatus } from "../../components/UploadStatus";
import "./style/filebrowserStyle.css";
import "../../components/style.css";
import {
    AuthHeaderT,
    FileBrowserAction,
    FBDispatchT,
    FbStateT,
    MediaDataT,
    ScanMeta,
    FileInfoT,
    getBlankFile,
    UserContextT,
} from "../../types/Types";
import { humanFileSize } from "../../util";
import { AlbumScoller } from "./FileBrowserAlbums";
import {
    HandleDrop,
    HandleUploadButton,
    MoveSelected,
    SetFileData,
    uploadViaUrl,
    downloadSelected,
    fileBrowserReducer,
    getRealId,
    getSortFunc,
    useKeyDownFileBrowser,
    usePaste,
    useSubscribe,
} from "./FileBrowserLogic";
import { ShareBox } from "./FilebrowserShareMenu";
import { GetFilesContext, GetItemsList } from "./FilesContext";
import {
    ColumnBox,
    DirViewWrapper,
    FileInfoDisplay,
    GetStartedCard,
    IconDisplay,
    PresentationFile,
    RowBox,
    TaskProgCard,
    TransferCard,
    WebsocketStatus,
} from "./FilebrowserStyles";
import { SelectIcon, WeblensButton } from "../../components/WeblensButton";
import { FileRows } from "./FileRows";
import {
    useResize,
    useResizeDrag,
    useWindowSize,
} from "../../components/hooks";
import { WeblensProgress } from "../../components/WeblensProgress";
import { IconFiles } from "@tabler/icons-react";
import { FilesPane } from "./FileInfoPane";
import { StatTree } from "./FileStatTree";
import FileHistoryMenu, { SnapshotMenu } from "./FileHistory";
import { FileSortBox } from "./FileSortBox";

function PasteImageDialogue({
    img,
    dirMap,
    folderId,
    authHeader,
    dispatch,
    wsSend,
}: {
    img: ArrayBuffer;
    dirMap: Map<string, FileInfoT>;
    folderId;
    authHeader;
    dispatch;
    wsSend;
}) {
    if (!img) {
        return null;
    }
    const media: MediaDataT = {} as MediaDataT;
    media.mediaId = "paste";
    media.thumbnail = img;

    return (
        <PresentationContainer
            shadeOpacity={"0.25"}
            onClick={() => dispatch({ type: "paste_image", img: null })}
        >
            <ColumnBox
                style={{
                    position: "absolute",
                    justifyContent: "center",
                    alignItems: "center",
                    zIndex: 2,
                }}
            >
                <Text fw={700} size="40px" style={{ paddingBottom: "50px" }}>
                    Upload from clipboard?
                </Text>
                <ColumnBox
                    onClick={(e) => {
                        e.stopPropagation();
                    }}
                    style={{
                        height: "50%",
                        width: "max-content",
                        backgroundColor: "#222277ee",
                        padding: "10px",
                        borderRadius: "8px",
                        overflow: "hidden",
                    }}
                >
                    <MediaImage media={media} quality="thumbnail" />
                </ColumnBox>
                <RowBox
                    style={{
                        justifyContent: "space-between",
                        width: "300px",
                        height: "150px",
                    }}
                >
                    <Button
                        size="xl"
                        variant="default"
                        onClick={(e) => {
                            e.stopPropagation();
                            dispatch({ type: "paste_image", img: null });
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        size="xl"
                        color="#4444ff"
                        onClick={(e) => {
                            e.stopPropagation();
                            uploadViaUrl(
                                img,
                                folderId,
                                dirMap,
                                authHeader,
                                dispatch,
                                wsSend
                            );
                        }}
                    >
                        Upload
                    </Button>
                </RowBox>
            </ColumnBox>
        </PresentationContainer>
    );
}

const SIDEBAR_BREAKPOINT = 650;

const GlobalActions = memo(
    ({
        fbState: fb,
        uploadState,
        dispatch,
        wsSend,
        uploadDispatch,
    }: {
        fbState: FbStateT;
        uploadState;
        dispatch: FBDispatchT;
        wsSend: (action: string, content: any) => void;
        uploadDispatch;
    }) => {
        const nav = useNavigate();
        const { usr, authHeader }: UserContextT = useContext(userContext);
        const windowSize = useWindowSize();
        const [trashSize, trashUnits] = humanFileSize(fb.trashDirSize);

        const [resizing, setResizing] = useState(false);
        const [resizeOffset, setResizeOffset] = useState(
            windowSize?.width > SIDEBAR_BREAKPOINT ? 300 : 75
        );
        useResizeDrag(resizing, setResizing, (s) => {
            setResizeOffset(Math.min(s > 200 ? s : 75, 600));
        });

        useEffect(() => {
            if (windowSize.width < SIDEBAR_BREAKPOINT && resizeOffset >= 300) {
                setResizeOffset(75);
            } else if (
                windowSize.width >= SIDEBAR_BREAKPOINT &&
                resizeOffset < 300
            ) {
                setResizeOffset(300);
            }
        }, [windowSize.width]);

        return (
            <RowBox
                style={{
                    width: resizeOffset,
                    height: "100%",
                    flexGrow: 0,
                    flexShrink: 0,
                    alignItems: "flex-start",
                }}
            >
                {usr.isLoggedIn === false && (
                    <Box className="login-required-background">
                        <WeblensButton
                            label="Login"
                            Left={<IconLogin className="button-icon" />}
                            centerContent
                            onClick={() => nav("/login")}
                            style={{ maxWidth: 300 }}
                        />
                    </Box>
                )}

                <Box
                    style={{
                        display: "flex",
                        height: "100%",
                        width: "100%",
                        flexDirection: "column",
                        alignItems: "center",
                        flexShrink: 0,
                        padding: 20,
                    }}
                >
                    <WeblensButton
                        label="Home"
                        toggleOn={
                            fb.folderInfo.id === usr?.homeId &&
                            fb.fbMode === "default"
                        }
                        width={"100%"}
                        allowRepeat={false}
                        Left={<IconHome className="button-icon" />}
                        onMouseUp={(e) => {
                            e.stopPropagation();
                            if (fb.draggingState !== 0) {
                                moveFiles(
                                    Array.from(fb.selected.keys()),
                                    usr.homeId,
                                    authHeader
                                );
                                dispatch({
                                    type: "set_dragging",
                                    dragging: false,
                                });
                            } else {
                                nav("/files/home");
                            }
                        }}
                    />

                    <WeblensButton
                        label="Shared"
                        toggleOn={fb.fbMode === "share"}
                        disabled={fb.draggingState !== 0}
                        allowRepeat={false}
                        Left={<IconUsers className="button-icon" />}
                        width={"100%"}
                        onClick={() => {
                            nav("/files/shared");
                        }}
                    />

                    <WeblensButton
                        label="Trash"
                        toggleOn={fb.folderInfo.id === usr?.trashId}
                        allowRepeat={false}
                        Left={<IconTrash className="button-icon" />}
                        width={"100%"}
                        postScript={
                            trashSize && resizeOffset >= 150
                                ? `${trashSize}${trashUnits}`
                                : ""
                        }
                        onMouseUp={(e) => {
                            e.stopPropagation();
                            if (fb.draggingState !== 0) {
                                moveFiles(
                                    Array.from(fb.selected.keys()),
                                    usr.trashId,
                                    authHeader
                                );
                                dispatch({
                                    type: "set_dragging",
                                    dragging: false,
                                });
                            } else {
                                nav("/files/trash");
                            }
                        }}
                    />

                    <Space h={"md"} />

                    {usr.admin && (
                        <ColumnBox style={{ height: "max-content" }}>
                            <WeblensButton
                                label="External"
                                toggleOn={fb.fbMode === "external"}
                                allowRepeat={false}
                                Left={<IconServer className="button-icon" />}
                                width={"100%"}
                                disabled={fb.draggingState !== 0}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    nav("/files/external");
                                }}
                            />
                            <Space h={"md"} />
                        </ColumnBox>
                    )}

                    <WeblensButton
                        label="New Folder"
                        Left={<IconFolderPlus className="button-icon" />}
                        showSuccess={false}
                        disabled={
                            fb.draggingState !== 0 || !fb.folderInfo.modifiable
                        }
                        onClick={(e) => {
                            e.stopPropagation();
                            dispatch({ type: "new_dir" });
                        }}
                        width={"100%"}
                    />

                    <FileButton
                        onChange={(files) => {
                            HandleUploadButton(
                                files,
                                fb.folderInfo.id,
                                false,
                                "",
                                authHeader,
                                uploadDispatch,
                                wsSend
                            );
                        }}
                        accept="file"
                        multiple
                    >
                        {(props) => {
                            return (
                                <WeblensButton
                                    label="Upload"
                                    showSuccess={false}
                                    disabled={
                                        fb.draggingState !== 0 ||
                                        !fb.folderInfo.modifiable
                                    }
                                    Left={
                                        <IconUpload className="button-icon" />
                                    }
                                    width={"100%"}
                                    // height={"48px"}
                                    onClick={() => props.onClick()}
                                />
                            );
                        }}
                    </FileButton>

                    <Divider w={"100%"} my="lg" size={1.5} />

                    <UsageInfo
                        inHome={fb.folderInfo.id === usr.homeId}
                        homeDirSize={fb.homeDirSize}
                        currentFolderSize={fb.folderInfo.size}
                        displayCurrent={fb.folderInfo.id !== "shared"}
                        trashSize={fb.trashDirSize}
                        selected={Array.from(fb.selected.keys()).map((v) =>
                            fb.dirMap.get(v)
                        )}
                        dirId={fb.contentId}
                        mode={fb.fbMode}
                    />
                    <TaskProgress
                        scanProgress={fb.scanProgress}
                        dispatch={dispatch}
                    />
                    <Box
                        style={{
                            display: "flex",
                            flexDirection: "column",
                            height: "40px",
                        }}
                    />
                    <UploadStatus
                        uploadState={uploadState}
                        uploadDispatch={uploadDispatch}
                    />
                </Box>
                <Box
                    draggable={false}
                    className="resize-bar-wrapper"
                    onMouseDown={(e) => {
                        e.preventDefault();
                        setResizing(true);
                    }}
                >
                    <Box className="resize-bar" />
                </Box>
            </RowBox>
        );
    },
    (p, n) => {
        if (p.fbState.folderInfo !== n.fbState.folderInfo) {
            return false;
        } else if (p.fbState.scanProgress !== n.fbState.scanProgress) {
            return false;
        } else if (p.uploadState !== n.uploadState) {
            return false;
        } else if (p.fbState.selected !== n.fbState.selected) {
            return false;
        } else if (p.fbState.trashDirSize !== n.fbState.trashDirSize) {
            return false;
        } else if (p.fbState.draggingState !== n.fbState.draggingState) {
            return false;
        }
        return true;
    }
);

const UsageInfo = memo(
    ({
        inHome,
        homeDirSize,
        currentFolderSize,
        displayCurrent,
        trashSize,
        selected,
        mode,
        dirId,
    }: {
        inHome: boolean;
        homeDirSize: number;
        currentFolderSize: number;
        displayCurrent: boolean;
        trashSize: number;
        selected: FileInfoT[];
        mode: string;
        dirId: string;
    }) => {
        const [box, setBox] = useState(null);
        const size = useResize(box);
        const nav = useNavigate();

        if (!displayCurrent) {
            return null;
        }
        if (homeDirSize === 0) {
            homeDirSize = currentFolderSize;
        }
        if (!trashSize) {
            trashSize = 0;
        }
        if (inHome) {
            currentFolderSize = currentFolderSize - trashSize;
        }

        const selectedSize = selected.reduce((acc: number, x: FileInfoT) => {
            return acc + x.size;
        }, 0);

        if (homeDirSize < currentFolderSize) {
            currentFolderSize = homeDirSize;
        }
        let usagePercent =
            selected.length === 0
                ? (currentFolderSize / homeDirSize) * 100
                : (selectedSize / currentFolderSize) * 100;
        if (!usagePercent) {
            usagePercent = 0;
        }

        const miniMode = size.width < 100;

        let startIcon =
            selected.length === 0 ? (
                <IconFolder size={20} />
            ) : (
                <IconFiles size={20} />
            );
        let endIcon =
            selected.length === 0 ? (
                <IconHome size={20} />
            ) : (
                <IconFolder size={20} />
            );
        if (miniMode) {
            [startIcon, endIcon] = [endIcon, startIcon];
        }

        return (
            <Box
                ref={setBox}
                style={{
                    display: "flex",
                    flexDirection: "column",
                    height: "max-content",
                    width: "100%",
                    alignItems: miniMode ? "center" : "flex-start",
                    gap: 10,
                }}
            >
                <RowBox
                    style={{
                        height: "max-content",
                        display: miniMode ? "none" : "flex",
                        gap: 8,
                    }}
                >
                    <Text fw={600} style={{ userSelect: "none" }}>
                        Usage
                    </Text>
                    <Text fw={650} style={{ userSelect: "none" }}>
                        {usagePercent ? usagePercent.toFixed(2) : 0}%
                    </Text>
                    <Box style={{ flexGrow: 1 }} />
                    <IconFileAnalytics
                        style={{ cursor: "pointer" }}
                        onClick={() =>
                            nav(
                                `/files/stats/${
                                    mode === "external" ? mode : dirId
                                }`
                            )
                        }
                    />
                </RowBox>
                {miniMode && startIcon}
                <WeblensProgress
                    key={miniMode ? "usage-vertical" : "usage-horizontal"}
                    color="#4444ff"
                    value={usagePercent}
                    orientation={miniMode ? "vertical" : "horizontal"}
                    style={{
                        height: miniMode ? 100 : 20,
                        width: miniMode ? 20 : "100%",
                    }}
                />
                <RowBox
                    style={{
                        height: "max-content",
                        width: miniMode ? "max-content" : "98%",
                    }}
                >
                    {displayCurrent && !miniMode && (
                        <RowBox>
                            {startIcon}
                            <Text
                                style={{
                                    userSelect: "none",
                                    display: miniMode ? "none" : "block",
                                }}
                                size="14px"
                                pl={3}
                            >
                                {selected.length === 0
                                    ? humanFileSize(currentFolderSize)
                                    : humanFileSize(selectedSize)}
                            </Text>
                        </RowBox>
                    )}
                    <RowBox
                        style={{
                            justifyContent: "right",
                            width: "max-content",
                        }}
                    >
                        <Text
                            style={{
                                userSelect: "none",
                                display: miniMode ? "none" : "block",
                            }}
                            size="14px"
                            pr={3}
                        >
                            {selected.length === 0
                                ? humanFileSize(homeDirSize)
                                : humanFileSize(currentFolderSize)}
                        </Text>
                        {endIcon}
                    </RowBox>
                </RowBox>
            </Box>
        );
    },
    (p, n) => {
        if (p.inHome !== n.inHome) {
            return false;
        }
        if (p.homeDirSize !== n.homeDirSize) {
            return false;
        }
        if (p.currentFolderSize !== n.currentFolderSize) {
            return false;
        }
        if (p.displayCurrent !== n.displayCurrent) {
            return false;
        }
        if (p.trashSize !== n.trashSize) {
            return false;
        }
        if (p.selected.length !== n.selected.length) {
            return false;
        }
        return true;
    }
);

const TaskProgress = ({
    scanProgress,
    dispatch,
}: {
    scanProgress: ScanMeta[];
    dispatch: FBDispatchT;
}) => {
    if (scanProgress.length == 0) {
        return null;
    }
    const cards = useMemo(() => {
        return scanProgress.map((sp) => (
            <TaskProgCard key={sp.taskId} prog={sp} dispatch={dispatch} />
        ));
    }, [scanProgress]);
    return <ColumnBox style={{ height: "max-content" }}>{cards}</ColumnBox>;
};

function FileSearch({
    fb,
    defaultOpen = false,
    searchRef,
    dispatch,
}: {
    fb: FbStateT;
    defaultOpen?: boolean;
    searchRef;
    dispatch;
}) {
    const [searchOpen, setSearchOpen] = useState(defaultOpen);
    const [hintOpen, setHintOpen] = useState(false);
    const [error, setError] = useState(false);
    const nav = useNavigate();

    useEffect(() => {
        if (Boolean(fb.searchContent) && !searchOpen) {
            setSearchOpen(true);
        }
    }, [searchOpen, fb.searchContent]);

    useEffect(() => {
        if (fb.fbMode !== "search") {
            setHintOpen(false);
            setSearchOpen(false);
        }
    }, [fb.fbMode]);

    useEffect(() => {
        if (
            !Boolean(fb.searchContent) ||
            document.activeElement !== searchRef.current
        ) {
            setHintOpen(false);
            return;
        }
        try {
            new RegExp(fb.searchContent);
            setError(false);
            setHintOpen(true);
        } catch {
            setHintOpen(false);
        }
    }, [setHintOpen, fb.searchContent]);

    return (
        <ColumnBox
            style={{
                height: "max-content",
                width: "max-content",
                alignItems: "flex-start",
                marginRight: 5,
            }}
        >
            <Box className="search-box">
                <IconSearch
                    color="white"
                    className="search-icon"
                    onClick={() => {
                        setSearchOpen(true);
                        searchRef.current.focus();
                    }}
                />
                <Input
                    mod={{ "data-open": "false" }}
                    onBlur={() => {
                        if (fb.searchContent === "") {
                            setSearchOpen(false);
                            setHintOpen(false);
                            searchRef.current.blur();
                        } else if (hintOpen) {
                            setHintOpen(false);
                        }
                    }}
                    onFocus={() => {
                        if (fb.searchContent === "") {
                            return;
                        }
                        try {
                            new RegExp(fb.searchContent);
                            setError(false);
                            setHintOpen(true);
                        } catch {
                            setHintOpen(false);
                        }
                    }}
                    classNames={{
                        input: `search-input search-input-${
                            searchOpen ? "open" : "closed"
                        }`,
                    }}
                    unstyled
                    value={fb.searchContent}
                    ref={searchRef}
                    onChange={(e) =>
                        dispatch({
                            type: "set_search",
                            search: e.target.value,
                        })
                    }
                    onKeyDown={(e) => {
                        if (e.key === "Enter" && !hintOpen) {
                            console.log("HE?");
                            e.stopPropagation();
                            if (!Boolean(fb.searchContent)) {
                                nav(`/files/${fb.contentId}`);
                                return;
                            }
                            setError(true);
                            setTimeout(() => setError(false), 2000);
                        }
                    }}
                />
            </Box>
            {hintOpen && (
                <Box className="search-hint-box">
                    <Box className="key-line">
                        <Text>Press</Text>
                        <Text className="key-display">Enter</Text>
                        <Text>to search all files</Text>
                    </Box>
                </Box>
            )}
            {error && (
                <Box className="search-hint-box" mod={{ "data-error": "true" }}>
                    <Text>Not valid regex</Text>
                </Box>
            )}
        </ColumnBox>
    );
}

function DraggingCounter({ dragging, dirMap, selected, dispatch }) {
    const position = useMouse();
    const selectedKeys = Array.from(selected.keys());
    const { files, folders } = useMemo(() => {
        let files = 0;
        let folders = 0;

        selectedKeys.forEach((k: string) => {
            if (dirMap.get(k)?.isDir) {
                folders++;
            } else {
                files++;
            }
        });
        return { files, folders };
    }, [JSON.stringify(selectedKeys)]);

    if (dragging !== 1) {
        return null;
    }

    return (
        <ColumnBox
            style={{
                position: "fixed",
                top: position.y + 8,
                left: position.x + 8,
                zIndex: 10,
            }}
            onMouseUp={() => {
                dispatch({ type: "set_dragging", dragging: false });
            }}
        >
            {Boolean(files) && (
                <RowBox style={{ height: "max-content" }}>
                    <IconFile size={30} />
                    <Space w={10} />
                    <StyledBreadcrumb
                        label={files.toString()}
                        fontSize={30}
                        compact
                    />
                </RowBox>
            )}
            {Boolean(folders) && (
                <RowBox style={{ height: "max-content" }}>
                    <IconFolder size={30} />
                    <Space w={10} />
                    <StyledBreadcrumb
                        label={folders.toString()}
                        fontSize={30}
                        compact
                    />
                </RowBox>
            )}
        </ColumnBox>
    );
}

function FileContextMenu({
    itemId,
    fbState,
    open,
    setOpen,
    menuPos,
    dispatch,
    wsSend,
    authHeader,
}: {
    itemId: string;
    fbState: FbStateT;
    open;
    setOpen;
    menuPos;
    dispatch;
    wsSend;
    authHeader;
}) {
    const { usr }: UserContextT = useContext(userContext);
    const [shareMenu, setShareMenu] = useState(false);
    const [historyMenu, setHistoryMenu] = useState(false);
    const [addToAlbumMenu, setAddToAlbumMenu] = useState(false);
    const itemInfo: FileInfoT = fbState.dirMap.get(itemId) || ({} as FileInfoT);
    const selected: boolean = Boolean(fbState.selected.get(itemId));

    useEffect(() => {
        dispatch({ type: "set_block_focus", block: open });
    }, [dispatch, open]);

    const { items } = useMemo(() => {
        if (fbState.dirMap.size === 0) {
            return { items: [], anyDisplayable: false };
        }
        const itemIds = selected
            ? Array.from(fbState.selected.keys())
            : [itemId];
        let mediaCount = 0;
        const items = itemIds.map((i) => {
            const item = fbState.dirMap.get(i);
            if (!item) {
                return null;
            }
            if (item.displayable || item.isDir) {
                mediaCount++;
            }
            return item;
        });

        return { items: items.filter((i) => Boolean(i)), mediaCount };
    }, [
        itemId,
        JSON.stringify(fbState.dirMap.get(itemId)),
        selected,
        fbState.selected,
    ]);

    let extraString;
    if (selected && fbState.selected.size > 1) {
        extraString = ` +${fbState.selected.size - 1} more`;
    }

    const wormholeId = useMemo(() => {
        if (itemInfo.shares) {
            const whs = itemInfo.shares.filter((s) => s.Wormhole);
            if (whs.length !== 0) {
                return whs[0].shareId;
            }
        }
    }, [itemInfo.shares]);
    const selectedMedia = useMemo(
        () => items.filter((i) => i.displayable).map((i) => i.id),
        [items]
    );
    const selectedFolders = useMemo(
        () => items.filter((i) => i.isDir).map((i) => i.id),
        [items]
    );
    const inTrash = fbState.folderInfo.id === usr.trashId;
    const inShare = fbState.folderInfo.id === "shared";
    let trashName;
    if (inTrash) {
        trashName = "Delete Forever";
    } else if (inShare) {
        trashName = "Unshare Me";
    } else {
        trashName = "Delete";
    }

    return (
        <Menu
            opened={open}
            closeDelay={0}
            openDelay={0}
            onClose={() => setOpen(false)}
            closeOnClickOutside={!(addToAlbumMenu || shareMenu)}
            position="right-start"
            closeOnItemClick={false}
            transitionProps={{ duration: 100, exitDuration: 0 }}
            styles={{
                dropdown: {
                    boxShadow: "0px 0px 20px -5px black",
                    width: "max-content",
                    padding: 10,
                    border: 0,
                },
            }}
        >
            <Menu.Target>
                <Box
                    style={{
                        position: "absolute",
                        top: menuPos.y,
                        left: menuPos.x,
                    }}
                />
            </Menu.Target>

            <Menu.Dropdown
                onClick={(e) => e.stopPropagation()}
                onDoubleClick={(e) => e.stopPropagation()}
            >
                <Menu.Label>
                    <RowBox style={{ gap: 8, justifyContent: "center" }}>
                        {itemInfo.isDir && <IconFolder />}
                        {!itemInfo.isDir && <IconFile />}
                        <Text truncate="end" style={{ maxWidth: "250px" }}>
                            {itemInfo.filename}
                        </Text>
                        {extraString}
                    </RowBox>
                </Menu.Label>

                <Divider my={10} />

                <Menu
                    opened={addToAlbumMenu}
                    trigger="hover"
                    disabled={inTrash}
                    offset={0}
                    position="right-start"
                    onOpen={() => setAddToAlbumMenu(true)}
                    onClose={() => setAddToAlbumMenu(false)}
                    styles={{
                        dropdown: {
                            boxShadow: "0px 0px 20px -5px black",
                            width: "max-content",
                            padding: 10,
                            border: 0,
                        },
                    }}
                >
                    <Menu.Target>
                        {/* <Box className={`menu-item${inTrash ? '-disabled' : ''}`}> */}
                        <Box
                            className="menu-item"
                            mod={{ "data-disabled": inTrash.toString() }}
                        >
                            <IconPhotoPlus />
                            <Text className="menu-item-text">Add to Album</Text>
                            <IconChevronRight />
                        </Box>
                    </Menu.Target>
                    <Menu.Dropdown onMouseOver={(e) => e.stopPropagation()}>
                        <AlbumScoller
                            selectedMedia={selectedMedia}
                            selectedFolders={selectedFolders}
                            authHeader={authHeader}
                        />
                    </Menu.Dropdown>
                </Menu>

                {/* Wormhole menu */}
                {itemInfo.isDir && (
                    <Box
                        className="menu-item"
                        mod={{ "data-disabled": inTrash.toString() }}
                        style={{
                            pointerEvents:
                                fbState.selected.size > 1 && selected
                                    ? "none"
                                    : "all",
                        }}
                        onClick={(e) => {
                            e.stopPropagation();
                            if (!wormholeId) {
                                NewWormhole(itemId, authHeader);
                            } else {
                                navigator.clipboard.writeText(
                                    `${window.location.origin}/wormhole/${wormholeId}`
                                );
                                setOpen(false);
                                notifications.show({
                                    message: "Link to wormhole copied",
                                    color: "green",
                                });
                            }
                        }}
                    >
                        <IconSpiral
                            color={fbState.selected.size > 1 ? "grey" : "white"}
                        />
                        <Text
                            className="menu-item-text"
                            truncate="end"
                            c={fbState.selected.size > 1 ? "grey" : "white"}
                        >
                            {!wormholeId ? "Attach" : "Copy"} Wormhole
                        </Text>
                    </Box>
                )}

                {/* Share menu */}
                <Menu
                    opened={shareMenu}
                    disabled={inTrash}
                    trigger="hover"
                    closeOnClickOutside={false}
                    offset={0}
                    position="right-start"
                    onOpen={() => setShareMenu(true)}
                    onClose={() => setShareMenu(false)}
                    styles={{
                        dropdown: {
                            boxShadow: "0px 0px 20px -5px black",
                            width: "max-content",
                            padding: 0,
                            border: 0,
                        },
                    }}
                >
                    <Menu.Target>
                        <Box
                            className="menu-item"
                            mod={{ "data-disabled": inTrash.toString() }}
                        >
                            <IconShare />
                            <Text className="menu-item-text">Share</Text>
                            <IconChevronRight />
                        </Box>
                    </Menu.Target>
                    <Menu.Dropdown>
                        <ShareBox candidates={items} authHeader={authHeader} />
                    </Menu.Dropdown>
                </Menu>

                <Box
                    className="menu-item"
                    onClick={(e) => {
                        e.stopPropagation();
                        downloadSelected(
                            selected
                                ? Array.from(fbState.selected.keys()).map(
                                      (fId) => fbState.dirMap.get(fId)
                                  )
                                : [fbState.dirMap.get(itemId)],
                            dispatch,
                            wsSend,
                            authHeader,
                            itemInfo.shares[0].shareId
                        );
                    }}
                >
                    <IconDownload />
                    <Text className="menu-item-text">Download</Text>
                </Box>
                {historyMenu && <FileHistoryMenu fileId={itemInfo.id} />}
                <Box
                    className="menu-item"
                    onClick={(e) => {
                        e.stopPropagation();

                        setHistoryMenu(true);
                    }}
                >
                    <IconHistory />
                    <Text className="menu-item-text">File History</Text>
                </Box>

                {itemInfo.isDir && (
                    <Box
                        className="menu-item"
                        onClick={() =>
                            dispatchSync(
                                items.map((i: FileInfoT) => i.id),
                                wsSend,
                                true,
                                true
                            )
                        }
                    >
                        <IconScan />
                        <Text className="menu-item-text">Scan</Text>
                    </Box>
                )}

                <Divider w={"100%"} my="sm" />

                {wormholeId && (
                    <Box
                        className="menu-item"
                        mod={{ "data-disabled": inTrash.toString() }}
                        onClick={(e) => {
                            e.stopPropagation();
                            DeleteShare(wormholeId, authHeader);
                        }}
                    >
                        <IconSpiral color="#ff8888" />
                        <Text
                            className="menu-item-text"
                            truncate="end"
                            c="#ff8888"
                        >
                            Remove Wormhole
                        </Text>
                    </Box>
                )}
                {inTrash && (
                    <Box
                        className="menu-item"
                        onClick={(e) => {
                            e.stopPropagation();
                            UnTrashFiles(
                                items.map((i) => i.id),
                                authHeader
                            );
                            setOpen(false);
                        }}
                    >
                        <IconArrowBackUp />
                        <Text className="menu-item-text">{"Put back"}</Text>
                    </Box>
                )}
                <Box
                    className="menu-item"
                    onClick={(e) => {
                        e.stopPropagation();

                        if (inTrash) {
                            DeleteFiles(
                                items.map((i) => i.id),
                                authHeader
                            );
                        } else if (inShare) {
                            let thisShare =
                                fbState.dirMap.get(itemId).shares[0];
                            UpdateFileShare(
                                thisShare.shareId,
                                thisShare.Public,
                                thisShare.Accessors.filter(
                                    (u) => u !== usr.username
                                ),
                                authHeader
                            );
                        } else {
                            TrashFiles(
                                items.map((i: FileInfoT) => i.id),
                                authHeader
                            );
                        }
                        setOpen(false);
                    }}
                >
                    {inShare ? (
                        <IconUserMinus color="#ff4444" />
                    ) : (
                        <IconTrash color="#ff4444" />
                    )}
                    <Text className="menu-item-text" c="#ff4444">
                        {trashName}
                    </Text>
                </Box>
            </Menu.Dropdown>
        </Menu>
    );
}

function SingleFile({
    file,
    doDownload,
}: {
    file: FileInfoT;
    doDownload: (file: FileInfoT) => void;
}) {
    if (!file.id) {
        return (
            <NotFound
                resourceType="Share"
                link="/files/home"
                setNotFound={() => {}}
            />
        );
    }

    return (
        <Box
            style={{
                width: "100%",
                height: "100%",
                display: "flex",
                flexDirection: "row",
                justifyContent: "space-around",
                paddingBottom: 8,
            }}
        >
            <Box
                className="icon-display-wrapper"
                style={{
                    display: "flex",
                    width: 150,
                    maxWidth: "65%",
                    flexGrow: 1,
                    alignItems: "center",
                }}
            >
                <IconDisplay file={file} allowMedia size={"65%"} />
            </Box>
            <ColumnBox style={{ flexGrow: 1, maxWidth: "50%", padding: 10 }}>
                <FileInfoDisplay file={file} />
                <Box style={{ minHeight: "40%" }}>
                    <RowBox
                        onClick={() => doDownload(file)}
                        style={{
                            backgroundColor: "#4444ff",
                            borderRadius: 4,
                            padding: 10,
                            height: "max-content",
                            cursor: "pointer",
                        }}
                    >
                        <IconDownload />
                        <Text c="white" style={{ paddingLeft: 10 }}>
                            Download {file.filename}
                        </Text>
                    </RowBox>
                </Box>
            </ColumnBox>
        </Box>
    );
}

function Files({
    fb: fb,
    notFound,
    setNotFound,
    searchRef,
    dispatch,
    wsSend,
    uploadDispatch,
    authHeader,
}: {
    fb: FbStateT;
    notFound;
    setNotFound;
    searchRef;
    dispatch: (action: FileBrowserAction) => void;
    wsSend: (action: string, content: any) => void;
    uploadDispatch;
    authHeader;
}) {
    const { usr, serverInfo }: UserContextT = useContext(userContext);
    const nav = useNavigate();
    const [debouncedSearch] = useDebouncedValue(fb.searchContent, 200);

    const [fullViewRef, setFullViewRef] = useState(null);
    useResize(fullViewRef);
    const [contentViewRef, setContentViewRef] = useState(null);

    const moveSelectedTo = useCallback(
        (folderId: string) => {
            MoveSelected(fb.selected, folderId, authHeader);
            dispatch({ type: "clear_selected" });
        },
        [fb.selected.size, fb.contentId, authHeader]
    );

    const itemsList: ItemProps[] = useMemo(() => {
        return GetItemsList(fb, usr, debouncedSearch);
    }, [
        fb.dirMap,
        fb.selected,
        debouncedSearch,
        usr,
        fb.lastSelected,
        fb.sortFunc,
        fb.sortDirection,
    ]);

    useEffect(() => {
        dispatch({ type: "add_loading", loading: "fileSearch" });
    }, [debouncedSearch]);

    useEffect(() => {
        const fileIds = itemsList.map((v) => v.itemId);
        dispatch({
            type: "set_files_list",
            fileIds: fileIds,
        });
        dispatch({ type: "remove_loading", loading: "fileSearch" });
    }, [itemsList, dispatch]);

    const selectedInfo = useMemo(() => {
        return Array.from(fb.selected.keys()).map((fId) => fb.dirMap.get(fId));
    }, [fb.selected]);

    const itemsCtx: GlobalContextType = useMemo(() => {
        return GetFilesContext(fb, itemsList, nav, authHeader, dispatch);
    }, [
        itemsList,
        fb.contentId,
        fb.dirMap,
        fb.selected,
        fb.fbMode,
        fb.draggingState,
        fb.hovering,
        fb.holdingShift,
        dispatch,
    ]);

    if (notFound) {
        return (
            <NotFound
                resourceType="Folder"
                link="/files/home"
                setNotFound={setNotFound}
            />
        );
    }

    return (
        <Box
            ref={setFullViewRef}
            style={{
                display: "flex",
                flexDirection: "column",
                width: "100%",
                height: "100%",
                paddingLeft: 10,
            }}
        >
            <TransferCard
                action="Move"
                destination={fb.moveDest}
                boundRef={fullViewRef}
            />
            <RowBox
                style={{
                    height: "max-content",
                    justifyContent: "space-between",
                    padding: 8,
                }}
            >
                <Crumbs
                    finalFile={fb.folderInfo}
                    parents={fb.parents}
                    postText={
                        fb.viewingPast
                            ? `@ ${fb.viewingPast.toDateString()} ${fb.viewingPast.toLocaleTimeString()}`
                            : ""
                    }
                    navOnLast={false}
                    dragging={fb.draggingState}
                    moveSelectedTo={moveSelectedTo}
                    setMoveDest={(itemName) =>
                        dispatch({ type: "set_move_dest", fileName: itemName })
                    }
                />
                {serverInfo.role === "backup" && (
                    <SnapshotMenu dispatch={dispatch} />
                )}
                <FileSearch fb={fb} searchRef={searchRef} dispatch={dispatch} />
                <FileSortBox fb={fb} dispatch={dispatch} />
                <WeblensButton
                    Left={
                        fb.fileInfoMenu ? (
                            <IconX className="button-icon" />
                        ) : (
                            <IconInfoCircle className="button-icon" />
                        )
                    }
                    height={42}
                    width={42}
                    subtle
                    onClick={() =>
                        dispatch({
                            type: "set_file_info_menu",
                            open: !fb.fileInfoMenu,
                        })
                    }
                />
            </RowBox>
            <Box
                style={{
                    display: "flex",
                    flexDirection: "row",
                    height: "200px",
                    flexGrow: 1,
                    maxWidth: "100%",
                }}
            >
                <Box
                    ref={setContentViewRef}
                    style={{ flexGrow: 1, flexShrink: 1, width: 0 }}
                >
                    {(itemsList.length !== 0 && (
                        <ItemScroller
                            itemsContext={itemsList}
                            globalContext={itemsCtx}
                            parentNode={contentViewRef}
                            dispatch={dispatch}
                        />
                    )) ||
                        (fb.loading.length === 0 &&
                            fb.searchContent === "" &&
                            fb.searchContent === debouncedSearch && (
                                <GetStartedCard
                                    fb={fb}
                                    moveSelectedTo={moveSelectedTo}
                                    dispatch={dispatch}
                                    uploadDispatch={uploadDispatch}
                                    authHeader={authHeader}
                                    wsSend={wsSend}
                                />
                            )) ||
                        (fb.loading.length === 0 &&
                            fb.folderInfo.id === "shared" && (
                                <NotFound
                                    resourceType="any files shared with you"
                                    link="/files/home"
                                    setNotFound={setNotFound}
                                />
                            )) ||
                        (fb.loading.length === 0 && fb.searchContent !== "" && (
                            <ColumnBox
                                style={{
                                    justifyContent: "flex-end",
                                    height: "20%",
                                }}
                            >
                                <RowBox
                                    style={{
                                        height: "max-content",
                                        width: "max-content",
                                    }}
                                >
                                    <Text size="20px">
                                        No items match your search in
                                    </Text>
                                    <IconFolder style={{ marginLeft: 4 }} />
                                    <Text size="20px">
                                        {fb.folderInfo.filename}
                                    </Text>
                                </RowBox>
                                <Space h={10} />

                                <Box className="key-line">
                                    <Text size="16px">Press</Text>
                                    <Text className="key-display">Enter</Text>
                                    <Text>to search all files</Text>
                                </Box>
                            </ColumnBox>
                        ))}
                </Box>
                <FilesPane
                    open={fb.fileInfoMenu}
                    setOpen={(o) =>
                        dispatch({ type: "set_file_info_menu", open: o })
                    }
                    selectedFiles={selectedInfo}
                    contentId={fb.contentId}
                />
            </Box>
        </Box>
    );
}

function SearchResults({
    fbState,
    searchQuery,
    filter,
    searchRef,
    dispatch,
}: {
    fbState: FbStateT;
    searchQuery: string;
    filter: string;
    searchRef;
    dispatch: FBDispatchT;
}) {
    const nav = useNavigate();
    let titleString: string = "Searching ";
    if (searchQuery) {
        titleString += `for ${searchQuery}`;
    } else {
        titleString += `all files`;
    }
    titleString += ` in ${fbState.folderInfo.filename}`;
    if (filter) {
        titleString += ` ending with .${filter}`;
    }

    return (
        <ColumnBox>
            <RowBox style={{ height: 58, flexShrink: 0 }}>
                <IconArrowLeft
                    style={{ width: 40, height: 32, cursor: "pointer" }}
                    onClick={() => nav(-1)}
                />
                <Text className="crumb-text">{titleString}</Text>

                <Text
                    className="crumb-text"
                    c="#aaaaaa"
                    style={{ fontSize: "14px", marginLeft: 10 }}
                >
                    {fbState.searchResults.length} results
                </Text>
                <Space flex={1} />
                <FileSearch
                    fb={fbState}
                    defaultOpen
                    searchRef={searchRef}
                    dispatch={dispatch}
                />
            </RowBox>
            <Space h={10} />
            <FileRows files={fbState.searchResults} />
        </ColumnBox>
    );
}

function DirView({
    fb: fb,
    notFound,
    setNotFound,
    searchRef,
    searchQuery,
    searchFilter,
    dispatch,
    wsSend,
    uploadDispatch,
    authHeader,
}: {
    fb: FbStateT;
    notFound: boolean;
    setNotFound;
    searchRef;
    searchQuery: string;
    searchFilter: string;
    dispatch: FBDispatchT;
    wsSend: (action: string, content: any) => void;
    uploadDispatch;
    authHeader: AuthHeaderT;
}) {
    const download = useCallback(
        (file: FileInfoT) =>
            downloadSelected([file], dispatch, wsSend, authHeader, fb.shareId),
        [authHeader, wsSend, dispatch, fb.fbMode, fb.contentId, fb.shareId]
    );

    if (fb.loading.includes("files")) {
        return null;
    }

    if (fb.fbMode === "default" && fb.folderInfo.id && !fb.folderInfo.isDir) {
        return <SingleFile file={fb.folderInfo} doDownload={download} />;
    } else if (fb.fbMode === "stats") {
        return <StatTree folderInfo={fb.folderInfo} authHeader={authHeader} />;
    } else if (fb.fbMode === "search") {
        return (
            <SearchResults
                fbState={fb}
                searchQuery={searchQuery}
                filter={searchFilter}
                searchRef={searchRef}
                dispatch={dispatch}
            />
        );
    } else {
        return (
            <Files
                fb={fb}
                notFound={notFound}
                setNotFound={setNotFound}
                searchRef={searchRef}
                dispatch={dispatch}
                wsSend={wsSend}
                uploadDispatch={uploadDispatch}
                authHeader={authHeader}
            />
        );
    }
}

function useQuery() {
    const { search } = useLocation();
    const q = new URLSearchParams(search);
    const getQuery = useCallback(
        (s) => {
            const r = q.get(s);
            if (!r) {
                return "";
            }
            return r;
        },
        [q]
    );
    return getQuery;
}

const FileBrowser = () => {
    const urlPath = useParams()["*"];
    const query = useQuery();
    const searchQuery = query("query");
    const searchFilter = query("filter");
    const nav = useNavigate();
    const { authHeader, usr }: UserContextT = useContext(userContext);

    const searchRef = useRef();

    const [notFound, setNotFound] = useState(false);
    const { uploadState, uploadDispatch } = useUploadStatus();

    const [fb, dispatch]: [FbStateT, (action: FileBrowserAction) => void] =
        useReducer(fileBrowserReducer, {
            uploadMap: new Map<string, boolean>(),
            selected: new Map<string, boolean>(),
            dirMap: new Map<string, FileInfoT>(),
            viewingPast: null,
            folderInfo: getBlankFile(),
            menuPos: { x: 0, y: 0 },
            waitingForNewName: "",
            holdingShift: false,
            fileInfoMenu: false,
            blockFocus: false,
            searchResults: [],
            searchContent: "",
            lastSelected: "",
            draggingState: 0,
            scanProgress: [],
            menuTargetId: "",
            presentingId: "",
            sortDirection: 1,
            sortFunc: "Name",
            menuOpen: false,
            trashDirSize: 0,
            pasteImg: null,
            homeDirSize: 0,
            contentId: "",
            filesList: [],
            hovering: "",
            scrollTo: "",
            moveDest: "",
            shareId: "",
            parents: [],
            loading: [],
            numCols: 0,
            fbMode: "",
        });

    if (fb.fbMode && fb.fbMode !== "share" && usr.isLoggedIn === false) {
        nav("/login");
    }

    useEffect(() => {
        if (usr.isLoggedIn === undefined) {
            return;
        }
        dispatch({ type: "add_loading", loading: "files" });

        let mode: string;
        let contentId: string;
        let shareId: string;
        const splitPath = urlPath.split("/").filter((s) => s.length !== 0);

        if (splitPath.length === 0) {
            return;
        }

        if (splitPath[0] === "shared") {
            mode = "share";
            shareId = splitPath[1];
            contentId = splitPath[2];
        } else if (splitPath[0] === "external") {
            mode = "external";
            contentId = splitPath[1];
        } else if (splitPath[0] === "stats") {
            mode = "stats";
            contentId = splitPath[1];
        } else if (splitPath[0] === "search") {
            mode = "search";
            contentId = splitPath[1];
        } else {
            mode = "default";
            contentId = splitPath[0];
        }

        getRealId(contentId, mode, usr, authHeader).then((realId) => {
            dispatch({
                type: "set_location_state",
                realId: realId,
                mode: mode,
                shareId: shareId,
            });
            dispatch({ type: "remove_loading", loading: "files" });
        });
    }, [urlPath, dispatch, authHeader, usr]);

    const { wsSend, readyState } = useSubscribe(
        fb.contentId,
        fb.folderInfo.id,
        fb.fbMode,
        usr,
        dispatch,
        authHeader
    );

    useKeyDownFileBrowser(
        fb,
        searchQuery,
        usr,
        dispatch,
        authHeader,
        wsSend,
        searchRef
    );

    // Hook to handle uploading images from the clipboard
    usePaste(fb.contentId, usr, searchRef, fb.blockFocus, dispatch);

    // Reset most of the state when we change folders
    useEffect(() => {
        const syncState = async () => {
            if (!urlPath || urlPath === usr?.homeId || urlPath === undefined) {
                nav("/files/home", { replace: true });
            }

            // If we're not ready, leave
            if (!fb.fbMode || usr.isLoggedIn === undefined) {
                return;
            }

            if (fb.fbMode === "search") {
                dispatch({
                    type: "set_search_results",
                    fileInfos: [],
                });

                const folderData = await GetFileInfo(
                    fb.contentId,
                    "",
                    authHeader
                );
                SetFileData({ self: folderData }, dispatch, usr);

                const searchResults = await SearchFolder(
                    fb.contentId,
                    searchQuery,
                    searchFilter,
                    authHeader
                );

                dispatch({ type: "set_search", search: searchQuery });
                dispatch({
                    type: "set_search_results",
                    fileInfos: searchResults,
                });
                dispatch({ type: "remove_loading", loading: "files" });
                return;
            }
            // Kinda just reset everything...
            setNotFound(false);

            dispatch({ type: "clear_files" });
            dispatch({ type: "set_search", search: "" });

            let fileData;
            if (fb.viewingPast !== null) {
                fileData = await getPastFolderInfo(
                    fb.contentId,
                    fb.viewingPast,
                    authHeader
                );
            } else {
                fileData = await GetFolderData(
                    fb.contentId,
                    fb.fbMode,
                    fb.shareId,
                    usr,
                    dispatch,
                    authHeader
                ).catch((r) => {
                    if (r === 400 || r === 404) {
                        setNotFound(true);
                    } else {
                        notifications.show({
                            title: "Could not get folder info",
                            message: String(r),
                            color: "red",
                            autoClose: 5000,
                        });
                    }
                });
            }

            if (fileData) {
                SetFileData(fileData, dispatch, usr);
            }

            const jumpItem = query("jumpTo");
            if (jumpItem) {
                dispatch({ type: "set_scroll_to", fileId: jumpItem });
                dispatch({
                    type: "set_selected",
                    fileId: jumpItem,
                    selected: true,
                });
            }

            dispatch({ type: "remove_loading", loading: "files" });
        };
        syncState();
    }, [
        usr.username,
        authHeader,
        fb.contentId,
        fb.fbMode,
        searchQuery,
        fb.viewingPast,
    ]);

    return (
        <ColumnBox style={{ height: "100vh" }}>
            <HeaderBar
                dispatch={dispatch}
                page={"files"}
                loading={fb.loading}
            />
            <DraggingCounter
                dragging={fb.draggingState}
                dirMap={fb.dirMap}
                selected={fb.selected}
                dispatch={dispatch}
            />
            <Presentation
                itemId={fb.presentingId}
                mediaData={fb.dirMap.get(fb.presentingId)?.mediaData}
                element={() =>
                    PresentationFile({
                        file: fb.dirMap.get(fb.presentingId),
                    })
                }
                dispatch={dispatch}
            />
            <PasteImageDialogue
                img={fb.pasteImg}
                folderId={fb.contentId}
                dirMap={fb.dirMap}
                authHeader={authHeader}
                dispatch={dispatch}
                wsSend={wsSend}
            />
            <FileContextMenu
                itemId={fb.menuTargetId}
                fbState={fb}
                open={fb.menuOpen}
                setOpen={(o) => dispatch({ type: "set_menu_open", open: o })}
                menuPos={fb.menuPos}
                dispatch={dispatch}
                wsSend={wsSend}
                authHeader={authHeader}
            />
            <WebsocketStatus ready={readyState} />
            <RowBox
                style={{
                    alignItems: "flex-start",
                    height: "90vh",
                    flexGrow: 1,
                }}
            >
                <GlobalActions
                    fbState={fb}
                    uploadState={uploadState}
                    dispatch={dispatch}
                    wsSend={wsSend}
                    uploadDispatch={uploadDispatch}
                />
                <DirViewWrapper
                    fbState={fb}
                    folderName={fb.folderInfo?.filename}
                    dragging={fb.draggingState}
                    onDrop={(e) =>
                        HandleDrop(
                            e.dataTransfer.items,
                            fb.contentId,
                            Array.from(fb.dirMap.values()).map(
                                (value: FileInfoT) => value.filename
                            ),
                            false,
                            "",
                            authHeader,
                            uploadDispatch,
                            wsSend
                        )
                    }
                    dispatch={dispatch}
                >
                    {/* <Space h={10} /> */}
                    <DirView
                        fb={fb}
                        notFound={notFound}
                        setNotFound={setNotFound}
                        searchRef={searchRef}
                        searchQuery={searchQuery}
                        searchFilter={searchFilter}
                        dispatch={dispatch}
                        wsSend={wsSend}
                        uploadDispatch={uploadDispatch}
                        authHeader={authHeader}
                    />
                </DirViewWrapper>
            </RowBox>
        </ColumnBox>
    );
};

export default FileBrowser;
