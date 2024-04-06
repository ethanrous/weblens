import {
    Box,
    Card,
    MantineStyleProp,
    Text,
    Tooltip,
    ActionIcon,
    Space,
    Menu,
    Divider,
    FileButton,
    Center,
    Skeleton,
} from '@mantine/core';
import { useCallback, useContext, useMemo, useState } from 'react';
import {
    handleDragOver,
    HandleDrop,
    HandleUploadButton,
} from './FileBrowserLogic';
import {
    FBDispatchT,
    ScanMeta,
    FileInfoT,
    FbStateT,
    UserContextT,
} from '../../types/Types';

import {
    IconFile,
    IconFileZip,
    IconFolder,
    IconFolderCancel,
    IconFolderPlus,
    IconPhoto,
    IconRefresh,
    IconSpiral,
    IconUpload,
} from '@tabler/icons-react';
import { userContext } from '../../Context';
import '../../components/style.css';
import '../../components/filebrowserStyle.css';
import { humanFileSize, nsToHumanTime } from '../../util';
import { ContainerMedia } from '../../components/Presentation';
import { IconX } from '@tabler/icons-react';
import { WeblensProgress } from '../../components/WeblensProgress';
import { useResize } from '../../components/hooks';

export const ColumnBox = ({
    children,
    style,
    reff,
    className,
    onClick,
    onMouseOver,
    onMouseLeave,
    onContextMenu,
    onBlur,
    onDragOver,
    onMouseUp,
}: {
    children?;
    style?: MantineStyleProp;
    reff?;
    className?: string;
    onClick?;
    onMouseOver?;
    onMouseLeave?;
    onContextMenu?;
    onBlur?;
    onDragOver?;
    onMouseUp?;
}) => {
    return (
        <Box
            draggable={false}
            ref={reff}
            children={children}
            onClick={onClick}
            onMouseOver={onMouseOver}
            onMouseLeave={onMouseLeave}
            onContextMenu={onContextMenu}
            onBlur={onBlur}
            onDrag={(e) => e.preventDefault()}
            onDragOver={onDragOver}
            onMouseUp={onMouseUp}
            style={{
                display: 'flex',
                height: '100%',
                width: '100%',
                flexDirection: 'column',
                alignItems: 'center',
                ...style,
            }}
            className={`column-box ${className ? className : ''}`}
        />
    );
};

export const RowBox = ({
    children,
    style,
    onClick,
    onBlur,
}: {
    children;
    style?: MantineStyleProp;
    onClick?;
    onBlur?;
}) => {
    return (
        <Box
            draggable={false}
            children={children}
            onClick={onClick}
            onBlur={onBlur}
            onDrag={(e) => e.preventDefault()}
            style={{
                height: '100%',
                width: '100%',
                display: 'flex',
                flexDirection: 'row',
                alignItems: 'center',
                ...style,
            }}
        />
    );
};

export const TransferCard = ({
    action,
    destination,
    boundRef,
}: {
    action: string;
    destination: string;
    boundRef?;
}) => {
    let width;
    let left;
    if (boundRef) {
        width = boundRef.clientWidth;
        left = boundRef.getBoundingClientRect()['left'];
    }
    if (!destination) {
        return;
    }

    return (
        <Box
            className="transfer-info-box"
            style={{
                pointerEvents: 'none',
                width: width ? width : '100%',
                left: left ? left : 0,
            }}
        >
            <Card style={{ height: 'max-content' }}>
                <RowBox>
                    <Text style={{ userSelect: 'none' }}>{action} to</Text>
                    <IconFolder style={{ marginLeft: '7px' }} />
                    <Text
                        fw={700}
                        style={{ marginLeft: 3, userSelect: 'none' }}
                    >
                        {destination}
                    </Text>
                </RowBox>
            </Card>
        </Box>
    );
};

