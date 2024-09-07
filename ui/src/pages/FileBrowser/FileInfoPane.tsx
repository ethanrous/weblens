import { Divider } from '@mantine/core'
import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconChevronLeft,
    IconChevronRight,
    IconFile,
    IconFolder,
} from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { getFileHistory } from '@weblens/api/FileBrowserApi'
import { useResizeDrag, useWindowSize } from '@weblens/components/hooks'
import WeblensLoader from '@weblens/components/Loading'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import '@weblens/pages/FileBrowser/style/history.scss'
import {
    FbModeT,
    useFileBrowserStore,
} from '@weblens/pages/FileBrowser/FBStateControl'
import { historyDate } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import {
    FriendlyFile,
    FriendlyPath,
} from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
import { clamp } from '@weblens/util'

import { memo, useEffect, useMemo, useState } from 'react'

const SIDEBAR_BREAKPOINT = 650

export const FileInfoPane = () => {
    const windowSize = useWindowSize()
    const [resizing, setResizing] = useState(false)
    const [resizeOffset, setResizeOffset] = useState(
        windowSize?.width > SIDEBAR_BREAKPOINT ? 450 : 75
    )
    const [open, setOpen] = useState<boolean>(false)
    const [tab, setTab] = useState('info')
    useResizeDrag(
        resizing,
        setResizing,
        (v) => {
            setResizeOffset(clamp(v, 300, 800))
        },
        true
    )

    return (
        <div
            className="file-info-pane"
            data-resizing={resizing}
            data-open={open}
            style={{ width: open ? resizeOffset : 20 }}
        >
            <div className="open-arrow-container">
                <WeblensButton
                    squareSize={20}
                    Left={open ? IconChevronRight : IconChevronLeft}
                    onClick={() => {
                        setOpen(!open)
                    }}
                />
            </div>
            <div
                draggable={false}
                className="resize-bar-wrapper"
                onMouseDown={(e) => {
                    e.preventDefault()
                    setResizing(true)
                }}
            >
                <div className="resize-bar" />
            </div>
            <div className="flex flex-col w-[75px] grow h-full">
                <div className="flex flex-row h-max w-full gap-1 justify-between p-1 pl-0">
                    <WeblensButton
                        fillWidth
                        centerContent
                        label="File Info"
                        squareSize={50}
                        toggleOn={tab === 'info'}
                        onClick={() => setTab('info')}
                    />

                    <WeblensButton
                        fillWidth
                        centerContent
                        label="History"
                        squareSize={50}
                        toggleOn={tab === 'history'}
                        onClick={() => setTab('history')}
                    />
                </div>
                {tab === 'info' && open && <FileInfo />}
                {tab === 'history' && open && <FileHistory />}
            </div>
        </div>
    )
}

