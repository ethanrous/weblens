import { Divider, FileButton, Text } from '@mantine/core'
import {
    IconFile,
    IconFileZip,
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconPhoto,
    IconServer,
    IconSlash,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import WeblensLoader from '@weblens/components/Loading'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useResize } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensTooltip from '@weblens/lib/WeblensTooltip'
import { ButtonIcon } from '@weblens/lib/buttonTypes'
import historyStyle from '@weblens/pages/FileBrowser/style/historyStyle.module.scss'
import { ErrorHandler } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { FbMenuModeT, WeblensFile } from '@weblens/types/files/File'
import { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import User from '@weblens/types/user/User'
import { friendlyFolderName, humanFileSize } from '@weblens/util'
import { FC, ReactElement, memo, useState } from 'react'

import { FbModeT, useFileBrowserStore } from '../../store/FBStateControl'
import { HandleUploadButton, filenameFromPath } from './FileBrowserLogic'
import fbStyle from './style/fileBrowserStyle.module.scss'

export const DirViewWrapper = memo(
    ({ children }: { children: ReactElement }) => {
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
                className="h-full shrink-0 grow w-0"
                onDrag={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                }}
                onMouseUp={(e) => {
                    if (draggingState) {
                        e.stopPropagation()
                        setTimeout(() => setDragging(DraggingStateT.NoDrag), 10)
                    }
                }}
                onClick={(e) => {
                    e.stopPropagation()
                    if (!draggingState) {
                        clearSelected()
                    }
                }}
                onContextMenu={(e) => {
                    e.preventDefault()
                    setMenu({
                        menuTarget: '',
                        menuPos: { x: e.clientX, y: e.clientY },
                        menuState: FbMenuModeT.Default,
                    })
                }}
            >
                <div className="w-full h-full p-2">{children}</div>
            </div>
        )
    },
    (prev, next) => {
        return prev.children === next.children
    }
)

export const FileIcon = ({
    filename,
    id,
    Icon,
    usr,
    as,
}: {
    filename: string
    id: string
    Icon: ButtonIcon
    usr: User
    as?: string
}) => {
    return (
        <div className="flex items-center">
            <Icon className={fbStyle['icon-noshrink']} />
            <p className="font-medium text-white truncate text-nowrap p-2 shrink">
                {friendlyFolderName(filename, id, usr)}
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
    const [containerRef, setContainerRef] = useState<HTMLDivElement>(null)
    const containerSize = useResize(containerRef)
    const mediaData = useMediaStore((state) =>
        state.mediaMap.get(file.GetContentId())
    )

    if (!file) {
        return null
    }

    if (file.IsFolder()) {
        if (file.GetContentId() !== '') {
            const containerQuanta = Math.ceil(containerSize.height / 100)
            return (
                <div
                    ref={setContainerRef}
                    className="relative flex w-full h-full justify-center items-center "
                >
                    <div
                        className="relative w-[90%] h-[90%] z-20"
                        style={{
                            translate: `${containerQuanta * -3}px ${containerQuanta * -3}px`,
                        }}
                    >
                        <MediaImage
                            media={mediaData}
                            quality={PhotoQuality.LowRes}
                        />
                    </div>
                    <div className="absolute w-[88%] h-[88%] bg-wl-outline-subtle outline outline-2 outline-theme-text opacity-75 rounded z-10" />
                    <div
                        className="absolute w-[88%] h-[88%] bg-wl-outline-subtle outline outline-2 outline-theme-text opacity-50 rounded"
                        style={{
                            translate: `${containerQuanta * 3}px ${containerQuanta * 3}px`,
                        }}
                    />
                </div>
            )
        } else {
            return (
                <IconFolder
                    stroke={1}
                    className="h-3/4 w-3/4 z-10 shrink-0 text-[--wl-file-text-color]"
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

const EmptyIcon = ({ folderId, usr }: { folderId: string; usr: User }) => {
    if (folderId === usr.homeId) {
        return <IconHome size={500} className="text-wl-barely-visible " />
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

export function GetStartedCard() {
    const user = useSessionStore((state) => state.user)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const loading = useFileBrowserStore((state) => state.loading)
    const [viewRef, setViewRef] = useState<HTMLDivElement>()
    const size = useResize(viewRef)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    if (draggingState === DraggingStateT.ExternalDrag) {
        return null
    }

    if (loading && loading.length !== 0) {
        return <WeblensLoader />
    }

    if (!folderInfo && mode === FbModeT.share) {
        return (
            <div className="flex justify-center items-center m-auto p-2">
                <div className="h-max w-max p-4 wl-outline">
                    <h4>You have no files shared with you</h4>
                </div>
            </div>
        )
    } else if (!folderInfo) {
        return null
    }

    const doVertical = size.width < 500

    let buttonSize = 128
    if (doVertical) {
        buttonSize = 64
    }

    return (
        <div
            ref={setViewRef}
            className="flex w-[50vw] min-w-[250px] justify-center items-center animate-fade h-max m-auto p-4 z-[3]"
        >
            <div className="flex flex-col w-max max-w-full h-fit justify-center items-center">
                <div className="flex items-center p-30 absolute -z-1 pointer-events-none h-max max-w-full min-w-[200px]">
                    <EmptyIcon folderId={folderInfo.Id()} usr={user} />
                </div>

                <p className="text-2xl w-max h-max select-none z-10">
                    {`This folder ${folderInfo.IsPastFile() ? 'was' : 'is'} empty`}
                </p>

                {folderInfo.IsModifiable() && (
                    <div
                        className="flex p-5 w-350 max-w-full z-10 items-center"
                        style={{ flexDirection: doVertical ? 'column' : 'row' }}
                    >
                        <FileButton
                            onChange={(files) => {
                                HandleUploadButton(
                                    files,
                                    folderInfo.Id(),
                                    false,
                                    ''
                                ).catch(ErrorHandler)
                            }}
                            accept="file"
                            multiple
                        >
                            {(props) => {
                                return (
                                    <WeblensButton
                                        subtle
                                        Left={IconUpload}
                                        squareSize={buttonSize}
                                        tooltip="Upload"
                                        onClick={props.onClick}
                                    />
                                )
                            }}
                        </FileButton>
                        <Divider
                            orientation={doVertical ? 'horizontal' : 'vertical'}
                            m={30 * (buttonSize / 128)}
                        />

                        <WeblensButton
                            Left={IconFolderPlus}
                            squareSize={buttonSize}
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
        let color: string
        let status: string

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
    while (parts.includes('.user_trash')) {
        parts.shift()
        StartIcon = IconTrash
    }

    if (!StartIcon) {
        if (parts[0] === useSessionStore.getState().user.username) {
            parts.shift()
        }
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
                        <IconSlash
                            className="text-[--wl-text-color] shrink-0"
                            size={18}
                        />
                        <p className={historyStyle['path-text']}>{part}</p>
                    </div>
                )
            })}
        </div>
    )
}

export function FileFmt({ pathName }: { pathName: string }) {
    let nameText = '---'
    let StartIcon: FC<{ className: string }> = IconFolder
    if (pathName) {
        ;({ nameText, StartIcon } = filenameFromPath(pathName))
    }

    return (
        <div className="flex items-center w-max min-w-0 max-w-full">
            {StartIcon && (
                <StartIcon className="theme-text m-1 shrink-0 text-[--wl-text-color]" />
            )}
            <p className="theme-text select-none font-semibold text-lg text-nowrap truncate">
                {nameText}
            </p>
        </div>
    )
}