const Dropspot = ({
    onDrop,
    dropspotTitle,
    dragging,
    dropAllowed,
    handleDrag,
    wrapperRef,
}: {
    onDrop;
    dropspotTitle;
    dragging;
    dropAllowed;
    handleDrag: React.DragEventHandler<HTMLDivElement>;
    wrapperRef?;
}) => {
    return (
        <Box
            draggable={false}
            className="dropspot-wrapper"
            onDragOver={(e) => {
                if (dragging === 0) {
                    handleDrag(e);
                }
            }}
            style={{
                pointerEvents: dragging === 2 ? 'all' : 'none',
                cursor: !dropAllowed && dragging === 2 ? 'no-drop' : 'auto',
                height: wrapperRef ? wrapperRef.clientHeight : '100%',
                width: wrapperRef ? wrapperRef.clientWidth : '100%',
            }}
            onDragLeave={handleDrag}
        >
            {dragging === 2 && (
                <Box
                    className="dropbox"
                    onMouseLeave={handleDrag}
                    onDrop={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        if (dropAllowed) {
                            onDrop(e);
                        }
                    }}
                    // required for onDrop to work https://stackoverflow.com/questions/50230048/react-ondrop-is-not-firing
                    onDragOver={(e) => e.preventDefault()}
                    style={{
                        outlineColor: `${dropAllowed ? '#ffffff' : '#dd2222'}`,
                        cursor:
                            !dropAllowed && dragging === 2 ? 'no-drop' : 'auto',
                    }}
                >
                    {!dropAllowed && (
                        <ColumnBox
                            style={{
                                position: 'relative',
                                justifyContent: 'center',
                                cursor: 'no-drop',
                                width: 'max-content',
                                pointerEvents: 'none',
                            }}
                        >
                            <IconFolderCancel size={100} color="#dd2222" />
                        </ColumnBox>
                    )}
                    {dropAllowed && (
                        <TransferCard
                            action="Upload"
                            destination={dropspotTitle}
                        />
                    )}
                </Box>
            )}
        </Box>
    );
};

const FilebrowserMenu = ({
    folderName,
    menuPos,
    menuOpen,
    setMenuOpen,
    newFolder,
}) => {
    return (
        <Menu opened={menuOpen} onClose={() => setMenuOpen(false)}>
            <Menu.Target>
                <Box
                    style={{
                        position: 'fixed',
                        top: menuPos.y,
                        left: menuPos.x,
                    }}
                />
            </Menu.Target>

            <Menu.Dropdown>
                <Menu.Label>{folderName}</Menu.Label>
                <Menu.Item
                    leftSection={<IconFolderPlus />}
                    onClick={() => newFolder()}
                >
                    New Folder
                </Menu.Item>
            </Menu.Dropdown>
        </Menu>
    );
};

type DirViewWrapperProps = {
    fbState: FbStateT;
    folderName: string;
    dragging: number;
    dispatch: FBDispatchT;
    onDrop: (e: any) => void;
    children: JSX.Element;
};

export const DirViewWrapper = ({
    fbState,
    folderName,
    dragging,
    dispatch,
    onDrop,
    children,
}: DirViewWrapperProps) => {
    const { usr }: UserContextT = useContext(userContext);
    const [menuOpen, setMenuOpen] = useState(false);
    const [menuPos, setMenuPos] = useState({ x: 0, y: 0 });
    const [wrapperRef, setWrapperRef] = useState(null);
    const dropAllowed = useMemo(() => {
        return !(
            fbState.fbMode === 'share' || fbState.contentId === usr.trashId
        );
    }, [fbState.contentId, usr.trashId, fbState.fbMode]);

    return (
        <Box
            draggable={false}
            style={{
                // marginRight: 10,
                height: '100%',
                flexShrink: 0,
                minWidth: 400,
                flexGrow: 1,
                width: 0,
            }}
            onDrag={(e) => {
                e.preventDefault();
                e.stopPropagation();
            }}
            ref={setWrapperRef}
            // If dropping is not allowed, and we drop, we want to clear the window when we detect the mouse moving again
            // We have to wait (a very short time, 10ms) to make sure the drop event fires and gets captured by the dropbox, otherwise
            // we set dragging to 0 too early, the dropbox gets removed, and chrome handles the drop event, opening the image in another tab
            // onMouseMove={e => { if (dragging) { setTimeout(() => dispatch({ type: 'set_dragging', dragging: false }), 10) } }}

            onMouseUp={(e) => {
                if (dragging) {
                    setTimeout(
                        () =>
                            dispatch({ type: 'set_dragging', dragging: false }),
                        10
                    );
                }
            }}
            onClick={(e) => {
                if (dragging) {
                    return;
                }
                dispatch({ type: 'clear_selected' });
            }}
            onContextMenu={(e) => {
                e.preventDefault();
                if (fbState.fbMode === 'share') {
                    return;
                }
                setMenuPos({ x: e.clientX, y: e.clientY });
                setMenuOpen(true);
            }}
        >
            <FilebrowserMenu
                folderName={folderName === usr.username ? 'Home' : folderName}
                menuPos={menuPos}
                menuOpen={menuOpen}
                setMenuOpen={setMenuOpen}
                newFolder={() => dispatch({ type: 'new_dir' })}
            />
            <Dropspot
                onDrop={(e) => {
                    onDrop(e);
                    dispatch({ type: 'set_dragging', dragging: false });
                }}
                dropspotTitle={folderName}
                dragging={dragging}
                dropAllowed={dropAllowed}
                handleDrag={(event) =>
                    handleDragOver(event, dispatch, dragging)
                }
                wrapperRef={wrapperRef}
            />
            <ColumnBox
                style={{ width: '100%', padding: 8 }}
                onDragOver={(event) => {
                    if (!dragging) {
                        handleDragOver(event, dispatch, dragging);
                    }
                }}
            >
                {children}
            </ColumnBox>
        </Box>
    );
};

