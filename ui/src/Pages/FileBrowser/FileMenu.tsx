import {
    IconArrowLeft,
    IconFileAnalytics,
    IconFolder,
    IconFolderPlus,
    IconLink,
    IconMinus,
    IconPlus,
    IconTrash,
    IconUser,
    IconUsers,
    IconUsersPlus,
} from '@tabler/icons-react'
import { useNavigate } from 'react-router-dom'
import { useCallback, useContext, useEffect, useMemo, useState } from 'react'

import {
    CreateFolder,
    DeleteFiles,
    shareFiles,
    TrashFiles,
    updateFileShare,
} from '../../api/FileBrowserApi'
import { FbStateT, UserContextT } from '../../types/Types'
import { WeblensFile } from '../../classes/File'
import { UserContext } from '../../Context'
import { FbMenuModeT } from './FileBrowserStyles'
import {
    useClick,
    useKeyDown,
    useResize,
    useWindowSize,
} from '../../components/hooks'
import { WeblensButton } from '../../components/WeblensButton'
import { FbContext, FbModeT } from './FileBrowser'
import { WeblensShare } from '../../classes/Share'
import WeblensInput from '../../components/WeblensInput'
import { AutocompleteUsers } from '../../api/ApiFetch'

import './style/fileBrowserMenuStyle.scss'
import { clamp } from '../../util'

const activeItemsFromState = (fbState: FbStateT, selected: boolean) => {
    if (fbState.dirMap.size === 0) {
        return { items: [], anyDisplayable: false }
    }
    const itemIds = selected
        ? Array.from(fbState.selected.keys())
        : [fbState.menuTargetId]
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
}

