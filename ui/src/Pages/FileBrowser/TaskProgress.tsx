import { Text } from '@mantine/core'
import { IconX } from '@tabler/icons-react'
import { Dispatch, useContext, useMemo, useState } from 'react'

import { WeblensProgress } from '../../components/WeblensProgress'
import { TaskProgContext } from '../../Files/FBTypes'
import { nsToHumanTime } from '../../util'
import WeblensButton from '../../components/WeblensButton'
import { WebsocketContext } from '../../Context'

export class TasksProgress {
    private tasks: Map<string, TaskProgress>

    constructor(prev?: TasksProgress) {
        if (prev) {
            this.tasks = prev.tasks
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
            return
        }

        if (stage === TaskStage.Complete && task.taskType === 'download_file') {
            this.removeTask(taskId)
            return
        }

        task.setTaskStage(stage)
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
        progress?: number,
        completeCount?: number | string,
        totalCount?: number | string
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

    timeNs: number
    progressPercent: number
    tasksComplete: number | string
    tasksTotal: number | string

    stage: TaskStage

    hidden: boolean

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

    FormatTaskType(): string {
        switch (this.taskType) {
            case 'scan_directory':
                return 'Folder Scan'
            case 'create_zip':
                return 'Zip'
            case 'download_file':
                return 'Download'
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
        return this.progressPercent
    }

    hide(): void {
        this.hidden = true
    }
}

export const TasksDisplay = () => {
    const { progState, progDispatch } = useContext(TaskProgContext)
    const cards = useMemo(() => {
        return progState.getTasks().map((sp) => {
            if (sp.taskType === 'do_backup') {
                return null
            }
            return <TaskProgCard key={sp.taskId} prog={sp} />
        })
    }, [progState])

    if (progState.tasksCount() == 0) {
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
                    progDispatch({ type: 'clear_tasks' })
                }}
                disabled={progState.tasksCount() === 0}
            />
            <div className="flex shrink w-full overflow-y-scroll h-full pb-4 no-scrollbar">
                <div className="flex flex-col h-max w-full">{cards}</div>
            </div>
        </div>
    )
}

const TaskProgCard = ({ prog }: { prog: TaskProgress }) => {
    const { progDispatch } = useContext(TaskProgContext)
    const wsSend = useContext(WebsocketContext)

    const [cancelWarning, setCancelWarning] = useState(false)

    let totalNum
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
                        <p className="text-sm select-none text-nowrap">
                            {prog.FormatTaskType()}
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
                                    progDispatch({
                                        type: 'remove_task_progress',
                                        taskId: prog.taskId,
                                    })
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
                    <p className="text-lg font-semibold select-none text-nowrap truncate">
                        {prog.target}
                    </p>
                </div>
            </div>
            <div className="relative h-6 shrink-0 w-full m-3">
                <WeblensProgress
                    value={prog.getProgress()}
                    // complete={prog.stage === TaskStage.Complete}
                    loading={prog.stage === TaskStage.Queued}
                    failure={
                        prog.stage === TaskStage.Failure ||
                        prog.stage === TaskStage.Cancelled
                    }
                />
            </div>
            {prog.stage !== TaskStage.Complete && (
                <div className="flex flex-row w-full justify-between h-max gap-3">
                    <Text
                        size="10px"
                        truncate="end"
                        style={{ userSelect: 'none' }}
                    >
                        {prog.workingOn}
                    </Text>
                    {totalNum > 0 && (
                        <Text size="10px" style={{ userSelect: 'none' }}>
                            {prog.tasksComplete}/{prog.tasksTotal}
                        </Text>
                    )}
                </div>
            )}

            <div className="flex flex-row w-full justify-between h-max gap-3">
                {prog.stage === TaskStage.Complete && (
                    <Text
                        size="10px"
                        style={{ width: 'max-content', userSelect: 'none' }}
                    >
                        Finished{' '}
                        {prog.timeNs !== 0 ? `in ${prog.getTime()}` : ''}
                    </Text>
                )}
                {prog.stage === TaskStage.Queued && (
                    <Text
                        size="10px"
                        style={{ width: 'max-content', userSelect: 'none' }}
                    >
                        Queued...
                    </Text>
                )}
                {prog.stage === TaskStage.Failure && (
                    <Text
                        size="10px"
                        style={{ width: 'max-content', userSelect: 'none' }}
                    >
                        {prog.note}
                    </Text>
                )}
            </div>
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
    tasksTotal?: number | string
}

export type TasksProgressDispatch = Dispatch<TasksProgressAction>

export function taskProgressReducer(
    state: TasksProgress,
    action: TasksProgressAction
): TasksProgress {
    switch (action.type) {
        case 'new_task': {
            const prog = new TaskProgress(action.taskId, action.taskType)
            prog.setTarget(action.target)
            state.addTask(prog)

            break
        }
        case 'task_complete': {
            state.setTaskStage(action.taskId, TaskStage.Complete)

            if (action.time) {
                state.setTaskTime(action.taskId, action.time)
            }
            if (action.note) {
                state.setTaskNote(action.taskId, action.note)
            }

            break
        }

        case 'task_failure': {
            state.setTaskStage(action.taskId, TaskStage.Failure)
            if (action.note) {
                state.setTaskNote(action.taskId, action.note)
            }

            break
        }

        case 'task_cancelled': {
            state.setTaskStage(action.taskId, TaskStage.Cancelled)
            break
        }

        case 'update_scan_progress': {
            if (!state.has(action.taskId)) {
                state.addTask(new TaskProgress(action.taskId, action.taskType))
                console.error(
                    'Trying to update scan progress on task that does not exist [' +
                        action.taskId +
                        ']'
                )
                return state
                // task = new TaskProgress(action.taskId, action.taskType);
                // state.addTask(task);
            }

            state.setTaskStage(action.taskId, TaskStage.InProgress)

            state.updateTaskProgress(
                action.taskId,
                action.progress,
                action.tasksComplete,
                action.tasksTotal
            )

            state.setWorkingOn(action.taskId, action.workingOn)
            state.setTaskTarget(action.taskId, action.target)
            state.setTaskNote(action.taskId, action.note)

            break
        }

        case 'add_pool_to_progress': {
            if (!state.has(action.taskId)) {
                const newTask = new TaskProgress(action.taskId, action.taskType)
                state.addTask(newTask)
            }
            state.linkPoolToTask(action.taskId, action.poolId)
            break
        }

        case 'remove_task_progress': {
            state.removeTask(action.taskId)
            break
        }

        case 'clear_tasks': {
            for (const task of state.getTasks()) {
                if (task.getTaskStage() === TaskStage.Complete) {
                    state.removeTask(task.GetTaskId())
                }
            }
            break
        }

        case 'refresh': {
            break
        }

        default: {
            console.error(
                'Unknown action type in task progress reducer ',
                action.type
            )
            return state
        }
    }

    return new TasksProgress(state)
}
