import { Divider, FileButton, Text } from '@mantine/core'

// Icons
import {
    IconArrowLeft,
    IconFiles,
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconPlus,
    IconSearch,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import React, {
    memo,
    ReactElement,
    useCallback,
    useEffect,
    useReducer,
    useState,
} from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import {
    CreateFolder,
    GetFileInfo,
    getFileShare,
    GetFolderData,
    getPastFolderInfo,
    moveFiles,
    searchFolder,
} from '../../api/FileBrowserApi'
import { useSubscribe } from '../../api/Websocket'
import Crumbs from '../../components/Crumbs'
import FilesGrid from '../../Files/FileScroller'
import HeaderBar from '../../components/HeaderBar'
import { useResize, useResizeDrag, useWindowSize } from '../../components/hooks'
import NotFound from '../../components/NotFound'
import Presentation, {
    PresentationContainer,
    PresentationVisual,
} from '../../components/Presentation'
import WeblensButton from '../../components/WeblensButton'
import WeblensInput from '../../components/WeblensInput'
import { WeblensProgress } from '../../components/WeblensProgress'

// Weblens
import { WebsocketContext } from '../../Context'
import { WeblensFile, WeblensFileParams } from '../../Files/File'
import { FileContextMenu } from '../../Files/FileMenu'
import { FileRows } from '../../Files/FileRows'
import { DraggingStateT, TaskProgContext } from '../../Files/FBTypes'
import { AuthHeaderT } from '../../types/Types'
import { humanFileSize } from '../../util'

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
    GetStartedCard,
    PresentationFile,
    TransferCard,
    WebsocketStatus,
} from './FileBrowserMiscComponents'
import { FileInfoPane } from './FileInfoPane'

import FileSortBox from './FileSortBox'
import { StatTree } from './FileStatTree'
import {
    taskProgressReducer,
    TaskProgressState,
    TasksDisplay,
    TasksProgressAction,
    TasksProgressDispatch,
} from './TaskProgress'
import UploadStatus from './UploadStatus'
import './style/fileBrowserStyle.scss'
import '../../components/style.scss'
import { FbModeT, useFileBrowserStore } from './FBStateControl'
import { useShallow } from 'zustand/react/shallow'
import { useSessionStore } from '../../components/UserInfo'
import SearchDialogue from './SearchDialogue'
import { MediaImage } from '../../Media/PhotoContainer'
import WeblensMedia from '../../Media/Media'
import { useMediaStore } from '../../Media/MediaStateControl'

function PasteImageDialogue() {
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const pasteImage = useFileBrowserStore((state) => state.pasteImgBytes)

    const auth = useSessionStore((state) => state.auth)

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
                    <MediaImage media={media} quality="thumbnail" />
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
                            uploadViaUrl(
                                pasteImage,
                                contentId,
                                filesMap,
                                auth
                            ).then(() => setPasteImgBytes(null))
                        }}
                    />
                </div>
            </div>
        </PresentationContainer>
    )
}

const SIDEBAR_BREAKPOINT = 650

