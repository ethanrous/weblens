import { IconArrowLeft, IconClock } from '@tabler/icons-react'
import { FileApi } from '@weblens/api/FileBrowserApi'
import FilesErrorDisplay from '@weblens/components/NotFound.tsx'
import { useSessionStore } from '@weblens/components/UserInfo'
import FileHistoryPane from '@weblens/components/filebrowser/historyPane.tsx'
import Crumbs from '@weblens/lib/Crumbs.tsx'
import WeblensButton from '@weblens/lib/WeblensButton'
import { useKeyDown } from '@weblens/lib/hooks'
import { TransferCard } from '@weblens/pages/FileBrowser/DropSpot'
import { historyDateTime } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { DirViewModeT } from '@weblens/pages/FileBrowser/FileBrowserTypes'
import FileSortBox from '@weblens/pages/FileBrowser/FileSortBox'
import { ErrorHandler } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import WeblensFile, { FbMenuModeT } from '@weblens/types/files/File'
import FileColumns from '@weblens/types/files/FileColumns'
import FileGrid from '@weblens/types/files/FileGrid'
import { FileRows } from '@weblens/types/files/FileRows'
import { ReactElement, useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { FbModeT, useFileBrowserStore } from '../../store/FBStateControl'
import { PresentationFile } from '../Presentation'
import dirViewStyle from './directoryView.module.scss'

function SingleFile({ file }: { file: WeblensFile }) {
    if (!file.Id()) {
        return (
            <FilesErrorDisplay
                resourceType="Share"
                link="/files/home"
                setNotFound={() => {}}
                error={404}
                notFound={true}
            />
        )
    }

    return (
        <div className="flex h-full w-full flex-row justify-around pb-2">
            <div className="flex h-full w-full items-center justify-center p-6">
                <PresentationFile file={file} />
            </div>
        </div>
    )
}

function DirView({
    filesError,
    setFilesError,
}: {
    filesError: number
    setFilesError: (err: number) => void
    searchFilter: string
}) {
    const nav = useNavigate()

    const fullViewRef = useRef<HTMLDivElement>(null)
    const dragBoxRef = useRef<HTMLDivElement>(null)

    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesList = useFileBrowserStore((state) => state.filesLists)
    const viewOpts = useFileBrowserStore((state) => state.viewOpts)
    const moveDest = useFileBrowserStore((state) => state.moveDest)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const setViewOptions = useFileBrowserStore((state) => state.setViewOptions)
    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)

    const user = useSessionStore((state) => state.user)

    const activeList = filesList.get(folderInfo?.Id()) || []

    const isSingleSharedFile =
        (mode === FbModeT.default || mode === FbModeT.share) &&
        folderInfo?.Id() &&
        !folderInfo.IsFolder()

    // useEffect(() => {
    //     if (isSingleSharedFile && presentingId !== folderInfo.Id()) {
    //         setPresentingId(folderInfo.Id())
    //     }
    // }, [folderInfo])

    useKeyDown('Escape', (e) => {
        if (isSingleSharedFile) {
            e.stopPropagation()
            nav('/files/shared')
        }
    })

    let fileDisplay: ReactElement
    if (filesError) {
        fileDisplay = (
            <FilesErrorDisplay
                error={filesError}
                resourceType="Folder"
                link="/files/home"
                setNotFound={setFilesError}
                notFound={true}
            />
        )
    } else if (
        (mode === FbModeT.default || mode === FbModeT.share) &&
        folderInfo?.Id() &&
        !folderInfo.IsFolder()
    ) {
        fileDisplay = <SingleFile file={folderInfo} />
    } else if (viewOpts.dirViewMode === DirViewModeT.List) {
        fileDisplay = <FileRows files={activeList} />
    } else if (viewOpts.dirViewMode === DirViewModeT.Grid) {
        fileDisplay = <FileGrid files={activeList} />
    } else if (viewOpts.dirViewMode === DirViewModeT.Columns) {
        fileDisplay = <FileColumns />
    } else {
        console.error(
            'Could not find valid directory view from state. View mode:',
            viewOpts.dirViewMode
        )
        console.debug('Defaulting view mode to grid')
        setViewOptions({ dirViewMode: DirViewModeT.Grid })
        return null
    }

    let dropAction = ''
    if (draggingState === DraggingStateT.InternalDrag) {
        dropAction = 'Move'
    } else if (draggingState === DraggingStateT.ExternalDrag) {
        dropAction = 'Upload'
    }

    return (
        <div className="flex h-full flex-col" ref={fullViewRef}>
            <DirViewHeader />
            <TransferCard
                action={dropAction}
                destination={moveDest}
                boundRef={fullViewRef}
            />
            <div
                className="flex h-0 w-full grow"
                ref={dragBoxRef}
                onDragEnter={(e) => {
                    e.stopPropagation()
                    e.preventDefault()
                    setDragging(DraggingStateT.ExternalDrag)
                    setMoveDest(folderInfo?.Id())
                }}
                onDragOver={(e) => e.preventDefault()}
                onDragLeave={(e) => {
                    e.stopPropagation()

                    if (dragBoxRef.current?.contains(e.relatedTarget as Node)) {
                        return
                    }

                    setDragging(DraggingStateT.NoDrag)
                }}
            >
                <div className="flex h-full w-full min-w-[20vw] flex-col">
                    <div className="flex h-[200px] max-w-full grow flex-row">
                        <div className="ml-3 w-0 shrink grow mr-0.5">
                            {fileDisplay}
                        </div>
                    </div>
                </div>
                {user.isLoggedIn && <FileHistoryPane />}
            </div>
        </div>
    )
}

