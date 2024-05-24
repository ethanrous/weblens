import { Box, Space, Text } from "@mantine/core";
import { FixedSizeList as ScrollList } from "react-window";
import { FBDispatchT, FbStateT } from "../../types/Types";
import { WeblensFile } from "../../classes/File";
import { useEffect, useMemo, useState } from "react";
import { IconDisplay } from "./FileBrowserStyles";
import { useNavigate } from "react-router-dom";
import { humanFileSize } from "../../util";
import { useResize } from "../../components/hooks";

function FileRow({
    data,
    index,
    style,
}: {
    data: {
        files: WeblensFile[];
        selected: Map<string, boolean>;
        lastSelectedIndex: number;
        hoveringIndex: number;
        dispatch: FBDispatchT;
    };
    index: number;
    style;
}) {
    const nav = useNavigate();
    const inRange =
        (index - data.lastSelectedIndex) * (index - data.hoveringIndex) < 1;

    return (
        <Box style={{ ...style, display: "flex" }}>
            <Box
                className="file-row"
                mod={{
                    "data-selected":
                        data.selected.get(data.files[index].Id()) === true,
                    "data-in-range": inRange === true,
                }}
                style={{
                    border:
                        index === data.lastSelectedIndex
                            ? "2px solid #442299"
                            : "2px solid #00000000",
                }}
                onClick={(e) => {
                    e.stopPropagation();
                    data.dispatch({
                        type: "set_selected",
                        fileId: data.files[index].Id(),
                        selected:
                            data.selected.get(data.files[index].Id()) !== true,
                    });
                }}
                onMouseOver={(e) => {
                    data.dispatch({
                        type: "set_hovering",
                        hovering: data.files[index].Id(),
                    });
                }}
                onContextMenu={(e) => {
                    e.stopPropagation();
                    e.preventDefault();
                    data.dispatch({
                        type: "set_menu_target",
                        fileId: data.files[index].Id(),
                    });
                    data.dispatch({ type: "set_menu_open", open: true });
                    data.dispatch({
                        type: "set_menu_pos",
                        pos: { x: e.clientX, y: e.clientY },
                    });
                }}
                onDoubleClick={() => {
                    if (data.files[index].IsFolder()) {
                        nav(`/files/${data.files[index].Id()}`);
                    } else {
                        nav(
                            `/files/${data.files[
                                index
                            ].ParentId()}?jumpTo=${data.files[index].Id()}`
                        );
                    }
                }}
            >
                <Box style={{ width: 24, height: 24 }}>
                    <IconDisplay file={data.files[index]} />
                </Box>
                <Text>{data.files[index].GetFilename()}</Text>
                <Space flex={1} />
                <Text>{data.files[index].FormatSize()}</Text>
            </Box>
        </Box>
    );
}

export function FileRows({
    fb,
    dispatch,
}: {
    fb: FbStateT;
    dispatch: FBDispatchT;
}) {
    const [boxRef, setBoxRef] = useState(null);
    const size = useResize(boxRef);

    const files = useMemo(() => {
        const filesList = Array.from(fb.dirMap.values());
        return filesList;
    }, [fb.dirMap]);

    useEffect(() => {
        dispatch({
            type: "set_files_list",
            fileIds: files.map((v) => v.Id()),
        });
    }, [files]);

    const lastSelectedIndex = useMemo(() => {
        return files.findIndex((v) => v.Id() === fb.lastSelected);
    }, [files, fb.lastSelected]);

    const hoveringIndex = useMemo(() => {
        if (!fb.holdingShift || !files) {
            return { hoveringIndex: -1, lastSelectedIndex: -1 };
        }
        return files.findIndex((v) => v.Id() === fb.hovering);
    }, [files, fb.holdingShift, fb.hovering]);

    return (
        <Box
            ref={setBoxRef}
            className="flex flex-col items-center w-full h-full ml-1"
        >
            <ScrollList
                className="no-scrollbar"
                height={boxRef?.clientHeight ? boxRef.clientHeight : 0}
                width={size.width}
                itemSize={52}
                itemCount={files.length}
                itemData={{
                    files: files,
                    selected: fb.selected,
                    lastSelectedIndex: lastSelectedIndex,
                    hoveringIndex: hoveringIndex,
                    dispatch: dispatch,
                }}
                overscanRowCount={50}
            >
                {FileRow}
            </ScrollList>
        </Box>
    );
}
