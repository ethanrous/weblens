import {
    IconArrowLeft,
    IconDownload,
    IconFileAnalytics,
    IconFileExport,
    IconFolderPlus,
    IconLibraryPlus,
    IconLink,
    IconMinus,
    IconPencil,
    IconPhotoMinus,
    IconPhotoShare,
    IconPhotoUp,
    IconPlus,
    IconRestore,
    IconScan,
    IconTrash,
    IconUser,
    IconUsers,
    IconUsersPlus,
} from '@tabler/icons-react'
import { useQuery, UseQueryResult } from '@tanstack/react-query'
import { AutocompleteUsers } from '@weblens/api/ApiFetch'

import '@weblens/pages/FileBrowser/style/fileBrowserMenuStyle.scss'

import {
    CreateFolder,
    DeleteFiles,
    RenameFile,
    SetFolderImage,
    TrashFiles,
    UnTrashFiles,
} from '@weblens/api/FileBrowserApi'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import {
    FbModeT,
    useFileBrowserStore,
} from '@weblens/pages/FileBrowser/FBStateControl'
import { downloadSelected } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { MiniAlbumCover } from '@weblens/types/albums/AlbumDisplay'
import {
    addMediaToAlbum,
    createAlbum,
    getAlbums,
} from '@weblens/types/albums/AlbumQuery'
import {
    FbMenuModeT,
    SelectedState,
    WeblensFile,
} from '@weblens/types/files/File'
import { getFoldersMedia, restoreFiles } from '@weblens/types/files/FilesQuery'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { getMedias } from '@weblens/types/media/MediaQuery'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { WeblensShare } from '@weblens/types/share/share'
import { shareFile } from '@weblens/types/share/shareQuery'
import {
    useClick,
    useKeyDown,
    useResize,
    useWindowSize,
} from 'components/hooks'
import { useSessionStore } from 'components/UserInfo'
import React, {
    ReactElement,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react'
import { useNavigate } from 'react-router-dom'
import { AlbumData, UserInfoT } from 'types/Types'
import { clamp } from '@weblens/util'
import { FileFmt } from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
import SearchDialogue from '@weblens/pages/FileBrowser/SearchDialogue'
import { MediaImage } from '../media/PhotoContainer'
import { useWebsocketStore } from '@weblens/api/Websocket'

type footerNote = {
    hint: string
    danger: boolean
}

const activeItemsFromState = (
    filesMap: Map<string, WeblensFile>,
    selected: Map<string, boolean>,
    menuTargetId: string
): {
    items: WeblensFile[]
    anyDisplayable: boolean
    mediaCount: number
} => {
    if (filesMap.size === 0) {
        return { items: [], anyDisplayable: false, mediaCount: 0 }
    }
    const isSelected = Boolean(selected.get(menuTargetId))
    const itemIds = isSelected ? Array.from(selected.keys()) : [menuTargetId]
    let mediaCount = 0
    const items = itemIds.map((i) => {
        const item = filesMap.get(i)
        if (!item) {
            return null
        }
        if (item.GetContentId() || item.IsFolder()) {
            mediaCount++
        }
        return item
    })

    return {
        items: items.filter((i) => Boolean(i)),
        anyDisplayable: undefined,
        mediaCount,
    }
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
        <div className="file-menu-title">
            {menuMode === FbMenuModeT.NameFolder && (
                <div className="flex flex-grow absolute w-full">
                    <WeblensButton
                        Left={IconArrowLeft}
                        onClick={(e) => {
                            e.stopPropagation()
                            setMenu({ menuState: FbMenuModeT.Default })
                        }}
                    />
                </div>
            )}

            <div className="flex flex-row items-center justify-center w-full h-8 gap-1">
                <FileFmt pathName={targetItem?.portablePath} />

                {extrasText && (
                    <p className="flex w-max items-center justify-end text-xs select-none h-3">
                        {extrasText}
                    </p>
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
        <div className="flex absolute flex-grow w-full justify-center h-max bottom-0 z-[100] translate-y-[120%]">
            <div
                className="footer-wrapper animate-fade"
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
    const viewingPast = useFileBrowserStore((state) => state.pastTime)
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

    useClick((e) => {
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
        if (viewingPast) {
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
        menuBody = <AddToAlbum activeItems={activeItems.items} />
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
            <div className="flex w-[50vw] h-[40vh] p-2 gap-2 menu-body-below-header items-center">
                <div className="flex grow rounded-md h-[39vh]">
                    <MediaImage
                        media={targetMedia}
                        quality={PhotoQuality.LowRes}
                    />
                </div>
                <div className="flex w-[50%] h-[39vh]">
                    <SearchDialogue
                        text={text}
                        visitFunc={(folderId: string) => {
                            SetFolderImage(folderId, targetMedia.Id()).then(
                                () => setMenu({ menuState: FbMenuModeT.Closed })
                            )
                        }}
                    />
                </div>
            </div>
        )
    }

    return (
        <div
            className="backdrop-menu-wrapper"
            data-mode={menuMode}
            style={menuPosStyle}
        >
            <div
                className={'backdrop-menu'}
                data-mode={menuMode}
                ref={setMenuRef}
                onClick={(e) => {
                    e.stopPropagation()
                    setMenu({ menuState: FbMenuModeT.Closed })
                }}
            >
                <MenuTitle />
                {/* {viewingPast !== null && <div />} */}
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
    const wsSend = useWebsocketStore((state) => state.wsSend)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)
    const filesMap = useFileBrowserStore((state) => state.filesMap)

    const targetFile = filesMap.get(menuTarget)

    if (user.trashId === folderInfo.Id()) {
        return <></>
    }

    if (menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className={'default-grid no-scrollbar'}
            data-visible={menuMode === FbMenuModeT.Default && menuTarget !== ''}
        >
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconPencil}
                    disabled={activeItems.items.length > 1}
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Rename', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        setFooterNote({ hint: '', danger: false })
                        setMenu({ menuState: FbMenuModeT.RenameFile })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconUsersPlus}
                    disabled={activeItems.items.length > 1}
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Share', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        setFooterNote({ hint: '', danger: false })
                        setMenu({ menuState: FbMenuModeT.Sharing })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconDownload}
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Download', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={async (e) => {
                        e.stopPropagation()
                        return await downloadSelected(
                            activeItems.items,
                            removeLoading,
                            wsSend,
                            shareId
                        )
                            .then(() => true)
                            .catch(() => false)
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconPhotoShare}
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({
                            hint: 'Add Medias to Album',
                            danger: false,
                        })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        setFooterNote({ hint: '', danger: false })
                        setMenu({ menuState: FbMenuModeT.AddToAlbum })
                    }}
                />
            </div>
            {folderInfo.IsModifiable() && (
                <div className="default-menu-icon">
                    <WeblensButton
                        Left={IconFolderPlus}
                        squareSize={100}
                        centerContent
                        onMouseOver={() =>
                            setFooterNote({
                                hint: `New Folder From Selection (${activeItems.items.length})`,
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
                </div>
            )}
            {targetFile &&
                (!targetFile.IsFolder() || targetFile.GetContentId()) && (
                    <div className="default-menu-icon">
                        {targetFile.IsFolder() &&
                            targetFile.GetContentId() !== '' && (
                                <WeblensButton
                                    Left={IconPhotoMinus}
                                    squareSize={100}
                                    centerContent
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
                                        SetFolderImage(
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
                                squareSize={100}
                                centerContent
                                onMouseOver={() =>
                                    setFooterNote({
                                        hint: 'Set Folder Image',
                                        danger: false,
                                    })
                                }
                                onMouseLeave={() =>
                                    setFooterNote({ hint: '', danger: false })
                                }
                                onClick={async (e) => {
                                    e.stopPropagation()
                                    setMenu({
                                        menuState: FbMenuModeT.SearchForFile,
                                    })
                                }}
                            />
                        )}
                    </div>
                )}
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconScan}
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Scan Folder', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        activeItems.items.forEach((i) =>
                            wsSend('scan_directory', { folderId: i.Id() })
                        )
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconTrash}
                    danger
                    squareSize={100}
                    centerContent
                    disabled={!folderInfo.IsModifiable()}
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Delete', danger: true })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={async (e) => {
                        e.stopPropagation()
                        activeItems.items.forEach((f) =>
                            f.SetSelected(SelectedState.Moved)
                        )
                        return TrashFiles(
                            activeItems.items.map((f) => f.Id()),
                            shareId
                        ).then(() => {
                            setMenu({ menuState: FbMenuModeT.Closed })
                        })
                    }}
                />
            </div>
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
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const folderId = useFileBrowserStore((state) => state.folderInfo.Id())
    const restoreTime = useFileBrowserStore((state) => state.pastTime)

    return (
        <div
            className={'default-grid no-scrollbar'}
            data-visible={menuMode === FbMenuModeT.Default}
        >
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconRestore}
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Restore', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={async (e) => {
                        e.stopPropagation()
                        restoreFiles(
                            activeItems.map((f) => f.Id()),
                            folderId,
                            restoreTime
                        ).then((res) => {
                            setFooterNote({ hint: '', danger: false })
                            setMenu({ menuState: FbMenuModeT.Closed })
                            console.log('going to', res.newParentId)
                            nav(`/files/${res.newParentId}`)
                        })
                    }}
                />
            </div>
        </div>
    )
}

