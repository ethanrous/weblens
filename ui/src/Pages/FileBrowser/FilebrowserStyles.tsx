import {
    Box,
    Card,
    MantineStyleProp,
    Text,
    Tooltip,
    ActionIcon,
    Space,
    Menu,
    Divider,
    FileButton,
    Center,
    Skeleton,
} from "@mantine/core";
import { useContext, useMemo, useState } from "react";
import {
    handleDragOver,
    HandleDrop,
    HandleUploadButton,
} from "./FileBrowserLogic";
import {
    FBDispatchT,
    FbStateT,
    UserContextT,
    UserInfoT,
} from "../../types/Types";
import { WeblensFile } from "../../classes/File";

import {
    IconDatabase,
    IconFile,
    IconFileZip,
    IconFolder,
    IconFolderCancel,
    IconFolderPlus,
    IconHome,
    IconHome2,
    IconPhoto,
    IconRefresh,
    IconServer,
    IconServer2,
    IconSpiral,
    IconTrash,
    IconUpload,
    IconUser,
    IconUsers,
} from "@tabler/icons-react";
import { UserContext } from "../../Context";
import { friendlyFolderName, humanFileSize, nsToHumanTime } from "../../util";
import { ContainerMedia } from "../../components/Presentation";
import { IconX } from "@tabler/icons-react";
import { WeblensProgress } from "../../components/WeblensProgress";
import { useResize } from "../../components/hooks";
import { BackdropMenu } from "./FileMenu";

import "./style/fileBrowserStyle.css";
import { DraggingState } from "./FileBrowser";

export const ColumnBox = ({
    children,
    style,
    reff,
    className,
    onClick,
    onMouseOver,
    onMouseLeave,
    onContextMenu,
    onBlur,
    onDragOver,
    onMouseUp,
}: {
    children?;
    style?: MantineStyleProp;
    reff?;
    className?: string;
    onClick?;
    onMouseOver?;
    onMouseLeave?;
    onContextMenu?;
    onBlur?;
    onDragOver?;
    onMouseUp?;
}) => {
    return (
        <Box
            draggable={false}
            ref={reff}
            children={children}
            onClick={onClick}
            onMouseOver={onMouseOver}
            onMouseLeave={onMouseLeave}
            onContextMenu={onContextMenu}
            onBlur={onBlur}
            onDrag={(e) => e.preventDefault()}
            onDragOver={onDragOver}
            onMouseUp={onMouseUp}
            style={{
                display: "flex",
                height: "100%",
                width: "100%",
                flexDirection: "column",
                alignItems: "center",
                ...style,
            }}
            className={`column-box ${className ? className : ""}`}
        />
    );
};

export const RowBox = ({
    children,
    style,
    onClick,
    onBlur,
}: {
    children;
    style?: MantineStyleProp;
    onClick?;
    onBlur?;
}) => {
    return (
        <Box
            draggable={false}
            children={children}
            onClick={onClick}
            onBlur={onBlur}
            onDrag={(e) => e.preventDefault()}
            style={{
                height: "100%",
                width: "100%",
                display: "flex",
                flexDirection: "row",
                alignItems: "center",
                ...style,
            }}
        />
    );
};

export const TransferCard = ({
    action,
    destination,
    boundRef,
}: {
    action: string;
    destination: string;
    boundRef?;
}) => {
    let width;
    let left;
    if (boundRef) {
        width = boundRef.clientWidth;
        left = boundRef.getBoundingClientRect()["left"];
    }
    if (!destination) {
        return;
    }

    return (
        <Box
            className="transfer-info-box"
            style={{
                pointerEvents: "none",
                width: width ? width : "100%",
                left: left ? left : 0,
            }}
        >
            <Card style={{ height: "max-content" }}>
                <RowBox>
                    <Text style={{ userSelect: "none" }}>{action} to</Text>
                    <IconFolder style={{ marginLeft: "7px" }} />
                    <Text
                        fw={700}
                        style={{ marginLeft: 3, userSelect: "none" }}
                    >
                        {destination}
                    </Text>
                </RowBox>
            </Card>
        </Box>
    );
};

