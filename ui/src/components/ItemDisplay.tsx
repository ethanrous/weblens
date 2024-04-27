import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
    AspectRatio,
    Box,
    Divider,
    Loader,
    Text,
    Tooltip,
} from "@mantine/core";
import { MediaImage } from "./PhotoContainer";
import { WeblensFile } from "../classes/File";
import { getVisitRoute } from "../Pages/FileBrowser/FileBrowserLogic";
import { useNavigate } from "react-router-dom";
import { DraggingState } from "../Pages/FileBrowser/FileBrowser";

import "./itemDisplayStyle.css";
import { IconFolder } from "@tabler/icons-react";

export type GlobalContextType = {
    setDragging: (d: DraggingState) => void;
    blockFocus: (b: boolean) => void;
    rename: (itemId: string, newName: string) => void;

    setMenuOpen: (o: boolean) => void;
    setMenuPos: ({ x, y }: { x: number; y: number }) => void;
    setMenuTarget: (itemId: string) => void;

    setHovering?: (itemId: string) => void;
    setSelected?: (itemId: string, selected?: boolean) => void;
    selectAll?: (itemId: string, selected?: boolean) => void;
    moveSelected?: (itemId: string) => void;
    doSelectMany?: () => void;
    setMoveDest?: (itemName) => void;

    dragging?: number;
    numCols?: number;
    itemWidth?: number;
    initialScrollIndex?: number;
    hoveringIndex?: number;
    lastSelectedIndex?: number;
    doMediaFetch?: boolean;
    allowEditing?: boolean;
};

// export type ItemProps = {
//     itemId: string;
//     itemTitle: string;
//     itemSize?: number;
//     itemSizeBytes?: number;
//     itemSizeUnits?: string;
//     modifyDate?: Date;
//     selected: number;
//     mediaData: WeblensMedia;
//     droppable: boolean;
//     isDir: boolean;
//     imported: boolean;
//     displayable: boolean;
//     dragging?: number;
//     dispatch?: any;
//     shares?: any[];
//     extraIcons?: any[];
//     index?: number;
// };

type WrapperProps = {
    itemInfo: WeblensFile;
    fileRef;

    editing: boolean;

    selected: SelectedState;

    width: number;
    index: number;
    lastSelectedIndex?: number;
    hoverIndex?: number;
    dragging: DraggingState;

    setSelected: (itemId: string, selected?: boolean) => void;
    doSelectMany: () => void;
    moveSelected: (entryId: string) => void;
    setMoveDest: (itemName: string) => void;

    setDragging: (d: DraggingState) => void;
    setHovering: (i: string) => void;

    setMenuOpen: (o: boolean) => void;
    setMenuPos: ({ x, y }: { x: number; y: number }) => void;
    setMenuTarget: (itemId: string) => void;

    children;
};

type TitleProps = {
    itemId: string;
    itemTitle: string;
    secondaryInfo?: string;
    editing: boolean;
    setEditing: (e: boolean) => void;
    allowEditing: boolean;
    height: number;
    blockFocus: (b: boolean) => void;
    rename: (itemId: string, newName: string) => void;
};

const MARGIN = 6;

function selectedStyles(selected: SelectedState): {
    backgroundColor: string;
    outline: string;
} {
    let backgroundColor = "#222222";
    let outline;
    if (selected & SelectedState.Hovering) {
        backgroundColor = "#333333";
    }

    if (selected & SelectedState.InRange) {
        backgroundColor = "#373365";
    }

    if (selected & SelectedState.Selected) {
        backgroundColor = "#331177bb";
    }

    if (selected & SelectedState.LastSelected) {
        outline = "2px solid #442299";
    }

    if (selected & SelectedState.Droppable) {
        backgroundColor = "#1c1049";
        outline = "2px solid #4444ff";
    }

    return { backgroundColor, outline };
}

