import { Box, Divider, Text } from "@mantine/core";
import { FileInfoT, UserContextT, UserInfoT } from "../../types/Types";
import { useResizeDrag, useWindowSize } from "../../components/hooks";
import { memo, useContext, useEffect, useMemo, useState } from "react";
import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconFile,
    IconFileImport,
    IconFileMinus,
    IconFilePlus,
    IconFolder,
    IconFolderShare,
} from "@tabler/icons-react";
import { ColumnBox, RowBox } from "./FilebrowserStyles";
import { clamp, friendlyFolderName } from "../../util";
import { WeblensButton } from "../../components/WeblensButton";
import { getFileHistory } from "../../api/FileBrowserApi";
import { userContext } from "../../Context";
import { FileEventT } from "./FileBrowserTypes";

const SIDEBAR_BREAKPOINT = 650;

export const FilesPane = memo(
    ({
        open,
        selectedFiles,
        contentId,
    }: {
        open: boolean;
        setOpen: (o) => void;
        selectedFiles: FileInfoT[];
        contentId: string;
    }) => {
        const windowSize = useWindowSize();
        const [resizing, setResizing] = useState(false);
        const [resizeOffset, setResizeOffset] = useState(
            windowSize?.width > SIDEBAR_BREAKPOINT ? 450 : 75
        );
        const [tab, setTab] = useState("info");
        useResizeDrag(
            resizing,
            setResizing,
            (v) => setResizeOffset(clamp(v, 300, 800)),
            true
        );

        if (!open) {
            return <Box className="file-info-pane" style={{ width: 0 }} />;
        }

        return (
            <Box
                className="file-info-pane"
                mod={{ "data-resizing": resizing.toString() }}
                style={{ width: resizeOffset }}
            >
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
                <Box style={{ width: 75, flexGrow: 1 }}>
                    <Box
                        style={{
                            display: "flex",
                            flexDirection: "row",
                            height: "max-content",
                            width: "100%",
                            gap: 5,
                        }}
                    >
                        <WeblensButton
                            label="File Info"
                            width={"50%"}
                            toggleOn={tab === "info"}
                            onClick={() => setTab("info")}
                            style={{ margin: 0, borderRadius: 2 }}
                        />
                        <WeblensButton
                            label="History"
                            width={"50%"}
                            toggleOn={tab === "history"}
                            onClick={() => setTab("history")}
                            style={{ margin: 0, borderRadius: 2 }}
                        />
                    </Box>
                    {tab === "info" && (
                        <FileInfo selectedFiles={selectedFiles} />
                    )}
                    {tab === "history" && <FileHistory fileId={contentId} />}
                </Box>
            </Box>
        );
    },
    (prev, next) => {
        if (prev.open !== next.open) {
            console.log(1);
            return false;
        }

        if (prev.contentId !== next.contentId) {
            console.log(2);
            return false;
        }

        if (prev.selectedFiles !== next.selectedFiles) {
            console.log(3);
            return false;
        }
        return true;
    }
);

function FileInfo({ selectedFiles }) {
    const titleText = useMemo(() => {
        if (selectedFiles.length === 0) {
            return "No files selected";
        } else if (selectedFiles.length === 1) {
            return selectedFiles[0].filename;
        } else {
            return `${selectedFiles.length} files selected`;
        }
    }, [selectedFiles]);

    const singleItem = selectedFiles.length === 1;
    const itemIsFolder = selectedFiles[0]?.isDir;

    return (
        <Box className="file-info-content">
            <RowBox
                style={{
                    height: "58px",
                    justifyContent: "space-between",
                }}
            >
                <Text
                    size="24px"
                    fw={600}
                    style={{ textWrap: "nowrap", paddingRight: 32 }}
                >
                    {titleText}
                </Text>
            </RowBox>
            {selectedFiles.length > 0 && (
                <ColumnBox style={{ height: "max-content" }}>
                    <RowBox>
                        {singleItem && itemIsFolder && (
                            <IconFolder size={"48px"} />
                        )}
                        {(!singleItem || !itemIsFolder) && (
                            <IconFile size={"48px"} />
                        )}
                        <Text size="20px">
                            {itemIsFolder ? "Folder" : "File"}
                            {singleItem ? "" : "s"} Info
                        </Text>
                    </RowBox>
                    <Divider h={2} w={"90%"} m={10} />
                </ColumnBox>
            )}
        </Box>
    );
}