export const Dropspot = ({
    onDrop,
    dropspotTitle,
    dragging,
    dropAllowed,
    handleDrag,
    wrapperRef,
}: {
    onDrop;
    dropspotTitle;
    dragging: DraggingState;
    dropAllowed;
    handleDrag: React.DragEventHandler<HTMLDivElement>;
    wrapperRef?;
}) => {
    const wrapperSize = useResize(wrapperRef);
    return (
        <Box
            draggable={false}
            className="dropspot-wrapper"
            onDragOver={(e) => {
                if (dragging === 0) {
                    handleDrag(e);
                }
            }}
            style={{
                pointerEvents: dragging === 2 ? "all" : "none",
                cursor: !dropAllowed && dragging === 2 ? "no-drop" : "auto",
                height: wrapperSize ? wrapperSize.height - 2 : "100%",
                width: wrapperSize ? wrapperSize.width - 2 : "100%",
            }}
            onDragLeave={handleDrag}
        >
            {dragging === 2 && (
                <Box
                    className="dropbox"
                    onMouseLeave={handleDrag}
                    onDrop={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        if (dropAllowed) {
                            onDrop(e);
                        }
                    }}
                    // required for onDrop to work https://stackoverflow.com/questions/50230048/react-ondrop-is-not-firing
                    onDragOver={(e) => e.preventDefault()}
                    style={{
                        outlineColor: `${dropAllowed ? "#ffffff" : "#dd2222"}`,
                        cursor:
                            !dropAllowed && dragging === 2 ? "no-drop" : "auto",
                    }}
                >
                    {!dropAllowed && (
                        <ColumnBox
                            style={{
                                position: "relative",
                                justifyContent: "center",
                                cursor: "no-drop",
                                width: "max-content",
                                pointerEvents: "none",
                            }}
                        >
                            <IconFolderCancel size={100} color="#dd2222" />
                        </ColumnBox>
                    )}
                    {dropAllowed && (
                        <TransferCard
                            action="Upload"
                            destination={dropspotTitle}
                        />
                    )}
                </Box>
            )}
        </Box>
    );
};

type DirViewWrapperProps = {
    fbState: FbStateT;
    folderName: string;
    dragging: number;
    dispatch: FBDispatchT;
    children: JSX.Element;
};

export const DirViewWrapper = ({
    fbState,
    folderName,
    dragging,
    dispatch,

    children,
}: DirViewWrapperProps) => {
    const { usr }: UserContextT = useContext(UserContext);
    const [menuOpen, setMenuOpen] = useState(false);
    const [menuPos, setMenuPos] = useState({ x: 0, y: 0 });

    return (
        <Box
            draggable={false}
            style={{
                // marginRight: 10,
                height: "100%",
                flexShrink: 0,
                minWidth: 400,
                flexGrow: 1,
                width: 0,
            }}
            onDrag={(e) => {
                e.preventDefault();
                e.stopPropagation();
            }}
            onMouseUp={(e) => {
                if (dragging) {
                    setTimeout(
                        () =>
                            dispatch({
                                type: "set_dragging",
                                dragging: DraggingState.NoDrag,
                            }),
                        10
                    );
                }
            }}
            onClick={(e) => {
                if (dragging) {
                    return;
                }
                dispatch({ type: "clear_selected" });
            }}
            onContextMenu={(e) => {
                e.preventDefault();
                if (fbState.fbMode === "share") {
                    return;
                }
                setMenuPos({ x: e.clientX, y: e.clientY });
                setMenuOpen(true);
            }}
        >
            <BackdropMenu
                folderName={folderName === usr.username ? "Home" : folderName}
                menuPos={menuPos}
                menuOpen={menuOpen}
                setMenuOpen={setMenuOpen}
                newFolder={() => dispatch({ type: "new_dir" })}
            />

            <ColumnBox
                style={{ width: "100%", padding: 8 }}
                onDragOver={(event) => {
                    if (!dragging) {
                        handleDragOver(event, dispatch, dragging);
                    }
                }}
            >
                {children}
            </ColumnBox>
        </Box>
    );
};

