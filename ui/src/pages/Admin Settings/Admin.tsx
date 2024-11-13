import {
    IconClipboard,
    IconLockOpen,
    IconPlus,
    IconTrash,
    IconUserMinus,
    IconUserShield,
    IconUserUp,
    IconX,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { clearCache, resetServer } from '@weblens/api/ApiFetch'
import UsersApi from '@weblens/api/UserApi'
import { useKeyDown } from '@weblens/components/hooks'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { useEffect, useMemo, useState } from 'react'
import RemoteStatus from '@weblens/components/RemoteStatus'
import './adminStyle.scss'
import {
    HandleWebsocketMessage,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import { AdminWebsocketHandler } from './adminLogic'
import { BackupProgressT } from '../Backup/BackupLogic'
import {
    TaskProgress,
    TaskStage,
    useTaskState,
} from '../FileBrowser/TaskStateControl'
import { ApiKeyInfo, ServerInfo } from '@weblens/api/swag'
import { ServersApi } from '@weblens/api/ServersApi'
import User from '@weblens/types/user/User'
import AccessApi from '@weblens/api/AccessApi'

// function PathAutocomplete() {
//     const [pathSearch, setPathSearch] = useState('')
//     const [hoverOffset, setHoverOffset] = useState(0)
//     const [bouncedSearch] = useDebouncedValue(pathSearch, 100)
//     const { data: names } = useQuery<FolderInfo>({
//         queryKey: ['pathAutocomplete', bouncedSearch],
//         initialData: { children: [], folder: null } as FolderInfo,
//         queryFn: () =>
//             FileApi.autocompletePath(bouncedSearch).then((res) => {
//                 const data = res.data
//                 data.children.sort((f1, f2) =>
//                     f1.filename.localeCompare(f2.filename)
//                 )
//                 return data
//             }),
//         retry: false,
//     })
//
//     useEffect(() => {
//         setHoverOffset(0)
//     }, [pathSearch])
//
//     const setBlockFocus = useFileBrowserStore((state) => state.setBlockFocus)
//
//     useKeyDown('Tab', (e) => {
//         e.preventDefault()
//         const tabFile = names.children[names.children.length - hoverOffset - 1]
//
//         if (!tabFile) {
//             return
//         }
//
//         setPathSearch((s) => {
//             if (!s.endsWith('/')) {
//                 s += '/'
//             }
//             return (
//                 s.slice(0, s.lastIndexOf('/')) +
//                 '/' +
//                 tabFile.filename +
//                 (tabFile.isDir ? '/' : '')
//             )
//         })
//     })
//
//     useKeyDown('ArrowUp', (e) => {
//         e.preventDefault()
//         e.stopPropagation()
//         setHoverOffset((offset) => Math.min(offset + 1, names.children.length))
//     })
//
//     useKeyDown('ArrowDown', (e) => {
//         e.preventDefault()
//         e.stopPropagation()
//         setHoverOffset((offset) => Math.max(offset - 1, 0))
//     })
//
//     const result = useMemo(() => {
//         const lastSlash = pathSearch.lastIndexOf('/')
//         if ((lastSlash === -1 || names.children.length === 0) && names.self) {
//             return names.self
//         }
//
//         if (pathSearch.slice(lastSlash + 1) === '') {
//             return names.self
//         }
//
//         if (names.children.length !== 0) {
//             return names.children[0]
//         }
//
//         return null
//     }, [names, pathSearch])
//
//     return (
//         <div className="w-[50%] h-10 m-2">
//             <WeblensInput
//                 value={pathSearch}
//                 valueCallback={setPathSearch}
//                 fillWidth
//                 openInput={() => setBlockFocus(true)}
//                 closeInput={() => setBlockFocus(false)}
//                 failed={names.self === null && pathSearch !== ''}
//                 placeholder={'File Search'}
//             />
//             <div
//                 className="flex flex-col gap-1 absolute -translate-y-[100%] pb-12 pointer-events-none"
//                 style={{ paddingLeft: pathSearch.length * 9 }}
//             >
//                 {names.children.map((cn, i) => {
//                     if (pathSearch.endsWith(cn.filename)) {
//                         return null
//                     }
//                     return (
//                         <div
//                             key={cn.filename}
//                             className="flex justify-center items-center p-2 h-10 bg-[#1c1049] rounded pointer-events-auto cursor-pointer"
//                             style={{
//                                 backgroundColor:
//                                     i ===
//                                     names.children.length - hoverOffset - 1
//                                         ? '#2549ff'
//                                         : '#1c1049',
//                             }}
//                             onClick={() => {
//                                 setPathSearch(
//                                     (s) =>
//                                         s.slice(0, s.lastIndexOf('/')) +
//                                         '/' +
//                                         cn.filename +
//                                         (cn.isDir ? '/' : '')
//                                 )
//                             }}
//                         >
//                             <p>{cn.filename}</p>
//                         </div>
//                     )
//                 })}
//             </div>
//             {result && (
//                 <div className="flex flex-row gap-2">
//                     <p>{result.filename}</p>
//                     <p>{result.id}</p>
//                 </div>
//             )}
//         </div>
//     )
// }

function CreateUserBox({ refetchUsers }: { refetchUsers: () => void }) {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')
    const [makeAdmin, setMakeAdmin] = useState(false)
    return (
        <div className="theme-outline flex flex-col p-2 h-max w-full rounded gap-2">
            <p className="w-full h-max font-semibold text-xl select-none">
                Add User
            </p>
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
                                }).then(() => {
                                    refetchUsers()
                                    setUserInput('')
                                    setPassInput('')
                                })
                            }
                        />
                    </div>
                </div>
            </div>
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
    refetchUsers: () => void
}) => {
    const [changingPass, setChangingPass] = useState(false)
    return (
        <div key={rowUser.username} className="admin-user-row">
            <div className="flex flex-col justify-center w-max h-max">
                <p className="font-bold w-max theme-text">{rowUser.username}</p>
                {rowUser.admin && !rowUser.owner && (
                    <p className="theme-text">Admin</p>
                )}
                {rowUser.owner && <p className="theme-text">Owner</p>}
            </div>
            <div className="flex">
                {rowUser.activated === false && (
                    <WeblensButton
                        label="Activate"
                        squareSize={35}
                        onClick={() => {
                            UsersApi.activateUser(rowUser.username, true).then(
                                () => refetchUsers()
                            )
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
                                    'Cannot update password to empty string'
                                )
                            }
                            return UsersApi.updateUserPassword(
                                rowUser.username,
                                { newPassword: newPass }
                            ).then(() => true)
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
                            UsersApi.setUserAdmin(rowUser.username, true).then(
                                () => refetchUsers()
                            )
                        }}
                    />
                )}
                {!rowUser.owner && rowUser.admin && accessor.owner && (
                    <WeblensButton
                        tooltip="Remove Admin"
                        Left={IconUserMinus}
                        squareSize={35}
                        onClick={() => {
                            UsersApi.setUserAdmin(rowUser.username, false).then(
                                () => refetchUsers()
                            )
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

function UsersBox({
    thisUserInfo,
    allUsersInfo,
    refetchUsers,
}: {
    thisUserInfo: User
    allUsersInfo: User[]
    refetchUsers: () => void
}) {
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
                accessor={thisUserInfo}
                refetchUsers={refetchUsers}
            />
        ))
    }, [allUsersInfo])

    return (
        <div className="theme-outline flex flex-col p-2 shrink w-full h-max min-h-96 rounded gap-2 no-scrollbar">
            <h4 className="pl-1">Users</h4>
            <div className="flex flex-col grow relative gap-1 overflow-y-scroll overflow-x-visible max-h-[50vh]">
                {usersList}
            </div>
            <CreateUserBox refetchUsers={refetchUsers} />
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
    return (
        <div key={keyInfo.id} className="admin-user-row">
            <div className="flex flex-col grow w-1/2">
                <p className="theme-text font-bold text-nowrap w-full truncate select-none">
                    {keyInfo.key}
                </p>
                {keyInfo.remoteUsing !== '' && (
                    <p className="select-none">
                        Used by:{' '}
                        {
                            remotes.find((r) => r.id === keyInfo.remoteUsing)
                                ?.name
                        }
                    </p>
                )}
                {keyInfo.remoteUsing === '' && (
                    <p className="select-none">Unused</p>
                )}
            </div>
            <WeblensButton
                Left={IconClipboard}
                tooltip="Copy Key"
                onClick={() => {
                    if (!window.isSecureContext) {
                        return false
                    }
                    navigator.clipboard.writeText(keyInfo.key)
                    return true
                }}
            />
            <WeblensButton
                Left={IconTrash}
                danger
                requireConfirm
                tooltip="Delete Key"
                onClick={() => {
                    AccessApi.deleteApiKey(keyInfo.key).then(() => refetch())
                }}
            />
        </div>
    )
}

function Servers() {
    const server = useSessionStore((state) => state.server)

    const { data: keys, refetch: refetchKeys } = useQuery<ApiKeyInfo[]>({
        queryKey: ['apiKeys'],
        initialData: [],
        queryFn: () => AccessApi.getApiKeys().then((res) => res.data),
        retry: false,
    })

    const { data: remotes, refetch: refetchRemotes } = useQuery<ServerInfo[]>({
        queryKey: ['remotes'],
        initialData: [],
        queryFn: async () =>
            await ServersApi.getRemotes().then((res) => res.data),
        retry: false,
    })

    const lastMessage = useWebsocketStore((state) => state.lastMessage)

    const [backupProgress, setBackupProgress] = useState<
        Map<string, BackupProgressT>
    >(new Map())

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            AdminWebsocketHandler(setBackupProgress, refetchRemotes)
        )
    }, [lastMessage])

    useEffect(() => {
        refetchKeys()
    }, [remotes?.length])

    return (
        <div className="theme-outline flex flex-col rounded items-center p-1 w-full">
            <h4 className="pl-2 w-full">API Keys</h4>

            {Boolean(keys?.length) && (
                <div className="flex flex-col items-center p-1 rounded w-full gap-1">
                    {keys.map((k) => (
                        <ApiKeyRow
                            key={k.id}
                            keyInfo={k}
                            refetch={refetchKeys}
                            remotes={remotes}
                        />
                    ))}
                </div>
            )}
            <WeblensButton
                squareSize={40}
                label="New Api Key"
                Left={IconPlus}
                onClick={() => {
                    AccessApi.createApiKey().then(() => refetchKeys())
                }}
            />
            <div className="flex flex-row w-full items-center pr-4">
                <h4 className="pl-2">Remotes</h4>
            </div>

            <div className="flex flex-col items-center p-2 rounded w-full gap-2 overflow-scroll">
                {remotes.map((r) => {
                    if (r.id === server.id) {
                        return null
                    }

                    return (
                        <RemoteStatus
                            key={r.id}
                            remoteInfo={r}
                            refetchRemotes={refetchRemotes}
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
        </div>
    )
}

function BackupProgress() {
    const tasks = useTaskState((state) => state.tasks)
    const [backupTask, setBackupTask] = useState<TaskProgress>()
    const [backupTaskId, setBackupTaskId] = useState<string>()

    useEffect(() => {
        if (backupTaskId) {
            setBackupTask(tasks.get(backupTaskId))
        } else {
            const backupTasks = Array.from(tasks.values()).filter(
                (t) => t.taskType === 'do_backup'
            )
            if (backupTasks.length !== 0) {
                setBackupTaskId(backupTasks[0].GetTaskId())
                setBackupTask(backupTasks[0])
            }
        }
    }, [tasks])

    if (!backupTask) {
        return null
    }

    const completeTasks = backupTask.getTasksComplete() || 0
    const totalTasks = backupTask.getTasksTotal() || 0

    return (
        <div className="flex flex-col h-max w-[90%] shrink-0 bg-slate-800 rounded p-4 gap-4 m-2">
            <p className="w-full h-max font-semibold text-xl select-none">
                {backupTask.getTaskStage() === TaskStage.Complete
                    ? 'Backup Complete'
                    : 'Backup in progress...'}
            </p>
            <WeblensProgress value={backupTask.getProgress()} />
            {Boolean(completeTasks) &&
                backupTask.getTaskStage() !== TaskStage.Complete && (
                    <p className="font-light select-none">
                        Downloading files: {completeTasks} / {totalTasks}
                    </p>
                )}
        </div>
    )
}

export function Admin({ closeAdminMenu }) {
    const user = useSessionStore((state) => state.user)
    const wsSend = useWebsocketStore((state) => state.wsSend)

    useKeyDown('Escape', closeAdminMenu)

    const { data: allUsersInfo, refetch: refetchUsers } = useQuery<User[]>({
        queryKey: ['users'],
        initialData: [],
        queryFn: () =>
            UsersApi.getUsers().then((res) =>
                res.data.map((info) => new User(info))
            ),
    })

    useEffect(() => {
        wsSend('task_subscribe', { taskType: 'do_backup' })
        return () => wsSend('unsubscribe', { taskType: 'do_backup' })
    }, [])

    const server = useSessionStore((state) => state.server)

    if (user === null) {
        return null
    }

    return (
        <div className="settings-menu-container" data-open={true}>
            <div className="settings-menu" onClick={(e) => e.stopPropagation()}>
                <div className="flex flex-col gap-2 select-none">
                    <h1 className="text-3xl pt-4 font-bold">Admin Settings</h1>
                    <div className="flex flex-row justify-between">
                        <p>{server.name}</p>
                        <p className="text-main-accent">
                            {server.role.toUpperCase()}
                        </p>
                    </div>
                </div>

                <div className="top-0 left-0 m-1 absolute">
                    <WeblensButton
                        Left={IconX}
                        squareSize={35}
                        onClick={closeAdminMenu}
                        disabled={server.role !== 'core'}
                    />
                </div>
                <div className="flex flex-col w-full h-full items-center p-4 no-scrollbar">
                    <div className="flex flex-row w-full h-full gap-2">
                        <div className="flex flex-col w-1/2 gap-2">
                            <UsersBox
                                thisUserInfo={user}
                                allUsersInfo={allUsersInfo}
                                refetchUsers={refetchUsers}
                            />
                        </div>
                        <div className="flex flex-col w-1/2 gap-2 items-center">
                            <Servers />
                            <BackupProgress />
                        </div>
                    </div>
                    {/* <PathAutocomplete /> */}
                    <div className="flex flex-row w-full justify-center gap-2 m-2">
                        <WeblensButton
                            label="Clear Cache"
                            squareSize={40}
                            danger
                            requireConfirm
                            onClick={() => {
                                clearCache().then(() => closeAdminMenu())
                            }}
                        />
                        <WeblensButton
                            label={'Reset Server'}
                            danger
                            requireConfirm
                            onClick={() => resetServer()}
                        />
                    </div>
                </div>

                <div className="h-10" />
            </div>
        </div>
    )
}

export default Admin
