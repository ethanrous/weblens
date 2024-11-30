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
import { ServersApi } from '@weblens/api/ServersApi'
import UsersApi from '@weblens/api/UserApi'
import {
    HandleWebsocketMessage,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import { ThemeToggleButton } from '@weblens/components/HeaderBar'
import Logo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useKeyDown } from '@weblens/components/hooks'
import setupStyle from '@weblens/components/setupStyle.module.scss'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import User from '@weblens/types/user/User'
import { require_css } from '@weblens/util'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { setupWebsocketHandler } from './SetupLogic'

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
            <div className={setupStyle['caution-box']}>
                <div className={setupStyle['caution-header']}>
                    <Text
                        className={setupStyle['subheader-text']}
                        c="white"
                        style={{ paddingTop: 0 }}
                    >
                        This Server Already Has An Owner
                    </Text>
                    <IconExclamationCircle color="white" />
                </div>
                <p className="text-sm text-white">
                    Log in as {owner.username} to continue
                </p>
                <Input
                    variant="unstyled"
                    className={setupStyle['weblens-input-wrapper']}
                    type="password"
                    placeholder="Password"
                    style={{ width: '100%' }}
                    onChange={(e) => setPassword(e.target.value)}
                />
            </div>
        )
    }

    return (
        <div className={setupStyle['caution-box']}>
            <div className={setupStyle['caution-header']}>
                <Text
                    className={setupStyle['subheader-text']}
                    c="#ffffff"
                    style={{ paddingTop: 0 }}
                >
                    This Server Already Has Users
                </Text>
                <IconExclamationCircle color="white" />
            </div>
            <p className="text-white">Select a user to make owner</p>
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
                    <p className="text-white">
                        Log in as {username} to continue
                    </p>
                    {/* <Input */}
                    {/*     variant="unstyled" */}
                    {/*     className={css(["weblens-input-wrapper"])} */}
                    {/*     type="password" */}
                    {/*     placeholder="Password" */}
                    {/*     style={{ width: '100%' }} */}
                    {/*     onChange={(e) => setPassword(e.target.value)} */}
                    {/* /> */}
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

    const { data: users } = useQuery<User[]>({
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
        <div
            className={setupStyle['setup-content-box']}
            data-on-deck={onDeck}
        >
            <div className="w-[90%] absolute">
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    onClick={() => setPage('landing')}
                />
            </div>
            <div className="flex items-center pl-12">
                <h1 className="text-4xl font-bold">Core</h1>
            </div>
            {users.length === 0 && (
                <div className="flex flex-col w-full">
                    <p className=" m-2">Create an Owner Account *</p>
                    <div className="flex flex-col p-4 outline-gray-700 outline rounded gap-2">
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
                <div className={setupStyle['caution-box']}>
                    <div className={setupStyle['caution-header']}>
                        <p
                            className={require_css(
                                setupStyle['subheader-text'],
                                'text-white'
                            )}
                        >
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
                    <p className="m-2">Server Name *</p>
                    <WeblensInput
                        value={serverName}
                        squareSize={50}
                        placeholder="My Radical Weblens Server"
                        valueCallback={setServerName}
                    />
                </div>
            )}
            <div className="flex flex-col w-full">
                <p className="m-2">Server Address</p>
                <WeblensInput
                    squareSize={50}
                    placeholder={location.origin}
                    // valueCallback={}
                />
            </div>

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
                    const res = await ServersApi.initializeServer({
                        name: serverName,
                        role: 'core',
                        username: username,
                        password: password,
                    })

                    if (res.status !== 201) {
                        console.error(res.statusText)
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

                    return fetchServerInfo()
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
        <div
            className={setupStyle['setup-content-box']}
            data-on-deck={onDeck}
        >
            <div className="w-[90%] absolute">
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    onClick={() => setPage('landing')}
                />
            </div>
            <div className="flex pl-16 items-center max-w-max">
                <h1 className="text-4xl font-bold">Backup</h1>
            </div>

            <div className="w-full h-14">
                <p className="m-2">Server Name *</p>
                <WeblensInput
                    placeholder={'My Rad Backup Server'}
                    valueCallback={setServerName}
                />
            </div>

            <div className="w-full h-14">
                <p className="m-2">Core Server Address *</p>
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
                <p className="m-2">Core API Key *</p>
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
                    const res = await ServersApi.initializeServer({
                        name: serverName,
                        role: 'backup',
                        coreAddress: coreAddress,
                        coreKey: apiKey,
                    })
                    if (res.status !== 201) {
                        console.error(res.statusText)
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
    const lastMessage = useWebsocketStore((state) => state.lastMessage)

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            setupWebsocketHandler(setRestoreInProgress, nav)
        )
    }, [lastMessage])

    return (
        <div
            className={setupStyle['setup-content-box']}
            data-on-deck={onDeck}
        >
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
        <div
            className={require_css(
                setupStyle['setup-content-box'],
                'max-h-[60%] mt-40'
            )}
            data-on-deck={onDeck}
        >
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

    let logoSize = 100
    if (page !== 'landing') {
        logoSize = 48
    }

    return (
        <div className={setupStyle['setup-container']}>
            <div className="absolute bottom-4 right-4">
                <ThemeToggleButton />
            </div>
            <div
                className={setupStyle['setup-content-pane']}
                data-active={true}
            >
                <div
                    className={setupStyle['setup-logo']}
                    data-page={page}
                >
                    <Logo size={logoSize} />
                </div>
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
