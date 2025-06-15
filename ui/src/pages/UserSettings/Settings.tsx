import {
    IconBrush,
    IconClipboard,
    IconExternalLink,
    IconKey,
    IconLock,
    IconLockOpen,
    IconLogout,
    IconServer,
    IconTerminal2,
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
import MediaApi from '@weblens/api/MediaApi'
import { TowersApi } from '@weblens/api/ServersApi.js'
import UsersApi from '@weblens/api/UserApi'
import {
    HandleWebsocketMessage,
    WsAction,
    WsSubscriptionType,
    useWeblensSocket,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import { TokenInfo, TowerInfo } from '@weblens/api/swag/api.js'
import HeaderBar, { ThemeToggleButton } from '@weblens/components/HeaderBar'
import WeblensLoader from '@weblens/components/Loading.tsx'
import RemoteStatus from '@weblens/components/RemoteStatus.tsx'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import WeblensInput from '@weblens/lib/WeblensInput.tsx'
import { useKeyDown } from '@weblens/lib/hooks'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { useMessagesController } from '@weblens/store/MessagesController'
import { ErrorHandler } from '@weblens/types/Types'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import User, { UserPermissions } from '@weblens/types/user/User'
import { FC, useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { BackupProgressT } from '../Backup/BackupLogic.js'
import { historyDate } from '../FileBrowser/FileBrowserLogic.js'
import { SettingsWebsocketHandler } from './SettingsLogic.js'
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
    {
        id: 'dev',
        name: 'Developer',
        icon: IconTerminal2,
        pageComp: DeveloperTab,
    },
]

function TabGroup({
    tabs,
    activeTab,
    setActiveTab,
}: {
    tabs: settingsTab[]
    activeTab: string
    setActiveTab: (tab: string) => void
}) {
    return (
        <div>
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
                            <span className={settingsStyle.tabIcon}>
                                <tab.icon size={20} />
                            </span>
                            <span>{tab.name}</span>
                        </a>
                    </li>
                )
            })}
        </div>
    )
}

