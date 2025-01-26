import {
    IconBrush,
    IconClipboard,
    IconExternalLink,
    IconLock,
    IconLockOpen,
    IconLogout,
    IconServer,
    IconTrash,
    IconUser,
    IconUserMinus,
    IconUserShield,
    IconUserUp,
    IconUsers,
} from '@tabler/icons-react'
import {
    QueryObserverResult,
    RefetchOptions,
    useQuery,
} from '@tanstack/react-query'
import AccessApi from '@weblens/api/AccessApi'
import { ServersApi } from '@weblens/api/ServersApi'
import UsersApi from '@weblens/api/UserApi'
import {
    HandleWebsocketMessage,
    useWeblensSocket,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import { ApiKeyInfo, ServerInfo } from '@weblens/api/swag'
import HeaderBar, { ThemeToggleButton } from '@weblens/components/HeaderBar'
import WeblensLoader from '@weblens/components/Loading'
import RemoteStatus from '@weblens/components/RemoteStatus'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useKeyDown } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { ErrorHandler } from '@weblens/types/Types'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import User from '@weblens/types/user/User'
import { FC, useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { AdminWebsocketHandler } from '../AdminSettings/adminLogic'
import { BackupProgressT } from '../Backup/BackupLogic'
import { historyDate } from '../FileBrowser/FileBrowserLogic'
import settingsStyle from './settingsStyle.module.scss'

type settingsTab = {
    id: string
    name: string
    icon: FC<{ size: number }>
    pageComp: FC
}

const tabs: settingsTab[] = [
    {
        id: 'account',
        name: 'Account',
        icon: IconUser,
        pageComp: AccountTab,
    },
    {
        id: 'security',
        name: 'Securty',
        icon: IconLock,
        pageComp: SecurityTab,
    },
    {
        id: 'appearance',
        name: 'Appearance',
        icon: IconBrush,
        pageComp: AppearanceTab,
    },
]

const adminTabs: settingsTab[] = [
    {
        id: 'servers',
        name: 'Servers',
        icon: IconServer,
        pageComp: ServersTab,
    },
    {
        id: 'users',
        name: 'Users',
        icon: IconUsers,
        pageComp: UsersTab,
    },
]

export function SettingsMenu() {
    const user = useSessionStore((state) => state.user)
    const setUser = useSessionStore((state) => state.setUser)
    const nav = useNavigate()

    const [activeTab, setActiveTab] = useState(
        window.location.pathname.replace('/settings/', '')
    )

    useWeblensSocket()

    useEffect(() => {
        const desiredLoc = `/settings/${activeTab}`
        if (window.location.pathname !== desiredLoc) {
            nav(desiredLoc)
        }
    }, [activeTab])

    const ActivePage = useMemo(() => {
        const allTabs = [...tabs]
        if (user.admin) {
            allTabs.push(...adminTabs)
        }

        const ActivePage = allTabs.find((val) => val.id === activeTab)?.pageComp
        return ActivePage
    }, [activeTab])

    useEffect(() => {
        if (!user.isLoggedIn) {
            nav('/login', { state: { returnTo: window.location.pathname } })
            return
        }
        if (!ActivePage) {
            setActiveTab('account')
            nav('/settings/account')
        }
    }, [user])

    return (
        <div className={settingsStyle['settings-menu']}>
            <HeaderBar />
            <div className="flex flex-col grow p-8">
                <div className="flex h-max items-center gap-2 w-full mb-2 pb-2 border-b-[--wl-outline-subtle] border-b">
                    <IconUser size={25} />
                    <h3>{user.username}</h3>
                </div>
                <div className="flex grow">
                    <div className={settingsStyle['sidebar']}>
                        <ul className="flex flex-col h-full">
                            {tabs.map((tab) => {
                                return (
                                    <li key={tab.id}>
                                        <a
                                            data-active={activeTab === tab.id}
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                e.preventDefault()
                                                setActiveTab(tab.id)
                                            }}
                                        >
                                            <span
                                                className={
                                                    settingsStyle['tab-icon']
                                                }
                                            >
                                                <tab.icon size={20} />
                                            </span>
                                            <span>{tab.name}</span>
                                        </a>
                                    </li>
                                )
                            })}
                            {user.admin && (
                                <p
                                    className={
                                        settingsStyle['settings-tabs-group']
                                    }
                                >
                                    Admin
                                </p>
                            )}
                            {user.admin &&
                                adminTabs.map((tab) => {
                                    return (
                                        <li key={tab.id}>
                                            <a
                                                data-active={
                                                    activeTab === tab.id
                                                }
                                                onClick={(e) => {
                                                    e.stopPropagation()
                                                    e.preventDefault()
                                                    setActiveTab(tab.id)
                                                }}
                                            >
                                                <span
                                                    className={
                                                        settingsStyle[
                                                            'tab-icon'
                                                        ]
                                                    }
                                                >
                                                    <tab.icon size={16} />
                                                </span>
                                                <span>{tab.name}</span>
                                            </a>
                                        </li>
                                    )
                                })}
                            <li className="mt-auto">
                                <WeblensButton
                                    label={'Logout'}
                                    Left={IconLogout}
                                    danger
                                    centerContent
                                    squareSize={32}
                                    onClick={async () => {
                                        useMediaStore.getState().clear()
                                        await UsersApi.logoutUser()
                                        const loggedOut = new User()
                                        loggedOut.isLoggedIn = false
                                        setUser(loggedOut)
                                        nav('/login')
                                    }}
                                />
                            </li>
                        </ul>
                    </div>
                    {ActivePage && <ActivePage />}
                </div>
            </div>
        </div>
    )
}