const ItemWrapper = memo(
    ({
        itemInfo: file,
        fileRef,
        width,
        selected,
        index,
        lastSelectedIndex,
        hoverIndex,
        setSelected,
        doSelectMany,
        dragging = 0,
        setDragging,
        setHovering,
        moveSelected,
        setMenuOpen,
        setMenuPos,
        setMenuTarget,
        setMoveDest,
        children,
    }: WrapperProps) => {
        const [mouseDown, setMouseDown] = useState(null);
        const nav = useNavigate();

        const { outline, backgroundColor } = selectedStyles(selected);

        return (
            <Box
                ref={fileRef}
                style={{ margin: MARGIN }}
                onMouseOver={(e) => {
                    e.stopPropagation();
                    // setH(true);s
                    file.SetHovering(true);
                    setHovering(file.Id());
                    if (dragging && !file.IsSelected() && file.IsFolder()) {
                        setMoveDest(file.GetFilename());
                    }
                }}
                onMouseDown={(e) => {
                    setMouseDown({ x: e.clientX, y: e.clientY });
                }}
                onMouseMove={(e) => {
                    if (
                        mouseDown &&
                        !dragging &&
                        (Math.abs(mouseDown.x - e.clientX) > 20 ||
                            Math.abs(mouseDown.y - e.clientY) > 20)
                    ) {
                        setSelected(file.Id(), true);
                        setDragging(DraggingState.InternalDrag);
                    }
                }}
                onClick={(e) => {
                    e.stopPropagation();
                    if (e.shiftKey) {
                        doSelectMany();
                    } else {
                        setSelected(file.Id());
                    }
                }}
                onMouseUp={(e) => {
                    // e.stopPropagation();
                    if (dragging !== 0) {
                        if (!file.IsSelected() && file.IsFolder()) {
                            moveSelected(file.Id());
                        }
                        setMoveDest("");
                        setDragging(DraggingState.NoDrag);
                    }
                    setMouseDown(null);
                }}
                onDoubleClick={(e) => {
                    e.stopPropagation();
                    const jump = getVisitRoute(
                        "default",
                        file.Id(),
                        "",
                        file.IsFolder(),
                        file.GetMedia()?.IsDisplayable(),
                        () => {}
                    );
                    if (jump) {
                        nav(jump);
                    }
                }}
                onContextMenu={(e) => {
                    e.preventDefault();
                    e.stopPropagation();

                    setMenuTarget(file.Id());
                    setMenuPos({ x: e.clientX, y: e.clientY });
                    setMenuOpen(true);
                }}
                onMouseLeave={(e) => {
                    file.SetHovering(false);
                    setHovering("");
                    if (dragging && file.IsFolder()) {
                        setMoveDest("");
                    }
                    if (mouseDown) {
                        setMouseDown(null);
                    }
                }}
            >
                <Box
                    className="item-child"
                    children={children}
                    style={{
                        outline: outline,
                        backgroundColor: backgroundColor,
                        height: (width - MARGIN * 2) * 1.1,
                        width: width - MARGIN * 2,
                        cursor:
                            dragging !== 0 && !file.IsFolder()
                                ? "default"
                                : "pointer",
                    }}
                />
                {(file.IsSelected() || !file.IsFolder()) && dragging !== 0 && (
                    <Box
                        className="no-drop-cover"
                        style={{
                            height: (width - MARGIN * 2) * 1.1,
                            width: width - MARGIN * 2,
                        }}
                        onMouseLeave={(e) => {
                            file.SetHovering(false);
                            setHovering("");
                        }}
                        onClick={(e) => e.stopPropagation()}
                    />
                )}
            </Box>
        );
    },
    (prev, next) => {
        if (prev.itemInfo !== next.itemInfo) {
            return false;
        } else if (prev.index !== next.index) {
            return false;
        } else if (prev.hoverIndex !== next.hoverIndex) {
            return false;
        } else if (prev.selected !== next.selected) {
            return false;
        } else if (prev.lastSelectedIndex !== next.lastSelectedIndex) {
            return false;
        } else if (prev.editing !== next.editing) {
            return false;
        } else if (prev.dragging !== next.dragging) {
            return false;
        } else if (prev.width !== next.width) {
            return false;
        }
        return true;
    }
);

