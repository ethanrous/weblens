import { Box, Divider, Text } from "@mantine/core";
import {
    FBDispatchT,
    FileInfoT,
    UserContextT,
    UserInfoT,
} from "../../types/Types";
import { useResizeDrag, useWindowSize } from "../../components/hooks";
import { memo, useContext, useEffect, useMemo, useState } from "react";
import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconFile,
    IconFileMinus,
    IconFilePlus,
    IconFolder,
    IconReorder,
} from "@tabler/icons-react";
import { ColumnBox, RowBox } from "./FilebrowserStyles";
import { clamp, friendlyFolderName } from "../../util";
import { WeblensButton } from "../../components/WeblensButton";
import { getFileHistory } from "../../api/FileBrowserApi";
import { userContext } from "../../Context";
import { FileEventT } from "./FileBrowserTypes";
import "./style/history.css";

const SIDEBAR_BREAKPOINT = 650;

export const FilesPane = memo(
    ({
        open,
        selectedFiles,
        contentId,
        timestamp,
        dispatch,
    }: {
        open: boolean;
        setOpen: (o) => void;
        selectedFiles: FileInfoT[];
        contentId: string;
        timestamp: number;
        dispatch: FBDispatchT;
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
                <Box
                    style={{
                        width: 75,
                        flexGrow: 1,
                        display: "flex",
                        flexDirection: "column",
                    }}
                >
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
                    {tab === "history" && (
                        <FileHistory
                            fileId={contentId}
                            timestamp={timestamp}
                            dispatch={dispatch}
                        />
                    )}
                </Box>
            </Box>
        );
    },
    (prev, next) => {
        if (prev.open !== next.open) {
            return false;
        }

        if (prev.contentId !== next.contentId) {
            return false;
        }

        if (prev.selectedFiles !== next.selectedFiles) {
            return false;
        }

        if (prev.timestamp !== next.timestamp) {
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

const portableToFolderName = (path: string) => {
    const folderPath = path.slice(path.indexOf(":") + 1, path.lastIndexOf("/"));
    const folderName = folderPath.slice(folderPath.lastIndexOf("/") + 1);
    return folderName;
};
const fileBase = (path: string) => {
    if (path[path.length - 1] === "/") {
        path = path.slice(0, path.length - 1);
    }
    return path.slice(path.lastIndexOf("/") + 1);
};

const FileIcon = ({
    fileName,
    id,
    Icon,
    usr,
    as,
}: {
    fileName: string;
    id: string;
    Icon;
    usr: UserInfoT;
    as?: string;
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

const HistoryRow = ({
    eventGroup,
    folderPath,
    viewing,
    usr,
    dispatch,
}: {
    eventGroup: {
        count: number;
        action: string;
        time: number;
        events: FileEventT[];
    };
    viewing: boolean;
    folderPath: string;
    usr: UserInfoT;
    dispatch: FBDispatchT;
}) => {
    const [open, setOpen] = useState(false);
    const timeStr = historyDate(eventGroup.time);

    const folderName = portableToFolderName(folderPath);

    return (
        <Box
            style={{
                width: "100%",
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                height: "max-content",
                padding: 10,
                borderRadius: 8,
                backgroundColor: viewing ? "#1c1049" : "",
                outline: viewing ? "1px solid #4444ff" : "",
            }}
        >
            <Box key={eventGroup.time} className="file-history-summary">
                <Box
                    className="file-history-accordion-header"
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
                            width: "max-content",
                            padding: "0px 5px 0px 5px",

                            textWrap: "nowrap",
                        }}
                        fw={600}
                    >
                        {eventGroup.count} File
                        {eventGroup.count !== 1 ? "s" : ""}{" "}
                        {eventGroup.action.slice(4)}d
                    </Text>
                </Box>
                <Box style={{ flexGrow: 1 }} />
                <Text
                    className="event-time-string"
                    fw={viewing ? 600 : 400}
                    c={viewing ? "white" : ""}
                    onClick={() =>
                        dispatch({
                            type: "set_past_time",
                            past: viewing
                                ? null
                                : new Date(eventGroup.events[0].Timestamp),
                        })
                    }
                >
                    {timeStr}
                </Text>
            </Box>
            <Box
                className="file-history-detail-accordion"
                mod={{ "data-open": open.toString() }}
                style={{ height: open ? eventGroup.count * 30 + 35 : 0 }}
            >
                <Box className={"file-history-detail"}>
                    {eventGroup.events.map((e) => {
                        const fromFolder = portableToFolderName(e.FromPath);
                        const toFolder = portableToFolderName(e.Path);

                        const fromFile = fileBase(e.FromPath);
                        const toFile = fileBase(e.Path);

                        return (
                            <Box
                                key={`${e.Path}-${e.millis}`}
                                className="history-detail-action-row"
                            >
                                {e.Action === "fileMove" &&
                                    folderName === fromFolder && (
                                        <FileIcon
                                            id={e.FromFileId}
                                            fileName={fromFile}
                                            Icon={IconFile}
                                            usr={usr}
                                        />
                                    )}
                                {e.Action === "fileMove" &&
                                    folderName !== fromFolder && (
                                        <FileIcon
                                            id={e.FromFileId}
                                            fileName={fromFolder}
                                            Icon={IconFolder}
                                            usr={usr}
                                            as={
                                                toFile !== fromFile
                                                    ? fromFile
                                                    : ""
                                            }
                                        />
                                    )}
                                {e.Action === "fileCreate" && (
                                    <FileIcon
                                        id={e.FileId}
                                        fileName={toFile}
                                        Icon={IconFilePlus}
                                        usr={usr}
                                    />
                                )}
                                {e.Action === "fileDelete" && (
                                    <FileIcon
                                        fileName={fromFile}
                                        Icon={IconFileMinus}
                                        id={e.FromFileId}
                                        usr={usr}
                                    />
                                    // <IconFileMinus className="icon-noshrink" />
                                )}
                                {e.Action === "fileRestore" && (
                                    <FileIcon
                                        fileName={toFile}
                                        Icon={IconReorder}
                                        id={e.FileId}
                                        usr={usr}
                                    />
                                    // <IconFileMinus className="icon-noshrink" />
                                )}
                                {e.Action === "fileMove" && (
                                    <IconArrowRight className="icon-noshrink" />
                                )}

                                {e.Action === "fileMove" &&
                                    folderName === toFolder && (
                                        <FileIcon
                                            id={e.FileId}
                                            fileName={toFile}
                                            Icon={IconFile}
                                            usr={usr}
                                        />
                                    )}

                                {e.Action === "fileMove" &&
                                    folderName !== toFolder && (
                                        <FileIcon
                                            id={e.FileId}
                                            fileName={toFolder}
                                            Icon={IconFolder}
                                            usr={usr}
                                            as={
                                                toFile !== fromFile
                                                    ? toFile
                                                    : ""
                                            }
                                        />
                                    )}

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

function FileHistory({
    fileId,
    timestamp,
    dispatch,
}: {
    fileId: string;
    timestamp: number;
    dispatch: FBDispatchT;
}) {
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
    }, [fileId, timestamp]);

    const { createEvent, eventGroups } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return {};
        }

        const createEvent = fileHistory.shift();

        const groupMap = new Map<string, FileEventT[]>();
        fileHistory.forEach((e) => {
            let millis = Date.parse(e.Timestamp);
            let groupMillis = millis - (millis % 60000);

            e.millis = millis;
            const groupKey = `${groupMillis}-${e.Action}`;
            let group = groupMap.get(groupKey);
            if (group === undefined) {
                group = [e];
            } else {
                group.push(e);
            }
            groupMap.set(groupKey, group);
        });

        const groups = Array.from(groupMap.values());

        return {
            createEvent: createEvent,
            eventGroups: groups
                .map((g) => {
                    g.sort((a, b) => b.millis - a.millis);
                    return {
                        count: g.length,
                        action: g[0].Action,
                        time: g[0].millis,
                        events: g,
                    };
                })
                .sort((a, b) => b.time - a.time),
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
                overflow: "scroll",
                height: 200,
                flexGrow: 1,
            }}
        >
            {eventGroups.map((eg) => {
                return (
                    <HistoryRow
                        key={eg.events[0].FileId + eg.time}
                        eventGroup={eg}
                        folderPath={createEvent.Path}
                        viewing={eg.time === timestamp}
                        usr={usr}
                        dispatch={dispatch}
                    />
                );
            })}
            <Box
                style={{
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    padding: 10,
                    paddingTop: 40,
                }}
            >
                <FileIcon
                    id={createEvent.FileId}
                    fileName={createEvent.Path}
                    Icon={IconFolder}
                    usr={usr}
                />

                <Text style={{ textWrap: "nowrap" }}>
                    created on {createTimeString}
                </Text>
            </Box>
        </Box>
    );
}
