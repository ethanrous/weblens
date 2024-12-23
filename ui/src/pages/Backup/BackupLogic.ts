import { WsMsgEvent, wsMsgInfo } from '../../api/Websocket'
import { TaskStageT } from '../../store/TaskStateControl'

export type RestoreProgress = {
    stage: string
    error: string
    timestamp: Date
    progressTotal: number
    progressCurrent: number
}

type FileBackupProgress = {
    name: string
    start: Date
}

export type BackupProgressT = {
    stages: TaskStageT[]
    error: string
    timestamp: Date
    progressTotal: number
    progressCurrent: number
    files: Map<string, FileBackupProgress>
    totalTime: number
}

export function backupPageWebsocketHandler(
    setRestoreStage: React.Dispatch<React.SetStateAction<RestoreProgress>>,
    setBackupProgress: React.Dispatch<
        React.SetStateAction<Map<string, BackupProgressT>>
    >,
    refetchRemotes: () => void
) {
    return (msgData: wsMsgInfo) => {
        switch (msgData.eventTag) {
            case WsMsgEvent.RestoreStartedEvent: {
                setRestoreStage((p) => {
                    if (msgData.content.stage) {
                        p.stage = msgData.content.stage
                    }
                    if (msgData.content.filesTotal) {
                        p.progressTotal = msgData.content.filesTotal
                        p.progressCurrent = msgData.content.filesRestored
                    }
                    if (msgData.content.timestamp) {
                        p.timestamp = new Date(msgData.content.timestamp)
                    }

                    return { ...p }
                })
                break
            }

            case WsMsgEvent.RestoreCompleteEvent: {
                setRestoreStage((p: RestoreProgress) => {
                    p.progressCurrent = 1
                    p.progressTotal = 1
                    p.timestamp = null
                    return { ...p }
                })
                refetchRemotes()
                break
            }

            case WsMsgEvent.RestoreFailedEvent: {
                setRestoreStage((p) => {
                    p.error = msgData.content.error
                    return { ...p }
                })
                break
            }

            case WsMsgEvent.BackupProgressEvent: {
                const stages: TaskStageT[] = msgData.content.stages
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog || !stages[0].finished) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    prog.stages = [...stages]
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                break
            }

            case WsMsgEvent.BackupCompleteEvent: {
                const stages: TaskStageT[] = msgData.content.stages
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    prog.stages = [...stages]
                    prog.totalTime = msgData.content.totalTime
                    prog.files.clear()
                    prog.progressCurrent = prog.progressTotal
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                refetchRemotes()
                break
            }

            case WsMsgEvent.CopyFileStartedEvent: {
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    prog.files.set(msgData.content.filename, {
                        name: msgData.content.filename,
                        start: new Date(msgData.content.timestamp),
                    })
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                break
            }

            case WsMsgEvent.BackupFailedEvent: {
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    prog.error = msgData.content.error
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                break
            }

            case WsMsgEvent.CopyFileCompleteEvent: {
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    prog.progressCurrent = msgData.content.tasksComplete
                    prog.progressTotal = msgData.content.tasksTotal
                    prog.files.delete(msgData.content.filename)
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                break
            }

            case WsMsgEvent.CopyFileFailedEvent: {
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    // prog.error = msgData.content.error
                    prog.files.delete(msgData.content.filename)
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                break
            }

            case WsMsgEvent.TaskCreatedEvent: {
                if (msgData.content.taskType === 'do_backup') {
                    // Reset backup progress when a new backup task is created
                    setBackupProgress((p) => {
                        const prog = { files: new Map() } as BackupProgressT
                        p.set(msgData.content.coreId, prog)
                        return new Map(p)
                    })
                }
                break
            }

            case WsMsgEvent.RemoteConnectionChangedEvent: {
                refetchRemotes()
                break
            }

            case WsMsgEvent.PoolCreatedEvent:
            case WsMsgEvent.WeblensLoadedEvent: {
                break
            }

            default: {
                console.error('Unknown eventTag: ' + msgData.eventTag)
            }
        }
    }
}
