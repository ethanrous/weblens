import {
    IconArrowBackUp,
    IconChevronRight,
    IconDownload,
    IconFile,
    IconFolder,
    IconFolderPlus,
    IconHistory,
    IconPhotoPlus,
    IconReorder,
    IconScan,
    IconShare,
    IconSpiral,
    IconTrash,
    IconUserMinus,
} from "@tabler/icons-react";
import {
    DeleteFiles,
    DeleteShare,
    NewWormhole,
    restoreFiles,
    TrashFiles,
    UnTrashFiles,
    UpdateFileShare,
} from "../../api/FileBrowserApi";
import { FbStateT, FileInfoT, UserContextT } from "../../types/Types";
import { useContext, useEffect, useMemo, useState } from "react";
import { userContext } from "../../Context";
import { Box, Divider, Menu, Text } from "@mantine/core";
import { RowBox } from "./FileBrowserStyles";
import { notifications } from "@mantine/notifications";
import { dispatchSync } from "../../api/Websocket";
import { AlbumScoller } from "./FileBrowserAlbums";
import { downloadSelected } from "./FileBrowserLogic";
import { ShareBox } from "./FilebrowserShareMenu";

export function FileContextMenu({
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
    if (fbState.viewingPast !== null) {
        return (
            <FileHistoryMenu
                itemId={itemId}
                fbState={fbState}
                open={open}
                setOpen={setOpen}
                menuPos={menuPos}
                dispatch={dispatch}
                wsSend={wsSend}
                authHeader={authHeader}
            />
        );
    } else {
        return (
            <StandardFileMenu
                itemId={itemId}
                fbState={fbState}
                open={open}
                setOpen={setOpen}
                menuPos={menuPos}
                dispatch={dispatch}
                wsSend={wsSend}
                authHeader={authHeader}
            />
        );
    }
}

const FileMenuHeader = ({ itemInfo, extraString }) => {
    return (
        <Box>
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
        </Box>
    );
};

function StandardFileMenu({
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
                <FileMenuHeader itemInfo={itemInfo} extraString={extraString} />

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
                            itemInfo.shares[0]?.shareId
                        );
                    }}
                >
                    <IconDownload />
                    <Text className="menu-item-text">Download</Text>
                </Box>
                {/* {historyMenu && <FileHistoryMenu fileId={itemInfo.id} />} */}
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

function FileHistoryMenu({
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
    const itemInfo: FileInfoT = fbState.dirMap.get(itemId) || ({} as FileInfoT);
    const selected: boolean = Boolean(fbState.selected.get(itemId));
    let extraString;
    if (selected && fbState.selected.size > 1) {
        extraString = ` +${fbState.selected.size - 1} more`;
    }

    return (
        <Menu opened={open} onClose={() => setOpen(false)}>
            <Menu.Target>
                <Box
                    style={{
                        position: "absolute",
                        top: menuPos.y,
                        left: menuPos.x,
                    }}
                />
            </Menu.Target>
            <Menu.Dropdown>
                <FileMenuHeader itemInfo={itemInfo} extraString={extraString} />
                <Box
                    className="menu-item"
                    onClick={(e) => {
                        e.stopPropagation();
                        restoreFiles(
                            Array.from(fbState.selected.keys()),
                            fbState.viewingPast,
                            authHeader
                        )
                            .then(() => {
                                setOpen(false);
                                dispatch({ type: "set_past_time", past: null });
                            })
                            .catch(() =>
                                notifications.show({
                                    message: "Failed to restore files",
                                    color: "red",
                                })
                            );
                    }}
                >
                    <IconReorder />
                    <Text className="menu-item-text">Bring to Present</Text>
                </Box>
            </Menu.Dropdown>
        </Menu>
    );
}

export const BackdropMenu = ({
    folderName,
    menuPos,
    menuOpen,
    setMenuOpen,
    newFolder,
}) => {
    return (
        <Menu opened={menuOpen} onClose={() => setMenuOpen(false)}>
            <Menu.Target>
                <Box
                    style={{
                        position: "fixed",
                        top: menuPos.y,
                        left: menuPos.x,
                    }}
                />
            </Menu.Target>

            <Menu.Dropdown>
                <Menu.Label>{folderName}</Menu.Label>
                <Menu.Item
                    leftSection={<IconFolderPlus />}
                    onClick={() => newFolder()}
                >
                    New Folder
                </Menu.Item>
            </Menu.Dropdown>
        </Menu>
    );
};
