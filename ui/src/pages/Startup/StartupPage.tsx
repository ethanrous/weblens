import {
    HandleWebsocketMessage,
    useWeblensSocket,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import Logo from '@weblens/components/Logo'
import WebsocketStatusDot from '@weblens/components/filebrowser/websocketStatus'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { useEffect, useState } from 'react'

import { StartupTask, startupWebsocketHandler } from './StartupLogic'
import './startup.scss'

export default function StartUp() {
    const [setupProgress, setSetupProgress] = useState(0)
    const [waitingOn, setWaitingOn] = useState<StartupTask[]>([])
    const [lastTask, setLastTask] = useState<string>('')
    const lastMessage = useWebsocketStore((state) => state.lastMessage)
    const readyState = useWebsocketStore((state) => state.readyState)

    useWeblensSocket()

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            startupWebsocketHandler(setSetupProgress, setWaitingOn, setLastTask)
        )
    }, [lastMessage])

    return (
        <div className="theme-background flex h-screen w-screen flex-col items-center justify-center">
            <Logo size={150} />
            <div className="absolute bottom-1 left-1">
                <WebsocketStatusDot ready={readyState} />
            </div>
            {setupProgress !== 0 && (
                <div className="relative mt-16 flex h-14 w-[670px] max-w-full flex-col gap-2 p-2">
                    <WeblensProgress value={setupProgress} />
                    <p className="theme-text">{lastTask}</p>
                </div>
            )}
            {waitingOn && (
                <div className="absolute bottom-5 flex flex-col">
                    {waitingOn.map((startupTask: StartupTask) => {
                        return (
                            <p
                                key={startupTask.Name}
                                className="theme-text h-4 w-full"
                            >
                                {startupTask.Description}
                            </p>
                        )
                    })}
                </div>
            )}
        </div>
    )
}
