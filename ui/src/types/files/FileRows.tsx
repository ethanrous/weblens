import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { historyDate } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { IconDisplay } from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
import { SelectedState, WeblensFile } from '@weblens/types/files/File'
import {
    fileHandleContextMenu,
    handleMouseLeave,
    handleMouseOver,
    handleMouseUp,
    mouseMove,
    visitFile,
} from '@weblens/types/files/FileDragLogic'
import { FileTextBox } from '@weblens/types/files/FileSquare'
import { useResize } from 'components/hooks'
import { useSessionStore } from 'components/UserInfo'
import React, { MouseEvent, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { FixedSizeList as WindowList } from 'react-window'
import '@weblens/types/files/filesStyle.scss'

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

    const {
        draggingState,
        fbMode,
        shareId,
        menuMode,
        folderInfo,
        selected,
        setMoveDest,
        setHovering,
        setSelected,
        setDragging,
        setPresentationTarget,
        setMenu,
        clearSelected,
        setSelectedMoved,
    } = useFileBrowserStore()

    const selState = useFileBrowserStore((state) =>
        state.filesMap.get(file.Id()).GetSelectedState()
    )

    return (
        <div style={style}>
            <div
                className="weblens-file animate-fade-short"
                data-row={true}
                data-clickable={!draggingState || file.IsFolder()}
                data-hovering={selState & SelectedState.Hovering}
                data-in-range={(selState & SelectedState.InRange) >> 1}
                data-selected={(selState & SelectedState.Selected) >> 2}
                data-last-selected={
                    (selState & SelectedState.LastSelected) >> 3
                }
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
                        fbMode,
                        shareId,
                        file,
                        folderInfo.IsTrash(),
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
                        Array.from(selected.keys()),
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
                <div className="flex h-full w-full items-center">
                    <div className="flex shrink-0 h-full aspect-square rounded overflow-hidden m-1 justify-center items-center">
                        <IconDisplay file={file} allowMedia={true} />
                    </div>
                    <FileTextBox itemTitle={file.GetFilename()} />
                </div>
                <div className="flex flex-col h-full">
                    <div
                        className="file-size-box"
                        data-moved={(selState & SelectedState.Moved) >> 5}
                    >
                        <p>{file.FormatSize()}</p>
                    </div>
                    <p className="absolute bottom-3 right-3 w-max text-sm">
                        {historyDate(file.GetModified().getTime())}
                    </p>
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
