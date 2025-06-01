import {
    IconChevronLeft,
    IconChevronRight,
    IconExclamationCircle,
    IconX,
} from '@tabler/icons-react'
import { CancelTask } from '@weblens/api/FileBrowserApi'
import LoaderDots from '@weblens/lib/LoaderDots.tsx'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import WeblensProgress from '@weblens/lib/WeblensProgress.tsx'
import { useClick } from '@weblens/lib/hooks'
import { useMessagesController } from '@weblens/store/MessagesController'
import { msToHumanTime } from '@weblens/util'
import { useEffect, useMemo, useRef, useState } from 'react'

import {
    TaskProgress,
    TaskStage,
    useTaskState,
} from '../../store/TaskStateControl'

export function TasksPane({
    calloutDivRef,
    // paneRef,
    setPaneRef,
}: {
    calloutDivRef: HTMLDivElement
    // paneRef: HTMLDivElement
    setPaneRef: (d: HTMLDivElement) => void
}) {
    const tasksMap = useTaskState((state) => state.tasks)
    const clearTasks = useTaskState((state) => state.clearTasks)

    const showingMenu = useTaskState((state) => state.showingMenu)
    const setShowingMenu = useTaskState((state) => state.setShowingMenu)

    useClick(() => {
        useMessagesController.getState().addMessage({
            text: 'Tasks menu closed',
            duration: 5000,
            severity: 'debug',
        })

        if (showingMenu) {
            setShowingMenu(false)
        }
    }, calloutDivRef)

    const sortedTasks = useMemo(() => {
        const sortedTasks = Array.from(tasksMap.values()).sort((a, b) => {
            if (
                a.stage === TaskStage.InProgress &&
                b.stage !== TaskStage.InProgress
            ) {
                return -1
            } else if (
                a.stage !== TaskStage.InProgress &&
                b.stage === TaskStage.InProgress
            ) {
                return 1
            }

            return a.taskId.localeCompare(b.taskId)
        })
        return sortedTasks
    }, [tasksMap])

    return (
        <div
            ref={setPaneRef}
            className="border-color-theme-primary bg-background-primary absolute top-[-100%] left-full z-50 ml-10 flex h-max w-96 max-w-[50vw] grow flex-col rounded-md border p-4 pt-4 shadow-2xl"
        >
            <div className="mb-2 flex flex-row items-center">
                <WeblensButton
                    Left={IconChevronLeft}
                    size="tiny"
                    className="h-max w-max"
                    onClick={() => {
                        setShowingMenu(false)
                    }}
                />
                <h4 className="m-auto">Tasks</h4>
            </div>
            {sortedTasks.length === 0 && (
                <h3 className="my-12 text-center">No tasks running</h3>
            )}
            <div className="no-scrollbar flex h-max max-h-full w-full shrink gap-2 overflow-y-scroll pb-4">
                <div className="flex h-max w-full flex-col gap-2">
                    {sortedTasks.map((sp) => (
                        <TaskProgCard key={sp.taskId} prog={sp} />
                    ))}
                </div>
            </div>
            {sortedTasks.length > 0 && (
                <WeblensButton
                    label={'Clear Tasks'}
                    flavor="outline"
                    onClick={() => {
                        clearTasks()
                    }}
                />
            )}
        </div>
    )
}

function TaskProgTimers({ prog }: { prog: TaskProgress }) {
    const startTime = useRef(0)
    const [elapsed, setElapsed] = useState(0)

    useEffect(() => {
        const update = (now: number) => {
            const epochHighRes = Math.floor(performance.timeOrigin + now)
            if (startTime.current === 0) {
                startTime.current = epochHighRes
            }
            for (const [i] of prog.workingOn.entries()) {
                prog.workingOn[i].elapsedTimeMs =
                    epochHighRes - prog.workingOn[i].startTimeEpochMs
            }
            requestAnimationFrame(update)
            setElapsed(epochHighRes - startTime.current)
        }

        const req = requestAnimationFrame(update)

        return () => cancelAnimationFrame(req)
    }, [])

    return (
        <>
            <span>{msToHumanTime(elapsed)}</span>
            {prog.workingOn.map((workingOn, i) => {
                return (
                    <div
                        className="flex justify-between text-nowrap select-none"
                        key={i}
                    >
                        <p key={workingOn.itemId} className="truncate">
                            {workingOn.itemName}
                        </p>
                        <p>{msToHumanTime(workingOn.elapsedTimeMs)}</p>
                    </div>
                )
            })}
        </>
    )
}

