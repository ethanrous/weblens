import { Divider, Input, Text } from '@mantine/core'
import {
    IconArrowLeft,
    IconDatabase,
    IconDatabaseImport,
    IconExclamationCircle,
    IconPackage,
    IconRocket,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { initServer } from '@weblens/api/ApiFetch'
import UsersApi from '@weblens/api/UserApi'
import { useKeyDown } from '@weblens/components/hooks'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import '@weblens/components/setup.scss'
import WeblensInput from '@weblens/lib/WeblensInput'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import Logo from '@weblens/components/Logo'
import { ThemeToggleButton } from '@weblens/components/HeaderBar'
import {
    HandleWebsocketMessage,
    useWeblensSocket,
} from '@weblens/api/Websocket'
import { setupWebsocketHandler } from './SetupLogic'
import User from '@weblens/types/user/user'

const UserSelect = ({
    users,
    username,
    setUsername,
    setPassword,
    owner,
}: {
    users: User[]
    username: string
    setUsername: (username: string) => void
    setPassword: (password: string) => void
    owner: User
}) => {
    if (users.length === 0) {
        return null
    }

    if (owner) {
        return (
            <div className="caution-box">
                <div className="caution-header">
                    <Text
                        className="subheader-text"
                        c="white"
                        style={{ paddingTop: 0 }}
                    >
                        This Server Already Has An Owner
                    </Text>
                    <IconExclamationCircle color="white" />
                </div>
                <Text className="" style={{ fontSize: '12px' }} c="white">
                    Log in as {owner.username} to continue
                </Text>
                <Input
                    variant="unstyled"
                    className="weblens-input-wrapper"
                    type="password"
                    placeholder="Password"
                    style={{ width: '100%' }}
                    onChange={(e) => setPassword(e.target.value)}
                />
            </div>
        )
    }

    return (
        <div className="caution-box">
            <div className="caution-header">
                <Text
                    className="subheader-text"
                    c="#ffffff"
                    style={{ paddingTop: 0 }}
                >
                    This Server Already Has Users
                </Text>
                <IconExclamationCircle color="white" />
            </div>
            <Text className="" c="#ffffff" style={{ padding: 0 }}>
                Select a user to make owner
            </Text>
            <div className="w-full h-max max-h-[100px] shrink-0 overflow-scroll">
                {users.map((u) => {
                    return (
                        <WeblensButton
                            toggleOn={u.username === username}
                            squareSize={40}
                            allowRepeat={false}
                            key={u.username}
                            label={u.username}
                            style={{ padding: 2 }}
                            onClick={() => setUsername(u.username)}
                        />
                    )
                })}
            </div>
            {username && (
                <div className="w-full">
                    <Text className="" c="#ffffff">
                        Log in as {username} to continue
                    </Text>
                    <Input
                        variant="unstyled"
                        className="weblens-input-wrapper"
                        type="password"
                        placeholder="Password"
                        style={{ width: '100%' }}
                        onChange={(e) => setPassword(e.target.value)}
                    />
                </div>
            )}
        </div>
    )
}

const Core = ({
    page,
    setPage,
    existingName,
}: {
    page: string
    setPage: (page: string) => void
    existingName: string
}) => {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [passwordVerify, setPasswordVerify] = useState('')
    const [buttonRef, setButtonRef] = useState<HTMLDivElement>()
    const [serverName, setServerName] = useState(
        existingName ? existingName : ''
    )

    const fetchServerInfo = useSessionStore((state) => state.fetchServerInfo)
    const setUser = useSessionStore((state) => state.setUser)
    useKeyDown('Enter', () => buttonRef.click())

    const { data: users } = useQuery({
        queryKey: ['setupUsers'],
        queryFn: () => {
            // if (serverInfo.info.role !== 'init') {
            //     UsersApi.
            //     return getUsers()
            // }
            return []
        },
        initialData: [],
    })

    const owner = useMemo(() => {
        const owner: User = users.filter((u) => u.owner)[0]
        if (owner) {
            setUsername(owner.username)
        }
        return owner
    }, [users])

    let onDeck: string
    if (page === 'core') {
        onDeck = 'active'
    } else if (page === 'landing') {
        onDeck = 'next'
    }

    return (
        <div className="setup-content-box" data-on-deck={onDeck}>
            <div className="w-[90%] absolute">
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    onClick={() => setPage('landing')}
                />
            </div>
            <div className="flex items-center gap-[20]">
                <IconPackage className="text-theme-text" size={'60px'} />
                <h1 className="pl-4">Core</h1>
            </div>
            {users.length === 0 && (
                <div className="flex flex-col w-full">
                    <p className=" m-2">Create an Owner Account</p>
                    <div className="flex flex-col p-4 outline-gray-700 outline rounded">
                        <WeblensInput
                            placeholder={'Username'}
                            squareSize={50}
                            valueCallback={setUsername}
                        />
                        <WeblensInput
                            placeholder={'Password'}
                            squareSize={50}
                            password={true}
                            valueCallback={setPassword}
                        />
                        <WeblensInput
                            placeholder={'Verify Password'}
                            squareSize={50}
                            valueCallback={setPasswordVerify}
                            failed={
                                passwordVerify !== '' &&
                                password !== passwordVerify
                            }
                            password={true}
                        />
                    </div>
                </div>
            )}

            <UserSelect
                users={users}
                username={username}
                setUsername={setUsername}
                setPassword={setPassword}
                owner={owner}
            />

            <Divider />

            {existingName && (
                <div className="caution-box">
                    <div className="caution-header">
                        <p className="subheader-text text-white">
                            This server already has a name
                        </p>
                        <IconExclamationCircle color="white" />
                    </div>

                    <WeblensInput
                        // disabled={true}
                        value={serverName}
                    />
                </div>
            )}
            {!existingName && (
                <div className="flex flex-col w-full">
                    <p className="m-2">Name This Weblens Server</p>
                    <WeblensInput
                        // disabled={Boolean(existingName)}
                        value={serverName}
                        squareSize={50}
                        placeholder="My Radical Weblens Server"
                        valueCallback={setServerName}
                    />
                </div>
            )}

            <WeblensButton
                label="Start Weblens"
                setButtonRef={setButtonRef}
                squareSize={50}
                Left={IconRocket}
                disabled={
                    serverName === '' ||
                    username === '' ||
                    password === '' ||
                    password !== passwordVerify
                }
                doSuper
                onClick={async () => {
                    const ret = await initServer(
                        serverName,
                        'core',
                        username,
                        password,
                        '',
                        ''
                    )
                    if (ret.status !== 201) {
                        console.error(ret.statusText)
                        return false
                    }

                    await new Promise((r) => setTimeout(r, 200))

                    const gotInfo = await UsersApi.getUser()
                        .then((res) => {
                            const user = new User(res.data)
                            user.isLoggedIn = true
                            setUser(user)
                            return true
                        })
                        .catch((r) => {
                            console.error(r)
                            const user = new User()
                            user.isLoggedIn = false
                            setUser(user)
                            return false
                        })

                    if (!gotInfo) {
                        return false
                    }

                    fetchServerInfo()

                    return true
                }}
            />
        </div>
    )
}

const Backup = ({
    page,
    setPage,
}: {
    page: string
    setPage: (page: string) => void
}) => {
    const [serverName, setServerName] = useState('')
    const [coreAddress, setCoreAddress] = useState('')
    const [apiKey, setApiKey] = useState('')

    const fetchServerInfo = useSessionStore((state) => state.fetchServerInfo)

    const nav = useNavigate()

    let onDeck: string
    if (page === 'backup') {
        onDeck = 'active'
    } else if (page === 'landing' || page === 'core') {
        onDeck = 'next'
    }
    return (
        <div className="setup-content-box" data-on-deck={onDeck}>
            <div className="w-[90%] absolute">
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    onClick={() => setPage('landing')}
                />
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
                <IconDatabaseImport className="text-theme-text" size={'60px'} />
                <h1>Backup</h1>
            </div>

            <div className="w-full h-14">
                <p className="m-2">Name This Server</p>
                <WeblensInput
                    placeholder={'My Rad Backup Server'}
                    valueCallback={setServerName}
                />
            </div>

            <div className="w-full h-14">
                <p className="m-2">Remote (Core) Weblens Address</p>
                <WeblensInput
                    placeholder={'https://myremoteweblens.net/'}
                    valueCallback={setCoreAddress}
                    failed={
                        coreAddress &&
                        coreAddress.match(
                            '^http(s)?:\\/\\/[^:]+(:\\d{2,5})?/?$'
                        ) === null
                    }
                />
            </div>

            <div className="w-full h-14">
                <p className="m-2">API Key</p>
                <WeblensInput
                    placeholder={'RUH8gHMH4EgQvw_n2...'}
                    valueCallback={setApiKey}
                    password
                />
            </div>

            <div />

            <WeblensButton
                label="Attach to Core"
                squareSize={40}
                Left={IconRocket}
                disabled={
                    serverName === '' || coreAddress === '' || apiKey === ''
                }
                doSuper
                onClick={async () => {
                    const ret = await initServer(
                        serverName,
                        'backup',
                        '',
                        '',
                        coreAddress,
                        apiKey
                    )
                    if (ret.status !== 201) {
                        console.error(ret.statusText)
                        return false
                    }

                    await new Promise((r) => setTimeout(r, 200))

                    await fetchServerInfo()
                    nav('/backup')

                    return true
                }}
            />
        </div>
    )
}

const Restore = ({
    page,
    setPage,
}: {
    page: string
    setPage: (page: string) => void
}) => {
    let onDeck: string
    if (page === 'restore') {
        onDeck = 'active'
    } else if (page === 'landing' || page === 'core') {
        onDeck = 'next'
    }

    const nav = useNavigate()
    const [restoreInProgress, setRestoreInProgress] = useState(false)

    const { lastMessage } = useWeblensSocket()
    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            setupWebsocketHandler(setRestoreInProgress, nav)
        )
    }, [lastMessage])

    return (
        <div className="setup-content-box" data-on-deck={onDeck}>
            <div className="w-[90%] absolute">
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    subtle
                    onClick={() => setPage('landing')}
                />
            </div>
            {restoreInProgress && (
                <h1 className="flex h-[80%] items-center">
                    Restore In Progress
                </h1>
            )}
            {!restoreInProgress && (
                <h1 className="flex h-[80%] items-center">
                    Waiting for Restore
                </h1>
            )}
        </div>
    )
}

