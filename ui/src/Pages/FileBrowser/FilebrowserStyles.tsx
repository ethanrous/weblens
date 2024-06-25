import {
    Box,
    Center,
    Divider,
    FileButton,
    Space,
    Text,
    Tooltip,
} from '@mantine/core'

import {
    IconFile,
    IconFileZip,
    IconFolder,
    IconFolderCancel,
    IconFolderPlus,
    IconHome,
    IconPhoto,
    IconServer,
    IconSpiral,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import {
    DragEventHandler,
    memo,
    useContext,
    useEffect,
    useMemo,
    useState,
} from 'react'
import { useMedia, useResize } from '../../components/hooks'

import './style/fileBrowserStyle.scss'
import { ContainerMedia } from '../../components/Presentation'
import { UserContext } from '../../Context'
import { FbMenuModeT, WeblensFile } from '../../Files/File'
import { DraggingStateT, FbContext } from '../../Files/filesContext'
import { FBDispatchT, FbStateT, UserInfoT } from '../../types/Types'
import { friendlyFolderName, humanFileSize } from '../../util'
import { handleDragOver, HandleUploadButton } from './FileBrowserLogic'

export const TransferCard = ({
    action,
    destination,
    boundRef,
}: {
    action: string
    destination: string
    boundRef?
}) => {
    let width: number
    let left: number
    if (boundRef) {
        width = boundRef.clientWidth
        left = boundRef.getBoundingClientRect()['left']
    }
    if (!destination) {
        return null
    }

    return (
        <div
            className="transfer-info-wrapper"
            style={{
                width: width ? width : '100%',
                left: left ? left : 0,
            }}
        >
            <div className="transfer-info-box">
                <p className="select-none">{action} to</p>
                <IconFolder />
                <p className="font-bold select-none">{destination}</p>
            </div>
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
    dragging: DraggingStateT
    dropAllowed
    handleDrag: DragEventHandler<HTMLDivElement>
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
                    // required for onDrop to work
                    // https://stackoverflow.com/questions/50230048/react-ondrop-is-not-firing
                    onDragOver={(e) => e.preventDefault()}
                    style={{
                        outlineColor: `${dropAllowed ? '#ffffff' : '#dd2222'}`,
                        cursor:
                            !dropAllowed && dragging === 2 ? 'no-drop' : 'auto',
                    }}
                >
                    {!dropAllowed && (
                        <div className="flex justify-center items-center relative cursor-no-drop w-max pointer-events-none">
                            <IconFolderCancel
                                className="pointer-events-none"
                                size={100}
                                color="#dd2222"
                            />
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

export const DirViewWrapper = memo(
    ({ children }: { children }) => {
        const { fbState, fbDispatch } = useContext(FbContext)
        const [menuMode, setMenuMode] = useState(FbMenuModeT.Closed)

        useEffect(() => {
            if (menuMode === FbMenuModeT.Sharing) {
                fbDispatch({ type: 'set_block_focus', block: true })
            } else {
                fbDispatch({ type: 'set_block_focus', block: false })
            }
        }, [menuMode])

        return (
            <div
                draggable={false}
                className="h-full shrink-0 min-w-[400px] grow w-0"
                onDrag={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                }}
                onMouseUp={() => {
                    if (fbState.draggingState) {
                        setTimeout(
                            () =>
                                fbDispatch({
                                    type: 'set_dragging',
                                    dragging: DraggingStateT.NoDrag,
                                }),
                            10
                        )
                    }
                }}
                onClick={() => {
                    if (fbState.draggingState) {
                        return
                    }
                    fbDispatch({ type: 'clear_selected' })
                }}
                onContextMenu={(e) => {
                    e.preventDefault()
                    fbDispatch({ type: 'set_menu_target', fileId: '' })
                    fbDispatch({
                        type: 'set_menu_pos',
                        pos: { x: e.clientX, y: e.clientY },
                    })
                    fbDispatch({
                        type: 'set_menu_open',
                        menuMode: FbMenuModeT.Default,
                    })
                }}
            >
                <div
                    className="w-full h-full p-2"
                    onDragOver={(event) => {
                        if (!fbState.draggingState) {
                            handleDragOver(
                                event,
                                fbDispatch,
                                fbState.draggingState
                            )
                        }
                    }}
                >
                    {children}
                </div>
            </div>
        )
    },
    (prev, next) => {
        return prev.children === next.children
    }
)

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
        <div className="flex items-center">
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
                <div className="flex flex-row items-center">
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
                </div>
            )}
        </div>
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
        <div className="flex w-full h-full items-center justify-center">
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
        </div>
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
    const mediaData = useMedia(file.GetMediaId())

    if (!file) {
        return null
    }

    if (file.IsFolder()) {
        return <FolderIcon shares={file.GetShare()} size={size} />
    }

    if (!file.IsImported() && mediaData && allowMedia) {
        return (
            <Center style={{ height: '100%', width: '100%' }}>
                {/* <Skeleton height={"100%"} width={"100%"} />
                 <Text pos={"absolute"} style={{ userSelect: "none" }}>
                 Processing...
                 </Text> */}
                <IconPhoto />
            </Center>
        )
    } else if (mediaData && allowMedia) {
        return (
            <Box ref={setContainerRef} style={{ justifyContent: 'center' }}>
                <ContainerMedia
                    mediaData={mediaData}
                    containerRef={containerRef}
                />
            </Box>
            // <MediaImage media={file.mediaData} quality={quality} />
        )
    } else if (mediaData) {
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
        <div
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
                <div className="flex flex-row h-max w-full items-center justify-center">
                    <Text style={{ fontSize: '25px', maxWidth: '100%' }}>
                        {file.GetChildren().length} Item
                        {file.GetChildren().length !== 1 ? 's' : ''}
                    </Text>
                    <Divider orientation="vertical" size={2} mx={10} />
                    <Text style={{ fontSize: '25px' }}>
                        {size}
                        {units}
                    </Text>
                </div>
            )}
            {!file.IsFolder() && (
                <Text style={{ fontSize: '25px' }}>
                    {size}
                    {units}
                </Text>
            )}
        </div>
    )
}