export const ScanFolderButton = ({ folderId, holdingShift, doScan }) => {
    return (
        <Box>
            {folderId !== "shared" && folderId !== "trash" && (
                <Tooltip
                    label={holdingShift ? "Deep scan folder" : "Scan folder"}
                >
                    <ActionIcon color="#00000000" size={35} onClick={doScan}>
                        <IconRefresh
                            color={holdingShift ? "#4444ff" : "white"}
                            size={35}
                        />
                    </ActionIcon>
                </Tooltip>
            )}
            {(folderId === "shared" || folderId === "trash") && (
                <Space w={35} />
            )}
        </Box>
    );
};

export const FileIcon = ({
    fileName,
    id,
    Icon,
    usr,
    as,
    includeText = true,
}: {
    fileName: string;
    id: string;
    Icon;
    usr: UserInfoT;
    as?: string;
    includeText?: boolean;
}) => {
    return (
        <Box
            style={{
                display: "flex",
                alignItems: "center",
            }}
        >
            <Icon className="icon-noshrink" />
            <Text
                fw={550}
                c="white"
                truncate="end"
                style={{
                    fontFamily: "monospace",
                    textWrap: "nowrap",
                    padding: 6,
                    flexShrink: 1,
                }}
            >
                {friendlyFolderName(fileName, id, usr)}
            </Text>
            {as && (
                <Box
                    style={{
                        display: "flex",
                        flexDirection: "row",
                        alignItems: "center",
                    }}
                >
                    <Text size="12px">as</Text>
                    <Text
                        size="12px"
                        truncate="end"
                        style={{
                            fontFamily: "monospace",
                            textWrap: "nowrap",
                            padding: 3,
                            flexShrink: 2,
                        }}
                    >
                        {as}
                    </Text>
                </Box>
            )}
        </Box>
    );
};

export const FolderIcon = ({ shares, size }: { shares; size }) => {
    const [copied, setCopied] = useState(false);
    const wormholeId = useMemo(() => {
        if (shares) {
            const whs = shares.filter((s) => s.Wormhole);
            if (whs.length !== 0) {
                return whs[0].shareId;
            }
        }
    }, [shares]);
    return (
        <Box
            style={{
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                width: "100%",
                height: "100%",
            }}
        >
            <IconFolder size={size} />
            {wormholeId && (
                <Tooltip label={copied ? "Copied" : "Copy Wormhole"}>
                    <IconSpiral
                        color={copied ? "#4444ff" : "white"}
                        style={{ position: "absolute", right: 0, top: 0 }}
                        onClick={(e) => {
                            e.stopPropagation();
                            navigator.clipboard.writeText(
                                `${window.location.origin}/wormhole/${shares[0].ShareId}`
                            );
                            setCopied(true);
                            setTimeout(() => setCopied(false), 1000);
                        }}
                        // onDoubleClick={(e) => e.stopPropagation()}
                    />
                </Tooltip>
            )}
        </Box>
    );
};