export const WormholeWrapper = ({
    wormholeId,
    wormholeName,
    fileId,
    validWormhole,
    uploadDispatch,
    children,
}: {
    wormholeId: string;
    wormholeName: string;
    fileId: string;
    validWormhole: boolean;
    uploadDispatch;
    children;
}) => {
    const { authHeader }: UserContextT = useContext(userContext);
    const [dragging, setDragging] = useState(0);
    const handleDrag = useCallback(
        (e) => {
            e.preventDefault();
            if (e.type === 'dragenter' || e.type === 'dragover') {
                if (!dragging) {
                    setDragging(2);
                }
            } else if (dragging) {
                setDragging(0);
            }
        },
        [dragging]
    );

    return (
        <Box className="wormhole-wrapper">
            <Box
                style={{ position: 'relative', width: '98%', height: '98%' }}
                //                    See DirViewWrapper \/
                onMouseMove={(e) => {
                    if (dragging) {
                        setTimeout(() => setDragging(0), 10);
                    }
                }}
            >
                <Dropspot
                    onDrop={(e) =>
                        HandleDrop(
                            e.dataTransfer.items,
                            fileId,
                            [],
                            true,
                            wormholeId,
                            authHeader,
                            uploadDispatch,
                            () => {}
                        )
                    }
                    dropspotTitle={wormholeName}
                    dragging={dragging}
                    dropAllowed={validWormhole}
                    handleDrag={handleDrag}
                />
                <ColumnBox
                    style={{ justifyContent: 'center' }}
                    onDragOver={handleDrag}
                >
                    {children}
                </ColumnBox>
            </Box>
        </Box>
    );
};

export const ScanFolderButton = ({ folderId, holdingShift, doScan }) => {
    return (
        <Box>
            {folderId !== 'shared' && folderId !== 'trash' && (
                <Tooltip
                    label={holdingShift ? 'Deep scan folder' : 'Scan folder'}
                >
                    <ActionIcon color="#00000000" size={35} onClick={doScan}>
                        <IconRefresh
                            color={holdingShift ? '#4444ff' : 'white'}
                            size={35}
                        />
                    </ActionIcon>
                </Tooltip>
            )}
            {(folderId === 'shared' || folderId === 'trash') && (
                <Space w={35} />
            )}
        </Box>
    );
};

export const FolderIcon = ({ shares, size }: { shares; size }) => {
    const [copied, setCopied] = useState(false);
    const wormholeId = useMemo(() => {
        if (shares) {
            const whs = shares.filter((s) => s.Wormhole);
            if (whs.length !== 0) {
                return whs[0].shareId;
            }
        }
    }, [shares]);
    return (
        <RowBox style={{ justifyContent: 'center', width: '100%' }}>
            <IconFolder size={size} />
            {wormholeId && (
                <Tooltip label={copied ? 'Copied' : 'Copy Wormhole'}>
                    <IconSpiral
                        color={copied ? '#4444ff' : 'white'}
                        style={{ position: 'absolute', right: 0, top: 0 }}
                        onClick={(e) => {
                            e.stopPropagation();
                            navigator.clipboard.writeText(
                                `${window.location.origin}/wormhole/${shares[0].ShareId}`
                            );
                            setCopied(true);
                            setTimeout(() => setCopied(false), 1000);
                        }}
                        onDoubleClick={(e) => e.stopPropagation()}
                    />
                </Tooltip>
            )}
        </RowBox>
    );
};

