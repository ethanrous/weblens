import { SelectedState, WeblensFile } from '@weblens/types/files/File'
import '@weblens/types/files/filesStyle.scss'
import '@weblens/components/style.scss'
import {
    fileHandleContextMenu,
    goToFile,
    handleMouseLeave,
    handleMouseOver,
    handleMouseUp,
    mouseMove,
    visitFile,
} from './FileDragLogic'
import {
    createRef,
    CSSProperties,
    MouseEvent,
    useEffect,
    useMemo,
    useState,
} from 'react'
import {
    FbModeT,
    useFileBrowserStore,
} from '@weblens/pages/FileBrowser/FBStateControl'
import {
    GetStartedCard,
    IconDisplay,
} from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
import { GetFolderData } from '@weblens/api/FileBrowserApi'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useKeyDown, useResize, useResizeDrag } from '@weblens/components/hooks'
import { FixedSizeList as WindowList } from 'react-window'
import WeblensLoader from '@weblens/components/Loading'
import { IconChevronRight, IconFile } from '@tabler/icons-react'
import { DirViewModeT } from '@weblens/pages/FileBrowser/FileBrowserTypes'
import { DraggingStateT } from './FBTypes'
import { clamp } from '@mantine/hooks'
import { MediaImage } from '../media/PhotoContainer'
import { PhotoQuality } from '../media/Media'
import { useMediaStore } from '../media/MediaStateControl'
import { humanFileSize } from '@weblens/util'

function ColumnRow(p: { data; index: number; style: CSSProperties }) {
    const file = p.data.files[p.index]
    const [mouseDown, setMouseDown] = useState<{ x: number; y: number }>(null)
    let selState = useFileBrowserStore((state) => {
        return state.filesMap.get(file?.Id())?.GetSelectedState()
    })
    if (p.data.selectedChildId && file.Id() === p.data.selectedChildId) {
        selState = SelectedState.Selected
    }
    const [fileRef, setFileRef] = useState<HTMLDivElement>()

    const {
        draggingState,
        folderInfo,
        selected,
        lastSelectedId,
        setMoveDest,
        setHovering,
        setSelected,
        setDragging,
        setPresentationTarget,
        setMenu,
        clearSelected,
        setSelectedMoved,
    } = useFileBrowserStore()

    useEffect(() => {
        if (file.Id() === lastSelectedId) {
            if (fileRef) {
                fileRef.scrollIntoView({
                    behavior: 'instant',
                    block: 'nearest',
                    inline: 'nearest',
                })
            }
        }
    }, [lastSelectedId])

    return (
        <div ref={setFileRef} style={p.style}>
            <div
                key={file.Id()}
                className="weblens-file animate-fade-short"
                data-column-row
                data-clickable={!draggingState || file.IsFolder()}
                data-hovering={selState & SelectedState.Hovering}
                data-in-range={(selState & SelectedState.InRange) >> 1}
                data-selected={(selState & SelectedState.Selected) >> 2}
                data-last-selected={
                    (selState & SelectedState.LastSelected) >> 3 && selected.size < 2
                }
                data-current-view={file.Id() === lastSelectedId}
                data-droppable={(selState & SelectedState.Droppable) >> 4}
                data-moved={(selState & SelectedState.Moved) >> 5}
                onMouseOver={(e: MouseEvent<HTMLDivElement>) =>
                    handleMouseOver(
                        e,
                        file,
                        draggingState,
                        setHovering,
                        setMoveDest
                    )
                }
                onMouseDown={(e) => {
                    setMouseDown({ x: e.clientX, y: e.clientY })
                }}
                onMouseUp={(e) => {
                    e.stopPropagation()
                    return handleMouseUp(
                        file,
                        draggingState,
                        Array.from(selected.keys()),
                        setSelectedMoved,
                        clearSelected,
                        setMoveDest,
                        setDragging,
                        setMouseDown,
                        DirViewModeT.Columns
                    )
                }}
                onContextMenu={(e) => fileHandleContextMenu(e, setMenu, file)}
                onClick={(e) => {
                    e.stopPropagation()
                    if (draggingState) {
                        return
                    }
                    setSelected([file.Id()])
                }}
                onDoubleClick={(e) =>
                    visitFile(
                        e,
                        file,
                        folderInfo.IsTrash(),
                        setPresentationTarget
                    )
                }
                onMouseMove={(e) =>
                    mouseMove(
                        e,
                        file,
                        draggingState,
                        mouseDown,
                        setSelected,
                        setDragging
                    )
                }
                onMouseLeave={() =>
                    handleMouseLeave(
                        file,
                        draggingState,
                        setMoveDest,
                        setHovering,
                        Boolean(mouseDown),
                        setMouseDown
                    )
                }
            >
                <div className="flex items-center h-[40px] max-h-full gap-3 w-full">
                    <div className="flex shrink-0 justify-center items-center w-[40px] h-[40px] max-w-[40px] max-h-full">
                        <IconDisplay file={file} allowMedia={true} />
                    </div>
                    <div className="file-text-container">
                        <p className="file-text">{file.GetFilename()}</p>
                    </div>
                </div>
                {file.IsFolder() && (
                    <IconChevronRight
                        className="text-wl-outline-subtle"
                        style={{
                            color:
                                file.Id() === p.data.selectedChildId
                                    ? '#ffffff'
                                    : '',
                        }}
                    />
                )}
            </div>
        </div>
    )
}

