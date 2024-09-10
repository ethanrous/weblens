import { useDebouncedValue } from '@mantine/hooks'
import {
    IconClipboard,
    IconLockOpen,
    IconTrash,
    IconUserShield,
    IconUserUp,
    IconX,
} from '@tabler/icons-react'
import { DefinedUseQueryResult, useQuery } from '@tanstack/react-query'
import {
    adminCreateUser,
    autocompletePath,
    clearCache,
    deleteApiKey,
    deleteRemote,
    doBackup,
    getApiKeys,
    getRemotes,
    newApiKey,
} from '@weblens/api/ApiFetch'
import {
    ActivateUser,
    DeleteUser,
    GetUsersInfo,
    SetUserAdmin,
    UpdatePassword,
} from '@weblens/api/UserApi'
import { useKeyDown } from '@weblens/components/hooks'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { TaskProgContext } from '@weblens/types/files/FBTypes'
import { WeblensFileParams } from '@weblens/types/files/File'
import { ApiKeyInfo, UserInfoT as UserInfoT } from '@weblens/types/Types'
import React, { useContext, useEffect, useMemo, useState } from 'react'
import { WebsocketContext } from '../../Context'
import { useFileBrowserStore } from '../FileBrowser/FBStateControl'
import { WebsocketStatus } from '../FileBrowser/FileBrowserMiscComponents'
import { TaskProgress, TaskStage } from '../FileBrowser/TaskProgress'

function PathAutocomplete() {
    const [pathSearch, setPathSearch] = useState('')
    const [hoverOffset, setHoverOffset] = useState(0)
    const [bouncedSearch] = useDebouncedValue(pathSearch, 100)
    const names: DefinedUseQueryResult<
        { children: WeblensFileParams[]; folder: WeblensFileParams },
        Error
    > = useQuery<{ children: WeblensFileParams[]; folder: WeblensFileParams }>({
        queryKey: ['pathAutocomplete', bouncedSearch],
        initialData: { children: [], folder: null },
        queryFn: () =>
            autocompletePath(bouncedSearch).then((names) => {
                names.children.sort((f1, f2) =>
                    f1.filename.localeCompare(f2.filename)
                )
                return names
            }),
        retry: false,
    })

    useEffect(() => {
        setHoverOffset(0)
    }, [pathSearch])

    const setBlockFocus = useFileBrowserStore((state) => state.setBlockFocus)

    useKeyDown('Tab', (e) => {
        e.preventDefault()
        const tabFile =
            names.data.children[names.data.children.length - hoverOffset - 1]

        if (!tabFile) {
            return
        }

        setPathSearch((s) => {
            if (!s.endsWith('/')) {
                s += '/'
            }
            return (
                s.slice(0, s.lastIndexOf('/')) +
                '/' +
                tabFile.filename +
                (tabFile.isDir ? '/' : '')
            )
        })
    })

    useKeyDown('ArrowUp', (e) => {
        e.preventDefault()
        e.stopPropagation()
        setHoverOffset((offset) =>
            Math.min(offset + 1, names.data.children.length)
        )
    })

    useKeyDown('ArrowDown', (e) => {
        e.preventDefault()
        e.stopPropagation()
        setHoverOffset((offset) => Math.max(offset - 1, 0))
    })

    const result = useMemo(() => {
        const lastSlash = pathSearch.lastIndexOf('/')
        if (
            (lastSlash === -1 || names.data.children.length === 0) &&
            names.data.folder
        ) {
            return names.data.folder
        }

        if (pathSearch.slice(lastSlash + 1) === '') {
            return names.data.folder
        }

        if (names.data.children.length !== 0) {
            return names.data.children[0]
        }

        return null
    }, [names, pathSearch])

    return (
        <div className="w-[50%] h-10">
            <WeblensInput
                value={pathSearch}
                valueCallback={setPathSearch}
                fillWidth
                openInput={() => setBlockFocus(true)}
                closeInput={() => setBlockFocus(false)}
                failed={names.data.folder === null && pathSearch !== ''}
                placeholder={'File Search'}
            />
            <div
                className="flex flex-col gap-1 absolute -translate-y-[100%] pb-12 pointer-events-none"
                style={{ paddingLeft: pathSearch.length * 9 }}
            >
                {names.data.children.map((cn, i) => {
                    if (pathSearch.endsWith(cn.filename)) {
                        return null
                    }
                    return (
                        <div
                            key={cn.filename}
                            className="flex justify-center items-center p-2 h-10 bg-[#1c1049] rounded pointer-events-auto cursor-pointer"
                            style={{
                                backgroundColor:
                                    i ===
                                    names.data.children.length - hoverOffset - 1
                                        ? '#2549ff'
                                        : '#1c1049',
                            }}
                            onClick={() => {
                                setPathSearch(
                                    (s) =>
                                        s.slice(0, s.lastIndexOf('/')) +
                                        '/' +
                                        cn.filename +
                                        (cn.isDir ? '/' : '')
                                )
                            }}
                        >
                            <p>{cn.filename}</p>
                        </div>
                    )
                })}
            </div>
            {result && (
                <div className="flex flex-row gap-2">
                    <p>{result.filename}</p>
                    <p>{result.id}</p>
                </div>
            )}
        </div>
    )
}

