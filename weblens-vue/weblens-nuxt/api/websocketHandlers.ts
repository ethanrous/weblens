import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'
import { BackupInfo } from '~/types/backupTypes'
import { TaskType, type TaskParams } from '~/types/task'
import WeblensFile from '~/types/weblensFile'
import { WsEvent, type WsMessage } from '~/types/websocket'

function handleModified(msg: WsMessage) {
    if (msg.content.fileInfo.id === useLocationStore().activeFolderID) {
        return
    }

    if (msg.content.fileInfo.parentID !== useLocationStore().activeFolderID) {
        useFilesStore().removeFiles(msg.content.fileInfo.id)
    } else {
        useFilesStore().addFile(msg.content.fileInfo)
    }
}

export function handleWebsocketMessage(msg: WsMessage) {
    if (msg.error) {
        console.error('WebSocket error:', msg.error)
        return
    }

    if (msg.eventTag.toLowerCase().includes('fail')) {
        console.error('WebSocket message received at', new Date(msg.sentTime).toISOString(), ':', msg, 'at')
    } else {
        console.debug('WebSocket message received at', new Date(msg.sentTime).toISOString(), ':', msg, 'at')
    }

    switch (msg.eventTag) {
        case WsEvent.FileCreatedEvent:
            useFilesStore().addFile(msg.content.fileInfo)
            break

        case WsEvent.FileUpdatedEvent:
            handleModified(msg)
            break

        case WsEvent.FileDeletedEvent:
            useFilesStore().removeFiles(msg.content.fileInfo.id)
            break

        case WsEvent.TaskCreatedEvent: {
            let targetFile: WeblensFile | undefined
            let taskParams: TaskParams
            if (msg.taskType === TaskType.ScanDirectory) {
                targetFile = new WeblensFile({ portablePath: msg.content.filename, isDir: true })
                taskParams = {
                    taskID: msg.subscribeKey,
                    taskType: TaskType.ScanDirectory,

                    percentComplete: 0,
                    target: targetFile,
                    countComplete: 0,
                    countTotal: 0,
                }
            } else if (msg.taskType === TaskType.CreateZip) {
                taskParams = {
                    taskID: msg.subscribeKey,
                    taskType: TaskType.CreateZip,

                    bytesSoFar: 0,
                    bytesTotal: 0,
                    completedFiles: 0,
                    speedBytes: 0,
                    totalFiles: 0,
                }
            } else {
                console.warn('Unknown task type for TaskCreatedEvent:', msg.taskType)
                break
            }

            useTasksStore().upsertTask(msg.subscribeKey, taskParams)
            break
        }

        case WsEvent.TaskCanceledEvent: {
            useTasksStore().cancelTask(msg.subscribeKey)

            break
        }

        case WsEvent.TaskFailedEvent: {
            useTasksStore().failTask(msg.subscribeKey, { tasksFailed: msg.content.failedCount })

            break
        }

        case WsEvent.FileScanStartedEvent:
        case WsEvent.FileScanFailedEvent:
        case WsEvent.FileScanCompleteEvent: {
            const targetFile = new WeblensFile({ portablePath: msg.content.filename, isDir: true })

            useTasksStore().upsertTask(msg.subscribeKey, {
                taskID: msg.subscribeKey,
                taskType: TaskType.ScanDirectory,

                percentComplete: msg.content.percentProgress,
                countComplete: msg.content.tasksComplete,
                countTotal: msg.content.tasksTotal,
                tasksFailed: msg.content.tasksFailed,

                target: targetFile,
            })
            break
        }

        case WsEvent.FolderScanCompleteEvent: {
            const targetFile = new WeblensFile({ portablePath: msg.content.filename, isDir: true })

            useTasksStore().upsertTask(msg.subscribeKey, {
                taskID: msg.subscribeKey,
                taskType: TaskType.ScanDirectory,

                percentComplete: msg.content.percentProgress,
                countComplete: msg.content.tasksComplete,
                countTotal: msg.content.tasksTotal,
                target: targetFile,
                executionTime: msg.content.runtime,
            })

            useTasksStore().setTaskComplete(msg.subscribeKey, msg.content)
            break
        }

        case WsEvent.ZipProgressEvent: {
            useTasksStore().upsertTask(msg.subscribeKey, {
                taskID: msg.subscribeKey,
                taskType: TaskType.CreateZip,

                bytesSoFar: msg.content.bytesSoFar,
                bytesTotal: msg.content.bytesTotal,
                completedFiles: msg.content.completedFiles,
                speedBytes: msg.content.speedBytes,
                totalFiles: msg.content.totalFiles,
            })

            break
        }

        case WsEvent.ZipCompleteEvent: {
            useTasksStore().setTaskComplete(msg.subscribeKey, msg.content)

            break
        }

        case WsEvent.RemoteConnectionChangedEvent: {
            useRemotesStore().refreshRemotes()

            break
        }

        case WsEvent.BackupStartedEvent:
        case WsEvent.BackupProgressEvent: {
            if (!msg.content['coreID']) {
                console.error('Backup event missing towerID in content:', msg)
                break
            }

            const backupInfo = BackupInfo.fromWsMsg(msg.content['coreID'], msg)
            useBackupStore().updateBackup(backupInfo)

            break
        }

        case WsEvent.BackupCompleteEvent: {
            if (!msg.content['coreID']) {
                console.error('Backup complete event missing towerID in content:', msg)
                break
            }

            useBackupStore().setBackupComplete(msg.content['coreID'], msg.constructedTime)

            break
        }

        default:
            console.warn('Unhandled WebSocket event:', msg.eventTag)
    }
}
