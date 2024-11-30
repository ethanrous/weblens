import { Divider } from '@mantine/core'
import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconChevronLeft,
    IconChevronRight,
    IconExclamationCircle,
    IconFile,
    IconFolder,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { FolderApi } from '@weblens/api/FileBrowserApi'
import { FileActionInfo } from '@weblens/api/swag'
import WeblensLoader from '@weblens/components/Loading'
import { useSessionStore } from '@weblens/components/UserInfo'
import {
    useResize,
    useResizeDrag,
    useWindowSize,
} from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import {
    FbModeT,
    useFileBrowserStore,
} from '@weblens/pages/FileBrowser/FBStateControl'
import { historyDate } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import {
    FileFmt,
    PathFmt,
} from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
import fbStyle from '@weblens/pages/FileBrowser/style/fileBrowserStyle.module.scss'
import historyStyle from '@weblens/pages/FileBrowser/style/historyStyle.module.scss'
import { ErrorHandler } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import { clamp } from '@weblens/util'
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

const SIDEBAR_BREAKPOINT = 650

export default function FileInfoPane() {
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
    const [tab] = useState<string>('history')

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
            setResizeOffset(clamp(v, 300, 800))
        },
        true
    )

    return (
        <div
            className={fbStyle['file-info-pane']}
            data-resizing={dragging}
            data-open={open}
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
                        {/* <div className="flex flex-row h-max w-full gap-1 justify-between p-1 pl-0"> */}
                        {/*     <WeblensButton */}
                        {/*         fillWidth */}
                        {/*         centerContent */}
                        {/*         label="File Info" */}
                        {/*         squareSize={50} */}
                        {/*         toggleOn={tab === 'info'} */}
                        {/*         onClick={() => setTab('info')} */}
                        {/*     /> */}
                        {/**/}
                        {/*     <WeblensButton */}
                        {/*         fillWidth */}
                        {/*         centerContent */}
                        {/*         label="History" */}
                        {/*         squareSize={50} */}
                        {/*         toggleOn={tab === 'history'} */}
                        {/*         onClick={() => setTab('history')} */}
                        {/*     /> */}
                        {/* </div> */}
                        {tab === 'info' && open && <FileInfo />}
                        {tab === 'history' && open && <FileHistory />}
                    </div>
                </div>
            )}
        </div>
    )
}

