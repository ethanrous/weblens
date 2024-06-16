import { Divider, Text } from '@mantine/core'
import { FBDispatchT, UserContextT, UserInfoT } from '../../types/Types'
import { WeblensFile } from '../../Files/File'
import { useResizeDrag, useWindowSize } from '../../components/hooks'
import { memo, useContext, useMemo, useState } from 'react'
import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconFile,
    IconFileMinus,
    IconFilePlus,
    IconFolder,
    IconReorder,
} from '@tabler/icons-react'
import { FileIcon } from './FileBrowserStyles'
import { clamp } from '../../util'
import { WeblensButton } from '../../components/WeblensButton'
import { getFileHistory } from '../../api/FileBrowserApi'
import { UserContext } from '../../Context'
import { FileEventT } from './FileBrowserTypes'
import './style/history.css'
import { useQuery } from '@tanstack/react-query'

const SIDEBAR_BREAKPOINT = 650

export const FilesPane = memo(
    ({
        open,
        selectedFiles,
        contentId,
        timestamp,
        dispatch,
    }: {
        open: boolean
        setOpen: (o) => void
        selectedFiles: WeblensFile[]
        contentId: string
        timestamp: number
        dispatch: FBDispatchT
    }) => {
        const windowSize = useWindowSize()
        const [resizing, setResizing] = useState(false)
        const [resizeOffset, setResizeOffset] = useState(
            windowSize?.width > SIDEBAR_BREAKPOINT ? 450 : 75
        )
        const [tab, setTab] = useState('info')
        useResizeDrag(
            resizing,
            setResizing,
            (v) => setResizeOffset(clamp(v, 300, 800)),
            true
        )

        if (!open) {
            return <div className="file-info-pane" />
        }

        return (
            <div
                className="file-info-pane"
                data-resizing={resizing}
                style={{ width: resizeOffset }}
            >
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
                <div className="flex flex-col w-[75px] grow ">
                    <div className="flex flex-row h-max w-full gap-1 justify-between p-1 pl-0">
                        <div className="max-w-[49%]">
                            <WeblensButton
                                label="File Info"
                                squareSize={50}
                                fillWidth
                                toggleOn={tab === 'info'}
                                onClick={() => setTab('info')}
                            />
                        </div>
                        <div className="max-w-[49%]">
                            <WeblensButton
                                label="History"
                                squareSize={50}
                                fillWidth
                                toggleOn={tab === 'history'}
                                onClick={() => setTab('history')}
                            />
                        </div>
                    </div>
                    {tab === 'info' && (
                        <FileInfo selectedFiles={selectedFiles} />
                    )}
                    {tab === 'history' && (
                        <FileHistory
                            fileId={contentId}
                            timestamp={timestamp}
                            dispatch={dispatch}
                        />
                    )}
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.open !== next.open) {
            return false
        }

        if (prev.contentId !== next.contentId) {
            return false
        }

        if (prev.selectedFiles !== next.selectedFiles) {
            return false
        }

        if (prev.timestamp !== next.timestamp) {
            return false
        }

        return true
    }
)

function FileInfo({ selectedFiles }) {
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
                <Text
                    size="24px"
                    fw={600}
                    style={{ textWrap: 'nowrap', paddingRight: 32 }}
                >
                    {titleText}
                </Text>
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
                        <Text size="20px">
                            {itemIsFolder ? 'Folder' : 'File'}
                            {singleItem ? '' : 's'} Info
                        </Text>
                    </div>
                    <Divider h={2} w={'90%'} m={10} />
                </div>
            )}
        </div>
    )
}

