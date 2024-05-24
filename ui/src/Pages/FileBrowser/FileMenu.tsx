import {
    IconArrowBackUp,
    IconChevronRight,
    IconDownload,
    IconFile,
    IconFolder,
    IconFolderPlus,
    IconPhotoPlus,
    IconReorder,
    IconScan,
    IconShare,
    IconSpiral,
    IconTrash,
    IconUserMinus,
} from '@tabler/icons-react'
import {
    DeleteFiles,
    DeleteShare,
    NewWormhole,
    restoreFiles,
    TrashFiles,
    UnTrashFiles,
    UpdateFileShare,
} from '../../api/FileBrowserApi'
import { FbStateT, UserContextT } from '../../types/Types'
import { WeblensFile } from '../../classes/File'
import { useContext, useEffect, useMemo, useState } from 'react'
import { UserContext } from '../../Context'
import { Box, Divider, Menu, Text } from '@mantine/core'
import { RowBox } from './FileBrowserStyles'
import { notifications } from '@mantine/notifications'
import { dispatchSync } from '../../api/Websocket'
import { AlbumScoller } from './FileBrowserAlbums'
import { downloadSelected } from './FileBrowserLogic'
import { useClick, useKeyDown } from '../../components/hooks'
import { WeblensButton } from '../../components/WeblensButton'
import { ShareBox } from './FileBrowserShareMenu'
import { FbContext } from './FileBrowser'

export function FileContextMenu({
    itemId,
    setOpen,
    menuPos,
    wsSend,
}: {
    itemId: string
    setOpen
    menuPos
    wsSend
}) {
    const { authHeader } = useContext(UserContext)
    const { fbState, fbDispatch } = useContext(FbContext)

    useEffect(() => {
        fbDispatch({ type: 'set_block_focus', block: fbState.menuOpen })
    }, [fbState?.menuOpen])

    if (fbState.viewingPast !== null) {
        return (
            <FileHistoryMenu
                itemId={itemId}
                fbState={fbState}
                open={open}
                setOpen={setOpen}
                menuPos={menuPos}
                dispatch={fbDispatch}
                wsSend={wsSend}
                authHeader={authHeader}
            />
        )
    } else if (open && itemId !== '') {
        return (
            <StandardFileMenu
                itemId={itemId}
                fbState={fbState}
                open={open}
                setOpen={setOpen}
                menuPos={menuPos}
                dispatch={fbDispatch}
                wsSend={wsSend}
                authHeader={authHeader}
            />
        )
    }
}

