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
    IconFileAnalytics,
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconInfoCircle,
    IconLogin,
    IconSearch,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
    IconX,
} from '@tabler/icons-react'

// Mantine
import {
    Box,
    Button,
    Divider,
    FileButton,
    Input,
    Space,
    Text,
} from '@mantine/core'
import { useDebouncedValue, useMouse } from '@mantine/hooks'
import { notifications } from '@mantine/notifications'

// Weblens
import { UserContext } from '../../Context'
import {
    GetFileInfo,
    GetFolderData,
    SearchFolder,
    getPastFolderInfo,
    moveFiles,
} from '../../api/FileBrowserApi'
import Crumbs, { StyledBreadcrumb } from '../../components/Crumbs'
import HeaderBar from '../../components/HeaderBar'
import { GlobalContextType } from '../../components/ItemDisplay'
import { ItemScroller } from '../../components/ItemScroller'
import NotFound from '../../components/NotFound'
import { MediaImage } from '../../components/PhotoContainer'
import Presentation, {
    PresentationContainer,
} from '../../components/Presentation'
import UploadStatus, { useUploadStatus } from './UploadStatus'
import './style/fileBrowserStyle.css'
import '../../components/style.scss'
import {
    AuthHeaderT,
    FileBrowserAction,
    FBDispatchT,
    FbStateT,
    UserContextT,
} from '../../types/Types'
import WeblensMedia from '../../classes/Media'
import { WeblensFile } from '../../classes/File'
import { humanFileSize } from '../../util'

import {
    HandleDrop,
    HandleUploadButton,
    MoveSelected,
    SetFileData,
    uploadViaUrl,
    downloadSelected,
    fileBrowserReducer,
    getRealId,
    useKeyDownFileBrowser,
    usePaste,
    handleDragOver,
} from './FileBrowserLogic'

