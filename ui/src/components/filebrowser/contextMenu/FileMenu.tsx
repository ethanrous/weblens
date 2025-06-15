import {
    IconArrowLeft,
    IconDownload,
    IconFileExport,
    IconFolderPlus,
    IconPencil,
    IconPhotoMinus,
    IconPhotoUp,
    IconPlus,
    IconRestore,
    IconScan,
    IconTrash,
    IconUsersPlus,
} from '@tabler/icons-react'
import { FileApi, FolderApi, ScanDirectory } from '@weblens/api/FileBrowserApi'
import WeblensLoader from '@weblens/components/Loading'
import { useSessionStore } from '@weblens/components/UserInfo'
import { FileFmt } from '@weblens/components/filebrowser/filename'
import SearchDialogue from '@weblens/components/filebrowser/searchDialogue.tsx'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import WeblensInput from '@weblens/lib/WeblensInput.tsx'
import { ButtonActionHandler, ButtonIcon } from '@weblens/lib/buttonTypes'
import {
    useClick,
    useKeyDown,
    useResize,
    useWindowSize,
} from '@weblens/lib/hooks'
import useShare from '@weblens/lib/hooks/useShare'
import {
    calculateShareId,
    downloadSelected,
} from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { useMessagesController } from '@weblens/store/MessagesController'
import { ErrorHandler } from '@weblens/types/Types'
import WeblensFile, {
    FbMenuModeT,
    SelectedState,
} from '@weblens/types/files/File'
import { activeItemsFromState } from '@weblens/types/files/FileDragLogic'
import { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import { clamp, humanFileSize } from '@weblens/util'
import { ReactElement, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useShallow } from 'zustand/shallow'

import { ShareModal } from './ShareModal'

function FileMenuButton({
    Icon,
    name,
    action,
    show = true,
    danger = false,
    disabled = false,
    loading = false,
}: {
    Icon: ButtonIcon
    name: string
    action: ButtonActionHandler<void>
    show?: boolean
    danger?: boolean
    disabled?: boolean
    loading?: boolean
}) {
    if (!show) {
        return null
    }

    return (
        <div className="relative flex w-full items-center justify-center">
            {loading && (
                <WeblensLoader
                    size={14}
                    className="shadow-abyss-400 absolute z-10 m-auto shadow"
                />
            )}
            <WeblensButton
                Left={Icon}
                label={name}
                danger={danger}
                fillWidth
                centerContent={false}
                disabled={disabled || loading}
                onClick={(e) => {
                    e.stopPropagation()
                    return action()
                }}
            />
        </div>
    )
}

function MenuTitle() {
    const [targetItem, setTargetItem] = useState<WeblensFile>(null)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const selected = useFileBrowserStore((state) => state.selected)
    const menuMode = useFileBrowserStore((state) => state.menuMode)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    useEffect(() => {
        if (menuTarget === '' && targetItem?.Id() !== folderInfo?.Id()) {
            setTargetItem(folderInfo)
        } else if (menuTarget !== '') {
            const newTarget = filesMap.get(menuTarget)

            if (!newTarget || newTarget.Id() === targetItem?.Id()) {
                return
            }

            setTargetItem(newTarget)
        }
    }, [menuTarget, folderInfo, filesMap, targetItem])

    const extrasText = useMemo(() => {
        if (selected.get(targetItem?.Id()) && selected.size > 1) {
            return `+${selected.size - 1}`
        } else {
            return ''
        }
    }, [targetItem, selected])

    return (
        <div className="mb-2 flex h-max w-full max-w-96">
            {(menuMode === FbMenuModeT.NameFolder ||
                menuMode === FbMenuModeT.RenameFile) && (
                <WeblensButton
                    Left={IconArrowLeft}
                    size="small"
                    onClick={(e) => {
                        e.stopPropagation()
                        setMenu({ menuState: FbMenuModeT.Default })
                    }}
                />
            )}

            <div className="m-auto flex items-center">
                <FileFmt pathName={targetItem?.portablePath} />

                {extrasText && (
                    <span className="mt-auto mb-0.5 ml-1 select-none">
                        {extrasText}
                    </span>
                )}
            </div>
        </div>
    )
}

function MenuBodyWrapper({ children }: { children: ReactElement[] }) {
    return (
        <div
            className={
                'flex max-h-max w-full flex-col items-center justify-center gap-1'
            }
        >
            {children}
        </div>
    )
}

export function FileContextMenu() {
    const user = useSessionStore(useShallow((state) => state.user))
    const menuRef = useRef<HTMLDivElement>(null)

    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const menuPos = useFileBrowserStore((state) => state.menuPos)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const activeProps = useFileBrowserStore(
        useShallow((state) => ({
            filesMap: state.filesMap,
            selected: state.selected,
            menuTargetId: state.menuTargetId,
        }))
    )
    const activeItems = activeItemsFromState(
        activeProps.filesMap,
        activeProps.selected,
        activeProps.menuTargetId
    )

    const filesMap = useFileBrowserStore((state) => state.filesMap)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    useKeyDown(
        'Escape',
        (e) => {
            if (menuMode !== FbMenuModeT.Closed) {
                e.stopPropagation()
                setMenu({ menuState: FbMenuModeT.Closed })
            }
        },
        menuMode === FbMenuModeT.Closed
    )

    useClick((e: MouseEvent) => {
        if (menuMode !== FbMenuModeT.Closed && e.button === 0) {
            e.stopPropagation()
            setMenu({ menuState: FbMenuModeT.Closed })
        }
    }, menuRef)

    const { width, height } = useWindowSize()
    const { height: menuHeight, width: menuWidth } = useResize(menuRef)

    const menuPosStyle = useMemo(() => {
        return {
            top: clamp(
                menuPos.y,
                8 + menuHeight / 2,
                height - menuHeight / 2 - 8
            ),
            left: clamp(
                menuPos.x,
                8 + menuWidth / 2,
                width - menuWidth / 2 - 8
            ),
        }
    }, [menuPos, menuHeight, menuWidth, width, height])

    const targetFile = filesMap.get(
        menuTarget === '' ? folderInfo.Id() : menuTarget
    )

    const targetMedia = useMediaStore((state) =>
        state.mediaMap.get(targetFile?.GetContentId() ?? '')
    )

    if (!folderInfo || !targetFile) {
        return null
    }

    let menuBody: ReactElement = <div />
    if (user?.trashId === folderInfo.Id()) {
        menuBody = <InTrashMenu activeItems={activeItems.items} />
    } else if (menuMode === FbMenuModeT.Default) {
        if (pastTime.getTime() !== 0) {
            menuBody = <PastFileMenu activeItems={activeItems.items} />
        } else {
            menuBody = <StandardFileMenu activeItems={activeItems} />
        }
    } else if (menuMode === FbMenuModeT.NameFolder) {
        menuBody = <NewFolderName items={activeItems.items} />
    } else if (menuMode === FbMenuModeT.Sharing) {
        return <ShareModal targetFile={targetFile} ref={menuRef} />
    } else if (menuMode === FbMenuModeT.RenameFile) {
        menuBody = <FileRenameInput />
    } else if (menuMode === FbMenuModeT.SearchForFile) {
        const path = targetFile.portablePath ?? ''
        const text = '~' + path.slice(path.indexOf('/'), path.lastIndexOf('/'))
        menuBody = (
            <div className="menu-body-below-header flex h-[40vh] w-[50vw] items-center gap-2 p-2">
                <div className="flex h-[39vh] grow rounded-md">
                    <MediaImage
                        media={targetMedia}
                        quality={PhotoQuality.LowRes}
                    />
                </div>
                <div className="flex h-[39vh] w-[50%]">
                    <SearchDialogue
                        text={text}
                        visitFunc={(folderId: string) => {
                            FolderApi.setFolderCover(folderId, targetMedia.Id())
                                .then(() =>
                                    setMenu({ menuState: FbMenuModeT.Closed })
                                )
                                .catch((err) => {
                                    console.error(err)
                                })
                        }}
                    />
                </div>
            </div>
        )
    }

    if (menuBody === <div />) {
        console.error('Unknown menu mode:', menuMode)
        return null
    }

    return (
        <div
            key={'fileContextMenu-' + menuMode}
            className="absolute z-50 flex h-max w-max origin-center -translate-1/2 transition-[150ms] data-closed:pointer-events-none data-closed:opacity-0"
            data-closed={menuMode === FbMenuModeT.Closed ? true : null}
            data-mode={menuMode}
            style={menuPosStyle}
        >
            <div
                className="wl-floating-card flex h-max flex-col items-center justify-start"
                data-mode={menuMode}
                ref={menuRef}
                onClick={(e) => {
                    e.stopPropagation()
                }}
            >
                <MenuTitle />
                {menuBody}
            </div>
        </div>
    )
}

function StandardFileMenu({
    activeItems,
}: {
    activeItems: { items: WeblensFile[] }
}) {
    const user = useSessionStore((state) => state.user)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const menuMode = useFileBrowserStore((state) => state.menuMode)

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)

    const filesMap = useFileBrowserStore((state) => state.filesMap)

    const targetFile = useMemo(() => {
        return filesMap.get(menuTarget)
    }, [filesMap, menuTarget])

    const { share, shareLoading } = useShare(targetFile)

    if (user.trashId === folderInfo.Id()) {
        return null
    }

    if (menuMode === FbMenuModeT.Closed) {
        return null
    }

    return (
        <MenuBodyWrapper>
            <FileMenuButton
                Icon={IconPencil}
                name="Rename"
                disabled={
                    activeItems.items.length > 1 || !targetFile?.modifiable
                }
                action={() => setMenu({ menuState: FbMenuModeT.RenameFile })}
            />

            <FileMenuButton
                name="New Folder"
                Icon={IconFolderPlus}
                show={menuTarget === folderInfo.Id()}
                disabled={!folderInfo.IsModifiable()}
                action={() => setMenu({ menuState: FbMenuModeT.NameFolder })}
            />

            <FileMenuButton
                Icon={IconUsersPlus}
                name="Share"
                loading={shareLoading}
                disabled={
                    shareLoading ||
                    activeItems.items.length > 1 ||
                    (share && !share.checkPermission(user.username, 'canEdit'))
                }
                action={() => setMenu({ menuState: FbMenuModeT.Sharing })}
            />

            <FileMenuButton
                Icon={IconDownload}
                name={
                    'Download ' +
                    humanFileSize(
                        activeItems.items.reduce(
                            (acc, f) => acc + f.GetSize(),
                            0
                        )
                    ).join(' ')
                }
                disabled={
                    activeItems.items.length < 2 && targetFile?.size === 0
                }
                action={async () => {
                    const dlShareId = calculateShareId(activeItems.items)
                    return await downloadSelected(
                        activeItems.items,
                        removeLoading,
                        dlShareId
                    )
                        .then(() =>
                            useMessagesController.getState().addMessage({
                                severity: 'success',
                                text:
                                    activeItems.items.length === 1
                                        ? `Downloading ${activeItems.items[0].GetFilename()}`
                                        : `Downloading ${activeItems.items.length} files`,
                                duration: 2000,
                            })
                        )
                        .catch((e: Error) => {
                            ErrorHandler(e)
                            return false
                        })
                }}
            />

            <FileMenuButton
                Icon={IconFolderPlus}
                show={folderInfo.IsModifiable() && activeItems.items.length > 1}
                name="Folder From Selection"
                action={() => setMenu({ menuState: FbMenuModeT.NameFolder })}
            />

            <FileMenuButton
                Icon={IconPhotoMinus}
                show={
                    targetFile &&
                    targetFile.IsFolder() &&
                    targetFile.GetContentId() !== ''
                }
                name="Remove Folder Image"
                disabled={!targetFile?.modifiable}
                action={async () => {
                    if (!targetFile) {
                        return Promise.reject('No target file selected')
                    }

                    return FolderApi.setFolderCover(targetFile.Id(), '').then(
                        () => {
                            setMenu({
                                menuState: FbMenuModeT.Closed,
                            })
                            return true
                        }
                    )
                }}
            />

            <FileMenuButton
                Icon={IconPhotoUp}
                name="Set Folder Image"
                show={
                    targetFile &&
                    !targetFile.IsFolder() &&
                    activeItems.items.length === 1
                }
                disabled={targetFile?.owner !== user?.username}
                action={() =>
                    setMenu({
                        menuState: FbMenuModeT.SearchForFile,
                    })
                }
            />

            <FileMenuButton
                Icon={IconScan}
                name="Scan Directory"
                show={
                    activeItems.items.length === 1 &&
                    activeItems.items[0].IsFolder()
                }
                action={() => {
                    activeItems.items.forEach(ScanDirectory)
                    setMenu({ menuState: FbMenuModeT.Closed })
                }}
            />

            <FileMenuButton
                Icon={IconTrash}
                name="Delete"
                danger
                disabled={
                    shareLoading ||
                    !share.checkPermission(user.username, 'canDelete')
                }
                action={async () => {
                    activeItems.items.forEach((f) =>
                        f.SetSelected(SelectedState.Moved)
                    )

                    return FileApi.moveFiles(
                        {
                            fileIds: activeItems.items.map((f) => f.Id()),
                            newParentId: user.trashId,
                        },
                        share.shareId
                    )
                        .then(() => setMenu({ menuState: FbMenuModeT.Closed }))
                        .catch((err) => {
                            ErrorHandler(err, 'Error deleting files')
                            activeItems.items.forEach((f) =>
                                f.UnsetSelected(SelectedState.Moved)
                            )

                            return false
                        })
                }}
            />
        </MenuBodyWrapper>
    )
}

