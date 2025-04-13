import {
    IconArrowLeft,
    IconDatabase,
    IconDatabaseImport,
    IconExclamationCircle,
    IconPackage,
    IconRocket,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { TowersApi } from '@weblens/api/ServersApi'
import UsersApi from '@weblens/api/UserApi'
import {
    HandleWebsocketMessage,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import Logo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
import setupStyle from '@weblens/components/setupStyle.module.scss'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { useKeyDown } from '@weblens/lib/hooks'
import User from '@weblens/types/user/User'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { SignupInputForm } from '../Signup/Signup'
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
            <div className={setupStyle.cautionBox}>
                <div className={setupStyle.cautionHeader}>
                    <span>This Server Already Has An Owner</span>
                    <IconExclamationCircle color="white" />
                </div>
                <p className="text-sm text-white">
                    Log in as {owner.username} to continue
                </p>
                <WeblensInput
                    className={setupStyle.weblensInputWrapper}
                    password
                    placeholder="Password"
                    valueCallback={(v) => setPassword(v)}
                />
            </div>
        )
    }

    return (
        <div className={setupStyle.cautionBox}>
            <div className={setupStyle.cautionHeader}>
                <span>This Server Already Has Users</span>
                <IconExclamationCircle color="white" />
            </div>
            <p className="text-white">Select a user to make owner</p>
            <div className="h-max max-h-[100px] w-full shrink-0 overflow-scroll">
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
    const [fullName, setFullName] = useState('')
    const [buttonRef, setButtonRef] = useState<HTMLButtonElement>()
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
        <div className={setupStyle.setupContentBox} data-on-deck={onDeck}>
            <div className="absolute w-[90%] mb-10">
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
                <div className="border-color-border-primary flex w-full flex-col rounded-md border p-4 mt-6">
                    <h4 className="mb-4">Create an Owner Account</h4>
                    <SignupInputForm
                        setFullName={setFullName}
                        setUsername={setUsername}
                        setPassword={setPassword}
                        setError={() => {}}
                        disabled={false}
                    />
                </div>
            )}

            <UserSelect
                users={users}
                username={username}
                setUsername={setUsername}
                setPassword={setPassword}
                owner={owner}
            />

            {existingName && (
                <div className={setupStyle.cautionBox}>
                    <div className={setupStyle.cautionHeader}>
                        <p className={setupStyle.subheaderText}>
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
                <>
                    <label
                        className="mb-1 mt-8 flex items-center mr-auto"
                        htmlFor="serverName"
                    >
                        <span>Server Name</span>
                        <sup className="h-max text-red-500 ">*</sup>
                    </label>

                    <WeblensInput
                        value={serverName}
                        squareSize={50}
                        placeholder="My Radical Weblens Server"
                        valueCallback={setServerName}
                        autoComplete="serverName"
                    />
                </>
            )}
            <label htmlFor="serverAddress" className="mt-4 mb-1 mr-auto">
                <span>Server Address</span>
            </label>

            <WeblensInput
                squareSize={50}
                value={location.origin}
                autoComplete="serverAddress"
				className='mb-6'
            />

            <WeblensButton
                label="Start Weblens"
                setButtonRef={setButtonRef}
                squareSize={50}
                Left={IconRocket}
                disabled={
                    serverName === '' || username === '' || password === ''
                }
                doSuper
                onClick={async () => {
                    const res = await TowersApi.initializeTower({
                        name: serverName,
                        role: 'core',
                        username: username,
                        password: password,
                        fullName: fullName,
                    })

                    if (res.status !== 201) {
                        console.error(res.statusText)
                        return false
                    }

                    // let srvInfo: ServerInfo | void
                    // let tries = 0
                    // while (
                    //     !srvInfo ||
                    //     !srvInfo.started ||
                    //     srvInfo.role === 'init'
                    // ) {
                    //     await new Promise((r) => setTimeout(r, 500))
                    //     srvInfo = await ServersApi.getServerInfo()
                    //         .then((res) => res.data)
                    //         .catch((r: Error) =>
                    //             ErrorHandler(r, 'Waiting for server to start')
                    //         )
                    //     if (
                    //         srvInfo &&
                    //         srvInfo.started &&
                    //         srvInfo.role !== 'init'
                    //     ) {
                    //         break
                    //     }
                    //     if (tries > 25) {
                    //         console.error('Server failed to start')
                    //         return false
                    //     }
                    //     tries++
                    // }

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

    const addressIsValid =
        coreAddress.match('^http(s)?:\\/\\/[^:]+(:\\d{2,5})?/?$') !== null

    return (
        <div className={setupStyle.setupContentBox} data-on-deck={onDeck}>
            <div className="absolute w-[90%] mb-10">
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    onClick={() => setPage('landing')}
                />
            </div>
            <div className="flex max-w-max items-center pl-16">
                <h1 className="text-4xl font-bold">Backup</h1>
            </div>

            <div className="h-14 w-full">
                <p className="m-2">Server Name *</p>
                <WeblensInput
                    placeholder={'My Rad Backup Server'}
                    valueCallback={setServerName}
                />
            </div>

            <div className="h-14 w-full">
                <p className="m-2">Core Server Address *</p>
                <WeblensInput
                    placeholder={'https://myremoteweblens.net/'}
                    valueCallback={setCoreAddress}
                    valid={coreAddress === '' || addressIsValid ? null : false}
                />
            </div>

            <div className="h-14 w-full">
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
                    const res = await TowersApi.initializeTower({
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
        <div className={setupStyle.setupContentBox} data-on-deck={onDeck}>
            <div className="absolute w-[90%] mb-10">
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
        <div className={setupStyle.setupContentBox} data-on-deck={onDeck}>
            <h3>Welcome to Weblens!</h3>
            <h5>Choose how to set up this server</h5>
            <div className="my-auto flex w-full">
                <WeblensButton
                    label="Set Up Weblens Core"
                    Left={IconPackage}
                    centerContent
                    onClick={() => setPage('core')}
                    className="mx-auto h-20 max-w-96 justify-center"
                    size="jumbo"
                    fillWidth
                />
            </div>
            <div className="bg-border-primary h-[1px] w-[90%]" />
            <div className="my-auto flex h-[10%] flex-row gap-2">
                <WeblensButton
                    label="Set Up As Backup"
                    Left={IconDatabase}
                    className="w-52"
                    onClick={() => setPage('backup')}
                    flavor="outline"
                />
                <WeblensButton
                    label="Restore From Backup"
                    Left={IconDatabaseImport}
                    className="w-52"
                    onClick={() => setPage('restore')}
                    flavor="outline"
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
        <div className={setupStyle.setupContainer}>
            <div className="flex w-full justify-center text-center sm:mt-16">
                <Logo size={100} />
                <h1 className="mt-auto">EBLENS</h1>
            </div>
            <div className={setupStyle.setupContentPane}>
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
