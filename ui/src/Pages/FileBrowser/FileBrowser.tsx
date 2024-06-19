// React
import {
    createContext,
    memo,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useReducer,
    useRef,
    useState,
} from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

// Icons
import {
    IconArrowLeft,
    IconDownload,
    IconFile,
    IconFiles,
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconInfoCircle,
    IconPlus,
    IconSearch,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
    IconX,
} from '@tabler/icons-react'

// Mantine
import { Button, Divider, FileButton, Input, Space, Text } from '@mantine/core'
import { useDebouncedValue, useMouse } from '@mantine/hooks'

// Weblens
import { UserContext } from '../../Context'
import {
    CreateFolder,
    GetFileInfo,
    getFileShare,
    GetFolderData,
    getPastFolderInfo,
    moveFiles,
    searchFolder,
} from '../../api/FileBrowserApi'
import Crumbs from '../../components/Crumbs'
import HeaderBar from '../../components/HeaderBar'
import { GlobalContextType } from '../../Files/FileDisplay'
import FileScroller from '../../components/FileScroller'
import NotFound from '../../components/NotFound'
import { MediaImage } from '../../Media/PhotoContainer'
import Presentation, {
    PresentationContainer,
} from '../../components/Presentation'
import UploadStatus, { useUploadStatus } from './UploadStatus'
import './style/fileBrowserStyle.scss'
import '../../components/style.scss'
import {
    AuthHeaderT,
    FBDispatchT,
    FbStateT,
    FileBrowserAction,
    UserContextT,
} from '../../types/Types'
import WeblensMedia from '../../Media/Media'
import { FileInitT, WeblensFile } from '../../Files/File'
import { humanFileSize } from '../../util'

import {
    downloadSelected,
    fileBrowserReducer,
    getRealId,
    handleDragOver,
    HandleDrop,
    HandleUploadButton,
    MoveSelected,
    SetFileData,
    uploadViaUrl,
    useKeyDownFileBrowser,
    usePaste,
} from './FileBrowserLogic'

import { GetFilesContext, GetItemsList } from './FilesContext'
import {
    DirViewWrapper,
    DropSpot,
    FbMenuModeT,
    FileInfoDisplay,
    GetStartedCard,
    IconDisplay,
    PresentationFile,
    TransferCard,
    WebsocketStatus,
} from './FileBrowserStyles'
import { WeblensButton } from '../../components/WeblensButton'
import { FileRows } from '../../Files/FileRows'
import { useResize, useResizeDrag, useWindowSize } from '../../components/hooks'
import { WeblensProgress } from '../../components/WeblensProgress'
import { FilesPane } from './FileInfoPane'
import { StatTree } from './FileStatTree'
import FileSortBox from './FileSortBox'
import { FileContextMenu } from '../../Files/FileMenu'
import { useSubscribe } from '../../api/Websocket'
import { TasksDisplay } from './TaskProgress'
import WeblensInput from '../../components/WeblensInput'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

type FbContextT = {
    fbState: FbStateT
    fbDispatch: FBDispatchT
}

export const FbContext = createContext<FbContextT>({
    fbState: null,
    fbDispatch: null,
})

export enum DraggingState {
    NoDrag, // No dragging is taking place
    InternalDrag, // Dragging is of only internal elements

    // Dragging is from external source, such as
    // dragging files from your computer over the browser
    ExternalDrag,
}