function PastFileMenu({ activeItems }: { activeItems: WeblensFile[] }) {
    const nav = useNavigate()

    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const folderId = useFileBrowserStore((state) => state.folderInfo.Id())
    const restoreTime = useFileBrowserStore((state) => state.pastTime)
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const setPastTime = useFileBrowserStore((state) => state.setPastTime)

    const canRestore = activeItems.find((f) => !f.hasRestoreMedia) === undefined

    return (
        <div
            className="no-scrollbar grid grid-flow-row grid-cols-2 items-center justify-center justify-items-center px-1 pt-1 pb-4"
            data-visible={menuMode === FbMenuModeT.Default}
        >
            <WeblensButton
                Left={IconRestore}
                className="mx-auto h-24 w-24"
                size="jumbo"
                disabled={!canRestore}
                tooltip={
                    canRestore
                        ? ''
                        : 'One or more selected files are missing restore media, and cannot be recovered'
                }
                onClick={async (e) => {
                    e.stopPropagation()
                    return FileApi.restoreFiles({
                        fileIds: activeItems.map((f) => f.Id()),
                        newParentId: folderId,
                        timestamp: restoreTime.getTime(),
                    }).then((res) => {
                        setMenu({ menuState: FbMenuModeT.Closed })
                        setPastTime(new Date(0))
                        nav(`/files/${res.data.newParentId}`)
                    })
                }}
            />
        </div>
    )
}