function GlobalActions() {
    const nav = useNavigate()
    const user = useSessionStore((state) => state.user)
    const authHeader = useSessionStore((state) => state.auth)
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
    const setPastTime = useFileBrowserStore((state) => state.setPastTime)

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

    const homeMouseUp = useCallback(
        (e) => {
            e.stopPropagation()
            setMoveDest('')
            if (draggingState !== DraggingStateT.NoDrag) {
                moveFiles(selectedIds, user.homeId, authHeader)
                setDragging(DraggingStateT.NoDrag)
            } else {
                setPastTime(null)
                nav('/files/home')
            }
        },
        [selectedIds, draggingState]
    )

    const trashMouseOver = useCallback(() => {
        if (draggingState !== DraggingStateT.NoDrag) {
            setMoveDest('Trash')
        }
    }, [draggingState])

    const trashMouseUp = useCallback(
        (e) => {
            e.stopPropagation()
            setMoveDest('')
            if (draggingState !== DraggingStateT.NoDrag) {
                moveFiles(selectedIds, user.trashId, authHeader)
                setDragging(DraggingStateT.NoDrag)
            } else {
                nav('/files/trash')
            }
        },
        [selectedIds, draggingState]
    )

    const [namingFolder, setNamingFolder] = useState(false)

    const navToShared = useCallback(() => {
        nav('/files/shared')
    }, [nav])

    const navToExternal = useCallback(
        (e) => {
            e.stopPropagation()
            nav('/files/external')
        },
        [nav]
    )

    const newFolder = useCallback(
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
                        squareSize={48}
                        toggleOn={
                            folderInfo?.Id() === user?.homeId &&
                            mode === FbModeT.default
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
                        squareSize={48}
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
                        <div className="relative w-full translate-y-1 z-20">
                            <div className="file-size-box">
                                <p>{`${trashSizeValue}${trashSizeUnit}`}</p>
                            </div>
                        </div>
                    )}
                    <WeblensButton
                        label="Trash"
                        fillWidth
                        squareSize={48}
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
                            squareSize={48}
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
                            squareSize={48}
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
                            height={48}
                            placeholder={'New Folder'}
                            buttonIcon={IconPlus}
                            closeInput={() => setNamingFolder(false)}
                            onComplete={(newName) =>
                                CreateFolder(
                                    folderInfo.Id(),
                                    newName,
                                    [],
                                    false,
                                    shareId,
                                    authHeader
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
                                '',
                                authHeader
                            )
                        }}
                        accept="file"
                        multiple
                    >
                        {(props) => {
                            return (
                                <WeblensButton
                                    label="Upload"
                                    squareSize={48}
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

                <Divider w={'100%'} my="lg" size={1.5} />

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

    let startIcon = doGlobalSize ? (
        <IconFolder size={20} />
    ) : (
        <IconFiles size={20} />
    )
    let endIcon = doGlobalSize ? (
        <IconHome size={20} />
    ) : (
        <IconFolder size={20} />
    )
    if (miniMode) {
        ;[startIcon, endIcon] = [endIcon, startIcon]
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
                    <div className="flex flex-row items-center h-full select-none font-semibold text-lg">
                        <p>Usage</p>
                        <div className="p-1" />
                        <p className=" text-ellipsis">
                            {usagePercent ? usagePercent.toFixed(2) : 0}%
                        </p>
                    </div>
                </div>
            )}
            {miniMode && startIcon}
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
                        {startIcon}
                        <Text
                            style={{
                                userSelect: 'none',
                                display: miniMode ? 'none' : 'block',
                            }}
                            size="14px"
                            pl={3}
                        >
                            {doGlobalSize
                                ? humanFileSize(displaySize)
                                : humanFileSize(selectedSize)}
                        </Text>
                    </div>
                )}
                <div className="flex flex-row justify-end w-max items-center">
                    <Text
                        style={{
                            userSelect: 'none',
                            display: miniMode ? 'none' : 'block',
                        }}
                        size="14px"
                        pr={3}
                    >
                        {doGlobalSize
                            ? humanFileSize(homeSize)
                            : humanFileSize(displaySize)}
                    </Text>
                    {endIcon}
                </div>
            </div>
        </div>
    )
}

function EmptySearch() {
    const nav = useNavigate()

    const filename = useFileBrowserStore((state) =>
        state.folderInfo.GetFilename()
    )
    const contentId = useFileBrowserStore((state) => state.contentId)
    const searchContent = useFileBrowserStore((state) => state.searchContent)

    return (
        <div className="flex flex-col items-center justify-end h-1/5">
            <div className="flex flex-row h-max w-max">
                <p className="text-lg">No items match your search in</p>
                <IconFolder style={{ marginLeft: 4 }} />
                <p className="text-lg">{filename}</p>
            </div>
            <div className="h-10" />
            <WeblensButton
                squareSize={40}
                centerContent
                Left={IconSearch}
                label={'Search all files'}
                onClick={() => {
                    nav(`/files/search/${contentId}?query=${searchContent}`)
                }}
            />
        </div>
    )
}