function PasteImageDialogue({
    img,
    dirMap,
    folderId,
    authHeader,
    dispatch,
    wsSend,
}: {
    img: ArrayBuffer
    dirMap: Map<string, WeblensFile>
    folderId
    authHeader
    dispatch
    wsSend
}) {
    if (!img) {
        return <></>
    }
    const media = new WeblensMedia({ contentId: 'paste' })
    media.SetThumbnailBytes(img)

    return (
        <PresentationContainer
            onClick={() => dispatch({ type: 'paste_image', img: null })}
        >
            <div className="flex absolute justify-center items-center z-[2]">
                <Text fw={700} size="40px" style={{ paddingBottom: '50px' }}>
                    Upload from clipboard?
                </Text>
                <div
                    className="h-1/2 w-max bg-bottom-grey p-3 rounded-lg overflow-hidden"
                    onClick={(e) => {
                        e.stopPropagation()
                    }}
                >
                    <MediaImage media={media} quality="thumbnail" />
                </div>
                <div className="flex flex-row justify-between w-[300px] h-[150px]">
                    <Button
                        size="xl"
                        variant="default"
                        onClick={(e) => {
                            e.stopPropagation()
                            dispatch({ type: 'paste_image', img: null })
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        size="xl"
                        color="#4444ff"
                        onClick={(e) => {
                            e.stopPropagation()
                            uploadViaUrl(
                                img,
                                folderId,
                                dirMap,
                                authHeader,
                                dispatch,
                                wsSend
                            )
                        }}
                    >
                        Upload
                    </Button>
                </div>
            </div>
        </PresentationContainer>
    )
}

const SIDEBAR_BREAKPOINT = 650

const GlobalActions = memo(
    ({
        uploadState,
        wsSend,
        uploadDispatch,
    }: {
        uploadState
        wsSend: (action: string, content: any) => void
        uploadDispatch
    }) => {
        const nav = useNavigate()
        const { usr, authHeader }: UserContextT = useContext(UserContext)
        const { fbState, fbDispatch }: FbContextT = useContext(FbContext)
        const windowSize = useWindowSize()
        const [trashSize, trashUnits] = humanFileSize(fbState.trashDirSize)

        const [resizing, setResizing] = useState(false)
        const [resizeOffset, setResizeOffset] = useState(
            windowSize?.width > SIDEBAR_BREAKPOINT ? 300 : 75
        )
        useResizeDrag(resizing, setResizing, (s) => {
            setResizeOffset(Math.min(s > 200 ? s : 75, 600))
        })

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

        const homeMouseOver = useCallback(() => {
            if (fbState.draggingState !== DraggingState.NoDrag) {
                fbDispatch({
                    type: 'set_move_dest',
                    fileName: 'Home',
                })
            }
        }, [])

        const homeMouseLeave = useCallback(() => {
            if (fbState.draggingState !== DraggingState.NoDrag) {
                fbDispatch({
                    type: 'set_move_dest',
                    fileName: '',
                })
            }
        }, [])

        const homeMouseUp = useCallback((e) => {
            e.stopPropagation()
            fbDispatch({
                type: 'set_move_dest',
                fileName: '',
            })
            if (fbState.draggingState !== DraggingState.NoDrag) {
                moveFiles(
                    Array.from(fbState.selected.keys()),
                    usr.homeId,
                    authHeader
                )
                fbDispatch({
                    type: 'set_dragging',
                    dragging: DraggingState.NoDrag,
                })
            } else {
                nav('/files/home')
            }
        }, [])

        const trashMouseOver = useCallback(() => {
            if (fbState.draggingState !== DraggingState.NoDrag) {
                fbDispatch({
                    type: 'set_move_dest',
                    fileName: 'Trash',
                })
            }
        }, [])

        const trashMouseLeave = useCallback(() => {
            if (fbState.draggingState !== DraggingState.NoDrag) {
                fbDispatch({
                    type: 'set_move_dest',
                    fileName: '',
                })
            }
        }, [])

        const trashMouseUp = useCallback((e) => {
            e.stopPropagation()
            fbDispatch({
                type: 'set_move_dest',
                fileName: '',
            })
            if (fbState.draggingState !== DraggingState.NoDrag) {
                moveFiles(
                    Array.from(fbState.selected.keys()),
                    usr.trashId,
                    authHeader
                )
                fbDispatch({
                    type: 'set_dragging',
                    dragging: DraggingState.NoDrag,
                })
            } else {
                nav('/files/trash')
            }
        }, [])

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
                    <div className="flex flex-col w-full gap-1">
                        <WeblensButton
                            label="Home"
                            fillWidth
                            squareSize={48}
                            toggleOn={
                                fbState.folderInfo.Id() === usr?.homeId &&
                                fbState.fbMode === FbModeT.default
                            }
                            disabled={!usr.isLoggedIn}
                            allowRepeat={false}
                            Left={IconHome}
                            onMouseOver={homeMouseOver}
                            onMouseLeave={homeMouseLeave}
                            onMouseUp={homeMouseUp}
                        />

                        <WeblensButton
                            label="Shared"
                            fillWidth
                            squareSize={48}
                            toggleOn={
                                fbState.fbMode === FbModeT.share &&
                                fbState.shareId === ''
                            }
                            disabled={
                                fbState.draggingState !==
                                    DraggingState.NoDrag || !usr.isLoggedIn
                            }
                            allowRepeat={false}
                            Left={IconUsers}
                            onClick={navToShared}
                        />

                        <WeblensButton
                            label="Trash"
                            fillWidth
                            squareSize={48}
                            toggleOn={
                                fbState.folderInfo.Id() === usr?.trashId &&
                                fbState.fbMode === FbModeT.default
                            }
                            disabled={
                                (fbState.draggingState !==
                                    DraggingState.NoDrag &&
                                    fbState.folderInfo.Id() === usr?.trashId &&
                                    fbState.fbMode === FbModeT.default) ||
                                !usr.isLoggedIn
                            }
                            allowRepeat={false}
                            Left={IconTrash}
                            postScript={
                                trashSize && resizeOffset >= 150
                                    ? `${trashSize}${trashUnits}`
                                    : ''
                            }
                            onMouseOver={trashMouseOver}
                            onMouseLeave={trashMouseLeave}
                            onMouseUp={trashMouseUp}
                        />

                        <div className="p-1" />

                        {usr.admin && (
                            <WeblensButton
                                label="External"
                                fillWidth
                                squareSize={48}
                                toggleOn={fbState.fbMode === FbModeT.external}
                                allowRepeat={false}
                                Left={IconServer}
                                disabled={
                                    fbState.draggingState !==
                                    DraggingState.NoDrag
                                }
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
                                    fbState.draggingState !== 0 ||
                                    !fbState.folderInfo.IsModifiable()
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
                                onComplete={(newName) => {
                                    CreateFolder(
                                        fbState.folderInfo.Id(),
                                        newName,
                                        [],
                                        false,
                                        fbState.shareId,
                                        authHeader
                                    )
                                        .then(() => setNamingFolder(false))
                                        .catch((r) => {
                                            console.error(r)
                                        })
                                }}
                            />
                        )}

                        <FileButton
                            onChange={(files) => {
                                HandleUploadButton(
                                    files,
                                    fbState.folderInfo.Id(),
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
                                    <WeblensButton
                                        label="Upload"
                                        squareSize={48}
                                        fillWidth
                                        showSuccess={false}
                                        disabled={
                                            fbState.draggingState !== 0 ||
                                            !fbState.folderInfo.IsModifiable()
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

                    <TasksDisplay scanProgress={fbState.scanProgress} />

                    <div className="flex grow" />

                    <UploadStatus
                        uploadState={uploadState}
                        uploadDispatch={uploadDispatch}
                    />
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
    },
    (p, n) => {
        return p.uploadState === n.uploadState
    }
)

const UsageInfo = ({}) => {
    const [box, setBox] = useState(null)
    const size = useResize(box)

    const { usr } = useContext(UserContext)
    const { fbState } = useContext(FbContext)

    if (fbState.folderInfo.Id() === 'shared') {
        return null
    }

    let displaySize = fbState.folderInfo.GetSize()
    if (fbState.folderInfo.Id() === usr.homeId) {
        displaySize = displaySize - fbState.trashDirSize
    }

    const selected = Array.from(fbState.selected.keys()).map((v) =>
        fbState.dirMap.get(v)
    )

    const selectedSize = selected.reduce((acc: number, x: WeblensFile) => {
        return acc + (x ? x.GetSize() : 0)
    }, 0)

    if (fbState.homeDirSize < displaySize) {
        displaySize = fbState.homeDirSize
    }

    let usagePercent =
        selected.length === 0
            ? (displaySize / fbState.homeDirSize) * 100
            : (selectedSize / displaySize) * 100
    if (!usagePercent) {
        usagePercent = 0
    }

    const miniMode = size.width < 100

    let startIcon =
        selected.length === 0 ? (
            <IconFolder size={20} />
        ) : (
            <IconFiles size={20} />
        )
    let endIcon =
        selected.length === 0 ? (
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
            <WeblensProgress
                key={miniMode ? 'usage-vertical' : 'usage-horizontal'}
                value={usagePercent}
                orientation={miniMode ? 'vertical' : 'horizontal'}
                style={{
                    height: miniMode ? 100 : 20,
                    width: miniMode ? 20 : '100%',
                }}
            />
            <div
                className="flex flex-row h-max justify-between items-center"
                style={{
                    width: miniMode ? 'max-content' : '98%',
                }}
            >
                {fbState.folderInfo.Id() !== 'shared' && !miniMode && (
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
                            {selected.length === 0
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
                        {selected.length === 0
                            ? humanFileSize(fbState.homeDirSize)
                            : humanFileSize(displaySize)}
                    </Text>
                    {endIcon}
                </div>
            </div>
        </div>
    )
}

function FileSearch({
    fb,
    defaultOpen = false,
    searchRef,
    dispatch,
}: {
    fb: FbStateT
    defaultOpen?: boolean
    searchRef: any
    dispatch: FBDispatchT
}) {
    const [searchOpen, setSearchOpen] = useState(defaultOpen)
    const [hintOpen, setHintOpen] = useState(false)
    const [error, setError] = useState(false)
    const nav = useNavigate()

    useEffect(() => {
        if (Boolean(fb.searchContent) && !searchOpen) {
            setSearchOpen(true)
        }
    }, [searchOpen, fb.searchContent])

    useEffect(() => {
        if (fb.fbMode !== FbModeT.search) {
            setHintOpen(false)
            setSearchOpen(false)
        }
    }, [fb.fbMode])

    useEffect(() => {
        if (
            !Boolean(fb.searchContent) ||
            document.activeElement !== searchRef.current
        ) {
            setHintOpen(false)
            return
        }
        try {
            new RegExp(fb.searchContent)
            setError(false)
            setHintOpen(true)
        } catch {
            setHintOpen(false)
        }
    }, [setHintOpen, fb.searchContent])

    return (
        <div
            style={{
                height: 'max-content',
                width: 'max-content',
                alignItems: 'flex-start',
                marginRight: 5,
            }}
        >
            <div className="search-box">
                <IconSearch
                    color="white"
                    className="search-icon"
                    onClick={() => {
                        setSearchOpen(true)
                        searchRef.current.focus()
                    }}
                />
                <Input
                    mod={{ 'data-open': 'false' }}
                    onBlur={() => {
                        if (fb.searchContent === '') {
                            setSearchOpen(false)
                            setHintOpen(false)
                            searchRef.current.blur()
                        } else if (hintOpen) {
                            setHintOpen(false)
                        }
                    }}
                    onFocus={() => {
                        if (fb.searchContent === '') {
                            return
                        }
                        try {
                            new RegExp(fb.searchContent)
                            setError(false)
                            setHintOpen(true)
                        } catch {
                            setHintOpen(false)
                        }
                    }}
                    classNames={{
                        input: `search-input search-input-${
                            searchOpen ? 'open' : 'closed'
                        }`,
                    }}
                    unstyled
                    value={fb.searchContent}
                    ref={searchRef}
                    onChange={(e) =>
                        dispatch({
                            type: 'set_search',
                            search: e.target.value,
                        })
                    }
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' && !hintOpen) {
                            e.stopPropagation()
                            if (!Boolean(fb.searchContent)) {
                                nav(`/files/${fb.contentId}`)
                                return
                            }
                            setError(true)
                            setTimeout(() => setError(false), 2000)
                        }
                    }}
                />
            </div>
            {hintOpen && (
                <div className="search-hint-box">
                    <div className="key-line">
                        <Text>Press</Text>
                        <Text className="key-display">Enter</Text>
                        <Text>to search all files</Text>
                    </div>
                </div>
            )}
            {error && (
                <div className="search-hint-box" data-error={true}>
                    <Text>Not valid regex</Text>
                </div>
            )}
        </div>
    )
}

function DraggingCounter({ dragging, dirMap, selected, dispatch }) {
    const position = useMouse()
    const selectedKeys = Array.from(selected.keys())
    const { files, folders } = useMemo(() => {
        let files = 0
        let folders = 0

        selectedKeys.forEach((k: string) => {
            if (dirMap.get(k)?.IsFolder()) {
                folders++
            } else {
                files++
            }
        })
        return { files, folders }
    }, [JSON.stringify(selectedKeys)])

    if (dragging !== 1) {
        return null
    }

    return (
        <div
            className="fixed z-10"
            style={{
                top: position.y + 8,
                left: position.x + 8,
            }}
            onMouseUp={() => {
                dispatch({ type: 'set_dragging', dragging: false })
            }}
        >
            {Boolean(files) && (
                <div className="flex flex-row h-max">
                    <IconFile size={30} />
                    <Space w={10} />
                    <p>{files}</p>
                </div>
            )}
            {Boolean(folders) && (
                <div className="flex flex-row h-max">
                    <IconFolder size={30} />
                    <Space w={10} />
                    <p>{folders}</p>
                </div>
            )}
        </div>
    )
}

function SingleFile({
    file,
    doDownload,
}: {
    file: WeblensFile
    doDownload: (file: WeblensFile) => void
}) {
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
            <div className="icon-display-wrapper">
                <IconDisplay file={file} allowMedia size={'65%'} />
            </div>
            <div className="grow max-w-1/2 p-3">
                <FileInfoDisplay file={file} />
                <div className="min-h-[40%]">
                    <div
                        className="flex flex-row"
                        onClick={() => doDownload(file)}
                        style={{
                            backgroundColor: '#4444ff',
                            borderRadius: 4,
                            padding: 10,
                            height: 'max-content',
                            cursor: 'pointer',
                        }}
                    >
                        <IconDownload />
                        <Text c="white" style={{ paddingLeft: 10 }}>
                            Download {file.GetFilename()}
                        </Text>
                    </div>
                </div>
            </div>
        </div>
    )
}

function Files({
    fbState,
    notFound,
    setNotFound,
    searchRef,
    dispatch,
    wsSend,
    uploadDispatch,
    authHeader,
}: {
    fbState: FbStateT
    notFound: boolean
    setNotFound: (f: boolean) => void
    searchRef
    dispatch: (action: FileBrowserAction) => void
    wsSend: (action: string, content: any) => void
    uploadDispatch
    authHeader: AuthHeaderT
}) {
    const { usr }: UserContextT = useContext(UserContext)
    const nav = useNavigate()
    const [debouncedSearch] = useDebouncedValue(fbState.searchContent, 200)

    const [fullViewRef, setFullViewRef] = useState(null)
    useResize(fullViewRef)
    const [contentViewRef, setContentViewRef] = useState(null)

    const moveSelectedTo = useCallback(
        (folderId: string) => {
            MoveSelected(fbState.selected, folderId, authHeader).then(() => {
                dispatch({ type: 'clear_selected' })
            })
        },
        [fbState.selected.size, fbState.contentId, authHeader]
    )

    const { files, hoveringIndex, lastSelectedIndex } = useMemo(() => {
        return GetItemsList(fbState, usr, debouncedSearch)
    }, [
        fbState.dirMap,
        fbState.holdingShift,
        fbState.selected,
        fbState.hovering,
        debouncedSearch,
        usr,
        fbState.lastSelected,
        fbState.sortFunc,
        fbState.sortDirection,
    ])

    useEffect(() => {
        dispatch({ type: 'add_loading', loading: 'fileSearch' })
    }, [debouncedSearch])

    useEffect(() => {
        const fileIds = files.map((v) => v.file.Id())
        dispatch({
            type: 'set_files_list',
            fileIds: fileIds,
        })
        dispatch({ type: 'remove_loading', loading: 'fileSearch' })
    }, [files, dispatch])

    const selectedInfo = useMemo(() => {
        return Array.from(fbState.selected.keys()).map((fId) =>
            fbState.dirMap.get(fId)
        )
    }, [fbState.selected])

    const itemsCtx: GlobalContextType = useMemo(() => {
        return GetFilesContext(
            fbState,
            files,
            hoveringIndex,
            lastSelectedIndex,
            authHeader,
            dispatch
        )
    }, [
        files,
        fbState.contentId,
        fbState.dirMap,
        fbState.selected,
        fbState.fbMode,
        fbState.draggingState,
        fbState.hovering,
        fbState.holdingShift,
        dispatch,
    ])
    const dropAllowed = useMemo(() => {
        return (
            fbState.folderInfo.IsModifiable() &&
            !(
                fbState.fbMode === FbModeT.share ||
                fbState.contentId === usr.trashId
            )
        )
    }, [fbState.contentId, usr.trashId, fbState.fbMode, fbState.folderInfo])

    const openInfoPane = useCallback(
        () =>
            dispatch({
                type: 'set_file_info_menu',
                open: !fbState.fileInfoMenu,
            }),
        [fbState.fileInfoMenu]
    )

    if (notFound) {
        return (
            <NotFound
                resourceType="Folder"
                link="/files/home"
                setNotFound={setNotFound}
            />
        )
    }

    return (
        <div ref={setFullViewRef} className="flex flex-col w-full h-full pl-3">
            <TransferCard
                action="Move"
                destination={fbState.moveDest}
                boundRef={fullViewRef}
            />
            <div className="flex flex-row h-max justify-between p-2">
                <Crumbs
                    postText={
                        fbState.viewingPast
                            ? `@ ${fbState.viewingPast.toDateString()} ${fbState.viewingPast.toLocaleTimeString()}`
                            : ''
                    }
                    navOnLast={false}
                    dragging={fbState.draggingState}
                    moveSelectedTo={moveSelectedTo}
                    setMoveDest={(itemName) =>
                        dispatch({ type: 'set_move_dest', fileName: itemName })
                    }
                />

                <FileSearch
                    fb={fbState}
                    searchRef={searchRef}
                    dispatch={dispatch}
                />
                <FileSortBox />
                <WeblensButton
                    Left={fbState.fileInfoMenu ? IconX : IconInfoCircle}
                    subtle
                    squareSize={42}
                    onClick={openInfoPane}
                />
            </div>
            <div className="flex flex-row h-[200px] grow max-w-full">
                <div className="grow shrink w-0" ref={setContentViewRef}>
                    <DropSpot
                        onDrop={(e) => {
                            HandleDrop(
                                e.dataTransfer.items,
                                fbState.contentId,
                                Array.from(fbState.dirMap.values()).map(
                                    (value: WeblensFile) => value.GetFilename()
                                ),
                                false,
                                '',
                                authHeader,
                                uploadDispatch,
                                wsSend
                            )
                            dispatch({
                                type: 'set_dragging',
                                dragging: DraggingState.NoDrag,
                            })
                        }}
                        dropSpotTitle={fbState.folderInfo.GetFilename()}
                        dragging={fbState.draggingState}
                        dropAllowed={dropAllowed}
                        handleDrag={(event) =>
                            handleDragOver(
                                event,
                                dispatch,
                                fbState.draggingState
                            )
                        }
                        wrapperRef={contentViewRef}
                    />
                    {(files.length !== 0 && (
                        <FileScroller
                            itemsContext={files}
                            globalContext={itemsCtx}
                            parentNode={contentViewRef}
                            dispatch={dispatch}
                        />
                    )) ||
                        (fbState.loading.length === 0 &&
                            fbState.searchContent === '' &&
                            fbState.searchContent === debouncedSearch && (
                                <GetStartedCard
                                    fb={fbState}
                                    dispatch={dispatch}
                                    uploadDispatch={uploadDispatch}
                                    wsSend={wsSend}
                                />
                            )) ||
                        (fbState.loading.length === 0 &&
                            fbState.folderInfo.Id() === 'shared' && (
                                <NotFound
                                    resourceType="any files shared with you"
                                    link="/files/home"
                                    setNotFound={setNotFound}
                                />
                            )) ||
                        (fbState.loading.length === 0 &&
                            fbState.searchContent !== '' && (
                                <div className="flex flex-col items-center justify-end h-1/5">
                                    <div className="flex flex-row h-max w-max">
                                        <Text size="20px">
                                            No items match your search in
                                        </Text>
                                        <IconFolder style={{ marginLeft: 4 }} />
                                        <Text size="20px">
                                            {fbState.folderInfo.GetFilename()}
                                        </Text>
                                    </div>
                                    <div className="h-10" />
                                    <WeblensButton
                                        squareSize={40}
                                        centerContent
                                        Left={IconSearch}
                                        label={'Search all files'}
                                        onClick={() => {
                                            nav(
                                                `/files/search/${fbState.contentId}?query=${fbState.searchContent}`
                                            )
                                        }}
                                    />
                                </div>
                            ))}
                </div>
                <FilesPane
                    open={fbState.fileInfoMenu}
                    setOpen={(o) =>
                        dispatch({ type: 'set_file_info_menu', open: o })
                    }
                    selectedFiles={selectedInfo}
                    timestamp={fbState.viewingPast?.getTime()}
                    contentId={fbState.contentId}
                    dispatch={dispatch}
                />
            </div>
        </div>
    )
}

function SearchResults({
    searchQuery,
    filter,
    searchRef,
}: {
    searchQuery: string
    filter: string
    searchRef
}) {
    const nav = useNavigate()
    const { fbState, fbDispatch } = useContext(FbContext)
    let titleString: string = 'Searching '
    if (searchQuery) {
        titleString += `for ${searchQuery}`
    } else {
        titleString += `all files`
    }
    titleString += ` in ${fbState.folderInfo.GetFilename()}`
    if (filter) {
        titleString += ` ending with .${filter}`
    }

    return (
        <div className="flex flex-col h-full">
            <div className="flex h-14 flex-shrink-0 w-full items-center">
                <IconArrowLeft
                    style={{ width: 40, height: 32, cursor: 'pointer' }}
                    onClick={() => nav(-1)}
                />
                <Text className="crumb-text">{titleString}</Text>

                <Text
                    className="crumb-text"
                    c="#aaaaaa"
                    style={{ fontSize: '14px', marginLeft: 10 }}
                >
                    {fbState.dirMap.size} results
                </Text>
                <Space flex={1} />
                <FileSearch
                    fb={fbState}
                    defaultOpen
                    searchRef={searchRef}
                    dispatch={fbDispatch}
                />
            </div>
            <Space h={10} />
            <FileRows fb={fbState} dispatch={fbDispatch} />
        </div>
    )
}

function DirView({
    fb: fb,
    notFound,
    setNotFound,
    searchRef,
    searchQuery,
    searchFilter,
    dispatch,
    wsSend,
    uploadDispatch,
    authHeader,
}: {
    fb: FbStateT
    notFound: boolean
    setNotFound
    searchRef
    searchQuery: string
    searchFilter: string
    dispatch: FBDispatchT
    wsSend: (action: string, content: any) => void
    uploadDispatch
    authHeader: AuthHeaderT
}) {
    const download = useCallback(
        (file: WeblensFile) =>
            downloadSelected([file], dispatch, wsSend, authHeader, fb.shareId),
        [authHeader, wsSend, dispatch, fb.fbMode, fb.contentId, fb.shareId]
    )

    if (fb.loading.includes('files')) {
        return null
    }

    if (
        fb.fbMode === FbModeT.default &&
        fb.folderInfo.Id() &&
        !fb.folderInfo.IsFolder()
    ) {
        return <SingleFile file={fb.folderInfo} doDownload={download} />
    } else if (fb.fbMode === FbModeT.stats) {
        return <StatTree folderInfo={fb.folderInfo} authHeader={authHeader} />
    } else if (fb.fbMode === FbModeT.search) {
        return (
            <SearchResults
                searchQuery={searchQuery}
                filter={searchFilter}
                searchRef={searchRef}
            />
        )
    } else {
        return (
            <Files
                fbState={fb}
                notFound={notFound}
                setNotFound={setNotFound}
                searchRef={searchRef}
                dispatch={dispatch}
                wsSend={wsSend}
                uploadDispatch={uploadDispatch}
                authHeader={authHeader}
            />
        )
    }
}

function useQuery() {
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

export enum FbModeT {
    unset,
    default,
    share,
    external,
    stats,
    search,
}

const FileBrowser = () => {
    const urlPath = useParams()['*']
    const query = useQuery()
    const searchQuery = query('query')
    const searchFilter = query('filter')
    const nav = useNavigate()
    const { authHeader, usr }: UserContextT = useContext(UserContext)

    const searchRef = useRef()

    const [notFound, setNotFound] = useState(false)
    const { uploadState, uploadDispatch } = useUploadStatus()

    const [fb, dispatch]: [FbStateT, (action: FileBrowserAction) => void] =
        useReducer(fileBrowserReducer, {
            folderInfo: new WeblensFile({ isDir: true }),
            uploadMap: new Map<string, boolean>(),
            selected: new Map<string, boolean>(),
            dirMap: new Map<string, WeblensFile>(),
            menuMode: FbMenuModeT.Closed,
            menuPos: { x: 0, y: 0 },
            holdingShift: false,
            fileInfoMenu: false,
            blockFocus: false,
            viewingPast: null,
            searchContent: '',
            lastSelected: '',
            draggingState: 0,
            scanProgress: [],
            menuTargetId: '',
            presentingId: '',
            sortDirection: 1,
            sortFunc: 'Name',
            trashDirSize: 0,
            pasteImg: null,
            homeDirSize: 0,
            contentId: '',
            filesList: [],
            hovering: '',
            scrollTo: '',
            moveDest: '',
            shareId: '',
            parents: [],
            loading: [],
            numCols: 0,
            fbMode: FbModeT.unset,
        })

    if (fb.fbMode && fb.fbMode !== FbModeT.share && usr.isLoggedIn === false) {
        nav('/login')
    }

    useEffect(() => {
        if (usr.isLoggedIn === undefined) {
            return
        }
        dispatch({ type: 'add_loading', loading: 'files' })

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
            getRealId(contentId, mode, usr, authHeader).then((realId) => {
                dispatch({
                    type: 'set_location_state',
                    realId: realId,
                    mode: mode,
                    shareId: shareId,
                })
                dispatch({ type: 'remove_loading', loading: 'files' })
            })
        }
    }, [urlPath, dispatch, authHeader, usr])

    const { wsSend, readyState } = useSubscribe(
        fb.contentId,
        fb.shareId,
        fb.fbMode,
        usr,
        dispatch,
        authHeader
    )

    useKeyDownFileBrowser(
        fb,
        searchQuery,
        usr,
        dispatch,
        authHeader,
        wsSend,
        searchRef
    )

    // Hook to handle uploading images from the clipboard
    usePaste(fb.contentId, usr, searchRef, fb.blockFocus, dispatch)

    // Reset most of the state when we change folders
    const syncState = useCallback(async () => {
        if (!urlPath) {
            nav('/files/home', { replace: true })
        }

        if (urlPath === usr?.homeId) {
            let redirect = '/files/home'
            const jumpItem = query('jumpTo')
            if (jumpItem) {
                redirect += `?jumpTo=${jumpItem}`
            }
            nav(redirect, { replace: true })
        }

        // If we're not ready, leave
        if (!fb.fbMode || usr.isLoggedIn === undefined) {
            return
        }

        setNotFound(false)
        dispatch({ type: 'clear_files' })

        if (fb.fbMode === FbModeT.search) {
            const folderData = await GetFileInfo(
                fb.contentId,
                fb.shareId,
                authHeader
            )
            if (!folderData) {
                console.error('No folder data')
                return
            }

            const searchResults = await searchFolder(
                fb.contentId,
                searchQuery,
                searchFilter,
                authHeader
            )

            dispatch({ type: 'set_search', search: searchQuery })
            SetFileData(
                { children: searchResults, self: folderData },
                dispatch,
                usr
            )
            dispatch({ type: 'remove_loading', loading: 'files' })
            return
        }

        dispatch({ type: 'set_search', search: '' })

        let fileData: {
            self?: FileInitT
            children?: FileInitT[]
            parents?: FileInitT[]
            error?: any
        }
        if (fb.viewingPast !== null) {
            fileData = await getPastFolderInfo(
                fb.contentId,
                fb.viewingPast,
                authHeader
            )
        } else {
            fileData = await GetFolderData(
                fb.contentId,
                fb.fbMode,
                fb.shareId,
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
            SetFileData(fileData, dispatch, usr)
        }

        const jumpItem = query('jumpTo')
        if (jumpItem) {
            dispatch({ type: 'set_scroll_to', fileId: jumpItem })
            dispatch({
                type: 'set_selected',
                fileId: jumpItem,
                selected: true,
            })
        }
    }, [
        usr.username,
        authHeader,
        fb.contentId,
        fb.shareId,
        fb.fbMode,
        searchQuery,
        fb.viewingPast,
    ])

    useEffect(() => {
        syncState().then(() =>
            dispatch({ type: 'remove_loading', loading: 'files' })
        )
    }, [syncState])

    const queryClient = new QueryClient()

    if (!fb) {
        return <></>
    }

    return (
        <QueryClientProvider client={queryClient}>
            <FbContext.Provider value={{ fbState: fb, fbDispatch: dispatch }}>
                <div className="h-screen flex flex-col">
                    <HeaderBar
                        dispatch={dispatch}
                        page={'files'}
                        loading={fb.loading}
                    />
                    <DraggingCounter
                        dragging={fb.draggingState}
                        dirMap={fb.dirMap}
                        selected={fb.selected}
                        dispatch={dispatch}
                    />
                    <Presentation
                        itemId={fb.presentingId}
                        mediaData={fb.dirMap.get(fb.presentingId)?.GetMedia()}
                        element={() =>
                            PresentationFile({
                                file: fb.dirMap.get(fb.presentingId),
                            })
                        }
                        dispatch={dispatch}
                    />
                    <PasteImageDialogue
                        img={fb.pasteImg}
                        folderId={fb.contentId}
                        dirMap={fb.dirMap}
                        authHeader={authHeader}
                        dispatch={dispatch}
                        wsSend={wsSend}
                    />
                    <FileContextMenu />
                    <WebsocketStatus ready={readyState} />
                    <div className="flex flex-row grow h-[90vh] items-start">
                        <GlobalActions
                            uploadState={uploadState}
                            wsSend={wsSend}
                            uploadDispatch={uploadDispatch}
                        />
                        <DirViewWrapper>
                            <DirView
                                fb={fb}
                                notFound={notFound}
                                setNotFound={setNotFound}
                                searchRef={searchRef}
                                searchQuery={searchQuery}
                                searchFilter={searchFilter}
                                dispatch={dispatch}
                                wsSend={wsSend}
                                uploadDispatch={uploadDispatch}
                                authHeader={authHeader}
                            />
                        </DirViewWrapper>
                    </div>
                </div>
            </FbContext.Provider>
        </QueryClientProvider>
    )
}

export default FileBrowser
