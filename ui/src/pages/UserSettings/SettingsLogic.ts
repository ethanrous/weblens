import { WsEvent, wsMsgInfo } from '@weblens/api/Websocket'
import { TaskStageT } from '@weblens/store/TaskStateControl'

import { BackupProgressT } from '../Backup/BackupLogic'

export function SettingsWebsocketHandler(
	setBackupProgress: React.Dispatch<
		React.SetStateAction<Map<string, BackupProgressT>>
	>,
	refetchRemotes: () => void
) {
	return (msgData: wsMsgInfo) => {
		if (!msgData.eventTag) {
			console.error('Empty websocket message', msgData)
			return
		}

		switch (msgData.eventTag) {
			case WsEvent.BackupProgressEvent: {
				const stages: TaskStageT[] = msgData.content.stages
				setBackupProgress((p: Map<string, BackupProgressT>) => {
					let prog = p.get(msgData.relaySource)
					if (!prog || !stages[0].started) {
						prog = { files: new Map() } as BackupProgressT
					}
					prog.stages = [...stages]
					p.set(msgData.relaySource, prog)
					return new Map(p)
				})
				break
			}

			case WsEvent.BackupFailedEvent: {
				const stages: TaskStageT[] = msgData.content.stages
				setBackupProgress((p: Map<string, BackupProgressT>) => {
					let prog = p.get(msgData.relaySource)
					if (!prog) {
						prog = { files: new Map() } as BackupProgressT
					}
					prog.stages = [...stages]
					prog.error = msgData.content.error
					return new Map(p)
				})
				break
			}

			case WsEvent.BackupCompleteEvent: {
				const stages: TaskStageT[] = msgData.content.stages
				setBackupProgress((p: Map<string, BackupProgressT>) => {
					let prog = p.get(msgData.relaySource)
					if (!prog) {
						prog = { files: new Map() } as BackupProgressT
					}
					prog.stages = [...stages]
					prog.totalTime = msgData.content.totalTime
					p.set(msgData.relaySource, prog)
					return new Map(p)
				})
				refetchRemotes()
				break
			}

			case WsEvent.CopyFileStartedEvent: {
				setBackupProgress((p) => {
					let prog = p.get(msgData.relaySource)
					if (!prog) {
						prog = { files: new Map() } as BackupProgressT
					}
					prog.files.set(msgData.content.filename, {
						name: msgData.content.filename,
						start: new Date(msgData.content.timestamp),
					})
					p.set(msgData.relaySource, prog)
					return new Map(p)
				})
				break
			}

			case WsEvent.CopyFileCompleteEvent: {
				setBackupProgress((p) => {
					let prog = p.get(msgData.relaySource)
					if (!prog) {
						prog = { files: new Map() } as BackupProgressT
					}
					prog.progressCurrent = msgData.content.tasksComplete
					prog.progressTotal = msgData.content.tasksTotal
					prog.files.delete(msgData.content.filename)
					p.set(msgData.relaySource, prog)
					return new Map(p)
				})
				break
			}

			case WsEvent.RemoteConnectionChangedEvent: {
				refetchRemotes()
				break
			}
		}
	}
}