const historyDate = (timestamp: number) => {
    const dateObj = new Date(timestamp)
    const options: Intl.DateTimeFormatOptions = {
        month: 'long',
        day: 'numeric',
        minute: 'numeric',
        hour: 'numeric',
    }
    if (dateObj.getFullYear() !== new Date().getFullYear()) {
        options.year = 'numeric'
    }
    return dateObj.toLocaleDateString('en-US', options)
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

const HistoryRow = ({
    eventGroup,
    folderPath,
    viewing,
    usr,
    dispatch,
}: {
    eventGroup: {
        count: number
        action: string
        time: number
        events: FileEventT[]
    }
    viewing: boolean
    folderPath: string
    usr: UserInfoT
    dispatch: FBDispatchT
}) => {
    const [open, setOpen] = useState(false)
    const timeStr = historyDate(eventGroup.time)

    const folderName = portableToFolderName(folderPath)

    // TODO move style into data attributes
    return (
        <div
            className="flex flex-col w-full h-max justify-center p-2 rounded-lg"
            style={{
                backgroundColor: viewing ? '$dark-paper' : '',
                outline: viewing ? '1px solid #4444ff' : '',
            }}
        >
            <div key={eventGroup.time} className="file-history-summary">
                <div
                    className="file-history-accordion-header"
                    onClick={() => setOpen(!open)}
                >
                    {open ? (
                        <IconCaretDown size={20} style={{ flexShrink: 0 }} />
                    ) : (
                        <IconCaretRight size={20} style={{ flexShrink: 0 }} />
                    )}
                    <Text
                        c="white"
                        truncate="end"
                        style={{
                            lineHeight: '20px',
                            fontSize: '20px',
                            width: 'max-content',
                            padding: '0px 5px 0px 5px',

                            textWrap: 'nowrap',
                        }}
                        fw={600}
                    >
                        {eventGroup.count} File
                        {eventGroup.count !== 1 ? 's' : ''}{' '}
                        {eventGroup.action.slice(4)}d
                    </Text>
                </div>
                <div className="grow" />
                <Text
                    className="event-time-string"
                    fw={viewing ? 600 : 400}
                    c={viewing ? 'white' : ''}
                    onClick={() =>
                        dispatch({
                            type: 'set_past_time',
                            past: viewing
                                ? null
                                : new Date(eventGroup.events[0].Timestamp),
                        })
                    }
                >
                    {timeStr}
                </Text>
            </div>
            <div
                className="file-history-detail-accordion"
                data-open={open}
                style={{ height: open ? eventGroup.count * 30 + 35 : 0 }}
            >
                <div className={'file-history-detail'}>
                    {eventGroup.events.map((e) => {
                        const fromFolder = portableToFolderName(e.FromPath)
                        const toFolder = portableToFolderName(e.Path)

                        const fromFile = fileBase(e.FromPath)
                        const toFile = fileBase(e.Path)

                        return (
                            <div
                                key={`${e.Path}-${e.millis}`}
                                className="history-detail-action-row"
                            >
                                {e.Action === 'fileMove' &&
                                    folderName === fromFolder && (
                                        <FileIcon
                                            id={e.FromFileId}
                                            fileName={fromFile}
                                            Icon={IconFile}
                                            usr={usr}
                                        />
                                    )}
                                {e.Action === 'fileMove' &&
                                    folderName !== fromFolder && (
                                        <FileIcon
                                            id={e.FromFileId}
                                            fileName={fromFolder}
                                            Icon={IconFolder}
                                            usr={usr}
                                            as={
                                                toFile !== fromFile
                                                    ? fromFile
                                                    : ''
                                            }
                                        />
                                    )}
                                {e.Action === 'fileCreate' && (
                                    <FileIcon
                                        id={e.FileId}
                                        fileName={toFile}
                                        Icon={IconFilePlus}
                                        usr={usr}
                                    />
                                )}
                                {e.Action === 'fileDelete' && (
                                    <FileIcon
                                        fileName={fromFile}
                                        Icon={IconFileMinus}
                                        id={e.FromFileId}
                                        usr={usr}
                                    />
                                    // <IconFileMinus className="icon-noshrink" />
                                )}
                                {e.Action === 'fileRestore' && (
                                    <FileIcon
                                        fileName={toFile}
                                        Icon={IconReorder}
                                        id={e.FileId}
                                        usr={usr}
                                    />
                                    // <IconFileMinus className="icon-noshrink" />
                                )}
                                {e.Action === 'fileMove' && (
                                    <IconArrowRight className="icon-noshrink" />
                                )}

                                {e.Action === 'fileMove' &&
                                    folderName === toFolder && (
                                        <FileIcon
                                            id={e.FileId}
                                            fileName={toFile}
                                            Icon={IconFile}
                                            usr={usr}
                                        />
                                    )}

                                {e.Action === 'fileMove' &&
                                    folderName !== toFolder && (
                                        <FileIcon
                                            id={e.FileId}
                                            fileName={toFolder}
                                            Icon={IconFolder}
                                            usr={usr}
                                            as={
                                                toFile !== fromFile
                                                    ? toFile
                                                    : ''
                                            }
                                        />
                                    )}

                                <Text style={{ textWrap: 'nowrap' }}>
                                    {new Date(e.millis).toLocaleTimeString()}
                                </Text>
                            </div>
                        )
                    })}
                </div>
            </div>
        </div>
    )
}

function FileHistory({
    fileId,
    timestamp,
    dispatch,
}: {
    fileId: string
    timestamp: number
    dispatch: FBDispatchT
}) {
    const { authHeader, usr }: UserContextT = useContext(UserContext)
    // const [fileHistory, setFileHistory]: [
    //     fileHistory: FileEventT[],
    //     setFileHistory: any,
    // ] = useState([])
    //
    // useEffect(() => {
    //     setFileHistory([])
    //     getFileHistory(fileId, authHeader).then((r) => setFileHistory(r.events))
    // }, [fileId, timestamp])

    const { data: fileHistory } = useQuery({
        queryKey: ['fileHistory'],
        queryFn: () => getFileHistory(fileId, authHeader).then((r) => r.events),
    })

    const { createEvent, eventGroups } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return {}
        }

        const createEvent = fileHistory.shift()

        const groupMap = new Map<string, FileEventT[]>()
        fileHistory.forEach((e) => {
            let millis = Date.parse(e.Timestamp)
            let groupMillis = millis - (millis % 60000)

            e.millis = millis
            const groupKey = `${groupMillis}-${e.Action}`
            let group = groupMap.get(groupKey)
            if (group === undefined) {
                group = [e]
            } else {
                group.push(e)
            }
            groupMap.set(groupKey, group)
        })

        const groups = Array.from(groupMap.values())

        return {
            createEvent: createEvent,
            eventGroups: groups
                .map((g) => {
                    g.sort((a, b) => b.millis - a.millis)
                    return {
                        count: g.length,
                        action: g[0].Action,
                        time: g[0].millis,
                        events: g,
                    }
                })
                .sort((a, b) => b.time - a.time),
        }
    }, [fileHistory])

    if (!createEvent) {
        return null
    }

    const createTimeString = historyDate(Date.parse(createEvent.Timestamp))

    return (
        <div className="flex flex-col items-center p-2 overflow-scroll h-[200px] grow">
            {eventGroups.map((eg) => {
                return (
                    <HistoryRow
                        key={eg.events[0].FileId + eg.time}
                        eventGroup={eg}
                        folderPath={createEvent.Path}
                        viewing={eg.time === timestamp}
                        usr={usr}
                        dispatch={dispatch}
                    />
                )
            })}
            <div className="flex flex-row items-center p-2 pt-10">
                <FileIcon
                    id={createEvent.FileId}
                    fileName={createEvent.Path}
                    Icon={IconFolder}
                    usr={usr}
                />

                <Text style={{ textWrap: 'nowrap' }}>
                    created on {createTimeString}
                </Text>
            </div>
        </div>
    )
}