function FileInfo() {
    const selectedFiles = useFileBrowserStore((state) =>
        Array.from(state.selected.keys())
            .map((fId) => state.filesMap.get(fId))
            .filter((f) => Boolean(f))
    )

    const titleText = useMemo(() => {
        if (selectedFiles.length === 0) {
            return 'No files selected'
        } else if (selectedFiles.length === 1) {
            return selectedFiles[0].filename
        } else {
            return `${selectedFiles.length} files selected`
        }
    }, [selectedFiles])

    const singleItem = selectedFiles.length === 1
    const itemIsFolder = selectedFiles[0]?.isDir

    return (
        <div className="file-info-content">
            <div className="flex flex-row h-[58px] w-full items-center justify-between">
                <p className="text-2xl font-semibold text-nowrap pr-8">
                    {titleText}
                </p>
            </div>
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

function ActionRow({
    action,
    folderName,
}: {
    action: fileAction
    folderName: string
}) {
    const fromFolder = portableToFolderName(action.originPath)
    const toFolder = portableToFolderName(action.destinationPath)

    return (
        <div className="history-detail-action-row">
            {action.actionType === 'fileMove' && folderName === fromFolder && (
                <FriendlyFile pathName={action.originPath} />
            )}
            {action.actionType === 'fileMove' && folderName !== fromFolder && (
                <FriendlyPath pathName={action.originPath} />
            )}
            {action.actionType === 'fileCreate' && (
                <FriendlyFile pathName={action.destinationPath} />
            )}
            {action.actionType === 'fileDelete' && (
                <FriendlyFile pathName={action.originPath} />
            )}
            {action.actionType === 'fileRestore' && (
                <FriendlyFile pathName={action.destinationPath} />
            )}
            {action.actionType === 'fileMove' && (
                <IconArrowRight className="icon-noshrink" />
            )}
            {action.actionType === 'fileMove' && folderName !== toFolder && (
                <FriendlyPath pathName={action.destinationPath} />
            )}
            {action.actionType === 'fileMove' && folderName === toFolder && (
                <FriendlyFile pathName={action.destinationPath} />
            )}
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
        event: fileAction[]
        folderPath: string
        open: boolean
        setOpen: (o: boolean) => void
    }) => {
        const folderName = portableToFolderName(folderPath)

        let caretIcon
        if (open) {
            caretIcon = <IconCaretDown size={20} style={{ flexShrink: 0 }} />
        } else {
            caretIcon = <IconCaretRight size={20} style={{ flexShrink: 0 }} />
        }

        return (
            <div className="flex flex-col w-full h-max justify-center p-2 rounded-lg">
                {event.length == 1 && (
                    <div className="flex flex-row items-center outline-gray-700 outline p-2 rounded w-full justify-between">
                        <ActionRow action={event[0]} folderName={folderName} />
                        <p className="text-nowrap select-none">
                            File {event[0].actionType.slice(4)}d
                        </p>
                    </div>
                )}
                {event.length > 1 && (
                    <div
                        className="file-history-accordion-header"
                        data-open={open}
                        style={{ height: open ? 72 + event.length * 32 : 48 }}
                    >
                        <div
                            className="flex flex-row items-center pl-4 h-12 shrink-0 cursor-pointer"
                            onClick={() => setOpen(!open)}
                        >
                            <div className="event-caret">{caretIcon}</div>
                            <p className="text-white font-semibold truncate text-xl w-max text-nowrap p-2 select-none">
                                {event.length} File
                                {event.length !== 1 ? 's' : ''}{' '}
                                {event[0].actionType.slice(4)}d ...
                            </p>
                        </div>
                        <div
                            className="file-history-detail-accordion no-scrollbar"
                            data-open={open}
                        >
                            {open && (
                                <div className="file-history-detail">
                                    {event.map((a, i) => {
                                        return (
                                            <ActionRow
                                                key={a.eventId + i}
                                                action={a}
                                                folderName={folderName}
                                            />
                                        )
                                    })}
                                </div>
                            )}
                        </div>
                    </div>
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
}: {
    events: fileAction[][]
    openEvents: boolean[]
}) {
    const setPastTime = useFileBrowserStore((state) => state.setPastTime)

    const [steps, setSteps] = useState(0)
    const [dragging, setDragging] = useState(false)
    useResizeDrag(
        dragging,
        setDragging,
        (v) => {
            v = v - 205
            if (v < 0) {
                v = 0
            }

            let offset = 0
            let counter = 0
            while (true) {
                if (counter >= openEvents.length) {
                    break
                }
                let nextOffset
                if (!openEvents[counter]) {
                    nextOffset = 64
                } else {
                    nextOffset = Math.min(500, 88 + events[counter].length * 32)
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

    useEffect(() => {
        if (!dragging) {
            if (steps === 0) {
                setPastTime(null)
                return
            }
            const event = events[steps - 1]
            if (event) {
                setPastTime(
                    new Date(
                        Math.min(...events[steps - 1].map((a) => a.timestamp))
                    )
                )
            }
        }
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
                offset += Math.min(500, 88 + events[i].length * 32)
            }
        }
        return offset
    }, [openEvents, steps])

    return (
        <div
            className="rollback-bar-wrapper"
            style={{ top: offset }}
            onMouseDown={() => setDragging(true)}
            onMouseUp={() => setDragging(false)}
        >
            <div className="rollback-bar" data-moving={dragging} />
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

type fileAction = {
    actionType: string
    destinationId: string
    destinationPath: string
    eventId: string
    lifeId: string
    originId: string
    originPath: string
    timestamp: number
}

function FileHistory() {
    const user = useSessionStore((state) => state.user)
    const authHeader = useSessionStore((state) => state.auth)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const mode = useFileBrowserStore((state) => state.fbMode)

    const { data: fileHistory, refetch } = useQuery<fileAction[]>({
        queryKey: ['fileHistory', contentId],
        queryFn: () => {
            if (mode === FbModeT.share) {
                return null
            }
            return getFileHistory(contentId, authHeader)
        },
    })

    useEffect(() => {
        refetch()
    }, [filesMap.size])

    const { events, epoch } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return { events: [], epoch: null }
        }

        const events: fileAction[][] = []
        let epoch: fileAction

        fileHistory.forEach((a) => {
            if (a.lifeId === contentId) {
                epoch = a
                return
            }
            if (a.lifeId === user.trashId) {
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

    const [openEvents, setOpenEvents] = useState([])
    useEffect(() => {
        const openInit = events.map(() => false)
        setOpenEvents(openInit)
    }, [events])

    if (mode === FbModeT.share) {
        return (
            <div className="flex justify-center mt-10">
                <p>Cannot get file history of shared file</p>
            </div>
        )
    }
    if (!epoch) {
        return <WeblensLoader />
    }

    const createTimeString = historyDate(epoch.timestamp)

    return (
        <div className="flex flex-col items-center p-2 overflow-scroll h-[200px] grow relative pt-3">
            <RollbackBar events={events} openEvents={openEvents} />
            {events.map((e, i) => {
                return (
                    <HistoryEventRow
                        key={e[0].eventId}
                        event={e}
                        folderPath={epoch.destinationPath}
                        open={openEvents[i]}
                        setOpen={(o: boolean) =>
                            setOpenEvents((p) => {
                                p[i] = o
                                return [...p]
                            })
                        }
                    />
                )
            })}
            <div className="flex flex-col items-center p-2 pt-10">
                <FriendlyFile pathName={epoch.destinationPath} />
                <p className="text-nowrap select-none">
                    Created {createTimeString}
                </p>
            </div>
        </div>
    )
}