const historyDate = (timestamp: number) => {
    const dateObj = new Date(timestamp);
    const options: Intl.DateTimeFormatOptions = {
        month: "long",
        day: "numeric",
        minute: "numeric",
        hour: "numeric",
    };
    if (dateObj.getFullYear() !== new Date().getFullYear()) {
        options.year = "numeric";
    }
    return dateObj.toLocaleDateString("en-US", options);
};

const HistoryRow = ({
    eventGroup,
    filePath,
    usr,
}: {
    eventGroup: {
        count: number;
        action: string;
        time: number;
        events: FileEventT[];
    };
    filePath: string;
    usr: UserInfoT;
}) => {
    const [open, setOpen] = useState(false);
    const timeStr = historyDate(eventGroup.time);

    if (filePath.lastIndexOf("/") === filePath.length - 1) {
        filePath = filePath.slice(0, -1);
    }
    let folderName;
    let slashIndex = filePath.lastIndexOf("/");
    if (slashIndex === -1) {
        folderName = filePath.slice(filePath.indexOf(":") + 1);
    } else {
        folderName = filePath.slice(slashIndex + 1);
    }

    return (
        <Box
            style={{
                width: "100%",
                display: "flex",
                flexDirection: "column",
                height: "max-content",
            }}
        >
            <Box
                key={eventGroup.time}
                className="file-history-summary"
                onClick={() => setOpen(!open)}
            >
                {open ? (
                    <IconCaretDown size={20} style={{ flexShrink: 0 }} />
                ) : (
                    <IconCaretRight size={20} style={{ flexShrink: 0 }} />
                )}
                <Text
                    c="white"
                    truncate="end"
                    style={{
                        lineHeight: "20px",
                        fontSize: "20px",
                        width: "100%",
                        padding: "0px 5px 0px 5px",

                        textWrap: "nowrap",
                    }}
                    fw={600}
                >
                    {eventGroup.count} File{eventGroup.count !== 1 ? "s" : ""}{" "}
                    {eventGroup.action.slice(4)}d
                </Text>
                <Text style={{ textWrap: "nowrap" }}>{timeStr}</Text>
            </Box>
            <Box
                className="file-history-detail-accordion"
                mod={{ "data-open": open.toString() }}
                style={{ height: open ? eventGroup.count * 30 + 25 : 0 }}
            >
                <Box className={"file-history-detail"}>
                    {eventGroup.events.map((e) => {
                        const fileFolderPath = e.Path.slice(
                            e.Path.indexOf(":") + 1,
                            e.Path.lastIndexOf("/")
                        );
                        console.log(e.Path);

                        const fileFolderName = fileFolderPath.slice(
                            fileFolderPath.lastIndexOf("/") + 1
                        );

                        const fileName = e.Path.slice(
                            e.Path.lastIndexOf("/") + 1
                        );

                        let moveDirection;
                        console.log(fileFolderName);
                        if (
                            e.Action === "fileCreate" &&
                            fileFolderName === filePath
                        ) {
                        }

                        return (
                            <Box
                                key={`${e.Path}-${e.millis}`}
                                className="history-detail-action-row"
                            >
                                {e.Action === "fileMove" && (
                                    <IconFile className="icon-noshrink" />
                                )}
                                <Text
                                    fw={550}
                                    c="white"
                                    truncate="end"
                                    style={{
                                        textWrap: "nowrap",
                                        paddingRight: 6,
                                        flexShrink: 1,
                                    }}
                                >
                                    {e.fileName}
                                </Text>
                                {e.Action === "fileCreate" && (
                                    <IconFilePlus className="icon-noshrink" />
                                )}
                                {e.Action === "fileDelete" && (
                                    <IconFileMinus className="icon-noshrink" />
                                )}
                                {e.Action === "fileMove" && (
                                    <IconArrowRight className="icon-noshrink" />
                                )}
                                {e.Action === "fileMove" && (
                                    <IconFolder className="icon-noshrink" />
                                )}
                                <Text
                                    fw={550}
                                    c="white"
                                    style={{ flexGrow: 1, paddingLeft: 2 }}
                                    truncate="end"
                                >
                                    {eventGroup.action === "fileMove"
                                        ? friendlyFolderName(
                                              fileFolderName,
                                              "",
                                              usr
                                          )
                                        : fileName}
                                </Text>
                                <Text style={{ textWrap: "nowrap" }}>
                                    {new Date(e.millis).toLocaleTimeString()}
                                </Text>
                            </Box>
                        );
                    })}
                </Box>
            </Box>
        </Box>
    );
};

