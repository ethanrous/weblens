import { WsEvent, wsMsgInfo } from '../../api/Websocket'

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
			case WsEvent.StartupProgressEvent: {
				setSetupMostRecent(msgData.content?.waitingOn)
				break
			}
			case WsEvent.WeblensLoadedEvent: {
				setTimeout(() => location.reload(), 500)
				break
			}
			case WsEvent.TaskCompleteEvent: {
				setSetupProgress(
					(msgData.content.tasksComplete /
						msgData.content.tasksTotal) *
					100
				)
				setLastTask(msgData.content.filename)
				// setSetupMostRecent(
				//     `${msgData.taskType}: ${msgData.content.filename}`
				// )
				break
			}
			default: {
				// setSetupMostRecent(`${msgData.taskType}`)
			}
		}
	}
}