const FileMenuHeader = ({
    itemInfo,
    extraString,
}: {
    itemInfo: WeblensFile
    extraString: string
}) => {
    return (
        <div>
            <Menu.Label>
                <RowBox style={{ gap: 8, justifyContent: 'center' }}>
                    {itemInfo.IsFolder() && <IconFolder />}
                    {!itemInfo.IsFolder() && <IconFile />}
                    <Text truncate="end" style={{ maxWidth: '250px' }}>
                        {itemInfo.GetFilename()}
                    </Text>
                    {extraString}
                </RowBox>
            </Menu.Label>
            <Divider my={10} />
        </div>
    )
}

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
    itemId: string
    fbState: FbStateT
    open
    setOpen
    menuPos
    dispatch
    wsSend
    authHeader
}) {
    const { usr }: UserContextT = useContext(UserContext)
    const [shareMenu, setShareMenu] = useState(false)
    const [addToAlbumMenu, setAddToAlbumMenu] = useState(false)
    const [itemInfo, setItemInfo] = useState(new WeblensFile({}))
    const selected: boolean = Boolean(fbState.selected.get(itemId))

    useEffect(() => {
        const info = fbState.dirMap.get(itemId)
        if (info) {
            setItemInfo(info)
        }
    }, [fbState.dirMap.get(itemId)])

    const { items }: { items: WeblensFile[] } = useMemo(() => {
        if (fbState.dirMap.size === 0) {
            return { items: [], anyDisplayable: false }
        }
        const itemIds = selected
            ? Array.from(fbState.selected.keys())
            : [itemId]
        let mediaCount = 0
        const items = itemIds.map((i) => {
            const item = fbState.dirMap.get(i)
            if (!item) {
                return null
            }
            if (item.GetMedia()?.IsDisplayable() || item.IsFolder()) {
                mediaCount++
            }
            return item
        })

        return { items: items.filter((i) => Boolean(i)), mediaCount }
    }, [
        itemId,
        JSON.stringify(fbState.dirMap.get(itemId)),
        selected,
        fbState.selected,
    ])

    let extraString
    if (selected && fbState.selected.size > 1) {
        extraString = ` +${fbState.selected.size - 1} more`
    }

    const wormholeId = useMemo(() => {
        if (itemInfo?.GetShares()) {
            const whs = itemInfo.GetShares().filter((s) => s.Wormhole)
            if (whs.length !== 0) {
                return whs[0].shareId
            }
        }
    }, [itemInfo?.GetShares()])

    const selectedMedia = useMemo(
        () =>
            items
                .filter((i) => i.GetMedia()?.IsDisplayable())
                .map((i) => i.Id()),
        [items]
    )

    const selectedFolders = useMemo(
        () => items.filter((i) => i.IsFolder()).map((i) => i.Id()),
        [items]
    )
    const inTrash = fbState.folderInfo.Id() === usr.trashId
    const inShare = fbState.folderInfo.Id() === 'shared'
    let trashName
    if (inTrash) {
        trashName = 'Delete Forever'
    } else if (inShare) {
        trashName = 'Unshare Me'
    } else {
        trashName = 'Delete'
    }

    console.log(itemInfo)

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
                    boxShadow: '0px 0px 20px -5px black',
                    width: 'max-content',
                    padding: 10,
                    border: 0,
                },
            }}
        >
            <Menu.Target>
                <Box
                    style={{
                        position: 'absolute',
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
                            boxShadow: '0px 0px 20px -5px black',
                            width: 'max-content',
                            padding: 10,
                            border: 0,
                        },
                    }}
                >
                    <Menu.Target>
                        <Box
                            className="menu-item"
                            mod={{ 'data-disabled': inTrash.toString() }}
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
                {itemInfo.IsFolder() && (
                    <Box
                        className="menu-item"
                        mod={{ 'data-disabled': inTrash.toString() }}
                        style={{
                            pointerEvents:
                                fbState.selected.size > 1 && selected
                                    ? 'none'
                                    : 'all',
                        }}
                        onClick={(e) => {
                            e.stopPropagation()
                            if (!wormholeId) {
                                NewWormhole(itemId, authHeader)
                            } else {
                                navigator.clipboard.writeText(
                                    `${window.location.origin}/wormhole/${wormholeId}`
                                )
                                setOpen(false)
                                notifications.show({
                                    message: 'Link to wormhole copied',
                                    color: 'green',
                                })
                            }
                        }}
                    >
                        <IconSpiral
                            color={fbState.selected.size > 1 ? 'grey' : 'white'}
                        />
                        <Text
                            className="menu-item-text"
                            truncate="end"
                            c={fbState.selected.size > 1 ? 'grey' : 'white'}
                        >
                            {!wormholeId ? 'Attach' : 'Copy'} Wormhole
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
                            boxShadow: '0px 0px 20px -5px black',
                            width: 'max-content',
                            padding: 0,
                            border: 0,
                        },
                    }}
                >
                    <Menu.Target>
                        <div
                            className="menu-item"
                            data-disabled={inTrash.toString()}
                        >
                            <IconShare />
                            <Text className="menu-item-text">Share</Text>
                            <IconChevronRight />
                        </div>
                    </Menu.Target>
                    <Menu.Dropdown>
                        <ShareBox candidates={items} authHeader={authHeader} />
                    </Menu.Dropdown>
                </Menu>

                <div
                    className="menu-item"
                    onClick={(e) => {
                        e.stopPropagation()
                        downloadSelected(
                            selected
                                ? Array.from(fbState.selected.keys()).map(
                                      (fId) => fbState.dirMap.get(fId)
                                  )
                                : [fbState.dirMap.get(itemId)],
                            dispatch,
                            wsSend,
                            authHeader,
                            itemInfo.GetShares()[0]?.shareId
                        )
                    }}
                >
                    <IconDownload />
                    <Text className="menu-item-text">Download</Text>
                </div>

                {itemInfo.IsFolder() && (
                    <Box
                        className="menu-item"
                        onClick={() => {
                            dispatchSync(
                                items.map((i: WeblensFile) => i.Id()),
                                wsSend,
                                true,
                                true
                            )
                            setOpen(false)
                        }}
                    >
                        <IconScan />
                        <Text className="menu-item-text">Scan</Text>
                    </Box>
                )}

                <Divider w={'100%'} my="sm" />

                {wormholeId && (
                    <Box
                        className="menu-item"
                        mod={{ 'data-disabled': inTrash.toString() }}
                        onClick={(e) => {
                            e.stopPropagation()
                            DeleteShare(wormholeId, authHeader)
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
                            e.stopPropagation()
                            UnTrashFiles(
                                items.map((i) => i.Id()),
                                authHeader
                            )
                            setOpen(false)
                        }}
                    >
                        <IconArrowBackUp />
                        <Text className="menu-item-text">{'Put back'}</Text>
                    </Box>
                )}
                <Box
                    className="menu-item"
                    onClick={(e) => {
                        e.stopPropagation()

                        if (inTrash) {
                            DeleteFiles(
                                items.map((i) => i.Id()),
                                authHeader
                            )
                        } else if (inShare) {
                            let thisShare = fbState.dirMap
                                .get(itemId)
                                .GetShares()[0]
                            UpdateFileShare(
                                thisShare.shareId,
                                thisShare.Public,
                                thisShare.Accessors.filter(
                                    (u) => u !== usr.username
                                ),
                                authHeader
                            )
                        } else {
                            TrashFiles(
                                items.map((i: WeblensFile) => i.Id()),
                                authHeader
                            )
                        }
                        setOpen(false)
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
    )
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
    itemId: string
    fbState: FbStateT
    open
    setOpen
    menuPos
    dispatch
    wsSend
    authHeader
}) {
    const itemInfo: WeblensFile =
        fbState.dirMap.get(itemId) || ({} as WeblensFile)
    const selected: boolean = Boolean(fbState.selected.get(itemId))
    let extraString
    if (selected && fbState.selected.size > 1) {
        extraString = ` +${fbState.selected.size - 1} more`
    }

    return (
        <Menu opened={open} onClose={() => setOpen(false)}>
            <Menu.Target>
                <Box
                    style={{
                        position: 'absolute',
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
                        e.stopPropagation()
                        restoreFiles(
                            Array.from(fbState.selected.keys()),
                            fbState.viewingPast,
                            authHeader
                        )
                            .then(() => {
                                setOpen(false)
                                dispatch({ type: 'set_past_time', past: null })
                            })
                            .catch(() =>
                                notifications.show({
                                    message: 'Failed to restore files',
                                    color: 'red',
                                })
                            )
                    }}
                >
                    <IconReorder />
                    <Text className="menu-item-text">Bring to Present</Text>
                </Box>
            </Menu.Dropdown>
        </Menu>
    )
}

