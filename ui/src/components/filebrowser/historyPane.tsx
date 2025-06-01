import {
    IconCaretDown,
    IconCaretRight,
    IconChevronLeft,
    IconChevronRight,
    IconCircleMinus,
    IconCirclePlus,
    IconExternalLink,
    IconFile,
    IconFolderOpen,
    IconFolderSymlink,
    IconRestore,
    IconTrash,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { FolderApi } from '@weblens/api/FileBrowserApi'
import { FileActionInfo } from '@weblens/api/swag'
import WeblensLoader from '@weblens/components/Loading.tsx'
import { useSessionStore } from '@weblens/components/UserInfo'
import historyStyle from '@weblens/components/filebrowser/historyStyle.module.scss'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import { useResize, useResizeDrag, useWindowSize } from '@weblens/lib/hooks'
import {
    filenameFromPath,
    historyDateTime,
} from '@weblens/pages/FileBrowser/FileBrowserLogic'
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
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react'
import { VariableSizeList, VariableSizeList as WindowList } from 'react-window'

import { RecoverableError } from '../Error'
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
            className={fbStyle.fileInfoPane}
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
            <div className={fbStyle.openArrowContainer}>
                <WeblensButton
                    size="tiny"
                    className="h-6 w-6"
                    flavor="outline"
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
                        className={fbStyle.resizeBarWrapper}
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
                        <div className={fbStyle.resizeBar} />
                    </div>
                    <div className="flex h-full w-[75px] grow flex-col">
                        {open && <FileHistory />}
                    </div>
                </div>
            )}
        </div>
    )
}

const relevantOrigin = (action: FileActionInfo) => {
    return action.originPath ?? action.filepath ?? ''
}