export const IconDisplay = ({
    file,
    size = 24,
    allowMedia = false,
}: {
    file: WeblensFile;
    size?: string | number;
    allowMedia?: boolean;
}) => {
    const [containerRef, setContainerRef] = useState(null);
    const containerSize = useResize(containerRef);

    if (!file) {
        return null;
    }

    if (file.isDir) {
        return <FolderIcon shares={file.shares} size={size} />;
    }

    if (!file.imported && file.displayable && allowMedia) {
        return (
            <Center style={{ height: "100%", width: "100%" }}>
                {/* <Skeleton height={"100%"} width={"100%"} />
                <Text pos={"absolute"} style={{ userSelect: "none" }}>
                    Processing...
                </Text> */}
                <IconPhoto />
            </Center>
        );
    } else if (file.displayable && allowMedia) {
        return (
            <ColumnBox
                reff={setContainerRef}
                style={{ justifyContent: "center" }}
            >
                <ContainerMedia
                    mediaData={file.mediaData}
                    containerRef={containerRef}
                />
            </ColumnBox>
            // <MediaImage media={file.mediaData} quality={quality} />
        );
    } else if (file.displayable) {
        return <IconPhoto />;
    }
    const extIndex = file.filename.lastIndexOf(".");
    const ext = file.filename.slice(extIndex + 1, file.filename.length);
    const textSize = `${Math.floor(containerSize?.width / (ext.length + 5))}px`;

    switch (ext) {
        case "zip":
            return <IconFileZip />;
        default:
            return (
                <Box
                    ref={setContainerRef}
                    style={{
                        display: "flex",
                        justifyContent: "center",
                        alignItems: "center",
                        width: "100%",
                        height: "100%",
                    }}
                >
                    <IconFile size={size} />
                    {extIndex !== -1 && (
                        <Text
                            size={textSize}
                            fw={700}
                            style={{
                                position: "absolute",
                                userSelect: "none",
                                WebkitUserSelect: "none",
                            }}
                        >
                            .{ext}
                        </Text>
                    )}
                </Box>
            );
    }
};

export const FileInfoDisplay = ({ file }: { file: WeblensFile }) => {
    let [size, units] = humanFileSize(file.size);
    return (
        <ColumnBox
            style={{
                width: "max-content",
                whiteSpace: "nowrap",
                justifyContent: "center",
                maxWidth: "100%",
            }}
        >
            <Text fw={600} style={{ fontSize: "2.5vw", maxWidth: "100%" }}>
                {file.filename}
            </Text>
            {file.isDir && (
                <RowBox
                    style={{
                        height: "max-content",
                        justifyContent: "center",
                        width: "100%",
                    }}
                >
                    <Text style={{ fontSize: "25px", maxWidth: "100%" }}>
                        {file.children.length} Item
                        {file.children.length !== 1 ? "s" : ""}
                    </Text>
                    <Divider orientation="vertical" size={2} mx={10} />
                    <Text style={{ fontSize: "25px" }}>
                        {size}
                        {units}
                    </Text>
                </RowBox>
            )}
            {!file.isDir && (
                <Text style={{ fontSize: "25px" }}>
                    {size}
                    {units}
                </Text>
            )}
        </ColumnBox>
    );
};

export const PresentationFile = ({ file }: { file: WeblensFile }) => {
    if (!file) {
        return null;
    }
    let [size, units] = humanFileSize(file.size);
    if (file.displayable && file.mediaData) {
        return (
            <ColumnBox
                style={{
                    justifyContent: "center",
                    width: "40%",
                    height: "max-content",
                }}
                onClick={(e) => e.stopPropagation()}
            >
                <Text
                    fw={600}
                    style={{ fontSize: "2.1vw", wordBreak: "break-all" }}
                >
                    {file.filename}
                </Text>
                <Text style={{ fontSize: "25px" }}>
                    {size}
                    {units}
                </Text>
                <Text style={{ fontSize: "20px" }}>
                    {new Date(Date.parse(file.modTime)).toLocaleDateString(
                        "en-us",
                        {
                            year: "numeric",
                            month: "short",
                            day: "numeric",
                        }
                    )}
                </Text>
                <Divider />
                <Text style={{ fontSize: "20px" }}>
                    {new Date(
                        Date.parse(file.mediaData.createDate)
                    ).toLocaleDateString("en-us", {
                        year: "numeric",
                        month: "short",
                        day: "numeric",
                    })}
                </Text>
            </ColumnBox>
        );
    } else {
        return (
            <RowBox
                style={{ justifyContent: "center", height: "max-content" }}
                onClick={(e) => e.stopPropagation()}
            >
                <Box
                    style={{
                        width: "60%",
                        display: "flex",
                        justifyContent: "center",
                    }}
                >
                    <IconDisplay file={file} allowMedia />
                </Box>
                <Space w={30} />
                <ColumnBox style={{ width: "40%", justifyContent: "center" }}>
                    <Text fw={600} style={{ width: "100%" }}>
                        {file.filename}
                    </Text>
                    {file.isDir && (
                        <RowBox
                            style={{
                                height: "max-content",
                                justifyContent: "center",
                                width: "50vw",
                            }}
                        >
                            <Text style={{ fontSize: "25px" }}>
                                {file.children.length} Item
                                {file.children.length !== 1 ? "s" : ""}
                            </Text>
                            <Divider orientation="vertical" size={2} mx={10} />
                            <Text style={{ fontSize: "25px" }}>
                                {size}
                                {units}
                            </Text>
                        </RowBox>
                    )}
                    {!file.isDir && (
                        <Text style={{ fontSize: "25px" }}>
                            {size}
                            {units}
                        </Text>
                    )}
                </ColumnBox>
            </RowBox>
        );
    }
};

