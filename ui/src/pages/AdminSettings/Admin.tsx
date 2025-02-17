import {
    IconArrowLeft,
    IconClipboard,
    IconLockOpen,
    IconServer,
    IconTrash,
    IconUserMinus,
    IconUserShield,
    IconUserUp,
} from '@tabler/icons-react'
import {
    QueryObserverResult,
    RefetchOptions,
    useQuery,
} from '@tanstack/react-query'
import AccessApi from '@weblens/api/AccessApi'
import MediaApi from '@weblens/api/MediaApi'
import { ServersApi } from '@weblens/api/ServersApi'
import UsersApi from '@weblens/api/UserApi'
import {
    HandleWebsocketMessage,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import { ApiKeyInfo, ServerInfo } from '@weblens/api/swag'
import Logo from '@weblens/components/Logo'
import RemoteStatus from '@weblens/components/RemoteStatus'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useKeyDown } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { ErrorHandler } from '@weblens/types/Types'
import User from '@weblens/types/user/User'
import { useEffect, useMemo, useState } from 'react'

import { BackupProgressT } from '../Backup/BackupLogic'
import { AdminWebsocketHandler } from './adminLogic'
import adminStyle from './adminStyle.module.scss'

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
        <div className="theme-outline flex flex-col p-2 h-max w-full gap-2">
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
    return (
        <div key={rowUser.username} className={adminStyle.adminUserRow}>
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

function UsersBox() {
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
        <div className={adminStyle.contentBox}>
            <div className="flex flex-col p-2 shrink w-full h-full">
                <h4 className="p-1">Users</h4>
                <div className="flex flex-col grow relative gap-1 overflow-y-scroll overflow-x-visible max-h-[50vh]">
                    {usersList}
                </div>
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
        <div key={keyInfo.id} className={adminStyle.adminUserRow}>
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
                        return
                    }
                    navigator.clipboard
                        .writeText(keyInfo.key)
                        .catch(ErrorHandler)
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

function Servers() {
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
            (await ServersApi.getRemotes().then((res) => res.data)) || [],
        retry: false,
    })

    const lastMessage = useWebsocketStore((state) => state.lastMessage)

    const [backupProgress, setBackupProgress] = useState<
        Map<string, BackupProgressT>
    >(new Map())

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            AdminWebsocketHandler(setBackupProgress, () => {
                refetchRemotes().catch(ErrorHandler)
            })
        )
    }, [lastMessage])

    useEffect(() => {
        refetchKeys().catch(ErrorHandler)
    }, [remotes?.length])

    return (
        <div className={adminStyle.contentBox}>
            <div className="h-1/2">
                <div className={adminStyle.contentHeader}>
                    <h2>API Keys</h2>
                </div>

                {Boolean(keys?.length) && (
                    <div className="flex flex-col items-center p-1 rounded w-full gap-1">
                        {keys.map((k) => (
                            <ApiKeyRow
                                key={k.id}
                                keyInfo={k}
                                refetch={() => {
                                    refetchKeys().catch(ErrorHandler)
                                }}
                                remotes={remotes}
                            />
                        ))}
                    </div>
                )}
            </div>
            <div>
                <div className={adminStyle.contentHeader}>
                    <h2>Remotes</h2>
                </div>

                <div className="flex flex-col items-center p-2 rounded w-full gap-2 overflow-scroll">
                    {remotes.length === 0 && (
                        <div className="flex flex-col items-center mt-16">
                            <IconServer />
                            <h2 className="text-center">No remotes</h2>
                            <p>
                                After you set up a {''}
                                <a
                                    href="https://github.com/ethanrous/weblens?tab=readme-ov-file#weblens-backup"
                                    target="_blank"
                                >
                                    backup server
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
            </div>
        </div>
    )
}

// function BackupProgress() {
//     const tasks = useTaskState((state) => state.tasks)
//     const [backupTask, setBackupTask] = useState<TaskProgress>()
//     const [backupTaskId, setBackupTaskId] = useState<string>()
//
//     useEffect(() => {
//         if (backupTaskId) {
//             setBackupTask(tasks.get(backupTaskId))
//         } else {
//             const backupTasks = Array.from(tasks.values()).filter(
//                 (t) => t.taskType === TaskType.Backup
//             )
//             if (backupTasks.length !== 0) {
//                 setBackupTaskId(backupTasks[0].GetTaskId())
//                 setBackupTask(backupTasks[0])
//             }
//         }
//     }, [tasks])
//
//     if (!backupTask) {
//         return null
//     }
//
//     const completeTasks = backupTask.getTasksComplete() || 0
//     const totalTasks = backupTask.getTasksTotal() || 0
//
//     return (
//         <div className="flex flex-col h-max w-[90%] shrink-0 bg-slate-800 rounded p-4 gap-4 m-2">
//             <p className="w-full h-max font-semibold text-xl select-none">
//                 {backupTask.getTaskStage() === TaskStage.Complete
//                     ? 'Backup Complete'
//                     : 'Backup in progress...'}
//             </p>
//             <WeblensProgress value={backupTask.getProgress()} />
//             {Boolean(completeTasks) &&
//                 backupTask.getTaskStage() !== TaskStage.Complete && (
//                     <p className="font-light select-none">
//                         Downloading files: {completeTasks} / {totalTasks}
//                     </p>
//                 )}
//         </div>
//     )
// }

function InfoBox({ title, content }: { title: string; content: string }) {
    return (
        <div className="flex flex-col p-2 m-2 bg-[--wl-card-background] w-max rounded-md">
            <p className="text-sm text-[--wl-text-color-dull]">{title}</p>
            <h4>{content}</h4>
        </div>
    )
}

function GeneralInfo() {
    const server = useSessionStore((state) => state.server)
    const buildTag: string = import.meta.env.VITE_APP_BUILD_TAG
        ? String(import.meta.env.VITE_APP_BUILD_TAG)
        : 'local'

    return (
        <div className={adminStyle.contentBox}>
            <div className="flex flex-col h-full">
                <div className="flex items-center justify-center gap-4">
                    <Logo />
                    <h1>{server.name}</h1>
                </div>
                <div className="flex">
                    <InfoBox title="Server Role" content={server.role} />
                    <InfoBox title="Server ID" content={server.id} />
                    <InfoBox title="Build" content={buildTag} />
                </div>
            </div>
            <div className="flex flex-col p-2 theme-outline">
                <p className="m-1">Advanced</p>
                <WeblensButton
                    label="Clean Media"
                    onClick={() => MediaApi.cleanupMedia()}
                />
            </div>
        </div>
    )
}

enum Page {
    General = 'General',
    Users = 'Users',
    Servers = 'Servers',
}

function AdminPageContent({ pageName }: { pageName: Page }) {
    switch (pageName) {
        case Page.General:
            return <GeneralInfo />
        case Page.Users:
            return <UsersBox />
        case Page.Servers:
            return <Servers />
        default:
            return <div className={adminStyle.contentBox} />
    }
}

export function Admin({ closeAdminMenu }: { closeAdminMenu: () => void }) {
    const user = useSessionStore((state) => state.user)
    const wsSend = useWebsocketStore((state) => state.wsSend)
    const [pageName, setpageName] = useState<Page>(Page.General)

    useKeyDown('Escape', closeAdminMenu)

    useEffect(() => {
        wsSend('taskSubscribe', { taskType: 'do_backup' })
        return () => wsSend('unsubscribe', { taskType: 'do_backup' })
    }, [])

    const server = useSessionStore((state) => state.server)

    if (user === null) {
        return null
    }

    return (
        <div className={adminStyle.settingsMenuContainer} data-open={true}>
            <div
                className={adminStyle.settingsMenu}
                onClick={(e) => e.stopPropagation()}
            >
                <div className="flex flex-col w-[25vw] h-full">
                    <div className="flex flex-row items-center gap-2">
                        <WeblensButton
                            Left={IconArrowLeft}
                            subtle
                            squareSize={35}
                            onClick={closeAdminMenu}
                            disabled={server.role !== 'core'}
                        />
                        <h3 className="font-bold text-nowrap">
                            Server Settings
                        </h3>
                    </div>
                    <div className="flex flex-col mt-4 pr-4 gap-2">
                        <WeblensButton
                            label={Page.General}
                            fillWidth
                            onClick={() => setpageName(Page.General)}
                            toggleOn={pageName === Page.General}
                        />
                        <WeblensButton
                            label={Page.Users}
                            fillWidth
                            onClick={() => setpageName(Page.Users)}
                            toggleOn={pageName === Page.Users}
                        />
                        <WeblensButton
                            label={Page.Servers}
                            fillWidth
                            onClick={() => setpageName(Page.Servers)}
                            toggleOn={pageName === Page.Servers}
                        />
                    </div>
                </div>
                <AdminPageContent pageName={pageName} />
                {/* <div className="flex flex-col gap-2 select-none"> */}
                {/*     <div className="flex flex-row justify-between"> */}
                {/*         <p>{server.name}</p> */}
                {/*         <p className="text-main-accent"> */}
                {/*             {server.role.toUpperCase()} */}
                {/*         </p> */}
                {/*     </div> */}
                {/* </div> */}

                {/* <div className="flex flex-col w-full h-full items-center p-4 no-scrollbar"> */}
                {/*     <div className="flex flex-row w-full h-full gap-2"> */}
                {/*         <div className="flex flex-col w-1/2 gap-2"> */}
                {/*             <UsersBox */}
                {/*                 thisUserInfo={user} */}
                {/*                 allUsersInfo={allUsersInfo} */}
                {/*                 refetchUsers={() => { */}
                {/*                     refetchUsers().catch(ErrorHandler) */}
                {/*                 }} */}
                {/*             /> */}
                {/*         </div> */}
                {/*         <div className="flex flex-col w-1/2 gap-2 items-center"> */}
                {/*             <Servers /> */}
                {/*             <BackupProgress /> */}
                {/*         </div> */}
                {/*     </div> */}
                {/* <PathAutocomplete /> */}
                {/*     <div className="flex flex-row w-full justify-center gap-2 m-2"> */}
                {/*         <WeblensButton */}
                {/*             label="Clear Cache" */}
                {/*             squareSize={40} */}
                {/*             danger */}
                {/*             requireConfirm */}
                {/*             onClick={() => { */}
                {/*                 // clearCache() */}
                {/*                 //     .then(() => closeAdminMenu()) */}
                {/*                 //     .catch(ErrorHandler) */}
                {/*             }} */}
                {/*         /> */}
                {/*         <WeblensButton */}
                {/*             label={'Reset Server'} */}
                {/*             danger */}
                {/*             requireConfirm */}
                {/*             // onClick={() => resetServer()} */}
                {/*         /> */}
                {/*     </div> */}
                {/* </div> */}

                {/* <div className="h-10" /> */}
            </div>
        </div>
    )
}

export default Admin