function FileInfo() {
    const mediaMap = useMediaStore((state) => state.mediaMap)
    const selectedFiles = useFileBrowserStore((state) =>
        Array.from(state.selected.keys())
            .map((fId) => state.filesMap.get(fId))
            .filter((f) => Boolean(f))
    )

    const titleText = useMemo(() => {
        if (selectedFiles.length === 0) {
            return 'No files selected'
        } else if (selectedFiles.length === 1) {
            return selectedFiles[0].GetFilename()
        } else {
            return `${selectedFiles.length} files selected`
        }
    }, [selectedFiles])

    const singleItem = selectedFiles.length === 1
    const itemIsFolder = selectedFiles[0]?.isDir

    const singleFileMedia =
        singleItem && mediaMap.get(selectedFiles[0].GetContentId())

    return (
        <div className={fbStyle['file-info-content'] + ' no-scrollbar'}>
            <div className="flex flex-row h-[58px] w-full items-center justify-between">
                <p className="text-2xl font-semibold text-nowrap pr-8">
                    {titleText}
                </p>
            </div>
            {singleFileMedia && (
                <div className="px-[10%]">
                    <MediaImage
                        media={singleFileMedia}
                        quality={PhotoQuality.LowRes}
                    />
                </div>
            )}
            {selectedFiles.length > 0 && (
                <div className="h-max">
                    <div className="flex flex-row h-full w-full items-center">
                        {singleItem && itemIsFolder && (
                            <IconFolder size={'48px'} />
                        )}
                        {(!singleItem || !itemIsFolder) && (
                            <IconFile size={'48px'} />
                        )}
                        <p className="text-xl">
                            {itemIsFolder ? 'Folder' : 'File'}
                            {singleItem ? '' : 's'} Info
                        </p>
                    </div>
                    <Divider h={2} w={'90%'} m={10} />
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
        if (action.actionType == 'fileMove') {
            if (folderName === fromFolder) {
                fromNode = <FileFmt pathName={action.originPath} />
            } else {
                fromNode = <PathFmt pathName={action.originPath} />
            }
        } else if (
            action.actionType == 'fileCreate' ||
            action.actionType == 'fileRestore'
        ) {
            fromNode = <FileFmt pathName={action.destinationPath} />
        } else if (action.actionType == 'fileDelete') {
            fromNode = <FileFmt pathName={action.originPath} />
        }

        let toNode: ReactElement
        if (action.actionType == 'fileMove') {
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
            {action.actionType === 'fileMove' && (
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
    }
    index: number
    style: CSSProperties
}) {
    return (
        <div style={{ ...style, paddingRight: '10px' }}>
            <HistoryEventRow
                key={data.events[index][0].eventId}
                event={data.events[index]}
                folderPath={data.events[index][0].destinationPath}
                open={data.openEvents[index]}
                setOpen={(o: boolean) =>
                    data.setOpenEvents((p) => {
                        p[index] = o
                        return [...p]
                    })
                }
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
        <div style={{ ...style, paddingRight: '10px' }}>
            <ActionRow
                action={data.actions[index]}
                folderName={data.folderName}
            />
        </div>
    )
}

function ExpandableHistoryRow({
    event,
    folderName,
    open,
    setOpen,
}: {
    event: FileActionInfo[]
    folderName: string
    open: boolean
    setOpen: Dispatch<SetStateAction<boolean>>
}) {
    const [boxRef, setBoxRef] = useState<HTMLDivElement>()
    const boxSize = useResize(boxRef)

    const CaretIcon = open ? IconCaretDown : IconCaretRight

    return (
        <div
            className={historyStyle['file-history-accordion-header']}
            data-open={open}
            style={{ height: open ? 72 + event.length * 32 : 48 }}
        >
            <div
                className="flex flex-row items-center pl-4 h-12 shrink-0 cursor-pointer"
                onClick={() => setOpen(!open)}
            >
                <div className={historyStyle['event-caret']}>
                    <CaretIcon size={20} className="shrink-0" />
                </div>
                <p className="theme-text font-semibold truncate text-xl w-max text-nowrap p-2 select-none">
                    {event.length} File
                    {event.length !== 1 ? 's' : ''}{' '}
                    {event[0].actionType.slice(4)}d ...
                </p>
            </div>
            <div
                ref={setBoxRef}
                className={
                    historyStyle['file-history-detail-accordion'] +
                    ' no-scrollbar'
                }
                data-open={open}
            >
                {open && (
                    <div className={historyStyle['file-history-detail']}>
                        <WindowList
                            height={boxSize.height}
                            width={boxSize.width - 8}
                            itemSize={() => 36}
                            itemCount={event.length}
                            itemData={{ actions: event, folderName }}
                            overscanCount={25}
                        >
                            {ActionRowWrapper}
                        </WindowList>
                    </div>
                )}
            </div>
        </div>
    )
}

const HistoryEventRow = memo(
    ({
        event,
        folderPath,
        open,
        setOpen,
    }: {
        event: FileActionInfo[]
        folderPath: string
        open: boolean
        setOpen: Dispatch<SetStateAction<boolean>>
    }) => {
        const folderName = portableToFileName(folderPath)

        return (
            <div className="flex flex-col w-full h-max justify-center p-2 rounded-lg">
                {event.length == 1 && (
                    <div className="flex flex-row items-center outline-gray-700 outline p-2 rounded w-full justify-between gap-2">
                        <ActionRow action={event[0]} folderName={folderName} />
                        <p className={historyStyle['file-action-text']}>
                            File {event[0].actionType.slice(4)}d
                        </p>
                    </div>
                )}
                {event.length > 1 && (
                    <ExpandableHistoryRow
                        event={event}
                        folderName={folderName}
                        open={open}
                        setOpen={setOpen}
                    />
                )}
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
        }

        return true
    }
)

function RollbackBar({
    events,
    openEvents,
    historyScroll,
}: {
    events: FileActionInfo[][]
    openEvents: boolean[]
    historyScroll: number
}) {
    // const nav = useNavigate()
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const setLocation = useFileBrowserStore((state) => state.setLocationState)

    const [steps, setSteps] = useState(0)

    useEffect(() => {
        if (pastTime && pastTime.getTime() !== 0) {
            let counter = 0
            for (const e of events) {
                if (e[0].timestamp < pastTime.getTime()) {
                    break
                }
                counter++
            }
            setSteps(counter)
        } else {
            setSteps(0)
        }
    }, [pastTime, contentId])

    const [dragging, setDragging] = useState(false)

    useResizeDrag(
        dragging,
        (b) => {
            setDragging(b)
        },
        (v) => {
            v = v - 140 + historyScroll
            if (v < 0) {
                v = 0
            }

            let offset = 0
            let counter = 0
            while (true) {
                if (counter >= openEvents.length) {
                    break
                }
                let nextOffset: number
                if (!openEvents[counter]) {
                    nextOffset = 64
                } else {
                    nextOffset = Math.min(512, 88 + events[counter].length * 32)
                }

                if (offset + nextOffset / 2 > v) {
                    break
                } else if (offset + nextOffset > v) {
                    counter++
                    break
                }

                offset += nextOffset
                counter++
            }
            setSteps(counter)
        },
        false,
        true
    )

    const [dragging2, setDragging2] = useState(false)
    useEffect(() => {
        if (dragging2 && !dragging) {
            let newDate: Date = new Date(0)
            if (steps !== 0) {
                const timestamp = Math.min(
                    ...events[steps - 1].map((a) => a.timestamp)
                )
                newDate = new Date(timestamp)
            }
            if (newDate !== pastTime) {
                setLocation({ contentId: contentId, pastTime: newDate })
            }
        }
        setDragging2(dragging)
    }, [dragging])

    const currentTime = useMemo(() => {
        if (!dragging) {
            return ''
        }
        if (steps === 0) {
            return 'Now'
        }

        return historyDate(events[steps - 1][0].timestamp)
    }, [dragging, steps])

    const offset = useMemo(() => {
        if (steps === 0) {
            return 0
        }
        let offset = 5
        for (let i = 0; i < steps; i++) {
            if (!openEvents[i]) {
                offset += 64
            } else {
                offset += Math.min(516, 88 + events[i].length * 32)
            }
        }
        return offset - historyScroll
    }, [openEvents, steps, historyScroll])

    return (
        <div
            className={historyStyle['rollback-bar-wrapper']}
            style={{ top: offset }}
            onMouseDown={() => setDragging(true)}
        >
            <div
                className={historyStyle['rollback-bar']}
                data-moving={dragging}
            />
            {dragging && (
                <div className="bg-[#333333cc] p-1 rounded h-max w-max relative">
                    <p className="relative select-none z-10 text-white">
                        {currentTime}
                    </p>
                </div>
            )}
        </div>
    )
}

function FileHistory() {
    const user = useSessionStore((state) => state.user)

    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const pastTime = useFileBrowserStore((state) => state.pastTime)

    const [historyScroll, setHistoryScroll] = useState(0)
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

        fileHistory.forEach((a: FileActionInfo) => {
            if (a.lifeId === contentId) {
                epoch = a
                return
            }
            if (
                a.lifeId === user.trashId ||
                a.actionType === 'fileSizeChange'
            ) {
                return
            }

            const i = events.findLastIndex((e) => e[0].eventId === a.eventId)
            if (i != -1) {
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
    }, [openEvents])

    if (mode === FbModeT.share) {
        return (
            <div className="flex justify-center mt-10">
                <p>Cannot get file history of shared file</p>
            </div>
        )
    }

    let createTimeString = '---'
    if (epoch) {
        createTimeString = historyDate(epoch.timestamp)
    }

    return (
        <div
            ref={setBoxRef}
            className="flex flex-col items-center p-2 h-[2px] grow relative pt-3"
            // onScroll={(e) => setHistoryScroll(e.currentTarget.scrollTop)}
        >
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
                <div className="flex flex-col w-full h-1 grow">
                    <RollbackBar
                        events={events}
                        openEvents={openEvents}
                        historyScroll={historyScroll}
                    />
                    <WindowList
                        ref={setWindowRef}
                        height={boxSize.height}
                        width={boxSize.width}
                        itemSize={(i: number) => {
                            if (openEvents[i]) {
                                return Math.min(516, 88 + events[i].length * 32)
                            }
                            return 64
                        }}
                        itemCount={events.length}
                        itemData={{ events, openEvents, setOpenEvents }}
                        overscanCount={25}
                        onScroll={(e) => setHistoryScroll(e.scrollOffset)}
                    >
                        {HistoryRowWrapper}
                    </WindowList>
                </div>
            )}
            <div className="flex flex-col items-center pt-2">
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