const EmptyIcon = ({ folderId, usr }) => {
    if (folderId === usr.homeId) {
        return <IconHome size={500} color="#16181d" />;
    }
    if (folderId === usr.trashId) {
        return <IconTrash size={500} color="#16181d" />;
    }
    if (folderId === "shared") {
        return <IconUsers size={500} color="#16181d" />;
    }
    if (folderId === "EXTERNAL") {
        return <IconServer size={500} color="#16181d" />;
    }
    return null;
};

export const GetStartedCard = ({
    fb,
    dispatch,
    uploadDispatch,
    wsSend,
}: {
    fb: FbStateT;
    dispatch: FBDispatchT;
    uploadDispatch;
    wsSend;
}) => {
    const { authHeader, usr } = useContext(UserContext);
    return (
        <ColumnBox>
            <ColumnBox
                style={{
                    width: "max-content",
                    height: "max-content",
                    marginTop: "25vh",
                    justifyContent: "center",
                }}
            >
                <Box style={{ padding: 30, position: "absolute", zIndex: -1 }}>
                    <EmptyIcon folderId={fb.folderInfo.id} usr={usr} />
                </Box>

                <Text
                    size="28px"
                    style={{ width: "max-content", userSelect: "none" }}
                >
                    {`This folder ${
                        fb.folderInfo.pastFile ? "was" : "is"
                    } empty`}
                </Text>

                {fb.folderInfo.modifiable && !fb.viewingPast && (
                    <RowBox style={{ padding: 10, width: 350 }}>
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
                                    <ColumnBox
                                        className="get-started-box"
                                        onClick={() => {
                                            props.onClick();
                                        }}
                                    >
                                        <IconUpload
                                            size={100}
                                            stroke={"inherit"}
                                            style={{ padding: "10px" }}
                                        />
                                        <Text size="20px" fw={"inherit"}>
                                            Upload
                                        </Text>
                                        <Space h={4}></Space>
                                        <Text size="12px" fw={"inherit"}>
                                            Click or Drop
                                        </Text>
                                    </ColumnBox>
                                );
                            }}
                        </FileButton>
                        <Divider orientation="vertical" m={30} />

                        <ColumnBox
                            className="get-started-box"
                            onClick={(e) => {
                                e.stopPropagation();
                                dispatch({ type: "new_dir" });
                            }}
                        >
                            <IconFolderPlus
                                size={100}
                                stroke={"inherit"}
                                style={{ padding: "10px" }}
                            />
                            <Text
                                size="20px"
                                fw={"inherit"}
                                style={{ width: "max-content" }}
                            >
                                New Folder
                            </Text>
                        </ColumnBox>
                    </RowBox>
                )}
            </ColumnBox>
        </ColumnBox>
    );
};

export const WebsocketStatus = ({ ready }) => {
    let color;
    let status;

    switch (ready) {
        case 1:
            color = "#00ff0055";
            status = "Connected";
            break;
        case 2:
        case 3:
            color = "orange";
            status = "Connecting";
            break;
        case -1:
            color = "red";
            status = "Disconnected, try refreshing your page";
    }

    return (
        <Box style={{ position: "absolute", bottom: 0, left: 5 }}>
            <Tooltip label={status} color="#222222">
                <svg width="24" height="24" fill={color}>
                    <path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" />
                </svg>
            </Tooltip>
        </Box>
    );
};
