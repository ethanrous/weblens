import {
    IconArrowLeft,
    IconDownload,
    IconFileExport,
    IconFolderPlus,
    IconLink,
    IconMinus,
    IconPencil,
    IconPhotoMinus,
    IconPhotoUp,
    IconPlus,
    IconRestore,
    IconScan,
    IconTrash,
    IconUser,
    IconUsers,
    IconUsersPlus,
} from '@tabler/icons-react'
import { FileApi, FolderApi, ScanDirectory } from '@weblens/api/FileBrowserApi'
import SharesApi from '@weblens/api/SharesApi'
import UsersApi from '@weblens/api/UserApi'
import { UserInfo } from '@weblens/api/swag'
import { useSessionStore } from '@weblens/components/UserInfo'
import { FileFmt } from '@weblens/components/filebrowser/filename'
import SearchDialogue from '@weblens/components/filebrowser/searchDialogue'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import {
    useClick,
    useKeyDown,
    useResize,
    useWindowSize,
} from '@weblens/lib/hooks'
import {
    calculateShareId,
    downloadSelected,
} from '@weblens/pages/FileBrowser/FileBrowserLogic'
import menuStyle from '@weblens/pages/FileBrowser/style/fileBrowserMenuStyle.module.scss'
import { FbModeT, useFileBrowserStore } from '@weblens/store/FBStateControl'
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
import { WeblensShare } from '@weblens/types/share/share'
import { clamp } from '@weblens/util'
import React, {
    ReactElement,
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react'
import { useNavigate } from 'react-router-dom'

type footerNote = {
    hint: string
    danger: boolean
}

const MenuTitle = () => {
    const [targetItem, setTargetItem] = useState<WeblensFile>(null)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const selected = useFileBrowserStore((state) => state.selected)
    const menuMode = useFileBrowserStore((state) => state.menuMode)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    useEffect(() => {
        if (menuTarget === '') {
            setTargetItem(folderInfo)
        } else {
            setTargetItem(filesMap.get(menuTarget))
        }
    }, [menuTarget, folderInfo])

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
                    className="absolute"
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

const MenuFooter = ({
    footerNote,
    menuMode,
}: {
    footerNote: { hint: string; danger: boolean }
    menuMode: FbMenuModeT
}) => {
    if (!footerNote.hint || menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div className="absolute bottom-0 z-100 flex h-max w-full grow translate-y-[120%] justify-center">
            <div
                className={menuStyle.footerWrapper}
                data-danger={footerNote.danger}
            >
                <p className="theme-text-dark-bg text-nowrap">
                    {footerNote.hint}
                </p>
            </div>
        </div>
    )
}

export function FileContextMenu() {
    const user = useSessionStore((state) => state.user)
    const [menuRef, setMenuRef] = useState<HTMLDivElement>(null)
    const [footerNote, setFooterNote] = useState<footerNote>({} as footerNote)

    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const menuPos = useFileBrowserStore((state) => state.menuPos)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const activeItems = useFileBrowserStore((state) =>
        activeItemsFromState(state.filesMap, state.selected, state.menuTargetId)
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

    useEffect(() => {
        if (menuMode === FbMenuModeT.Closed) {
            setFooterNote({ hint: '', danger: false })
        }
    }, [menuMode])

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

    if (!folderInfo) {
        return null
    }

    const targetFile = filesMap.get(menuTarget)
    const targetMedia = useMediaStore
        .getState()
        .mediaMap.get(targetFile?.GetContentId())

    let menuBody: ReactElement
    if (user?.trashId === folderInfo.Id()) {
        menuBody = (
            <InTrashMenu
                activeItems={activeItems.items}
                setFooterNote={setFooterNote}
            />
        )
    } else if (menuMode === FbMenuModeT.Default) {
        if (pastTime.getTime() !== 0) {
            menuBody = (
                <PastFileMenu
                    setFooterNote={setFooterNote}
                    activeItems={activeItems.items}
                />
            )
        } else if (menuTarget === '') {
            menuBody = <BackdropDefaultItems setFooterNote={setFooterNote} />
        } else {
            menuBody = (
                <StandardFileMenu
                    setFooterNote={setFooterNote}
                    activeItems={activeItems}
                />
            )
        }
    } else if (menuMode === FbMenuModeT.NameFolder) {
        menuBody = <NewFolderName items={activeItems.items} />
    } else if (menuMode === FbMenuModeT.Sharing) {
        menuBody = <FileShareMenu targetFile={targetFile} />
    } else if (menuMode === FbMenuModeT.AddToAlbum) {
        // menuBody = <AddToAlbum activeItems={activeItems.items} />
    } else if (menuMode === FbMenuModeT.RenameFile) {
        menuBody = <FileRenameInput />
    } else if (menuMode === FbMenuModeT.SearchForFile) {
        const text =
            '~' +
            targetFile.portablePath.slice(
                targetFile.portablePath.indexOf('/'),
                targetFile.portablePath.lastIndexOf('/')
            )
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

    return (
        <div
            key={'fileContextMenu'}
            className="absolute z-50 flex h-max w-max origin-center -translate-1/2 data-closed:pointer-events-none data-closed:max-h-0 data-closed:max-w-0 data-closed:opacity-0"
            data-closed={menuMode === FbMenuModeT.Closed ? true : null}
            data-mode={menuMode}
            style={menuPosStyle}
        >
            <div
                className="wl-floating-card flex h-max flex-col items-center justify-start p-2"
                data-mode={menuMode}
                ref={setMenuRef}
                onClick={(e) => {
                    e.stopPropagation()
                }}
            >
                <MenuTitle />
                {menuBody}
            </div>
            <MenuFooter footerNote={footerNote} menuMode={menuMode} />
        </div>
    )
}

function StandardFileMenu({
    setFooterNote,
    activeItems,
}: {
    setFooterNote: (n: footerNote) => void
    activeItems: { items: WeblensFile[] }
}) {
    const user = useSessionStore((state) => state.user)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const mode = useFileBrowserStore((state) => state.fbMode)

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)
    const filesMap = useFileBrowserStore((state) => state.filesMap)

    const targetFile = filesMap.get(menuTarget)

    if (user.trashId === folderInfo.Id()) {
        return null
    }

    if (menuMode === FbMenuModeT.Closed) {
        return null
    }

    return (
        <div
            className={
                'no-scrollbar grid max-h-56 grid-flow-row grid-cols-2 items-center justify-center justify-items-center gap-2 p-1 pb-4'
            }
            data-visible={menuMode === FbMenuModeT.Default && menuTarget !== ''}
        >
            <WeblensButton
                Left={IconPencil}
                disabled={activeItems.items.length > 1}
                className="mx-auto h-24 w-24"
                size="jumbo"
                onMouseOver={() =>
                    setFooterNote({ hint: 'Rename', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={(e) => {
                    e.stopPropagation()
                    setFooterNote({ hint: '', danger: false })
                    setMenu({ menuState: FbMenuModeT.RenameFile })
                }}
            />
            <WeblensButton
                Left={IconUsersPlus}
                disabled={activeItems.items.length > 1}
                className="mx-auto h-24 w-24"
                size="jumbo"
                onMouseOver={() =>
                    setFooterNote({ hint: 'Share', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={(e) => {
                    e.stopPropagation()
                    setFooterNote({ hint: '', danger: false })
                    setMenu({ menuState: FbMenuModeT.Sharing })
                }}
            />

            <WeblensButton
                Left={IconDownload}
                className="mx-auto h-24 w-24"
                size="jumbo"
                onMouseOver={() =>
                    setFooterNote({ hint: 'Download', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={async (e) => {
                    e.stopPropagation()
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
            {folderInfo.IsModifiable() && (
                <WeblensButton
                    Left={IconFolderPlus}
                    className="mx-auto h-24 w-24"
                    size="jumbo"
                    onMouseOver={() =>
                        setFooterNote({
                            hint: 'New Folder From Selection',
                            danger: false,
                        })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        setMenu({ menuState: FbMenuModeT.NameFolder })
                    }}
                />
            )}
            {targetFile &&
                (!targetFile.IsFolder() || targetFile.GetContentId()) && (
                    <>
                        {targetFile.IsFolder() &&
                            targetFile.GetContentId() !== '' && (
                                <WeblensButton
                                    Left={IconPhotoMinus}
                                    className="mx-auto h-24 w-24"
                                    size="jumbo"
                                    onMouseOver={() =>
                                        setFooterNote({
                                            hint: 'Remove Folder Image',
                                            danger: false,
                                        })
                                    }
                                    onMouseLeave={() =>
                                        setFooterNote({
                                            hint: '',
                                            danger: false,
                                        })
                                    }
                                    onClick={async (e) => {
                                        e.stopPropagation()
                                        return FolderApi.setFolderCover(
                                            targetFile.Id(),
                                            ''
                                        ).then(() => {
                                            setMenu({
                                                menuState: FbMenuModeT.Closed,
                                            })
                                            return true
                                        })
                                    }}
                                />
                            )}
                        {!targetFile.IsFolder() && (
                            <WeblensButton
                                Left={IconPhotoUp}
                                className="mx-auto h-24 w-24"
                                size="jumbo"
                                disabled={targetFile.owner !== user?.username}
                                onMouseOver={() =>
                                    setFooterNote({
                                        hint: 'Set Folder Image',
                                        danger: false,
                                    })
                                }
                                onMouseLeave={() =>
                                    setFooterNote({ hint: '', danger: false })
                                }
                                onClick={(e) => {
                                    e.stopPropagation()
                                    setMenu({
                                        menuState: FbMenuModeT.SearchForFile,
                                    })
                                }}
                            />
                        )}
                    </>
                )}

            <WeblensButton
                Left={IconScan}
                className="mx-auto h-24 w-24"
                size="jumbo"
                onMouseOver={() =>
                    setFooterNote({ hint: 'Scan Folder', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={(e) => {
                    e.stopPropagation()
                    activeItems.items.forEach(ScanDirectory)
                    setMenu({ menuState: FbMenuModeT.Closed })
                }}
            />

            <WeblensButton
                Left={IconTrash}
                danger
                className="mx-auto h-24 w-24"
                size="jumbo"
                disabled={!folderInfo.IsModifiable() || mode === FbModeT.share}
                onMouseOver={() =>
                    setFooterNote({ hint: 'Delete', danger: true })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={async (e) => {
                    e.stopPropagation()
                    activeItems.items.forEach((f) =>
                        f.SetSelected(SelectedState.Moved)
                    )
                    setMenu({ menuState: FbMenuModeT.Closed })
                    return FileApi.moveFiles({
                        fileIds: activeItems.items.map((f) => f.Id()),
                        newParentId: user.trashId,
                    })
                }}
            />
        </div>
    )
}

function PastFileMenu({
    setFooterNote,
    activeItems,
}: {
    setFooterNote: (n: footerNote) => void
    activeItems: WeblensFile[]
}) {
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
                onMouseOver={() =>
                    setFooterNote({ hint: 'Restore', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={async (e) => {
                    e.stopPropagation()
                    return FileApi.restoreFiles({
                        fileIds: activeItems.map((f) => f.Id()),
                        newParentId: folderId,
                        timestamp: restoreTime.getTime(),
                    }).then((res) => {
                        setFooterNote({ hint: '', danger: false })
                        setMenu({ menuState: FbMenuModeT.Closed })
                        setPastTime(new Date(0))
                        nav(`/files/${res.data.newParentId}`)
                    })
                }}
            />
        </div>
    )
}

function FileShareMenu({ targetFile }: { targetFile: WeblensFile }) {
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const searchRef = useRef<HTMLInputElement>()

    const [accessors, setAccessors] = useState<UserInfo[]>([])
    const [isPublic, setIsPublic] = useState(false)
    const [share, setShare] = useState<WeblensShare>(null)

    useClick(() => {
        setSearchMenuOpen(false)
    }, searchRef.current)

    if (!targetFile) {
        targetFile = folderInfo
    }

    useEffect(() => {
        if (!targetFile) {
            return
        }
        const setShareData = async () => {
            const share = await targetFile.GetShare()
            if (share) {
                setShare(share)
                if (share.IsPublic() !== undefined) {
                    setIsPublic(share.IsPublic())
                }
                setAccessors(share.GetAccessors())
            } else {
                setIsPublic(false)
            }
        }
        setShareData().catch((err) => {
            console.error('Failed to set share data', err)
        })
    }, [targetFile])

    const [userSearch, setUserSearch] = useState('')
    const [searchMenuOpen, setSearchMenuOpen] = useState(false)
    const [userSearchResults, setUserSearchResults] = useState<UserInfo[]>([])
    useEffect(() => {
        if (userSearch.length < 2) {
            setUserSearchResults([])
            return
        }
        UsersApi.searchUsers(userSearch)
            .then((res) => {
                if (!res.data) {
                    setUserSearchResults([])
                    return
                }
                const searchResults = res.data.filter(
                    (u) =>
                        accessors.findIndex(
                            (val) => val.username === u.username
                        ) === -1
                )

                setUserSearchResults(searchResults)
            })
            .catch((err) => {
                console.error('Failed to search users', err)
            })
    }, [userSearch])

    useEffect(() => {
        if (menuMode !== FbMenuModeT.Sharing) {
            setUserSearch('')
            setUserSearchResults([])
        }
    }, [menuMode])

    const updateShare = useCallback(
        async (e: React.MouseEvent<HTMLElement>) => {
            e.stopPropagation()
            const share = await targetFile.GetShare()
            if (share?.Id()) {
                return await share
                    .UpdateShare(
                        isPublic,
                        accessors.map((u) => u.username)
                    )
                    .then(() => share)
                    .catch(ErrorHandler)
            } else {
                return await SharesApi.createFileShare({
                    fileId: targetFile.Id(),
                    public: isPublic,
                    users: accessors.map((u) => u.username),
                })
                    .then(async (res) => {
                        targetFile.SetShare(new WeblensShare(res.data))
                        const sh = await targetFile.GetShare()
                        return sh
                    })
                    .catch(ErrorHandler)
            }
        },
        [targetFile, isPublic, accessors, folderInfo]
    )

    if (menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className="flex h-max w-72 flex-col items-center"
            data-visible={menuMode === FbMenuModeT.Sharing}
            onClick={(e) => e.stopPropagation()}
        >
            <div className="flex w-full flex-row gap-1">
                <WeblensButton
                    label={isPublic ? 'Public' : 'Private'}
                    Left={isPublic ? IconUsers : IconUser}
                    flavor={isPublic ? 'default' : 'outline'}
                    fillWidth
                    onClick={(e) => {
                        e.stopPropagation()
                        setIsPublic((p) => !p)
                    }}
                />
                <WeblensButton
                    label={'Copy Link'}
                    Left={IconLink}
                    fillWidth
                    disabled={!isPublic && accessors.length === 0}
                    onClick={async (e) => {
                        e.stopPropagation()
                        return await updateShare(e)
                            .then(async (share) => {
                                if (!share) {
                                    console.error('No Shares!')
                                    return false
                                }
                                return navigator.clipboard
                                    .writeText(share.GetPublicLink())
                                    .then(() => {
                                        useMessagesController
                                            .getState()
                                            .addMessage({
                                                severity: 'success',
                                                text: 'Share link copied!',
                                                duration: 2000,
                                            })
                                    })
                                    .catch((r) => {
                                        useMessagesController
                                            .getState()
                                            .addMessage({
                                                severity: 'error',
                                                text: 'Failed to copy share link',
                                                duration: 5000,
                                            })
                                        console.error(r)
                                        return false
                                    })
                            })
                            .catch((r) => {
                                console.error(r)
                                return false
                            })
                    }}
                />
            </div>
            <div
                ref={searchRef}
                className="flex w-full flex-col items-center gap-1"
            >
                <div className="z-20 mt-3 mb-3 h-10 w-full">
                    <WeblensInput
                        value={userSearch}
                        valueCallback={setUserSearch}
                        placeholder="Add users"
                        onComplete={null}
                        Icon={IconUsersPlus}
                        openInput={() => setSearchMenuOpen(true)}
                    />
                </div>
                {userSearchResults.length !== 0 && searchMenuOpen && (
                    <div className="no-scrollbar bg-background-secondary absolute z-10 mt-14 flex max-h-32 w-11/12 flex-col gap-1 rounded-sm p-2 drop-shadow-xl">
                        {userSearchResults.map((u) => {
                            return (
                                <div
                                    className="hover:bg-background-tertiary flex h-10 w-full cursor-pointer flex-row items-center rounded p-4"
                                    key={u.username}
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        e.preventDefault()
                                        setAccessors((p) => {
                                            const newP = [...p]
                                            newP.push(u)
                                            return newP
                                        })
                                        setUserSearchResults((p) => {
                                            const newP = [...p]
                                            newP.splice(newP.indexOf(u), 1)
                                            return newP
                                        })
                                    }}
                                >
                                    <span>{u.fullName}</span>
                                    <span className="text-text-secondary ml-1">
                                        [{u.username}]
                                    </span>
                                    <IconPlus className="ml-auto" />
                                </div>
                            )
                        })}
                    </div>
                )}
                <span className="m-2">Shared With</span>
            </div>
            <div className="no-scrollbar m-2 mt-0 flex h-full max-h-60 min-h-20 w-full rounded-sm border p-2">
                {accessors.length === 0 && (
                    <span className="m-auto">Not Shared</span>
                )}
                {accessors.length !== 0 &&
                    accessors.map((u: UserInfo) => {
                        return (
                            <div
                                key={u.username}
                                className="bg-background-secondary hover:bg-background-tertiary group/user flex h-10 w-full items-center rounded p-2 transition"
                            >
                                <span>{u.fullName}</span>
                                <span className="text-color-text-secondary ml-1">
                                    [{u.username}]
                                </span>
                                <div
                                    className={
                                        'ml-auto opacity-0 group-hover/user:opacity-100'
                                    }
                                >
                                    <WeblensButton
                                        size="small"
                                        Left={IconMinus}
                                        onClick={(e) => {
                                            e.stopPropagation()
                                            setAccessors((p) => {
                                                const newP = [...p]
                                                newP.splice(
                                                    newP.findIndex(
                                                        (v) =>
                                                            v.username ===
                                                            u.username
                                                    ),
                                                    1
                                                )
                                                return newP
                                            })
                                        }}
                                    />
                                </div>
                            </div>
                        )
                    })}
            </div>
            <div className="flex w-full flex-row">
                <div className="m-1 flex w-1/4 grow justify-center">
                    <WeblensButton
                        label="Back"
                        fillWidth
                        Left={IconArrowLeft}
                        onClick={(e) => {
                            e.stopPropagation()
                            setMenu({ menuState: FbMenuModeT.Default })
                        }}
                    />
                </div>
                <div className="m-1 flex h-full w-1/4 grow justify-center">
                    <WeblensButton
                        fillWidth
                        label="Save"
                        disabled={
                            share &&
                            share.IsPublic() === isPublic &&
                            accessors === share.GetAccessors()
                        }
                        onClick={(e) =>
                            updateShare(e)
                                .then(() => true)
                                .catch(ErrorHandler)
                        }
                    />
                </div>
            </div>
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
            value={menuTarget.GetFilename()}
            placeholder="Rename File"
            autoFocus
            fillWidth
            squareSize={50}
            buttonIcon={IconPlus}
            onComplete={async (newName) => {
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

// function AlbumCover({
//     a,
//     medias,
//     refetch,
// }: {
//     a: AlbumInfo
//     medias: string[]
//     refetch: () => Promise<QueryObserverResult<MediaInfo[], Error>>
// }) {
//     const hasAll = medias?.filter((v) => !a.medias?.includes(v)).length === 0
//
//     return (
//         <div
//             className="h-max w-max"
//             key={a.id}
//             onClick={(e) => {
//                 e.stopPropagation()
//                 if (hasAll) {
//                     return
//                 }
//                 AlbumsApi.updateAlbum(a.id, undefined, undefined, medias)
//                     .then(() => refetch())
//                     .catch(ErrorHandler)
//             }}
//         >
//             <MiniAlbumCover
//                 album={a}
//                 disabled={!medias || medias.length === 0 || hasAll}
//             />
//         </div>
//     )
// }

// function AddToAlbum({ activeItems }: { activeItems: WeblensFile[] }) {
//     const [newAlbum, setNewAlbum] = useState(false)
//
//     const { data: albums } = useQuery<AlbumInfo[]>({
//         queryKey: ['albums'],
//         initialData: [],
//         queryFn: () =>
//             AlbumsApi.getAlbums().then((res) =>
//                 res.data.sort((a, b) => {
//                     return a.name.localeCompare(b.name)
//                 })
//             ),
//     })
//
//     const menuMode = useFileBrowserStore((state) => state.menuMode)
//     const setMenu = useFileBrowserStore((state) => state.setMenu)
//     const addMedias = useMediaStore((state) => state.addMedias)
//     const getMedia = useMediaStore((state) => state.getMedia)
//
//     useEffect(() => {
//         setNewAlbum(false)
//     }, [menuMode])
//
//     useEffect(() => {
//         const newMediaIds: string[] = []
//         for (const album of albums) {
//             if (album.cover && !getMedia(album.cover)) {
//                 newMediaIds.push(album.cover)
//             }
//         }
//         if (newMediaIds) {
//             MediaApi.getMedia(
//                 true,
//                 true,
//                 undefined,
//                 undefined,
//                 undefined,
//                 undefined,
//                 JSON.stringify(newMediaIds)
//             )
//                 .then((res) => {
//                     const medias = res.data.Media.map(
//                         (mediaParam) => new WeblensMedia(mediaParam)
//                     )
//                     addMedias(medias)
//                 })
//                 .catch((err) => {
//                     console.error(err)
//                 })
//         }
//     }, [albums.length])
//
//     const {
//         data: medias,
//         isLoading,
//         refetch,
//     } = useQuery<MediaInfo[]>({
//         queryKey: ['selected-medias', activeItems.map((i) => i.Id()), menuMode],
//         initialData: [],
//         queryFn: () => {
//             if (menuMode !== FbMenuModeT.AddToAlbum) {
//                 return [] as MediaInfo[]
//             }
//             return MediaApi.getMedia(
//                 true,
//                 true,
//                 undefined,
//                 undefined,
//                 undefined,
//                 JSON.stringify(activeItems.map((i) => i.Id()))
//             ).then((res) => res.data.Media)
//         },
//     })
//
//     if (menuMode !== FbMenuModeT.AddToAlbum) {
//         return <></>
//     }
//
//     return (
//         <div className="add-to-album-menu">
//             {medias && medias.length !== 0 && (
//                 <p className="animate-fade">
//                     Add {medias.length} media to Albums
//                 </p>
//             )}
//             {medias && medias.length === 0 && (
//                 <p className="animate-fade">No valid media selected</p>
//             )}
//             {isLoading && <p className="animate-fade">Loading media...</p>}
//             <div className="no-scrollbar grid grid-cols-2 gap-3 h-max max-h-[350px] overflow-y-scroll pt-1">
//                 {albums.map((a) => {
//                     return (
//                         <AlbumCover
//                             key={a.name}
//                             a={a}
//                             medias={medias.map((m) => m.contentId)}
//                             refetch={refetch}
//                         />
//                     )
//                 })}
//             </div>
//             {newAlbum && (
//                 <WeblensInput
//                     squareSize={40}
//                     autoFocus
//                     closeInput={() => setNewAlbum(false)}
//                     onComplete={async (v: string) =>
//                         AlbumsApi.createAlbum(v)
//                             .then(() => refetch())
//                             .then(() => {
//                                 setNewAlbum(false)
//                             })
//                     }
//                 />
//             )}
//             {!newAlbum && (
//                 <WeblensButton
//                     fillWidth
//                     label={'New Album'}
//                     Left={IconLibraryPlus}
//                     size="jumbo"
//                     onClick={(e) => {
//                         e.stopPropagation()
//                         setNewAlbum(true)
//                     }}
//                 />
//             )}
//             <WeblensButton
//                 fillWidth
//                 label={'Back'}
//                 Left={IconArrowLeft}
//                 size="jumbo"
//                 onClick={(e) => {
//                     e.stopPropagation()
//                     setMenu({ menuState: FbMenuModeT.Default })
//                 }}
//             />
//         </div>
//     )
// }

function InTrashMenu({
    activeItems,
    setFooterNote,
}: {
    activeItems: WeblensFile[]
    setFooterNote: (n: footerNote) => void
}) {
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
        <div className="no-scrollbar grid grid-flow-row grid-cols-2 items-center justify-center justify-items-center gap-2 p-1 pb-4">
            <WeblensButton
                className="mx-auto h-24 w-24"
                Left={IconFileExport}
                size="jumbo"
                disabled={menuTarget === ''}
                onMouseOver={() =>
                    setFooterNote({ hint: 'Put Back', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={async (e) => {
                    e.stopPropagation()
                    const ids = activeItems.map((f) => f.Id())
                    setSelectedMoved(ids)
                    setMenu({ menuState: FbMenuModeT.Closed })
                    return FileApi.unTrashFiles({
                        fileIds: ids,
                    })
                }}
            />
            <WeblensButton
                className="mx-auto h-24 w-24"
                Left={IconTrash}
                size="jumbo"
                danger
                disabled={menuTarget === '' && filesList.size === 0}
                onMouseOver={() =>
                    setFooterNote({
                        hint:
                            menuTarget === ''
                                ? 'Empty Trash'
                                : 'Delete Permanently',
                        danger: true,
                    })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={async (e) => {
                    e.stopPropagation()
                    let toDeleteIds: string[]
                    if (menuTarget === '') {
                        toDeleteIds = filesList
                            .get(user.trashId)
                            .map((f) => f.Id())
                    } else {
                        toDeleteIds = activeItems.map((f) => f.Id())
                    }
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

function BackdropDefaultItems({
    setFooterNote,
}: {
    setFooterNote: (n: footerNote) => void
}) {
    const user = useSessionStore((state) => state.user)

    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    if (menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className="no-scrollbar grid grid-flow-row grid-cols-2 items-center justify-center justify-items-center gap-2 p-1 pb-4"
            data-visible={menuMode === FbMenuModeT.Default && menuTarget === ''}
        >
            <WeblensButton
                Left={IconFolderPlus}
                className="mx-auto h-24 w-24"
                size="jumbo"
                disabled={!folderInfo.IsModifiable()}
                onMouseOver={() =>
                    setFooterNote({ hint: 'New Folder', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={(e) => {
                    e.stopPropagation()
                    setFooterNote({ hint: '', danger: false })
                    setMenu({ menuState: FbMenuModeT.NameFolder })
                }}
            />

            <WeblensButton
                Left={IconUsersPlus}
                className="mx-auto h-24 w-24"
                disabled={
                    folderInfo.Id() === user.homeId ||
                    folderInfo.Id() === user.trashId
                }
                size="jumbo"
                onMouseOver={() =>
                    setFooterNote({ hint: 'Share', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={(e) => {
                    e.stopPropagation()
                    setMenu({ menuState: FbMenuModeT.Sharing })
                    setFooterNote({ hint: '', danger: false })
                }}
            />

            <WeblensButton
                Left={IconScan}
                className="mx-auto h-24 w-24"
                size="jumbo"
                onMouseOver={() =>
                    setFooterNote({ hint: 'Scan Folder', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={(e) => {
                    e.stopPropagation()
                    ScanDirectory(folderInfo)
                }}
            />
        </div>
    )
}