function AccountTab() {
    return (
        <div className="flex flex-col gap-2">
            <p className="text-lg font-semibold p-2 w-max text-nowrap">
                Account
            </p>
        </div>
    )
}

function AppearanceTab() {
    return (
        <div className="flex flex-col gap-2">
            <p className="text-lg font-semibold p-2 w-max text-nowrap">
                Appearance
            </p>
            <ThemeToggleButton />
            <p className="text-[--wl-text-color-dull]">
                Hint: Press [T] to toggle theme
            </p>
        </div>
    )
}

function SecurityTab() {
    const user = useSessionStore((state) => state.user)
    const [oldP, setOldP] = useState('')
    const [newP, setNewP] = useState('')
    const [namingKey, setNamingKey] = useState(false)
    const [buttonRef, setButtonRef] = useState<HTMLDivElement>()

    const updatePass = useCallback(async () => {
        if (oldP == '' || newP == '' || oldP === newP) {
            return Promise.reject(
                new Error('Old and new password cannot be empty or match')
            )
        }
        return UsersApi.updateUserPassword(user.username, {
            oldPassword: oldP,
            newPassword: newP,
        }).then(() => {
            setNewP('')
            setOldP('')
        })
    }, [user.username, String(oldP), String(newP)])

    const {
        data: keys,
        refetch: refetchKeys,
        isLoading,
    } = useQuery<ApiKeyInfo[]>({
        queryKey: ['apiKeys'],
        initialData: [],
        queryFn: () => AccessApi.getApiKeys().then((res) => res.data),
        retry: false,
    })

    const { data: remotes, refetch: refetchRemotes } = useQuery<ServerInfo[]>({
        queryKey: ['remotes'],
        initialData: [],
        queryFn: async () =>
            (await ServersApi.getRemotes().then((res) => res.data)) || [],
        retry: false,
    })

    useKeyDown('Enter', () => {
        if (buttonRef) {
            buttonRef.click()
        }
    })

    return (
        <div className="flex flex-col w-full gap-2 relative">
            <div className={settingsStyle['settings-section']}>
                {namingKey && (
                    <div className="absolute flex w-full h-full z-10 backdrop-blur-sm rounded scale-105">
                        <div className="relative w-32 h-10 m-auto">
                            <WeblensInput
                                placeholder="New Key Name"
                                autoFocus
                                closeInput={() => setNamingKey(false)}
                                onComplete={async (v) => {
                                    await AccessApi.createApiKey({
                                        name: v,
                                    }).then(() => refetchKeys())
                                    await refetchKeys()
                                    return
                                }}
                            />
                        </div>
                    </div>
                )}
                <div className={settingsStyle['settings-header']}>
                    <h3>API Keys</h3>
                    <WeblensButton
                        squareSize={32}
                        label="New Api Key"
                        onClick={() => {
                            setNamingKey(true)
                        }}
                    />
                </div>
                {!isLoading &&
                    keys?.map((val) => {
                        return (
                            <ApiKeyRow
                                key={val.id}
                                keyInfo={val}
                                refetch={() => {
                                    refetchRemotes().catch(ErrorHandler)
                                    refetchKeys().catch(ErrorHandler)
                                }}
                                remotes={remotes}
                            />
                        )
                    })}
                {!isLoading && !keys && (
                    <p className="w-full text-center text-[#cccccc]">
                        You have no API keys
                    </p>
                )}
                {isLoading && <WeblensLoader />}
            </div>

            <div className={settingsStyle['settings-header']}>
                <h3>Change Password</h3>
            </div>
            <WeblensInput
                value={oldP}
                placeholder="Old Password"
                password
                valueCallback={setOldP}
                squareSize={50}
            />
            <WeblensInput
                value={newP}
                placeholder="New Password"
                valueCallback={setNewP}
                squareSize={50}
                password
            />
            <div className="p2" />
            <WeblensButton
                label="Update Password"
                squareSize={40}
                fillWidth
                showSuccess
                disabled={oldP == '' || newP == '' || oldP === newP}
                onClick={updatePass}
                setButtonRef={setButtonRef}
            />
        </div>
    )
}