const SingleFile = memo(
    ({ file }: { file: WeblensFile }) => {
        // const [containerRef, setContainerRef] = useState<HTMLDivElement>()
        const mediaData = useMediaStore((state) =>
            state.mediaMap.get(file.GetMediaId())
        )

        if (!file.Id()) {
            return (
                <NotFound
                    resourceType="Share"
                    link="/files/home"
                    setNotFound={() => {}}
                />
            )
        }

        return (
            <div className="flex flex-row w-full h-full justify-around pb-2">
                <div
                    // ref={setContainerRef}
                    className="flex justify-center items-center p-6 h-full w-full"
                >
                    <PresentationVisual
                        mediaData={mediaData}
                        Element={() => PresentationFile({ file })}
                    />
                </div>
            </div>
        )
    },
    (prev, next) => {
        return prev.file === next.file
    }
)

function DirViewHeader({ moveSelected, searchQuery }) {
    const nav = useNavigate()
    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesCount = useFileBrowserStore((state) => state.filesList.length)
    const viewingPast = useFileBrowserStore((state) => state.viewingPast)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)

    return (
        <div className="flex flex-row h-max justify-between items-center p-2">
            {mode === FbModeT.search && (
                <div className="flex h-14 w-full items-center">
                    <WeblensButton
                        Left={IconArrowLeft}
                        onClick={() => nav(-1)}
                    />
                    <div className="w-4" />
                    <div
                        className="flex items-center p-1 pr-2 rounded bg-dark-paper
                                    outline outline-main-accent gap-2 m-2"
                    >
                        <IconSearch />
                        <p className="crumb-text">{searchQuery}</p>
                    </div>
                    <p className="crumb-text m-2">in</p>
                    <IconFolder size={36} />
                    <p className="crumb-text"> {folderInfo?.GetFilename()}</p>
                    <div className="w-2" />
                    <p className="text-gray-400 select-none">
                        {filesCount} results
                    </p>
                </div>
            )}
            {(mode === FbModeT.default || mode === FbModeT.share) && (
                <Crumbs
                    postText={
                        viewingPast
                            ? `@ ${historyDate(viewingPast.getTime())}`
                            : ''
                    }
                    navOnLast={false}
                    dragging={draggingState}
                    moveSelectedTo={moveSelected}
                    setMoveDest={setMoveDest}
                />
            )}
            {folderInfo?.IsFolder() && <FileSortBox />}
        </div>
    )
}

function DirView({
    notFound,
    setNotFound,
    searchQuery,
    authHeader,
}: {
    notFound: boolean
    setNotFound: (boolean) => void
    searchQuery: string
    searchFilter: string
    authHeader: AuthHeaderT
}) {
    const [contentViewRef, setContentViewRef] = useState(null)
    const [fullViewRef, setFullViewRef] = useState(null)

    const mode = useFileBrowserStore((state) => state.fbMode)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const selected = useFileBrowserStore((state) => state.selected)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const loading = useFileBrowserStore((state) => state.loading)
    const searchContent = useFileBrowserStore((state) => state.searchContent)
    const filesList = useFileBrowserStore((state) => state.filesList)
    const viewOpts = useFileBrowserStore((state) => state.viewOpts)
    const moveDest = useFileBrowserStore((state) => state.moveDest)
    const draggingState = useFileBrowserStore((state) => state.draggingState)

    const user = useSessionStore((state) => state.user)

    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const setDragging = useFileBrowserStore((state) => state.setDragging)

    const moveSelectedTo = useCallback(
        (folderId: string) => {
            MoveSelected(
                Array.from(selected.keys()),
                folderId,
                authHeader
            ).then(() => {
                clearSelected()
            })
        },
        [selected.size, contentId, authHeader]
    )

    if (loading.includes('files')) {
        return null
    }

    let fileDisplay: ReactElement
    if (notFound) {
        fileDisplay = (
            <NotFound
                resourceType="Folder"
                link="/files/home"
                setNotFound={setNotFound}
            />
        )
    } else if (
        (mode === FbModeT.default || mode === FbModeT.share) &&
        folderInfo?.Id() &&
        !folderInfo.IsFolder()
    ) {
        fileDisplay = <SingleFile file={folderInfo} />
    } else if (
        loading.length === 0 &&
        searchContent === '' &&
        filesMap.size === 0
    ) {
        fileDisplay = <GetStartedCard />
    } else if (
        filesList.length === 0 &&
        loading.length === 0 &&
        searchContent !== ''
    ) {
        fileDisplay = <EmptySearch />
    } else if (
        filesList.length === 0 &&
        loading.length === 0 &&
        mode === FbModeT.share
    ) {
        fileDisplay = (
            <NotFound
                resourceType="any files shared with you"
                link="/files/home"
                setNotFound={setNotFound}
            />
        )
    } else if (mode === FbModeT.stats) {
        fileDisplay = (
            <StatTree folderInfo={folderInfo} authHeader={authHeader} />
        )
    } else if (viewOpts.dirViewMode === 'List') {
        fileDisplay = <FileRows files={filesList} />
    } else if (viewOpts.dirViewMode === 'Grid') {
        fileDisplay = <FilesGrid files={filesList} />
    } else {
        console.error('Could not find valid directory view from state')
        return null
    }

    return (
        <div className="flex flex-col h-full" ref={setFullViewRef}>
            <DirViewHeader
                searchQuery={searchQuery}
                moveSelected={moveSelectedTo}
            />
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
                            filesList.map((value: WeblensFile) =>
                                value.GetFilename()
                            ),
                            false,
                            '',
                            authHeader
                        )
                        setDragging(DraggingStateT.NoDrag)
                    }}
                    stopDragging={() => setDragging(DraggingStateT.NoDrag)}
                    dropSpotTitle={folderInfo?.GetFilename()}
                    dragging={draggingState}
                    dropAllowed={folderInfo?.IsModifiable()}
                    handleDrag={(event) =>
                        handleDragOver(event, setDragging, draggingState)
                    }
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
        (s) => {
            const r = q.get(s)
            if (!r) {
                return ''
            }
            return r
        },
        [q]
    )
}

