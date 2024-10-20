import { Divider, FileButton, Space, Text, Tooltip } from '@mantine/core'
import { useMouse } from '@mantine/hooks'

import {
    IconChevronRight,
    IconFile,
    IconFileZip,
    IconFolder,
    IconFolderCancel,
    IconFolderPlus,
    IconHome,
    IconPhoto,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import { useResize } from '@weblens/components/hooks'

import '@weblens/components/theme.scss'
import './style/fileBrowserStyle.scss'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { FbMenuModeT, WeblensFile } from '@weblens/types/files/File'
import { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import { UserInfoT } from '@weblens/types/Types'
import { friendlyFolderName, humanFileSize } from '@weblens/util'
import { DragEventHandler, FC, memo, useMemo, useState } from 'react'
import { FbModeT, useFileBrowserStore } from './FBStateControl'
import { handleDragOver, HandleUploadButton } from './FileBrowserLogic'
import '@weblens/components/theme.scss'
import WeblensTooltip from '@weblens/lib/WeblensTooltip'

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
                <FileFmt pathName={destination} />
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
    stopDragging,
}: {
    onDrop
    dropSpotTitle: string
    dragging: DraggingStateT
    dropAllowed
    handleDrag: DragEventHandler<HTMLDivElement>
    wrapperRef?
    stopDragging: () => void
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
                    onMouseLeave={() => {
                        stopDragging()
                    }}
                    onDrop={(e) => {
                        e.preventDefault()
                        e.stopPropagation()
                        if (dropAllowed) {
                            onDrop(e)
                        } else {
                            stopDragging()
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

export function DraggingCounter() {
    const drag = useFileBrowserStore((state) => state.draggingState)
    const setDrag = useFileBrowserStore((state) => state.setDragging)
    const selected = useFileBrowserStore((state) => state.selected)
    const filesMap = useFileBrowserStore((state) => state.filesMap)

    const position = useMouse()
    const selectedKeys = Array.from(selected.keys())
    const { files, folders } = useMemo(() => {
        let files = 0
        let folders = 0

        selectedKeys.forEach((k: string) => {
            if (filesMap.get(k)?.IsFolder()) {
                folders++
            } else {
                files++
            }
        })
        return { files, folders }
    }, [JSON.stringify(selectedKeys)])

    if (drag !== DraggingStateT.InternalDrag) {
        return null
    }

    return (
        <div
            className="fixed z-10 bg-wl-barely-visible wl-outline p-2"
            style={{
                top: position.y + 8,
                left: position.x + 8,
            }}
            onMouseUp={() => {
                setDrag(DraggingStateT.NoDrag)
            }}
        >
            {Boolean(files) && (
                <div className="flex flex-row h-max text-[--wl-text-color]">
                    <IconFile size={30} />
                    <Space w={10} />
                    <p>{files}</p>
                </div>
            )}
            {Boolean(folders) && (
                <div className="flex flex-row h-max text-[--wl-text-color]">
                    <IconFolder size={30} />
                    <Space w={10} />
                    <p>{folders}</p>
                </div>
            )}
        </div>
    )
}

export const DirViewWrapper = memo(
    ({ children }: { children }) => {
        const draggingState = useFileBrowserStore(
            (state) => state.draggingState
        )

        const setDragging = useFileBrowserStore((state) => state.setDragging)
        const clearSelected = useFileBrowserStore(
            (state) => state.clearSelected
        )
        const setMenu = useFileBrowserStore((state) => state.setMenu)

        return (
            <div
                draggable={false}
                className="h-full shrink-0 min-w-[400px] grow w-0"
                onDrag={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                }}
                onMouseUp={() => {
                    if (draggingState) {
                        setTimeout(() => setDragging(DraggingStateT.NoDrag), 10)
                    } else {
                        clearSelected()
                    }
                }}
                // onClick={() => {
                //     if (draggingState) {
                //         return
                //     }
                //     clearSelected()
                // }}
                onContextMenu={(e) => {
                    e.preventDefault()
                    setMenu({
                        menuTarget: '',
                        menuPos: { x: e.clientX, y: e.clientY },
                        menuState: FbMenuModeT.Default,
                    })
                }}
            >
                <div
                    className="w-full h-full p-2"
                    onDragOver={(event) => {
                        if (!draggingState) {
                            handleDragOver(event, setDragging, draggingState)
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
}: {
    fileName: string
    id: string
    Icon
    usr: UserInfoT
    as?: string
}) => {
    return (
        <div className="flex items-center">
            <Icon className="icon-noshrink" />
            <p className="font-medium text-white truncate text-nowrap p-2 shrink">
                {friendlyFolderName(fileName, id, usr)}
            </p>
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

export const IconDisplay = ({
    file,
    allowMedia = false,
}: {
    file: WeblensFile
    allowMedia?: boolean
}) => {
    const [containerRef, setContainerRef] = useState(null)
    const containerSize = useResize(containerRef)
    const mediaData = useMediaStore((state) =>
        state.mediaMap.get(file.GetContentId())
    )

    if (!file) {
        return null
    }

    if (file.IsFolder()) {
        if (file.GetContentId() !== '') {
            return (
                <div className="relative flex w-full h-full justify-center items-center">
                    {/* <IconFolder */}
                    {/*     stroke={0} */}
                    {/*     fill={'white'} */}
                    {/*     className="absolute h-1/4 w-1/4 z-10 shrink-0 text-[--wl-text-color] bottom-[12%] right-[3%]" */}
                    {/* /> */}
                    <div className="relative w-[90%] h-[90%] -translate-x-1 -translate-y-1 z-10">
                        <MediaImage
                            media={mediaData}
                            quality={PhotoQuality.LowRes}
                        />
                    </div>
                    <div className="absolute w-[88%] h-[88%] translate-x-[2px] translate-y-[2px] outline outline-2 outline-theme-text opacity-75 rounded" />
                    <div className="absolute w-[88%] h-[88%] translate-x-[8px] translate-y-[8px] outline outline-2 outline-theme-text opacity-50 rounded" />
                </div>
            )
        } else {
            return (
                <IconFolder
                    stroke={1}
                    className="h-3/4 w-3/4 z-10 shrink-0 text-[--wl-text-color]"
                />
            )
        }
    }

    if (mediaData && (!mediaData.IsImported() || !allowMedia)) {
        return <IconPhoto stroke={1} className="shrink-0" />
    } else if (mediaData && allowMedia && mediaData.IsImported()) {
        return <MediaImage media={mediaData} quality={PhotoQuality.LowRes} />
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
                <div
                    ref={setContainerRef}
                    className="flex justify-center items-center w-full h-full"
                >
                    <IconFile stroke={1} className="w-3/4 h-3/4" />
                    {extIndex !== -1 && (
                        <p
                            className="font-semibold absolute select-none"
                            style={{ fontSize: textSize }}
                        >
                            .{ext}
                        </p>
                    )}
                </div>
            )
    }
}

export const FileInfoDisplay = ({ file }: { file: WeblensFile }) => {
    const [size, units] = humanFileSize(file.GetSize())
    return (
        <div className="flex flex-col w-max whitespace-nowrap justify-center max-w-full ml-1 gap-1 mb-2">
            <p className="text-3xl font-semibold max-w-full">
                {file.GetFilename()}
            </p>
            {file.IsFolder() && (
                <div className="flex flex-row h-max w-full items-center justify-center">
                    <p className="text-sm max-w-full">
                        {file.GetChildren().length} Item
                        {file.GetChildren().length !== 1 ? 's' : ''}
                    </p>
                    <Divider orientation="vertical" size={2} mx={10} />
                    <Text style={{ fontSize: '25px' }}>
                        {size}
                        {units}
                    </Text>
                </div>
            )}
            {!file.IsFolder() && (
                <p className={'text-sm'}>
                    {size}
                    {units}
                </p>
            )}
        </div>
    )
}

const EmptyIcon = ({ folderId, usr }) => {
    if (folderId === usr.homeId) {
        return <IconHome size={500} className="text-wl-barely-visible" />
    }
    if (folderId === usr.trashId) {
        return <IconTrash size={500} className="text-wl-barely-visible" />
    }
    if (folderId === 'shared') {
        return <IconUsers size={500} className="text-wl-barely-visible" />
    }
    if (folderId === 'EXTERNAL') {
        return <IconServer size={500} className="text-wl-barely-visible" />
    }
    return <IconFolder size={500} className="text-wl-barely-visible" />
}

export const GetStartedCard = () => {
    const user = useSessionStore((state) => state.user)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const viewingPast = useFileBrowserStore((state) => state.pastTime)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    if (!folderInfo && mode === FbModeT.share) {
        return (
            <div className="flex justify-center items-center h-3/4">
                <div className="h-max w-max p-4 wl-outline">
                    <h4>You have no files shared with you</h4>
                </div>
            </div>
        )
    } else if (!folderInfo) {
        return null
    }

    return (
        <div className="flex w-full justify-center items-center animate-fade h-3/4">
            <div className="flex flex-col w-max h-fit justify-center items-center">
                <div className="flex items-center p-30 absolute -z-1 pointer-events-none h-max">
                    <EmptyIcon folderId={folderInfo.Id()} usr={user} />
                </div>

                <p className="text-2xl w-max h-max select-none z-10">
                    {`This folder ${folderInfo.IsPastFile() ? 'was' : 'is'} empty`}
                </p>

                {folderInfo.IsModifiable() && !viewingPast && (
                    <div className="flex flex-row p-5 w-350 z-10 items-center">
                        <FileButton
                            onChange={(files) => {
                                HandleUploadButton(
                                    files,
                                    folderInfo.Id(),
                                    false,
                                    ''
                                )
                            }}
                            accept="file"
                            multiple
                        >
                            {(props) => {
                                return (
                                    <WeblensButton
                                        subtle
                                        Left={IconUpload}
                                        squareSize={128}
                                        tooltip="Upload"
                                        onClick={props.onClick}
                                    />
                                )
                            }}
                        </FileButton>
                        <Divider orientation="vertical" m={30} />

                        <WeblensButton
                            Left={IconFolderPlus}
                            squareSize={128}
                            subtle
                            tooltip="New Folder"
                            onClick={(e) => {
                                setMenu({
                                    menuPos: { x: e.clientX, y: e.clientY },
                                    menuState: FbMenuModeT.NameFolder,
                                })
                            }}
                        />
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
                status = 'Disconnected'
        }

        return (
            <WeblensTooltip label={status}>
                <svg width="24" height="24" fill={color}>
                    <path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" />
                </svg>
            </WeblensTooltip>
        )
    },
    (prev, next) => {
        return prev.ready === next.ready
    }
)

export function PathFmt({ pathName }: { pathName: string }) {
    pathName = pathName.slice(pathName.indexOf(':') + 1)
    const parts = pathName.split('/')

    if (parts[parts.length - 1] === '') {
        parts.pop()
    }

    let StartIcon: FC<{ className: string }>
    if (parts[0] === '.user_trash') {
        parts.shift()
        StartIcon = IconTrash
    } else {
        parts.shift()
        StartIcon = IconHome
    }

    return (
        <div
            className="flex items-center min-w-0"
            style={{ flexShrink: parts.length ? 1 : 0 }}
        >
            <StartIcon className="shrink-0 text-[--wl-text-color]" />
            {parts.map((part) => {
                return (
                    <div
                        key={part}
                        className="flex w-max items-center shrink min-w-0"
                    >
                        <IconChevronRight
                            className="text-[--wl-text-color] shrink-0"
                            size={18}
                        />
                        <p className="text-[--wl-text-color] select-none font-semibold text-xl text-nowrap truncate">
                            {part}
                        </p>
                    </div>
                )
            })}
        </div>
    )
}

export function FileFmt({ pathName }: { pathName: string }) {
    if (!pathName) {
        return null
    }

    pathName = pathName.slice(pathName.indexOf(':') + 1)
    const parts = pathName.split('/')
    parts.shift()

    let StartIcon: FC<{ className: string }>
    let nameText: string
    if (parts.length == 0 || (parts.length === 1 && parts[0] === '')) {
        StartIcon = IconHome
        nameText = 'Home'
    } else if (
        parts.length === 2 &&
        parts[0] === '.user_trash' &&
        parts[1] === ''
    ) {
        parts.shift()
        StartIcon = IconTrash
        nameText = 'Trash'
    } else if (parts[parts.length - 1] === '') {
        parts.pop()
        StartIcon = IconFolder
        nameText = parts[parts.length - 1]
    } else {
        StartIcon = IconFile
        nameText = parts[parts.length - 1]
    }

    return (
        <div
            className="flex items-center w-max min-w-0"
            style={{ flexShrink: parts.length ? 1 : 0 }}
        >
            <StartIcon className="theme-text m-1 shrink-0 text-[--wl-text-color]" />
            <p className="theme-text select-none font-semibold text-lg text-nowrap truncate">
                {nameText}
            </p>
        </div>
    )
}
