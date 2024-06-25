import { Text } from '@mantine/core'
import { IconX } from '@tabler/icons-react'
import { useContext, useMemo, useState } from 'react'

import { WeblensProgress } from '../../components/WeblensProgress'
import { FbContext } from '../../Files/filesContext'
import { nsToHumanTime } from '../../util'
import WeblensButton from '../../components/WeblensButton'
import { WebsocketContext } from '../../Context'

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

    constructor(serverId: string, taskType: string, target: string) {
        if (!serverId || !target || !taskType) {
            console.error(
                'TaskId:',
                serverId,
                'Target:',
                target,
                'Task Type:',
                taskType
            )
            throw new Error('Empty prop in TaskProgress constructor')
        }

        this.taskId = serverId
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
    const wsSend = useContext(WebsocketContext)

    const [cancelWarning, setCancelWarning] = useState(false)

    if (typeof prog.tasksTotal === 'string') {
        console.log(
            prog.tasksComplete,
            prog.tasksTotal,
            parseInt(prog.tasksTotal)
        )
    }

    let totalNum
    if (typeof prog.tasksTotal === 'string') {
        totalNum = parseInt(prog.tasksTotal)
    } else {
        totalNum = prog.tasksTotal
    }

    return (
        <div className="task-progress-box">
            <div className="flex flex-row w-full max-w-full overflow-hidden h-max items-center">
                <div className="flex shrink grow flex-col w-36">
                    <p className="text-sm select-none text-nowrap">
                        {prog.GetTaskType()}
                    </p>
                    <p className="text-lg font-semibold select-none text-nowrap truncate">
                        {prog.target}
                    </p>
                </div>
                <div className="flex w-[30px] min-w-max max-w-[30px] justify-end">
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
                                fbDispatch({
                                    type: 'remove_task_progress',
                                    serverId: prog.taskId,
                                })
                                prog.hide()
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
            <div className="h-6 shrink-0 w-full m-3">
                <WeblensProgress
                    value={prog.getProgress()}
                    complete={prog.stage === TaskStage.Complete}
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