export const IconDisplay = ({
    file,
    size = 24,
    allowMedia = false,
}: {
    file: FileInfoT;
    size?: string | number;
    allowMedia?: boolean;
}) => {
    const [containerRef, setContainerRef] = useState(null);
    const containerSize = useResize(containerRef);

    if (!file) {
        return null;
    }

    if (file.isDir) {
        return <FolderIcon shares={file.shares} size={size} />;
    }

    if (!file.imported && file.displayable && allowMedia) {
        return (
            <Center style={{ height: '100%', width: '100%' }}>
                <Skeleton height={'100%'} width={'100%'} />
                <Text pos={'absolute'} style={{ userSelect: 'none' }}>
                    Processing...
                </Text>
            </Center>
        );
    } else if (file.displayable && allowMedia) {
        return (
            <ColumnBox
                reff={setContainerRef}
                style={{ justifyContent: 'center' }}
            >
                <ContainerMedia
                    mediaData={file.mediaData}
                    containerRef={containerRef}
                />
            </ColumnBox>
            // <MediaImage media={file.mediaData} quality={quality} />
        );
    } else if (file.displayable) {
        return <IconPhoto />;
    }
    const extIndex = file.filename.lastIndexOf('.');
    const ext = file.filename.slice(extIndex + 1, file.filename.length);
    const textSize = `${Math.floor(containerSize?.width / (ext.length + 5))}px`;

    switch (ext) {
        case 'zip':
            return <IconFileZip />;
        default:
            return (
                <Box
                    ref={setContainerRef}
                    style={{
                        display: 'flex',
                        justifyContent: 'center',
                        alignItems: 'center',
                        width: '100%',
                        height: '100%',
                    }}
                >
                    <IconFile size={size} />
                    {extIndex !== -1 && (
                        <Text
                            size={textSize}
                            fw={700}
                            style={{
                                position: 'absolute',
                                userSelect: 'none',
                                WebkitUserSelect: 'none',
                            }}
                        >
                            .{ext}
                        </Text>
                    )}
                </Box>
            );
    }
};

export const FileInfoDisplay = ({ file }: { file: FileInfoT }) => {
    let [size, units] = humanFileSize(file.size);
    return (
        <ColumnBox
            style={{
                width: 'max-content',
                whiteSpace: 'nowrap',
                justifyContent: 'center',
                maxWidth: '100%',
            }}
        >
            <Text fw={600} style={{ fontSize: '2.5vw', maxWidth: '100%' }}>
                {file.filename}
            </Text>
            {file.isDir && (
                <RowBox
                    style={{
                        height: 'max-content',
                        justifyContent: 'center',
                        width: '100%',
                    }}
                >
                    <Text style={{ fontSize: '25px', maxWidth: '100%' }}>
                        {file.children.length} Item
                        {file.children.length !== 1 ? 's' : ''}
                    </Text>
                    <Divider orientation="vertical" size={2} mx={10} />
                    <Text style={{ fontSize: '25px' }}>
                        {size}
                        {units}
                    </Text>
                </RowBox>
            )}
            {!file.isDir && (
                <Text style={{ fontSize: '25px' }}>
                    {size}
                    {units}
                </Text>
            )}
        </ColumnBox>
    );
};

export const PresentationFile = ({ file }: { file: FileInfoT }) => {
    if (!file) {
        return null;
    }
    let [size, units] = humanFileSize(file.size);
    if (file.displayable && file.mediaData) {
        return (
            <ColumnBox
                style={{
                    justifyContent: 'center',
                    width: '40%',
                    height: 'max-content',
                }}
                onClick={(e) => e.stopPropagation()}
            >
                <Text
                    fw={600}
                    style={{ fontSize: '2.1vw', wordBreak: 'break-all' }}
                >
                    {file.filename}
                </Text>
                <Text style={{ fontSize: '25px' }}>
                    {size}
                    {units}
                </Text>
                <Text style={{ fontSize: '20px' }}>
                    {new Date(Date.parse(file.modTime)).toLocaleDateString(
                        'en-us',
                        {
                            year: 'numeric',
                            month: 'short',
                            day: 'numeric',
                        }
                    )}
                </Text>
                <Divider />
                <Text style={{ fontSize: '20px' }}>
                    {new Date(
                        Date.parse(file.mediaData.createDate)
                    ).toLocaleDateString('en-us', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                    })}
                </Text>
            </ColumnBox>
        );
    } else {
        return (
            <RowBox
                style={{ justifyContent: 'center', height: 'max-content' }}
                onClick={(e) => e.stopPropagation()}
            >
                <Box
                    style={{
                        width: '60%',
                        display: 'flex',
                        justifyContent: 'center',
                    }}
                >
                    <IconDisplay file={file} allowMedia />
                </Box>
                <Space w={30} />
                <ColumnBox style={{ width: '40%', justifyContent: 'center' }}>
                    <Text fw={600} style={{ width: '100%' }}>
                        {file.filename}
                    </Text>
                    {file.isDir && (
                        <RowBox
                            style={{
                                height: 'max-content',
                                justifyContent: 'center',
                                width: '50vw',
                            }}
                        >
                            <Text style={{ fontSize: '25px' }}>
                                {file.children.length} Item
                                {file.children.length !== 1 ? 's' : ''}
                            </Text>
                            <Divider orientation="vertical" size={2} mx={10} />
                            <Text style={{ fontSize: '25px' }}>
                                {size}
                                {units}
                            </Text>
                        </RowBox>
                    )}
                    {!file.isDir && (
                        <Text style={{ fontSize: '25px' }}>
                            {size}
                            {units}
                        </Text>
                    )}
                </ColumnBox>
            </RowBox>
        );
    }
};

