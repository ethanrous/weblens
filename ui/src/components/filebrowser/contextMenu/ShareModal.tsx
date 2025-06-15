import {
    IconChevronRight,
    IconCopy,
    IconPlus,
    IconUser,
    IconUsers,
    IconUsersPlus,
} from '@tabler/icons-react'
import { QueryObserverResult, RefetchOptions } from '@tanstack/react-query'
import UsersApi from '@weblens/api/UserApi'
import { PermissionsParams, UserInfo } from '@weblens/api/swag'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { useClick } from '@weblens/lib/hooks'
import useShare from '@weblens/lib/hooks/useShare'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { useMessagesController } from '@weblens/store/MessagesController'
import WeblensFile, { FbMenuModeT } from '@weblens/types/files/File'
import { WeblensShare } from '@weblens/types/share/share'
import { Ref, useEffect, useRef, useState } from 'react'

async function toggleParam(
    username: string,
    share: WeblensShare,
    param: keyof PermissionsParams
) {
    const existingPerms = share.permissions[username]
    await share.updateAccessorPerms(username, {
        ...existingPerms,
        [param]: !existingPerms[param],
    })
}

function UserSearch({
    accessors,
    addUser,
}: {
    accessors: UserInfo[]
    addUser: (username: UserInfo) => Promise<void>
}) {
    const [userSearch, setUserSearch] = useState('')
    const [searchMenuOpen, setSearchMenuOpen] = useState(false)
    const [userSearchResults, setUserSearchResults] = useState<UserInfo[]>([])
    const searchRef = useRef<HTMLInputElement>(null)

    useClick(() => {
        setSearchMenuOpen(false)
    }, searchRef)

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
    }, [userSearch, accessors])

    return (
        <div
            ref={searchRef}
            className="relative flex w-full flex-col items-center gap-1"
        >
            <div className="z-20 mt-3 mb-3 h-10 w-full">
                <WeblensInput
                    value={userSearch}
                    valueCallback={setUserSearch}
                    placeholder="Add users"
                    Icon={IconUsersPlus}
                    openInput={() => setSearchMenuOpen(true)}
                />
            </div>
            {userSearchResults.length !== 0 && searchMenuOpen && (
                <div className="no-scrollbar bg-card-background-primary floating-card absolute z-10 mt-14 flex max-h-32 w-full flex-col gap-1 rounded-sm border">
                    {userSearchResults.map((user) => {
                        return (
                            <div
                                className="hover:bg-card-background-hover relative flex h-10 w-full cursor-pointer flex-row items-center rounded p-4"
                                key={user.username}
                                onClick={async (e) => {
                                    e.stopPropagation()
                                    e.preventDefault()
                                    await addUser(user)
                                    setUserSearchResults((p) => {
                                        const newP = [...p]
                                        newP.splice(newP.indexOf(user), 1)
                                        return newP
                                    })
                                }}
                            >
                                <span>{user.fullName}</span>
                                <span className="text-text-secondary ml-1">
                                    [{user.username}]
                                </span>
                                <IconPlus className="ml-auto" />
                            </div>
                        )
                    })}
                </div>
            )}
        </div>
    )
}

function UserPermissions({
    user,
    share,
    refetch,
}: {
    user: UserInfo
    share: WeblensShare
    refetch: (
        options?: RefetchOptions
    ) => Promise<QueryObserverResult<WeblensShare, Error>>
}) {
    const perms = share.permissions[user.username]

    return (
        <div className="border-l-border-primary ml-2 flex w-1/2 flex-col items-center border-l-2 px-2">
            <h5>{user.username} Permissions</h5>
            {perms && (
                <div className="flex flex-col gap-1 px-10">
                    <WeblensButton
                        label="Can Download"
                        fillWidth
                        flavor={perms.canDownload ? 'default' : 'outline'}
                        onClick={async () => {
                            await toggleParam(
                                user.username,
                                share,
                                'canDownload'
                            )
                            refetch()
                        }}
                    />
                    <WeblensButton
                        label="Can Edit"
                        fillWidth
                        flavor={perms.canEdit ? 'default' : 'outline'}
                        onClick={async () => {
                            await toggleParam(user.username, share, 'canEdit')
                            refetch()
                        }}
                    />
                    <WeblensButton
                        label="Can Delete"
                        fillWidth
                        flavor={perms.canDelete ? 'default' : 'outline'}
                        onClick={async () => {
                            await toggleParam(user.username, share, 'canDelete')
                            refetch()
                        }}
                    />

                    <WeblensButton
                        label="Unshare"
                        danger
                        className="mt-4"
                        fillWidth
                        onClick={async () => {
                            await share.removeAccessor(user.username)
                            refetch()
                        }}
                    />
                </div>
            )}
        </div>
    )
}

