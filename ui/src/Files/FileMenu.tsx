import {
    IconArrowLeft,
    IconDownload,
    IconFileAnalytics,
    IconFileExport,
    IconFolderPlus,
    IconLibraryPlus,
    IconLink,
    IconMinus,
    IconPhotoShare,
    IconPlus,
    IconScan,
    IconTrash,
    IconUser,
    IconUsers,
    IconUsersPlus,
} from '@tabler/icons-react'
import { useQuery, UseQueryResult } from '@tanstack/react-query'
import {
    ReactElement,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
} from 'react'
import { useNavigate } from 'react-router-dom'
import { MiniAlbumCover } from '../Albums/AlbumDisplay'
import { addMediaToAlbum, createAlbum, getAlbums } from '../Albums/AlbumQuery'
import { AutocompleteUsers } from '../api/ApiFetch'

import '../Pages/FileBrowser/style/fileBrowserMenuStyle.scss'

import {
    addUsersToFileShare,
    CreateFolder,
    DeleteFiles,
    setFileSharePublic,
    shareFile,
    TrashFiles,
    UnTrashFiles,
} from '../api/FileBrowserApi'
import {
    useClick,
    useKeyDown,
    useResize,
    useWindowSize,
} from '../components/hooks'
import WeblensButton from '../components/WeblensButton'
import WeblensInput from '../components/WeblensInput'
import { WebsocketContext } from '../Context'
import { downloadSelected } from '../Pages/FileBrowser/FileBrowserLogic'
import { AlbumData, AuthHeaderT, UserInfoT } from '../types/Types'
import { clamp } from '../util'
import { FbMenuModeT, SelectedState, WeblensFile } from './File'
import { TaskProgContext } from './FBTypes'
import { getFoldersMedia } from './FilesQuery'
import {
    FbModeT,
    useFileBrowserStore,
} from '../Pages/FileBrowser/FBStateControl'
import { useSessionStore } from '../components/UserInfo'

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
        if (item.GetMediaId() || item.IsFolder()) {
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

    const Icon = useMemo(() => {
        if (menuMode === FbMenuModeT.NameFolder) {
            return IconFolderPlus
        }
        return targetItem?.GetFileIcon()
    }, [targetItem, menuMode])

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
                <div className="flex shrink-0 justify-center items-center h-8 w-7">
                    {Icon && <Icon />}
                </div>
                <p className="font-semibold w-max select-none text-nowrap truncate text-sm">
                    {targetItem?.GetFilename()}
                </p>
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
                <p className="text-nowrap">{footerNote.hint}</p>
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
    const viewingPast = useFileBrowserStore((state) => state.viewingPast)
    const activeItems = useFileBrowserStore((state) =>
        activeItemsFromState(state.filesMap, state.selected, state.menuTargetId)
    )

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const setBlock = useFileBrowserStore((state) => state.setBlockFocus)

    useEffect(() => {
        setBlock(menuMode !== FbMenuModeT.Closed)
    }, [menuMode])

    useKeyDown('Escape', (e) => {
        if (menuMode !== FbMenuModeT.Closed) {
            e.stopPropagation()
            setMenu({ menuState: FbMenuModeT.Closed })
        }
    })

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

    let menuBody: ReactElement
    if (user.trashId === folderInfo.Id()) {
        menuBody = (
            <InTrashMenu
                activeItems={activeItems.items}
                setFooterNote={setFooterNote}
            />
        )
    } else if (menuMode === FbMenuModeT.Default) {
        if (viewingPast) {
            menuBody = <HistoryFileMenu setFooterNote={setFooterNote} />
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
        menuBody = <FileShareMenu activeItems={activeItems.items} />
    } else if (menuMode === FbMenuModeT.AddToAlbum) {
        menuBody = <AddToAlbum activeItems={activeItems.items} />
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
                {viewingPast !== null && <div />}
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
    const auth = useSessionStore((state) => state.auth)
    const { progDispatch } = useContext(TaskProgContext)
    const wsSend = useContext(WebsocketContext)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)

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
                    Left={IconUsersPlus}
                    subtle
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
                    Left={IconPhotoShare}
                    subtle
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Add to Album', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        setMenu({ menuState: FbMenuModeT.AddToAlbum })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconFolderPlus}
                    subtle
                    squareSize={100}
                    centerContent
                    disabled={!folderInfo.IsModifiable()}
                    onMouseOver={() =>
                        setFooterNote({
                            hint: 'New Folder With Selected',
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
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconDownload}
                    subtle
                    squareSize={100}
                    centerContent
                    onMouseOver={() =>
                        setFooterNote({ hint: 'Download', danger: false })
                    }
                    onMouseLeave={() =>
                        setFooterNote({ hint: '', danger: false })
                    }
                    onClick={(e) => {
                        e.stopPropagation()
                        downloadSelected(
                            activeItems.items,
                            removeLoading,
                            progDispatch,
                            wsSend,
                            auth,
                            shareId
                        )
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconScan}
                    subtle
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
                    subtle
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
                    onClick={(e) => {
                        e.stopPropagation()
                        activeItems.items.forEach((f) =>
                            f.SetSelected(SelectedState.Moved)
                        )
                        TrashFiles(
                            activeItems.items.map((f) => f.Id()),
                            shareId,
                            auth
                        )
                        setMenu({ menuState: FbMenuModeT.Closed })
                    }}
                />
            </div>
        </div>
    )
}

function HistoryFileMenu({
    setFooterNote,
}: {
    setFooterNote: (n: footerNote) => void
}) {
    return null
}

function FileShareMenu({ activeItems }: { activeItems: WeblensFile[] }) {
    const auth = useSessionStore((state) => state.auth)
    const [isPublic, setIsPublic] = useState(false)

    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const setMenu = useFileBrowserStore((state) => state.setMenu)

    const item: WeblensFile = useMemo(() => {
        if (activeItems.length > 1) {
            return null
        } else if (activeItems.length === 1) {
            return activeItems[0]
        } else {
            return folderInfo
        }
    }, [activeItems, folderInfo])

    const [accessors, setAccessors] = useState<string[]>([])
    useEffect(() => {
        if (!item) {
            return
        }
        item.LoadShare(auth).then((share) => {
            if (share) {
                setIsPublic(share.IsPublic())
                setAccessors(share.GetAccessors())
            } else {
                setIsPublic(false)
            }
        })
    }, [item])

    const [userSearch, setUserSearch] = useState('')
    const [userSearchResults, setUserSearchResults] = useState<UserInfoT[]>([])
    useEffect(() => {
        AutocompleteUsers(userSearch, auth).then((us) => {
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
            const share = await item.LoadShare(auth)
            let req: Promise<Response>
            if (share) {
                req = addUsersToFileShare(
                    share.Id(),
                    accessors.map((u) => u),
                    auth
                )
                req = setFileSharePublic(share.Id(), isPublic, auth)
            } else {
                req = shareFile(
                    item,
                    isPublic,
                    accessors.map((u) => u),
                    auth
                )
            }
            return req
                .then(() => {
                    return true
                })
                .catch((r) => {
                    console.error(r)
                    return false
                })
        },
        [item, isPublic, accessors, auth]
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
                            setIsPublic(!isPublic)
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
                        onClick={async (e) => {
                            e.stopPropagation()
                            return await updateShare(e)
                                .then(async () => {
                                    const share = await item.LoadShare(auth)
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
                        onClick={updateShare}
                    />
                </div>
            </div>
        </div>
    )
}

function NewFolderName({ items }: { items: WeblensFile[] }) {
    const auth = useSessionStore((state) => state.auth)

    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const setMoved = useFileBrowserStore((state) => state.setSelectedMoved)

    if (menuMode !== FbMenuModeT.NameFolder) {
        return <></>
    }

    return (
        <div className="new-folder-menu">
            <WeblensInput
                placeholder="New Folder Name"
                autoFocus
                fillWidth
                height={50}
                buttonIcon={IconPlus}
                onComplete={(newName) => {
                    const itemIds = items.map((f) => f.Id())
                    setMoved(itemIds)
                    CreateFolder(
                        folderInfo.Id(),
                        newName,
                        itemIds,
                        false,
                        shareId,
                        auth
                    )
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
    authHeader,
}: {
    a: AlbumData
    medias: string[]
    albums: UseQueryResult<AlbumData[], Error>
    authHeader: AuthHeaderT
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
                addMediaToAlbum(a.id, medias, [], authHeader).then(() =>
                    albums.refetch()
                )
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
    const auth = useSessionStore((state) => state.auth)
    const [newAlbum, setNewAlbum] = useState(false)

    const albums = useQuery({
        queryKey: ['albums'],
        queryFn: () =>
            getAlbums(false, auth).then((as) =>
                as.sort((a, b) => {
                    return a.name.localeCompare(b.name)
                })
            ),
    })

    const menuMode = useFileBrowserStore((state) => state.menuMode)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    useEffect(() => {
        setNewAlbum(false)
    }, [menuMode])

    const getMediasInFolders = useCallback(
        ({ queryKey }: { queryKey: [string, WeblensFile[], FbMenuModeT] }) => {
            if (queryKey[2] !== FbMenuModeT.AddToAlbum) {
                return []
            }
            return getFoldersMedia(
                queryKey[1].map((f) => f.Id()),
                auth
            )
        },
        [auth]
    )

    const medias = useQuery({
        queryKey: ['selected-medias', activeItems, menuMode],
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
                            authHeader={auth}
                        />
                    )
                })}
            </div>
            {newAlbum && (
                <WeblensInput
                    height={40}
                    autoFocus
                    closeInput={() => setNewAlbum(false)}
                    onComplete={(v: string) => {
                        createAlbum(v, auth).then(() => {
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
    const auth = useSessionStore((state) => state.auth)

    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const menuTarget = useFileBrowserStore((state) => state.menuTargetId)
    const filesList = useFileBrowserStore((state) => state.filesList)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

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
                        activeItems.map((f) => f.Id()),
                        auth
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
                    let toDeleteIds = []
                    if (menuTarget === '') {
                        toDeleteIds = filesList.map((f) => f.Id())
                    } else {
                        toDeleteIds = activeItems.map((f) => f.Id())
                    }
                    const res = await DeleteFiles(toDeleteIds, auth)

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
                    subtle
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
                    subtle
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
                    subtle
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