import { GetFilesContext, GetItemsList } from './FilesContext'
import {
    DirViewWrapper,
    DropSpot,
    FileInfoDisplay,
    GetStartedCard,
    IconDisplay,
    PresentationFile,
    RowBox,
    TransferCard,
    WebsocketStatus,
} from './FileBrowserStyles'
import { WeblensButton } from '../../components/WeblensButton'
import { FileRows } from './FileRows'
import { useResize, useResizeDrag, useWindowSize } from '../../components/hooks'
import { WeblensProgress } from '../../components/WeblensProgress'
import { IconFiles } from '@tabler/icons-react'
import { FilesPane } from './FileInfoPane'
import { StatTree } from './FileStatTree'
import { FileSortBox } from './FileSortBox'
import { FileContextMenu } from './FileMenu'
import { useSubscribe } from '../../api/Websocket'
import { TasksDisplay } from './TaskProgress'

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
        return null
    }
    const media = new WeblensMedia({ mediaId: 'paste' })
    media.SetThumbnailBytes(img)

    return (
        <PresentationContainer
            shadeOpacity={'0.25'}
            onClick={() => dispatch({ type: 'paste_image', img: null })}
        >
            <Box
                style={{
                    position: 'absolute',
                    justifyContent: 'center',
                    alignItems: 'center',
                    zIndex: 2,
                }}
            >
                <Text fw={700} size="40px" style={{ paddingBottom: '50px' }}>
                    Upload from clipboard?
                </Text>
                <Box
                    onClick={(e) => {
                        e.stopPropagation()
                    }}
                    style={{
                        height: '50%',
                        width: 'max-content',
                        backgroundColor: '#222277ee',
                        padding: '10px',
                        borderRadius: '8px',
                        overflow: 'hidden',
                    }}
                >
                    <MediaImage media={media} quality="thumbnail" />
                </Box>
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
            </Box>
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

        return (
            <div
                className="flex flex-row items-start w-full h-full grow-0 shrink-0"
                style={{
                    width: resizeOffset,
                }}
            >
                {usr.isLoggedIn === false && (
                    <Box className="login-required-background">
                        <WeblensButton
                            label="Login"
                            height={48}
                            Left={<IconLogin className="button-icon" />}
                            centerContent
                            onClick={() => nav('/login')}
                            style={{ maxWidth: 300 }}
                        />
                    </Box>
                )}

                <div className="sidebar-container">
                    <WeblensButton
                        label="Home"
                        height={48}
                        toggleOn={
                            fbState.folderInfo.Id() === usr?.homeId &&
                            fbState.fbMode === 'default'
                        }
                        width={'100%'}
                        allowRepeat={false}
                        Left={<IconHome className="button-icon" />}
                        onMouseOver={homeMouseOver}
                        onMouseLeave={homeMouseLeave}
                        onMouseUp={homeMouseUp}
                    />

                    <WeblensButton
                        label="Shared"
                        height={48}
                        toggleOn={fbState.fbMode === 'share'}
                        disabled={
                            fbState.draggingState !== DraggingState.NoDrag
                        }
                        allowRepeat={false}
                        Left={<IconUsers className="button-icon" />}
                        width={'100%'}
                        onClick={() => {
                            nav('/files/shared')
                        }}
                    />

                    <WeblensButton
                        label="Trash"
                        height={48}
                        toggleOn={
                            fbState.folderInfo.Id() === usr?.trashId &&
                            fbState.fbMode === 'default'
                        }
                        disabled={
                            fbState.draggingState !== DraggingState.NoDrag &&
                            fbState.folderInfo.Id() === usr?.trashId &&
                            fbState.fbMode === 'default'
                        }
                        allowRepeat={false}
                        Left={<IconTrash className="button-icon" />}
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
                            height={48}
                            toggleOn={fbState.fbMode === 'external'}
                            allowRepeat={false}
                            Left={<IconServer className="button-icon" />}
                            style={{ margin: 0 }}
                            disabled={
                                fbState.draggingState !== DraggingState.NoDrag
                            }
                            onClick={(e) => {
                                e.stopPropagation()
                                nav('/files/external')
                            }}
                        />
                    )}

                    <WeblensButton
                        label="New Folder"
                        height={48}
                        Left={<IconFolderPlus className="button-icon" />}
                        showSuccess={false}
                        disabled={
                            fbState.draggingState !== 0 ||
                            !fbState.folderInfo.IsModifiable()
                        }
                        onClick={(e) => {
                            e.stopPropagation()
                            fbDispatch({ type: 'new_dir' })
                        }}
                        width={'100%'}
                    />

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
                                    height={48}
                                    showSuccess={false}
                                    disabled={
                                        fbState.draggingState !== 0 ||
                                        !fbState.folderInfo.IsModifiable()
                                    }
                                    Left={
                                        <IconUpload className="button-icon" />
                                    }
                                    width={'100%'}
                                    onClick={() => props.onClick()}
                                />
                            )
                        }}
                    </FileButton>

                    <Divider w={'100%'} my="lg" size={1.5} />

                    <UsageInfo />

                    <TasksDisplay scanProgress={fbState.scanProgress} />

                    <UploadStatus
                        uploadState={uploadState}
                        uploadDispatch={uploadDispatch}
                    />
                </div>
                <Box
                    draggable={false}
                    className="resize-bar-wrapper"
                    onMouseDown={(e) => {
                        e.preventDefault()
                        setResizing(true)
                    }}
                >
                    <Box className="resize-bar" />
                </Box>
            </div>
        )
    },
    (p, n) => {
        if (p.uploadState !== n.uploadState) {
            return false
        }
        return true
    }
)

