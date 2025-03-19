import { nsToHumanTime } from '@weblens/util'
import { StateCreator, create } from 'zustand'

export enum TaskStage {
    Queued,
    InProgress,
    Complete,
    Cancelled,
    Failure,
}

export type TaskStageT = {
    key: string
    name: string
    started: number
    finished: number
}

type SubTaskInfo = {
    itemId: string
    itemName: string
    startTimeEpochMs: number

    elapsedTimeMs?: number
}

export class TaskProgress {
    taskId: string
    poolId: string
    taskType: TaskType
    target: string
    workingOn: SubTaskInfo[]
    note: string
    error: string

    timeNs: number
    progressPercent: number
    tasksComplete: number | string
    tasksFailed: number | string
    tasksTotal: number | string

    stage: TaskStage

    hidden: boolean

    progMeta: TaskUpdateProps

    constructor(serverId: string, taskType: TaskType) {
        if (!serverId || !taskType) {
            console.error('TaskId:', serverId, 'Task Type:', taskType)
            throw new Error('Empty prop in TaskProgress constructor')
        }

        this.taskId = serverId
        this.taskType = taskType

        this.timeNs = 0
        this.progressPercent = 0

        this.stage = TaskStage.Queued

        this.hidden = false
        this.workingOn = []
    }

    GetTaskId(): string {
        return this.taskId
    }

    FormatTaskType(): string {
        switch (this.taskType) {
            case TaskType.ScanDirectory:
                return 'Scan'
            case TaskType.CreateZip:
                return 'Zip'
            case TaskType.DownloadFile:
                return 'Download'
        }
    }

    FormatTaskName(): string {
        const name = this.FormatTaskType()
        if (this.target) {
            return `${name} ${this.target}`
        } else if (this.taskType === TaskType.ScanDirectory) {
            return `${name} folder`
        }
        return name
    }

    getTaskStage(): TaskStage {
        return this.stage
    }

    setTaskStage(newStage: TaskStage) {
        if (newStage < this.stage) {
            return
        }
        this.stage = newStage
    }

    setTarget(target: string): void {
        this.target = target
    }

    getTasksComplete(): number {
        return Number(this.tasksComplete)
    }

    getTasksTotal(): number {
        return Number(this.tasksTotal)
    }

    getTime(): string {
        return nsToHumanTime(this.timeNs)
    }

    // Returns the progress of the task as a percentage
    getProgress(): number {
        if (this.stage === TaskStage.Complete) {
            return 100
        }

        if (this.tasksFailed) {
            const healthyTasks =
                Number(this.tasksComplete) - Number(this.tasksFailed)
            return (healthyTasks * 100) / Number(this.tasksTotal)
        }

        return this.progressPercent
    }

    getErrorProgress(): number {
        if (this.stage === TaskStage.Complete) {
            return 100
        }

        if (this.tasksFailed) {
            return (Number(this.tasksComplete) * 100) / Number(this.tasksTotal)
        }

        return 0
    }

    hide(): void {
        this.hidden = true
    }
}
export type TasksProgressAction = {
    type: string

    taskId?: string
    poolId?: string
    taskType?: string
    note?: string
    workingOn?: string
    target?: string

    time?: number
    progress?: number
    tasksComplete?: number | string
    tasksFailed?: number | string
    tasksTotal?: number | string
}

export type NewTaskOptions = {
    target?: string
    progress?: number
}

export enum TaskType {
    ScanDirectory = 'scan_directory',
    CreateZip = 'create_zip',
    DownloadFile = 'download_file',
    Backup = 'do_backup',
}

type TaskUpdateProps = {
    progress: number
    tasksTotal?: number | string
    tasksComplete?: number | string
    tasksFailed?: number | string
    workingOn?: string
    finished?: string
    itemId?: string
    taskId?: string
    note?: string
    taskType?: TaskType
}