export function ShareModal({
    targetFile,
    ref,
}: {
    targetFile: WeblensFile | undefined
    ref: Ref<HTMLDivElement>
}) {
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    const [focusedUser, setFocusedUser] = useState<UserInfo | null>(null)

    if (!targetFile) {
        console.warn(
            'ShareModal: No targetFile provided, using folderInfo instead'
        )
        targetFile = folderInfo
    }

    const { share, refetchShare, shareLoading } = useShare(targetFile)

    // const {
    //     data: share,
    //     refetch: refetchShare,
    //     isLoading: shareLoading,
    // } = useQuery<WeblensShare | undefined>({
    //     queryKey: ['share', targetFile.Id()],
    //     initialData: undefined,
    //     queryFn: async () => {
    //         const share = await targetFile.GetShare(true).catch(ErrorHandler)
    //
    //         console.log('GOT SHREA', share)
    //         if (!share) {
    //             return
    //         }
    //
    //         return share
    //     },
    // })

    useEffect(() => {
        if (!share) {
            return
        }

        if (share.accessors.length === 0) {
            setFocusedUser(null)
        }
    }, [share])

    if (menuMode === FbMenuModeT.Closed) {
        return <></>
    }

    const shareLink = share?.GetLink()

    return (
        <div
            className="fullscreen-modal"
            onClick={(e) => e.stopPropagation()}
            ref={ref}
        >
            {shareLoading && (
                <div className="bg-background-primary absolute top-0 left-0 h-full w-full opacity-50" />
            )}

            <div className="bg-background-primary floating-card m-auto flex h-1/2 w-1/2 flex-col rounded-lg p-5">
                <h3 className="mb-3 flex min-h-max w-full max-w-full truncate text-nowrap">
                    {'Share "' + targetFile.GetFilename() + '"'}
                </h3>
                <div className="flex w-full flex-row gap-1">
                    <WeblensButton
                        label={share?.public ? 'Public' : 'Private'}
                        Left={share?.public ? IconUsers : IconUser}
                        flavor={share?.public ? 'default' : 'outline'}
                        fillWidth
                        onClick={async (e) => {
                            e.stopPropagation()
                            if (!share) {
                                console.error('Share is not defined')
                                return
                            }

                            await share.setPublic(!share.public)
                            await refetchShare()
                        }}
                    />
                    <div className="flex h-10 w-1/2 shrink-0 items-center gap-1 rounded-sm border p-1">
                        <span
                            className="no-scrollbar data-[has-share=true]:text-text-primary data-[has-share=false]:text-text-tertiary mr-1 ml-1 overflow-scroll select-all data-[has-share=false]:cursor-not-allowed data-[has-share=false]:select-none data-[has-share=true]:cursor-pointer"
                            data-has-share={Boolean(share?.shareId)}
                        >
                            {share?.shareId ? shareLink : 'Not Shared'}
                        </span>
                        <WeblensButton
                            Left={IconCopy}
                            disabled={!share?.shareId}
                            tooltip={
                                !share?.shareId ? 'Not shared' : 'Copy Link'
                            }
                            size="small"
                            containerClassName="ml-auto"
                            onClick={async (e) => {
                                e.stopPropagation()
                                if (!share || !share.shareId) {
                                    console.error(
                                        'Share is not defined fully defined'
                                    )
                                    return
                                }

                                if (navigator.clipboard === undefined) {
                                    useMessagesController
                                        .getState()
                                        .addMessage({
                                            severity: 'error',
                                            title: 'Copy Error',
                                            text: 'Copy is not supported in this context. You must use https.',
                                            duration: 5000,
                                        })
                                    return
                                }

                                const shareUrl = share.GetLink()
                                await navigator.clipboard.writeText(shareUrl)
                            }}
                        />
                    </div>
                </div>

                <UserSearch
                    accessors={share?.accessors || []}
                    addUser={async (user: UserInfo) => {
                        if (!share) {
                            console.error('Share is not defined')
                            return
                        }

                        await share.addAccessor(user.username)
                        await refetchShare()
                    }}
                />

                <h4 className="mb-2 w-max">Shared With</h4>

                <div className="no-scrollbar flex h-full w-full rounded-sm border p-2">
                    <div className="flex min-w-[75%] flex-col">
                        {share?.public && (
                            <div className="border-b pb-2 mb-2">
                                <UserShareRow
                                    u={{ username: 'PUBLIC' } as UserInfo}
                                    focusedUser={focusedUser}
                                    setFocusedUser={setFocusedUser}
                                />
                            </div>
                        )}
                        {share && share.accessors.length === 0 && (
                            <span className="m-auto">
                                Not Shared With Anyone Specific
                            </span>
                        )}
                        {share &&
                            share.accessors.length !== 0 &&
                            share.accessors.map((u: UserInfo) => {
                                return (
                                    <UserShareRow
                                        key={u.username}
                                        u={u}
                                        focusedUser={focusedUser}
                                        setFocusedUser={setFocusedUser}
                                    />
                                )
                            })}
                    </div>
                    {focusedUser && (
                        <UserPermissions
                            user={focusedUser}
                            share={share}
                            refetch={refetchShare}
                        />
                    )}
                </div>
                <div className="mt-3 ml-auto flex w-1/3">
                    <WeblensButton
                        fillWidth
                        label="Done"
                        onClick={() =>
                            setMenu({ menuState: FbMenuModeT.Closed })
                        }
                    />
                </div>
            </div>
        </div>
    )
}

function UserShareRow({
    u,
    focusedUser,
    setFocusedUser,
}: {
    u: UserInfo
    focusedUser: UserInfo | null
    setFocusedUser: (user: UserInfo | null) => void
}) {
    return (
        <div
            className="bg-background-secondary hover:bg-card-background-hover group/user flex h-10 w-full cursor-pointer items-center rounded p-2 transition"
            onClick={() => {
                if (focusedUser === u) {
                    setFocusedUser(null)
                } else {
                    setFocusedUser(u)
                }
            }}
        >
            {u.username === 'PUBLIC' && <span>{u.username}</span>}
            {u.username !== 'PUBLIC' && (
                <>
                    <IconUser className="text-text-tertiary" />
                    <span>{u.fullName}</span>
                    <span className="text-color-text-secondary ml-1">
                        [{u.username}]
                    </span>
                </>
            )}
            <IconChevronRight className="text-text-tertiary ml-auto" />
        </div>
    )
}