export const PresentationFile = ({ file }: { file: WeblensFile }) => {
    const mediaData = useMedia(file.GetMediaId())

    if (!file) {
        return null
    }
    let [size, units] = humanFileSize(file.GetSize())
    if (mediaData) {
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
                        Date.parse(mediaData.GetCreateDate())
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
            <div
                className="flex flex-row h-max w-full items-center justify-center"
                onClick={(e) => e.stopPropagation()}
            >
                <div className="w-flex [60%] justify-center">
                    <IconDisplay file={file} allowMedia />
                </div>
                <Space w={30} />
                <Box style={{ width: '40%', justifyContent: 'center' }}>
                    <Text fw={600} style={{ width: '100%' }}>
                        {file.GetFilename()}
                    </Text>
                    {file.IsFolder() && (
                        <div className="flex flex-row h-max w-1/2 items-center justify-center">
                            <Text style={{ fontSize: '25px' }}>
                                {file.GetChildren().length} Item
                                {file.GetChildren().length !== 1 ? 's' : ''}
                            </Text>
                            <Divider orientation="vertical" size={2} mx={10} />
                            <Text style={{ fontSize: '25px' }}>
                                {size}
                                {units}
                            </Text>
                        </div>
                    )}
                    {!file.IsFolder() && (
                        <Text style={{ fontSize: '25px' }}>
                            {size}
                            {units}
                        </Text>
                    )}
                </Box>
            </div>
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
        <div className="flex w-full justify-center items-center animate-fade">
            <div className="flex flex-col w-max h-fit mt-[25vh] justify-center items-center">
                <div className="flex items-center p-30 absolute -z-1 pointer-events-none h-max">
                    <EmptyIcon folderId={fb.folderInfo.Id()} usr={usr} />
                </div>

                <p className="text-2xl w-max h-max select-none z-10">
                    {`This folder ${
                        fb.folderInfo.IsPastFile() ? 'was' : 'is'
                    } empty`}
                </p>

                {fb.folderInfo.IsModifiable() && !fb.viewingPast && (
                    <div className="flex flex-row p-5 w-350 z-10">
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
                                    <div
                                        className="flex flex-col items-center w-32 text-gray-400 cursor-pointer m-5 font-normal stroke-1 transition-all duration-100 hover:text-white hover:font-semibold hover:stroke-2 z-10"
                                        onClick={() => {
                                            props.onClick()
                                        }}
                                    >
                                        <IconUpload
                                            size={100}
                                            stroke={'inherit'}
                                            className="p-3"
                                        />
                                        <Text
                                            size="20px"
                                            fw={'inherit'}
                                            className="select-none"
                                        >
                                            Upload
                                        </Text>
                                        <Space h={4}></Space>
                                        <Text
                                            size="12px"
                                            fw={'inherit'}
                                            className="select-none"
                                        >
                                            Click or Drop
                                        </Text>
                                    </div>
                                )
                            }}
                        </FileButton>
                        <Divider orientation="vertical" m={30} />

                        <div
                            className="flex flex-col items-center w-32 text-gray-400 cursor-pointer m-5 font-normal stroke-1 transition-all duration-100 hover:text-white hover:font-semibold hover:stroke-2 z-10"
                            onClick={(e) => {
                                e.stopPropagation()
                                dispatch({ type: 'new_dir' })
                            }}
                        >
                            <IconFolderPlus
                                size={100}
                                stroke={'inherit'}
                                className="p-3"
                            />
                            <Text
                                size="20px"
                                fw={'inherit'}
                                className="select-none w-max"
                            >
                                New Folder
                            </Text>
                        </div>
                    </div>
                )}
            </div>
        </div>
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