const MenuTitle = () => {
    const { fbState, fbDispatch } = useContext(FbContext)
    const [targetItem, setTargetItem] = useState<WeblensFile>(null)

    useEffect(() => {
        if (fbState.menuTargetId === '') {
            console.log(fbState.menuTargetId, fbState.folderInfo)
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
            return `+ ${fbState.selected.size - 1} more`
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
            <div className="flex flex-col w-max">
                <div className="flex flex-row items-center justify-center w-full h-8">
                    <div className="flex shrink-0 justify-center items-center h-8 w-8">
                        {Icon && <Icon />}
                    </div>
                    <p className="font-semibold select-none text-nowrap truncate text-sm">
                        {targetItem?.GetFilename()}
                    </p>
                </div>
                {extrasText && (
                    <p className="flex w-full justify-end text-xs select-none h-3">
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

    const setMenuMode = useCallback(
        (m: FbMenuModeT) => {
            // fbDispatch({ type: 'set_block_focus', block: m !== FbMenuModeT.Closed })
            fbDispatch({ type: 'set_menu_open', menuMode: m })
        },
        [fbDispatch]
    )

    useKeyDown('Escape', (e) => {
        if (fbState.menuMode !== FbMenuModeT.Closed) {
            e.stopPropagation()
            fbDispatch({ type: 'set_menu_open', menuMode: FbMenuModeT.Closed })
        }
    })

    useClick(() => {
        if (fbState.menuMode !== FbMenuModeT.Closed) {
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
            <FileShareMenu fileInfo={fbState.folderInfo} />
            <NewFolderName />
        </div>
    )
}

function StandardFileMenu() {
    const { usr, authHeader }: UserContextT = useContext(UserContext)
    const [itemInfo, setItemInfo] = useState(new WeblensFile({}))

    const { fbState, fbDispatch } = useContext(FbContext)
    const selected: boolean = Boolean(
        fbState.selected.get(fbState.menuTargetId)
    )

    useEffect(() => {
        const info = fbState.dirMap.get(fbState.menuTargetId)
        if (info) {
            setItemInfo(info)
        }
    }, [fbState.dirMap.get(fbState.menuTargetId)])

    const { items }: { items: WeblensFile[] } = useMemo(() => {
        return activeItemsFromState(fbState, selected)
    }, [fbState.menuTargetId, selected, fbState.selected])

    const wormholeId = useMemo(() => {
        if (itemInfo?.GetShares()) {
            const whs = itemInfo.GetShares().filter((s) => s.IsWormhole())
            if (whs.length !== 0) {
                return whs[0].Id()
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
    const inShare = fbState.fbMode === FbModeT.share

    let trashName: string
    if (inTrash) {
        trashName = 'Delete Forever'
    } else if (inShare) {
        trashName = 'Unshare Me'
    } else {
        trashName = 'Delete'
    }

    if (fbState.menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className={'default-grid'}
            data-visible={
                fbState.menuMode === FbMenuModeT.Default &&
                fbState.menuTargetId !== ''
            }
        >
            <div className="default-menu-icon">
                <WeblensButton
                    Left={IconUsersPlus}
                    subtle
                    width={80}
                    squareSize={80}
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
                    Left={IconFolderPlus}
                    subtle
                    width={80}
                    squareSize={80}
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
                    Left={IconTrash}
                    subtle
                    danger
                    width={80}
                    squareSize={80}
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

function FileShareMenu({ fileInfo }: { fileInfo: WeblensFile }) {
    const { authHeader } = useContext(UserContext)
    const [isPublic, setIsPublic] = useState(false)
    const { fbState, fbDispatch } = useContext(FbContext)

    const [accessors, setAccessors] = useState([])
    useEffect(() => {
        const shares: WeblensShare[] = fileInfo.GetShares()
        if (shares.length !== 0) {
            setIsPublic(shares[0].IsPublic())
            setAccessors(shares[0].GetAccessors())
        } else {
            setIsPublic(false)
        }
    }, [fileInfo])

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
            const shares = fileInfo.GetShares()
            let req: Promise<Response>
            if (shares.length !== 0) {
                req = updateFileShare(
                    shares[0].Id(),
                    isPublic,
                    accessors,
                    authHeader
                )
            } else {
                req = shareFiles([fileInfo], isPublic, accessors, authHeader)
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
        [fileInfo, isPublic, accessors, authHeader]
    )

    if (fbState.menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    return (
        <div
            className="file-share-menu"
            data-visible={fbState.menuMode === FbMenuModeT.Sharing}
        >
            <div className="flex flex-row w-full">
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
                                const shares = fileInfo.GetShares()
                                return navigator.clipboard
                                    .writeText(shares[0].GetPublicLink())
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
            <div className="flex flex-col w-full gap-1 items-center">
                <div className="h-10 w-full mt-3 mb-3 z-20">
                    <WeblensInput
                        value={userSearch}
                        valueCallback={setUserSearch}
                        placeholder="Add users"
                        onComplete={null}
                        icon={<IconUsersPlus />}
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
                                        width={40}
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
                <WeblensButton
                    squareSize={40}
                    centerContent
                    fillWidth
                    label="Save"
                    onClick={updateShare}
                />
            </div>
        </div>
    )
}

function NewFolderName() {
    const { fbState, fbDispatch } = useContext(FbContext)
    const { authHeader } = useContext(UserContext)

    const selected: boolean = useMemo(() => {
        return Boolean(fbState.selected.get(fbState.menuTargetId))
    }, [fbState.selected, fbState.menuTargetId])

    const { items }: { items: WeblensFile[] } = useMemo(() => {
        return activeItemsFromState(fbState, selected)
    }, [fbState.menuTargetId, selected, fbState.selected])

    if (fbState.menuMode !== FbMenuModeT.NameFolder) {
        return <></>
    }

    return (
        <div className="new-folder-menu">
            <IconFolder size={50} />
            <WeblensInput
                placeholder="New Folder Name"
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
                    width={80}
                    squareSize={80}
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
                    width={80}
                    squareSize={80}
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
                    width={80}
                    squareSize={80}
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