function Preview({ file }: { file: WeblensFile }) {
    const media = useMediaStore((state) =>
        state.mediaMap.get(file?.GetContentId())
    )

    if (!file || file.IsFolder()) {
        return null
    }

    return (
        <div
            className="flex flex-col w-full p-4"
            onClick={(e) => e.stopPropagation()}
        >
            <div className="flex justify-center items-center w-full p-4 h-[75%]">
                {media && (
                    <MediaImage
                        media={media}
                        quality={PhotoQuality.LowRes}
                        fitLogic="contain"
                    />
                )}
                {!media && <IconFile />}
            </div>
            <div>
                <p>{file.GetFilename()}</p>
                <p>{humanFileSize(file.size)}</p>
            </div>
        </div>
    )
}

function Column({
    files,
    parentId,
    selectedChildId,
    setColSize,
}: {
    files: WeblensFile[]
    parentId: string
    selectedChildId: string
    setColSize: (size: number) => void
}) {
    const { fbMode, shareId, folderInfo, filesMap, setFilesData } =
        useFileBrowserStore()
    const user = useSessionStore((state) => state.user)
    const [loading, setLoading] = useState(true)
    const listRef = createRef<WindowList>()
    const [didScroll, setDidScroll] = useState<boolean>()
    const draggingState = useFileBrowserStore((state) => state.draggingState)

    useEffect(() => {
        if (!folderInfo) {
            return
        }
        const parents = [...folderInfo.parents]
        if (
            folderInfo.Id() !== parentId &&
            parents.filter((p) => p.Id() === parentId).length === 0
        ) {
            return
        }

        const p = filesMap.get(parentId)
        if (p.GetChildren().length === files.length) {
            setLoading(false)
            return
        }
        setLoading(true)

        const fetch = async () => {
            const preId = useFileBrowserStore
                .getState()
                .folderInfo?.parents.map((p) => p.Id())
            preId.push(useFileBrowserStore.getState().folderInfo?.Id())

            const fileData = await GetFolderData(
                parentId,
                fbMode,
                shareId,
                null
            )

            // If the above request takes too long, and we change folders in that time...
            // do not add the files to the state
            if (
                !preId.includes(useFileBrowserStore.getState().folderInfo?.Id())
            ) {
                return
            }

            setFilesData({
                childrenInfo: fileData.children,
                parentsInfo: fileData.parents,
                mediaData: fileData.medias,
                user,
            })
        }
        fetch()
    }, [folderInfo, files.length])

    const [boxRef, setBoxRef] = useState<HTMLDivElement>()
    const size = useResize(boxRef)

    useEffect(() => {
        if (files.length === 0) {
            return
        }
        for (const [i, file] of files.entries()) {
            file.SetIndex(i)
        }
    }, [files])

    useEffect(() => {
        if (!selectedChildId) {
            return
        }

        if (listRef.current) {
            const child = useFileBrowserStore
                .getState()
                .filesMap.get(selectedChildId)
            listRef.current.scrollToItem(child ? child.GetIndex() : 0)
        }
    }, [selectedChildId])

    return (
        <div
            ref={setBoxRef}
            className="flex shrink-0 justify-center h-full no-scrollbar gap-1 w-full"
            onClick={(e) => {
                e.stopPropagation()
                if (draggingState !== DraggingStateT.NoDrag) {
                    return
                }
                const parent = filesMap.get(parentId)
                goToFile(parent)
            }}
        >
            {loading && (
                <div className="flex grow justify-center">
                    <WeblensLoader />
                </div>
            )}
            {!loading && (
                <WindowList
                    ref={listRef}
                    height={size.height}
                    width={size.width}
                    itemSize={56}
                    itemCount={files.length}
                    itemData={{ files, selectedChildId }}
                    overscan={100}
                    onItemsRendered={() => {
                        if (didScroll) {
                            return
                        }
                        // Grid ref is not ready yet even when this callback is called,
                        // but putting it in a timeout will push it off to the next tick,
                        // and the ref will be ready.
                        setTimeout(() => {
                            if (listRef.current && selectedChildId) {
                                const child = useFileBrowserStore
                                    .getState()
                                    .filesMap.get(selectedChildId)
                                if (child) {
                                    listRef.current.scrollToItem(
                                        child.GetIndex(),
                                        'smart'
                                    )
                                    setDidScroll(true)
                                } else {
                                    console.error(
                                        'Could not find child to scroll to',
                                        selectedChildId
                                    )
                                }
                            }
                        }, 1)
                    }}
                >
                    {ColumnRow}
                </WindowList>
            )}
            <ColumnResizer setColSize={setColSize} left={boxRef?.offsetLeft} />
        </div>
    )
}

