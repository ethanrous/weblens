import {
    IconFiles,
    IconFolderPlus,
    IconFolders,
    IconHome,
    IconPlus,
    IconServer,
    IconTrash,
    IconUsers,
} from '@tabler/icons-react'
import { FileApi, FolderApi } from '@weblens/api/FileBrowserApi'
import { useSessionStore } from '@weblens/components/UserInfo'
import UsageInfo from '@weblens/components/filebrowser/usageInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensFileButton from '@weblens/lib/WeblensFileButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { ButtonActionHandler } from '@weblens/lib/buttonTypes'
import { useResizeDrag, useWindowSize } from '@weblens/lib/hooks'
import { TaskProgressMini } from '@weblens/pages/FileBrowser/TaskProgress'
import UploadStatus from '@weblens/pages/FileBrowser/UploadStatus'
import fbStyle from '@weblens/pages/FileBrowser/style/fileBrowserStyle.module.scss'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import WeblensFile from '@weblens/types/files/File'
import { goToFile } from '@weblens/types/files/FileDragLogic'
import { humanFileSize } from '@weblens/util'
import { useCallback, useEffect, useMemo, useState } from 'react'

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
    const sidebarCollapsed = useFileBrowserStore(
        (state) => state.sidebarCollapsed
    )

    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setSidebarCollapsed = useFileBrowserStore(
        (state) => state.setSidebarCollapsed
    )
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

    useEffect(() => {
        if (resizeOffset < SIDEBAR_MIN_OPEN_WIDTH && !sidebarCollapsed) {
            setSidebarCollapsed(true)
        } else if (resizeOffset >= SIDEBAR_MIN_OPEN_WIDTH && sidebarCollapsed) {
            setSidebarCollapsed(false)
        }
    }, [resizeOffset])

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
            className="animate-fade-in flex h-full w-full shrink-0 grow-0 flex-row items-start"
            style={{
                width: resizeOffset,
            }}
        >
            <div
                className={fbStyle.sidebarContainer}
                data-mini={resizeOffset < SIDEBAR_MIN_OPEN_WIDTH}
            >
                <div className={fbStyle.sidebarUpperHalf}>
                    <div className="flex w-full flex-col items-center gap-2">
                        <WeblensButton
                            label="Home"
                            fillWidth
                            squareSize={40}
                            flavor={
                                folderInfo?.Id() === user?.homeId &&
                                mode === FbModeT.default
                                    ? 'default'
                                    : 'outline'
                            }
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
                            flavor={
                                mode === FbModeT.share && !shareId
                                    ? 'default'
                                    : 'outline'
                            }
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
                            flavor={
                                folderInfo?.Id() === user?.trashId &&
                                mode === FbModeT.default
                                    ? 'default'
                                    : 'outline'
                            }
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

                        <WeblensFileButton
                            folderId={folderInfo?.Id()}
                            shareId={shareId}
                            buttonProps={{
                                fillWidth: true,
                                label: 'Upload',
                            }}
                        />
                    </div>
                    <UsageInfo />
                </div>
                <SelectedBox />
                <TaskProgressMini />
                <UploadStatus />
            </div>
            <div
                draggable={false}
                className={fbStyle.resizeBarWrapper}
                onMouseDown={(e) => {
                    e.preventDefault()
                    setResizing(true)
                }}
            >
                <div className={fbStyle.resizeBar} />
            </div>
        </div>
    )
}

function TrashSize() {
    const trashSize = useFileBrowserStore((state) => state.trashDirSize)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    if (trashSize <= 0) {
        return null
    }
    const [trashSizeValue, trashSizeUnit] = humanFileSize(trashSize)

    if (pastTime.getTime() !== 0) {
        return null
    }

    return (
        <div className="z-20 ml-2 flex h-full w-max items-center contain-size">
            <span className="">{`${trashSizeValue}${trashSizeUnit}`}</span>
        </div>
    )
}

function SelectedBox() {
    const selected = useFileBrowserStore((state) => state.selected)
    const selectedLength = selected.size

    const [closed, setClosed] = useState(selected.size === 0)
    const sidebarCollapsed = useFileBrowserStore(
        (state) => state.sidebarCollapsed
    )
    const filesMap = useFileBrowserStore((state) => state.filesMap)

    const { selectedFolderCount, selectedFileCount } = useMemo(() => {
        if (selected.size !== 0) {
            setClosed(false)
        }

        let selectedFolderCount = 0
        let selectedFileCount = 0
        Array.from(selected.keys()).forEach((fileId) => {
            const f = filesMap.get(fileId)
            if (!f) {
                return
            }
            if (f.IsFolder()) {
                selectedFolderCount++
            } else {
                selectedFileCount++
            }
        })

        return { selectedFolderCount, selectedFileCount }
    }, [selected.size])

    return (
        <div
            className="animate-popup bg-background-secondary mt-2 h-max w-full flex-row items-center justify-between rounded-lg p-2 transition"
            onTransitionEnd={() => {
                if (selectedLength === 0) {
                    setClosed(true)
                } else {
                    setClosed(false)
                }
            }}
            style={{
                display: closed ? 'none' : 'flex',
                opacity: selectedLength > 0 ? 100 : 0,
                transform: selectedLength > 0 ? 'scale(1)' : 'scale(0.90)',
                flexDirection: sidebarCollapsed ? 'column' : 'row',
            }}
        >
            <div className="text-color-text-primary flex h-max w-min items-center">
                <IconFiles />
                <span className="p-1 select-none">{selectedFileCount}</span>
            </div>
            {!sidebarCollapsed && (
                <>
                    <span className="text-color-text-secondary">Selected</span>
                    <div className="text-color-text-primary flex h-max w-min items-center">
                        <span className="p-1 select-none">
                            {selectedFolderCount}
                        </span>
                        <IconFolders />
                    </div>
                </>
            )}
            {sidebarCollapsed && (
                <div className="text-color-text-primary flex h-max w-min items-center">
                    <IconFolders />
                    <span className="p-1 select-none">
                        {selectedFolderCount}
                    </span>
                </div>
            )}
        </div>
    )
}

export default FBSidebar