function CreateUserBox({ setAllUsersInfo }: { setAllUsersInfo }) {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')
    const [makeAdmin, setMakeAdmin] = useState(false)
    return (
        <div className="flex flex-col p-2 h-max w-full rounded bg-slate-800 gap-2">
            <p className="w-full h-max font-semibold text-xl select-none">
                Add User
            </p>
            <div className="flex flex-col gap-2 w-full">
                <div className="flex gap-1">
                    <WeblensInput
                        placeholder="Username"
                        squareSize={50}
                        onComplete={null}
                        fillWidth
                        valueCallback={setUserInput}
                    />
                    <WeblensInput
                        placeholder="Password"
                        squareSize={50}
                        onComplete={null}
                        fillWidth
                        password
                        valueCallback={setPassInput}
                    />
                </div>
                <div className="flex">
                    <WeblensButton
                        Left={IconUserShield}
                        label={'Admin'}
                        allowRepeat
                        squareSize={50}
                        toggleOn={makeAdmin}
                        onClick={() => setMakeAdmin(!makeAdmin)}
                    />
                    <WeblensButton
                        label="Create User"
                        squareSize={50}
                        disabled={userInput === '' || passInput === ''}
                        onClick={async () => {
                            await adminCreateUser(
                                userInput,
                                passInput,
                                makeAdmin
                            ).then(() => {
                                GetUsersInfo(setAllUsersInfo)
                                setUserInput('')
                                setPassInput('')
                            })
                            return true
                        }}
                    />
                </div>
            </div>
        </div>
    )
}

const UserRow = ({
    rowUser,
    accessor,
    setAllUsersInfo,
}: {
    rowUser: UserInfoT
    accessor: UserInfoT
    setAllUsersInfo
}) => {
    const [changingPass, setChangingPass] = useState(false)
    return (
        <div
            key={rowUser.username}
            className="flex flex-row w-full h-16 justify-between items-center bg-bottom-grey p-2 rounded"
        >
            <div className="flex flex-col justify-center  w-max h-max">
                <p className="font-semibold w-max text-white">
                    {rowUser.username}
                </p>
                {rowUser.admin && !rowUser.owner && (
                    <p className="text-gray-400">Admin</p>
                )}
                {rowUser.owner && <p className="text-[#aaaaaa]">Owner</p>}
            </div>
            <div className="flex">
                {!changingPass && accessor.owner && (
                    <WeblensButton
                        label="Change Password"
                        labelOnHover={true}
                        Left={IconLockOpen}
                        squareSize={35}
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
                            return UpdatePassword(rowUser.username, '', newPass)
                        }}
                    />
                )}
                {!rowUser.admin && accessor.owner && (
                    <WeblensButton
                        label="Promote to Admin"
                        Left={IconUserUp}
                        labelOnHover={true}
                        allowShrink={false}
                        squareSize={35}
                        onClick={() => {
                            SetUserAdmin(rowUser.username, true).then(() =>
                                GetUsersInfo(setAllUsersInfo)
                            )
                        }}
                    />
                )}
                {!rowUser.owner && rowUser.admin && accessor.owner && (
                    <WeblensButton
                        label="Remove Admin"
                        squareSize={35}
                        onClick={() => {
                            SetUserAdmin(rowUser.username, false).then(() =>
                                GetUsersInfo(setAllUsersInfo)
                            )
                        }}
                    />
                )}
                {rowUser.activated === false && (
                    <WeblensButton
                        label="Activate"
                        squareSize={35}
                        onClick={() => {
                            ActivateUser(rowUser.username).then(() =>
                                GetUsersInfo(setAllUsersInfo)
                            )
                        }}
                    />
                )}

                <WeblensButton
                    squareSize={35}
                    Left={IconTrash}
                    danger
                    centerContent
                    disabled={rowUser.admin && !accessor.owner}
                    onClick={() => {
                        DeleteUser(rowUser.username).then(() =>
                            GetUsersInfo(setAllUsersInfo)
                        )
                    }}
                />
            </div>
        </div>
    )
}

