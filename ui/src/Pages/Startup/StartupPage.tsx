import './startup.scss'
import { HandleWebsocketMessage, useWeblensSocket } from '../../api/Websocket'
import { useEffect, useState } from 'react'
import { startupWebsocketHandler } from './StartupLogic'
import { WeblensProgress } from '../../components/WeblensProgress'

export default function StartUp() {
    const { lastMessage } = useWeblensSocket()
    const [setupProgress, setSetupProgress] = useState(0)
    const [setupMostRecent, setSetupMostRecent] = useState('')
    useEffect(() => {
        HandleWebsocketMessage(
            lastMessage,
            startupWebsocketHandler(setSetupProgress, setSetupMostRecent)
        )
    }, [lastMessage])

    return (
        <div className="flex flex-col justify-center items-center w-screen h-screen bg-background">
            <p className="startup-header-text text-wrap">
                Weblens is Starting...
            </p>
            {setupProgress !== 0 && (
                <div className="flex flex-col relative w-[670px] max-w-full h-14 p-2 mt-16 gap-2">
                    <WeblensProgress value={setupProgress} />
                    <p className="w-full h-4">{setupMostRecent}</p>
                </div>
            )}
        </div>
    )
}
