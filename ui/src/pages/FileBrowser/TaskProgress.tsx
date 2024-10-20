import { IconX } from '@tabler/icons-react'
import { useWebsocketStore } from '@weblens/api/Websocket'
import WeblensButton from '@weblens/lib/WeblensButton'

import WeblensProgress from '@weblens/lib/WeblensProgress'
import { nsToHumanTime } from '@weblens/util'
import { useContext, useMemo, useState } from 'react'
import { create, StateCreator } from 'zustand'

export type TaskStageT = {
    key: string
    name: string
    started: number
    finished: number
}

export class TaskProgressState {
    private tasks: Map<string, TaskProgress>

    constructor(prev?: TaskProgressState) {
        if (prev) {
            this.tasks = new Map<string, TaskProgress>(prev.tasks)
        } else {
            this.tasks = new Map<string, TaskProgress>()
        }
    }

    public has(taskId: string): boolean {
        return this.tasks.has(taskId)
    }

    public tasksCount(): number {
        return Array.from(this.tasks.values()).filter((t) => {
            return t.taskType !== 'do_backup'
        }).length
    }

    public getTasks(): TaskProgress[] {
        const retTasks = []
        this.tasks.forEach((task, taskId) => {
            if (taskId === task.GetTaskId()) {
                retTasks.push(task)
            }
        })
        return retTasks
    }

    public addTask(task: TaskProgress) {
        this.tasks.set(task.GetTaskId(), task)
    }

    public getTask(taskId: string): TaskProgress {
        return this.tasks.get(taskId)
    }

    public setTaskStage(taskId: string, stage: TaskStage): void {
        const task = this.tasks.get(taskId)
        if (!task) {
            console.error('Could not find task to set stage', taskId)
            return
        }

        if (stage === TaskStage.Complete && task.taskType === 'download_file') {
            task.hide()
        }

        task.setTaskStage(stage)
        this.tasks.set(taskId, task)
    }

    public setTaskTime(taskId: string, taskTimeNs: number) {
        const task = this.tasks.get(taskId)
        if (!task) {
            return
        }
        task.timeNs = taskTimeNs
    }

    public setTaskNote(taskId: string, note: string) {
        if (!note) {
            return
        }

        const task = this.tasks.get(taskId)
        if (!task) {
            return
        }
        task.note = note
    }

    public setTaskTarget(taskId: string, target: string) {
        if (!target) {
            return
        }

        const task = this.tasks.get(taskId)
        if (!task) {
            return
        }
        task.setTarget(target)
    }

    public linkPoolToTask(taskId: string, poolId: string): void {
        if (!this.tasks.has(taskId)) {
            console.error(
                'Trying to link pool [' +
                    poolId +
                    '] to task with id [' +
                    taskId +
                    '] but task does not exist'
            )
            return
        }

        this.tasks.set(poolId, this.tasks.get(taskId))
    }

    updateTaskProgress(
        taskId: string,
        progress: number,
        completeCount: number | string,
        failedCount: number | string,
        totalCount: number | string
    ): void {
        const task = this.tasks.get(taskId)
        if (!task) {
            return
        }

        if (progress !== undefined) {
            task.progressPercent = progress
        }

        if (
            completeCount !== undefined &&
            (task.tasksTotal || totalCount !== undefined)
        ) {
            task.tasksComplete = completeCount
            if (totalCount !== undefined) {
                task.tasksTotal = totalCount
            }
        }

        if (failedCount) {
            task.tasksFailed = failedCount
        }
    }

    public setWorkingOn(taskId: string, workingOn: string): void {
        if (!workingOn) {
            return
        }

        const task = this.tasks.get(taskId)
        if (!task) {
            return
        }

        task.workingOn = workingOn
    }

    public removeTask(removeTaskId: string): void {
        const tasks = Array.from(this.tasks.entries())
        this.tasks.clear()
        for (const [tId, task] of tasks) {
            if (task.GetTaskId() !== removeTaskId) {
                this.tasks.set(tId, task)
            }
        }
    }
}

export enum TaskStage {
    Queued,
    InProgress,
    Complete,
    Cancelled,
    Failure,
}

export class TaskProgress {
    taskId: string
    poolId: string
    taskType: string
    target: string
    workingOn: string
    note: string
    error: string

    timeNs: number
    progressPercent: number
    tasksComplete: number | string
    tasksFailed: number | string
    tasksTotal: number | string

    stage: TaskStage

    hidden: boolean

    progMeta

    constructor(serverId: string, taskType: string) {
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
    }

    GetTaskId(): string {
        return this.taskId
    }

