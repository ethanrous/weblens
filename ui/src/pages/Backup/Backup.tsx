import {
    IconDatabaseImport,
    IconPlus,
    IconRocket,
    IconX,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { ServersApi } from '@weblens/api/ServersApi'
import {
    HandleWebsocketMessage,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import { ServerInfo } from '@weblens/api/swag'
import { ThemeToggleButton } from '@weblens/components/HeaderBar'
import Logo from '@weblens/components/Logo'
import RemoteStatus from '@weblens/components/RemoteStatus'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { ErrorHandler } from '@weblens/types/Types'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import {
    BackupProgressT,
    RestoreProgress,
    backupPageWebsocketHandler,
} from './BackupLogic'

function NewCoreMenu({ closeNewCore }: { closeNewCore: () => void }) {
    const [coreAddress, setCoreAddress] = useState('')
    const [apiKey, setApiKey] = useState('')

    return (
        <div className="absolute z-10 h-screen w-screen p-20 backdrop-blur-sm">
            <div className="bg-wl-background flex h-max flex-col p-10 wl-outline">
                <div className="mb-8 flex items-center gap-5">
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
                    valid={
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
                    onClick={() => {
                        ServersApi.createRemote({
                            coreAddress: coreAddress,
                            usingKey: apiKey,
                        }).catch(ErrorHandler)
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
            return ServersApi.getRemotes().then((res) => res.data)
        },
    })
    const lastMessage = useWebsocketStore((state) => state.lastMessage)
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
                () => {
                    refetch().catch(ErrorHandler)
                }
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
        <div className="flex h-full w-full flex-col items-end p-4">
            {newCoreMenu && (
                <NewCoreMenu closeNewCore={() => setNewCoreMenu(false)} />
            )}
            <div className="flex w-full items-center justify-between">
                <div className="flex flex-row gap-2">
                    <Logo />
                    <h1 className="text-3xl">Backup</h1>
                </div>
                <ThemeToggleButton />
            </div>
            <div className="mb-10 flex w-full gap-1 pt-4">
                {remotes.map((remote) => {
                    return (
                        <RemoteStatus
                            key={remote.id}
                            remoteInfo={remote}
                            refetchRemotes={() => {
                                refetch().catch(ErrorHandler)
                            }}
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
