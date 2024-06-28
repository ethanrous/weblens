import {
    IconArrowLeft,
    IconDownload,
    IconFileAnalytics,
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
import { useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { MiniAlbumCover } from '../Albums/AlbumDisplay'
import { addMediaToAlbum, createAlbum, getAlbums } from '../Albums/AlbumQuery'
import { AutocompleteUsers } from '../api/ApiFetch'

import '../Pages/FileBrowser/style/fileBrowserMenuStyle.scss'

import {
    CreateFolder,
    DeleteFiles,
    shareFile,
    TrashFiles,
    updateFileShare,
} from '../api/FileBrowserApi'
import {
    useClick,
    useKeyDown,
    useResize,
    useWindowSize,
} from '../components/hooks'
import WeblensButton from '../components/WeblensButton'
import WeblensInput from '../components/WeblensInput'
import { MediaContext, UserContext, WebsocketContext } from '../Context'
import { downloadSelected } from '../Pages/FileBrowser/FileBrowserLogic'
import { AlbumData, AuthHeaderT, FbStateT, UserContextT } from '../types/Types'
import { clamp } from '../util'
import { FbMenuModeT, WeblensFile } from './File'
import { FbContext, FbModeT } from './filesContext'
import { getFoldersMedia } from './FilesQuery'

const activeItemsFromState = (
    fbState: FbStateT
): {
    items: WeblensFile[]
    anyDisplayable: boolean
    mediaCount: number
} => {
    if (fbState.dirMap.size === 0) {
        return { items: [], anyDisplayable: false, mediaCount: 0 }
    }
    const selected = Boolean(fbState.selected.get(fbState.menuTargetId))
    const itemIds = selected
        ? Array.from(fbState.selected.keys())
        : [fbState.menuTargetId]
    let mediaCount = 0
    const items = itemIds.map((i) => {
        const item = fbState.dirMap.get(i)
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
    const { fbState, fbDispatch } = useContext(FbContext)
    const [targetItem, setTargetItem] = useState<WeblensFile>(null)

    useEffect(() => {
        if (fbState.menuTargetId === '') {
            setTargetItem(fbState.folderInfo)
        } else {
            setTargetItem(fbState.dirMap.get(fbState.menuTargetId))
        }
    }, [fbState.menuTargetId, fbState.folderInfo])

    const extrasText = useMemo(() => {
        if (
            fbState.selected.get(targetItem?.Id()) &&
            fbState.selected.size > 1
        ) {
            return `+${fbState.selected.size - 1}`
        } else {
            return ''
        }
    }, [targetItem, fbState.selected])

    const Icon = useMemo(() => {
        if (fbState.menuMode === FbMenuModeT.NameFolder) {
            return IconFolderPlus
        }
        return targetItem?.GetFileIcon()
    }, [targetItem, fbState.menuMode])

    return (
        <div className="file-menu-title">
            {fbState.menuMode === FbMenuModeT.NameFolder && (
                <div className="flex flex-grow absolute w-full">
                    <WeblensButton
                        Left={IconArrowLeft}
                        onClick={(e) => {
                            e.stopPropagation()
                            fbDispatch({
                                type: 'set_menu_open',
                                menuMode: FbMenuModeT.Default,
                            })
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

export function FileContextMenu() {
    const { authHeader } = useContext(UserContext)
    const { fbState, fbDispatch } = useContext(FbContext)
    const [menuRef, setMenuRef] = useState<HTMLDivElement>(null)

    useEffect(() => {
        fbDispatch({
            type: 'set_block_focus',
            block: fbState.menuMode !== FbMenuModeT.Closed,
        })
    }, [fbState.menuMode])

    useKeyDown('Escape', (e) => {
        if (fbState.menuMode !== FbMenuModeT.Closed) {
            e.stopPropagation()
            fbDispatch({ type: 'set_menu_open', menuMode: FbMenuModeT.Closed })
        }
    })

    useClick((e) => {
        if (fbState.menuMode !== FbMenuModeT.Closed && e.button === 0) {
            e.stopPropagation()
            fbDispatch({ type: 'set_menu_open', menuMode: FbMenuModeT.Closed })
        }
    }, menuRef)

    const { width, height } = useWindowSize()
    const { height: menuHeight, width: menuWidth } = useResize(menuRef)

    const menuPosStyle = useMemo(() => {
        return {
            top: clamp(
                fbState.menuPos.y,
                8 + menuHeight / 2,
                height - menuHeight / 2 - 8
            ),
            left: clamp(
                fbState.menuPos.x,
                8 + menuWidth / 2,
                width - menuWidth / 2 - 8
            ),
        }
    }, [
        fbState.menuPos.y,
        fbState.menuPos.x,
        menuHeight,
        menuWidth,
        width,
        height,
    ])

    return (
        <div
            className={'backdrop-menu'}
            data-mode={fbState.menuMode}
            ref={setMenuRef}
            onClick={(e) => {
                e.stopPropagation()
                fbDispatch({
                    type: 'set_menu_open',
                    menuMode: FbMenuModeT.Closed,
                })
            }}
            style={menuPosStyle}
        >
            <MenuTitle />
            {fbState.viewingPast !== null && <div />}
            <StandardFileMenu />
            <BackdropDefaultItems />
            <FileShareMenu />
            <NewFolderName />
            <AddToAlbum />
        </div>
    )
}

function StandardFileMenu() {
    const { usr, authHeader }: UserContextT = useContext(UserContext)
    const [itemInfo, setItemInfo] = useState(new WeblensFile({}))
    const { fbState, fbDispatch } = useContext(FbContext)
    const wsSend = useContext(WebsocketContext)

    useEffect(() => {
        const info = fbState.dirMap.get(fbState.menuTargetId)
        if (info) {
            setItemInfo(info)
        }
    }, [fbState.dirMap.get(fbState.menuTargetId)])

    const { items }: { items: WeblensFile[] } = useMemo(() => {
        return activeItemsFromState(fbState)
    }, [fbState.menuTargetId, fbState.menuTargetId, fbState.selected])

    const wormholeId = useMemo(() => {
        if (itemInfo?.GetShare()) {
            if (itemInfo.GetShare().IsWormhole()) {
                return itemInfo.GetShare().Id()
            }
        }
    }, [itemInfo?.GetShare()])

    const inTrash = fbState.folderInfo.Id() === usr.trashId

    if (fbState.menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className={'default-grid no-scrollbar'}
            data-visible={
                fbState.menuMode === FbMenuModeT.Default &&
                fbState.menuTargetId !== ''
            }
        >
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconUsersPlus}
                    subtle
                    disabled={items.length > 1}
                    squareSize={100}
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        fbDispatch({
                            type: 'set_menu_open',
                            menuMode: FbMenuModeT.Sharing,
                        })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconPhotoShare}
                    subtle
                    squareSize={100}
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        fbDispatch({
                            type: 'set_menu_open',
                            menuMode: FbMenuModeT.AddToAlbum,
                        })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconFolderPlus}
                    subtle
                    squareSize={100}
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        fbDispatch({
                            type: 'set_menu_open',
                            menuMode: FbMenuModeT.NameFolder,
                        })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconDownload}
                    subtle
                    squareSize={100}
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        downloadSelected(
                            items,
                            fbDispatch,
                            wsSend,
                            authHeader,
                            fbState.shareId
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
                    onClick={(e) => {
                        e.stopPropagation()
                        items.forEach((i) =>
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
                    onClick={(e) => {
                        e.stopPropagation()
                        if (inTrash) {
                            DeleteFiles(
                                items.map((f) => f.Id()),
                                authHeader
                            )
                        } else {
                            TrashFiles(
                                items.map((f) => f.Id()),
                                fbState.shareId,
                                authHeader
                            )
                        }
                        fbDispatch({
                            type: 'set_menu_open',
                            menuMode: FbMenuModeT.Closed,
                        })
                    }}
                />
            </div>
        </div>
    )
}

function FileShareMenu() {
    const { authHeader } = useContext(UserContext)
    const [isPublic, setIsPublic] = useState(false)
    const { fbState, fbDispatch } = useContext(FbContext)

    const item: WeblensFile = useMemo(() => {
        const { items } = activeItemsFromState(fbState)
        if (items.length > 1) {
            return null
        } else if (items.length === 1) {
            return items[0]
        } else {
            return fbState.folderInfo
        }
    }, [
        fbState.menuTargetId,
        fbState.menuTargetId,
        fbState.selected,
        fbState.folderInfo.Id(),
    ])

    const [accessors, setAccessors] = useState([])
    useEffect(() => {
        if (!item) {
            return
        }
        item.LoadShare(authHeader).then((share) => {
            if (share) {
                console.log(share)
                setIsPublic(share.IsPublic())
                setAccessors(share.GetAccessors())
            } else {
                setIsPublic(false)
            }
        })
    }, [item])

    const [userSearch, setUserSearch] = useState('')
    const [userSearchResults, setUserSearchResults] = useState([])
    useEffect(() => {
        AutocompleteUsers(userSearch, authHeader).then((us) => {
            us = us.filter((u) => !accessors.includes(u))
            setUserSearchResults(us)
        })
    }, [userSearch])

    useEffect(() => {
        if (fbState.menuMode !== FbMenuModeT.Sharing) {
            setUserSearch('')
            setUserSearchResults([])
        }
    }, [fbState.menuMode])

    const updateShare = useCallback(
        async (e) => {
            e.stopPropagation()
            const share = await item.LoadShare(authHeader)
            let req: Promise<Response>
            if (share) {
                req = updateFileShare(
                    share.Id(),
                    isPublic,
                    accessors,
                    authHeader
                )
            } else {
                req = shareFile(item, isPublic, accessors, authHeader)
            }
            return req
                .then((j) => {
                    return true
                })
                .catch((r) => {
                    console.error(r)
                    return false
                })
        },
        [item, isPublic, accessors, authHeader]
    )

    if (fbState.menuMode === FbMenuModeT.Closed) {
        return <></>
    }
    console.log(isPublic)

    return (
        <div
            className="file-share-menu"
            data-visible={fbState.menuMode === FbMenuModeT.Sharing}
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
                                .then(async (r) => {
                                    const share =
                                        await item.LoadShare(authHeader)
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
                    <div className="flex flex-col w-full bg-raised-grey absolute gap-1 rounded p-1 z-10 mt-14 max-h-32 overflow-y-scroll drop-shadow-xl">
                        {userSearchResults.map((u) => {
                            return (
                                <div
                                    className="user-autocomplete-row"
                                    key={u}
                                    onClick={(e) => {
                                        e.stopPropagation()
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
                                    <p>{u}</p>
                                    <IconPlus />
                                </div>
                            )
                        })}
                    </div>
                )}
                <p className="text-white">Shared With</p>
            </div>
            <div className="flex flex-row w-full h-full p-2 m-2 mt-0 rounded outline outline-main-accent justify-center">
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
                            fbDispatch({
                                type: 'set_menu_open',
                                menuMode: FbMenuModeT.Default,
                            })
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

function NewFolderName() {
    const { fbState, fbDispatch } = useContext(FbContext)
    const { authHeader } = useContext(UserContext)

    const { items }: { items: WeblensFile[] } = useMemo(() => {
        return activeItemsFromState(fbState)
    }, [fbState.menuTargetId, fbState.menuTargetId, fbState.selected])

    if (fbState.menuMode !== FbMenuModeT.NameFolder) {
        return <></>
    }

    return (
        <div className="new-folder-menu">
            {/*<IconFolder size={50} />*/}
            <WeblensInput
                placeholder="New Folder Name"
                autoFocus
                height={50}
                buttonIcon={IconPlus}
                onComplete={(newName) => {
                    CreateFolder(
                        fbState.folderInfo.Id(),
                        newName,
                        items.map((f) => f.Id()),
                        false,
                        fbState.shareId,
                        authHeader
                    )
                        .then(() =>
                            fbDispatch({
                                type: 'set_menu_open',
                                menuMode: FbMenuModeT.Closed,
                            })
                        )
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
                authHeader={authHeader}
            />
        </div>
    )
}

function AddToAlbum() {
    const { fbState, fbDispatch } = useContext(FbContext)
    const { mediaState, mediaDispatch } = useContext(MediaContext)
    const { authHeader } = useContext(UserContext)
    const [newAlbum, setNewAlbum] = useState(false)

    const albums = useQuery({
        queryKey: ['albums'],
        queryFn: () =>
            getAlbums(false, authHeader).then((as) =>
                as.sort((a, b) => {
                    return a.name.localeCompare(b.name)
                })
            ),
    })

    useEffect(() => {
        setNewAlbum(false)
    }, [fbState.menuMode])

    const activeItems = useMemo(() => {
        return activeItemsFromState(fbState).items.map((f) => f.Id())
    }, [fbState.selected, fbState.menuTargetId])

    const getMediasInFolders = useCallback(
        ({ queryKey }: { queryKey }) => {
            if (queryKey[2] !== FbMenuModeT.AddToAlbum) {
                return []
            }
            return getFoldersMedia(queryKey[1], authHeader)
        },
        [authHeader]
    )

    const medias = useQuery({
        queryKey: ['selected-medias', activeItems, fbState.menuMode],
        queryFn: getMediasInFolders,
    })

    if (fbState.menuMode !== FbMenuModeT.AddToAlbum) {
        return <></>
    }

    return (
        <div className="new-folder-menu">
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
                            a={a}
                            medias={medias.data}
                            albums={albums}
                            authHeader={authHeader}
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
                        createAlbum(v, authHeader).then(() => {
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
                    fbDispatch({
                        type: 'set_menu_open',
                        menuMode: FbMenuModeT.Default,
                    })
                }}
            />
        </div>
    )
}

function BackdropDefaultItems() {
    const nav = useNavigate()
    const { fbState, fbDispatch } = useContext(FbContext)
    const { usr } = useContext(UserContext)

    if (fbState.menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className="default-grid"
            data-visible={
                fbState.menuMode === FbMenuModeT.Default &&
                fbState.menuTargetId === ''
            }
        >
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconFolderPlus}
                    subtle
                    squareSize={100}
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        fbDispatch({
                            type: 'set_menu_open',
                            menuMode: FbMenuModeT.NameFolder,
                        })
                    }}
                />
            </div>
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconFileAnalytics}
                    subtle
                    squareSize={100}
                    centerContent
                    onClick={() =>
                        nav(
                            `/files/stats/${
                                fbState.fbMode === FbModeT.external
                                    ? fbState.fbMode
                                    : fbState.folderInfo.Id()
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
                        fbState.folderInfo.Id() === usr.homeId ||
                        fbState.folderInfo.Id() === usr.trashId
                    }
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        fbDispatch({
                            type: 'set_menu_open',
                            menuMode: FbMenuModeT.Sharing,
                        })
                    }}
                />
            </div>
        </div>
    )
}