const FileVisualWrapper = ({ children }) => {
    return (
        <AspectRatio ratio={1} w={"94%"} display={"flex"} m={"6px"}>
            <Box
                children={children}
                style={{ overflow: "hidden", borderRadius: "5px" }}
            />
        </AspectRatio>
    );
};

const FileVisual = ({
    file,
    doFetch,
}: {
    file: WeblensFile;
    doFetch: boolean;
}) => {
    if (file.IsFolder()) {
        return <IconFolder size={150} />;
    }

    if (file.GetMedia()?.IsDisplayable()) {
        return (
            <MediaImage
                media={file.GetMedia()}
                quality="thumbnail"
                doFetch={doFetch}
            />
        );
    }

    return null;
};

const useKeyDown = (
    itemId: string,
    oldName: string,
    newName: string,
    editing: boolean,
    setEditing: (b: boolean) => void,
    rename: (itemId: string, newName: string) => void
) => {
    const onKeyDown = useCallback(
        (event) => {
            if (!editing) {
                return;
            }
            if (event.key === "Enter") {
                if (oldName !== newName) {
                    rename(itemId, newName);
                }
                setEditing(false);
            } else if (event.key === "Escape") {
                setEditing(false);
                // Rename with empty name is a "cancel" to the rename
                rename(itemId, "");
            }
        },
        [itemId, oldName, newName, editing, setEditing, rename]
    );

    useEffect(() => {
        document.addEventListener("keydown", onKeyDown);
        return () => {
            document.removeEventListener("keydown", onKeyDown);
        };
    }, [onKeyDown]);
};

const TextBox = ({
    itemId,
    itemTitle,
    secondaryInfo,
    editing,
    setEditing,
    allowEditing,
    height,
    blockFocus,
    rename,
}: TitleProps) => {
    const editRef: React.Ref<HTMLInputElement> = useRef();
    const [renameVal, setRenameVal] = useState(itemTitle);

    const setEditingPlus = useCallback(
        (b: boolean) => {
            setEditing(b);
            setRenameVal((cur) => {
                if (cur === "") {
                    return itemTitle;
                } else {
                    return cur;
                }
            });
            blockFocus(b);
        },
        [itemTitle, setEditing, blockFocus]
    );
    useKeyDown(itemId, itemTitle, renameVal, editing, setEditingPlus, rename);

    useEffect(() => {
        if (editing && editRef.current) {
            editRef.current.select();
        }
    }, [editing, editRef]);

    useEffect(() => {
        if (itemId === "NEW_DIR") {
            setEditingPlus(true);
        }
    }, [itemId, setEditingPlus]);

    if (editing) {
        return (
            <Box
                className="item-info-box"
                style={{
                    height: height,
                }}
                onBlur={() => {
                    setEditingPlus(false);
                    rename(itemId, "");
                }}
            >
                <input
                    ref={editRef}
                    defaultValue={itemTitle}
                    onClick={(e) => {
                        e.stopPropagation();
                    }}
                    onDoubleClick={(e) => {
                        e.stopPropagation();
                    }}
                    onChange={(e) => {
                        setRenameVal(e.target.value);
                    }}
                    style={{
                        width: "90%",
                        backgroundColor: "#00000000",
                        border: 0,
                        outline: 0,
                    }}
                />
            </Box>
        );
    } else {
        return (
            <Box
                className="item-info-box"
                style={{
                    height: height,
                    cursor: allowEditing ? "text" : "default",
                    paddingBottom: MARGIN / 2,
                }}
                onClick={(e) => {
                    if (!allowEditing) {
                        return;
                    }
                    e.stopPropagation();
                    setEditingPlus(true);
                }}
            >
                <Box className="title-box">
                    <Text
                        size={`${height - MARGIN * 2}px`}
                        truncate={"end"}
                        style={{
                            color: "white",
                            userSelect: "none",
                            lineHeight: 1.5,
                        }}
                    >
                        {itemTitle}
                    </Text>
                    <Divider orientation="vertical" my={1} mx={6} />
                    <Box
                        style={{
                            width: "max-content",
                            justifyContent: "center",
                        }}
                    >
                        <Text
                            size={`${height - (MARGIN * 2 + 4)}px`}
                            lineClamp={1}
                            style={{
                                color: "white",
                                overflow: "visible",
                                userSelect: "none",
                                width: "max-content",
                            }}
                        >
                            {" "}
                            {secondaryInfo}{" "}
                        </Text>
                    </Box>
                </Box>
                <Tooltip openDelay={300} label={itemTitle}>
                    <Box
                        style={{
                            position: "absolute",
                            width: "90%",
                            height: height,
                        }}
                    />
                </Tooltip>
            </Box>
        );
    }
};

