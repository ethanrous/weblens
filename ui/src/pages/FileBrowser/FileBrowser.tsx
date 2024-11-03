import { Divider, FileButton } from '@mantine/core'

// Icons
import {
    IconArrowLeft,
    IconClock,
    IconFiles,
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconPlus,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import {
    CreateFolder,
    GetFileInfo,
    GetFolderData,
    moveFiles,
} from '@weblens/api/FileBrowserApi'
import { useSubscribe } from '@weblens/api/Websocket'
import HeaderBar from '@weblens/components/HeaderBar'
import {
    useResize,
    useResizeDrag,
    useWindowSize,
} from '@weblens/components/hooks'
import FilesErrorDisplay from '@weblens/components/NotFound'
import {
    PresentationContainer,
    PresentationFile,
} from '@weblens/components/Presentation'
import { useSessionStore } from '@weblens/components/UserInfo'
import Crumbs from '@weblens/lib/Crumbs'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { WeblensFile } from '@weblens/types/files/File'
import FileGrid from '@weblens/types/files/FileGrid'
import { FileContextMenu } from '@weblens/types/files/FileMenu'
import { FileRows } from '@weblens/types/files/FileRows'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import { getFileShare } from '@weblens/types/share/shareQuery'
import { humanFileSize } from '@weblens/util'
import { memo, ReactElement, useCallback, useEffect, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

// Weblens
import { FbModeT, useFileBrowserStore } from './FBStateControl'

import {
    getRealId,
    handleDragOver,
    HandleDrop,
    HandleUploadButton,
    historyDate,
    MoveSelected,
    uploadViaUrl,
    useKeyDownFileBrowser,
    usePaste,
} from './FileBrowserLogic'
import {
    DirViewWrapper,
    DraggingCounter,
    DropSpot,
    TransferCard,
    WebsocketStatus,
} from './FileBrowserMiscComponents'
import FileInfoPane from './FileInfoPane'

import FileSortBox from './FileSortBox'
import { StatTree } from './FileStatTree'
import SearchDialogue from './SearchDialogue'
import { TasksDisplay } from './TaskProgress'
import UploadStatus from './UploadStatus'
import './style/fileBrowserStyle.scss'
import '@weblens/components/style.scss'
import FileColumns from '@weblens/types/files/FileColumns'
import { goToFile } from '@weblens/types/files/FileDragLogic'
import { ButtonActionHandler } from '@weblens/lib/buttonTypes'
import { DirViewModeT } from './FileBrowserTypes'

function PasteImageDialogue() {
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const pasteImage = useFileBrowserStore((state) => state.pasteImgBytes)

    const setPasteImgBytes = useFileBrowserStore(
        (state) => state.setPasteImgBytes
    )

    const media = new WeblensMedia({ contentId: 'paste' })
    media.SetThumbnailBytes(pasteImage)

    return (
        <PresentationContainer
            onClick={() => {
                setPasteImgBytes(null)
            }}
        >
            <div className="flex flex-col absolute w-full h-full justify-center items-center z-[2]">
                <p className="font-bold text-[40px] pb-[50px]">
                    Upload from clipboard?
                </p>
                <div
                    className="h-1/2 w-max bg-bottom-grey p-3 rounded-lg overflow-hidden"
                    onClick={(e) => {
                        e.stopPropagation()
                    }}
                >
                    <MediaImage media={media} quality={PhotoQuality.LowRes} />
                </div>
                <div className="flex flex-row justify-between w-[50%] gap-6">
                    <WeblensButton
                        label={'Cancel'}
                        squareSize={50}
                        fillWidth
                        subtle
                        onClick={(e) => {
                            e.stopPropagation()

                            setPasteImgBytes(null)
                        }}
                    />
                    <WeblensButton
                        label={'Upload'}
                        squareSize={50}
                        fillWidth
                        onClick={(e) => {
                            e.stopPropagation()
                            uploadViaUrl(pasteImage, contentId, filesMap).then(
                                () => setPasteImgBytes(null)
                            )
                        }}
                    />
                </div>
            </div>
        </PresentationContainer>
    )
}

const SIDEBAR_BREAKPOINT = 650

function GlobalActions() {
    const user = useSessionStore((state) => state.user)

    const windowSize = useWindowSize()
    const [resizing, setResizing] = useState(false)
    const [resizeOffset, setResizeOffset] = useState(
        windowSize?.width > SIDEBAR_BREAKPOINT ? 300 : 75
    )
    useResizeDrag(resizing, setResizing, (s) => {
        setResizeOffset(Math.min(s > 200 ? s : 75, 600))
    })

    const trashSize = useFileBrowserStore((state) => state.trashDirSize)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const selectedIds = useFileBrowserStore((state) =>
        Array.from(state.selected.keys())
    )
    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)

    useEffect(() => {
        if (windowSize.width < SIDEBAR_BREAKPOINT && resizeOffset >= 300) {
            setResizeOffset(75)
        } else if (
            windowSize.width >= SIDEBAR_BREAKPOINT &&
            resizeOffset < 300
        ) {
            setResizeOffset(300)
        }
    }, [windowSize.width])

    const [trashSizeValue, trashSizeUnit] = humanFileSize(trashSize)

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
        (e) => {
            e.stopPropagation()
            setMoveDest('')
            if (draggingState !== DraggingStateT.NoDrag) {
                moveFiles(selectedIds, user.homeId)
                setDragging(DraggingStateT.NoDrag)
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
        [selectedIds, draggingState]
    )

    const trashMouseOver = useCallback(() => {
        if (draggingState !== DraggingStateT.NoDrag) {
            setMoveDest('Trash')
        }
    }, [draggingState])

    const trashMouseUp: ButtonActionHandler = useCallback(
        (e) => {
            e.stopPropagation()
            setMoveDest('')
            if (draggingState !== DraggingStateT.NoDrag) {
                moveFiles(selectedIds, user.trashId)
                setDragging(DraggingStateT.NoDrag)
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
        [selectedIds, draggingState]
    )

    const [namingFolder, setNamingFolder] = useState(false)

    const navToShared = useCallback(() => {
        goToFile(
            new WeblensFile({ id: 'shared', filename: 'SHARED', isDir: true }),
            true
        )
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
            <div className="sidebar-container">
                <div className="flex flex-col w-full gap-1 items-center">
                    <WeblensButton
                        label="Home"
                        fillWidth
                        squareSize={40}
                        toggleOn={
                            folderInfo?.Id() === user?.homeId &&
                            mode === FbModeT.default
                        }
                        float={draggingState === 1}
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
                        toggleOn={mode === FbModeT.share && shareId === ''}
                        disabled={
                            draggingState !== DraggingStateT.NoDrag ||
                            !user.isLoggedIn
                        }
                        allowRepeat={false}
                        Left={IconUsers}
                        onClick={navToShared}
                    />

                    {trashSize !== 0 && (
                        <div
                            className="relative w-full translate-y-[2px] z-20"
                            data-selected={
                                folderInfo?.Id() === user?.trashId &&
                                mode === FbModeT.default
                                    ? 1
                                    : 0
                            }
                        >
                            <div className="absolute file-size-box">
                                <p className="file-size-text">{`${trashSizeValue}${trashSizeUnit}`}</p>
                            </div>
                        </div>
                    )}
                    <WeblensButton
                        label="Trash"
                        fillWidth
                        squareSize={40}
                        float={draggingState === 1}
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
                            disabled={draggingState !== DraggingStateT.NoDrag}
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
                                draggingState !== 0 ||
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
                            onComplete={(newName) =>
                                CreateFolder(
                                    folderInfo.Id(),
                                    newName,
                                    [],
                                    false,
                                    shareId
                                ).then(() => setNamingFolder(false))
                            }
                        />
                    )}

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
                                    label="Upload"
                                    squareSize={40}
                                    fillWidth
                                    showSuccess={false}
                                    disabled={
                                        draggingState !== 0 ||
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

                <TasksDisplay />

                <div className="flex grow" />

                <UploadStatus />
            </div>
            <div
                draggable={false}
                className="resize-bar-wrapper"
                onMouseDown={(e) => {
                    e.preventDefault()
                    setResizing(true)
                }}
            >
                <div className="resize-bar" />
            </div>
        </div>
    )
}

const UsageInfo = () => {
    const [box, setBox] = useState(null)
    const size = useResize(box)

    const user = useSessionStore((state) => state.user)

    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const homeSize = useFileBrowserStore((state) => state.homeDirSize)
    const trashSize = useFileBrowserStore((state) => state.trashDirSize)
    const selectedSize = useFileBrowserStore((state) =>
        Array.from(state.selected.keys()).reduce((acc: number, x: string) => {
            return acc + (state.filesMap.get(x)?.GetSize() || 0)
        }, 0)
    )
    const selectedLength = useFileBrowserStore((state) => state.selected.size)
    const mode = useFileBrowserStore((state) => state.fbMode)

    if (folderInfo?.Id() === 'shared' || user === null) {
        return null
    }

    let displaySize = folderInfo?.GetSize() || 0
    if (folderInfo?.Id() === user.homeId) {
        displaySize = displaySize - trashSize
    }

    if (homeSize < displaySize) {
        displaySize = homeSize
    }

    const doGlobalSize = selectedLength === 0 && mode !== FbModeT.share

    let usagePercent = doGlobalSize
        ? (displaySize / homeSize) * 100
        : (selectedSize / displaySize) * 100
    if (!usagePercent || (selectedLength !== 0 && displaySize === 0)) {
        usagePercent = 0
    }

    const miniMode = size.width !== -1 && size.width < 100

    let StartIcon = doGlobalSize ? IconFolder : IconFiles
    let EndIcon = doGlobalSize ? IconHome : IconFolder
    if (miniMode) {
        ;[StartIcon, EndIcon] = [EndIcon, StartIcon]
    }

    return (
        <div
            ref={setBox}
            className="flex flex-col h-max w-full gap-3 mb-2"
            style={{
                alignItems: miniMode ? 'center' : 'flex-start',
            }}
        >
            {!miniMode && (
                <div className="flex flex-row h-max w-full gap-2 items-center justify-between">
                    <h1 className="font-bold text-lg">
                        Usage {usagePercent ? usagePercent.toFixed(2) : 0}%
                    </h1>
                </div>
            )}
            {miniMode && <StartIcon className="background-icon" />}
            <div
                className="relative h-max w-max"
                style={{
                    height: miniMode ? 100 : 20,
                    width: miniMode ? 20 : '100%',
                }}
            >
                <WeblensProgress
                    key={miniMode ? 'usage-vertical' : 'usage-horizontal'}
                    value={usagePercent}
                    orientation={miniMode ? 'vertical' : 'horizontal'}
                />
            </div>
            <div
                className="flex flex-row h-max justify-between items-center"
                style={{
                    width: miniMode ? 'max-content' : '98%',
                }}
            >
                {folderInfo?.Id() !== 'shared' && !miniMode && (
                    <div className="flex flex-row items-center">
                        {<StartIcon className="background-icon" />}
                        <p
                            className="select-none p-1"
                            style={{
                                display: miniMode ? 'none' : 'block',
                            }}
                        >
                            {doGlobalSize
                                ? humanFileSize(displaySize)
                                : humanFileSize(selectedSize)}
                        </p>
                    </div>
                )}
                <div className="flex flex-row justify-end w-max items-center">
                    <p
                        className="select-none p-1"
                        style={{
                            display: miniMode ? 'none' : 'block',
                        }}
                    >
                        {doGlobalSize
                            ? humanFileSize(homeSize)
                            : humanFileSize(displaySize)}
                    </p>
                    {<EndIcon className="background-icon" />}
                </div>
            </div>
        </div>
    )
}

const SingleFile = memo(
    ({ file }: { file: WeblensFile }) => {
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
            <div className="flex flex-row w-full h-full justify-around pb-2">
                <div className="flex justify-center items-center p-6 h-full w-full">
                    <PresentationFile file={file} />
                </div>
            </div>
        )
    },
    (prev, next) => {
        return prev.file === next.file
    }
)

function DirViewHeader({ moveSelected }) {
    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const viewingPast = useFileBrowserStore((state) => state.pastTime)
    const setPastTime = useFileBrowserStore((state) => state.setPastTime)

    const [viewingFolder, setViewingFolder] = useState<boolean>(false)
    const [hoverTime, setHoverTime] = useState<boolean>(false)

    useEffect(() => {
        if (!folderInfo) {
            return
        }

        setViewingFolder(folderInfo.IsFolder())
    }, [folderInfo])

    return (
        <div className="flex flex-col h-max">
            <div className="flex flex-row h-[60px] justify-between items-center pl-2 pt-0">
                {(mode === FbModeT.default || mode === FbModeT.share) && (
                    <Crumbs navOnLast={false} moveSelectedTo={moveSelected} />
                )}
                {(mode === FbModeT.share || viewingFolder) && <FileSortBox />}
            </div>
            {viewingPast && (
                <div
                    className="past-time-box"
                    onClick={(e) => {
                        e.stopPropagation()
                        setHoverTime(false)
                        setPastTime(null)
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
                        className="crumb-text absolute pointer-events-none ml-2 text-[#c4c4c4] text-xl"
                        style={{ opacity: hoverTime ? 1 : 0 }}
                    >
                        Back to present?
                    </p>
                    {hoverTime && <IconArrowLeft />}
                    {!hoverTime && <IconClock />}
                    <p
                        className="crumb-text ml-2 text-[#c4c4c4] text-xl"
                        style={{ opacity: hoverTime ? 0 : 1 }}
                    >
                        {historyDate(viewingPast.getTime())}
                    </p>
                </div>
            )}
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
    const [contentViewRef, setContentViewRef] = useState(null)
    const [fullViewRef, setFullViewRef] = useState(null)

    const mode = useFileBrowserStore((state) => state.fbMode)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const selected = useFileBrowserStore((state) => state.selected)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesList = useFileBrowserStore((state) => state.filesLists)
    const viewOpts = useFileBrowserStore((state) => state.viewOpts)
    const moveDest = useFileBrowserStore((state) => state.moveDest)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const setViewOptions = useFileBrowserStore((state) => state.setViewOptions)

    const user = useSessionStore((state) => state.user)

    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const setDragging = useFileBrowserStore((state) => state.setDragging)

    const activeList = filesList.get(folderInfo?.Id()) || []

    const moveSelectedTo = useCallback(
        (folderId: string) => {
            MoveSelected(Array.from(selected.keys()), folderId).then(() => {
                clearSelected()
            })
        },
        [selected.size, contentId]
    )

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
        // } else if (
        //     activeList.length === 0 &&
        //     loading.length === 0 &&
        //     searchContent !== ''
        // ) {
        //     fileDisplay = <EmptySearch />
    } else if (mode === FbModeT.stats) {
        fileDisplay = <StatTree folderInfo={folderInfo} />
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

    return (
        <div className="flex flex-col h-full" ref={setFullViewRef}>
            <DirViewHeader moveSelected={moveSelectedTo} />
            <TransferCard
                action="Move"
                destination={moveDest}
                boundRef={fullViewRef}
            />
            <div className="flex h-0 w-full grow" ref={setContentViewRef}>
                <DropSpot
                    onDrop={(e) => {
                        HandleDrop(
                            e.dataTransfer.items,
                            contentId,
                            activeList.map((value: WeblensFile) =>
                                value.GetFilename()
                            ),
                            false,
                            ''
                        )
                        setDragging(DraggingStateT.NoDrag)
                    }}
                    stopDragging={() => setDragging(DraggingStateT.NoDrag)}
                    dropSpotTitle={folderInfo?.GetFilename()}
                    dropAllowed={folderInfo?.IsModifiable()}
                    handleDrag={(event) => {
                        console.log('ITS ME??')
                        return
                        handleDragOver(event, setDragging, draggingState)
                    }}
                    wrapperRef={contentViewRef}
                />
                <div className="flex flex-col w-full h-full pl-3">
                    <div className="flex flex-row h-[200px] grow max-w-full">
                        <div className="grow shrink w-0">{fileDisplay}</div>
                    </div>
                </div>
                {user.isLoggedIn && <FileInfoPane />}
            </div>
        </div>
    )
}

function useSearch() {
    const { search } = useLocation()
    const q = new URLSearchParams(search)
    return useCallback(
        (s: string) => {
            const r = q.get(s)
            if (!r) {
                return ''
            }
            return r
        },
        [q]
    )
}

function FileBrowser() {
    const urlPath = useParams()['*']
    const jumpTo = window.location.hash.substring(1)
    const query = useSearch()
    const user = useSessionStore((state) => state.user)
    const nav = useNavigate()

    useEffect(() => {
        useFileBrowserStore.getState().setNav(nav)
    }, [nav])

    const [filesFetchErr, setFilesFetchErr] = useState(0)

    const {
        viewOpts,
        blockFocus,
        loading,
        filesMap,
        folderInfo,
        presentingId,
        isSearching,
        pasteImgBytes,
        fbMode,
        contentId,
        shareId,
        pastTime,
        addLoading,
        removeLoading,
        setLocationState,
        setSelected,
        setFilesData,
        setBlockFocus,
    } = useFileBrowserStore()

    useEffect(() => {
        localStorage.setItem('fbViewOpts', JSON.stringify(viewOpts))
    }, [viewOpts])

    useEffect(() => {
        if (!user) {
            return
        }

        let mode: FbModeT = 0
        let contentId: string = ''
        let shareId: string = ''
        const splitPath: string[] = urlPath
            .split('/')
            .filter((s) => s.length !== 0)

        if (splitPath.length === 0) {
            return
        }

        if (splitPath[0] === 'share') {
            mode = FbModeT.share
            shareId = splitPath[1]
            contentId = splitPath[2]
        } else if (splitPath[0] === 'shared') {
            mode = FbModeT.share
            shareId = ''
            contentId = ''
        } else if (splitPath[0] === 'external') {
            mode = FbModeT.external
            contentId = splitPath[1]
        } else if (splitPath[0] === 'stats') {
            mode = FbModeT.stats
            contentId = splitPath[1]
        } else if (splitPath[0] === 'search') {
            mode = FbModeT.search
            contentId = splitPath[1]
        } else {
            mode = FbModeT.default
            contentId = splitPath[0]
        }

        const timestamp = query('at')
        let pastTime: Date
        if (timestamp) {
            pastTime = new Date(Number(timestamp))
        }

        if (mode === FbModeT.share && shareId && !contentId) {
            getFileShare(shareId).then((s) => {
                nav(`/files/share/${shareId}/${s.fileId}`)
            })
        } else {
            getRealId(contentId, mode, user).then((contentId) => {
                setLocationState({ contentId, mode, shareId, pastTime, jumpTo })
                removeLoading('files')
            })
        }
    }, [urlPath, user, query('at'), jumpTo])

    const { readyState } = useSubscribe(contentId, shareId, user)

    useKeyDownFileBrowser()

    // Hook to handle uploading images from the clipboard
    usePaste(contentId, user, blockFocus)

    // Reset most of the state when we change folders
    useEffect(() => {
        if (contentId === null) {
            return
        }

        const syncState = async () => {
            setFilesFetchErr(0)

            if (!urlPath) {
                nav('/files/home', { replace: true })
            }

            if (urlPath === user?.homeId) {
                const redirect = '/files/home' + window.location.search
                nav(redirect, { replace: true })
            }

            // If we're not ready, leave
            if (fbMode == FbModeT.unset || !user) {
                console.debug(
                    'Not ready to sync state. Mode: ',
                    fbMode,
                    'User:',
                    user
                )
                return
            }

            if (!user.isLoggedIn && fbMode !== FbModeT.share) {
                console.debug('Going to login')
                nav('/login', { state: { returnTo: window.location.pathname } })
            }

            const folder = filesMap.get(contentId)
            if (folder && folder.modifiable !== undefined) {
                if (folder.Id() !== folderInfo.Id()) {
                    setFilesData({
                        selfInfo: folder,
                    })
                }
                return
            }

            const fileData = await GetFolderData(
                contentId,
                fbMode,
                shareId,
                pastTime
            ).catch((r) => {
                setFilesFetchErr(r)
            })

            // If request comes back after we have already navigated away, do nothing
            if (useFileBrowserStore.getState().contentId !== contentId) {
                console.error("Content ID don't match")
                return
            }

            if (fileData) {
                if (
                    fbMode === FbModeT.share &&
                    fileData.self?.owner == user.username
                ) {
                    nav(`/files/${fileData.self.id}`, { replace: true })
                    return
                }

                setFilesData({
                    selfInfo: fileData.self,
                    childrenInfo: fileData.children,
                    parentsInfo: fileData.parents,
                    mediaData: fileData.medias,
                    user: user,
                })
            }

            if (
                (fbMode !== FbModeT.share || contentId) &&
                useFileBrowserStore.getState().selected.size === 0
            ) {
                setSelected([jumpTo ? jumpTo : contentId], true)
            }
        }

        addLoading('files')
        syncState()
            .catch((e) => {
                console.error(e)
                setFilesFetchErr(e)
            })
            .finally(() => removeLoading('files'))
    }, [user, contentId, shareId, fbMode, pastTime, jumpTo])

    console.log('HER?')

    return (
        <div className="h-screen flex flex-col">
            <HeaderBar
                setBlockFocus={setBlockFocus}
                page={'files'}
                loading={loading}
            />
            <DraggingCounter />
            <PresentationFile file={filesMap.get(presentingId)} />
            {pasteImgBytes && <PasteImageDialogue />}
            {isSearching && (
                <div className="flex items-center justify-center w-screen h-screen absolute z-50 backdrop-blur-sm bg-[#00000088] px-[30%] py-[10%]">
                    <SearchDialogue
                        text={''}
                        visitFunc={(loc) => {
                            GetFileInfo(loc, '').then((f) => {
                                if (!f) {
                                    console.error(
                                        'Could not find file to nav to'
                                    )
                                    return
                                }

                                goToFile(new WeblensFile(f), true)
                            })
                        }}
                    />
                </div>
            )}
            <FileContextMenu />
            <div className="absolute bottom-1 left-1">
                <WebsocketStatus ready={readyState} />
            </div>
            <div className="flex flex-row grow h-[90vh] items-start">
                <GlobalActions />
                <DirViewWrapper>
                    <DirView
                        filesError={filesFetchErr}
                        setFilesError={setFilesFetchErr}
                        searchFilter={''}
                    />
                </DirViewWrapper>
            </div>
        </div>
    )
}

export default FileBrowser
