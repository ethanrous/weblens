import { WsMsgEvent, wsMsgInfo } from '../../api/Websocket'

export type StartupTask = {
    Name: string
    Description: string
    StartedAt: string
}

export function startupWebsocketHandler(
    setSetupProgress: (progress: number) => void,
    setSetupMostRecent: (waitingOn: StartupTask[]) => void,
    setLastTask: (lastTask: string) => void
) {
    return (msgData: wsMsgInfo) => {
        switch (msgData.eventTag) {
            case WsMsgEvent.StartupProgressEvent: {
                setSetupMostRecent(msgData.content?.waitingOn)
                break
            }
            case WsMsgEvent.WeblensLoadedEvent: {
                setTimeout(() => location.reload(), 500)
                break
            }
            case WsMsgEvent.TaskCompleteEvent: {
                setSetupProgress(
                    (msgData.content.tasks_complete /
                        msgData.content.tasks_total) *
                        100
                )
                setLastTask(msgData.content.filename)
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
