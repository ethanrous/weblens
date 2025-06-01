import {
    IconChevronRight,
    IconLink,
    IconPlus,
    IconUser,
    IconUsers,
    IconUsersPlus,
} from '@tabler/icons-react'
import {
    QueryObserverResult,
    RefetchOptions,
    useQuery,
} from '@tanstack/react-query'
import UsersApi from '@weblens/api/UserApi'
import { PermissionsParams, UserInfo } from '@weblens/api/swag'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { useClick } from '@weblens/lib/hooks'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
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
    }, searchRef.current)

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
                    onComplete={null}
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
    return (
        <div className="border-l-border-primary ml-2 flex w-1/2 flex-col items-center border-l-2 px-2">
            <h5>{user.username} Permissions</h5>
            <div className="flex flex-col gap-1 px-10">
                <WeblensButton
                    label="Can Download"
                    fillWidth
                    flavor={
                        share.permissions[user.username].canDownload
                            ? 'default'
                            : 'outline'
                    }
                    onClick={async () => {
                        await toggleParam(user.username, share, 'canDownload')
                        refetch()
                    }}
                />
                <WeblensButton
                    label="Can Edit"
                    fillWidth
                    flavor={
                        share.permissions[user.username].canEdit
                            ? 'default'
                            : 'outline'
                    }
                    onClick={async () => {
                        await toggleParam(user.username, share, 'canEdit')
                        refetch()
                    }}
                />
                <WeblensButton
                    label="Can Delete"
                    fillWidth
                    flavor={
                        share.permissions[user.username].canDelete
                            ? 'default'
                            : 'outline'
                    }
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
        </div>
    )
}

export function ShareModal({
    targetFile,
    ref,
}: {
    targetFile: WeblensFile
    ref: Ref<HTMLDivElement>
}) {
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    const [focusedUser, setFocusedUser] = useState<UserInfo>(null)

    if (!targetFile) {
        targetFile = folderInfo
    }

    const {
        data: share,
        refetch,
        isLoading,
    } = useQuery<WeblensShare>({
        queryKey: ['share', targetFile.Id()],
        initialData: null,
        queryFn: async () => {
            const share = await targetFile.GetShare(true).catch(ErrorHandler)

            console.log('GOT SHREA', share)
            if (!share) {
                return null
            }

            return share
        },
    })

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

    return (
        <div
            className="fullscreen-modal"
            onClick={(e) => e.stopPropagation()}
            ref={ref}
        >
            {isLoading && (
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
                            await share.setPublic(!share.public)
                            refetch()
                        }}
                    />
                    <WeblensButton
                        label={'Copy Link'}
                        Left={IconLink}
                        fillWidth
                        disabled={!share?.shareId}
                        tooltip={!share?.shareId ? 'Not shared' : ''}
                    />
                </div>

                <UserSearch
                    accessors={share?.accessors}
                    addUser={async (user: UserInfo) => {
                        await share.addAccessor(user.username)
                        refetch()
                    }}
                />

                <h4 className="mb-2 w-max">Shared With</h4>

                <div className="no-scrollbar flex h-full w-full rounded-sm border p-2">
                    {share && share.accessors.length === 0 && (
                        <span className="m-auto">Not Shared</span>
                    )}
                    {share &&
                        share.accessors.length !== 0 &&
                        share.accessors.map((u: UserInfo) => {
                            return (
                                <div
                                    key={u.username}
                                    className="bg-background-secondary hover:bg-card-background-hover group/user flex h-10 w-full cursor-pointer items-center rounded p-2 transition"
                                    onClick={() => {
                                        if (focusedUser === u) {
                                            setFocusedUser(null)
                                        } else {
                                            setFocusedUser(u)
                                        }
                                    }}
                                >
                                    <span>{u.fullName}</span>
                                    <span className="text-color-text-secondary ml-1">
                                        [{u.username}]
                                    </span>
                                    <IconChevronRight className="text-text-tertiary ml-auto" />
                                </div>
                            )
                        })}
                    {focusedUser && (
                        <UserPermissions
                            user={focusedUser}
                            share={share}
                            refetch={refetch}
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