export const BackdropMenu = ({
    folderName,
    menuPos,
    menuOpen,
    setMenuOpen,
    newFolder,
}) => {
    useKeyDown('Escape', (e) => {
        if (menuOpen) {
            e.stopPropagation()
            setMenuOpen(false)
        }
    })
    useClick(() => {
        if (menuOpen) {
            setMenuOpen(false)
        }
    })
    return (
        <Box
            key={'backdrop-menu'}
            className={`backdrop-menu backdrop-menu-${
                menuOpen ? 'open' : 'closed'
            }`}
            style={{
                top: menuPos.y,
                left: menuPos.x - 100,
            }}
        >
            <WeblensButton
                Left={
                    <IconFolderPlus style={{ width: '100%', height: '100%' }} />
                }
                subtle
                width={80}
                height={80}
                style={{ margin: 10 }}
                onClick={() => {
                    newFolder()
                    setMenuOpen(false)
                }}
            />

            {/* <Text fw={600}>New Folder</Text> */}
        </Box>
        // <Menu opened={true} onClose={() => setMenuOpen(false)}>
        //     <Menu.Target>
        //         <Box
        //             style={{
        //                 position: "fixed",
        //                 top: menuPos.y,
        //                 left: menuPos.x,
        //             }}
        //         />
        //     </Menu.Target>

        // </Menu>
    )
}