const Landing = ({
    page,
    setPage,
}: {
    page: string
    setPage: (page: string) => void
}) => {
    let onDeck: string
    if (page === 'landing') {
        onDeck = 'active'
    } else if (page === 'core' || page === 'backup') {
        onDeck = 'prev'
    }

    return (
        <div className="setup-content-box" data-on-deck={onDeck}>
            <div className="p-6">
                <Logo size={140} />
            </div>
            <WeblensButton
                label="Set Up Weblens Core"
                Left={IconPackage}
                squareSize={75}
                centerContent
                onClick={() => setPage('core')}
            />
            <div className="w-[90%] h-[1px] bg-[--wl-outline-subtle]" />
            <div className="flex flex-row gap-2">
                <WeblensButton
                    label="Set Up As Backup"
                    squareSize={40}
                    Left={IconDatabase}
                    style={{ width: '200px' }}
                    centerContent
                    subtle
                    onClick={() => setPage('backup')}
                />
                <WeblensButton
                    label="Restore From Backup"
                    squareSize={40}
                    Left={IconDatabaseImport}
                    style={{ width: '200px' }}
                    centerContent
                    subtle
                    onClick={() => setPage('restore')}
                />
            </div>
        </div>
    )
}

const Setup = () => {
    const server = useSessionStore((state) => state.server)
    const [page, setPage] = useState('landing')
    const nav = useNavigate()
    useEffect(() => {
        if (!server) {
            return
        }
        if (server.role !== 'init') {
            nav('/')
        }
    }, [server.role])

    if (!server) {
        return null
    }

    return (
        <div className="setup-container">
            <div className="absolute bottom-4 right-4">
                <ThemeToggleButton />
            </div>
            <div className="setup-content-pane" data-active={true}>
                <Landing page={page} setPage={setPage} />
                <Core
                    page={page}
                    setPage={setPage}
                    existingName={server.name}
                />
                <Backup page={page} setPage={setPage} />
                <Restore page={page} setPage={setPage} />
            </div>
        </div>
    )
}

export default Setup
