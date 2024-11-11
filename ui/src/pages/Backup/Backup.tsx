import { useQuery } from '@tanstack/react-query'
import { attachNewCore } from '@weblens/api/ApiFetch'
import RemoteStatus from '@weblens/components/RemoteStatus'
import { useEffect, useState } from 'react'
import {
    HandleWebsocketMessage,
    useWeblensSocket,
} from '@weblens/api/Websocket'
import {
    backupPageWebsocketHandler,
    BackupProgressT,
    RestoreProgress,
} from './BackupLogic'
import WeblensButton from '@weblens/lib/WeblensButton'
import {
    IconDatabaseImport,
    IconPlus,
    IconRocket,
    IconX,
} from '@tabler/icons-react'
import { ThemeToggleButton } from '@weblens/components/HeaderBar'
import WeblensInput from '@weblens/lib/WeblensInput'
import Logo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useNavigate } from 'react-router-dom'
import { RemoteApi } from '@weblens/api/RemotesApi'
import { ServerInfo } from '@weblens/api/swag'

function NewCoreMenu({ closeNewCore }: { closeNewCore: () => void }) {
    const [coreAddress, setCoreAddress] = useState('')
    const [apiKey, setApiKey] = useState('')

    const user = useSessionStore((state) => state.user)
    console.log(user)

    return (
        <div className="absolute backdrop-blur h-screen w-screen p-20 z-10">
            <div className="flex flex-col bg-wl-background p-10 wl-outline h-max">
                <div className="flex items-center gap-5 mb-8">
                    <WeblensButton
                        Left={IconX}
                        squareSize={35}
                        subtle
                        onClick={() => closeNewCore()}
                    />
                    <IconDatabaseImport color="white" size={'60px'} />
                    <h2 className="header-text">Add Core to Backup</h2>
                </div>

                <p className="body-text m-2">Remote (Core) Weblens Address</p>
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

                <p className="body-text m-2">API Key</p>
                <WeblensInput
                    placeholder={'RUH8gHMH4EgQvw_n2...'}
                    valueCallback={setApiKey}
                    password
                />

                <WeblensButton
                    label="Attach to Core"
                    squareSize={40}
                    Left={IconRocket}
                    disabled={coreAddress === '' || apiKey === ''}
                    doSuper
                    onClick={async () => {
                        return attachNewCore(coreAddress, apiKey).then(() =>
                            closeNewCore()
                        )
                    }}
                />
            </div>
        </div>
    )
}

export default function Backup() {
    const { data: remotes, refetch } = useQuery<ServerInfo[]>({
        queryKey: ['remotes'],
        initialData: [],
        queryFn: async () => {
            return RemoteApi.getRemotes().then((res) => res.data)
        },
    })
    const { lastMessage } = useWeblensSocket()
    const [restoreStage, setRestoreStage] = useState<RestoreProgress>(
        {} as RestoreProgress
    )

    const [backupProgress, setBackupProgress] = useState<
        Map<string, BackupProgressT>
    >(new Map())

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            backupPageWebsocketHandler(
                setRestoreStage,
                setBackupProgress,
                refetch
            )
        )
    }, [lastMessage])

    const server = useSessionStore((state) => state.server)
    const nav = useNavigate()
    useEffect(() => {
        if (server.role && server.role !== 'backup') {
            nav('/')
        }
    }, [])

    const [newCoreMenu, setNewCoreMenu] = useState(false)
    //const local = useSessionStore((state) => state.server)
    if (server.role !== 'backup') {
        return <></>
    }

    return (
        <div className="flex flex-col w-full h-full p-4 items-end">
            {newCoreMenu && (
                <NewCoreMenu closeNewCore={() => setNewCoreMenu(false)} />
            )}
            <div className="flex w-full justify-between items-center">
                <div className="flex flex-row gap-2">
                    <Logo />
                    <h1 className="text-3xl">Backup</h1>
                </div>
                <ThemeToggleButton />
            </div>
            <div className="flex w-full pt-4 gap-1 mb-10">
                {remotes.map((remote) => {
                    return (
                        <RemoteStatus
                            key={remote.id}
                            remoteInfo={remote}
                            refetchRemotes={refetch}
                            restoreProgress={restoreStage}
                            backupProgress={backupProgress.get(remote.id)}
                            setBackupProgress={(progress) => {
                                setBackupProgress((old) => {
                                    const newMap = new Map(old)
                                    newMap.set(remote.id, progress)
                                    return newMap
                                })
                            }}
                        />
                    )
                })}
            </div>
            <WeblensButton
                label="Add Core"
                Left={IconPlus}
                onClick={() => setNewCoreMenu(true)}
            />
        </div>
    )
}
