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
} from '@mantine/core'
import { memo, useContext, useMemo, useState } from 'react'
import {
    handleDragOver,
    HandleDrop,
    HandleUploadButton,
} from './FileBrowserLogic'
import {
    FBDispatchT,
    FbStateT,
    UserContextT,
    UserInfoT,
} from '../../types/Types'
import { WeblensFile } from '../../classes/File'

import {
    IconDatabase,
    IconFile,
    IconFileZip,
    IconFolder,
    IconFolderCancel,
    IconFolderPlus,
    IconHome,
    IconHome2,
    IconPhoto,
    IconRefresh,
    IconServer,
    IconServer2,
    IconSpiral,
    IconTrash,
    IconUpload,
    IconUser,
    IconUsers,
} from '@tabler/icons-react'
import { UserContext } from '../../Context'
import { friendlyFolderName, humanFileSize, nsToHumanTime } from '../../util'
import { ContainerMedia } from '../../components/Presentation'
import { IconX } from '@tabler/icons-react'
import { WeblensProgress } from '../../components/WeblensProgress'
import { useResize } from '../../components/hooks'
import { BackdropMenu } from './FileMenu'

import './style/fileBrowserStyle.css'
import { DraggingState, FbContext } from './FileBrowser'

// export const Box = ({
//     children,
//     style,
//     reff,
//     className,
//     onClick,
//     onMouseOver,
//     onMouseLeave,
//     onContextMenu,
//     onBlur,
//     onDragOver,
//     onMouseUp,
// }: {
//     children?;
//     style?: MantineStyleProp;
//     reff?;
//     className?: string;
//     onClick?;
//     onMouseOver?;
//     onMouseLeave?;
//     onContextMenu?;
//     onBlur?;
//     onDragOver?;
//     onMouseUp?;
// }) => {
//     return (
//         <Box
//             draggable={false}
//             ref={reff}
//             children={children}
//             onClick={onClick}
//             onMouseOver={onMouseOver}
//             onMouseLeave={onMouseLeave}
//             onContextMenu={onContextMenu}
//             onBlur={onBlur}
//             onDrag={(e) => e.preventDefault()}
//             onDragOver={onDragOver}
//             onMouseUp={onMouseUp}
//             style={{
//                 display: "flex",
//                 height: "100%",
//                 width: "100%",
//                 flexDirection: "column",
//                 alignItems: "center",
//                 ...style,
//             }}
//             className={`column-box ${className ? className : ""}`}
//         />
//     );
// };

export const RowBox = ({
    children,
    style,
    onClick,
    onBlur,
}: {
    children
    style?: MantineStyleProp
    onClick?
    onBlur?
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
    )
}

export const TransferCard = ({
    action,
    destination,
    boundRef,
}: {
    action: string
    destination: string
    boundRef?
}) => {
    let width
    let left
    if (boundRef) {
        width = boundRef.clientWidth
        left = boundRef.getBoundingClientRect()['left']
    }
    if (!destination) {
        return
    }

    return (
        <div
            className="transfer-info-box"
            style={{
                width: width ? width : '100%',
                left: left ? left : 0,
            }}
        >
            <Card style={{ height: 'max-content' }}>
                <div className="flex flex-row w-full">
                    <Text style={{ userSelect: 'none' }}>{action} to</Text>
                    <IconFolder style={{ marginLeft: '7px' }} />
                    <Text
                        fw={700}
                        style={{ marginLeft: 3, userSelect: 'none' }}
                    >
                        {destination}
                    </Text>
                </div>
            </Card>
        </div>
    )
}

