import { IconX } from '@tabler/icons-react'
import { useWebsocketStore } from '@weblens/api/Websocket'
import WeblensButton from '@weblens/lib/WeblensButton'

import WeblensProgress from '@weblens/lib/WeblensProgress'
import { useMemo, useState } from 'react'
import {
    TaskProgress,
    TaskStage,
    TaskType,
    useTaskState,
} from './TaskStateControl'

export const TasksDisplay = () => {
    const tasks = useTaskState((state) => state.tasks)
    const clearTasks = useTaskState((state) => state.clearTasks)

    const cards = useMemo(() => {
        return Array.from(tasks.values()).map((sp) => {
            if (sp.taskType === TaskType.Backup) {
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
                        <p className="select-none text-nowrap truncate max-w-full">
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
