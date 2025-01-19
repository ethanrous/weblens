import { WsMsgEvent, wsMsgInfo } from '@weblens/api/Websocket'

import { TaskStageT } from '../../store/TaskStateControl'
import { BackupProgressT } from '../Backup/BackupLogic'

export function AdminWebsocketHandler(
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
            case WsMsgEvent.BackupProgressEvent: {
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

            case WsMsgEvent.BackupFailedEvent: {
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

            case WsMsgEvent.BackupCompleteEvent: {
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

            case WsMsgEvent.CopyFileStartedEvent: {
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

            case WsMsgEvent.CopyFileCompleteEvent: {
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

            case WsMsgEvent.RemoteConnectionChangedEvent: {
                refetchRemotes()
                break
            }
        }
    }
}