function ColumnResizer({
    setColSize,
    left,
}: {
    setColSize: (size: number) => void
    left: number
}) {
    const dragging = useFileBrowserStore(
        (state) => state.draggingState === DraggingStateT.InterfaceDrag
    )
    const [localDragging, setLocalDragging] = useState(false)
    const setDraggingGlobal = useFileBrowserStore((state) => state.setDragging)
    const setDragging = (d: DraggingStateT) => {
        setDraggingGlobal(d)
        setLocalDragging(d === DraggingStateT.InterfaceDrag)
    }

    useEffect(() => {
        if (!dragging && localDragging) {
            setDragging(DraggingStateT.NoDrag)
        }
    }, [dragging])

    useResizeDrag(
        dragging,
        (dragging: boolean) => {
            if (dragging) {
                setDragging(DraggingStateT.InterfaceDrag)
            } else {
                setDragging(DraggingStateT.NoDrag)
            }
        },
        (v) => {
            if (!left || !localDragging) {
                return
            }
            const newSize = v - left + 8

            setColSize(clamp(newSize, 200, 800))
        },
        false
    )

    return (
        <div
            draggable={false}
            className="resize-bar-wrapper"
            onMouseDown={(e) => {
                e.preventDefault()
                setDragging(DraggingStateT.InterfaceDrag)
            }}
            onMouseUp={(e) => {
                e.stopPropagation()
                setDragging(DraggingStateT.NoDrag)
            }}
            onClick={(e) => {
                e.stopPropagation()
            }}
        >
            <div className="resize-bar" />
        </div>
    )
}

