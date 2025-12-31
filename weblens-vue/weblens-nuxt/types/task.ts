import type WeblensFile from './weblensFile'

export enum TaskType {
    ScanDirectory = 'scan_directory',
    CreateZip = 'create_zip',
}

export enum TaskStatus {
    Pending = 'pending',
    InProgress = 'in-progress',
    Completed = 'completed',
    Canceled = 'canceled',
    Failed = 'failed',
}

type TaskPayloads = {
    [TaskType.ScanDirectory]: {
        taskType: TaskType.ScanDirectory
        percentComplete: number
        countComplete: number
        countTotal: number
        target: WeblensFile
        executionTime?: number
        tasksFailed?: number
    }
    [TaskType.CreateZip]: {
        taskType: TaskType.CreateZip
        bytesSoFar: number
        bytesTotal: number
        completedFiles: number
        speedBytes: number
        totalFiles: number
    }
}

export type TaskParams<T extends TaskType = TaskType> = { taskID: string } & TaskPayloads[T]

export class Task<T extends TaskType = TaskType> {
    private _params!: TaskParams<T>

    taskID: string
    taskType: T

    targetFile?: WeblensFile
    _executionTime: number = 0
    percentComplete: number = 0
    countComplete: number = 0
    countTotal: number = 0
    bytesSoFar: number = 0
    bytesTotal: number = 0
    speedBytes: number = 0

    private _status: TaskStatus = TaskStatus.Pending

    constructor(params: TaskParams<T>) {
        this.taskID = params.taskID
        this.taskType = params.taskType as T

        this.updateProgress(params)
    }

    public isScanDirectoryTask(): this is Task<TaskType.ScanDirectory> {
        return this.taskType === TaskType.ScanDirectory
    }

    public isCreateZipTask(): this is Task<TaskType.CreateZip> {
        return this.taskType === TaskType.CreateZip
    }

    public updateProgress(params: TaskParams<T>) {
        this._params = params

        if (this.isScanDirectoryTask()) {
            if (this._params.target) {
                this.targetFile = this._params.target
            }

            if (this._params.percentComplete !== undefined) {
                this.percentComplete = this._params.percentComplete
            } else if (this._params.countTotal) {
                this.percentComplete = (this._params.countComplete / this._params.countTotal) * 100
            }

            if (this._params.executionTime !== undefined) {
                this._executionTime = this._params.executionTime
            }

            this.countComplete = this._params.countComplete || this.countComplete
            this.countTotal = this._params.countTotal || this.countTotal
        } else if (this.isCreateZipTask()) {
            if (this._params.bytesTotal) {
                this.percentComplete = (this._params.bytesSoFar / this._params.bytesTotal) * 100
            }
        }

        if (this.percentComplete !== 0 && this.percentComplete < 100) {
            this._status = TaskStatus.InProgress
        }
    }

    public setCanceled() {
        this._status = TaskStatus.Canceled
    }

    public setFailed(opts?: { tasksFailed?: number }) {
        this._status = TaskStatus.Failed
        if (this.isScanDirectoryTask() && opts?.tasksFailed !== undefined) {
            this._params.tasksFailed = opts.tasksFailed
        }
    }

    public setComplete() {
        this._status = TaskStatus.Completed
        this.percentComplete = 100
    }

    public setExeTime(executionTime: number) {
        this._executionTime = executionTime
    }

    public get status(): TaskStatus {
        return this._status
    }

    public get isRunning(): boolean {
        return this._status === TaskStatus.InProgress || this._status === TaskStatus.Pending
    }

    public get failCount(): number {
        return (this._params as { tasksFailed: number }).tasksFailed ?? 0
    }

    executionTime(this: Task<TaskType.ScanDirectory>): number
    executionTime(this: Task<Exclude<TaskType, TaskType.ScanDirectory>>): undefined
    executionTime(this: Task<TaskType>): number | undefined {
        if (!this.isScanDirectoryTask()) return undefined
        return this._executionTime
    }
}