const UsageInfo = ({}) => {
    const [box, setBox] = useState(null)
    const size = useResize(box)
    const nav = useNavigate()

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
        <Box
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

                    <WeblensButton
                        centerContent
                        toggleOn={fbState.fbMode === 'stats'}
                        height={40}
                        width={40}
                        Left={<IconFileAnalytics size={'30px'} />}
                        onClick={() =>
                            nav(
                                `/files/stats/${
                                    fbState.fbMode === 'external'
                                        ? fbState.fbMode
                                        : fbState.folderInfo.Id()
                                }`
                            )
                        }
                    />
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
        </Box>
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
        if (fb.fbMode !== 'search') {
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
        <Box
            style={{
                height: 'max-content',
                width: 'max-content',
                alignItems: 'flex-start',
                marginRight: 5,
            }}
        >
            <Box className="search-box">
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
            </Box>
            {hintOpen && (
                <Box className="search-hint-box">
                    <Box className="key-line">
                        <Text>Press</Text>
                        <Text className="key-display">Enter</Text>
                        <Text>to search all files</Text>
                    </Box>
                </Box>
            )}
            {error && (
                <Box className="search-hint-box" mod={{ 'data-error': 'true' }}>
                    <Text>Not valid regex</Text>
                </Box>
            )}
        </Box>
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
        <Box
            style={{
                position: 'fixed',
                top: position.y + 8,
                left: position.x + 8,
                zIndex: 10,
            }}
            onMouseUp={() => {
                dispatch({ type: 'set_dragging', dragging: false })
            }}
        >
            {Boolean(files) && (
                <div className="flex flex-row h-max">
                    <IconFile size={30} />
                    <Space w={10} />
                    <StyledBreadcrumb
                        label={files.toString()}
                        fontSize={30}
                        compact
                    />
                </div>
            )}
            {Boolean(folders) && (
                <div className="flex flex-row h-max">
                    <IconFolder size={30} />
                    <Space w={10} />
                    <StyledBreadcrumb
                        label={folders.toString()}
                        fontSize={30}
                        compact
                    />
                </div>
            )}
        </Box>
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
        <Box
            style={{
                width: '100%',
                height: '100%',
                display: 'flex',
                flexDirection: 'row',
                justifyContent: 'space-around',
                paddingBottom: 8,
            }}
        >
            <Box
                className="icon-display-wrapper"
                style={{
                    display: 'flex',
                    width: 150,
                    maxWidth: '65%',
                    flexGrow: 1,
                    alignItems: 'center',
                }}
            >
                <IconDisplay file={file} allowMedia size={'65%'} />
            </Box>
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
        </Box>
    )
}