function NewFolderName({ items }: { items: WeblensFile[] }) {
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const [newName, setNewName] = useState('')

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const setMoved = useFileBrowserStore((state) => state.setSelectedMoved)

    const badName = useMemo(() => {
        if (newName.includes('/')) {
            return true
        }

        return false
    }, [newName])

    if (menuMode !== FbMenuModeT.NameFolder) {
        return <></>
    }

    return (
        <WeblensInput
            placeholder="New Folder Name"
            autoFocus
            fillWidth
            buttonIcon={IconPlus}
            valid={badName ? false : undefined}
            valueCallback={setNewName}
            onComplete={async (newName) => {
                const itemIds = items.map((f) => f.Id())
                setMoved(itemIds)
                await FolderApi.createFolder(
                    {
                        parentFolderId: folderInfo.Id(),
                        newFolderName: newName,
                        children: itemIds,
                    },
                    shareId
                )
                setMenu({ menuState: FbMenuModeT.Closed })
            }}
        />
    )
}

function FileRenameInput() {
    const menuTarget = useFileBrowserStore((state) =>
        state.filesMap.get(state.menuTargetId)
    )
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    return (
        <WeblensInput
            value={menuTarget?.GetFilename()}
            placeholder="Rename File"
            autoFocus
            fillWidth
            squareSize={50}
            buttonIcon={IconPlus}
            onComplete={async (newName) => {
                if (!menuTarget || !newName) {
                    setMenu({ menuState: FbMenuModeT.Closed })
                    return false
                }

                return FileApi.updateFile(
                    menuTarget.Id(),
                    {
                        newName: newName,
                    },
                    shareId
                ).then(() => {
                    setMenu({ menuState: FbMenuModeT.Closed })
                    return true
                })
            }}
        />
    )
}

