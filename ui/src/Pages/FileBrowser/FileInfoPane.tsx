import { Divider } from '@mantine/core'

import { useResizeDrag, useWindowSize } from '../../components/hooks'
import { memo, useEffect, useMemo, useState } from 'react'
import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconChevronLeft,
    IconChevronRight,
    IconFile,
    IconFolder,
} from '@tabler/icons-react'
import { FriendlyFile, FriendlyPath } from './FileBrowserMiscComponents'
import WeblensButton from '../../components/WeblensButton'
import './style/history.scss'
import { useQuery } from '@tanstack/react-query'
import WeblensLoader from '../../components/Loading'
import { clamp } from '../../util'
import { getFileHistory } from '../../api/FileBrowserApi'
import { historyDate } from './FileBrowserLogic'
import { FbModeT, useFileBrowserStore } from './FBStateControl'
import { useSessionStore } from '../../components/UserInfo'

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
        (v) => setResizeOffset(clamp(v, 300, 800)),
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
    const folderPath = path.slice(path.indexOf(':') + 1, path.lastIndexOf('/'))
    const folderName = folderPath.slice(folderPath.lastIndexOf('/') + 1)
    return folderName
}
const fileBase = (path: string) => {
    if (path[path.length - 1] === '/') {
        path = path.slice(0, path.length - 1)
    }
    return path.slice(path.lastIndexOf('/') + 1)
}

const HistoryEventRow = memo(
    ({
        event,
        folderPath,
        viewing,
    }: {
        event: fileAction[]
        viewing: boolean
        folderPath: string
    }) => {
        const [open, setOpen] = useState(false)
        const setPastTime = useFileBrowserStore((state) => state.setPastTime)

        const timeStr = historyDate(event[0].timestamp)
        const folderName = portableToFolderName(folderPath)
        return (
            <div className="flex flex-col w-full h-max justify-center p-2 rounded-lg">
                <div className="file-history-summary">
                    <div
                        className="file-history-accordion-header"
                        onClick={() => setOpen(!open)}
                    >
                        {open ? (
                            <IconCaretDown
                                size={20}
                                style={{ flexShrink: 0 }}
                            />
                        ) : (
                            <IconCaretRight
                                size={20}
                                style={{ flexShrink: 0 }}
                            />
                        )}
                        <p className="text-white font-semibold truncate text-xl w-max text-nowrap p-2">
                            {event.length} File
                            {event.length !== 1 ? 's' : ''}{' '}
                            {event[0].actionType.slice(4)}d
                        </p>
                    </div>
                    <div className="grow" />
                    <WeblensButton
                        toggleOn={viewing}
                        label={timeStr}
                        centerContent
                        allowRepeat
                        onClick={() =>
                            setPastTime(
                                viewing ? null : new Date(event[0].timestamp)
                            )
                        }
                    />
                </div>
                <div
                    className="file-history-detail-accordion"
                    data-open={open}
                    style={{ height: open ? event.length * 36 + 24 : 0 }}
                >
                    <div className="file-history-detail no-scrollbar">
                        {event.map((a, i) => {
                            const fromFolder = portableToFolderName(
                                a.originPath
                            )
                            const toFolder = portableToFolderName(
                                a.destinationPath
                            )

                            const fromFile = fileBase(a.originPath)
                            const toFile = fileBase(a.destinationPath)

                            return (
                                <div
                                    key={`${fromFile}-${toFile}-${i}`}
                                    className="history-detail-action-row"
                                >
                                    {a.actionType === 'fileMove' &&
                                        folderName === fromFolder && (
                                            <FriendlyFile
                                                pathName={a.originPath}
                                            />
                                        )}
                                    {a.actionType === 'fileMove' &&
                                        folderName !== fromFolder && (
                                            <FriendlyPath
                                                pathName={a.originPath}
                                            />
                                        )}
                                    {a.actionType === 'fileCreate' && (
                                        <FriendlyFile
                                            pathName={a.destinationPath}
                                        />
                                    )}
                                    {a.actionType === 'fileDelete' && (
                                        <FriendlyFile pathName={a.originPath} />
                                    )}
                                    {a.actionType === 'fileRestore' && (
                                        <FriendlyFile
                                            pathName={a.destinationPath}
                                        />
                                    )}
                                    {a.actionType === 'fileMove' && (
                                        <IconArrowRight className="icon-noshrink" />
                                    )}

                                    {a.actionType === 'fileMove' &&
                                        folderName === toFolder && (
                                            <FriendlyPath
                                                pathName={a.destinationPath}
                                            />
                                        )}

                                    {a.actionType === 'fileMove' &&
                                        folderName !== toFolder && (
                                            <FriendlyFile
                                                pathName={a.destinationPath}
                                            />
                                        )}

                                    {/*<p className="text-nowrap text-md">{new Date(a.timestamp).toLocaleTimeString()}</p>*/}
                                </div>
                            )
                        })}
                    </div>
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.event !== next.event) {
            return false
        } else if (prev.viewing !== next.viewing) {
            return false
        } else if (prev.folderPath !== next.folderPath) {
            return false
        }

        return true
    }
)

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
    const pastTimestamp = useFileBrowserStore((state) =>
        state.viewingPast?.getTime()
    )

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
    }, [filesMap])

    const { events, epoch } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return { events: [], epoch: null }
        }

        const events: fileAction[][] = []
        let epoch: fileAction

        fileHistory.forEach((a) => {
            if (a.destinationId === contentId) {
                epoch = a
                return
            }
            if (a.destinationId === user.trashId) {
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
        <div className="flex flex-col items-center p-2 overflow-scroll h-[200px] grow">
            {events.map((e) => {
                return (
                    <HistoryEventRow
                        key={e[0].eventId}
                        event={e}
                        folderPath={epoch.destinationPath}
                        viewing={e[0].timestamp === pastTimestamp}
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