export const DropSpot = ({
    onDrop,
    dropSpotTitle,
    dragging,
    dropAllowed,
    handleDrag,
    wrapperRef,
}: {
    onDrop
    dropSpotTitle: string
    dragging: DraggingState
    dropAllowed
    handleDrag: React.DragEventHandler<HTMLDivElement>
    wrapperRef?
}) => {
    const wrapperSize = useResize(wrapperRef)
    return (
        <div
            draggable={false}
            className="dropspot-wrapper"
            onDragOver={(e) => {
                if (dragging === 0) {
                    handleDrag(e)
                }
            }}
            style={{
                pointerEvents: dragging === 2 ? 'all' : 'none',
                cursor: !dropAllowed && dragging === 2 ? 'no-drop' : 'auto',
                height: wrapperSize ? wrapperSize.height - 2 : '100%',
                width: wrapperSize ? wrapperSize.width - 2 : '100%',
            }}
            onDragLeave={handleDrag}
        >
            {dragging === 2 && (
                <div
                    className="dropbox"
                    onMouseLeave={handleDrag}
                    onDrop={(e) => {
                        e.preventDefault()
                        e.stopPropagation()
                        if (dropAllowed) {
                            onDrop(e)
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
                        <div
                            style={{
                                position: 'relative',
                                justifyContent: 'center',
                                cursor: 'no-drop',
                                width: 'max-content',
                                pointerEvents: 'none',
                            }}
                        >
                            <IconFolderCancel size={100} color="#dd2222" />
                        </div>
                    )}
                    {dropAllowed && (
                        <TransferCard
                            action="Upload"
                            destination={dropSpotTitle}
                        />
                    )}
                </div>
            )}
        </div>
    )
}

type DirViewWrapperProps = {
    folderName: string
    dragging: number
    children: JSX.Element
}

export const DirViewWrapper = memo(
    ({ folderName, dragging, children }: DirViewWrapperProps) => {
        const { usr }: UserContextT = useContext(UserContext)
        const { fbState, fbDispatch } = useContext(FbContext)
        const [menuOpen, setMenuOpen] = useState(false)
        const [menuPos, setMenuPos] = useState({ x: 0, y: 0 })

        return (
            <div
                draggable={false}
                className="h-full shrink-0 min-w-[400px] grow w-0"
                onDrag={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                }}
                onMouseUp={(e) => {
                    if (dragging) {
                        setTimeout(
                            () =>
                                fbDispatch({
                                    type: 'set_dragging',
                                    dragging: DraggingState.NoDrag,
                                }),
                            10
                        )
                    }
                }}
                onClick={(e) => {
                    if (dragging) {
                        return
                    }
                    fbDispatch({ type: 'clear_selected' })
                }}
                onContextMenu={(e) => {
                    e.preventDefault()
                    if (fbState.fbMode === 'share') {
                        return
                    }
                    setMenuPos({ x: e.clientX, y: e.clientY })
                    setMenuOpen(true)
                }}
            >
                <BackdropMenu
                    folderName={
                        folderName === usr.username ? 'Home' : folderName
                    }
                    menuPos={menuPos}
                    menuOpen={menuOpen}
                    setMenuOpen={setMenuOpen}
                    newFolder={() => fbDispatch({ type: 'new_dir' })}
                />

                <div
                    className="w-full h-full p-2"
                    onDragOver={(event) => {
                        if (!dragging) {
                            handleDragOver(event, fbDispatch, dragging)
                        }
                    }}
                >
                    {children}
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.dragging !== next.dragging) {
            return false
        } else if (prev.folderName !== next.folderName) {
            return false
        } else if (prev.children !== next.children) {
            return false
        }

        return true
    }
)

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
    )
}

export const FileIcon = ({
    fileName,
    id,
    Icon,
    usr,
    as,
    includeText = true,
}: {
    fileName: string
    id: string
    Icon
    usr: UserInfoT
    as?: string
    includeText?: boolean
}) => {
    return (
        <Box
            style={{
                display: 'flex',
                alignItems: 'center',
            }}
        >
            <Icon className="icon-noshrink" />
            <Text
                fw={550}
                c="white"
                truncate="end"
                style={{
                    fontFamily: 'monospace',
                    textWrap: 'nowrap',
                    padding: 6,
                    flexShrink: 1,
                }}
            >
                {friendlyFolderName(fileName, id, usr)}
            </Text>
            {as && (
                <Box
                    style={{
                        display: 'flex',
                        flexDirection: 'row',
                        alignItems: 'center',
                    }}
                >
                    <Text size="12px">as</Text>
                    <Text
                        size="12px"
                        truncate="end"
                        style={{
                            fontFamily: 'monospace',
                            textWrap: 'nowrap',
                            padding: 3,
                            flexShrink: 2,
                        }}
                    >
                        {as}
                    </Text>
                </Box>
            )}
        </Box>
    )
}

export const FolderIcon = ({ shares, size }: { shares; size }) => {
    const [copied, setCopied] = useState(false)
    const wormholeId = useMemo(() => {
        if (shares) {
            const whs = shares.filter((s) => s.Wormhole)
            if (whs.length !== 0) {
                return whs[0].shareId
            }
        }
    }, [shares])
    return (
        <Box
            style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                width: '100%',
                height: '100%',
            }}
        >
            <IconFolder size={size} />
            {wormholeId && (
                <Tooltip label={copied ? 'Copied' : 'Copy Wormhole'}>
                    <IconSpiral
                        color={copied ? '#4444ff' : 'white'}
                        style={{ position: 'absolute', right: 0, top: 0 }}
                        onClick={(e) => {
                            e.stopPropagation()
                            navigator.clipboard.writeText(
                                `${window.location.origin}/wormhole/${shares[0].ShareId}`
                            )
                            setCopied(true)
                            setTimeout(() => setCopied(false), 1000)
                        }}
                        // onDoubleClick={(e) => e.stopPropagation()}
                    />
                </Tooltip>
            )}
        </Box>
    )
}