type TaskStateT = {
    tasks: Map<string, TaskProgress>
    showingMenu: boolean

    addTask: (taskId: string, taskType: string, opts?: NewTaskOptions) => void
    removeTask: (taskId: string) => void
    clearTasks: () => void
    updateTaskProgress: (taskId: string, opts: TaskUpdateProps) => void
    handleTaskCompete: (taskId: string, time: number, note: string) => void
    handleTaskFailure: (taskId: string, error: string) => void
    handleTaskCancel: (taskId: string) => void
    setShowingMenu: (showing: boolean) => void
}

const TaskStateControl: StateCreator<TaskStateT, [], []> = (set) => ({
    tasks: new Map<string, TaskProgress>(),
    showingMenu: false,

    addTask: (taskId: string, taskType: TaskType, opts?: NewTaskOptions) => {
        set((state) => {
            const newProg = new TaskProgress(taskId, taskType)

            if (opts?.target) {
                newProg.setTarget(opts?.target)
            }
            if (opts?.progress) {
                newProg.progressPercent = opts?.progress
            }

            state.tasks.set(taskId, newProg)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },

    removeTask: (taskId: string) => {
        set((state) => {
            state.tasks.delete(taskId)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },

    clearTasks: () => {
        set({ tasks: new Map<string, TaskProgress>() })
    },

    handleTaskCompete: (taskId: string, time: number, note: string) => {
        set((state) => {
            const task = state.tasks.get(taskId)
            if (!task) {
                return state
            }

            task.timeNs = time
            task.note = note
            task.setTaskStage(TaskStage.Complete)

            state.tasks.set(taskId, task)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },

    handleTaskFailure: (taskId: string, error: string) => {
        set((state) => {
            const task = state.tasks.get(taskId)
            if (!task) {
                console.error('Could not find task to set failure', taskId)
                return state
            }

            task.setTaskStage(TaskStage.Failure)
            task.error = error

            state.tasks.set(taskId, task)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },

    handleTaskCancel: (taskId: string) => {
        set((state) => {
            const task = state.tasks.get(taskId)
            if (!task) {
                console.error('Could not find task to cancel', taskId)
                return state
            }

            task.setTaskStage(TaskStage.Cancelled)

            state.tasks.set(taskId, task)
            console.log('Task cancelled:', taskId)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },

    updateTaskProgress: (taskId: string, opts) => {
        set((state) => {
            const task = state.tasks.get(taskId)
            if (!task) {
                console.error('Could not find task to update progress', taskId)
                return state
            }

            if (
                task.stage === TaskStage.Complete ||
                task.stage === TaskStage.Cancelled
            ) {
                return state
            }

            task.progressPercent = opts.progress
            task.stage = TaskStage.InProgress
            switch (task.taskType) {
                case TaskType.ScanDirectory:
                    if (opts.workingOn) {
                        task.workingOn.push({
                            itemName: opts.workingOn,
                            itemId: opts.itemId,
                            startTimeEpochMs: Math.floor(
                                performance.timeOrigin + performance.now()
                            ),
                            elapsedTimeMs: 0,
                        })
                    } else {
                        const finishedIndex = task.workingOn.findIndex(
                            (val) => val.itemId === opts.itemId
                        )
                        task.workingOn.splice(finishedIndex, 1)
                    }
                    task.tasksComplete = opts.tasksComplete
                    task.tasksTotal = opts.tasksTotal
                    task.tasksFailed = opts.tasksFailed
                    break

                case TaskType.CreateZip:
                    task.workingOn = [
                        {
                            itemName: opts.workingOn,
                            itemId: '',
                            startTimeEpochMs: Math.floor(
                                performance.timeOrigin + performance.now()
                            ),
                        },
                    ]
                    task.tasksComplete = opts.tasksComplete
                    task.tasksTotal = opts.tasksTotal
                    task.tasksFailed = opts.tasksFailed
            }

            task.progMeta = { ...task.progMeta, ...opts }

            state.tasks.set(taskId, task)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },
    setShowingMenu: (showing: boolean) => {
        set({ showingMenu: showing })
    },
})

export const useTaskState = create<TaskStateT>()(TaskStateControl)