function UsersBox({
    thisUserInfo,
    allUsersInfo,
    setAllUsersInfo,
}: {
    thisUserInfo: UserInfoT
    allUsersInfo: UserInfoT[]
    setAllUsersInfo
}) {
    const usersList = useMemo(() => {
        if (!allUsersInfo) {
            return null
        }
        allUsersInfo.sort((a, b) => {
            return a.username.localeCompare(b.username)
        })

        const usersList = allUsersInfo.map((val) => (
            <UserRow
                key={val.username}
                rowUser={val}
                accessor={thisUserInfo}
                setAllUsersInfo={setAllUsersInfo}
            />
        ))
        return usersList
    }, [allUsersInfo, setAllUsersInfo])

    return (
        <div className="flex flex-col p-2 shrink w-full h-max min-h-96 overflow-x-hidden bg-slate-800 rounded gap-2 no-scrollbar">
            <p className="w-full h-max font-semibold text-xl select-none p-2">
                Users
            </p>
            {usersList}
            <CreateUserBox setAllUsersInfo={setAllUsersInfo} />
        </div>
    )
}

export function ApiKeys() {
    const server = useSessionStore((state) => state.server)

    const keys: DefinedUseQueryResult<ApiKeyInfo[], Error> = useQuery<
        ApiKeyInfo[]
    >({
        queryKey: ['apiKeys'],
        initialData: [],
        queryFn: () => getApiKeys(),
        retry: false,
    })

    const [remotes, setRemotes] = useState([])
    useEffect(() => {
        getRemotes().then((r) => {
            if (r >= 400) {
                return
            }
            setRemotes(r.remotes)
        })
    }, [])

    return (
        <div className="flex flex-col bg-slate-800 rounded items-center p-1 w-full">
            <p className="w-full h-max font-semibold text-xl select-none p-2">
                API Keys
            </p>

            {Boolean(keys.data?.length) && (
                <div className="flex flex-col items-center p-1 rounded w-full">
                    {keys.data.map((k) => {
                        return (
                            <div
                                key={k.id}
                                className="flex flex-row items-center max-w-full w-full bg-bottom-grey rounded p-2"
                            >
                                <div className="flex flex-col grow w-1/2">
                                    <p className="text-white font-semibold text-nowrap w-full truncate select-none">
                                        {k.key}
                                    </p>
                                    {k.remoteUsing !== '' && (
                                        <p className="select-none">
                                            Used by:{' '}
                                            {
                                                remotes.find(
                                                    (r) =>
                                                        r.id === k.remoteUsing
                                                )?.name
                                            }
                                        </p>
                                    )}
                                    {k.remoteUsing === '' && (
                                        <p className="select-none">Unused</p>
                                    )}
                                </div>
                                <WeblensButton
                                    Left={IconClipboard}
                                    onClick={() => {
                                        if (!window.isSecureContext) {
                                            return false
                                        }
                                        navigator.clipboard.writeText(k.key)
                                        return true
                                    }}
                                />
                                <WeblensButton
                                    Left={IconTrash}
                                    danger
                                    onClick={() => {
                                        deleteApiKey(k.key).then(() =>
                                            keys.refetch()
                                        )
                                    }}
                                />
                            </div>
                        )
                    })}
                </div>
            )}
            <WeblensButton
                squareSize={40}
                label="New Api Key"
                onClick={() => {
                    newApiKey().then(() => keys.refetch())
                }}
            />
            <p className="w-full h-max font-semibold text-xl select-none p-2">
                Remotes
            </p>
            <div className="flex flex-col items-center p-2 rounded w-full gap-2">
                {remotes.map((r) => {
                    if (r.id === server.info.id) {
                        return null
                    }
                    return (
                        <div
                            key={r.id}
                            className="flex flex-row items-center w-full rounded p-2 justify-between bg-bottom-grey"
                        >
                            <div className="flex flex-col">
                                <div className="flex flex-row items-center gap-1">
                                    <p className="text-white font-semibold select-none">
                                        {r.name} ({r.role})
                                    </p>
                                    <WebsocketStatus
                                        ready={r.online ? 1 : -1}
                                    />
                                </div>
                                <p className="select-none">{r.id}</p>
                            </div>
                            <WeblensButton
                                label="Sync now"
                                squareSize={40}
                                onClick={async () => doBackup()}
                            />
                            <WeblensButton
                                Left={IconTrash}
                                danger
                                onClick={() => {
                                    deleteRemote(r.id).then(() => {
                                        getRemotes().then((r) => {
                                            if (r >= 400) {
                                                return
                                            }
                                            setRemotes(r.remotes)
                                        })
                                    })
                                }}
                            />
                        </div>
                    )
                })}
            </div>
        </div>
    )
}