function InTrashMenu({ activeItems }: { activeItems: WeblensFile[] }) {
    const user = useSessionStore((state) => state.user)

    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const filesList = useFileBrowserStore((state) => state.filesLists)

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const setSelectedMoved = useFileBrowserStore(
        (state) => state.setSelectedMoved
    )

    if (user.trashId !== folderInfo.Id()) {
        return <></>
    }

    return (
        <div className="flex max-h-max w-full flex-col items-center justify-center gap-1">
            <FileMenuButton
                name="Restore"
                Icon={IconFileExport}
                disabled={menuTarget === ''}
                action={async () => {
                    const ids = activeItems.map((f) => f.Id())
                    setSelectedMoved(ids)
                    setMenu({ menuState: FbMenuModeT.Closed })
                    return FileApi.unTrashFiles({
                        fileIds: ids,
                    })
                }}
            />
            <FileMenuButton
                name={menuTarget ? 'Delete' : 'Empty Trash'}
                Icon={IconTrash}
                danger
                disabled={menuTarget === '' && filesList.size === 0}
                action={async () => {
                    if (menuTarget === '') {
                        const inTrashIds = useFileBrowserStore
                            .getState()
                            .filesLists.get(user.trashId)
                            ?.map((f) => f.Id())
                        setSelectedMoved(inTrashIds ?? [])

                        await FileApi.deleteFiles(
                            {
                                fileIds: [
                                    useSessionStore.getState().user.trashId,
                                ],
                            },
                            false,
                            true
                        ).then(() => setMenu({ menuState: FbMenuModeT.Closed }))

                        return
                    }

                    const toDeleteIds = activeItems.map((f) => f.Id())
                    setSelectedMoved(toDeleteIds)
                    setMenu({ menuState: FbMenuModeT.Closed })

                    return FileApi.deleteFiles({
                        fileIds: toDeleteIds,
                    })
                }}
            />
        </div>
    )
}