function ApiKeyRow({
    keyInfo,
    refetch,
    remotes,
}: {
    keyInfo: ApiKeyInfo
    refetch: () => void
    remotes: ServerInfo[]
}) {
    const stars = '*'.repeat(keyInfo.key.slice(4).length)

    return (
        <div key={keyInfo.id} className={settingsStyle['settings-content-row']}>
            <div className="flex flex-col grow w-1/2">
                <strong className="theme-text font-bold text-nowrap w-full truncate select-none my-1">
                    {keyInfo.name}
                </strong>
                <code className="theme-text text-nowrap w-full truncate select-none text-[12px]">
                    {keyInfo.key.slice(0, 4)}
                    {stars}
                </code>
                <p className="text-[--wl-text-color-dull]">
                    Added {historyDate(keyInfo.createdTime)}
                </p>
                {keyInfo.lastUsedTime === 0 && (
                    <p className="select-none text-[--wl-text-color-dull]">Unused</p>
                )}
                {keyInfo.lastUsedTime !== 0 && (
                    <p className="select-none text-[--wl-text-color-dull]">
                        {historyDate(keyInfo.lastUsedTime)}
                    </p>
                )}
                {keyInfo.remoteUsing !== '' && (
                    <p className="select-none text-[--wl-text-color-dull]">
                        Linked to{' '}
                        {
                            remotes.find((r) => r.id === keyInfo.remoteUsing)
                                ?.name
                        }
                    </p>
                )}
                {keyInfo.remoteUsing === '' && (
                    <p className="select-none text-[--wl-text-color-dull]">
                        Not Linked
                    </p>
                )}
            </div>
            <WeblensButton
                Left={IconClipboard}
                tooltip="Copy Key"
                onClick={async () => {
                    if (!window.isSecureContext) {
                        return
                    }
                    await navigator.clipboard.writeText(keyInfo.key)
                    return true
                }}
            />
            <WeblensButton
                Left={IconTrash}
                danger
                requireConfirm
                tooltip="Delete Key"
                onClick={() => {
                    AccessApi.deleteApiKey(keyInfo.key)
                        .then(() => refetch())
                        .catch(ErrorHandler)
                }}
            />
        </div>
    )
}

