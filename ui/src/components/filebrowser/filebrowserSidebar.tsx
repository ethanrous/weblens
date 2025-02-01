import { Divider, FileButton } from '@mantine/core'
import {
    IconFolderPlus,
    IconHome,
    IconPlus,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import { FileApi, FolderApi } from '@weblens/api/FileBrowserApi'
import { useSessionStore } from '@weblens/components/UserInfo'
import UsageInfo from '@weblens/components/filebrowser/usageInfo'
import { useResizeDrag, useWindowSize } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { ButtonActionHandler } from '@weblens/lib/buttonTypes'
import { HandleUploadButton } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { TasksDisplay } from '@weblens/pages/FileBrowser/TaskProgress'
import UploadStatus from '@weblens/pages/FileBrowser/UploadStatus'
import fbStyle from '@weblens/pages/FileBrowser/style/fileBrowserStyle.module.scss'
import { ErrorHandler } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import WeblensFile from '@weblens/types/files/File'
import { goToFile } from '@weblens/types/files/FileDragLogic'
import filesStyle from '@weblens/types/files/filesStyle.module.scss'
import { humanFileSize } from '@weblens/util'
import { useCallback, useEffect, useState } from 'react'

import {
    FbModeT,
    ShareRoot,
    useFileBrowserStore,
} from '../../store/FBStateControl'

const SIDEBAR_BREAKPOINT = 650
const SIDEBAR_DEFAULT_WIDTH = 300
const SIDEBAR_MIN_OPEN_WIDTH = 200

function FBSidebar() {
    const user = useSessionStore((state) => state.user)

    const windowSize = useWindowSize()
    const [resizing, setResizing] = useState(false)
    const [resizeOffset, setResizeOffset] = useState(
        windowSize?.width > SIDEBAR_BREAKPOINT ? SIDEBAR_DEFAULT_WIDTH : 75
    )
    useResizeDrag(resizing, setResizing, (s) => {
        setResizeOffset(Math.min(s > SIDEBAR_MIN_OPEN_WIDTH ? s : 75, 600))
    })

    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const selected = useFileBrowserStore((state) => state.selected)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)
    const setSelectedMoved = useFileBrowserStore(
        (state) => state.setSelectedMoved
    )

    useEffect(() => {
        if (
            windowSize.width < SIDEBAR_BREAKPOINT &&
            resizeOffset >= SIDEBAR_DEFAULT_WIDTH
        ) {
            setResizeOffset(75)
        } else if (
            windowSize.width >= SIDEBAR_BREAKPOINT &&
            resizeOffset < SIDEBAR_DEFAULT_WIDTH
        ) {
            setResizeOffset(SIDEBAR_DEFAULT_WIDTH)
        }
    }, [windowSize.width])

    const homeMouseOver = useCallback(() => {
        if (draggingState !== DraggingStateT.NoDrag) {
            setMoveDest('Home')
        }
    }, [draggingState])

    const mouseLeave = useCallback(() => {
        if (draggingState !== DraggingStateT.NoDrag) {
            setMoveDest('')
        }
    }, [draggingState])

    const homeMouseUp: ButtonActionHandler = useCallback(
        async (e) => {
            e.stopPropagation()
            setMoveDest('')
            if (draggingState !== DraggingStateT.NoDrag) {
                const selectedIds = Array.from(selected.keys())
                setSelectedMoved(selectedIds)
                setDragging(DraggingStateT.NoDrag)
                return FileApi.moveFiles({
                    fileIds: selectedIds,
                    newParentId: user.homeId,
                })
            } else {
                goToFile(
                    new WeblensFile({
                        id: user.homeId,
                        isDir: true,
                    }),
                    true
                )
            }
        },
        [selected, draggingState]
    )

    const trashMouseOver = useCallback(() => {
        if (draggingState !== DraggingStateT.NoDrag) {
            setMoveDest('.user_trash')
        }
    }, [draggingState])

    const trashMouseUp: ButtonActionHandler = useCallback(
        async (e) => {
            e.stopPropagation()
            setMoveDest('')
            if (draggingState !== DraggingStateT.NoDrag) {
                setSelectedMoved(Array.from(selected.keys()))
                setDragging(DraggingStateT.NoDrag)
                return FileApi.moveFiles({
                    fileIds: Array.from(selected.keys()),
                    newParentId: user.trashId,
                })
            } else {
                goToFile(
                    new WeblensFile({
                        id: user.trashId,
                        filename: '.user_trash',
                        isDir: true,
                        modifiable: false,
                    }),
                    true
                )
            }
        },
        [selected, draggingState]
    )

    const [namingFolder, setNamingFolder] = useState(false)

    const navToShared = useCallback(() => {
        goToFile(ShareRoot, true)
    }, [])

    const navToExternal: ButtonActionHandler = useCallback((e) => {
        e.stopPropagation()
        goToFile(new WeblensFile({ id: 'external', isDir: true }), true)
    }, [])

    const newFolder: ButtonActionHandler = useCallback(
        (e) => {
            e.stopPropagation()
            setNamingFolder(true)
        },
        [setNamingFolder]
    )

    return (
        <div
            className="flex flex-row items-start w-full h-full grow-0 shrink-0"
            style={{
                width: resizeOffset,
            }}
        >
            <div
                className={fbStyle['sidebar-container']}
                data-mini={resizeOffset < SIDEBAR_MIN_OPEN_WIDTH}
            >
                <div className={fbStyle['sidebar-upper-half']}>
                    <div className="flex flex-col w-full gap-1 items-center">
                        <WeblensButton
                            label="Home"
                            fillWidth
                            squareSize={40}
                            toggleOn={
                                folderInfo?.Id() === user?.homeId &&
                                mode === FbModeT.default
                            }
                            float={
                                draggingState === DraggingStateT.InternalDrag
                            }
                            disabled={!user.isLoggedIn}
                            allowRepeat={false}
                            Left={IconHome}
                            onMouseOver={homeMouseOver}
                            onMouseLeave={mouseLeave}
                            onMouseUp={homeMouseUp}
                        />

                        <WeblensButton
                            label="Shared"
                            fillWidth
                            squareSize={40}
                            toggleOn={mode === FbModeT.share && !shareId}
                            disabled={
                                draggingState !== DraggingStateT.NoDrag ||
                                !user.isLoggedIn
                            }
                            allowRepeat={false}
                            Left={IconUsers}
                            onClick={navToShared}
                        />

                        <WeblensButton
                            label="Trash"
                            fillWidth
                            squareSize={40}
                            float={
                                draggingState === DraggingStateT.InternalDrag
                            }
                            toggleOn={
                                folderInfo?.Id() === user?.trashId &&
                                mode === FbModeT.default
                            }
                            disabled={
                                (draggingState !== DraggingStateT.NoDrag &&
                                    folderInfo?.Id() === user?.trashId &&
                                    mode === FbModeT.default) ||
                                !user.isLoggedIn
                            }
                            allowRepeat={false}
                            Left={IconTrash}
                            onMouseOver={trashMouseOver}
                            onMouseLeave={mouseLeave}
                            onMouseUp={trashMouseUp}
                            Right={resizeOffset > 200 ? TrashSize : null}
                        />

                        <div className="p-1" />

                        {user?.admin && (
                            <WeblensButton
                                label="External"
                                fillWidth
                                squareSize={40}
                                toggleOn={mode === FbModeT.external}
                                allowRepeat={false}
                                Left={IconServer}
                                // disabled={draggingState !== DraggingStateT.NoDrag}
								tooltip={'Coming Soon'}
                                disabled={true}
                                onClick={navToExternal}
                            />
                        )}

                        <div className="p-1" />

                        {!namingFolder && (
                            <WeblensButton
                                label="New Folder"
                                fillWidth
                                squareSize={40}
                                Left={IconFolderPlus}
                                showSuccess={false}
                                disabled={
                                    draggingState !== DraggingStateT.NoDrag ||
                                    !folderInfo?.IsModifiable()
                                }
                                onClick={newFolder}
                            />
                        )}
                        {namingFolder && (
                            <WeblensInput
                                squareSize={40}
                                placeholder={'New Folder'}
                                buttonIcon={IconPlus}
                                closeInput={() => setNamingFolder(false)}
                                autoFocus
                                onComplete={async (newName) => {
                                    return FolderApi.createFolder(
                                        {
                                            parentFolderId: folderInfo.Id(),
                                            newFolderName: newName,
                                        },
                                        shareId
                                    ).then(() => setNamingFolder(false))
                                }}
                            />
                        )}

                        <FileButton
                            onChange={(files) => {
                                HandleUploadButton(
                                    files,
                                    folderInfo.Id(),
                                    false,
                                    shareId
                                ).catch(ErrorHandler)
                            }}
                            accept="file"
                            multiple
                        >
                            {(props) => {
                                return (
                                    <WeblensButton
                                        label="Upload"
                                        squareSize={40}
                                        fillWidth
                                        showSuccess={false}
                                        disabled={
                                            draggingState !==
                                                DraggingStateT.NoDrag ||
                                            !folderInfo?.IsModifiable()
                                        }
                                        Left={IconUpload}
                                        onClick={props.onClick}
                                    />
                                )
                            }}
                        </FileButton>
                    </div>
                    <Divider w={'100%'} my="md" size={1.5} />
                    <UsageInfo />
                </div>
                <TasksDisplay />
                <UploadStatus />
            </div>
            <div
                draggable={false}
                className={fbStyle['resize-bar-wrapper']}
                onMouseDown={(e) => {
                    e.preventDefault()
                    setResizing(true)
                }}
            >
                <div className={fbStyle['resize-bar']} />
            </div>
        </div>
    )
}

function TrashSize() {
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const trashSize = useFileBrowserStore((state) => state.trashDirSize)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const user = useSessionStore((state) => state.user)
    if (trashSize <= 0) {
        return null
    }
    const [trashSizeValue, trashSizeUnit] = humanFileSize(trashSize)

    if (pastTime.getTime() !== 0) {
        return null
    }

    return (
        <div
            className={filesStyle['trash-size-box']}
            data-selected={
                folderInfo?.Id() === user?.trashId && mode === FbModeT.default
                    ? 1
                    : 0
            }
        >
            <div className={filesStyle['file-size-box']}>
                <p
                    className={filesStyle['file-size-text']}
                >{`${trashSizeValue}${trashSizeUnit}`}</p>
            </div>
        </div>
    )
}

export default FBSidebar
