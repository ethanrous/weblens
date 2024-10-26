import { WsMsgEvent, wsMsgInfo } from '../../api/Websocket'
import { TaskStageT } from '../FileBrowser/TaskStateControl'

export type RestoreProgress = {
    stage: string
    error: string
    timestamp: Date
    progress_total: number
    progress_current: number
}

type FileBackupProgress = {
    name: string
    start: Date
}

export type BackupProgressT = {
    stages: TaskStageT[]
    error: string
    timestamp: Date
    progress_total: number
    progress_current: number
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
            case 'restore_progress': {
                setRestoreStage((p) => {
                    if (msgData.content.stage) {
                        p.stage = msgData.content.stage
                    }
                    if (msgData.content.files_total) {
                        p.progress_total = msgData.content.files_total
                        p.progress_current = msgData.content.files_restored
                    }
                    if (msgData.content.timestamp) {
                        p.timestamp = new Date(msgData.content.timestamp)
                    }

                    return { ...p }
                })
                break
            }

            case 'restore_complete': {
                setRestoreStage((p: RestoreProgress) => {
                    p.progress_current = 1
                    p.progress_total = 1
                    p.timestamp = null
                    return { ...p }
                })
                refetchRemotes()
                break
            }

            case 'restore_failed': {
                setRestoreStage((p) => {
                    p.error = msgData.content.error
                    return { ...p }
                })
                break
            }

            case 'backup_progress': {
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

            case 'backup_complete': {
                const stages: TaskStageT[] = msgData.content.stages
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    prog.stages = [...stages]
                    prog.totalTime = msgData.content.totalTime
                    prog.files.clear()
                    prog.progress_current = prog.progress_total
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                refetchRemotes()
                break
            }

            case 'copy_file_started': {
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

            case 'backup_failed': {
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

            case 'copy_file_complete': {
                setBackupProgress((p) => {
                    let prog = p.get(msgData.content.coreId)
                    if (!prog) {
                        prog = { files: new Map() } as BackupProgressT
                    }
                    prog.progress_current = msgData.content.tasks_complete
                    prog.progress_total = msgData.content.tasks_total
                    prog.files.delete(msgData.content.filename)
                    p.set(msgData.content.coreId, prog)
                    return new Map(p)
                })
                break
            }

            case 'copy_file_failed': {
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
