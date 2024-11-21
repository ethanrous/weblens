import { Divider, FileButton } from '@mantine/core'
import {
    IconArrowLeft,
    IconClock,
    IconFiles,
    IconFolder,
    IconFolderPlus,
    IconFolders,
    IconHome,
    IconPlus,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import { FileApi, FolderApi, GetFolderData } from '@weblens/api/FileBrowserApi'
import SharesApi from '@weblens/api/SharesApi'
import { useSubscribe as useFolderSubscribe } from '@weblens/api/Websocket'
import HeaderBar from '@weblens/components/HeaderBar'
import FilesErrorDisplay from '@weblens/components/NotFound'
import {
    PresentationContainer,
    PresentationFile,
} from '@weblens/components/Presentation'
import { useSessionStore } from '@weblens/components/UserInfo'
import {
    useResize,
    useResizeDrag,
    useWindowSize,
} from '@weblens/components/hooks'
import theme from '@weblens/components/theme.module.scss'
import Crumbs from '@weblens/lib/Crumbs'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { ButtonActionHandler } from '@weblens/lib/buttonTypes'
import fbStyle from '@weblens/pages/FileBrowser/style/fileBrowserStyle.module.scss'
import { ErrorHandler } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { WeblensFile } from '@weblens/types/files/File'
import FileColumns from '@weblens/types/files/FileColumns'
import { goToFile } from '@weblens/types/files/FileDragLogic'
import FileGrid from '@weblens/types/files/FileGrid'
import { FileContextMenu } from '@weblens/types/files/FileMenu'
import { FileRows } from '@weblens/types/files/FileRows'
import filesStyle from '@weblens/types/files/filesStyle.module.scss'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import { humanFileSize } from '@weblens/util'
import {
    ReactElement,
    memo,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { DraggingCounter, TransferCard } from './DropSpot'
import { FbModeT, useFileBrowserStore } from './FBStateControl'
import {
    HandleUploadButton,
    getRealId,
    historyDate,
    uploadViaUrl,
    useKeyDownFileBrowser,
    usePaste,
} from './FileBrowserLogic'
import { DirViewWrapper, WebsocketStatus } from './FileBrowserMiscComponents'
import { DirViewModeT } from './FileBrowserTypes'
import FileInfoPane from './FileInfoPane'
import FileSortBox from './FileSortBox'
import SearchDialogue from './SearchDialogue'
import { TasksDisplay } from './TaskProgress'
import UploadStatus from './UploadStatus'
import dirViewHeaderStyles from './style/dirViewHeader.module.scss'

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
                            uploadViaUrl(pasteImage, contentId, filesMap)
                                .then(() => setPasteImgBytes(null))
                                .catch(ErrorHandler)
                        }}
                    />
                </div>
            </div>
        </PresentationContainer>
    )
}