export const IconDisplay = ({
    file,
    size = 24,
    allowMedia = false,
}: {
    file: WeblensFile
    size?: string | number
    allowMedia?: boolean
}) => {
    const [containerRef, setContainerRef] = useState(null)
    const containerSize = useResize(containerRef)

    if (!file) {
        return null
    }

    if (file.IsFolder()) {
        return <FolderIcon shares={file.GetShares()} size={size} />
    }

    if (!file.IsImported() && file.GetMedia().IsDisplayable() && allowMedia) {
        return (
            <Center style={{ height: '100%', width: '100%' }}>
                {/* <Skeleton height={"100%"} width={"100%"} />
                <Text pos={"absolute"} style={{ userSelect: "none" }}>
                    Processing...
                </Text> */}
                <IconPhoto />
            </Center>
        )
    } else if (file.GetMedia().IsDisplayable() && allowMedia) {
        return (
            <Box ref={setContainerRef} style={{ justifyContent: 'center' }}>
                <ContainerMedia
                    mediaData={file.GetMedia()}
                    containerRef={containerRef}
                />
            </Box>
            // <MediaImage media={file.mediaData} quality={quality} />
        )
    } else if (file.GetMedia().IsDisplayable()) {
        return <IconPhoto />
    }
    const extIndex = file.GetFilename().lastIndexOf('.')
    const ext = file
        .GetFilename()
        .slice(extIndex + 1, file.GetFilename().length)
    const textSize = `${Math.floor(containerSize?.width / (ext.length + 5))}px`

    switch (ext) {
        case 'zip':
            return <IconFileZip />
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
            )
    }
}

export const FileInfoDisplay = ({ file }: { file: WeblensFile }) => {
    let [size, units] = humanFileSize(file.GetSize())
    return (
        <Box
            style={{
                width: 'max-content',
                whiteSpace: 'nowrap',
                justifyContent: 'center',
                maxWidth: '100%',
            }}
        >
            <Text fw={600} style={{ fontSize: '2.5vw', maxWidth: '100%' }}>
                {file.GetFilename()}
            </Text>
            {file.IsFolder() && (
                <RowBox
                    style={{
                        height: 'max-content',
                        justifyContent: 'center',
                        width: '100%',
                    }}
                >
                    <Text style={{ fontSize: '25px', maxWidth: '100%' }}>
                        {file.GetChildren().length} Item
                        {file.GetChildren().length !== 1 ? 's' : ''}
                    </Text>
                    <Divider orientation="vertical" size={2} mx={10} />
                    <Text style={{ fontSize: '25px' }}>
                        {size}
                        {units}
                    </Text>
                </RowBox>
            )}
            {!file.IsFolder() && (
                <Text style={{ fontSize: '25px' }}>
                    {size}
                    {units}
                </Text>
            )}
        </Box>
    )
}

export const PresentationFile = ({ file }: { file: WeblensFile }) => {
    if (!file) {
        return null
    }
    let [size, units] = humanFileSize(file.GetSize())
    if (file.GetMedia() && file.GetMedia().IsDisplayable()) {
        return (
            <Box
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
                    {file.GetFilename()}
                </Text>
                <Text style={{ fontSize: '25px' }}>
                    {size}
                    {units}
                </Text>
                <Text style={{ fontSize: '20px' }}>
                    {file.GetModified().toLocaleDateString('en-us', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                    })}
                </Text>
                <Divider />
                <Text style={{ fontSize: '20px' }}>
                    {new Date(
                        Date.parse(file.GetMedia().GetCreateDate())
                    ).toLocaleDateString('en-us', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                    })}
                </Text>
            </Box>
        )
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
                <Box style={{ width: '40%', justifyContent: 'center' }}>
                    <Text fw={600} style={{ width: '100%' }}>
                        {file.GetFilename()}
                    </Text>
                    {file.IsFolder() && (
                        <RowBox
                            style={{
                                height: 'max-content',
                                justifyContent: 'center',
                                width: '50vw',
                            }}
                        >
                            <Text style={{ fontSize: '25px' }}>
                                {file.GetChildren().length} Item
                                {file.GetChildren().length !== 1 ? 's' : ''}
                            </Text>
                            <Divider orientation="vertical" size={2} mx={10} />
                            <Text style={{ fontSize: '25px' }}>
                                {size}
                                {units}
                            </Text>
                        </RowBox>
                    )}
                    {!file.IsFolder() && (
                        <Text style={{ fontSize: '25px' }}>
                            {size}
                            {units}
                        </Text>
                    )}
                </Box>
            </RowBox>
        )
    }
}