function FileColumns() {
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const filesLists = useFileBrowserStore((state) => state.filesLists)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const lastSelectedId = useFileBrowserStore((state) => state.lastSelectedId)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const clearFiles = useFileBrowserStore((state) => state.clearFiles)

    const [endRef, setEndRef] = useState<HTMLDivElement>()
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const [colWidths, setColWidths] = useState([])

    const { lastSelected, currentCol } = useMemo(() => {
        const lastSelected = filesMap.get(lastSelectedId)

        let currentCol: WeblensFile[] = []
        if (lastSelected) {
            currentCol = filesLists.get(lastSelected?.parentId)
        }

        return { lastSelected, currentCol }
    }, [lastSelectedId, filesMap])

    useKeyDown(
        (e: KeyboardEvent) =>
            [
                'ArrowLeft',
                'ArrowRight',
                'ArrowDown',
                'ArrowUp',
                'h',
                'j',
                'k',
                'l',
            ].includes(e.key),
        (e) => {
            e.stopPropagation()
            e.preventDefault()

            // Allow for vim keybindings
            let key = e.key
            if (key === 'h') {
                key = 'ArrowLeft'
            } else if (key === 'j') {
                key = 'ArrowDown'
            } else if (key === 'k') {
                key = 'ArrowUp'
            } else if (key === 'l') {
                key = 'ArrowRight'
            }

            if (mode === FbModeT.share && !folderInfo?.Id()) {
                if (key === 'ArrowUp') {
                    clearFiles()
                    useFileBrowserStore.getState().nav('/files/home')
                    return
                } else if (key === 'ArrowDown') {
                    clearFiles()
                    useFileBrowserStore.getState().nav('/files/trash')
                    return
                }
            }

            let nextItem: WeblensFile

            if (!lastSelectedId || !lastSelected) {
                console.error('No lastSelected in column keydown')
                return
            }

            if (key === 'ArrowLeft') {
                nextItem = filesMap.get(lastSelected.parentId)
            } else if (key === 'ArrowRight') {
                if (!lastSelected.isDir) {
                    return
                }
                const nextCol = filesLists.get(lastSelected.Id())
                if (!nextCol?.length) {
                    // There is no files in the next column, so we can't go right
                    return
                }
                nextItem = nextCol[0]
            } else if (key === 'ArrowDown' || key === 'ArrowUp') {
                if (
                    key === 'ArrowDown' &&
                    folderInfo.Id() ===
                        useSessionStore.getState().user.homeId &&
                    lastSelectedId === folderInfo.Id()
                ) {
                    clearFiles()
                    useFileBrowserStore.getState().nav('/files/shared')
                    return
                }

                const nextIndex =
                    lastSelected.GetIndex() + (key === 'ArrowDown' ? 1 : -1)
                if (nextIndex < 0) {
                    // We are already at the top, so we can't go up
                    return
                }

                if (nextIndex >= currentCol.length) {
                    // We are already at the bottom, so we can't go down
                    return
                }

                nextItem = currentCol[nextIndex]
            }
            if (!nextItem) {
                console.error('No nextItem in column keydown')
                return
            }
            goToFile(nextItem, false)
        }
    )

    const lists = useMemo(() => {
        if (!folderInfo) {
            return []
        }

        const lists: { parentId: string; files: WeblensFile[] }[] = []
        const parents = [...folderInfo.parents]
        parents.push(folderInfo)
        for (const p of parents) {
            const files = filesLists.get(p.Id())
            if (!files) {
                lists.push({ parentId: p.Id(), files: [] })
                continue
            }
            lists.push({ parentId: p.Id(), files })
        }

        return lists
    }, [filesLists, folderInfo])

    useEffect(() => {
        setColWidths((widths) => {
            if (widths.length === lists.length) {
                return widths
            }

            if (widths.length < lists.length + 1) {
                return [...widths, 300]
            }

            return widths
        })
    }, [lists])

    const size = useResize(containerRef)
    useEffect(() => {
        return
        if (endRef) {
            endRef.scrollIntoView({
                behavior: 'instant',
                block: 'nearest',
                inline: 'start',
            })
        }
    }, [size.width, lists.length])

    if (lists && lists.length < 2 && !lists[0]?.files?.length) {
        return <GetStartedCard />
    }

    return (
        <div
            ref={setContainerRef}
            className="flex flex-row h-full w-full outline-0 overflow-x-scroll"
        >
            {lists.map((col, i) => {
                const selectedChildId =
                    lastSelected?.parentId === col.parentId
                        ? lastSelectedId
                        : lists[i + 1]?.parentId

                return (
                    <div
                        key={col.parentId}
                        className="flex shrink-0"
                        style={{ width: colWidths[i] ?? 300 }}
                    >
                        <Column
                            key={col.parentId}
                            files={col.files}
                            parentId={col.parentId}
                            selectedChildId={selectedChildId}
                            setColSize={(size: number) =>
                                setColWidths((w) => {
                                    w[i] = size
                                    return [...w]
                                })
                            }
                        />
                    </div>
                )
            })}
            <div
                className="flex shrink-0 grow w-[40vw]"
                // style={{ width: colWidths[lists.length + 1] ?? 300 }}
            >
                <Preview file={lastSelected} />
            </div>
            <div ref={setEndRef} />
        </div>
    )
}

export default FileColumns