function FileShareMenu({ targetFile }: { targetFile: WeblensFile }) {
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const setMenu = useFileBrowserStore((state) => state.setMenu)

    const [accessors, setAccessors] = useState<string[]>([])
    const [isPublic, setIsPublic] = useState(false)
    const [share, setShare] = useState<WeblensShare>(null)

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
        setShareData()
    }, [targetFile])

    const [userSearch, setUserSearch] = useState('')
    const [userSearchResults, setUserSearchResults] = useState<UserInfoT[]>([])
    useEffect(() => {
        AutocompleteUsers(userSearch).then((us) => {
            us = us.filter((u) => !accessors.includes(u.username))
            setUserSearchResults(us)
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
            if (share) {
                return await share
                    .UpdateShare(isPublic, accessors)
                    .then(() => true)
                    .catch(() => false)
            } else {
                return await shareFile(
                    targetFile,
                    isPublic,
                    accessors.map((u) => u)
                )
                    .then((si) => {
                        targetFile.SetShare(new WeblensShare(si))
                        return true
                    })
                    .catch(() => false)
            }
        },
        [targetFile, isPublic, accessors]
    )

    if (menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className="file-share-menu"
            data-visible={menuMode === FbMenuModeT.Sharing}
        >
            <div className="flex flex-row w-full">
                <div className="flex justify-center w-1/4 m-1 grow">
                    <WeblensButton
                        squareSize={40}
                        label={isPublic ? 'Public' : 'Private'}
                        allowRepeat
                        fillWidth
                        centerContent
                        toggleOn={isPublic}
                        Left={isPublic ? IconUsers : IconUser}
                        onClick={(e) => {
                            e.stopPropagation()
                            setIsPublic((p) => !p)
                        }}
                    />
                </div>
                <div className="flex justify-center w-1/4 m-1 grow">
                    <WeblensButton
                        squareSize={40}
                        label={'Copy Link'}
                        fillWidth
                        centerContent
                        Left={IconLink}
                        disabled={!isPublic && accessors.length === 0}
                        onClick={async (e) => {
                            e.stopPropagation()
                            return await updateShare(e)
                                .then(async () => {
                                    const share = await targetFile.GetShare()
                                    if (!share) {
                                        console.error('No Shares!')
                                        return false
                                    }
                                    return navigator.clipboard
                                        .writeText(share.GetPublicLink())
                                        .then(() => true)
                                        .catch((r) => {
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
            </div>
            <div className="flex flex-col w-full gap-1 items-center">
                <div className="h-10 w-full mt-3 mb-3 z-20">
                    <WeblensInput
                        value={userSearch}
                        valueCallback={setUserSearch}
                        placeholder="Add users"
                        onComplete={null}
                        Icon={IconUsersPlus}
                    />
                </div>
                {userSearchResults.length !== 0 && (
                    <div
                        className="flex flex-col w-full bg-raised-grey absolute gap-1 rounded
                                    p-1 z-10 mt-14 max-h-32 overflow-y-scroll drop-shadow-xl"
                    >
                        {userSearchResults.map((u) => {
                            return (
                                <div
                                    className="user-autocomplete-row"
                                    key={u.username}
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        setAccessors((p) => {
                                            const newP = [...p]
                                            newP.push(u.username)
                                            return newP
                                        })
                                        setUserSearchResults((p) => {
                                            const newP = [...p]
                                            newP.splice(newP.indexOf(u), 1)
                                            return newP
                                        })
                                    }}
                                >
                                    <p>{u.username}</p>
                                    <IconPlus />
                                </div>
                            )
                        })}
                    </div>
                )}
                <p className="text-white">Shared With</p>
            </div>
            <div
                className="flex flex-row w-full h-full p-2 m-2 mt-0 rounded
                            outline outline-main-accent justify-center"
            >
                {accessors.length === 0 && <p>Not Shared</p>}
                {accessors.length !== 0 &&
                    accessors.map((u: string) => {
                        return (
                            <div key={u} className="user-autocomplete-row">
                                <p>{u}</p>
                                <div className="user-minus-button">
                                    <WeblensButton
                                        squareSize={40}
                                        Left={IconMinus}
                                        onClick={(e) => {
                                            e.stopPropagation()
                                            setAccessors((p) => {
                                                const newP = [...p]
                                                newP.splice(newP.indexOf(u), 1)
                                                return newP
                                            })
                                        }}
                                    />
                                </div>
                            </div>
                        )
                    })}
            </div>
            <div className="flex flex-row w-full">
                <div className="flex justify-center w-1/4 m-1 grow">
                    <WeblensButton
                        squareSize={40}
                        centerContent
                        label="Back"
                        fillWidth
                        Left={IconArrowLeft}
                        onClick={(e) => {
                            e.stopPropagation()
                            setMenu({ menuState: FbMenuModeT.Default })
                        }}
                    />
                </div>
                <div className="flex justify-center w-1/4 m-1 grow">
                    <WeblensButton
                        squareSize={40}
                        centerContent
                        fillWidth
                        label="Save"
                        disabled={
                            share &&
                            share.IsPublic() === isPublic &&
                            accessors === share.GetAccessors()
                        }
                        onClick={updateShare}
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
        <div className="new-folder-menu">
            <WeblensInput
                placeholder="New Folder Name"
                autoFocus
                fillWidth
                squareSize={50}
                buttonIcon={IconPlus}
                failed={badName}
                valueCallback={setNewName}
                onComplete={async (newName) => {
                    const itemIds = items.map((f) => f.Id())
                    setMoved(itemIds)
                    return await CreateFolder(
                        folderInfo.Id(),
                        newName,
                        itemIds,
                        false,
                        shareId
                    )
                        .then(() => setMenu({ menuState: FbMenuModeT.Closed }))
                        .catch((r) => console.error(r))
                }}
            />
            <div className="w-[220px]"></div>
        </div>
    )
}

function FileRenameInput() {
    const menuTarget = useFileBrowserStore((state) =>
        state.filesMap.get(state.menuTargetId)
    )

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    return (
        <div className="new-folder-menu">
            <WeblensInput
                value={menuTarget.GetFilename()}
                placeholder="Rename File"
                autoFocus
                fillWidth
                squareSize={50}
                buttonIcon={IconPlus}
                onComplete={async (newName) => {
                    await RenameFile(menuTarget.Id(), newName)
                        .then(() => setMenu({ menuState: FbMenuModeT.Closed }))
                        .catch((r) => console.error(r))
                }}
            />
            <div className="w-[220px]"></div>
        </div>
    )
}

function AlbumCover({
    a,
    medias,
    albums,
}: {
    a: AlbumData
    medias: string[]
    albums: UseQueryResult<AlbumData[], Error>
}) {
    const hasAll = medias?.filter((v) => !a.medias?.includes(v)).length === 0

    return (
        <div
            className="h-max w-max"
            key={a.id}
            onClick={(e) => {
                e.stopPropagation()
                if (hasAll) {
                    return
                }
                addMediaToAlbum(a.id, medias, []).then(() => albums.refetch())
            }}
        >
            <MiniAlbumCover
                album={a}
                disabled={!medias || medias.length === 0 || hasAll}
            />
        </div>
    )
}

function AddToAlbum({ activeItems }: { activeItems: WeblensFile[] }) {
    const [newAlbum, setNewAlbum] = useState(false)

    const albums = useQuery<AlbumData[]>({
        queryKey: ['albums'],
        queryFn: () =>
            getAlbums(false).then((as) =>
                as.sort((a, b) => {
                    return a.name.localeCompare(b.name)
                })
            ),
        initialData: [],
    })

    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const addMedias = useMediaStore((state) => state.addMedias)
    const getMedia = useMediaStore((state) => state.getMedia)

    useEffect(() => {
        setNewAlbum(false)
    }, [menuMode])

    useEffect(() => {
        const newMediaIds: string[] = []
        for (const album of albums.data) {
            if (album.cover && !getMedia(album.cover)) {
                newMediaIds.push(album.cover)
            }
        }
        if (newMediaIds) {
            getMedias(newMediaIds).then((mediaParams) => {
                const medias = mediaParams.map(
                    (mediaParam) => new WeblensMedia(mediaParam)
                )
                addMedias(medias)
            })
        }
    }, [albums?.data.length])

    const getMediasInFolders = useCallback(
        ({ queryKey }: { queryKey: [string, string[], FbMenuModeT] }) => {
            if (queryKey[2] !== FbMenuModeT.AddToAlbum) {
                return []
            }
            return getFoldersMedia(queryKey[1])
        },
        []
    )

    const medias = useQuery({
        queryKey: ['selected-medias', activeItems.map((i) => i.Id()), menuMode],
        queryFn: getMediasInFolders,
    })

    if (menuMode !== FbMenuModeT.AddToAlbum) {
        return <></>
    }

    return (
        <div className="add-to-album-menu">
            {medias.data && medias.data.length !== 0 && (
                <p className="animate-fade">
                    Add {medias.data.length} media to Albums
                </p>
            )}
            {medias.data && medias.data.length === 0 && (
                <p className="animate-fade">No valid media selected</p>
            )}
            {medias.isLoading && (
                <p className="animate-fade">Loading media...</p>
            )}
            <div className="no-scrollbar grid grid-cols-2 gap-3 h-max max-h-[350px] overflow-y-scroll pt-1">
                {albums.data?.map((a) => {
                    return (
                        <AlbumCover
                            key={a.name}
                            a={a}
                            medias={medias.data}
                            albums={albums}
                        />
                    )
                })}
            </div>
            {newAlbum && (
                <WeblensInput
                    squareSize={40}
                    autoFocus
                    closeInput={() => setNewAlbum(false)}
                    onComplete={async (v: string) => {
                        await createAlbum(v).then(() => {
                            setNewAlbum(false)
                            albums.refetch()
                        })
                    }}
                />
            )}
            {!newAlbum && (
                <WeblensButton
                    fillWidth
                    label={'New Album'}
                    Left={IconLibraryPlus}
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        setNewAlbum(true)
                    }}
                />
            )}
            <WeblensButton
                fillWidth
                label={'Back'}
                Left={IconArrowLeft}
                centerContent
                onClick={(e) => {
                    e.stopPropagation()
                    setMenu({ menuState: FbMenuModeT.Default })
                }}
            />
        </div>
    )
}

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
        <div className="default-grid no-scrollbar">
            <WeblensButton
                squareSize={100}
                Left={IconFileExport}
                centerContent
                disabled={menuTarget === ''}
                onMouseOver={() =>
                    setFooterNote({ hint: 'Put Back', danger: false })
                }
                onMouseLeave={() => setFooterNote({ hint: '', danger: false })}
                onClick={async (e) => {
                    e.stopPropagation()
                    const res = await UnTrashFiles(
                        activeItems.map((f) => f.Id())
                    )

                    if (!res.ok) {
                        return false
                    }

                    setMenu({ menuState: FbMenuModeT.Closed })
                }}
            />
            <WeblensButton
                squareSize={100}
                Left={IconTrash}
                centerContent
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
                    const res = await DeleteFiles(toDeleteIds)

                    if (!res.ok) {
                        return false
                    }
                    setMenu({ menuState: FbMenuModeT.Closed })
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
    const nav = useNavigate()
    const user = useSessionStore((state) => state.user)

    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const mode = useFileBrowserStore((state) => state.fbMode)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    if (menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className="default-grid no-scrollbar"
            data-visible={menuMode === FbMenuModeT.Default && menuTarget === ''}
        >
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconFolderPlus}
                    squareSize={100}
                    centerContent
                    disabled={!folderInfo.IsModifiable()}
                    onMouseOver={() =>
                        setFooterNote({ hint: 'New Folder', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        setFooterNote({ hint: '', danger: false })
                        setMenu({ menuState: FbMenuModeT.NameFolder })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconFileAnalytics}
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Folder Stats', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={() =>
                        nav(
                            `/files/stats/${
                                mode === FbModeT.external
                                    ? mode
                                    : folderInfo.Id()
                            }`
                        )
                    }
                />
            </div>

            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconUsersPlus}
                    squareSize={100}
                    disabled={
                        folderInfo.Id() === user.homeId ||
                        folderInfo.Id() === user.trashId
                    }
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Share', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        setMenu({ menuState: FbMenuModeT.Sharing })
                    }}
                />
            </div>
        </div>
    )
}
