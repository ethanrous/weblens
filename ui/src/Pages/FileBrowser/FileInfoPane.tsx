import { Divider, Text } from '@mantine/core';
import { FBDispatchT, UserContextT, UserInfoT } from '../../types/Types';
import { WeblensFile } from '../../Files/File';
import { useResizeDrag, useWindowSize } from '../../components/hooks';
import { memo, useContext, useMemo, useState } from 'react';
import {
    IconArrowRight,
    IconCaretDown,
    IconCaretRight,
    IconChevronLeft,
    IconChevronRight,
    IconFile,
    IconFileMinus,
    IconFilePlus,
    IconFolder,
    IconReorder,
} from '@tabler/icons-react';
import { FileIcon } from './FileBrowserStyles';
import { clamp } from '../../util';
import WeblensButton from '../../components/WeblensButton';
import { getFileHistory } from '../../api/FileBrowserApi';
import { UserContext } from '../../Context';
import './style/history.css';
import { useQuery } from '@tanstack/react-query';

const SIDEBAR_BREAKPOINT = 650;

export const FilesPane = memo(
    ({
        selectedFiles,
        contentId,
        timestamp,
        dispatch,
    }: {
        selectedFiles: WeblensFile[];
        contentId: string;
        timestamp: number;
        dispatch: FBDispatchT;
    }) => {
        const windowSize = useWindowSize();
        const [resizing, setResizing] = useState(false);
        const [resizeOffset, setResizeOffset] = useState(windowSize?.width > SIDEBAR_BREAKPOINT ? 450 : 75);
        const [open, setOpen] = useState<boolean>(false);
        const [tab, setTab] = useState('info');
        useResizeDrag(resizing, setResizing, v => setResizeOffset(clamp(v, 300, 800)), true);

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
                            setOpen(!open);
                        }}
                    />
                </div>
                <div
                    draggable={false}
                    className="resize-bar-wrapper"
                    onMouseDown={e => {
                        e.preventDefault();
                        setResizing(true);
                    }}
                >
                    <div className="resize-bar" />
                </div>
                <div className="flex flex-col w-[75px] grow h-full">
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
                    {tab === 'info' && <FileInfo selectedFiles={selectedFiles} />}
                    {tab === 'history' && <FileHistory fileId={contentId} timestamp={timestamp} dispatch={dispatch} />}
                </div>
            </div>
        );
    },
    (prev, next) => {
        if (prev.contentId !== next.contentId) {
            return false;
        }

        if (prev.selectedFiles !== next.selectedFiles) {
            return false;
        }

        if (prev.timestamp !== next.timestamp) {
            return false;
        }

        return true;
    },
);

function FileInfo({ selectedFiles }) {
    const titleText = useMemo(() => {
        if (selectedFiles.length === 0) {
            return 'No files selected';
        } else if (selectedFiles.length === 1) {
            return selectedFiles[0].filename;
        } else {
            return `${selectedFiles.length} files selected`;
        }
    }, [selectedFiles]);

    const singleItem = selectedFiles.length === 1;
    const itemIsFolder = selectedFiles[0]?.isDir;

    return (
        <div className="file-info-content">
            <div className="flex flex-row h-[58px] w-full items-center justify-between">
                <Text size="24px" fw={600} style={{ textWrap: 'nowrap', paddingRight: 32 }}>
                    {titleText}
                </Text>
            </div>
            {selectedFiles.length > 0 && (
                <div className="h-max">
                    <div className="flex flex-row h-full w-full items-center">
                        {singleItem && itemIsFolder && <IconFolder size={'48px'} />}
                        {(!singleItem || !itemIsFolder) && <IconFile size={'48px'} />}
                        <Text size="20px">
                            {itemIsFolder ? 'Folder' : 'File'}
                            {singleItem ? '' : 's'} Info
                        </Text>
                    </div>
                    <Divider h={2} w={'90%'} m={10} />
                </div>
            )}
        </div>
    );
}

const historyDate = (timestamp: number) => {
    if (timestamp < 10000000000) {
        timestamp = timestamp * 1000;
    }
    const dateObj = new Date(timestamp);
    const options: Intl.DateTimeFormatOptions = {
        month: 'long',
        day: 'numeric',
        minute: 'numeric',
        hour: 'numeric',
    };
    if (dateObj.getFullYear() !== new Date().getFullYear()) {
        options.year = 'numeric';
    }
    return dateObj.toLocaleDateString('en-US', options);
};

const portableToFolderName = (path: string) => {
    const folderPath = path.slice(path.indexOf(':') + 1, path.lastIndexOf('/'));
    const folderName = folderPath.slice(folderPath.lastIndexOf('/') + 1);
    return folderName;
};
const fileBase = (path: string) => {
    if (path[path.length - 1] === '/') {
        path = path.slice(0, path.length - 1);
    }
    return path.slice(path.lastIndexOf('/') + 1);
};