const EmptyIcon = ({ folderId, usr }) => {
    if (folderId === usr.homeId) {
        return <IconHome size={500} color="#16181d" />
    }
    if (folderId === usr.trashId) {
        return <IconTrash size={500} color="#16181d" />
    }
    if (folderId === 'shared') {
        return <IconUsers size={500} color="#16181d" />
    }
    if (folderId === 'EXTERNAL') {
        return <IconServer size={500} color="#16181d" />
    }
    return null
}

export const GetStartedCard = ({
    fb,
    dispatch,
    uploadDispatch,
    wsSend,
}: {
    fb: FbStateT
    dispatch: FBDispatchT
    uploadDispatch
    wsSend
}) => {
    const { authHeader, usr } = useContext(UserContext)
    return (
        <Box className="flex w-full justify-center items-center">
            <Box className="flex flex-col w-max h-fit mt-[25vh] justify-center items-center">
                <Box className="flex items-center p-30 absolute -z-1 pointer-events-none h-max">
                    <EmptyIcon folderId={fb.folderInfo.Id()} usr={usr} />
                </Box>

                <p className="text-2xl w-max h-max select-none z-10">
                    {`This folder ${
                        fb.folderInfo.IsPastFile() ? 'was' : 'is'
                    } empty`}
                </p>

                {fb.folderInfo.IsModifiable() && !fb.viewingPast && (
                    <Box className="flex flex-row p-5 w-350 z-10">
                        <FileButton
                            onChange={(files) => {
                                HandleUploadButton(
                                    files,
                                    fb.folderInfo.Id(),
                                    false,
                                    '',
                                    authHeader,
                                    uploadDispatch,
                                    wsSend
                                )
                            }}
                            accept="file"
                            multiple
                        >
                            {(props) => {
                                return (
                                    <Box
                                        className="flex flex-col items-center w-32 text-gray-400 cursor-pointer m-5 font-normal stroke-1 transition-all duration-100 hover:text-white hover:font-semibold hover:stroke-2 z-10"
                                        onClick={() => {
                                            props.onClick()
                                        }}
                                    >
                                        <IconUpload
                                            size={100}
                                            stroke={'inherit'}
                                            style={{ padding: '10px' }}
                                        />
                                        <Text size="20px" fw={'inherit'}>
                                            Upload
                                        </Text>
                                        <Space h={4}></Space>
                                        <Text size="12px" fw={'inherit'}>
                                            Click or Drop
                                        </Text>
                                    </Box>
                                )
                            }}
                        </FileButton>
                        <Divider orientation="vertical" m={30} />

                        <Box
                            className="flex flex-col items-center w-32 text-gray-400 cursor-pointer m-5 font-normal stroke-1 transition-all duration-100 hover:text-white hover:font-semibold hover:stroke-2 z-10"
                            onClick={(e) => {
                                e.stopPropagation()
                                dispatch({ type: 'new_dir' })
                            }}
                        >
                            <IconFolderPlus
                                size={100}
                                stroke={'inherit'}
                                style={{ padding: '10px' }}
                            />
                            <Text
                                size="20px"
                                fw={'inherit'}
                                style={{ width: 'max-content' }}
                            >
                                New Folder
                            </Text>
                        </Box>
                    </Box>
                )}
            </Box>
        </Box>
    )
}

export const WebsocketStatus = memo(
    ({ ready }: { ready: number }) => {
        let color
        let status

        switch (ready) {
            case 1:
                color = '#00ff0055'
                status = 'Connected'
                break
            case 2:
            case 3:
                color = 'orange'
                status = 'Connecting'
                break
            case -1:
                color = 'red'
                status = 'Disconnected, try refreshing your page'
        }

        return (
            <div className="absolute bottom-1 left-1">
                <Tooltip label={status} color="#222222">
                    <svg width="24" height="24" fill={color}>
                        <path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" />
                    </svg>
                </Tooltip>
            </div>
        )
    },
    (prev, next) => {
        return prev.ready === next.ready
    }
)