export const GetStartedCard = ({
    fb,
    moveSelectedTo,
    dispatch,
    uploadDispatch,
    authHeader,
    wsSend,
}) => {
    return (
        <ColumnBox>
            <ColumnBox
                style={{
                    width: 'max-content',
                    height: 'max-content',
                    marginTop: '20vh',
                }}
            >
                <Text size="28px" style={{ width: 'max-content' }}>
                    This folder is empty
                </Text>

                {fb.folderInfo.modifiable && (
                    <RowBox style={{ padding: 10 }}>
                        <FileButton
                            onChange={(files) => {
                                HandleUploadButton(
                                    files,
                                    fb.folderInfo.id,
                                    false,
                                    '',
                                    authHeader,
                                    uploadDispatch,
                                    wsSend
                                );
                            }}
                            accept="file"
                            multiple
                        >
                            {(props) => {
                                return (
                                    <ColumnBox
                                        onClick={() => {
                                            props.onClick();
                                        }}
                                        style={{
                                            cursor: 'pointer',
                                            padding: 10,
                                        }}
                                    >
                                        <IconUpload
                                            size={100}
                                            style={{ padding: '10px' }}
                                        />
                                        <Text size="20px" fw={600}>
                                            Upload
                                        </Text>
                                        <Space h={4}></Space>
                                        <Text size="12px">Click or Drop</Text>
                                    </ColumnBox>
                                );
                            }}
                        </FileButton>
                        <Divider orientation="vertical" m={30} />

                        <ColumnBox
                            onClick={(e) => {
                                e.stopPropagation();
                                dispatch({ type: 'new_dir' });
                            }}
                            style={{ cursor: 'pointer', padding: 10 }}
                        >
                            <IconFolderPlus
                                size={100}
                                style={{ padding: '10px' }}
                            />
                            <Text
                                size="20px"
                                fw={600}
                                style={{ width: 'max-content' }}
                            >
                                New Folder
                            </Text>
                        </ColumnBox>
                    </RowBox>
                )}
            </ColumnBox>
        </ColumnBox>
    );
};

export const TaskProgCard = ({
    prog,
    dispatch,
}: {
    prog: ScanMeta;
    dispatch: FBDispatchT;
}) => {
    const timeString = useMemo(() => nsToHumanTime(prog.time), [prog.time]);

    return (
        <Box className="task-progress-box">
            <RowBox style={{ height: 'max-content' }}>
                <Box style={{ width: '100%' }}>
                    <Text size="12px">{prog.taskType}</Text>
                    <Text size="16px" fw={600}>
                        {prog.target}
                    </Text>
                </Box>
                <IconX
                    size={20}
                    cursor={'pointer'}
                    onClick={() =>
                        dispatch({
                            type: 'remove_task_progress',
                            taskId: prog.taskId,
                        })
                    }
                />
            </RowBox>
            <Box
                style={{ height: 25, flexShrink: 0, width: '100%', margin: 10 }}
            >
                <WeblensProgress
                    value={prog.complete ? 100 : prog.progress}
                    color={prog.complete ? '#22bb33' : '#4444ff'}
                />
            </Box>
            {!prog.complete && (
                <RowBox
                    style={{
                        justifyContent: 'space-between',
                        height: 'max-content',
                        gap: 10,
                    }}
                >
                    <Text size="10px" truncate="end">
                        {prog.mostRecent}
                    </Text>
                    {prog.tasksTotal > 0 && (
                        <Text size="10px">
                            {prog.tasksComplete}/{prog.tasksTotal}
                        </Text>
                    )}
                </RowBox>
            )}
            {prog.complete && (
                <RowBox
                    style={{
                        justifyContent: 'space-between',
                        height: 'max-content',
                        gap: 10,
                    }}
                >
                    <Text size="10px" style={{ width: 'max-content' }}>
                        Finished in {timeString}
                    </Text>
                    <Text
                        size="10px"
                        style={{ width: 'max-content', textWrap: 'nowrap' }}
                    >
                        {prog.note}
                    </Text>
                </RowBox>
            )}
        </Box>
    );
};
