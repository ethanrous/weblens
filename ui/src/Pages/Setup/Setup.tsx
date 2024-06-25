import { Divider, Input, Text } from '@mantine/core'
import WeblensButton from '../../components/WeblensButton'
import '../../style/setup.css'
import {
    IconArrowLeft,
    IconDatabaseImport,
    IconExclamationCircle,
    IconPackage,
    IconRocket,
} from '@tabler/icons-react'
import { useContext, useEffect, useMemo, useState } from 'react'
import { getUsers, initServer } from '../../api/ApiFetch'
import { useNavigate } from 'react-router-dom'
import { UserInfoT } from '../../types/Types'
import { UserContext } from '../../Context'

const UserSelect = ({
    users,
    username,
    setUsername,
    setPassword,
    owner,
}: {
    users: UserInfoT[]
    username
    setUsername
    setPassword
    owner
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
                <Text
                    className="body-text"
                    style={{ fontSize: '12px' }}
                    c="white"
                >
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
            <Text className="body-text" c="#ffffff" style={{ padding: 0 }}>
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
                    <Text className="body-text" c="#ffffff">
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
    const [serverName, setServerName] = useState(
        existingName ? existingName : ''
    )
    const [users, setUsers] = useState([])
    const nav = useNavigate()

    useEffect(() => {
        getUsers(null).then((r) => {
            setUsers(r)
        })
    }, [])

    const owner = useMemo(() => {
        const owner: UserInfoT = users.filter((u) => u.owner)[0]
        if (owner) {
            setUsername(owner.username)
        }
        return owner
    }, [users])

    let onDeck
    if (page === 'core') {
        onDeck = 'active'
    } else if (page === 'landing') {
        onDeck = 'next'
    }

    return (
        <div className="setup-content-box" data-onDeck={onDeck}>
            <div className="w-full">
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    subtle
                    onClick={() => setPage('landing')}
                />
            </div>
            <div className="flex items-center gap-[20]">
                <IconPackage color="white" size={'60px'} />
                <Text className="header-text">Weblens Core</Text>
            </div>
            {users.length === 0 && (
                <div className="w-full">
                    <Text className="body-text">Create the Owner Account</Text>
                    <Input
                        className="weblens-input-wrapper"
                        variant="unstyled"
                        placeholder="Username"
                        onChange={(e) => {
                            setUsername(e.target.value)
                        }}
                    />
                    <Input
                        className="weblens-input-wrapper"
                        variant="unstyled"
                        type="password"
                        placeholder="Password"
                        onChange={(e) => {
                            setPassword(e.target.value)
                        }}
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

            <Divider />

            {existingName && (
                <div className="caution-box">
                    <div className="caution-header">
                        <Text
                            className="subheader-text"
                            c="#ffffff"
                            style={{ paddingTop: 0 }}
                        >
                            This server already has a name
                        </Text>
                        <IconExclamationCircle color="white" />
                    </div>

                    <Input
                        className="weblens-input-wrapper"
                        styles={{
                            input: {
                                backgroundColor: '#00000000',
                                color: 'white',
                            },
                        }}
                        variant="unstyled"
                        disabled
                        value={serverName}
                    />
                </div>
            )}
            {!existingName && (
                <div>
                    <Text className="body-text">Name This Weblens Server</Text>
                    <Input
                        className="weblens-input-wrapper"
                        // classNames={{ input: "weblens-input-wrapper" }}
                        variant="unstyled"
                        disabled={Boolean(existingName)}
                        value={serverName}
                        placeholder="My Radical Weblens Server"
                        onChange={(e) => {
                            setServerName(e.target.value)
                        }}
                    />
                </div>
            )}

            <WeblensButton
                label="Start Weblens"
                squareSize={50}
                Left={IconRocket}
                disabled={
                    serverName === '' || username === '' || password === ''
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

                    nav('/login')

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
    const nav = useNavigate()

    let onDeck
    if (page === 'backup') {
        onDeck = 'active'
    } else if (page === 'landing' || page === 'core') {
        onDeck = 'next'
    }
    return (
        <div className="setup-content-box" data-onDeck={onDeck}>
            <div style={{ width: '100%' }}>
                <WeblensButton
                    Left={IconArrowLeft}
                    squareSize={35}
                    subtle
                    onClick={() => setPage('landing')}
                />
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
                <IconDatabaseImport color="white" size={'60px'} />
                <Text className="header-text">Weblens Backup</Text>
            </div>
            <Text className="body-text">Name Local (Backup) Server</Text>
            <Input
                className="weblens-input-wrapper"
                variant="unstyled"
                placeholder="Rad Backup Server"
                onChange={(e) => {
                    setServerName(e.target.value)
                }}
            />
            <Text className="body-text">Remote (Core) Weblens Address</Text>
            <Input
                className="weblens-input-wrapper"
                variant="unstyled"
                placeholder="https://myremoteweblens.net/"
                onChange={(e) => {
                    setCoreAddress(e.target.value)
                }}
            />
            <Text className="body-text">API Key</Text>
            <Input
                className="weblens-input-wrapper"
                variant="unstyled"
                type="password"
                placeholder="RUH8gHMH4EgQvw_n2..."
                onChange={(e) => {
                    setApiKey(e.target.value)
                }}
            />

            <WeblensButton
                label="Attach to Remote"
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

                    nav('/login')

                    return true
                }}
            />
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
    let onDeck
    if (page === 'landing') {
        onDeck = 'active'
    } else if (page === 'core' || page === 'backup') {
        onDeck = 'prev'
    }

    return (
        <div className="setup-content-box" data-onDeck={onDeck}>
            <Text className="title-text">WEBLENS</Text>
            {/* <Text className="content-title-text">Set Up Weblens</Text> */}
            <WeblensButton
                label="Set Up Weblens Core"
                Left={IconPackage}
                squareSize={75}
                centerContent
                onClick={() => setPage('core')}
            />
            <Text>Or...</Text>
            <WeblensButton
                label="Set Up As Backup"
                squareSize={40}
                Left={IconDatabaseImport}
                style={{ width: '200px' }}
                centerContent
                subtle
                onClick={() => setPage('backup')}
            />
        </div>
    )
}

const Setup = () => {
    const { serverInfo } = useContext(UserContext)
    const [page, setPage] = useState('landing')

    if (!serverInfo) {
        return null
    }

    return (
        <div className="setup-container">
            <div className="setup-content-pane" data-active={true}>
                <Landing page={page} setPage={setPage} />
                <Core
                    page={page}
                    setPage={setPage}
                    existingName={serverInfo.name}
                />
                <Backup page={page} setPage={setPage} />
            </div>
            {/* <ScatteredPhotos /> */}
        </div>
    )
}

export default Setup
