import { WsMsgEvent } from '../../api/Websocket'

export type StartupTask = {
    Name: string
    Description: string
    StartedAt: string
}

export function startupWebsocketHandler(
    setSetupProgress,
    setSetupMostRecent: (waitingOn: StartupTask[]) => void,
    setLastTask: (lastTask: string) => void
) {
    return (msgData) => {
        switch (msgData.eventTag) {
            case 'startup_progress': {
                setSetupMostRecent(msgData.content.waitingOn)
                break
            }
            case WsMsgEvent.WeblensLoadedEvent: {
                setTimeout(() => location.reload(), 500)
                break
            }
            case 'task_complete': {
                setSetupProgress(
                    (msgData.content.tasksComplete /
                        msgData.content.tasksTotal) *
                        100
                )
                setLastTask(msgData.content.fileName)
                // setSetupMostRecent(
                //     `${msgData.taskType}: ${msgData.content.fileName}`
                // )
                break
            }
            default: {
                // setSetupMostRecent(`${msgData.taskType}`)
            }
        }
    }
}