const relevantDestination = (action: FileActionInfo) => {
    return action.destinationPath ?? action.filepath ?? ''
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
    isSelected,
    inOpenEvent,
}: {
    action: FileActionInfo
    folderName: string
    isSelected: boolean
    inOpenEvent?: boolean
}) {
    const { fromNode, toNode, moveOut } = useMemo(() => {
        const fromFolder = portableToFolderName(relevantOrigin(action))
        const toFolder = portableToFolderName(relevantDestination(action))
        const originName = portableToFileName(relevantOrigin(action))

        let fromNode: ReactElement
        const pathClass =
            'font-semibold border rounded p-1 bg-card-background-primary hover:bg-card-background-hover'
        if (action.actionType === FbActionT.FileMove.valueOf()) {
            if (folderName === fromFolder) {
                fromNode = (
                    <FileFmt
                        pathName={relevantOrigin(action)}
                        className={pathClass}
                    />
                )
            } else {
                fromNode = (
                    <PathFmt
                        pathName={relevantOrigin(action)}
                        className={pathClass}
                    />
                )
            }
        } else if (
            action.actionType === FbActionT.FileCreate.valueOf() ||
            action.actionType === FbActionT.FileRestore.valueOf()
        ) {
            fromNode = (
                <FileFmt
                    pathName={relevantDestination(action)}
                    className={pathClass}
                />
            )
        } else if (action.actionType === FbActionT.FileDelete.valueOf()) {
            fromNode = (
                <FileFmt
                    pathName={relevantOrigin(action)}
                    className={pathClass}
                />
            )
        }

        let toNode: ReactElement
        let moveOut = false
        if (action.actionType === FbActionT.FileMove.valueOf()) {
            if (folderName !== toFolder) {
                moveOut = true
                toNode = (
                    <PathFmt
                        pathName={relevantDestination(action)}
                        className={pathClass}
                        excludeBasenameMatching={originName}
                    />
                )
            } else {
                toNode = (
                    <FileFmt
                        pathName={relevantDestination(action)}
                        className={pathClass}
                    />
                )
            }
        }

        return { fromNode, toNode, moveOut }
    }, [action])

    let ActionIcon: typeof IconFile = IconFile
    let actionColor: string
    if (action.actionType === FbActionT.FileMove.valueOf()) {
        if (moveOut) {
            if (relevantDestination(action).includes('.user_trash')) {
                ActionIcon = IconTrash
                actionColor = 'var(--color-danger)'
            } else {
                ActionIcon = IconExternalLink
            }
        } else {
            ActionIcon = IconFolderSymlink
        }
    } else if (action.actionType === FbActionT.FileCreate.valueOf()) {
        ActionIcon = IconCirclePlus
        actionColor = 'var(--color-valid)'
    } else if (action.actionType === FbActionT.FileDelete.valueOf()) {
        ActionIcon = IconCircleMinus
        actionColor = 'var(--color-danger)'
    } else if (action.actionType === FbActionT.FileRestore.valueOf()) {
        ActionIcon = IconRestore
    }

    return (
        <>
            <ActionIcon
                size={16}
                color={actionColor}
                className="mx-1 shrink-0"
            />
            <div className="mr-1 flex flex-col items-end">
                <span>{action.actionType.slice(4)}d</span>
            </div>

            {fromNode}

            {action.actionType === FbActionT.FileMove.valueOf() && (
                <span>to</span>
            )}

            {toNode}

            {!inOpenEvent && (
                <span
                    className="text-text-tertiary data-selected:text-text-primary ml-auto text-nowrap"
                    data-selected={isSelected ? true : undefined}
                >
                    {historyDateTime(action.timestamp, true)}
                </span>
            )}
        </>
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
                    if (v.fileId === data.events[index][0].fileId) {
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

    if (!data.events[index]) {
        return null
    }

    const thisEvent = data.events[index]

    return (
        <div
            style={{
                ...style,
                display: 'flex',
                alignItems: 'center',
                paddingTop: '2px',
                paddingBottom: '2px',
                paddingRight: '12px',
            }}
        >
            <HistoryEventRow
                key={thisEvent[0].eventId}
                event={thisEvent}
                folderPath={
                    thisEvent[0].filepath ?? thisEvent[0].destinationPath
                }
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

function EventDropdownRow({
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
        <div
            style={{
                ...style,
                alignItems: 'center',
                display: 'flex',
                gap: 4,
            }}
        >
            <ActionRow
                action={data.actions[index]}
                folderName={data.folderName}
                isSelected={false}
                inOpenEvent={true}
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
    const boxRef = useRef<HTMLDivElement>(null)
    const boxSize = useResize(boxRef)
    const pastTime = useFileBrowserStore((state) => state.pastTime)

    const CaretIcon = open ? IconCaretDown : IconCaretRight

    return (
        <div
            className="flex w-full flex-col justify-start gap-2 p-1"
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
                    : 36,
            }}
        >
            <div className="flex h-[28px] w-full shrink-0 cursor-pointer items-center">
                <div
                    className="text-text-secondary hover:text-text-primary hover:bg-background-secondary hover:border-text-primary rounded-md border transition"
                    onClick={(e) => {
                        e.stopPropagation()
                        setOpen(!open)
                    }}
                >
                    <CaretIcon size={20} className="shrink-0" />
                </div>
                <span className="theme-text w-max truncate p-2 font-semibold text-nowrap select-none">
                    {event.length} File
                    {event.length !== 1 ? 's' : ''}{' '}
                    {event[0].actionType.slice(4)}d ...
                </span>
                <span
                    className="text-text-tertiary data-selected:text-text-primary ml-auto text-nowrap"
                    // data-selected={isSelected ? true : undefined}
                >
                    {historyDateTime(event[0].timestamp, true)}
                </span>
            </div>
            <div
                ref={boxRef}
                className="relative flex h-0 w-full flex-col rounded-md data-open:h-full data-open:max-h-full"
                data-open={open ? true : undefined}
            >
                <WindowList
                    height={boxSize.height}
                    width={boxSize.width}
                    itemSize={() => 24}
                    itemCount={event.length}
                    itemData={{ actions: event, folderName }}
                    overscanCount={5}
                >
                    {EventDropdownRow}
                </WindowList>
            </div>
        </div>
    )
}

function HistoryEventRow({
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
}) {
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const contentId = useFileBrowserStore((state) => state.activeFileId)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const setLocation = useFileBrowserStore((state) => state.setLocationState)

    const folderName = portableToFileName(folderPath)

    // const date = historyDateTime(event[0].timestamp, true)

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

    let content: ReactElement
    if (
        event.length === 1 &&
        event[0].actionType === FbActionT.FileSizeChange.valueOf()
    ) {
        content = (
            <div className="flex max-h-[48px] w-full flex-row items-center justify-between gap-2 rounded-sm">
                <div className={historyStyle.sizeChangeDivider}>
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
            <>
                <FileFmt pathName={event[0].destinationPath} />
                <span className={historyStyle.fileActionText}>
                    Folder {event[0].actionType.slice(4)}d
                </span>
            </>
        )
    } else if (event.length === 1) {
        content = (
            <ActionRow
                action={event[0]}
                folderName={folderInfo?.GetFilename()}
                isSelected={isSelected}
            />
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
            className="data-selected:border-text-secondary data-selected:bg-background-secondary hover:bg-background-secondary my-1 flex h-full w-full cursor-pointer items-center justify-center gap-1 rounded border p-1"
            data-resize={
                event[0].actionType === FbActionT.FileSizeChange.valueOf()
            }
            data-selected={isSelected ? true : undefined}
            onClick={(e) => {
                e.stopPropagation()
                if (
                    event[0].actionType === FbActionT.FileSizeChange.valueOf()
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
}

function getEventHeight(
    events: FileActionInfo[][],
    openEvents: boolean[],
    epochPath: string,
    i: number,
    showResize: boolean
) {
    if (openEvents[i]) {
        return Math.min(516, 72 + events[i].length * 24)
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
    return 44
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

    const { nameText, StartIcon } = filenameFromPath(epoch?.destinationPath)

    return (
        <div className="flex w-full flex-col justify-around p-2">
            <WeblensButton
                allowRepeat
                label={'Size Changes'}
                toggleOn={showResize}
                onClick={() => setShowResize((r) => !r)}
                className="hidden"
            />
            <div className="border-t-color-border-primary mt-4 flex flex-col items-center border-t pt-2">
                <div className="flex flex-row items-center">
                    {StartIcon && <StartIcon />}
                    <h4>{nameText}</h4>
                    <p className="ml-2 h-max text-xl select-none">History</p>
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
    const activeFileId = useFileBrowserStore((state) => state.activeFileId)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const setLocation = useFileBrowserStore((state) => state.setLocationState)

    const [showResize, setShowResize] = useState<boolean>(false)

    const windowRef = useRef<VariableSizeList>(null)
    const boxRef = useRef<HTMLDivElement>(null)

    const boxSize = useResize(boxRef)

    const {
        data: fileHistory,
        refetch,
        isPending,
        error,
    } = useQuery<FileActionInfo[]>({
        queryKey: ['fileHistory', activeFileId],
        retry: false,
        queryFn: () => {
            if (mode === FbModeT.share) {
                return []
            }

            let timestamp = Date.now()
            if (pastTime) {
                timestamp = pastTime.getTime()
            }

            return FolderApi.getFolderHistory(activeFileId, timestamp).then(
                (res) => {
                    return res.data
                }
            )
        },
    })

    if (error) {
        console.error('Failed to fetch file history:', error)
    }

    useEffect(() => {
        refetch().catch(ErrorHandler)
    }, [filesMap.size, refetch])

    const { events, epoch } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return { events: [], epoch: null }
        }

        console.log('File history:', fileHistory)

        const events: FileActionInfo[][] = []
        let epoch: FileActionInfo
        // let lastSizeChangeIndex = -1

        fileHistory.forEach((a: FileActionInfo) => {
            if (a.fileId === activeFileId) {
                epoch = a
                return
            }
            if (a.fileId === user.trashId) {
                return
            }

            const i = events.findLastIndex(
                (e) =>
                    e[0].eventId === a.eventId || e[0].timestamp === a.timestamp
            )

            if (i !== -1) {
                events[i].push(a)
            } else {
                events.push([a])
            }
        })

        return { events, epoch }
    }, [fileHistory, user.trashId, activeFileId])

    const [openEvents, setOpenEvents] = useState<boolean[]>([])
    useEffect(() => {
        const openInit = events.map(() => false)
        setOpenEvents(openInit)
    }, [events])

    useEffect(() => {
        windowRef.current?.resetAfterIndex(0)
    }, [openEvents, showResize])

    if (mode === FbModeT.share) {
        return (
            <div className="mt-10 flex justify-center">
                <p>Cannot get file history of shared file</p>
            </div>
        )
    }

    return (
        <div className="relative flex h-[2px] w-full grow flex-col items-center justify-center p-2">
            <div
                className={historyStyle.historyRowContent}
                data-now={true}
                data-selected={pastTime.getTime() === 0}
                onClick={() => {
                    setLocation({
                        contentId: activeFileId,
                        pastTime: new Date(0),
                    })
                }}
            >
                <p className="relative z-10 text-(--color-text) select-none">
                    Now
                </p>
            </div>
            {isPending && (
                <div className="flex h-1 w-full grow items-center justify-center">
                    <WeblensLoader />
                </div>
            )}
            {error && (
                <div className="flex h-1 w-full grow flex-col items-center">
                    <div className="m-auto inline-flex w-max flex-row items-center justify-center gap-1 p-2">
                        <RecoverableError
                            message="Failed to get file history"
                            error={error}
                        />
                    </div>
                </div>
            )}
            {!isPending && !error && (!epoch || events.length <= 0) && (
                <div className="flex h-1 w-full grow flex-col items-center justify-center">
                    <span className="z-10">
                        Nothing has happened here... yet
                    </span>
                    <IconFolderOpen
                        className="text-nearly-invisible"
                        size={150}
                    />
                </div>
            )}
            {epoch && !error && (
                <div
                    ref={boxRef}
                    className="relative flex h-1 w-full grow flex-col pt-1"
                >
                    <WindowList
                        ref={windowRef}
                        height={boxSize.height}
                        width={boxSize.width}
                        style={{ position: 'relative' }}
                        itemSize={(i: number) =>
                            getEventHeight(
                                events,
                                openEvents,
                                epoch.filepath,
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