    FormatTaskName(): string {
        switch (this.taskType) {
            case 'scan_directory':
                return `Import ${this.target ? this.target : 'folder'}`
            case 'create_zip':
                return `Zip ${this.target ? this.target : ''}`
            case 'download_file':
                return `Download ${this.target ? this.target : ''}`
        }
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

export const TasksDisplay = () => {
    const tasks = useTaskState((state) => state.tasks)
    const clearTasks = useTaskState((state) => state.clearTasks)

    const cards = useMemo(() => {
        return Array.from(tasks.values()).map((sp) => {
            if (sp.taskType === 'do_backup') {
                return null
            }
            return <TaskProgCard key={sp.taskId} prog={sp} />
        })
    }, [tasks])

    if (tasks.size === 0) {
        return null
    }

    return (
        <div className="flex flex-col relative h-max w-full pt-4">
            <WeblensButton
                label={'Clear Tasks'}
                subtle
                centerContent
                fillWidth
                squareSize={32}
                onClick={() => {
                    clearTasks()
                }}
            />
            <div className="flex shrink w-full overflow-y-scroll h-full pb-4 pt-2 no-scrollbar">
                <div className="flex flex-col h-max w-full gap-2">{cards}</div>
            </div>
        </div>
    )
}

const TaskProgCard = ({ prog }: { prog: TaskProgress }) => {
    const removeTask = useTaskState((state) => state.removeTask)
    const wsSend = useWebsocketStore((state) => state.wsSend)

    const [cancelWarning, setCancelWarning] = useState(false)

    let totalNum: number
    if (typeof prog.tasksTotal === 'string') {
        totalNum = parseInt(prog.tasksTotal)
    } else {
        totalNum = prog.tasksTotal
    }

    return (
        <div
            className="task-progress-box animate-fade"
            data-hidden={prog.hidden}
        >
            <div className="flex flex-row w-full max-w-full h-max items-center">
                <div className="flex flex-col w-full">
                    <div className="flex flex-row justify-between items-center w-full h-max">
                        <p className="select-none text-nowrap truncaate">
                            {prog.FormatTaskName()}
                        </p>
                        <WeblensButton
                            Right={IconX}
                            subtle
                            centerContent
                            label={cancelWarning ? 'Cancel?' : ''}
                            danger={cancelWarning}
                            squareSize={30}
                            onClick={() => {
                                if (
                                    prog.stage === TaskStage.Complete ||
                                    prog.stage === TaskStage.Cancelled ||
                                    prog.stage === TaskStage.Failure
                                ) {
                                    removeTask(prog.taskId)
                                    return
                                }
                                if (cancelWarning) {
                                    wsSend('cancel_task', {
                                        taskPoolId: prog.poolId
                                            ? prog.poolId
                                            : prog.taskId,
                                    })

                                    setCancelWarning(false)
                                } else {
                                    setCancelWarning(true)
                                }
                            }}
                        />
                    </div>
                </div>
            </div>
            <div className="relative h-6 shrink-0 w-full m-2">
                <WeblensProgress
                    value={prog.getProgress()}
                    secondaryValue={prog.getErrorProgress()}
                    loading={prog.stage === TaskStage.Queued}
                    secondaryColor={'red'}
                />
            </div>
            {prog.stage !== TaskStage.Complete && (
                <div className="flex flex-row w-full justify-between h-max gap-3 m-2">
                    <p className="text-sm select-none text-nowrap truncate">
                        {prog.workingOn}
                    </p>
                    {totalNum > 0 && (
                        <p className="text-sm select-none">
                            {prog.tasksComplete}/{prog.tasksTotal}
                        </p>
                    )}
                </div>
            )}

            {prog.stage !== TaskStage.InProgress && (
                <div className="flex flex-row w-full justify-between h-max gap-3 mt-2">
                    {prog.stage === TaskStage.Complete && (
                        <p className="text-sm w-max select-none truncate">
                            Finished{' '}
                            {prog.timeNs !== 0 ? `in ${prog.getTime()}` : ''}
                        </p>
                    )}
                    {prog.stage === TaskStage.Queued && (
                        <p className="text-sm w-max select-none truncate">
                            Queued...
                        </p>
                    )}
                    {prog.stage === TaskStage.Failure && (
                        <p className="text-sm w-max select-none truncate">
                            {prog.note}
                        </p>
                    )}
                </div>
            )}
        </div>
    )
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

enum TaskType {
    ScanDirectory = 'scan_directory',
    CreateZip = 'create_zip',
    DownloadFile = 'download_file',
}

type TaskStateT = {
    tasks: Map<string, TaskProgress>

    addTask: (taskId: string, taskType: string, opts?: NewTaskOptions) => void
    removeTask: (taskId: string) => void
    clearTasks: () => void
    updateTaskProgress: (taskId: string, opts) => void
    handleTaskCompete: (taskId: string, time: number, note: string) => void
    handleTaskFailure: (taskId: string, error: string) => void
    handleTaskCancel: (taskId: string) => void
}

const TaskStateControl: StateCreator<TaskStateT, [], []> = (set) => ({
    tasks: new Map<string, TaskProgress>(),

    addTask: (taskId: string, taskType: string, opts?: NewTaskOptions) => {
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
            task.stage = TaskStage.Complete

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

            task.stage = TaskStage.Failure
            task.error = error

            state.tasks.set(taskId, task)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },

    handleTaskCancel: (taskId: string) => {
        console.error('handleTaskCancel not impl')
    },

    updateTaskProgress: (taskId: string, opts) => {
        set((state) => {
            const task = state.tasks.get(taskId)
            if (!task) {
                console.error('Could not find task to update progress', taskId)
                return state
            }

            task.progressPercent = opts.progress
            task.stage = TaskStage.InProgress
            switch (task.taskType) {
                case TaskType.ScanDirectory:
                case TaskType.CreateZip:
                    task.workingOn = opts.workingOn
                    task.tasksComplete = opts.tasksComplete
                    task.tasksTotal = opts.tasksTotal
            }

            task.progMeta = { ...task.progMeta, ...opts }

            state.tasks.set(taskId, task)
            return { tasks: new Map<string, TaskProgress>(state.tasks) }
        })
    },
})

export const useTaskState = create<TaskStateT>()(TaskStateControl)
