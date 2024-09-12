import './startup.scss'
import {
    HandleWebsocketMessage,
    useWeblensSocket,
} from '@weblens/api/Websocket'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { useEffect, useState } from 'react'
import { StartupTask, startupWebsocketHandler } from './StartupLogic'

export default function StartUp() {
    const { lastMessage } = useWeblensSocket()
    const [setupProgress, setSetupProgress] = useState(0)
    const [waitingOn, setWaitingOn] = useState<StartupTask[]>([])
    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            startupWebsocketHandler(setSetupProgress, setWaitingOn)
        )
    }, [lastMessage])

    console.log(waitingOn)

    return (
        <div className="flex flex-col justify-center items-center w-screen h-screen bg-background">
            <p className="startup-header-text text-wrap">
                Weblens is Starting...
            </p>
            {setupProgress !== 0 && (
                <div className="flex flex-col relative w-[670px] max-w-full h-14 p-2 mt-16 gap-2">
                    <WeblensProgress value={setupProgress} />
                </div>
            )}
            {waitingOn && (
                <div className="flex flex-col">
                    {waitingOn.map((startupTask: StartupTask) => {
                        return <p key={startupTask.Name} className="w-full h-4">{startupTask.Description}</p>
                    })}
                </div>
            )}
        </div>
    )
}
