import { WsMsgEvent } from '@weblens/api/Websocket'
import { BackupProgressT } from '../Backup/BackupLogic'
import { TaskStageT } from '../FileBrowser/TaskProgress'

export function AdminWebsocketHandler(setBackupProgress, refetchRemotes) {
    return (msgData) => {
        switch (msgData.eventTag) {
            case 'backup_progress': {
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

            case 'backup_failed': {
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

            case 'backup_complete': {
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

            case WsMsgEvent.RemoteConnectionChangedEvent: {
                refetchRemotes()
                break
            }
        }
    }
}