function FileHistory({ fileId }: { fileId: string }) {
    const { authHeader, usr }: UserContextT = useContext(userContext);
    const [fileHistory, setFileHistory]: [
        fileHistory: FileEventT[],
        setFileHistory: any
    ] = useState([]);

    useEffect(() => {
        setFileHistory([]);
        getFileHistory(fileId, authHeader).then((r) =>
            setFileHistory(r.events)
        );
    }, [fileId]);

    const { createEvent, eventGroups } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return {};
        }

        const createEvent = fileHistory.shift();
        if (createEvent.FileId === usr.homeId) {
            createEvent.fileName = "Home";
        } else if (createEvent.FileId === usr.trashId) {
            createEvent.fileName = "Trash";
        } else {
            let lastSlash = createEvent.Path.lastIndexOf("/");

            createEvent.fileName =
                lastSlash !== createEvent.Path.length
                    ? createEvent.Path.slice(lastSlash + 1)
                    : createEvent.Path.slice(createEvent.Path.indexOf(":") + 1);
        }

        const groupMap = new Map<number, FileEventT[]>();
        const filenameMap = new Map<string, string>();
        fileHistory.forEach((e) => {
            let millis = Date.parse(e.Timestamp);
            let groupMillis =
                Math.floor((millis - (millis % 1000)) / 6000) * 6000;

            e.millis = millis;
            let group = groupMap.get(groupMillis);
            if (group === undefined) {
                group = [e];
            } else {
                group.push(e);
            }
            groupMap.set(groupMillis, group);
            if (e.Path !== "") {
                let name = e.Path.slice(e.Path.lastIndexOf("/") + 1);
                if (e.Action === "fileCreate") {
                    filenameMap.set(e.FileId, name);
                } else if (e.Action === "fileMove") {
                    let dir = e.Path.slice(0, e.Path.lastIndexOf("/"));
                    if (dir !== createEvent.Path) {
                        e.Path = e.fileName = filenameMap.get(e.FileId);
                        e.fileName = e.fileName?.slice(
                            e.fileName?.lastIndexOf("/") + 1
                        );
                    } else {
                        e.oldPath = filenameMap.get(e.FileId);
                    }
                    filenameMap.set(e.SecondaryFileId, name);
                }
            } else {
                e.Path = filenameMap.get(e.FileId);
            }
        });

        const groups = Array.from(groupMap.values());

        return {
            createEvent: createEvent,
            eventGroups: groups.map((g) => {
                return {
                    count: g.length,
                    action: g[0].Action,
                    time: g[0].millis,
                    events: g,
                };
            }),
        };
    }, [fileHistory]);

    if (!createEvent) {
        return null;
    }

    const createTimeString = historyDate(Date.parse(createEvent.Timestamp));

    return (
        <Box
            style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                padding: 10,
            }}
        >
            <Box
                style={{
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    padding: 10,
                }}
            >
                <IconFolder />
                <Text fw={600} style={{ padding: 5 }}>
                    {createEvent.fileName}
                </Text>
                <Text style={{ textWrap: "nowrap" }}>
                    created on {createTimeString}
                </Text>
            </Box>
            {eventGroups.map((eg) => {
                return (
                    <HistoryRow
                        key={eg.events[0].FileId + eg.time}
                        eventGroup={eg}
                        filePath={createEvent.Path}
                        usr={usr}
                    />
                );
            })}
        </Box>
    );
}