function Files({
    fb: fb,
    notFound,
    setNotFound,
    searchRef,
    dispatch,
    wsSend,
    uploadDispatch,
    authHeader,
}: {
    fb: FbStateT
    notFound
    setNotFound
    searchRef
    dispatch: (action: FileBrowserAction) => void
    wsSend: (action: string, content: any) => void
    uploadDispatch
    authHeader
}) {
    const { usr }: UserContextT = useContext(UserContext)
    const nav = useNavigate()
    const [debouncedSearch] = useDebouncedValue(fb.searchContent, 200)

    const [fullViewRef, setFullViewRef] = useState(null)
    useResize(fullViewRef)
    const [contentViewRef, setContentViewRef] = useState(null)

    const moveSelectedTo = useCallback(
        (folderId: string) => {
            MoveSelected(fb.selected, folderId, authHeader)
            dispatch({ type: 'clear_selected' })
        },
        [fb.selected.size, fb.contentId, authHeader]
    )

    const { files, hoveringIndex, lastSelectedIndex } = useMemo(() => {
        return GetItemsList(fb, usr, debouncedSearch)
    }, [
        fb.dirMap,
        fb.holdingShift,
        fb.selected,
        fb.hovering,
        debouncedSearch,
        usr,
        fb.lastSelected,
        fb.sortFunc,
        fb.sortDirection,
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
        return Array.from(fb.selected.keys()).map((fId) => fb.dirMap.get(fId))
    }, [fb.selected])

    const itemsCtx: GlobalContextType = useMemo(() => {
        return GetFilesContext(
            fb,
            files,
            hoveringIndex,
            lastSelectedIndex,
            authHeader,
            dispatch
        )
    }, [
        files,
        fb.contentId,
        fb.dirMap,
        fb.selected,
        fb.fbMode,
        fb.draggingState,
        fb.hovering,
        fb.holdingShift,
        dispatch,
    ])
    const dropAllowed = useMemo(() => {
        return (
            fb.folderInfo.IsModifiable() &&
            !(fb.fbMode === 'share' || fb.contentId === usr.trashId)
        )
    }, [fb.contentId, usr.trashId, fb.fbMode, fb.folderInfo])

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
                destination={fb.moveDest}
                boundRef={fullViewRef}
            />
            <div className="flex flex-row h-max justify-between p-2">
                <Crumbs
                    finalFile={fb.folderInfo}
                    postText={
                        fb.viewingPast
                            ? `@ ${fb.viewingPast.toDateString()} ${fb.viewingPast.toLocaleTimeString()}`
                            : ''
                    }
                    navOnLast={false}
                    dragging={fb.draggingState}
                    moveSelectedTo={moveSelectedTo}
                    setMoveDest={(itemName) =>
                        dispatch({ type: 'set_move_dest', fileName: itemName })
                    }
                />

                <FileSearch fb={fb} searchRef={searchRef} dispatch={dispatch} />
                <FileSortBox />
                <WeblensButton
                    Left={
                        fb.fileInfoMenu ? (
                            <IconX className="button-icon" />
                        ) : (
                            <IconInfoCircle className="button-icon" />
                        )
                    }
                    height={42}
                    width={42}
                    subtle
                    onClick={() =>
                        dispatch({
                            type: 'set_file_info_menu',
                            open: !fb.fileInfoMenu,
                        })
                    }
                />
            </div>
            <div className="flex flex-row h-[200px] grow max-w-full">
                <div className="grow shrink w-0" ref={setContentViewRef}>
                    <DropSpot
                        onDrop={(e) => {
                            HandleDrop(
                                e.dataTransfer.items,
                                fb.contentId,
                                Array.from(fb.dirMap.values()).map(
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
                        dropSpotTitle={fb.folderInfo.GetFilename()}
                        dragging={fb.draggingState}
                        dropAllowed={dropAllowed}
                        handleDrag={(event) =>
                            handleDragOver(event, dispatch, fb.draggingState)
                        }
                        wrapperRef={contentViewRef}
                    />
                    {(files.length !== 0 && (
                        <ItemScroller
                            itemsContext={files}
                            globalContext={itemsCtx}
                            parentNode={contentViewRef}
                            dispatch={dispatch}
                        />
                    )) ||
                        (fb.loading.length === 0 &&
                            fb.searchContent === '' &&
                            fb.searchContent === debouncedSearch && (
                                <GetStartedCard
                                    fb={fb}
                                    dispatch={dispatch}
                                    uploadDispatch={uploadDispatch}
                                    wsSend={wsSend}
                                />
                            )) ||
                        (fb.loading.length === 0 &&
                            fb.folderInfo.Id() === 'shared' && (
                                <NotFound
                                    resourceType="any files shared with you"
                                    link="/files/home"
                                    setNotFound={setNotFound}
                                />
                            )) ||
                        (fb.loading.length === 0 && fb.searchContent !== '' && (
                            <Box className="flex justify-flex-end h-1/5">
                                <Box className="flex flex-row h-max w-max">
                                    <Text size="20px">
                                        No items match your search in
                                    </Text>
                                    <IconFolder style={{ marginLeft: 4 }} />
                                    <Text size="20px">
                                        {fb.folderInfo.GetFilename()}
                                    </Text>
                                </Box>
                                <Space h={10} />

                                <Box className="key-line">
                                    <Text size="16px">Press</Text>
                                    <Text className="key-display">Enter</Text>
                                    <Text>to search all files</Text>
                                </Box>
                            </Box>
                        ))}
                </div>
                <FilesPane
                    open={fb.fileInfoMenu}
                    setOpen={(o) =>
                        dispatch({ type: 'set_file_info_menu', open: o })
                    }
                    selectedFiles={selectedInfo}
                    timestamp={fb.viewingPast?.getTime()}
                    contentId={fb.contentId}
                    dispatch={dispatch}
                />
            </div>
        </div>
    )
}

function SearchResults({
    fb,
    searchQuery,
    filter,
    searchRef,
    dispatch,
}: {
    fb: FbStateT
    searchQuery: string
    filter: string
    searchRef
    dispatch: FBDispatchT
}) {
    const nav = useNavigate()
    let titleString: string = 'Searching '
    if (searchQuery) {
        titleString += `for ${searchQuery}`
    } else {
        titleString += `all files`
    }
    titleString += ` in ${fb.folderInfo.GetFilename()}`
    if (filter) {
        titleString += ` ending with .${filter}`
    }

    return (
        <Box className="flex flex-col h-full">
            <Box className="flex h-14 flex-shrink-0 w-full items-center">
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
                    {fb.dirMap.size} results
                </Text>
                <Space flex={1} />
                <FileSearch
                    fb={fb}
                    defaultOpen
                    searchRef={searchRef}
                    dispatch={dispatch}
                />
            </Box>
            <Space h={10} />
            <FileRows fb={fb} dispatch={dispatch} />
        </Box>
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
        fb.fbMode === 'default' &&
        fb.folderInfo.Id() &&
        !fb.folderInfo.IsFolder()
    ) {
        return <SingleFile file={fb.folderInfo} doDownload={download} />
    } else if (fb.fbMode === 'stats') {
        return <StatTree folderInfo={fb.folderInfo} authHeader={authHeader} />
    } else if (fb.fbMode === 'search') {
        return (
            <SearchResults
                fb={fb}
                searchQuery={searchQuery}
                filter={searchFilter}
                searchRef={searchRef}
                dispatch={dispatch}
            />
        )
    } else {
        return (
            <Files
                fb={fb}
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
    const getQuery = useCallback(
        (s) => {
            const r = q.get(s)
            if (!r) {
                return ''
            }
            return r
        },
        [q]
    )
    return getQuery
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
            uploadMap: new Map<string, boolean>(),
            selected: new Map<string, boolean>(),
            dirMap: new Map<string, WeblensFile>(),
            viewingPast: null,
            folderInfo: new WeblensFile({ isDir: true }),
            menuPos: { x: 0, y: 0 },
            waitingForNewName: '',
            holdingShift: false,
            fileInfoMenu: false,
            blockFocus: false,
            searchResults: [],
            searchContent: '',
            lastSelected: '',
            draggingState: 0,
            scanProgress: [],
            menuTargetId: '',
            presentingId: '',
            sortDirection: 1,
            sortFunc: 'Name',
            menuOpen: false,
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
            fbMode: '',
        })

    if (fb.fbMode && fb.fbMode !== 'share' && usr.isLoggedIn === false) {
        nav('/login')
    }

    useEffect(() => {
        if (usr.isLoggedIn === undefined) {
            return
        }
        dispatch({ type: 'add_loading', loading: 'files' })

        let mode: string
        let contentId: string
        let shareId: string
        const splitPath = urlPath.split('/').filter((s) => s.length !== 0)

        if (splitPath.length === 0) {
            return
        }

        if (splitPath[0] === 'shared') {
            mode = 'share'
            shareId = splitPath[1]
            contentId = splitPath[2]
        } else if (splitPath[0] === 'external') {
            mode = 'external'
            contentId = splitPath[1]
        } else if (splitPath[0] === 'stats') {
            mode = 'stats'
            contentId = splitPath[1]
        } else if (splitPath[0] === 'search') {
            mode = 'search'
            contentId = splitPath[1]
        } else {
            mode = 'default'
            contentId = splitPath[0]
        }

        getRealId(contentId, mode, usr, authHeader).then((realId) => {
            dispatch({
                type: 'set_location_state',
                realId: realId,
                mode: mode,
                shareId: shareId,
            })
            dispatch({ type: 'remove_loading', loading: 'files' })
        })
    }, [urlPath, dispatch, authHeader, usr])

    const { wsSend, readyState } = useSubscribe(
        fb.contentId,
        fb.folderInfo.Id(),
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
    useEffect(() => {
        const syncState = async () => {
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

            if (fb.fbMode === 'search') {
                const folderData = await GetFileInfo(
                    fb.contentId,
                    '',
                    authHeader
                )
                if (!folderData) {
                    return
                }

                const searchResults = await SearchFolder(
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

            let fileData
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
                        notifications.show({
                            title: 'Could not get folder info',
                            message: String(r),
                            color: 'red',
                            autoClose: 5000,
                        })
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

            dispatch({ type: 'remove_loading', loading: 'files' })
        }
        syncState()
    }, [
        usr.username,
        authHeader,
        fb.contentId,
        fb.fbMode,
        searchQuery,
        fb.viewingPast,
    ])

    return (
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
                <FileContextMenu
                    itemId={fb.menuTargetId}
                    setOpen={(o) =>
                        dispatch({ type: 'set_menu_open', open: o })
                    }
                    menuPos={fb.menuPos}
                    wsSend={wsSend}
                />
                <WebsocketStatus ready={readyState} />
                <div className="flex flex-row grow h-[90vh] items-start">
                    <GlobalActions
                        uploadState={uploadState}
                        wsSend={wsSend}
                        uploadDispatch={uploadDispatch}
                    />
                    <DirViewWrapper
                        folderName={fb.folderInfo?.GetFilename()}
                        dragging={fb.draggingState}
                    >
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
    )
}

export default FileBrowser