const HistoryEventRow = ({
    event,
    folderPath,
    viewing,
    usr,
    dispatch,
}: {
    event: fileAction[];
    viewing: boolean;
    folderPath: string;
    usr: UserInfoT;
    dispatch: FBDispatchT;
}) => {
    const [open, setOpen] = useState(false);
    const timeStr = historyDate(event[0].timestamp);

    const folderName = portableToFolderName(folderPath);
    return (
        <div
            className="flex flex-col w-full h-max justify-center p-2 rounded-lg"
            style={{
                backgroundColor: viewing ? '$dark-paper' : '',
                outline: viewing ? '1px solid #4444ff' : '',
            }}
        >
            <div className="file-history-summary">
                <div className="file-history-accordion-header" onClick={() => setOpen(!open)}>
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
                        {event.length} File
                        {event.length !== 1 ? 's' : ''} {event[0].actionType.slice(4)}d
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
                            past: viewing ? null : new Date(event[0].timestamp),
                        })
                    }
                >
                    {timeStr}
                </Text>
            </div>
            <div
                className="file-history-detail-accordion"
                data-open={open}
                style={{ height: open ? event.length * 30 + 35 : 0 }}
            >
                <div key={`${event[0].eventId}-detail`} className={'file-history-detail'}>
                    {event.map((a, i) => {
                        const fromFolder = portableToFolderName(a.originPath);
                        const toFolder = portableToFolderName(a.destinationPath);

                        const fromFile = fileBase(a.originPath);
                        const toFile = fileBase(a.destinationPath);

                        return (
                            <div key={`${fromFile}-${toFile}-${i}`} className="history-detail-action-row">
                                {a.actionType === 'fileMove' && folderName === fromFolder && (
                                    <FileIcon
                                        key={`${a.originId}-move-1`}
                                        id={a.originId}
                                        fileName={fromFile}
                                        Icon={IconFile}
                                        usr={usr}
                                    />
                                )}
                                {a.actionType === 'fileMove' && folderName !== fromFolder && (
                                    <FileIcon
                                        key={`${a.originId}-move-2`}
                                        id={a.originId}
                                        fileName={fromFolder}
                                        Icon={IconFolder}
                                        usr={usr}
                                        as={toFile !== fromFile ? fromFile : ''}
                                    />
                                )}
                                {a.actionType === 'fileCreate' && (
                                    <FileIcon
                                        key={`${a.destinationId}-create-1`}
                                        id={a.destinationId}
                                        fileName={toFile}
                                        Icon={IconFilePlus}
                                        usr={usr}
                                    />
                                )}
                                {a.actionType === 'fileDelete' && (
                                    <FileIcon
                                        key={`${a.originId}-delete-1`}
                                        fileName={fromFile}
                                        Icon={IconFileMinus}
                                        id={a.originId}
                                        usr={usr}
                                    />
                                    // <IconFileMinus className="icon-noshrink" />
                                )}
                                {a.actionType === 'fileRestore' && (
                                    <FileIcon
                                        key={`${a.destinationId}-restore-1`}
                                        fileName={toFile}
                                        Icon={IconReorder}
                                        id={a.destinationId}
                                        usr={usr}
                                    />
                                    // <IconFileMinus className="icon-noshrink" />
                                )}
                                {a.actionType === 'fileMove' && <IconArrowRight className="icon-noshrink" />}

                                {a.actionType === 'fileMove' && folderName === toFolder && (
                                    <FileIcon
                                        key={`${a.originId}-move-3`}
                                        id={a.originId}
                                        fileName={toFile}
                                        Icon={IconFile}
                                        usr={usr}
                                    />
                                )}

                                {a.actionType === 'fileMove' && folderName !== toFolder && (
                                    <FileIcon
                                        key={`${a.originId}-move-4`}
                                        id={a.originId}
                                        fileName={toFolder}
                                        Icon={IconFolder}
                                        usr={usr}
                                        as={toFile !== fromFile ? toFile : ''}
                                    />
                                )}

                                <Text style={{ textWrap: 'nowrap' }}>{new Date(a.timestamp).toLocaleTimeString()}</Text>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

type fileAction = {
    actionType: string;
    destinationId: string;
    destinationPath: string;
    eventId: string;
    lifeId: string;
    originId: string;
    originPath: string;
    timestamp: number;
};

function FileHistory({ fileId, timestamp, dispatch }: { fileId: string; timestamp: number; dispatch: FBDispatchT }) {
    const { authHeader, usr }: UserContextT = useContext(UserContext);
    // const [fileHistory, setFileHistory]: [
    //     fileHistory: FileEventT[],
    //     setFileHistory: any,
    // ] = useState([])
    //
    // useEffect(() => {
    //     setFileHistory([])
    //     getFileHistory(fileId, authHeader).then((r) => setFileHistory(r.events))
    // }, [fileId, timestamp])

    const { data: fileHistory } = useQuery<fileAction[]>({
        queryKey: ['fileHistory'],
        queryFn: () => getFileHistory(fileId, authHeader),
    });

    const { events, epoch } = useMemo(() => {
        if (!fileHistory || !fileHistory.length) {
            return { events: [], epoch: null };
        }

        const events: fileAction[][] = [];
        let epoch: fileAction;

        fileHistory.forEach(a => {
            if (a.destinationId === fileId) {
                epoch = a;
                return;
            }

            const i = events.findLastIndex(e => e[0].eventId === a.eventId);
            if (i != -1) {
                events[i].push(a);
            } else {
                events.push([a]);
            }
        });

        return { events, epoch };
    }, [fileHistory]);

    if (!epoch) {
        return null;
    }

    const createTimeString = historyDate(epoch.timestamp);

    return (
        <div className="flex flex-col items-center p-2 overflow-scroll h-[200px] grow">
            {events.map((e, i) => {
                return (
                    <HistoryEventRow
                        key={e[0].eventId}
                        event={e}
                        folderPath={epoch.destinationPath}
                        viewing={e[0].timestamp === timestamp}
                        usr={usr}
                        dispatch={dispatch}
                    />
                );
            })}
            <div className="flex flex-row items-center p-2 pt-10">
                <FileIcon id={epoch.destinationId} fileName={epoch.destinationPath} Icon={IconFolder} usr={usr} />

                <Text style={{ textWrap: 'nowrap' }}>created on {createTimeString}</Text>
            </div>
        </div>
    );
}