function BackupProgress() {
    const { progState } = useContext(TaskProgContext)
    const [backupTask, setBackupTask] = useState<TaskProgress>()
    const [backupTaskId, setBackupTaskId] = useState<string>()

    useEffect(() => {
        if (backupTaskId) {
            setBackupTask(progState.getTask(backupTaskId))
        } else {
            const backupTasks = progState
                .getTasks()
                .filter((t) => t.taskType === 'do_backup')
            if (backupTasks.length !== 0) {
                setBackupTaskId(backupTasks[0].GetTaskId())
                setBackupTask(backupTasks[0])
            }
        }
    }, [progState])

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

export function Admin({ open, closeAdminMenu }) {
    const user = useSessionStore((state) => state.user)
    const [allUsersInfo, setAllUsersInfo] = useState(null)
    const wsSend = useContext(WebsocketContext)

    useKeyDown('Escape', closeAdminMenu)

    useEffect(() => {
        if (open && !allUsersInfo) {
            GetUsersInfo(setAllUsersInfo)
        }
    }, [open])

    useEffect(() => {
        wsSend('task_subscribe', { taskType: 'do_backup' })
        return () => wsSend('unsubscribe', { taskType: 'do_backup' })
    }, [])

    if (user === null || !open) {
        return null
    }

    return (
        <div className="settings-menu-container" data-open={open}>
            <div
                className="settings-menu no-scrollbar"
                onClick={(e) => e.stopPropagation()}
            >
                <div className="top-0 left-0 m-1 absolute">
                    <WeblensButton
                        Left={IconX}
                        squareSize={35}
                        onClick={closeAdminMenu}
                    />
                </div>
                <div className="flex flex-col w-full h-full items-center p-11">
                    <div className="flex flex-row w-full h-full gap-2">
                        <div className="flex flex-col w-1/2 gap-2">
                            <UsersBox
                                thisUserInfo={user}
                                allUsersInfo={allUsersInfo}
                                setAllUsersInfo={setAllUsersInfo}
                            />
                        </div>
                        <div className="flex flex-col w-1/2 gap-2 items-center">
                            <ApiKeys />
                            <div className="flex flex-row w-full justify-around">
                                <WeblensButton
                                    label="Clear Cache"
                                    squareSize={40}
                                    danger
                                    onClick={() => {
                                        clearCache().then(() =>
                                            closeAdminMenu()
                                        )
                                    }}
                                />
                            </div>
                            <BackupProgress />
                        </div>
                    </div>
                    <PathAutocomplete />
                </div>
            </div>
        </div>
    )
}

export default Admin