function DirViewHeader() {
    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const pastTime = useFileBrowserStore((state) => state.pastTime)
    const selected = useFileBrowserStore((state) => state.selected)
    const historyPaneOpen = useFileBrowserStore(
        (state) => state.historyPaneOpen
    )

    const setLocationState = useFileBrowserStore(
        (state) => state.setLocationState
    )
    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setSelectedMoved = useFileBrowserStore(
        (state) => state.setSelectedMoved
    )
    const setHistoryPaneOpen = useFileBrowserStore(
        (state) => state.setHistoryPaneOpen
    )

    const [viewingFolder, setViewingFolder] = useState<boolean>(false)
    const [hoverTime, setHoverTime] = useState<boolean>(false)

    const moveSelectedTo = useCallback(
        (folderId: string) => {
            const activeIds = Array.from(selected.keys())
            setSelectedMoved(activeIds)
            setDragging(DraggingStateT.NoDrag)

            FileApi.moveFiles({
                fileIds: activeIds,
                newParentId: folderId,
            }).catch(ErrorHandler)
        },
        [selected, folderInfo?.id, setDragging, setSelectedMoved]
    )

    useEffect(() => {
        if (!folderInfo) {
            return
        }

        setViewingFolder(folderInfo.IsFolder())
    }, [folderInfo])

    return (
        <div className="flex h-max flex-col">
            <div className={dirViewStyle.dirViewHeaderWrapper}>
                {(mode === FbModeT.default || mode === FbModeT.share) && (
                    <Crumbs navOnLast={false} moveSelectedTo={moveSelectedTo} />
                )}
                {(mode === FbModeT.share || viewingFolder) && <FileSortBox />}
                <WeblensButton
                    Left={IconClock}
                    className="mr-3 ml-3"
                    tooltip={{
                        content: 'View History',
                        position: 'right',
                        className: 'mr-8',
                    }}
                    toggleOn={historyPaneOpen}
                    onClick={() => {
                        setHistoryPaneOpen(!historyPaneOpen)
                    }}
                />
            </div>
            {pastTime && pastTime.getTime() !== 0 && (
                <div
                    className={dirViewStyle.pastTimeBox}
                    onClick={(e) => {
                        e.stopPropagation()
                        setHoverTime(false)
                        setLocationState({
                            contentId: folderInfo.Id(),
                            pastTime: new Date(0),
                        })
                    }}
                    onMouseOver={(e) => {
                        e.stopPropagation()
                        setHoverTime(true)
                    }}
                    onMouseLeave={(e) => {
                        e.stopPropagation()
                        setHoverTime(false)
                    }}
                >
                    <p
                        className="crumb-text pointer-events-none absolute ml-2 text-xl"
                        style={{ opacity: hoverTime ? 1 : 0 }}
                    >
                        Back to the future
                    </p>
                    {hoverTime && <IconArrowLeft />}
                    {!hoverTime && <IconClock />}
                    <p
                        className="crumb-text ml-2 text-xl"
                        style={{ opacity: hoverTime ? 0 : 1 }}
                    >
                        {historyDateTime(pastTime.getTime())}
                    </p>
                </div>
            )}
        </div>
    )
}

function DirectoryView({
    filesError,
    setFilesError,
}: {
    filesError: number
    setFilesError: (err: number) => void
    searchFilter: string
}) {
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const setMenu = useFileBrowserStore((state) => state.setMenu)

    return (
        <div
            draggable={false}
            className="h-full w-0 shrink-0 grow"
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
                    menuTarget: folderInfo?.Id() || '',
                    menuPos: { x: e.clientX, y: e.clientY },
                    menuState: FbMenuModeT.Default,
                })
            }}
        >
            <div className="h-full w-full">
                <DirView
                    filesError={filesError}
                    setFilesError={setFilesError}
                    searchFilter={''}
                />
            </div>
        </div>
    )
}

export default DirectoryView