export enum SelectedState {
    NotSelected = 0x0,
    Hovering = 0x1,
    InRange = 0x10,
    Selected = 0x100,
    LastSelected = 0x1000,
    Droppable = 0x10000,
}

export const FileDisplay = memo(
    ({
        file,
        selected,
        index,
        context,
    }: {
        file: WeblensFile;
        selected: SelectedState;
        index: number;
        context: GlobalContextType;
    }) => {
        const wrapRef = useRef();
        const [editing, setEditing] = useState(false);

        return (
            <ItemWrapper
                itemInfo={file}
                fileRef={wrapRef}
                index={index}
                selected={selected}
                lastSelectedIndex={context.lastSelectedIndex}
                hoverIndex={context.hoveringIndex}
                setSelected={context.setSelected}
                doSelectMany={context.doSelectMany}
                width={context.itemWidth}
                moveSelected={context.moveSelected}
                dragging={context.dragging}
                setDragging={context.setDragging}
                setHovering={context.setHovering}
                setMoveDest={context.setMoveDest}
                setMenuOpen={context.setMenuOpen}
                setMenuPos={context.setMenuPos}
                setMenuTarget={context.setMenuTarget}
                editing={editing}
            >
                <FileVisualWrapper>
                    <FileVisual file={file} doFetch={context.doMediaFetch} />
                </FileVisualWrapper>

                <TextBox
                    itemId={file.Id()}
                    itemTitle={file.GetFilename()}
                    secondaryInfo={file.FormatSize()}
                    editing={editing}
                    setEditing={(e) => {
                        if (!context.allowEditing) {
                            return;
                        }
                        setEditing(e);
                    }}
                    allowEditing={context.allowEditing}
                    height={context.itemWidth * 0.1}
                    blockFocus={context.blockFocus}
                    rename={(id, newName) => {
                        if (
                            newName === file.GetFilename() ||
                            (newName === "" && file.Id() !== "NEW_DIR")
                        ) {
                            return;
                        }
                        context.rename(id, newName);
                    }}
                />

                {file.Id() === "NEW_DIR" && !editing && (
                    <Loader
                        color="white"
                        size={20}
                        style={{ position: "absolute", top: 20, right: 20 }}
                    />
                )}
            </ItemWrapper>
        );
    },
    (prev, next) => {
        if (prev.file.Id() !== next.file.Id()) {
            return false;
        } else if (prev.context !== next.context) {
            return false;
        } else if (prev.context.itemWidth !== next.context.itemWidth) {
            return false;
        } else if (prev.context.dragging !== next.context.dragging) {
            return false;
        } else if (prev.file.IsSelected() !== next.file.IsSelected()) {
            return false;
        } else if (prev.file.IsHovering() !== next.file.IsHovering()) {
            return false;
        }
        return true;
    }
);
