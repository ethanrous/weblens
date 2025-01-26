import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconChevronLeft,
    IconChevronRight,
    IconExclamationCircle,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { FolderApi } from '@weblens/api/FileBrowserApi'
import { FileActionInfo } from '@weblens/api/swag'
import WeblensLoader from '@weblens/components/Loading'
import { useSessionStore } from '@weblens/components/UserInfo'
import historyStyle from '@weblens/components/filebrowser/historyStyle.module.scss'
import {
    useResize,
    useResizeDrag,
    useWindowSize,
} from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import { historyDateTime } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { FbActionT } from '@weblens/pages/FileBrowser/FileBrowserTypes'
import fbStyle from '@weblens/pages/FileBrowser/style/fileBrowserStyle.module.scss'
import { FbModeT, useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { clamp, humanFileSize } from '@weblens/util'
import {
    CSSProperties,
    Dispatch,
    ReactElement,
    SetStateAction,
    memo,
    useEffect,
    useMemo,
    useState,
} from 'react'
import { VariableSizeList, VariableSizeList as WindowList } from 'react-window'

import { FileFmt, PathFmt } from './filename'

const SIDEBAR_BREAKPOINT = 650

function FileHistoryPane() {
    const windowSize = useWindowSize()

    const dragging = useFileBrowserStore(
        (state) => state.draggingState === DraggingStateT.InterfaceDrag
    )
    const [localDragging, setLocalDragging] = useState(false)
    const setDraggingGlobal = useFileBrowserStore((state) => state.setDragging)
    const setDragging = (d: DraggingStateT) => {
        setDraggingGlobal(d)
        setLocalDragging(d === DraggingStateT.InterfaceDrag)
    }

    const [resizeOffset, setResizeOffset] = useState(
        windowSize?.width > SIDEBAR_BREAKPOINT ? 550 : 75
    )
    const [open, setOpen] = useState<boolean>(false)

    useEffect(() => {
        if (!dragging && localDragging) {
            setDragging(DraggingStateT.NoDrag)
        }
    }, [dragging])

    useResizeDrag(
        localDragging,
        (dragging: boolean) => {
            if (dragging) {
                setDragging(DraggingStateT.InterfaceDrag)
            } else {
                setDragging(DraggingStateT.NoDrag)
            }
        },
        (v) => {
            setResizeOffset(clamp(v, 300, windowSize.width / 2))
        },
        true
    )

    return (
        <div
            className={fbStyle['file-info-pane']}
            data-resizing={dragging}
            data-open={open}
            onClick={(e) => {
                e.stopPropagation()
            }}
            onContextMenu={(e) => {
                e.preventDefault()
                e.stopPropagation()
            }}
        >
            <div className={fbStyle['open-arrow-container']}>
                <WeblensButton
                    squareSize={20}
                    Left={open ? IconChevronRight : IconChevronLeft}
                    onClick={(e) => {
                        e.stopPropagation()
                        setOpen(!open)
                    }}
                />
            </div>
            {open && (
                <div
                    className="flex h-full max-w-full"
                    style={{ width: open ? resizeOffset : 20 }}
                >
                    <div
                        draggable={false}
                        className={fbStyle['resize-bar-wrapper']}
                        onMouseDown={(e) => {
                            e.preventDefault()
                            setDragging(DraggingStateT.InterfaceDrag)
                        }}
                        onMouseUp={(e) => {
                            e.stopPropagation()
                            setDragging(DraggingStateT.NoDrag)
                        }}
                        onClick={(e) => {
                            if (dragging) {
                                e.preventDefault()
                            }
                        }}
                    >
                        <div className={fbStyle['resize-bar']} />
                    </div>
                    <div className="flex flex-col w-[75px] grow h-full">
                        {open && <FileHistory />}
                    </div>
                </div>
            )}
        </div>
    )
}

const portableToFolderName = (path: string) => {
    if (path.endsWith('/')) {
        path = path.substring(0, path.length - 1)
    }
    const lastSlash = path.lastIndexOf('/')
    let folderPath = path.slice(
        path.indexOf(':') + 1,
        lastSlash === -1 ? path.length : lastSlash
    )
    folderPath = folderPath.slice(folderPath.lastIndexOf('/') + 1)
    return folderPath
}

const portableToFileName = (path: string) => {
    let filePath = path
    if (path.endsWith('/')) {
        filePath = filePath.substring(0, path.length - 1)
    }
    filePath = filePath.slice(filePath.indexOf(':') + 1)
    filePath = filePath.slice(filePath.lastIndexOf('/') + 1)
    return filePath
}

function ActionRow({
    action,
    folderName,
}: {
    action: FileActionInfo
    folderName: string
}) {
    const { fromNode, toNode } = useMemo(() => {
        const fromFolder = portableToFolderName(action.originPath)
        const toFolder = portableToFolderName(action.destinationPath)

        let fromNode: ReactElement
        if (action.actionType === FbActionT.FileMove.valueOf()) {
            if (folderName === fromFolder) {
                fromNode = <FileFmt pathName={action.originPath} />
            } else {
                fromNode = <PathFmt pathName={action.originPath} />
            }
        } else if (
            action.actionType === FbActionT.FileCreate.valueOf() ||
            action.actionType === FbActionT.FileRestore.valueOf()
        ) {
            fromNode = <FileFmt pathName={action.destinationPath} />
        } else if (action.actionType === FbActionT.FileDelete.valueOf()) {
            fromNode = <FileFmt pathName={action.originPath} />
        }

        let toNode: ReactElement
        if (action.actionType === FbActionT.FileMove.valueOf()) {
            if (folderName !== toFolder) {
                toNode = <PathFmt pathName={action.destinationPath} />
            } else {
                toNode = <FileFmt pathName={action.destinationPath} />
            }
        }

        return { fromNode, toNode }
    }, [action])

    return (
        <div className={historyStyle['history-detail-action-row']}>
            {fromNode}
            {action.actionType === FbActionT.FileMove.valueOf() && (
                <IconArrowRight className="theme-text icon-noshrink" />
            )}
            {toNode}
        </div>
    )
}

function HistoryRowWrapper({
    data,
    index,
    style,
}: {
    data: {
        events: FileActionInfo[][]
        openEvents: boolean[]
        setOpenEvents: Dispatch<SetStateAction<boolean[]>>
        showResize: boolean
    }
    index: number
    style: CSSProperties
}) {
    const previousSize = useMemo(() => {
        let backCounter = 0
        let previousSize = 0
        if (
            data.events[index].length === 1 &&
            data.events[index][0].actionType ===
                FbActionT.FileSizeChange.valueOf()
        ) {
            let i = -1
            while (index + backCounter < data.events.length) {
                backCounter++
                i = data.events[index + backCounter].findIndex((v) => {
                    if (v.lifeId === data.events[index][0].lifeId) {
                        return true
                    }
                })
                if (i !== -1) {
                    previousSize = data.events[index + backCounter][i].size
                    break
                }
            }
        }
        return previousSize
    }, [data])

    return (
        <div
            style={{
                ...style,
                display: 'flex',
                alignItems: 'center',
            }}
        >
            <HistoryEventRow
                key={data.events[index][0].eventId}
                event={data.events[index]}
                folderPath={data.events[index][0].destinationPath}
                previousSize={previousSize}
                open={data.openEvents[index]}
                setOpen={(o: boolean) =>
                    data.setOpenEvents((p) => {
                        p[index] = o
                        return [...p]
                    })
                }
                showResize={data.showResize}
            />
        </div>
    )
}

function ActionRowWrapper({
    data,
    index,
    style,
}: {
    data: {
        actions: FileActionInfo[]
        folderName: string
    }
    index: number
    style: CSSProperties
}) {
    return (
        <div style={{ ...style, paddingRight: '10px', alignItems: 'center' }}>
            <ActionRow
                action={data.actions[index]}
                folderName={data.folderName}
            />
        </div>
    )
}

function ExpandableEventRow({
    event,
    folderName,
    open,
    setOpen,
    showResize,
}: {
    event: FileActionInfo[]
    folderName: string
    open: boolean
    setOpen: Dispatch<SetStateAction<boolean>>
    showResize: boolean
}) {
    const [boxRef, setBoxRef] = useState<HTMLDivElement>()
    const boxSize = useResize(boxRef)
    const pastTime = useFileBrowserStore((state) => state.pastTime)

    const CaretIcon = open ? IconCaretDown : IconCaretRight

    return (
        <div
            className={historyStyle['history-row-content']}
            data-selected={pastTime.getTime() === event[0].timestamp}
            data-expandable={true}
            style={{
                height: open
                    ? getEventHeight(
                          [event],
                          [open],
                          folderName,
                          0,
                          showResize
                      ) - 16
                    : 48,
            }}
        >
            <div className="flex flex-row items-center pl-2 shrink-0 cursor-pointer w-full h-[2rem]">
                <div
                    className={historyStyle['event-caret']}
                    onClick={(e) => {
                        e.stopPropagation()
                        setOpen(!open)
                    }}
                >
                    <CaretIcon size={20} className="shrink-0" />
                </div>
                <p className="theme-text font-semibold truncate text-xl w-max text-nowrap p-2 select-none">
                    {event.length} File
                    {event.length !== 1 ? 's' : ''}{' '}
                    {event[0].actionType.slice(4)}d ...
                </p>
                <p className={historyStyle['file-action-text'] + ' ml-auto'}>
                    {historyDateTime(event[0].timestamp, true)}
                </p>
            </div>
            {open && (
                <div
                    className={historyStyle['file-history-detail-accordion']}
                    ref={setBoxRef}
                >
                    <WindowList
                        height={boxSize.height}
                        width={boxSize.width - 8}
                        itemSize={() => 36}
                        itemCount={event.length}
                        itemData={{ actions: event, folderName }}
                        overscanCount={5}
                    >
                        {ActionRowWrapper}
                    </WindowList>
                </div>
            )}
        </div>
    )
}

const HistoryEventRow = memo(
    ({
        event,
        folderPath,
        previousSize,
        open,
        setOpen,
        showResize,
    }: {
        event: FileActionInfo[]
        folderPath: string
        previousSize: number
        open: boolean
        setOpen: Dispatch<SetStateAction<boolean>>
        showResize: boolean
    }) => {
        const pastTime = useFileBrowserStore((state) => state.pastTime)
        const contentId = useFileBrowserStore((state) => state.contentId)
        const setLocation = useFileBrowserStore(
            (state) => state.setLocationState
        )

        const folderName = portableToFileName(folderPath)

        const date = historyDateTime(event[0].timestamp, true)
        const folderInfo = useFileBrowserStore((state) => state.folderInfo)

        let isSelected = false
        if (pastTime.getTime() === event[0].timestamp) {
            isSelected = true
        }

        if (
            !showResize &&
            event[0].actionType === FbActionT.FileSizeChange.valueOf()
        ) {
            return null
        }

        let content: JSX.Element
        if (
            event.length === 1 &&
            event[0].actionType === FbActionT.FileSizeChange.valueOf()
        ) {
            content = (
                <div className="flex flex-row items-center rounded w-full justify-between gap-2 max-h-[48px] ">
                    <div className={historyStyle['size-change-divider']}>
                        <FileFmt pathName={event[0].destinationPath} />
                        <p>
                            {humanFileSize(previousSize)}
                            {' -> '}
                            {humanFileSize(event[0].size)}
                        </p>
                    </div>
                </div>
            )
        } else if (
            event.length === 1 &&
            event[0].destinationPath === folderInfo?.portablePath
        ) {
            content = (
                <div className={historyStyle['history-row-content']}>
                    <FileFmt pathName={event[0].destinationPath} />
                    <p className={historyStyle['file-action-text']}>
                        Folder {event[0].actionType.slice(4)}d
                    </p>
                </div>
            )
        } else if (event.length === 1) {
            content = (
                <div
                    className={historyStyle['history-row-content']}
                    data-selected={isSelected}
                >
                    <ActionRow action={event[0]} folderName={folderName} />
                    <div className="flex flex-col items-end">
                        <p className={historyStyle['file-action-text']}>
                            File {event[0].actionType.slice(4)}d
                        </p>
                        <p className={historyStyle['file-action-text']}>
                            {date}
                        </p>
                    </div>
                </div>
            )
        } else {
            content = (
                <ExpandableEventRow
                    event={event}
                    folderName={folderName}
                    open={open}
                    setOpen={setOpen}
                    showResize={showResize}
                />
            )
        }

        return (
            <div
                className={historyStyle['history-event-row']}
                data-resize={
                    event[0].actionType === FbActionT.FileSizeChange.valueOf()
                }
                onClick={(e) => {
                    e.stopPropagation()
                    if (
                        event[0].actionType ===
                        FbActionT.FileSizeChange.valueOf()
                    ) {
                        return
                    }

                    if (isSelected) {
                        setLocation({
                            contentId: contentId,
                            pastTime: new Date(0),
                        })
                        return
                    }

                    let newDate = new Date(0)
                    const timestamp = Math.max(...event.map((a) => a.timestamp))
                    newDate = new Date(timestamp)

                    if (newDate !== pastTime) {
                        setLocation({
                            contentId: contentId,
                            pastTime: newDate,
                        })
                    }
                }}
            >
                {content}
            </div>
        )
    },
    (prev, next) => {
        if (prev.event !== next.event) {
            return false
        } else if (prev.open !== next.open) {
            return false
        } else if (prev.setOpen !== next.setOpen) {
            return false
        } else if (prev.folderPath !== next.folderPath) {
            return false
        } else if (prev.previousSize !== next.previousSize) {
            return false
        }

        return true
    }
)

function getEventHeight(
    events: FileActionInfo[][],
    openEvents: boolean[],
    epochPath: string,
    i: number,
    showResize: boolean
) {
    if (openEvents[i]) {
        return Math.min(516, 72 + events[i].length * 36)
    }

    if (events[i][0].actionType === FbActionT.FileSizeChange.valueOf()) {
        if (!showResize) {
            return 0
        } else if (events[i].length === 1) {
            return 48
        }
    }

    if (events[i].length === 1 && events[i][0].destinationPath === epochPath) {
        return 90
    }
    return 64
}

function FileHistoryFooter({
    epoch,
    showResize,
    setShowResize,
}: {
    epoch: FileActionInfo
    showResize: boolean
    setShowResize: Dispatch<SetStateAction<boolean>>
}) {
    let createTimeString = '---'
    if (epoch) {
        createTimeString = historyDateTime(epoch.timestamp)
    }

    return (
        <div className="flex flex-col w-full justify-around p-2">
            <WeblensButton
                allowRepeat
                label={'Size Changes'}
                toggleOn={showResize}
                onClick={() => setShowResize((r) => !r)}
            />
            <div className="flex flex-col items-center pt-2 border-t border-t-[--wl-outline-subtle] mt-4">
                <div className="flex flex-row gap-2 items-center">
                    <FileFmt pathName={epoch?.destinationPath} />
                    <p className="h-max text-xl select-none">History</p>
                </div>
                <p className="text-nowrap select-none">
                    Created {createTimeString}
                </p>
            </div>
        </div>
    )
}

function FileHistory() {
    const user = useSessionStore((state) => state.user)

    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const setLocation = useFileBrowserStore((state) => state.setLocationState)

    const [showResize, setShowResize] = useState<boolean>(false)

    const [windowRef, setWindowRef] = useState<VariableSizeList>()
    const [boxRef, setBoxRef] = useState<HTMLDivElement>()

    const boxSize = useResize(boxRef)

    const {
        data: fileHistory,
        refetch,
        error,
    } = useQuery<FileActionInfo[]>({
        queryKey: ['fileHistory', contentId],
        queryFn: () => {
            if (mode === FbModeT.share) {
                return []
            }

            let timestamp = Date.now()
            if (pastTime) {
                timestamp = pastTime.getTime()
            }

            return FolderApi.getFolderHistory(contentId, timestamp).then(
                (res) => {
                    return res.data
                }
            )
        },
    })

    useEffect(() => {
        refetch().catch(ErrorHandler)
    }, [filesMap.size])

    const { events, epoch } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return { events: [], epoch: null }
        }

        const events: FileActionInfo[][] = []
        let epoch: FileActionInfo
        // let lastSizeChangeIndex = -1

        fileHistory.forEach((a: FileActionInfo) => {
            if (a.lifeId === contentId) {
                epoch = a
                return
            }
            if (a.lifeId === user.trashId) {
                return
            }

            const i = events.findLastIndex(
                (e) =>
                    e[0].eventId === a.eventId || e[0].timestamp === a.timestamp
            )

            // const isSizeChange =
            //     a.actionType === String(FbActionT.FileSizeChange)
            // if (isSizeChange && lastSizeChangeIndex !== -1) {
            //     events[lastSizeChangeIndex][0].size = a.size
            //     return
            // } else if (!isSizeChange) {
            //     lastSizeChangeIndex = -1
            // } else {
            //     lastSizeChangeIndex = i
            // }

            if (i !== -1) {
                events[i].push(a)
            } else {
                events.push([a])
            }
        })

        return { events, epoch }
    }, [fileHistory])

    const [openEvents, setOpenEvents] = useState<boolean[]>([])
    useEffect(() => {
        const openInit = events.map(() => false)
        setOpenEvents(openInit)
    }, [events])

    useEffect(() => {
        windowRef?.resetAfterIndex(0)
    }, [openEvents, showResize])

    if (mode === FbModeT.share) {
        return (
            <div className="flex justify-center mt-10">
                <p>Cannot get file history of shared file</p>
            </div>
        )
    }

    return (
        <div className="flex flex-col items-center p-2 h-[2px] grow relative w-full justify-center">
            <div
                className={historyStyle['history-row-content']}
                data-now={true}
                data-selected={pastTime.getTime() === 0}
                onClick={() => {
                    setLocation({
                        contentId: contentId,
                        pastTime: new Date(0),
                    })
                }}
            >
                <p className="relative select-none z-10 text-[--wl-text-color]">Now</p>
            </div>
            {!epoch && !error && (
                <div className="flex items-center justify-center w-full h-1 grow">
                    <WeblensLoader />
                </div>
            )}
            {error && (
                <div className="flex items-center w-full h-1 grow">
                    <div className="inline-flex flex-row w-max p-2 m-auto gap-1">
                        <IconExclamationCircle
                            size={24}
                            className="ml-2 text-red-500"
                        />
                        <p>Failed to get file history</p>
                    </div>
                </div>
            )}
            {epoch && (
                <div
                    ref={setBoxRef}
                    className="relative flex flex-col w-full h-1 grow pt-1"
                >
                    <WindowList
                        ref={setWindowRef}
                        height={boxSize.height}
                        width={boxSize.width}
                        style={{ position: 'relative' }}
                        itemSize={(i: number) =>
                            getEventHeight(
                                events,
                                openEvents,
                                epoch.destinationPath,
                                i,
                                showResize
                            )
                        }
                        itemCount={events.length}
                        itemData={{
                            events,
                            openEvents,
                            setOpenEvents,
                            showResize,
                        }}
                        overscanCount={5}
                    >
                        {HistoryRowWrapper}
                    </WindowList>
                </div>
            )}
            <FileHistoryFooter
                epoch={epoch}
                showResize={showResize}
                setShowResize={setShowResize}
            />
        </div>
    )
}

export default FileHistoryPane