function ServersTab() {
    const readyState = useWebsocketStore((state) => state.readyState)
    const wsSend = useWebsocketStore((state) => state.wsSend)

    const { data: remotes, refetch: refetchRemotes } = useQuery<ServerInfo[]>({
        queryKey: ['remotes'],
        initialData: [],
        queryFn: async () =>
            (await ServersApi.getRemotes().then((res) => res.data)) || [],
        retry: false,
    })

    const lastMessage = useWebsocketStore((state) => state.lastMessage)

    const [backupProgress, setBackupProgress] = useState<
        Map<string, BackupProgressT>
    >(new Map())

    useEffect(() => {
        if (!readyState) {
            return
        }

        wsSend('taskSubscribe', { taskType: 'do_backup' })
        return () => wsSend('unsubscribe', { taskType: 'do_backup' })
    }, [readyState])

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            AdminWebsocketHandler(setBackupProgress, () => {
                refetchRemotes().catch(ErrorHandler)
            })
        )
    }, [lastMessage])

    return (
        <div className="flex flex-col items-center p-2 rounded w-full gap-2 overflow-scroll">
            {remotes.length === 0 && (
                <div className="flex flex-col items-center mt-16">
                    <IconServer />
                    <h2 className="text-center">No remote servers</h2>
                    <p className="text-sm mt-2 text-[--wl-text-color-dull]">
                        After you set up a {''}
                        <a
                            href="https://github.com/ethanrous/weblens?tab=readme-ov-file#weblens-backup"
                            target="_blank"
                            className="p-1"
                            data-subtle
                        >
                            backup server
                            <IconExternalLink size={18} />
                        </a>
                        , it will appear here
                    </p>
                </div>
            )}
            {remotes.map((r) => {
                return (
                    <RemoteStatus
                        key={r.id}
                        remoteInfo={r}
                        refetchRemotes={() => {
                            refetchRemotes().catch(ErrorHandler)
                        }}
                        restoreProgress={null}
                        backupProgress={backupProgress.get(r.id)}
                        setBackupProgress={(progress) => {
                            setBackupProgress((old) => {
                                const newMap = new Map(old)
                                newMap.set(r.id, progress)
                                return newMap
                            })
                        }}
                    />
                )
            })}
        </div>
    )
}

function UsersTab() {
    const user = useSessionStore((state) => state.user)
    const { data: allUsersInfo, refetch: refetchUsers } = useQuery<User[]>({
        queryKey: ['users'],
        initialData: [],
        queryFn: () =>
            UsersApi.getUsers().then((res) =>
                res.data.map((info) => new User(info))
            ),
    })
    const usersList = useMemo(() => {
        if (!allUsersInfo) {
            return null
        }
        allUsersInfo.sort((a, b) => {
            return a.username.localeCompare(b.username)
        })

        return allUsersInfo.map((val) => (
            <UserRow
                key={val.username}
                rowUser={val}
                accessor={user}
                refetchUsers={refetchUsers}
            />
        ))
    }, [allUsersInfo])

    return (
        <div className="flex flex-col w-full">
            <div className="flex flex-col p-2 w-full h-0 grow shrink overflow-scroll">
                <div className="grid h-max">{usersList}</div>
            </div>
            <CreateUserBox refetchUsers={refetchUsers} />
        </div>
    )
}

