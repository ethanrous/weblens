import './startup.scss'
import 'components/theme.scss'
import {
    HandleWebsocketMessage,
    useWebsocketStore,
} from '@weblens/api/Websocket'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { useEffect, useState } from 'react'
import { StartupTask, startupWebsocketHandler } from './StartupLogic'
import Logo from '@weblens/components/Logo'
import { WebsocketStatus } from '../FileBrowser/FileBrowserMiscComponents'

export default function StartUp() {
    const [setupProgress, setSetupProgress] = useState(0)
    const [waitingOn, setWaitingOn] = useState<StartupTask[]>([])
    const [lastTask, setLastTask] = useState<string>('')
    const lastMessage = useWebsocketStore((state) => state.lastMessage)
    const readyState = useWebsocketStore((state) => state.readyState)

    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            startupWebsocketHandler(setSetupProgress, setWaitingOn, setLastTask)
        )
    }, [lastMessage])

    return (
        <div className="flex flex-col justify-center items-center w-screen h-screen theme-background">
            <Logo size={150} />
            <div className="absolute bottom-1 left-1">
                <WebsocketStatus ready={readyState} />
            </div>
            {setupProgress !== 0 && (
                <div className="flex flex-col relative w-[670px] max-w-full h-14 p-2 mt-16 gap-2">
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
                                className="theme-text w-full h-4"
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
