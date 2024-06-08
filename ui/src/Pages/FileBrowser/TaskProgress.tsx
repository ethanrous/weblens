import { useContext, useMemo } from 'react'
import { nsToHumanTime } from '../../util'
import { Text } from '@mantine/core'
import { IconX } from '@tabler/icons-react'
import { WeblensProgress } from '../../components/WeblensProgress'
import { FbContext } from './FileBrowser'

export enum TaskStage {
    Queued,
    InProgress,
    Complete,
    Failure,
}

export class TaskProgress {
    taskId: string
    taskType: string
    target: string
    workingOn: string
    note: string

    timeNs: number
    progressPercent: number
    tasksComplete: number
    tasksTotal: number

    stage: TaskStage

    hidden: boolean

    constructor(taskId: string, taskType: string, target: string) {
        if (!taskId || !target || !taskType) {
            console.error(
                'TaskId:',
                taskId,
                'Target:',
                target,
                'Task Type:',
                taskType
            )
            throw new Error('Empty prop in TaskProgress constructor')
        }

        this.taskId = taskId
        this.taskType = taskType
        this.target = target

        this.timeNs = 0
        this.progressPercent = 0

        this.stage = TaskStage.Queued

        this.hidden = false
    }

    GetTaskId(): string {
        return this.taskId
    }

    GetTaskType(): string {
        switch (this.taskType) {
            case 'scan_directory':
                return 'Folder Scan'
        }
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

export const TasksDisplay = ({
    scanProgress,
}: {
    scanProgress: TaskProgress[]
}) => {
    if (scanProgress.length == 0) {
        return null
    }
    const cards = useMemo(() => {
        return scanProgress.map((sp) => {
            if (sp.hidden) {
                return null
            }
            return <TaskProgCard key={sp.taskId} prog={sp} />
        })
    }, [scanProgress])
    return <div className="h-max w-full">{cards}</div>
}

const TaskProgCard = ({ prog }: { prog: TaskProgress }) => {
    const { fbDispatch } = useContext(FbContext)

    return (
        <div className="task-progress-box">
            <div className="flex flex-row w-full h-max">
                <div className="w-full">
                    <Text size="12px" style={{ userSelect: 'none' }}>
                        {prog.GetTaskType()}
                    </Text>
                    <Text size="16px" fw={600} style={{ userSelect: 'none' }}>
                        {prog.target}
                    </Text>
                </div>
                <IconX
                    size={20}
                    cursor={'pointer'}
                    onClick={() => {
                        prog.hide()
                        fbDispatch({
                            type: 'remove_task_progress',
                            taskId: prog.taskId,
                        })
                    }}
                />
            </div>
            <div className="h-6 shrink-0 w-full m-3">
                <WeblensProgress
                    value={prog.getProgress()}
                    complete={prog.stage === TaskStage.Complete}
                    loading={prog.stage === TaskStage.Queued}
                    failure={prog.stage === TaskStage.Failure}
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
                    {prog.tasksTotal > 0 && (
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
                {/* <Text
                    size="10px"
                    style={{
                        width: "max-content",
                        textWrap: "nowrap",
                        userSelect: "none",
                    }}
                >
                    {prog.note}
                </Text> */}
            </div>
        </div>
    )
}