const TaskProgCard = ({ prog }: { prog: TaskProgress }) => {
    const removeTask = useTaskState((state) => state.removeTask)

    const [cancelWarning, setCancelWarning] = useState(false)

    let totalNum: number
    if (typeof prog.tasksTotal === 'string') {
        totalNum = parseInt(prog.tasksTotal)
    } else {
        totalNum = prog.tasksTotal
    }

    return (
        <div className="bg-background-primary hover:bg-background-secondary/50 m-1 flex h-max min-w-0 flex-col items-center rounded-md p-4 transition">
            <div className="flex h-max w-full max-w-full flex-row items-center">
                <div className="flex w-full flex-col">
                    <div className="flex h-max w-full flex-row items-center justify-between">
                        <div className="flex w-[50%] grow flex-row items-center gap-2">
                            <h5 className="max-w-full truncate text-nowrap select-none">
                                {prog.FormatTaskName()}
                            </h5>
                            {Number(prog.tasksFailed) > 0 && (
                                <IconExclamationCircle
                                    className="text-red-500"
                                    size={24}
                                />
                            )}
                        </div>
                        <WeblensButton
                            Right={IconX}
                            size="tiny"
                            flavor="outline"
                            label={cancelWarning ? 'Cancel?' : ''}
                            danger={cancelWarning}
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
                                    CancelTask(
                                        prog.poolId ? prog.poolId : prog.taskId
                                    )

                                    setCancelWarning(false)
                                } else {
                                    setCancelWarning(true)
                                }
                            }}
                        />
                    </div>
                </div>
            </div>
            {prog.stage !== TaskStage.InProgress && (
                <div className="text-color-text-secondary flex h-max w-full flex-row justify-between gap-3 select-none">
                    {prog.stage === TaskStage.Complete && (
                        <span className="w-max">
                            Finished{' '}
                            {prog.timeNs !== 0 ? `in ${prog.getTime()}` : ''}
                        </span>
                    )}
                    {prog.stage === TaskStage.Queued && (
                        <span className="w-max truncate text-sm select-none">
                            Queued...
                        </span>
                    )}
                    {prog.stage === TaskStage.Failure && (
                        <span className="w-max truncate text-sm select-none">
                            {prog.tasksFailed || prog.FormatTaskType()} failed
                        </span>
                    )}
                    {prog.stage === TaskStage.Cancelled && (
                        <span className="w-max truncate text-sm select-none">
                            Cancelled
                        </span>
                    )}
                </div>
            )}
            <div className="relative m-2 h-2 w-full shrink-0">
                <WeblensProgress
                    value={prog.getProgress()}
                    secondaryValue={prog.getErrorProgress()}
                    failure={prog.stage === TaskStage.Failure}
                    loading={prog.stage === TaskStage.Queued}
                    primaryColor={
                        prog.stage === TaskStage.Cancelled ? '#313171' : ''
                    }
                    secondaryColor={'red'}
                />
            </div>
            {prog.stage === TaskStage.InProgress && (
                <div className="m-2 flex h-max w-full flex-col justify-between gap-3">
                    {totalNum > 0 && (
                        <p className="border-color-border-primary border-b pb-1 text-sm select-none">
                            {prog.tasksComplete}/{prog.tasksTotal}
                        </p>
                    )}
                    <TaskProgTimers prog={prog} />
                </div>
            )}
        </div>
    )
}

export function TaskProgressMini() {
    const tasksMap = useTaskState((state) => state.tasks)
    const showingMenu = useTaskState((state) => state.showingMenu)
    const setShowingMenu = useTaskState((state) => state.setShowingMenu)
    const [calloutRef, setCalloutRef] = useState<HTMLDivElement>()
    const [paneRef, setPaneRef] = useState<HTMLDivElement>()

    const taskPoolProgress = useMemo(() => {
        let totalProgress = 0
        for (const task of tasksMap.values()) {
            totalProgress += task.getProgress()
        }
        return totalProgress / tasksMap.size
    }, [tasksMap])

    if (tasksMap.size === 0 && !showingMenu) {
        return null
    }

    return (
        <div
            className="relative flex w-full flex-row"
            onClick={(e) => {
                if (paneRef && paneRef.contains(e.target as Node)) {
                    return
                }
                setShowingMenu(!showingMenu)
            }}
            ref={setCalloutRef}
        >
            {taskPoolProgress < 100 && (
                <div className="group text-color-text-secondary hover:bg-background-secondary hover:text-color-theme-primary mt-2 flex w-full cursor-pointer flex-col items-center gap-2 rounded-sm p-2 transition">
                    <div className="flex w-full items-center">
                        <span className="text-color-text-secondary group-hover:text-color-text-primary transition">
                            Tasks are running
                        </span>
                        <LoaderDots className="group-hover:text-color-text-primary mt-auto mb-1 ml-1" />
                        <IconChevronRight className="text-color-text-secondary group-hover:text-color-text-primary ml-auto transition" />
                    </div>
                    <WeblensProgress
                        value={taskPoolProgress}
                        className="min-h-3 w-full"
                    />
                </div>
            )}
            {taskPoolProgress === 100 && (
                <div className="group text-color-text-secondary hover:bg-background-secondary hover:text-color-theme-primary mt-2 flex w-full cursor-pointer flex-col items-center gap-2 rounded-sm p-2 transition">
                    <div className="flex w-full items-center">
                        <span className="text-color-text-secondary group-hover:text-color-text-primary transition">
                            Tasks complete
                        </span>
                        <IconChevronRight className="text-color-text-secondary group-hover:text-color-text-primary ml-auto transition group-hover:translate-x-1" />
                    </div>
                </div>
            )}
            {tasksMap.size === 0 && (
                <div className="group text-color-text-secondary hover:bg-background-secondary hover:text-color-theme-primary mt-2 flex w-full cursor-pointer flex-col items-center gap-2 rounded-sm p-2 transition">
                    <div className="flex w-full items-center">
                        <span className="text-color-text-secondary group-hover:text-color-text-primary transition">
                            No Tasks
                        </span>
                        <IconChevronRight className="text-color-text-secondary group-hover:text-color-text-primary ml-auto transition group-hover:translate-x-1" />
                    </div>
                </div>
            )}
            {showingMenu && (
                <TasksPane calloutDivRef={calloutRef} setPaneRef={setPaneRef} />
            )}
        </div>
    )
}
