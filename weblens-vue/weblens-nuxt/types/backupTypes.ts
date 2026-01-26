import type { WsMessage } from './websocket'

export interface BackupInfoWSMsg {
    backupSize: number
    totalBytesCopied: number
    totalBytesToCopy: number
    totalTime: number
    taskStartTime: number
}

export class BackupInfo {
    TowerID: string = ''
    BytesSoFar: number = 0
    TotalBytes: number = 0
    StartTime: number = 0
    EndTime: number = 0
    Completed: boolean = false

    static fromWsMsg(towerID: string, msg: WsMessage<BackupInfoWSMsg>): BackupInfo {
        const info = new BackupInfo()
        info.TowerID = towerID
        info.BytesSoFar = msg.content.totalBytesCopied
        info.TotalBytes = msg.content.totalBytesToCopy
        info.StartTime = msg.taskStartTime || 0
        info.EndTime = info.Completed ? info.StartTime + msg.content.totalTime : 0
        info.Completed = msg.content.totalBytesCopied >= msg.content.totalBytesToCopy
        return info
    }

    clone(): BackupInfo {
        const copy = new BackupInfo()
        copy.TowerID = this.TowerID
        copy.BytesSoFar = this.BytesSoFar
        copy.TotalBytes = this.TotalBytes
        copy.StartTime = this.StartTime
        copy.EndTime = this.EndTime
        copy.Completed = this.Completed
        return copy
    }

    merge(other: BackupInfo) {
        if (this.TowerID !== other.TowerID) {
            throw new Error('Cannot merge BackupInfo with different TowerIDs')
        }
        this.BytesSoFar = other.BytesSoFar
        this.TotalBytes = other.TotalBytes
        this.StartTime = other.StartTime
        this.EndTime = other.EndTime
        this.Completed = this.Completed || other.Completed
    }

    complete(endTime: number) {
        this.Completed = true
        this.EndTime = endTime
    }
}
