import { IconArrowLeft, IconClock } from '@tabler/icons-react'
import { FileApi } from '@weblens/api/FileBrowserApi'
import FilesErrorDisplay from '@weblens/components/NotFound'
import { useSessionStore } from '@weblens/components/UserInfo'
import FileHistoryPane from '@weblens/components/filebrowser/historyPane'
import Crumbs from '@weblens/lib/Crumbs'
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
import { ReactElement, useCallback, useEffect, useState } from 'react'

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
    const [fullViewRef, setFullViewRef] = useState<HTMLDivElement>(null)

    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesList = useFileBrowserStore((state) => state.filesLists)
    const viewOpts = useFileBrowserStore((state) => state.viewOpts)
    const moveDest = useFileBrowserStore((state) => state.moveDest)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const setViewOptions = useFileBrowserStore((state) => state.setViewOptions)
    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)
    const [dragBoxRef, setDragBoxRef] = useState<HTMLDivElement>()

    const user = useSessionStore((state) => state.user)

    const activeList = filesList.get(folderInfo?.Id()) || []

    let fileDisplay: ReactElement
    if (filesError) {
        fileDisplay = (
            <FilesErrorDisplay
                error={filesError}
                resourceType="Folder"
                link="/files/home"
                setNotFound={setFilesError}
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
        <div className="flex h-full flex-col" ref={setFullViewRef}>
            <DirViewHeader />
            <TransferCard
                action={dropAction}
                destination={moveDest}
                boundRef={fullViewRef}
            />
            <div
                className="flex h-0 w-full grow"
                ref={setDragBoxRef}
                onDragEnter={(e) => {
                    e.stopPropagation()
                    e.preventDefault()
                    setDragging(DraggingStateT.ExternalDrag)
                    setMoveDest(folderInfo?.Id())
                }}
                onDragOver={(e) => e.preventDefault()}
                onDragLeave={(e) => {
                    e.stopPropagation()
                    if (dragBoxRef.contains(e.relatedTarget as Node)) {
                        return
                    }
                    setDragging(DraggingStateT.NoDrag)
                }}
            >
                <div className="flex h-full w-full min-w-[20vw] flex-col">
                    <div className="flex h-[200px] max-w-full grow flex-row">
                        <div className="ml-1 w-0 shrink grow p-1">
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

    const setLocationState = useFileBrowserStore(
        (state) => state.setLocationState
    )
    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setSelectedMoved = useFileBrowserStore(
        (state) => state.setSelectedMoved
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
        [selected, folderInfo?.Id()]
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
                    menuTarget: '',
                    menuPos: { x: e.clientX, y: e.clientY },
                    menuState: FbMenuModeT.Default,
                })
            }}
        >
            <div className="h-full w-full p-2">
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