function TrashSize() {
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const trashSize = useFileBrowserStore((state) => state.trashDirSize)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const user = useSessionStore((state) => state.user)
    if (trashSize <= 0) {
        return null
    }
    const [trashSizeValue, trashSizeUnit] = humanFileSize(trashSize)

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

const SIDEBAR_BREAKPOINT = 650
const SIDEBAR_DEFAULT_WIDTH = 300
const SIDEBAR_MIN_OPEN_WIDTH = 200

function GlobalActions() {
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
                            Right={TrashSize}
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

const UsageInfo = () => {
    const [box, setBox] = useState<HTMLDivElement>(null)
    const size = useResize(box)

    const user = useSessionStore((state) => state.user)

    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    let homeSize = useFileBrowserStore((state) => state.homeDirSize)
    const trashSize = useFileBrowserStore((state) => state.trashDirSize)
    const selected = useFileBrowserStore((state) => state.selected)
    const filesMap = useFileBrowserStore((state) => state.filesMap)

    const selectedLength = selected.size

    const mode = useFileBrowserStore((state) => state.fbMode)

    const { selectedFileSize, selectedFolderCount, selectedFileCount } =
        useMemo(() => {
            let selectedFileSize = 0
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
                selectedFileSize += f.size || 0
            })

            return { selectedFileSize, selectedFolderCount, selectedFileCount }
        }, [selectedLength])

    if (
        folderInfo?.Id() === 'shared' ||
        user === null ||
        homeSize === -1 ||
        trashSize === -1
    ) {
        return null
    }

    let displaySize = folderInfo?.GetSize() || 0

    if (folderInfo?.Id() !== user.trashId) {
        homeSize = homeSize - trashSize
    }

    if (homeSize < displaySize) {
        displaySize = homeSize
    }

    const doGlobalSize = selectedLength === 0 && mode !== FbModeT.share

    let usagePercent = doGlobalSize
        ? (displaySize / homeSize) * 100
        : (selectedFileSize / displaySize) * 100
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
            {miniMode && <StartIcon className={theme['background-icon']} />}
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
                        {<StartIcon className={theme['background-icon']} />}
                        <p
                            className="select-none p-1"
                            style={{
                                display: miniMode ? 'none' : 'block',
                            }}
                        >
                            {doGlobalSize
                                ? humanFileSize(displaySize)
                                : humanFileSize(selectedFileSize)}
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
                    {<EndIcon className={theme['background-icon']} />}
                </div>
            </div>
            <div
                className="flex flex-row h-max justify-between items-center w-full bg-[--wl-barely-visible] rounded-lg p-2"
                style={{
                    display: selectedLength > 0 ? 'flex' : 'none',
                    flexDirection: miniMode ? 'column' : 'row',
                }}
            >
                <div className="flex h-max items-center text-[--wl-text-color] w-min">
                    <IconFiles />
                    <p className="select-none p-1">{selectedFileCount}</p>
                </div>
                <div className="flex h-max items-center text-[--wl-text-color] w-min">
                    <IconFolders />
                    <p className="select-none p-1">{selectedFolderCount}</p>
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

function DirViewHeader() {
    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const viewingPast = useFileBrowserStore((state) => state.pastTime)
    const selected = useFileBrowserStore((state) => state.selected)

    const setPastTime = useFileBrowserStore((state) => state.setPastTime)
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
        <div className="flex flex-col h-max">
            <div className={dirViewHeaderStyles['dir-view-header-wrapper']}>
                {(mode === FbModeT.default || mode === FbModeT.share) && (
                    <Crumbs navOnLast={false} moveSelectedTo={moveSelectedTo} />
                )}
                {(mode === FbModeT.share || viewingFolder) && <FileSortBox />}
            </div>
            {viewingPast && (
                <div
                    className={fbStyle['past-time-box']}
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
        <div className="flex flex-col h-full" ref={setFullViewRef}>
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
                    setDragging(DraggingStateT.ExternalDrag)
                    setMoveDest(folderInfo?.Id())
                }}
                onDragLeave={(e) => {
                    e.stopPropagation()
                    if (dragBoxRef.contains(e.relatedTarget as Node)) {
                        return
                    }
                    setDragging(DraggingStateT.NoDrag)
                }}
            >
                <div className="flex flex-col w-full h-full min-w-[20vw]">
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
        filesMap,
        filesLists,
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
        clearSelected,
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
            SharesApi.getFileShare(shareId)
                .then((res) => {
                    nav(`/files/share/${shareId}/${res.data.fileId}`)
                })
                .catch(ErrorHandler)
        } else {
            contentId = getRealId(contentId, mode, user)
            setLocationState({ contentId, mode, shareId, pastTime, jumpTo })
            removeLoading('files')
        }
    }, [urlPath, user, query('at'), jumpTo])

    const { readyState } = useFolderSubscribe()

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

            if (viewOpts.dirViewMode !== DirViewModeT.Columns) {
                clearSelected()
            }

            const folder = filesMap.get(contentId)
            if (
                folder &&
                (folder.GetFetching() ||
                    (folder.modifiable !== undefined &&
                        folder.childrenIds &&
                        folder.childrenIds.filter((f) => f !== user.trashId)
                            .length === filesLists.get(folder.Id())?.length))
            ) {
                console.debug('Exiting sync state early')
                if (folder.Id() !== folderInfo.Id()) {
                    setFilesData({
                        selfInfo: folder,
                    })
                }
                return
            }

            folder?.SetFetching(true)

            const fileData = await GetFolderData(
                contentId,
                fbMode,
                shareId,
                pastTime
            ).catch((r: number) => {
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

                console.debug('Setting main files data', fileData)
                setFilesData({
                    selfInfo: fileData.self,
                    childrenInfo: fileData.children,
                    parentsInfo: fileData.parents,
                    mediaData: fileData.medias,
                })

                folder?.SetFetching(false)
            }

            if (
                (jumpTo || viewOpts.dirViewMode === DirViewModeT.Columns) &&
                (fbMode !== FbModeT.share || contentId) &&
                useFileBrowserStore.getState().selected.size === 0
            ) {
                setSelected([jumpTo ? jumpTo : contentId], true)
            }
        }

        addLoading('files')
        syncState()
            .catch((e: number) => {
                console.error(e)
                setFilesFetchErr(e)
            })
            .finally(() => removeLoading('files'))
    }, [user, contentId, shareId, fbMode, pastTime, jumpTo])

    useEffect(() => {
        const selectedSize = useFileBrowserStore.getState().selected.size

        if (
            viewOpts.dirViewMode !== DirViewModeT.Columns &&
            selectedSize === 1
        ) {
            clearSelected()
        }

        if (
            (jumpTo || viewOpts.dirViewMode === DirViewModeT.Columns) &&
            (fbMode !== FbModeT.share || contentId) &&
            selectedSize === 0
        ) {
            setSelected([jumpTo ? jumpTo : contentId], true)
        }
    }, [viewOpts.dirViewMode])

    const searchVisitFunc = (loc: string) => {
        FileApi.getFile(loc)
            .then((f) => {
                if (!f.data) {
                    console.error('Could not find file to nav to')
                    return
                }

                goToFile(new WeblensFile(f.data), true)
            })
            .catch(ErrorHandler)
    }

    return (
        <div className="h-screen flex flex-col">
            <HeaderBar setBlockFocus={setBlockFocus} page={'files'} />
            <DraggingCounter />
            <PresentationFile file={filesMap.get(presentingId)} />
            {pasteImgBytes && <PasteImageDialogue />}
            {isSearching && (
                <div className="flex items-center justify-center w-screen h-screen absolute z-50 backdrop-blur-sm bg-[#00000088] px-[30%] py-[10%]">
                    <SearchDialogue text={''} visitFunc={searchVisitFunc} />
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