export default function Settings() {
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
        if (user.permissionLevel >= UserPermissions.Admin) {
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
        <div className={settingsStyle.settingsMenu}>
            <HeaderBar />
            <div className="flex grow flex-col p-8">
                <div className="border-b-color-border-primary mb-2 flex h-max w-full items-center gap-2 border-b-2 pb-2">
                    <IconUser size={25} />
                    <h3>{user.username}</h3>
                </div>
                <div className="flex grow">
                    <div className={settingsStyle.sidebar}>
                        <ul className="flex h-full flex-col">
                            <TabGroup
                                tabs={tabs}
                                activeTab={activeTab}
                                setActiveTab={setActiveTab}
                            />
                            {user.permissionLevel >= UserPermissions.Admin && (
                                <p className={settingsStyle.settingsTabsGroup}>
                                    Admin
                                </p>
                            )}
                            {user.permissionLevel >= UserPermissions.Admin && (
                                <TabGroup
                                    tabs={adminTabs}
                                    activeTab={activeTab}
                                    setActiveTab={setActiveTab}
                                />
                            )}
                            <li className="mt-auto">
                                <WeblensButton
                                    label={'Logout'}
                                    Left={IconLogout}
                                    danger
                                    centerContent
                                    squareSize={32}
                                    onClick={async () => {
                                        useMediaStore.getState().clear()
                                        useFileBrowserStore.getState().reset()
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
    const user = useSessionStore((state) => state.user)
    const setUser = useSessionStore((state) => state.setUser)
    const [userFullName, setUserFullName] = useState(user.fullName)

    return (
        <div className="flex flex-col gap-2">
            <h3>Account</h3>
            <h4>Username: {user.username}</h4>
            <h4>Full Name: {user.fullName}</h4>
            <div className="flex w-64 items-center gap-2">
                <WeblensInput
                    squareSize={50}
                    value={userFullName}
                    valueCallback={(val) => setUserFullName(val)}
                    placeholder="New Full Name"
                    valid={userFullName === '' ? false : null}
                    autoComplete="name"
                />
                <WeblensButton
                    squareSize={50}
                    centerContent
                    allowShrink={false}
                    label="Update"
                    disabled={
                        userFullName === '' || userFullName === user.fullName
                    }
                    onClick={() =>
                        UsersApi.changeDisplayName(user.username, userFullName)
                            .then(() => {
                                user.fullName = userFullName
                                setUser(user)
                                return true
                            })
                            .catch(ErrorHandler)
                    }
                />
            </div>
        </div>
    )
}

function AppearanceTab() {
    return (
        <div className="flex flex-col gap-2">
            <p className="w-max p-2 text-lg font-semibold text-nowrap">
                Appearance
            </p>
            <ThemeToggleButton />
            <p className="text-color-text-secondary">
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
    const [buttonRef, setButtonRef] = useState<HTMLButtonElement>()

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
    } = useQuery<TokenInfo[]>({
        queryKey: ['apiKeys'],
        initialData: [],
        queryFn: () => AccessApi.getApiKeys().then((res) => res.data),
        retry: false,
    })

    const { data: remotes, refetch: refetchRemotes } = useQuery<TowerInfo[]>({
        queryKey: ['remotes'],
        initialData: [],
        queryFn: async () =>
            (await TowersApi.getRemotes().then((res) => res.data)) || [],
        retry: false,
    })

    useKeyDown('Enter', () => {
        if (buttonRef) {
            buttonRef.click()
        }
    })

    return (
        <div className="relative flex w-full flex-col gap-2">
            <div className={settingsStyle.settingsSection}>
                {namingKey && (
                    <div className="absolute z-10 flex h-full w-full scale-105 rounded-sm backdrop-blur-xs">
                        <div className="relative m-auto h-10 w-32">
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
                <div className={settingsStyle.settingsHeader}>
                    <h3>API Keys</h3>
                    <WeblensButton
                        squareSize={32}
                        label="New Api Key"
                        centerContent
                        onClick={() => {
                            setNamingKey(true)
                        }}
                    />
                </div>
                <div className="flex w-full flex-col">
                    {!isLoading &&
                        keys?.map((val) => {
                            return (
                                <ApiKeyRow
                                    key={val.id}
                                    tokenInfo={val}
                                    refetch={() => {
                                        refetchRemotes().catch(ErrorHandler)
                                        refetchKeys().catch(ErrorHandler)
                                    }}
                                    remotes={remotes}
                                />
                            )
                        })}
                </div>
                {!isLoading && !keys && (
                    <p className="text-color-text-primary w-full text-center">
                        You have no API keys
                    </p>
                )}
                {isLoading && <WeblensLoader />}
            </div>

            <div className={settingsStyle.settingsHeader}>
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
                centerContent
                showSuccess
                disabled={oldP == '' || newP == '' || oldP === newP}
                onClick={updatePass}
                setButtonRef={setButtonRef}
            />
        </div>
    )
}

function ApiKeyRow({
    tokenInfo,
    refetch,
    remotes,
}: {
    tokenInfo: TokenInfo
    refetch: () => void
    remotes: TowerInfo[]
}) {
    const stars = '*'.repeat(tokenInfo.token.slice(4).length)

    return (
        <div
            key={tokenInfo.id}
            className="border-color-border-primary relative -my-[1px] flex h-max w-full flex-row items-center border-2 p-4 first:rounded-t-md last:rounded-b-md"
        >
            <div className="mx-6 flex flex-col items-center">
                <IconKey size={50} />
                {tokenInfo.remoteUsing !== '' && (
                    <span className="text-color-text-primary select-none">
                        Linked to{' '}
                        {
                            remotes.find((r) => r.id === tokenInfo.remoteUsing)
                                ?.name
                        }
                    </span>
                )}
                {tokenInfo.remoteUsing === '' && (
                    <span className="text-color-text-primary select-none">
                        Personal
                    </span>
                )}
            </div>
            <div className="flex w-max flex-col">
                <h4 className="theme-text w-full truncate text-nowrap">
                    {tokenInfo.nickname}
                </h4>
                <code className="theme-text mb-2 w-full truncate text-nowrap select-none">
                    {tokenInfo.token.slice(0, 4)}
                    {stars}
                </code>
                <span className="text-color-text-secondary">
                    Added on {historyDate(tokenInfo.createdTime, true)}
                </span>
                {tokenInfo.lastUsed === 0 && (
                    <span className="text-color-text-secondary select-none">
                        Never Used
                    </span>
                )}
                {tokenInfo.lastUsed !== 0 && (
                    <span className="text-color-text-secondary select-none">
                        {historyDate(tokenInfo.lastUsed)}
                    </span>
                )}
            </div>
            <div className="ml-auto flex gap-1">
                <WeblensButton
                    Left={IconClipboard}
                    tooltip="Copy Key"
                    onClick={async () => {
                        if (!window.isSecureContext) {
                            return
                        }
                        await navigator.clipboard.writeText(
                            String(tokenInfo.token)
                        )
                        return true
                    }}
                />
                <WeblensButton
                    Left={IconTrash}
                    danger
                    requireConfirm
                    tooltip="Delete Key"
                    onClick={() => {
                        AccessApi.deleteApiKey(tokenInfo.id)
                            .then(() => refetch())
                            .catch(ErrorHandler)
                    }}
                />
            </div>
        </div>
    )
}

function ServersTab() {
    const readyState = useWebsocketStore((state) => state.readyState)
    const wsSend = useWebsocketStore((state) => state.wsSend)

    const { data: remotes, refetch: refetchRemotes } = useQuery<TowerInfo[]>({
        queryKey: ['remotes'],
        initialData: [],
        queryFn: async () =>
            (await TowersApi.getRemotes().then((res) => res.data)) || [],
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

        wsSend({
            action: WsAction.Subscribe,
            subscriptionType: WsSubscriptionType.TaskType,
            content: { taskType: 'do_backup' },
        })

        return () =>
            wsSend({
                action: WsAction.Unsubscribe,
                content: { taskType: 'do_backup' },
            })
    }, [readyState])

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            SettingsWebsocketHandler(setBackupProgress, () => {
                refetchRemotes().catch(ErrorHandler)
            })
        )
    }, [lastMessage])

    return (
        <div className="flex w-full flex-col items-center gap-2 overflow-scroll rounded-sm p-2">
            {remotes.length === 0 && (
                <div className="mt-16 flex flex-col items-center">
                    <IconServer />
                    <h2 className="text-center">No remote servers</h2>
                    <div className="text-color-text-primary mt-2 flex flex-row items-center">
                        <span>After you set up a {''}</span>
                        <a
                            href="https://github.com/ethanrous/weblens?tab=readme-ov-file#weblens-backup"
                            target="_blank"
                            className="ml-1 inline-flex items-center"
                            data-subtle
                            rel="noreferrer"
                        >
                            backup server
                            <IconExternalLink size={18} />
                        </a>
                        <span>, it will appear here</span>
                    </div>
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
    const {
        data: allUsersInfo,
        refetch: refetchUsers,
        isLoading,
    } = useQuery<User[]>({
        queryKey: ['users'],
        initialData: [],
        queryFn: async () => {
            const users = await UsersApi.getUsers().then((res) =>
                res.data.map((info) => new User(info))
            )
            users.sort((a, b) => a.username.localeCompare(b.username))
            return users
        },
    })

    return (
        <div className="flex h-full w-full flex-col">
            <h3>Users</h3>
            <div className="mt-3 flex h-full w-full flex-row gap-4">
                <div className="border-color-border-primary mr-2 h-full w-1/2 border-r pr-6">
                    {isLoading && <WeblensLoader />}
                    {!isLoading && (
                        <UsersTable
                            users={allUsersInfo}
                            refetchUsers={refetchUsers}
                        />
                    )}
                </div>
                <div>
                    <CreateUserBox refetchUsers={refetchUsers} />
                </div>
            </div>
        </div>
    )
}

function UsersTable({
    users,
    refetchUsers,
}: {
    users: User[]
    refetchUsers: () => Promise<QueryObserverResult<User[], Error>>
}) {
    const user = useSessionStore((state) => state.user)

    return (
        <table className="text-color-text-primary h-max w-full caption-bottom border-collapse align-top">
            <thead className="border-color-border-primary table-header-group h-9 border-b align-top">
                <tr>
                    <th className="text-left text-lg">Full Name</th>
                    <th className="text-left text-lg">Username</th>
                    <th className="text-left text-lg">Role</th>
                    <th className="text-right text-lg">Actions</th>
                </tr>
            </thead>
            <tbody className="h-max">
                {!users
                    ? null
                    : users.map((rowUser) => (
                          <tr
                              key={rowUser.username}
                              className="hover:bg-background-secondary h-12"
                          >
                              <td className="p-2">{rowUser.fullName}</td>
                              <td>{rowUser.username}</td>
                              <td>
                                  {UserPermissions[rowUser.permissionLevel]}
                              </td>
                              <td className="p-2">
                                  <UserRowActions
                                      rowUser={rowUser}
                                      accessor={user}
                                      refetchUsers={refetchUsers}
                                  />
                              </td>
                          </tr>
                      ))}
            </tbody>
        </table>
    )
}

function UserRowActions({
    rowUser,
    accessor,
    refetchUsers,
}: {
    rowUser: User
    accessor: User
    refetchUsers: () => Promise<QueryObserverResult<User[], Error>>
}) {
    const [changingPass, setChangingPass] = useState(false)

    return (
        <div className="flex items-center justify-end gap-2">
            {rowUser.activated === false && (
                <WeblensButton
                    label="Activate"
                    size="small"
                    onClick={() => {
                        UsersApi.activateUser(rowUser.username, true)
                            .then(() => refetchUsers())
                            .catch(ErrorHandler)
                    }}
                />
            )}
            {!changingPass &&
                accessor.permissionLevel >= UserPermissions.Owner && (
                    <WeblensButton
                        tooltip="Change Password"
                        size="small"
                        Left={IconLockOpen}
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
                    className="max-w-[8vw]"
                    closeInput={() => setChangingPass(false)}
                    onComplete={async (newPass) => {
                        if (newPass === '') {
                            return Promise.reject(
                                new Error(
                                    'Cannot update password to empty string'
                                )
                            )
                        }
                        return UsersApi.updateUserPassword(rowUser.username, {
                            newPassword: newPass,
                        })
                    }}
                />
            )}
            {rowUser.permissionLevel < UserPermissions.Admin &&
                accessor.permissionLevel >= UserPermissions.Owner && (
                    <WeblensButton
                        tooltip="Make Admin"
                        Left={IconUserUp}
                        size="small"
                        labelOnHover={true}
                        onClick={() => {
                            UsersApi.setUserAdmin(rowUser.username, true)
                                .then(() => refetchUsers())
                                .catch(ErrorHandler)
                        }}
                    />
                )}
            {rowUser.permissionLevel < UserPermissions.Owner &&
                rowUser.permissionLevel >= UserPermissions.Admin &&
                accessor.permissionLevel >= UserPermissions.Owner && (
                    <WeblensButton
                        tooltip="Remove Admin"
                        Left={IconUserMinus}
                        size="small"
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
                size="small"
                danger
                requireConfirm
                disabled={
                    (rowUser.permissionLevel >= UserPermissions.Admin &&
                        accessor.permissionLevel < UserPermissions.Owner) ||
                    accessor.username === rowUser.username
                }
                onClick={() =>
                    UsersApi.deleteUser(rowUser.username).then(() =>
                        refetchUsers()
                    )
                }
            />
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
        <div className="mt-auto flex h-max w-full flex-col gap-2 p-2">
            <div className="flex w-full flex-col gap-2">
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
                    <div className="flex w-max grow flex-row">
                        <WeblensButton
                            Left={IconUserShield}
                            tooltip="Admin"
                            className="mx-2"
                            flavor={makeAdmin ? 'default' : 'outline'}
                            onClick={() => setMakeAdmin(!makeAdmin)}
                        />
                        <WeblensButton
                            label="Create User"
                            squareSize={50}
                            disabled={userInput === '' || passInput === ''}
                            onClick={() =>
                                UsersApi.createUser({
                                    fullName: '',
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

function DeveloperTab() {
    return (
        <div className="flex w-max flex-col gap-2">
            <span className="text-danger">Here be dragons</span>
            <WeblensButton
                label="Clear Media"
                danger
                onClick={() =>
                    MediaApi.dropMedia().then(() =>
                        useMessagesController.getState().addMessage({
                            text: 'Media Cleared',
                            duration: 5000,
                            severity: 'success',
                        })
                    )
                }
            />
            <WeblensButton
                label="Reset Server"
                danger
                onClick={() => {
                    TowersApi.resetTower().then(() => window.location.reload())
                }}
            />
        </div>
    )
}