const UserRow = ({
    rowUser,
    accessor,
    refetchUsers,
}: {
    rowUser: User
    accessor: User
    refetchUsers: (
        opts?: RefetchOptions
    ) => Promise<QueryObserverResult<User[], Error>>
}) => {
    const [changingPass, setChangingPass] = useState(false)

    let userLevel = ''
    if (rowUser.owner) {
        userLevel = 'Owner'
    } else if (rowUser.admin) {
        userLevel = 'Admin'
    }

    return (
        <div
            key={rowUser.username}
            className={settingsStyle['settings-content-row']}
        >
            <div className="flex flex-col justify-center w-max h-max">
                <p className="font-bold w-max theme-text">{rowUser.username}</p>
                <p className="theme-text">{userLevel}</p>
            </div>
            <div className="flex">
                {rowUser.activated === false && (
                    <WeblensButton
                        label="Activate"
                        squareSize={35}
                        onClick={() => {
                            UsersApi.activateUser(rowUser.username, true)
                                .then(() => refetchUsers())
                                .catch(ErrorHandler)
                        }}
                    />
                )}
                {!changingPass && accessor.owner && (
                    <WeblensButton
                        tooltip="Change Password"
                        labelOnHover={true}
                        Left={IconLockOpen}
                        squareSize={35}
                        disabled={!rowUser.activated}
                        onClick={() => {
                            setChangingPass(true)
                        }}
                    />
                )}
                {changingPass && (
                    <WeblensInput
                        placeholder="New Password"
                        autoFocus={true}
                        closeInput={() => setChangingPass(false)}
                        onComplete={async (newPass) => {
                            if (newPass === '') {
                                return Promise.reject(
                                    new Error(
                                        'Cannot update password to empty string'
                                    )
                                )
                            }
                            return UsersApi.updateUserPassword(
                                rowUser.username,
                                { newPassword: newPass }
                            )
                        }}
                    />
                )}
                {!rowUser.admin && accessor.owner && (
                    <WeblensButton
                        tooltip="Make Admin"
                        Left={IconUserUp}
                        labelOnHover={true}
                        allowShrink={false}
                        squareSize={35}
                        onClick={() => {
                            UsersApi.setUserAdmin(rowUser.username, true)
                                .then(() => refetchUsers())
                                .catch(ErrorHandler)
                        }}
                    />
                )}
                {!rowUser.owner && rowUser.admin && accessor.owner && (
                    <WeblensButton
                        tooltip="Remove Admin"
                        Left={IconUserMinus}
                        squareSize={35}
                        onClick={() => {
                            UsersApi.setUserAdmin(rowUser.username, false)
                                .then(() => refetchUsers())
                                .catch(ErrorHandler)
                        }}
                    />
                )}

                <WeblensButton
                    squareSize={35}
                    tooltip="Delete"
                    Left={IconTrash}
                    danger
                    requireConfirm
                    centerContent
                    disabled={rowUser.admin && !accessor.owner}
                    onClick={() =>
                        UsersApi.deleteUser(rowUser.username).then(() =>
                            refetchUsers()
                        )
                    }
                />
            </div>
        </div>
    )
}

function CreateUserBox({
    refetchUsers,
}: {
    refetchUsers: (
        opts?: RefetchOptions
    ) => Promise<QueryObserverResult<User[], Error>>
}) {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')
    const [makeAdmin, setMakeAdmin] = useState(false)
    return (
        <div className="flex flex-col p-2 h-max w-full gap-2">
            <div className="flex flex-col gap-2 w-full">
                <div className="flex gap-1">
                    <WeblensInput
                        placeholder="Username"
                        value={userInput}
                        squareSize={50}
                        onComplete={null}
                        fillWidth
                        valueCallback={setUserInput}
                    />
                    <WeblensInput
                        placeholder="Password"
                        value={passInput}
                        squareSize={50}
                        onComplete={null}
                        fillWidth
                        password
                        valueCallback={setPassInput}
                    />
                    <div className="flex flex-row grow w-max">
                        <WeblensButton
                            Left={IconUserShield}
                            tooltip="Admin"
                            allowRepeat
                            squareSize={50}
                            toggleOn={makeAdmin}
                            onClick={() => setMakeAdmin(!makeAdmin)}
                        />
                        <WeblensButton
                            label="Create User"
                            squareSize={50}
                            disabled={userInput === '' || passInput === ''}
                            onClick={() =>
                                UsersApi.createUser({
                                    admin: makeAdmin,
                                    autoActivate: true,
                                    username: userInput,
                                    password: passInput,
                                })
                                    .then(() => {
                                        setUserInput('')
                                        setPassInput('')
                                    })
                                    .then(() => refetchUsers())
                            }
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}
