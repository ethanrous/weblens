import { SelectedState, WeblensFile } from './File'
import React, { MouseEvent, useState } from 'react'
import { IconDisplay } from '../Pages/FileBrowser/FileBrowserMiscComponents'
import { useNavigate } from 'react-router-dom'
import {
    fileHandleContextMenu,
    handleMouseLeave,
    handleMouseOver,
    handleMouseUp,
    mouseMove,
    visitFile,
} from './FileDragLogic'
import { FileTextBox } from './FileSquare'
import { useResize } from '../components/hooks'
import { FixedSizeList as WindowList } from 'react-window'
import './filesStyle.scss'
import { useFileBrowserStore } from '../Pages/FileBrowser/FBStateControl'
import { useSessionStore } from '../components/UserInfo'

function FileRow({
    data,
    index,
    style,
}: {
    data: WeblensFile[]
    index: number
    style
}) {
    const file = data[index]
    const nav = useNavigate()
    const authHeader = useSessionStore((state) => state.auth)
    const [mouseDown, setMouseDown] = useState(null)

    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const selectedIds = useFileBrowserStore((state) =>
        Array.from(state.selected.keys())
    )

    const setBlockFocus = useFileBrowserStore((state) => state.setBlockFocus)
    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)
    const setHovering = useFileBrowserStore((state) => state.setHovering)
    const setSelected = useFileBrowserStore((state) => state.setSelected)
    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setPresentationTarget = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const setSelectedMoved = useFileBrowserStore(
        (state) => state.setSelectedMoved
    )

    const addLoading = useFileBrowserStore((state) => state.addLoading)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)

    const selected = useFileBrowserStore((state) =>
        state.filesMap.get(file.Id()).GetSelectedState()
    )

    return (
        <div style={style}>
            <div
                className="weblens-file animate-fade-short"
                data-row={true}
                data-clickable={!draggingState || file.IsFolder()}
                data-hovering={selected & SelectedState.Hovering}
                data-in-range={(selected & SelectedState.InRange) >> 1}
                data-selected={(selected & SelectedState.Selected) >> 2}
                data-last-selected={
                    (selected & SelectedState.LastSelected) >> 3
                }
                data-droppable={(selected & SelectedState.Droppable) >> 4}
                data-moved={(selected & SelectedState.Moved) >> 5}
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
                onClick={(e) => {
                    e.stopPropagation()
                    setSelected([file.Id()])
                }}
                onDoubleClick={(e) =>
                    visitFile(
                        e,
                        mode,
                        shareId,
                        file,
                        nav,
                        setPresentationTarget
                    )
                }
                onContextMenu={(e) =>
                    fileHandleContextMenu(e, menuMode, setMenu, file)
                }
                onMouseUp={() =>
                    handleMouseUp(
                        file,
                        draggingState,
                        selectedIds,
                        authHeader,
                        setSelectedMoved,
                        clearSelected,
                        setMoveDest,
                        setDragging,
                        setMouseDown
                    )
                }
                onMouseLeave={() =>
                    handleMouseLeave(
                        file,
                        draggingState,
                        setMoveDest,
                        setHovering,
                        mouseDown,
                        setMouseDown
                    )
                }
            >
                <div className="flex h-full items-center">
                    <div className="flex shrink-0 h-full aspect-square rounded overflow-hidden m-1 justify-center items-center">
                        <IconDisplay file={file} allowMedia={true} />
                    </div>
                    <FileTextBox itemTitle={file.GetFilename()} />
                </div>
                <div
                    className="file-size-box"
                    data-moved={(selected & SelectedState.Moved) >> 5}
                >
                    <p>{file.FormatSize()}</p>
                </div>
            </div>
        </div>
    )
}

export function FileRows({ files }: { files: WeblensFile[] }) {
    const [boxRef, setBoxRef] = useState<HTMLDivElement>()
    const size = useResize(boxRef)
    return (
        <div ref={setBoxRef} className="file-rows no-scrollbar">
            <WindowList
                height={size.height}
                width={size.width - 4}
                itemSize={88}
                itemCount={files.length}
                itemData={files}
                overscan={100}
            >
                {FileRow}
            </WindowList>
        </div>
    )
}