const FileBrowser = () => {
    const urlPath = useParams()['*']
    const query = useSearch()
    const searchQuery = query('query')
    const searchFilter = query('filter')
    const user = useSessionStore((state) => state.user)
    const authHeader = useSessionStore((state) => state.auth)
    const nav = useNavigate()

    const [notFound, setNotFound] = useState(false)

    const fbLocationContext = useFileBrowserStore(
        useShallow((state) => ({
            mode: state.fbMode,
            contentId: state.contentId,
            shareId: state.shareId,
        }))
    )
    const fbViewOpts = useFileBrowserStore((state) => state.viewOpts)
    const blockFocus = useFileBrowserStore((state) => state.blockFocus)
    const viewingPast = useFileBrowserStore((state) => state.viewingPast)
    const loading = useFileBrowserStore((state) => state.loading)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const presentingId = useFileBrowserStore((state) => state.presentingId)
    const isSearching = useFileBrowserStore((state) => state.isSearching)
    const pasteImage = useFileBrowserStore((state) => state.pasteImgBytes)

    const addLoading = useFileBrowserStore((state) => state.addLoading)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)
    const setLocation = useFileBrowserStore((state) => state.setLocationState)
    const clearFiles = useFileBrowserStore((state) => state.clearFiles)
    const setSearch = useFileBrowserStore((state) => state.setSearch)
    const setScrollTarget = useFileBrowserStore(
        (state) => state.setScrollTarget
    )
    const setSelected = useFileBrowserStore((state) => state.setSelected)
    const setFilesData = useFileBrowserStore((state) => state.setFilesData)
    const setBlockFocus = useFileBrowserStore((state) => state.setBlockFocus)
    const setPresentationTarget = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )

    const [taskProg, taskProgDispatch] = useReducer<
        (
            state: TaskProgressState,
            action: TasksProgressAction
        ) => TaskProgressState,
        TasksProgressDispatch
    >(taskProgressReducer, null, () => new TaskProgressState())

    useEffect(() => {
        localStorage.setItem('fbViewOpts', JSON.stringify(fbViewOpts))
    }, [fbViewOpts])

    useEffect(() => {
        if (!user) {
            return
        }
        addLoading('files')

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

        if (mode === FbModeT.share && shareId && !contentId) {
            getFileShare(shareId, authHeader).then((s) => {
                nav(`/files/share/${shareId}/${s.GetFileId()}`)
            })
        } else {
            getRealId(contentId, mode, user).then((contentId) => {
                setLocation(contentId, mode, shareId)
                removeLoading('files')
            })
        }
    }, [urlPath, authHeader, user])

    const { wsSend, readyState } = useSubscribe(
        fbLocationContext.contentId,
        fbLocationContext.shareId,
        user,
        taskProgDispatch,
        authHeader
    )

    useKeyDownFileBrowser()

    // Hook to handle uploading images from the clipboard
    usePaste(fbLocationContext.contentId, user, blockFocus)

    // Reset most of the state when we change folders
    const syncState = useCallback(async () => {
        if (!urlPath) {
            nav('/files/home', { replace: true })
        }

        if (urlPath === user?.homeId) {
            let redirect = '/files/home'
            const jumpItem = query('jumpTo')
            if (jumpItem) {
                redirect += `?jumpTo=${jumpItem}`
            }
            nav(redirect, { replace: true })
        }

        // If we're not ready, leave
        if (fbLocationContext.mode == FbModeT.unset || !user) {
            return
        }

        setNotFound(false)
        clearFiles()

        if (fbLocationContext.mode === FbModeT.search) {
            const folderData = await GetFileInfo(
                fbLocationContext.contentId,
                fbLocationContext.shareId,
                authHeader
            )

            if (!folderData) {
                console.error('No folder data')
                return
            }

            const searchResults = await searchFolder(
                fbLocationContext.contentId,
                searchQuery,
                searchFilter,
                authHeader
            )

            setSearch(searchQuery)
            setFilesData(folderData, searchResults, [], user)
            removeLoading('files')
            return
        }

        setSearch('')

        let fileData: {
            self?: WeblensFileParams
            children?: WeblensFileParams[]
            parents?: WeblensFileParams[]
            error?: string
        }
        if (viewingPast !== null) {
            fileData = await getPastFolderInfo(
                fbLocationContext.contentId,
                viewingPast,
                authHeader
            )
        } else {
            fileData = await GetFolderData(
                fbLocationContext.contentId,
                fbLocationContext.mode,
                fbLocationContext.shareId,
                authHeader
            ).catch((r) => {
                if (r === 400 || r === 404) {
                    setNotFound(true)
                } else {
                    console.error(r)
                }
            })
        }

        if (fileData) {
            setFilesData(
                fileData.self,
                fileData.children,
                fileData.parents,
                user
            )
        }

        const jumpItem = query('jumpTo')
        if (jumpItem) {
            setScrollTarget(jumpItem)
            setSelected([jumpItem])
        }
    }, [
        user,
        authHeader,
        fbLocationContext.contentId,
        fbLocationContext.shareId,
        fbLocationContext.mode,
        searchQuery,
        viewingPast,
    ])

    useEffect(() => {
        syncState().then(() => removeLoading('files'))
    }, [syncState])

    const presentingElement = useCallback(
        () =>
            PresentationFile({
                file: filesMap.get(presentingId),
            }),
        [filesMap, presentingId]
    )

    return (
        <WebsocketContext.Provider value={wsSend}>
            <TaskProgContext.Provider
                value={{ progState: taskProg, progDispatch: taskProgDispatch }}
            >
                <div className="h-screen flex flex-col">
                    <HeaderBar
                        setBlockFocus={setBlockFocus}
                        page={'files'}
                        loading={loading}
                    />
                    <DraggingCounter />
                    <Presentation
                        mediaId={filesMap.get(presentingId)?.GetMediaId()}
                        element={presentingElement}
                        dispatch={{ setPresentationTarget }}
                    />
                    {pasteImage && <PasteImageDialogue />}
                    {isSearching && <SearchDialogue />}
                    <FileContextMenu />
                    <div className="absolute bottom-1 left-1">
                        <WebsocketStatus ready={readyState} />
                    </div>
                    <div className="flex flex-row grow h-[90vh] items-start">
                        <GlobalActions />
                        <DirViewWrapper>
                            <DirView
                                notFound={notFound}
                                setNotFound={setNotFound}
                                searchQuery={searchQuery}
                                searchFilter={searchFilter}
                                authHeader={authHeader}
                            />
                        </DirViewWrapper>
                    </div>
                </div>
            </TaskProgContext.Provider>
        </WebsocketContext.Provider>
    )
}

export default FileBrowser
